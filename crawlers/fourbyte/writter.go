// Package fourbyte provides a crawler for processing pages and saving signatures to a database.
package fourbyte

import (
	"context"
	"encoding/binary"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/helpers"
	"github.com/txpull/unpack/scanners"
	"github.com/txpull/unpack/signatures"
	"go.uber.org/zap"
)

// BDB_NAME_LAST_PROCESSED_PAGE_KEY is the key for the last processed page number in BadgerDB.
const BDB_NAME_LAST_PROCESSED_PAGE_KEY = "last_processed_fourbyte_page_num"

// PreHook is a function that runs before Crawl.
type PreWriteHook func(signature *signatures.Signature) (*signatures.Signature, error)

// PostHook is a function that runs after Crawl.
type PostWriteHook func(signature *signatures.Signature) error

// FourByteWriter provides a crawler which processes pages and saves signatures to a BadgerDB.
type FourByteWriter struct {
	ctx           context.Context            // Context to control the crawling process.
	provider      *scanners.FourByteProvider // Provider used to fetch pages.
	db            *db.BadgerDB               // BadgerDB instance for storing signatures.
	cooldown      time.Duration              // Cooldown duration between page fetches.
	preWriteHook  PreWriteHook               // Pre-hook function that runs before Crawl.
	postWriteHook PostWriteHook              // Post-hook function that runs after Crawl.
}

// WriterOption is a functional option for customizing the FourByteWriter.
type WriterOption func(*FourByteWriter)

// WithProvider sets the FourByteProvider for the FourByteWriter.
//
// Example:
//
//		provider := scanners.NewFourByteProvider(httpClient)
//	 crawler := NewFourByteWriter(WithProvider(provider))
func WithProvider(provider *scanners.FourByteProvider) WriterOption {
	return func(c *FourByteWriter) {
		c.provider = provider
	}
}

// WithDB sets the BadgerDB for the FourByteWriter.
//
// Example:
//
//	db, _ := db.NewBadgerDB(db.WithContext(ctx), db.WithDbPath("/tmp/mydb"))
//	crawler := NewFourByteWriter(WithDB(db))
func WithDB(db *db.BadgerDB) WriterOption {
	return func(c *FourByteWriter) {
		c.db = db
	}
}

// WithContext sets the context for the FourByteWriter.
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	crawler := NewFourByteWriter(WithContext(ctx))
//	defer cancel()
func WithContext(ctx context.Context) WriterOption {
	return func(c *FourByteWriter) {
		c.ctx = ctx
	}
}

// WithCooldown sets the cooldown duration between crawling iterations.
//
// Example:
//
//	crawler := NewFourByteWriter(WithCooldown(1 * time.Second))
func WithCooldown(cooldown time.Duration) WriterOption {
	return func(c *FourByteWriter) {
		c.cooldown = cooldown
	}
}

// WithPreHook sets the pre-hook function for the FourByteWriter.
//
// Example:
//
//	preHook := func() error {
//		// ... pre-processing tasks ...
//		return nil
//	}
//
//	crawler := NewFourByteWriter(WithPreHook(preHook))
func WithPreHook(hook PreWriteHook) WriterOption {
	return func(c *FourByteWriter) {
		c.preWriteHook = hook
	}
}

// WithPostHook sets the post-hook function for the FourByteWriter.
//
// Example:
//
//	postHook := func() error {
//		// ... post-processing tasks ...
//		return nil
//	}
//
//	crawler := NewFourByteWriter(WithPostHook(postHook))
func WithPostHook(hook PostWriteHook) WriterOption {
	return func(c *FourByteWriter) {
		c.postWriteHook = hook
	}
}

// NewFourByteWriter creates a new FourByteWriter instance with the provided options.
//
// By default, FourByteWriter uses a background context and a cooldown period of 200ms.
// Options can be provided to change these defaults or to set the FourByteProvider and the BadgerDB instance.
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	db, _ := db.NewBadgerDB(db.WithContext(ctx), db.WithDbPath("/tmp/mydb"))
//	provider := scanners.NewFourByteProvider(httpClient)
//
//	writer := NewFourByteWriter(
//		WithContext(ctx),
//		WithDB(db),
//		WithProvider(provider),
//		WithCooldown(1 * time.Second),
//	)
func NewFourByteWriter(opts ...WriterOption) *FourByteWriter {
	writer := &FourByteWriter{
		ctx:           context.Background(),
		provider:      nil,
		db:            nil,
		cooldown:      200 * time.Millisecond,
		preWriteHook:  nil, // No pre-hook by default.
		postWriteHook: nil, // No post-hook by default.
	}

	for _, opt := range opts {
		opt(writer)
	}
	return writer
}

// Crawl starts crawling and processing pages.
//
// It follows these steps:
//
// 1. Retrieves the last processed page number from the BadgerDB.
// 2. Begins fetching and processing each page from the source (using FourByteProvider).
// 3. Extracts signatures from the page content.
// 4. Saves each unique signature to the database.
// 5. Updates the last processed page number in the BadgerDB.
// 6. Sleeps for a cooldown period between iterations.
//
// This method will continue until all pages have been processed.
//
// Example:
//
//	err := crawler.Crawl()
//	if err != nil {
//		log.Fatal("Failed to crawl pages", zap.Error(err))
//	}
func (w *FourByteWriter) Crawl() error {
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
			method, arguments := helpers.ExtractFourByteMethodAndArgumentTypes(result.Text)

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
				// If a pre-hook function has been set, run it.
				if w.preWriteHook != nil {
					var err error
					signature, err = w.preWriteHook(signature)
					if err != nil {
						zap.L().Error("Pre-hook function failed", zap.Error(err))
						return err
					}
				}

				// Save the signature to the database if it does not exist.
				if err := w.saveSignatureIfNotExists(signature); err != nil {
					zap.L().Error("Failed to save signature", zap.Error(err))
					return err
				}

				if w.postWriteHook != nil {
					if err := w.postWriteHook(signature); err != nil {
						zap.L().Error("failed to execute signature post write hook", zap.Error(err))
						return err
					}
				}

				os.Exit(1)
				return nil
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
func (w *FourByteWriter) getLastPageNum() (uint64, error) {
	pageNum := uint64(1)
	exists, err := w.db.Exists(BDB_NAME_LAST_PROCESSED_PAGE_KEY)
	if err != nil {
		return 0, err
	}

	if exists {
		val, err := w.db.Get(BDB_NAME_LAST_PROCESSED_PAGE_KEY)
		if err != nil {
			return 0, err
		}
		pageNum = binary.BigEndian.Uint64(val)
	}

	return pageNum, nil
}

// setLastPageNum sets the last processed page number in BadgerDB.
//
// The setLastPageNum method sets the last processed page number in the BadgerDB instance.
// It opens a transaction and stores the provided page number as a byte slice using the BDB_NAME_LAST_PROCESSED_PAGE_KEY key.
func (w *FourByteWriter) setLastPageNum(pageNum uint64) error {
	val := make([]byte, 8)
	binary.BigEndian.PutUint64(val, pageNum)
	return w.db.Write(BDB_NAME_LAST_PROCESSED_PAGE_KEY, val)
}

// saveSignatureIfNotExists saves the signature to the database if it doesn't already exist.
//
// The saveSignatureIfNotExists method checks if the signature already exists in the database.
// If it doesn't exist, it saves the signature to the database using the saveSignature method.
func (w *FourByteWriter) saveSignatureIfNotExists(signature *signatures.Signature) error {
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
func (w *FourByteWriter) signatureExists(signature *signatures.Signature) (bool, error) {
	exists, err := w.db.Exists(signature.Hex)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// saveSignature saves the signature to the database.
//
// The saveSignature method saves the given signature to the database.
// It opens a transaction, marshals the signature to JSON bytes, and stores the byte slice using the signature's hex string as the key.
// saveSignature saves the signature to the database.
func (w *FourByteWriter) saveSignature(signature *signatures.Signature) error {
	signatureBytes, err := signature.MarshalBytes()
	if err != nil {
		return err
	}

	// Save the signature to BadgerDB with the signature ID as the key.
	return w.db.Write(signature.Hex, signatureBytes)
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
