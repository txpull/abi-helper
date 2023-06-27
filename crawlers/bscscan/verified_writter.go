// Package bscscan provides utilities to interact with BSCScan for contract
// information and to persist that data in Redis.
//
// It is capable of reading contract data from CSV files, extracting
// contract details from BSCScan, and writing them to Redis.
//
// Usage:
//
//	ctx := context.Background()
//	db := db.NewRedis("path/to/db")
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
//	writer.WithRedis(db)
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
// data in the Redis for future use.
package bscscan

import (
	"context"
	"encoding/csv"
	"io"
	"math"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/txpull/unpack/clients"
	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/db/models"
	"github.com/txpull/unpack/helpers"
	"github.com/txpull/unpack/scanners"
	"github.com/txpull/unpack/types"
	"go.uber.org/zap"
)

const (
	defaultRequestLimit    = 10                     // Default maximum number of requests allowed per second
	defaultRequestInterval = 100 * time.Millisecond // Default time interval between each request
	defaultMaxRetry        = 5                      // Default maximum number of retries
	defaultBackoffFactor   = 2                      // Default backoff factor for exponential backoff
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
// and store their information in Redis.
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
	redis           *clients.Redis
	clickhouseDb    *db.ClickHouse
	ethClient       *clients.EthClient
	chainId         *big.Int
}

// WithScanner sets the BscScanProvider scanner for the BscscanWriter.
func WithScanner(scanner *scanners.BscScanProvider) Option {
	return func(bs *BscscanWriter) {
		bs.scanner = scanner
	}
}

// WithDataPath sets the path to the CSV file containing contract data for the BscscanWriter.
func WithDataPath(dataPath string) Option {
	return func(bs *BscscanWriter) {
		bs.dataPath = dataPath
	}
}

// WithRequestLimit sets the maximum number of requests allowed per second for the BscscanWriter.
func WithRequestLimit(requestLimit int) Option {
	return func(bs *BscscanWriter) {
		bs.requestLimit = requestLimit
	}
}

// WithRequestInterval sets the time interval between each request for the BscscanWriter.
func WithRequestInterval(requestInterval time.Duration) Option {
	return func(bs *BscscanWriter) {
		bs.requestInterval = requestInterval
	}
}

// WithMaxRetry sets the maximum number of retries for the BscscanWriter.
func WithMaxRetry(maxRetry int) Option {
	return func(bs *BscscanWriter) {
		bs.maxRetry = maxRetry
	}
}

// WithBackoffFactor sets the backoff factor for exponential backoff for the BscscanWriter.
func WithBackoffFactor(backoffFactor float64) Option {
	return func(bs *BscscanWriter) {
		bs.backoffFactor = backoffFactor
	}
}

func WithRedis(client *clients.Redis) Option {
	return func(c *BscscanWriter) {
		c.redis = client
	}
}

func WithClickHouseDb(clickhouseDb *db.ClickHouse) Option {
	return func(bs *BscscanWriter) {
		bs.clickhouseDb = clickhouseDb
	}
}

func WithEthClient(client *clients.EthClient) Option {
	return func(bs *BscscanWriter) {
		bs.ethClient = client
	}
}

func WithChainID(chainID *big.Int) Option {
	return func(bs *BscscanWriter) {
		bs.chainId = chainID
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

func (bs *BscscanWriter) tryScanContract(ctx context.Context, c CsvContract) error {
	for i := 0; i < bs.maxRetry; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			key := types.GetContractStorageKey(bs.chainId, c.ContractAddress)

			exists, err := bs.redis.Exists(bs.ctx, key)
			if err != nil {
				zap.L().Error(
					ErrFailedCheckExistenceInBadger.Error(),
					zap.String("contract_address", c.ContractAddress.Hex()),
					zap.Error(err),
				)
				return err
			}

			if bs.clickhouseDb != nil {
				dbExists, err := models.ContractExists(bs.ctx, bs.clickhouseDb, c.ContractAddress)
				if err != nil {
					zap.L().Error(
						ErrFailedToCheckIfMethodCacheKeyExists.Error(),
						zap.String("contract_address", c.ContractAddress.Hex()),
						zap.Error(err),
					)
					return err
				}

				if dbExists {
					continue
				}
			}

			// Just skip if the contract already exists in Redis
			if exists {
				zap.L().Info(
					"Contract already exists in Redis",
					zap.String("contract_address", c.ContractAddress.Hex()),
				)
				return nil
			}

			tx, _, err := helpers.GetTransactionByHash(bs.ctx, bs.ethClient, c.Txhash)
			if err != nil {
				zap.L().Error(
					ErrFailedGetTransactionByHash.Error(),
					zap.String("contract_address", c.ContractAddress.Hex()),
					zap.String("tx_hash", c.Txhash.Hex()),
					zap.Error(err),
				)
				time.Sleep(time.Duration(int64(math.Pow(float64(bs.backoffFactor), float64(i)))) * time.Second)
				continue
			}

			txReceipt, err := helpers.GetReceiptByHash(bs.ctx, bs.ethClient, c.Txhash)
			if err != nil {
				zap.L().Error(
					ErrFailedGetTransactionReceiptByHash.Error(),
					zap.String("contract_address", c.ContractAddress.Hex()),
					zap.String("tx_hash", c.Txhash.Hex()),
					zap.Error(err),
				)
				time.Sleep(time.Duration(int64(math.Pow(float64(bs.backoffFactor), float64(i)))) * time.Second)
				continue
			}

			contractResult, err := bs.scanner.ScanContract(c.ContractAddress.Hex())

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

			contract := &types.Contract{
				UUID:                 uuid.New(),
				BlockHash:            txReceipt.BlockHash,
				TransactionHash:      tx.Hash(),
				Name:                 contractResult.Name,
				ChainID:              bs.chainId,
				Address:              c.ContractAddress,
				CompilerVersion:      contractResult.CompilerVersion,
				ConstructorArguments: contractResult.ConstructorArguments,
				OptimizationUsed: func() string {
					if contractResult.OptimizationUsed == "1" {
						return "true"
					}
					return "false"
				}(),
				Language: func() types.ContractLanguage {
					if strings.Contains(contractResult.CompilerVersion, "vyper") {
						return types.ContractLanguageVyper
					}

					return types.ContractLanguageSolidity
				}(),
				Runs:        contractResult.Runs,
				EVMVersion:  strings.ToLower(contractResult.EVMVersion),
				Library:     contractResult.Library,
				LicenseType: contractResult.LicenseType,
				Proxy: func() string {
					if contractResult.Proxy == "1" {
						return "true"
					}
					return "false"
				}(),
				SourceCode:         contractResult.SourceCode,
				ABI:                contractResult.ABI,
				VerificationType:   types.ContractVerificationTypeBscscan,
				VerificationStatus: "perfect", // This is what sourcify returns so we'll just use it
			}

			if len(contractResult.SwarmSource) > 0 {
				contract.SourceUrls = append(contract.SourceUrls, contractResult.SwarmSource)
			}

			resBytes, err := contract.MarshalBytes()
			if err != nil {
				zap.L().Error(
					ErrFailedMarshalContractInfo.Error(),
					zap.String("contract_address", c.ContractAddress.Hex()),
					zap.Error(err),
				)
				return err
			}

			// Insert contract into the clickhouse database but only if clickhouse database is set
			if bs.clickhouseDb != nil {
				if err := models.InsertContract(bs.ctx, bs.clickhouseDb, contract); err != nil {
					return err
				}
			}

			// Process abi methods, events, resulting contract mappings and write them into the database for future use
			if err := bs.processAbi(ctx, contract); err != nil {
				zap.L().Error(
					ErrFailedProcessAbi.Error(),
					zap.String("contract_address", c.ContractAddress.Hex()),
					zap.Error(err),
				)
				return err
			}

			err = bs.redis.Write(bs.ctx, key, resBytes, 0)
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

func (bs *BscscanWriter) processAbi(ctx context.Context, contractResult *types.Contract) error {
	abi, err := abi.JSON(strings.NewReader(contractResult.ABI))
	if err != nil {
		zap.L().Error(
			ErrFailedParseAbi.Error(),
			zap.String("contract_address", contractResult.Address.Hex()),
			zap.Error(err),
		)
		return err
	}

	// Process abi methods and events and write them into the database for future use
	if err := bs.processAbiMethods(contractResult, abi.Methods); err != nil {
		zap.L().Error(
			ErrFailedProcessAbiMethods.Error(),
			zap.String("contract_address", contractResult.Address.Hex()),
			zap.Error(err),
		)
		return err
	}

	if err := bs.processAbiEvents(ctx, contractResult, abi.Events); err != nil {
		zap.L().Error(
			ErrFailedProcessAbiEvents.Error(),
			zap.String("contract_address", contractResult.Address.Hex()),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (bs *BscscanWriter) processAbiMethods(contractResult *types.Contract, methods map[string]abi.Method) error {
	for _, method := range methods {
		select {
		case <-bs.ctx.Done():
			return bs.ctx.Err()
		default:
			methodKey := types.GetMethodStorageKey(bs.chainId, method.ID)
			methodMapperKey := types.GetMethodMapperStorageKey(bs.chainId, method.ID)

			exists, err := bs.redis.Exists(bs.ctx, methodKey)
			if err != nil {
				zap.L().Error(
					ErrFailedCheckExistenceInBadger.Error(),
					zap.String("contract_address", contractResult.Address.Hex()),
					zap.String("method_name", method.Name),
					zap.Error(err),
				)
				return err
			}

			// Just skip if the method already exists in Redis
			if exists {
				continue
			}

			methodResult := types.NewFullMethod(method)
			methodMappingResult := types.NewMethodMapping(contractResult, methodResult)

			// Write into clickhouse but only if clickhouse database is set
			if bs.clickhouseDb != nil {
				methodExists, err := models.MethodExists(bs.ctx, bs.clickhouseDb, methodResult)
				if err != nil {
					zap.L().Error(
						ErrFailedToCheckIfMethodExists.Error(),
						zap.String("method_name", method.Name),
						zap.Error(err),
					)
					continue
				}

				if !methodExists {
					if err := models.InsertMethod(bs.ctx, bs.clickhouseDb, methodResult); err != nil {
						zap.L().Error(
							ErrFailedToInsertMethod.Error(),
							zap.Error(err),
							zap.String("contract_address", contractResult.Address.Hex()),
							zap.String("method_name", method.Name),
						)
						return err
					}

					if err := models.InsertMethodMapping(bs.ctx, bs.clickhouseDb, methodMappingResult); err != nil {
						zap.L().Error(
							ErrFailedToInsertMethodMapping.Error(),
							zap.Error(err),
							zap.String("contract_address", contractResult.Address.Hex()),
							zap.String("method_name", method.Name),
						)
						return err
					}
				}
			}

			methodBytes, err := methodResult.MarshalBytes()
			if err != nil {
				zap.L().Error(
					ErrFailedMarshalMethodInfo.Error(),
					zap.String("contract_address", contractResult.Address.Hex()),
					zap.String("method_name", method.Name),
					zap.Error(err),
				)
				return err
			}

			methodMappingBytes, err := methodMappingResult.MarshalBytes()
			if err != nil {
				zap.L().Error(
					ErrFailedMarshalMethodMappingInfo.Error(),
					zap.String("contract_address", contractResult.Address.Hex()),
					zap.String("method_name", method.Name),
					zap.Error(err),
				)
				return err
			}

			if err := bs.redis.Write(bs.ctx, methodKey, methodBytes, 0); err != nil {
				zap.L().Error(
					ErrFailedWriteMethodInfoToDB.Error(),
					zap.String("contract_address", contractResult.Address.Hex()),
					zap.String("method_name", method.Name),
					zap.Error(err),
				)
				return err
			}

			if err := bs.redis.Write(bs.ctx, methodMapperKey, methodMappingBytes, 0); err != nil {
				zap.L().Error(
					ErrFailedWriteMethodMappingInfoToDB.Error(),
					zap.String("contract_address", contractResult.Address.Hex()),
					zap.String("method_name", method.Name),
					zap.Error(err),
				)
				return err
			}
		}
	}

	return nil
}

func (bs *BscscanWriter) processAbiEvents(ctx context.Context, contractResult *types.Contract, events map[string]abi.Event) error {
	for _, event := range events {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			eventKey := types.GetEventStorageKey(bs.chainId, event.ID)
			eventMappingKey := types.GetEventMapperStorageKey(bs.chainId, event.ID)

			exists, err := bs.redis.Exists(bs.ctx, eventKey)
			if err != nil {
				zap.L().Error(
					ErrFailedCheckEventExistenceInBadger.Error(),
					zap.String("contract_address", contractResult.Address.Hex()),
					zap.String("event_name", event.Name),
					zap.Error(err),
				)
				return err
			}

			// Just skip if the event already exists in Redis
			if exists {
				continue
			}

			if bs.clickhouseDb != nil {
				dbExists, err := models.EventExist(bs.ctx, bs.clickhouseDb, event.ID)
				if err != nil {
					zap.L().Error(
						ErrFailedToCheckIfMethodCacheKeyExists.Error(),
						zap.String("contract_address", contractResult.Address.Hex()),
						zap.String("event_hash", event.ID.Hex()),
						zap.Error(err),
					)
					return err
				}

				if dbExists {
					continue
				}
			}

			// This is the deal.
			// We need to create a new event and event mapping and write it into the database.
			// Now event itself is 1-to-1 mapping with contract but event mapping will be 1-to-many.

			eventResult := types.NewFullEvent(event)

			// Write into clickhouse but only if clickhouse database is set
			if bs.clickhouseDb != nil {
				if err := models.InsertEvent(bs.ctx, bs.clickhouseDb, eventResult); err != nil {
					zap.L().Error(
						ErrFailedToInsertEvent.Error(),
						zap.Error(err),
						zap.String("contract_address", contractResult.Address.Hex()),
						zap.String("method_name", eventResult.Name),
					)
					return err
				}
			}

			eventMappingResult := types.NewEventMapping(contractResult, eventResult)
			if bs.clickhouseDb != nil {
				if err := models.InsertEventMapping(bs.ctx, bs.clickhouseDb, eventMappingResult); err != nil {
					zap.L().Error(
						ErrFailedToInsertEventMapping.Error(),
						zap.Error(err),
						zap.String("contract_address", contractResult.Address.Hex()),
						zap.String("method_name", eventResult.Name),
					)
					return err
				}
			}

			eventBytes, err := eventResult.MarshalBytes()
			if err != nil {
				zap.L().Error(
					ErrFailedToMarshalEventInfo.Error(),
					zap.String("contract_address", contractResult.Address.Hex()),
					zap.String("event_name", event.Name),
					zap.Error(err),
				)
				return err
			}

			if err := bs.redis.Write(bs.ctx, eventKey, eventBytes, 0); err != nil {
				zap.L().Error(
					ErrFailedToWriteEventInfoToRedis.Error(),
					zap.String("contract_address", contractResult.Address.Hex()),
					zap.String("event_name", event.Name),
					zap.Error(err),
				)
				return err
			}

			eventMapperBytes, err := eventMappingResult.MarshalBytes()
			if err != nil {
				zap.L().Error(
					ErrFailedToMarshalEventInfo.Error(),
					zap.String("contract_address", contractResult.Address.Hex()),
					zap.String("event_name", event.Name),
					zap.Error(err),
				)
				return err
			}

			if err := bs.redis.Write(bs.ctx, eventMappingKey, eventMapperBytes, 0); err != nil {
				zap.L().Error(
					ErrFailedToWriteEventMappingInfoToRedis.Error(),
					zap.String("contract_address", contractResult.Address.Hex()),
					zap.String("event_name", event.Name),
					zap.Error(err),
				)
				return err
			}
		}

	}

	return nil
}
