// Package fourbyte provides a crawler for processing pages and saving signatures to a database.
package fourbyte

import (
	"context"
	"encoding/binary"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/txpull/unpack/clients"
	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/db/models"
	"github.com/txpull/unpack/helpers"
	"github.com/txpull/unpack/scanners"
	"github.com/txpull/unpack/types"
	"go.uber.org/zap"
)

// LAST_PROCESSED_PAGE_KEY is the key for the last processed page number.
const LAST_PROCESSED_PAGE_KEY = "last_processed_fourbyte_page"

type FourByteWriter struct {
	ctx          context.Context            // Context to control the crawling process.
	provider     *scanners.FourByteProvider // Provider used to fetch pages.
	redis        *clients.Redis             // BadgerDB instance for storing signatures.
	cooldown     time.Duration              // Cooldown duration between page fetches.
	clickhouseDb *db.ClickHouse
	chainId      *big.Int
}

// WriterOption is a functional option for customizing the FourByteWriter.
type WriterOption func(*FourByteWriter)

func WithProvider(provider *scanners.FourByteProvider) WriterOption {
	return func(c *FourByteWriter) {
		c.provider = provider
	}
}

func WithRedis(client *clients.Redis) WriterOption {
	return func(c *FourByteWriter) {
		c.redis = client
	}
}

func WithCtx(ctx context.Context) WriterOption {
	return func(c *FourByteWriter) {
		c.ctx = ctx
	}
}

func WithCooldown(cooldown time.Duration) WriterOption {
	return func(c *FourByteWriter) {
		c.cooldown = cooldown
	}
}

func WithClickHouseDb(clickhouseDb *db.ClickHouse) WriterOption {
	return func(c *FourByteWriter) {
		c.clickhouseDb = clickhouseDb
	}
}

func WithChainID(chainID *big.Int) WriterOption {
	return func(c *FourByteWriter) {
		c.chainId = chainID
	}
}

func NewFourByteWriter(opts ...WriterOption) *FourByteWriter {
	writer := &FourByteWriter{
		ctx:      context.Background(),
		cooldown: 200 * time.Millisecond,
	}

	for _, opt := range opts {
		opt(writer)
	}
	return writer
}

func (w *FourByteWriter) Crawl() error {
	// Get the last page number from the BadgerDB.
	pageNum, err := w.getLastPageNum()
	if err != nil {
		zap.L().Error(ErrFailedToGetLastPageNumber.Error(), zap.Error(err))
		return err
	}

	for {
		zap.L().Info("Processing page...", zap.Uint64("page_number", pageNum))

		resp, err := w.provider.GetPage(pageNum)
		if err != nil {
			zap.L().Error(ErrFailedToGetPage.Error(), zap.Uint64("page number", pageNum), zap.Error(err))
			return err
		}

		// Process the page content here.
		// If processing is successful, update the last page number in the BadgerDB.
		for _, result := range resp.Results {
			methodName, methodArguments := helpers.ExtractFourByteMethodAndArgumentTypes(result.Text)

			method, err := types.NewFourByteMethod(
				result.Hex,
				methodName,
				result.Text,
				methodArguments,
			)
			if err != nil {
				// Silence invalid method length errors.
				if !strings.Contains(err.Error(), "invalid method length") {
					zap.L().Error(
						ErrFailedToConstructNewMethod.Error(),
						zap.Error(err),
						zap.String("name", result.Text),
					)
				}
				continue
			}

			cacheKey := types.GetMethodStorageKey(w.chainId, method.Bytes)

			exists, err := w.redis.Exists(w.ctx, cacheKey)
			if err != nil {
				zap.L().Error(
					ErrFailedToCheckIfMethodCacheKeyExists.Error(),
					zap.String("method_name", method.Name),
					zap.Error(err),
				)
				continue
			}

			methodBytes, err := method.MarshalBytes()
			if err != nil {
				zap.L().Error(
					ErrFailedMarshalMethod.Error(),
					zap.String("method_name", method.Name),
					zap.Error(err),
				)
				continue
			}

			// Alright, we don't have this signature processed yet, let's do it! :rocket:
			if !exists {
				methodExists, err := models.MethodExists(w.ctx, w.clickhouseDb, method)
				if err != nil {
					zap.L().Error(
						ErrFailedToCheckIfMethodExists.Error(),
						zap.String("method_name", method.Name),
						zap.Error(err),
					)
					continue
				}

				if !methodExists {
					if err := models.InsertMethod(w.ctx, w.clickhouseDb, method); err != nil {
						zap.L().Error(
							ErrFailedToInsertMethod.Error(),
							zap.String("method_name", method.Name),
							zap.Error(err),
						)
						continue
					}
				}

				if err := w.redis.Write(w.ctx, cacheKey, methodBytes, 0); err != nil {
					zap.L().Error(
						ErrFailedRedisWrite.Error(),
						zap.String("method_name", method.Name),
						zap.Error(err),
					)
					continue
				}
			}
		}

		if resp.Next == "" {
			break
		}

		pageNum, err = extractPageNumFromURL(resp.Next)
		if err != nil {
			zap.L().Error(ErrFailedToExtractPageNum.Error(), zap.Error(err))
			return err
		}

		// Update the last page number in the Redis.
		if err = w.setLastPageNum(pageNum); err != nil {
			zap.L().Error(ErrFailedToSetNextPageNumber.Error(), zap.Error(err))
			return err
		}

		// Sleep a bit between each iteration to not overload the API.
		time.Sleep(w.cooldown)
	}

	zap.L().Info("Successfully processed all pages!", zap.Uint64("last_page_number", pageNum))

	return nil
}

// getLastPageNum retrieves the last processed page number from BadgerDB.
//
// The getLastPageNum method retrieves the last processed page number from the BadgerDB instance.
// It opens a transaction to access the value associated with the BDB_NAME_LAST_PROCESSED_PAGE_KEY key.
// If the key is not found, it returns 0 as the last page number.
func (w *FourByteWriter) getLastPageNum() (uint64, error) {
	pageNum := uint64(1)
	exists, err := w.redis.Exists(w.ctx, LAST_PROCESSED_PAGE_KEY)
	if err != nil {
		return 0, err
	}

	if exists {
		val, err := w.redis.Get(w.ctx, LAST_PROCESSED_PAGE_KEY)
		if err != nil {
			return 0, err
		}
		pageNum = binary.BigEndian.Uint64(val)
	}

	return pageNum, nil
}

func (w *FourByteWriter) setLastPageNum(pageNum uint64) error {
	val := make([]byte, 8)
	binary.BigEndian.PutUint64(val, pageNum)
	return w.redis.Write(w.ctx, LAST_PROCESSED_PAGE_KEY, val, 0)
}

// extractPageNumFromURL extracts the page number from a URL.
//
// The extractPageNumFromURL method extracts the page number from the provided URL string.
// It parses the URL and retrieves the value of the "page" query parameter.
// The extracted page number is returned as a uint64.
func extractPageNumFromURL(pageURL string) (uint64, error) {
	u, err := url.Parse(pageURL)
	if err != nil {
		return 0, err
	}

	q := u.Query()
	pageStr := q.Get("page")

	pageNum, err := strconv.ParseUint(pageStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return pageNum, nil
}
