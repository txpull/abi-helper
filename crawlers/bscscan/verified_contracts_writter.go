// Package bscscan provides utilities to interact with BSCScan for contract
// information and to persist that data in BadgerDB.
//
// It is capable of reading contract data from CSV files, extracting
// contract details from BSCScan, and writing them to BadgerDB.
//
// Usage:
//
//	ctx := context.Background()
//	db := db.NewBadgerDB("path/to/db")
//
//	writer := bscscan.New(
//	    ctx,
//	    bscscan.WithDataPath("path/to/csv"),
//	    bscscan.WithScanner(&scanners.BscScanProvider{}),
//	    bscscan.WithRequestLimit(6),
//	    bscscan.WithRequestInterval(100*time.Millisecond),
//	    bscscan.WithMaxRetry(5),
//	    bscscan.WithBackoffFactor(2),
//	)
//
//	writer.WithBadgerDb(db)
//
//	contracts, err := writer.GatherVerifiedContracts()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if err := writer.ProcessVerifiedContracts(ctx, contracts); err != nil {
//	    log.Fatal(err)
//	}
//
// This will read verified contract details from BSCScan and persist the
// data in the BadgerDB for future use.
package bscscan

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/scanners"
	"go.uber.org/zap"
)

var (
	// ErrNoContractsToProcess is returned when there are no contracts to process.
	ErrNoContractsToProcess = errors.New("no contracts to process")

	// ErrMaxRateLimitReached is returned when the maximum rate limit is reached and retry is needed.
	ErrMaxRateLimitReached = errors.New("max rate limit reached, retrying...")

	// ErrFailedGetContractInfo is returned when failed to get contract information from BSCScan.
	ErrFailedGetContractInfo = errors.New("failed to get contract information from BSCScan")

	// ErrFailedMarshalContractInfo is returned when failed to marshal contract information to binary.
	ErrFailedMarshalContractInfo = errors.New("failed to marshal contract information to binary")

	// ErrFailedCheckExistenceInBadger is returned when failed to check existence of contract information in BadgerDB.
	ErrFailedCheckExistenceInBadger = errors.New("failed to check existence of contract information in BadgerDB")

	// ErrFailedWriteContractInfoToDB is returned when failed to write contract information to BadgerDB.
	ErrFailedWriteContractInfoToDB = errors.New("failed to write contract information to BadgerDB")

	// ErrExceededMaxRetryAttempts is returned when the maximum number of retry attempts is exceeded.
	ErrExceededMaxRetryAttempts = errors.New("exceeded max retry attempts")
)

const (
	databaseKeyPrefix      = "bscscan_verified_contracts:" // your prefix for keys
	defaultRequestLimit    = 6                             // Default maximum number of requests allowed per second
	defaultRequestInterval = 100 * time.Millisecond        // Default time interval between each request
	defaultMaxRetry        = 5                             // Default maximum number of retries
	defaultBackoffFactor   = 2                             // Default backoff factor for exponential backoff
)

// Option represents a functional option for configuring the BscscanWriter.
type Option func(*BscscanWriter)

// CsvContract holds a transaction hash, a contract address, and a contract name.
type CsvContract struct {
	Txhash          common.Hash
	ContractAddress common.Address
	ContractName    string
}

// BscscanWriter provides methods to gather verified contracts, process them,
// and store their information in BadgerDB.
type BscscanWriter struct {
	ctx             context.Context
	scanner         *scanners.BscScanProvider
	dataPath        string
	requestLimit    int
	requestInterval time.Duration
	maxRetry        int
	backoffFactor   float64
	semaphore       chan struct{}
	wg              sync.WaitGroup
	badgerDb        *db.BadgerDB
}

// WithScanner sets the BscScanProvider scanner for the BscscanWriter.
func WithScanner(scanner *scanners.BscScanProvider) Option {
	return func(w *BscscanWriter) {
		w.scanner = scanner
	}
}

// WithDataPath sets the path to the CSV file containing contract data for the BscscanWriter.
func WithDataPath(dataPath string) Option {
	return func(w *BscscanWriter) {
		w.dataPath = dataPath
	}
}

// WithRequestLimit sets the maximum number of requests allowed per second for the BscscanWriter.
func WithRequestLimit(requestLimit int) Option {
	return func(w *BscscanWriter) {
		w.requestLimit = requestLimit
	}
}

// WithRequestInterval sets the time interval between each request for the BscscanWriter.
func WithRequestInterval(requestInterval time.Duration) Option {
	return func(w *BscscanWriter) {
		w.requestInterval = requestInterval
	}
}

// WithMaxRetry sets the maximum number of retries for the BscscanWriter.
func WithMaxRetry(maxRetry int) Option {
	return func(w *BscscanWriter) {
		w.maxRetry = maxRetry
	}
}

// WithBackoffFactor sets the backoff factor for exponential backoff for the BscscanWriter.
func WithBackoffFactor(backoffFactor float64) Option {
	return func(w *BscscanWriter) {
		w.backoffFactor = backoffFactor
	}
}

// WithBadgerDb
func WithBadgerDb(badgerDb *db.BadgerDB) Option {
	return func(w *BscscanWriter) {
		w.badgerDb = badgerDb
	}
}

// New creates a new BscscanWriter with the provided options.
func NewVerifiedContractsWritter(ctx context.Context, opts ...Option) *BscscanWriter {
	writer := &BscscanWriter{
		ctx:             ctx,
		requestLimit:    defaultRequestLimit,
		requestInterval: defaultRequestInterval,
		maxRetry:        defaultMaxRetry,
		backoffFactor:   defaultBackoffFactor,
	}

	for _, opt := range opts {
		opt(writer)
	}

	writer.semaphore = make(chan struct{}, writer.requestLimit)

	return writer
}

// GatherVerifiedContracts reads contract data from a CSV file.
// It skips any records that do not contain valid contract data.
//
// Example:
//
//	contracts, err := bs.GatherVerifiedContracts()
//	if err != nil {
//	    return nil, err
//	}
func (bs *BscscanWriter) GatherVerifiedContracts() ([]CsvContract, error) {
	file, err := os.Open(bs.dataPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	var contracts []CsvContract

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		if strings.Contains(record[0], "Note:") || strings.Contains(record[0], "Txhash") {
			continue
		}

		contract := CsvContract{
			Txhash:          common.HexToHash(record[0]),
			ContractAddress: common.HexToAddress(record[1]),
			ContractName:    record[2],
		}

		contracts = append(contracts, contract)
	}

	return contracts, nil
}

// ProcessVerifiedContracts processes the provided contracts by scanning their contract information
// and storing it in BadgerDB. It uses concurrency to process multiple contracts simultaneously.
//
// Example:
//
//	if err := writer.ProcessVerifiedContracts(contracts); err != nil {
//	    return err
//	}
func (bs *BscscanWriter) ProcessVerifiedContracts(contracts []CsvContract) error {
	if len(contracts) < 1 {
		return ErrNoContractsToProcess
	}

	bs.semaphore = make(chan struct{}, bs.requestLimit)

	for _, contract := range contracts {
		select {
		case <-bs.ctx.Done():
			return bs.ctx.Err()
		default:
			bs.wg.Add(1)
			bs.semaphore <- struct{}{}

			go func(c CsvContract) {
				defer bs.wg.Done()

				zap.L().Info("Processing contract", zap.String("contract_address", c.ContractAddress.Hex()))

				err := bs.tryScanContract(bs.ctx, c)

				if err != nil {
					zap.L().Error(
						"Failed to scan contract",
						zap.String("contract_address", c.ContractAddress.Hex()),
						zap.Error(err),
					)
				}

				<-bs.semaphore
			}(contract)

			time.Sleep(bs.requestInterval)
		}
	}

	// Wait for all contracts to finish processing
	bs.wg.Wait()

	return nil
}

// tryScanContract tries to scan the contract with retry mechanism and store it in BadgerDB.
func (bs *BscscanWriter) tryScanContract(ctx context.Context, c CsvContract) error {
	for i := 0; i < bs.maxRetry; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Store the contract information in BadgerDB under key = databaseKeyPrefix + ContractAddress
			// only if it does not exist already
			key := databaseKeyPrefix + c.ContractAddress.Hex()

			exists, err := bs.badgerDb.Exists(key)
			if err != nil {
				zap.L().Error(
					ErrFailedCheckExistenceInBadger.Error(),
					zap.String("contract_address", c.ContractAddress.Hex()),
					zap.Error(err),
				)
				return err
			}

			// Just skip if the contract already exists in BadgerDB
			if exists {
				zap.L().Info(
					"Contract already exists in BadgerDB",
					zap.String("contract_address", c.ContractAddress.Hex()),
				)
				return nil
			}

			contractInfo, err := bs.scanner.ScanContract(c.ContractAddress.Hex())

			if err != nil {
				if strings.Contains(err.Error(), ErrMaxRateLimitReached.Error()) {
					zap.L().Error(
						ErrMaxRateLimitReached.Error(),
						zap.String("contract_address", c.ContractAddress.Hex()),
						zap.Int("retry_attempt", i+1),
						zap.Error(err),
					)
					time.Sleep(time.Duration(int64(math.Pow(float64(bs.backoffFactor), float64(i)))) * time.Second)
					continue
				} else {
					zap.L().Error(
						ErrFailedGetContractInfo.Error(),
						zap.String("contract_address", c.ContractAddress.Hex()),
						zap.Error(err),
					)
					return err
				}
			}

			resBytes, err := contractInfo.MarshalBytes()
			if err != nil {
				zap.L().Error(
					ErrFailedMarshalContractInfo.Error(),
					zap.String("contract_address", c.ContractAddress.Hex()),
					zap.Error(err),
				)
				return err
			}

			err = bs.badgerDb.Write(key, resBytes)
			if err != nil {
				zap.L().Error(
					ErrFailedWriteContractInfoToDB.Error(),
					zap.String("contract_address", c.ContractAddress.Hex()),
					zap.Error(err),
				)
				return err
			}

			return nil
		}
	}

	return ErrExceededMaxRetryAttempts
}
