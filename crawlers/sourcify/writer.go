package sourcify

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/txpull/sourcify-go"
	"github.com/txpull/unpack/clients"
	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/db/models"
	"github.com/txpull/unpack/helpers"
	"github.com/txpull/unpack/scanners"
	"github.com/txpull/unpack/types"
	"go.uber.org/zap"
)

var (
	queryChainToName = map[int64]string{
		1:  "ethereum",
		56: "bsc",
	}
)

type SourcifyWriter struct {
	ctx          context.Context  // Context to control the crawling process.
	provider     *sourcify.Client // Provider used to fetch pages.
	redis        *clients.Redis   // BadgerDB instance for storing signatures.
	clickhouseDb *db.ClickHouse
	bitquery     *scanners.BitQueryProvider
	ethClient    *clients.EthClient
	bscscan      *scanners.BscScanProvider
	chainId      *big.Int
}

// WriterOption is a functional option for customizing the SourcifyWriter.
type WriterOption func(*SourcifyWriter)

func WithSourcify(provider *sourcify.Client) WriterOption {
	return func(c *SourcifyWriter) {
		c.provider = provider
	}
}

func WithRedis(client *clients.Redis) WriterOption {
	return func(c *SourcifyWriter) {
		c.redis = client
	}
}

// WithBscScan sets the BscScanProvider scanner for the SourcifyWriter.
func WithBscScan(scanner *scanners.BscScanProvider) WriterOption {
	return func(w *SourcifyWriter) {
		w.bscscan = scanner
	}
}

func WithCtx(ctx context.Context) WriterOption {
	return func(c *SourcifyWriter) {
		c.ctx = ctx
	}
}

func WithClickHouseDb(clickhouseDb *db.ClickHouse) WriterOption {
	return func(c *SourcifyWriter) {
		c.clickhouseDb = clickhouseDb
	}
}

func WithChainID(chainID *big.Int) WriterOption {
	return func(c *SourcifyWriter) {
		c.chainId = chainID
	}
}

func WithBitQuery(bq *scanners.BitQueryProvider) WriterOption {
	return func(c *SourcifyWriter) {
		c.bitquery = bq
	}
}

func WithEthClient(client *clients.EthClient) WriterOption {
	return func(w *SourcifyWriter) {
		w.ethClient = client
	}
}

func NewSourcifyWriter(opts ...WriterOption) *SourcifyWriter {
	writer := &SourcifyWriter{
		ctx: context.Background(),
	}
	for _, opt := range opts {
		opt(writer)
	}
	return writer
}

func (w *SourcifyWriter) GetContractListByChainID(chainID *big.Int) (*sourcify.VerifiedContractAddresses, error) {
	return sourcify.GetAvailableContractAddresses(w.provider, int(chainID.Int64()))
}

func (w *SourcifyWriter) ProcessContractsByType(chainID *big.Int, contracts *sourcify.VerifiedContractAddresses, contractType sourcify.MethodMatchType) error {
	var slice []common.Address

	switch contractType {
	case sourcify.MethodMatchTypeFull:
		slice = contracts.Full
	case sourcify.MethodMatchTypePartial:
		slice = contracts.Partial
	}

	for _, address := range slice {
		zap.L().Info(
			"Processing contract...",
			zap.String("contract_address", address.Hex()),
		)

		cacheKey := types.GetContractStorageKey(w.chainId, address)

		exists, err := w.redis.Exists(w.ctx, cacheKey)
		if err != nil {
			zap.L().Error(
				ErrFailedToCheckIfMethodCacheKeyExists.Error(),
				zap.String("contract_address", address.Hex()),
				zap.Error(err),
			)
			return err
		}

		if exists {
			continue
		}

		if w.clickhouseDb != nil {
			dbExists, err := models.ContractExists(w.ctx, w.clickhouseDb, address)
			if err != nil {
				zap.L().Error(
					ErrFailedToCheckIfMethodCacheKeyExists.Error(),
					zap.String("contract_address", address.Hex()),
					zap.Error(err),
				)
				return err
			}

			if dbExists {
				continue
			}
		}

		metadata, err := sourcify.GetContractMetadata(w.provider, int(chainID.Uint64()), address, contractType)
		if err != nil {
			continue
		}

		metadataBytes, err := sourcify.GetContractMetadataAsBytes(w.provider, int(chainID.Uint64()), address, contractType)
		if err != nil {
			continue
		}

		sources, err := sourcify.GetContractSourceCode(w.provider, int(chainID.Uint64()), address, contractType)
		if err != nil {
			zap.L().Error(
				"failed to get contract files",
				zap.String("contract_address", address.Hex()),
				zap.Error(err),
			)
			continue
		}

		contract, err := types.NewContractFromSourcify(chainID, address, metadata, metadataBytes)
		if err != nil {
			zap.L().Error(
				"failed to create contract from sourcify metadata",
				zap.String("contract_address", address.Hex()),
				zap.Error(err),
			)
			continue
		}

		// Search for source code...
		for entryPointTarget, _ := range metadata.Settings.CompilationTarget {
			if metadata.Sources != nil && len(metadata.Sources) > 0 {
				for _, source := range sources.Code {
					if strings.Contains(source.Path, entryPointTarget) {
						contract.SourceCode = source.Content
					}
				}

			}
			break
		}

		// Figure out verification information...
		verification, err := sourcify.CheckContractByAddresses(w.provider, []string{address.Hex()}, []int{int(chainID.Uint64())}, contractType)
		if err != nil {
			zap.L().Error(
				ErrFailedContractValidationCheck.Error(),
				zap.String("contract_address", address.Hex()),
				zap.Error(err),
			)
			continue
		}

		if len(verification) > 0 {
			for _, v := range verification {
				if v.Status != "" {
					contract.VerificationType = types.ContractVerificationTypeSourcify
					contract.VerificationStatus = v.Status
				}
			}
		} else {
			contract.VerificationType = types.ContractVerificationTypeSourcify
			contract.VerificationStatus = "unknown"
		}

		// Now once we have all of the information that we are interested in
		// it's time to do some kebabs, and figure out block, transaction of the contract...
		// TODO: Figure out how to do this in a better way...
		queryData := map[string]string{
			"query": fmt.Sprintf(`{
			smartContractCreation: ethereum(network: %s) {
			  smartContractCalls(
				smartContractAddress: {is: "%s"}
				smartContractMethod: {is: "Contract Creation"}
			  ) {
				transaction {
				  hash
				}
				block {
				  height
				}
			  }
			}
		  }`, queryChainToName[chainID.Int64()], address.Hex()),
		}

		bitqueryInfo, err := w.bitquery.GetContractCreationInfo(queryData)
		if err != nil {
			zap.L().Error(
				"failed to get contract creation info from bitquery",
				zap.String("contract_address", address.Hex()),
				zap.Error(err),
			)
			continue
		}

		if len(bitqueryInfo.Data.SmartContractCreation.SmartContractCalls) > 0 {
			contract.TransactionHash = common.HexToHash(
				bitqueryInfo.Data.SmartContractCreation.SmartContractCalls[0].Transaction.Hash,
			)

			txReceipt, err := helpers.GetReceiptByHash(w.ctx, w.ethClient, contract.TransactionHash)
			if err != nil {
				zap.L().Error(
					ErrFailedGetTransactionReceiptByHash.Error(),
					zap.String("contract_address", address.Hex()),
					zap.String("tx_hash", contract.TransactionHash.Hex()),
					zap.Error(err),
				)
				continue
			}

			contract.BlockHash = txReceipt.BlockHash
		}

		// Figure out if we have constructor abi...
		for _, abi := range metadata.Output.Abi {
			if abi.Type == "constructor" {
				constructorAbi, err := json.Marshal(abi)
				if err != nil {
					zap.L().Error(
						ErrFailedToMarshalConstructorAbi.Error(),
						zap.String("contract_address", address.Hex()),
						zap.Error(err),
					)
					continue
				}
				contract.ConstructorABI = string(constructorAbi)
			}
		}

		// Now this is a bit of a hack to attempt filling up the contract information in case
		// sourcify is missing information and tools such as bscscan or etherscan have it...
		// TODO: Etherscan...
		// TODO: Seems that both BSCScan and Sourcify have the same source code but no ABI. See: 0x005D5631EF919DcDa961f0DE1539d62E3f0eBf37
		if contract.SourceCode == "" || contract.ABI == "" || contract.LicenseType == "" ||
			contract.Name == "" || contract.ConstructorArguments == "" || contract.Proxy == "" {
			contractResult, err := w.bscscan.ScanContract(contract.Address.Hex())
			if err == nil {
				if contractResult.SourceCode != "" && contract.SourceCode == "" {
					contract.SourceCode = contractResult.SourceCode
				}

				if contractResult.ABI != "" && (contract.ABI == "" || contract.ABI == "[]") {
					contract.ABI = contractResult.ABI
				}

				if contractResult.LicenseType != "" && contract.LicenseType == "" {
					contract.LicenseType = contractResult.LicenseType
				}

				// One from BSCScan is more reliable...
				if contractResult.Name != contract.Name {
					contract.Name = contractResult.Name
				}

				if contractResult.ConstructorArguments != "" {
					contract.ConstructorArguments = contractResult.ConstructorArguments
				}

				if contractResult.Proxy != "" {
					if contractResult.Proxy == "1" {
						contract.Proxy = "true"
					}
					contract.Proxy = "false"
				} else {
					contract.Proxy = "false"
				}
			}
		}

		// Now we have all of the information that we need to write the contract to the database...
		if err := w.WriteContract(contract); err != nil {
			zap.L().Error(
				ErrFailedToWriteContractToDatabase.Error(),
				zap.String("contract_address", address.Hex()),
				zap.Error(err),
			)
			continue
		}

	}

	return nil
}

func (w *SourcifyWriter) WriteContract(contract *types.Contract) error {
	cacheKey := types.GetContractStorageKey(w.chainId, contract.Address)

	exists, err := w.redis.Exists(w.ctx, cacheKey)
	if err != nil {
		zap.L().Error(
			ErrFailedToCheckIfMethodCacheKeyExists.Error(),
			zap.String("contract_address", contract.Address.Hex()),
			zap.Error(err),
		)
		return err
	}

	if exists {
		return ErrContractAlreadyExists
	}

	// Insert contract into the clickhouse database but only if clickhouse database is set
	if w.clickhouseDb != nil {
		if err := models.InsertContract(w.ctx, w.clickhouseDb, contract); err != nil {
			return err
		}
	}

	// Process abi methods, events, resulting contract mappings and write them into the database for future use
	if err := w.processAbi(w.ctx, contract); err != nil {
		zap.L().Error(
			ErrFailedProcessAbi.Error(),
			zap.String("contract_address", contract.Address.Hex()),
			zap.Error(err),
		)
		return err
	}

	resBytes, err := contract.MarshalBytes()
	if err != nil {
		zap.L().Error(
			ErrFailedMarshalContract.Error(),
			zap.String("contract_address", contract.Address.Hex()),
			zap.Error(err),
		)
		return err
	}

	err = w.redis.Write(w.ctx, cacheKey, resBytes, 0)
	if err != nil {
		zap.L().Error(
			ErrFailedWriteContractToRedis.Error(),
			zap.String("contract_address", contract.Address.Hex()),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (w *SourcifyWriter) processAbi(ctx context.Context, contract *types.Contract) error {
	abi, err := abi.JSON(strings.NewReader(contract.ABI))
	if err != nil {
		zap.L().Error(
			ErrFailedParseAbi.Error(),
			zap.String("contract_address", contract.Address.Hex()),
			zap.Error(err),
		)
		return err
	}

	// Process abi methods and events and write them into the database for future use
	if err := w.processAbiMethods(contract, abi.Methods); err != nil {
		zap.L().Error(
			ErrFailedProcessAbiMethods.Error(),
			zap.String("contract_address", contract.Address.Hex()),
			zap.Error(err),
		)
		return err
	}

	if err := w.processAbiEvents(ctx, contract, abi.Events); err != nil {
		zap.L().Error(
			ErrFailedProcessAbiEvents.Error(),
			zap.String("contract_address", contract.Address.Hex()),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func (w *SourcifyWriter) processAbiMethods(contract *types.Contract, methods map[string]abi.Method) error {
	for _, method := range methods {
		select {
		case <-w.ctx.Done():
			return w.ctx.Err()
		default:
			methodKey := types.GetMethodStorageKey(w.chainId, method.ID)
			methodMapperKey := types.GetMethodMapperStorageKey(w.chainId, method.ID)

			exists, err := w.redis.Exists(w.ctx, methodKey)
			if err != nil {
				zap.L().Error(
					ErrFailedToReadRedis.Error(),
					zap.String("contract_address", contract.Address.Hex()),
					zap.String("method_name", method.Name),
					zap.Error(err),
				)
				return err
			}

			// Just skip if the method already exists in BadgerDB
			if exists {
				continue
			}

			methodResult := types.NewFullMethod(method)
			methodMappingResult := types.NewMethodMapping(contract, methodResult)

			// Write into clickhouse but only if clickhouse database is set
			if w.clickhouseDb != nil {
				methodExists, err := models.MethodExists(w.ctx, w.clickhouseDb, methodResult)
				if err != nil {
					zap.L().Error(
						ErrFailedToCheckIfMethodExists.Error(),
						zap.String("method_name", method.Name),
						zap.Error(err),
					)
					continue
				}

				if !methodExists {
					if err := models.InsertMethod(w.ctx, w.clickhouseDb, methodResult); err != nil {
						zap.L().Error(
							ErrFailedToInsertMethod.Error(),
							zap.Error(err),
							zap.String("contract_address", contract.Address.Hex()),
							zap.String("method_name", method.Name),
						)
						return err
					}

					if err := models.InsertMethodMapping(w.ctx, w.clickhouseDb, methodMappingResult); err != nil {
						zap.L().Error(
							ErrFailedToInsertMethodMapping.Error(),
							zap.Error(err),
							zap.String("contract_address", contract.Address.Hex()),
							zap.String("method_name", method.Name),
						)
						return err
					}
				}
			}

			methodBytes, err := methodResult.MarshalBytes()
			if err != nil {
				zap.L().Error(
					ErrFailedMarshalMethod.Error(),
					zap.String("contract_address", contract.Address.Hex()),
					zap.String("method_name", method.Name),
					zap.Error(err),
				)
				return err
			}

			methodMappingBytes, err := methodMappingResult.MarshalBytes()
			if err != nil {
				zap.L().Error(
					ErrFailedMarshalMethodMapping.Error(),
					zap.String("contract_address", contract.Address.Hex()),
					zap.String("method_name", method.Name),
					zap.Error(err),
				)
				return err
			}

			if err := w.redis.Write(w.ctx, methodKey, methodBytes, 0); err != nil {
				zap.L().Error(
					ErrFailedWriteMethodToRedis.Error(),
					zap.String("contract_address", contract.Address.Hex()),
					zap.String("method_name", method.Name),
					zap.Error(err),
				)
				return err
			}

			if err := w.redis.Write(w.ctx, methodMapperKey, methodMappingBytes, 0); err != nil {
				zap.L().Error(
					ErrFailedWriteMethodMappingToRedis.Error(),
					zap.String("contract_address", contract.Address.Hex()),
					zap.String("method_name", method.Name),
					zap.Error(err),
				)
				return err
			}
		}
	}

	return nil
}

func (w *SourcifyWriter) processAbiEvents(ctx context.Context, contract *types.Contract, events map[string]abi.Event) error {
	for _, event := range events {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			eventKey := types.GetEventStorageKey(w.chainId, event.ID)
			eventMappingKey := types.GetEventMapperStorageKey(w.chainId, event.ID)

			exists, err := w.redis.Exists(w.ctx, eventKey)
			if err != nil {
				zap.L().Error(
					ErrFailedCheckEventExistenceInRedis.Error(),
					zap.String("contract_address", contract.Address.Hex()),
					zap.String("event_name", event.Name),
					zap.Error(err),
				)
				return err
			}

			// Just skip if the event already exists in BadgerDB
			if exists {
				continue
			}

			if w.clickhouseDb != nil {
				dbExists, err := models.EventExist(w.ctx, w.clickhouseDb, event.ID)
				if err != nil {
					zap.L().Error(
						ErrFailedToCheckIfMethodCacheKeyExists.Error(),
						zap.String("contract_address", contract.Address.Hex()),
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
			if w.clickhouseDb != nil {
				if err := models.InsertEvent(w.ctx, w.clickhouseDb, eventResult); err != nil {
					zap.L().Error(
						ErrFailedToInsertEvent.Error(),
						zap.Error(err),
						zap.String("contract_address", contract.Address.Hex()),
						zap.String("method_name", eventResult.Name),
					)
					return err
				}
			}

			eventMappingResult := types.NewEventMapping(contract, eventResult)
			if w.clickhouseDb != nil {
				if err := models.InsertEventMapping(w.ctx, w.clickhouseDb, eventMappingResult); err != nil {
					zap.L().Error(
						ErrFailedToInsertEventMapping.Error(),
						zap.Error(err),
						zap.String("contract_address", contract.Address.Hex()),
						zap.String("method_name", eventResult.Name),
					)
					return err
				}
			}

			eventBytes, err := eventResult.MarshalBytes()
			if err != nil {
				zap.L().Error(
					ErrFailedToMarshalEvent.Error(),
					zap.String("contract_address", contract.Address.Hex()),
					zap.String("event_name", event.Name),
					zap.Error(err),
				)
				return err
			}

			if err := w.redis.Write(w.ctx, eventKey, eventBytes, 0); err != nil {
				zap.L().Error(
					ErrFailedToWriteEventToRedis.Error(),
					zap.String("contract_address", contract.Address.Hex()),
					zap.String("event_name", event.Name),
					zap.Error(err),
				)
				return err
			}

			eventMapperBytes, err := eventMappingResult.MarshalBytes()
			if err != nil {
				zap.L().Error(
					ErrFailedToMarshalEvent.Error(),
					zap.String("contract_address", contract.Address.Hex()),
					zap.String("event_name", event.Name),
					zap.Error(err),
				)
				return err
			}

			if err := w.redis.Write(w.ctx, eventMappingKey, eventMapperBytes, 0); err != nil {
				zap.L().Error(
					ErrFailedToWriteEventMappingInfoToRedis.Error(),
					zap.String("contract_address", contract.Address.Hex()),
					zap.String("event_name", event.Name),
					zap.Error(err),
				)
				return err
			}
		}

	}

	return nil
}
