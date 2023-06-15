// Package fourbyte provides a crawler for processing pages and saving signatures to a database.
package fourbyte

import (
	"context"
	"encoding/binary"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/txpull/bytecode/scanners"
	"github.com/txpull/bytecode/signatures"
	"github.com/txpull/bytecode/utils"
	"go.uber.org/zap"
)

// BDB_NAME_LAST_PROCESSED_PAGE_KEY is the key for the last processed page number in BadgerDB.
const BDB_NAME_LAST_PROCESSED_PAGE_KEY = "last_processed_fourbyte_page_num"

// FourByteWritter crawls and processes pages, saving signatures to the database.
type FourByteWritter struct {
	ctx      context.Context
	provider *scanners.FourByteProvider
	db       *badger.DB
	cooldown time.Duration
}

// CrawlerOption is a functional option to customize the FourByteWritter.
type CrawlerOption func(*FourByteWritter)

// WithProvider sets the FourByteProvider for the FourByteWritter.
func WithProvider(provider *scanners.FourByteProvider) CrawlerOption {
	return func(c *FourByteWritter) {
		c.provider = provider
	}
}

// WithDB sets the BadgerDB for the FourByteWritter.
func WithDB(db *badger.DB) CrawlerOption {
	return func(c *FourByteWritter) {
		c.db = db
	}
}

// WithContext sets the context for the FourByteWritter.
func WithContext(ctx context.Context) CrawlerOption {
	return func(c *FourByteWritter) {
		c.ctx = ctx
	}
}

// WithCooldown sets the cooldown duration between crawling iterations.
func WithCooldown(cooldown time.Duration) CrawlerOption {
	return func(c *FourByteWritter) {
		c.cooldown = cooldown
	}
}

// NewFourByteWritter creates a new FourByteWritter instance with the provided options.
func NewFourByteWritter(opts ...CrawlerOption) *FourByteWritter {
	crawler := &FourByteWritter{
		ctx:      context.Background(),
		provider: nil,
		db:       nil,
		cooldown: 200 * time.Millisecond,
	}

	for _, opt := range opts {
		opt(crawler)
	}

	return crawler
}

// Crawl starts crawling and processing pages.
//
// The Crawl method initiates the crawling process, retrieving and processing pages until there are no more pages to process. It utilizes the FourByteProvider to fetch pages and saves the extracted signatures to the specified database using the BadgerDB instance.
// The crawling process includes the following steps:
// - Get the last processed page number from the BadgerDB.
// - Process each page:
//   - Get the page using the FourByteProvider.
//   - Extract signatures from the page content.
//   - Save each signature to the database if it doesn't already exist.
//   - Update the last page number in the BadgerDB.
//   - Sleep for a cooldown period between iterations.
//
// - Upon completion, the method logs the last processed page number and returns any encountered errors.
func (w *FourByteWritter) Crawl() error {
	// Get the last page number from the BadgerDB.
	pageNum, err := w.getLastPageNum()
	if err != nil {
		zap.L().Error("Failed to get last page number from DB", zap.Error(err))
		return err
	}

	for {
		zap.L().Info("Processing page...", zap.Uint64("page_number", pageNum))

		resp, err := w.provider.GetPage(pageNum)
		if err != nil {
			zap.L().Error("Failed to get page", zap.Uint64("page number", pageNum), zap.Error(err))
			return err
		}

		// Process the page content here.
		// If processing is successful, update the last page number in the BadgerDB.
		for _, result := range resp.Results {
			method, arguments := utils.ExtractFourByteMethodAndArgumentTypes(result.Text)

			signature := signatures.NewSignature(
				uint64(result.ID),
				result.Text,
				strings.TrimLeft(result.Hex, "0x"),
				signatures.ParseSignatureInputFromArray(arguments),
				method,
				[]signatures.InputOutput{},
				false,
				"not_known",
				result.CreatedAt,
			)

			if len(signature.Hex) > 0 {
				// Save the signature to the database if it does not exist.
				if err := w.saveSignatureIfNotExists(signature); err != nil {
					zap.L().Error("Failed to save signature", zap.Error(err))
					return err
				}
			}
		}

		if resp.Next == "" {
			break
		}

		pageNum, err = extractPageNumFromURL(resp.Next)
		if err != nil {
			zap.L().Error("Failed to extract page number from URL", zap.Error(err))
			return err
		}

		// Update the last page number in the BadgerDB.
		if err = w.setLastPageNum(pageNum); err != nil {
			zap.L().Error("Failed to set last page number in DB", zap.Error(err))
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
func (w *FourByteWritter) getLastPageNum() (uint64, error) {
	pageNum := uint64(1)
	err := w.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(BDB_NAME_LAST_PROCESSED_PAGE_KEY))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil // Key not found is not an error, it means we start from the first page.
			}
			return err
		}

		valCopy, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		pageNum = binary.BigEndian.Uint64(valCopy)
		return nil
	})

	return pageNum, err
}

// setLastPageNum sets the last processed page number in BadgerDB.
//
// The setLastPageNum method sets the last processed page number in the BadgerDB instance.
// It opens a transaction and stores the provided page number as a byte slice using the BDB_NAME_LAST_PROCESSED_PAGE_KEY key.
func (w *FourByteWritter) setLastPageNum(pageNum uint64) error {
	return w.db.Update(func(txn *badger.Txn) error {
		val := make([]byte, 8)
		binary.BigEndian.PutUint64(val, pageNum)
		return txn.Set([]byte(BDB_NAME_LAST_PROCESSED_PAGE_KEY), val)
	})
}

// saveSignatureIfNotExists saves the signature to the database if it doesn't already exist.
//
// The saveSignatureIfNotExists method checks if the signature already exists in the database.
// If it doesn't exist, it saves the signature to the database using the saveSignature method.
func (w *FourByteWritter) saveSignatureIfNotExists(signature *signatures.Signature) error {
	// Check if the signature already exists in the database.
	exists, err := w.signatureExists(signature)
	if err != nil {
		return err
	}

	if !exists {
		// Save the signature to the database.
		if err := w.saveSignature(signature); err != nil {
			return err
		}
	}

	return nil
}

// signatureExists checks if the signature exists in the database.
//
// The signatureExists method checks if the given signature exists in the database.
// It opens a transaction and attempts to retrieve the value associated with the signature's hex string.
// If the key is not found, it returns false, indicating that the signature does not exist.
func (w *FourByteWritter) signatureExists(signature *signatures.Signature) (bool, error) {
	exists := false

	err := w.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(signature.Hex))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil // Key not found means the signature does not exist.
			}
			return err
		}

		// Key found, signature exists.
		exists = true
		return nil
	})

	if err != nil {
		return false, err
	}

	return exists, nil
}

// saveSignature saves the signature to the database.
//
// The saveSignature method saves the given signature to the database.
// It opens a transaction, marshals the signature to JSON bytes, and stores the byte slice using the signature's hex string as the key.
func (w *FourByteWritter) saveSignature(signature *signatures.Signature) error {
	err := w.db.Update(func(txn *badger.Txn) error {
		// Marshal the signature to JSON bytes.
		signatureBytes, err := signature.MarshalBytes()
		if err != nil {
			return err
		}

		// Save the signature to BadgerDB with the signature ID as the key.
		err = txn.Set([]byte(signature.Hex), signatureBytes)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
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
