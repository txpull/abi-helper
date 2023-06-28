// Package contract provides a customizable Decoder
// which uses the Option pattern to set configurations.
package contracts

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/txpull/sourcify-go"
	"github.com/txpull/unpack/abis"
	"github.com/txpull/unpack/clients"
	"github.com/txpull/unpack/helpers"
	"github.com/txpull/unpack/readers"
	"github.com/txpull/unpack/scanners"
	"github.com/txpull/unpack/solidity"
	"github.com/txpull/unpack/types"
	"go.uber.org/zap"
)

// Decoder is a structure that holds a context, a BadgerDB instance and an EthClient instance.
// The context, the BadgerDB instance, and the EthClient instance within the Decoder
// can be customized via the Option functions.
type Decoder struct {
	// ctx holds the context to be used by the Decoder. It can be customized via WithCtx Option.
	ctx context.Context

	sourcify      *sourcify.Client
	readerManager *readers.Manager
	bitquery      *scanners.BitQueryProvider
	ethClient     *clients.EthClient
	bscscan       *scanners.BscScanProvider
	writer        *Writer
}

// Option defines a function type that applies configurations to a Decoder.
// It is used to customize the context held by the Decoder.
type Option func(*Decoder)

func WithBitQuery(bq *scanners.BitQueryProvider) Option {
	return func(c *Decoder) {
		c.bitquery = bq
	}
}

func WithEthClient(client *clients.EthClient) Option {
	return func(w *Decoder) {
		w.ethClient = client
	}
}

func WithReaderManager(manager *readers.Manager) Option {
	return func(w *Decoder) {
		w.readerManager = manager
	}
}

func WithBscScan(client *scanners.BscScanProvider) Option {
	return func(w *Decoder) {
		w.bscscan = client
	}
}

func WithSourcify(client *sourcify.Client) Option {
	return func(w *Decoder) {
		w.sourcify = client
	}
}

func NewDecoder(ctx context.Context, opts ...Option) (*Decoder, error) {
	decoder := &Decoder{ctx: ctx}

	// Apply all options to decoder
	for _, opt := range opts {
		opt(decoder)
	}

	if decoder.ethClient == nil {
		return nil, errors.New("eth client is required")
	}

	// We can look into redis, clickhouse or both, but we need at least one of them.
	if decoder.readerManager == nil {
		return nil, errors.New("reader manager is required")
	}

	if decoder.bitquery == nil {
		return nil, errors.New("bitquery client is required")
	}

	decoder.writer = &Writer{
		ctx:       ctx,
		ethClient: decoder.ethClient,
		bitquery:  decoder.bitquery,
		bscscan:   decoder.bscscan,
		sourcify:  decoder.sourcify,
	}

	return decoder, nil
}

func (c *Decoder) DecodeByAddress(chainId *big.Int, addr common.Address, abi *abis.Decoder) (*ContractResponse, bool, error) {
	for _, reader := range c.readerManager.GetSortedReaders() {
		contract, err := reader.GetContractByAddress(chainId, addr)
		if err != nil {
			zap.L().Error(
				"failed to get contract by address by selected reader",
				zap.String("address", addr.Hex()),
				zap.Int64("chain_id", chainId.Int64()),
				zap.String("reader_name", reader.String()),
				zap.Error(err),
			)
			continue
		}

		response, err := c.buildContractResponse(contract)
		if err != nil {
			return nil, false, err
		}

		return response, true, nil
	}

	zap.L().Info(
		"Contract not found in any of the database readers, trying to fetch it from the blockchain",
		zap.String("address", addr.Hex()),
		zap.Int64("chain_id", chainId.Int64()),
	)

	// From this moment onwards, response will take a bit of the time...
	// Before this it is usually around 0.3ms, but after this it can take up to 0.8s.
	// If we are here, it means that we couldn't find the contract in any of the readers.
	// Now we will try to figure out if there's anything new at bscscan or sourcify.
	// If it is, we will add it to the database and return it.
	// If it's not, we will return an error.
	// FUTURE WORK: We can use OpCodes, AST, CFG, etc. and database to figure out the contract information
	// without relying on bscscan or sourcify within the best of our abilities.
	valid, bytecode, err := c.GetDeployedBytecode(chainId, addr, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to check if address is a contract: %s", err)
	} else if !valid {
		return nil, false, errors.New("address is not a valid contract")
	}

	// We are going to cast workload into the ContractResponse and later on call ToModel()
	// to save it into the available databases.
	contract := ContractResponse{
		ChainId: chainId, Address: addr,
		ContractBytecode: bytecode, ContractBytecodeSize: len(bytecode),
	}

	// Get information from the node about the contract.
	if err := c.GetStateInformation(chainId, addr, &contract); err != nil {
		return nil, false, fmt.Errorf("failed to get block information: %s", err)
	}

	// Attempt retrieving information from bscscan
	if err := c.GetBscScanInformation(chainId, addr, &contract); err != nil {
		zap.L().Error(
			"failed to get bscscan information",
			zap.String("address", addr.Hex()),
			zap.Int64("chain_id", chainId.Int64()),
			zap.Error(err),
		)
	}

	// Attempt retrieving information from sourcify

	return &contract, true, nil
}

func (c *Decoder) GetBscScanInformation(chainId *big.Int, addr common.Address, contract *ContractResponse) error {
	response, err := c.bscscan.ScanContract(addr.Hex())
	if err != nil {
		return err
	}

	runs, err := strconv.Atoi(response.Runs)
	if err != nil {
		return fmt.Errorf("failed to convert runs to int: %s", err)
	}

	optimizationUsed, err := strconv.ParseBool(response.OptimizationUsed)
	if err != nil {
		return fmt.Errorf("failed to convert optimization used to bool: %s", err)
	}

	isProxy, err := strconv.ParseBool(response.Proxy)
	if err != nil {
		return fmt.Errorf("failed to convert optimization used to bool: %s", err)
	}

	contract.Name = response.Name
	contract.LicenseType = response.LicenseType
	contract.Runs = runs
	contract.ConstructorArguments = response.ConstructorArguments
	contract.ConstructorArgumentsSize = len(response.ConstructorArguments)
	contract.OptimizationUsed = optimizationUsed
	contract.CompilerVersion = response.CompilerVersion
	contract.SetLanguageFromCompilerVersion(response.CompilerVersion)
	contract.ProxyDetected = isProxy
	contract.Library = response.Library
	contract.Implementation = common.HexToAddress(response.Implementation)
	contract.SourceUrls = strings.Split(response.SwarmSource, ",")
	contract.VerificationType = types.ContractVerificationTypeBscscan
	contract.VerificationStatus = "verified:full"
	contract.ProcessStatus = ContractProcessStatusPending

	// If we have the source code, we can try to get the OpCodes, AST and CFG and other things...
	if len(response.SourceCode) > 0 {
		contract.SourceCode = solidity.NewDecoder(c.ctx, response.SourceCode)
	}

	// If we have ABI, we are actually really happy as it makes our lives much easier.
	// Problem is that I am not trusting third party sources to do ABIs for us as they can
	// be easily manipulated. So we will only use it if we have no other choice.
	if len(response.ABI) > 0 {
		contract.Abi, err = abis.NewDecoder(c.ctx, c.readerManager, response.ABI)
		if err != nil {
			return fmt.Errorf("failed to decode ABI: %s", err)
		}
	}

	return nil
}

func (c *Decoder) GetStateInformation(chainId *big.Int, addr common.Address, contract *ContractResponse) error {
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
	  }`, helpers.GetChainName(chainId), addr.Hex()),
	}

	bitqueryInfo, err := c.bitquery.GetContractCreationInfo(queryData)
	if err != nil {
		zap.L().Error(
			"failed to get contract creation info from bitquery",
			zap.String("contract_address", addr.Hex()),
			zap.Error(err),
		)
		return err
	}

	if len(bitqueryInfo.Data.SmartContractCreation.SmartContractCalls) == 0 {
		return errors.New("no contract information found on bitquery")
	}

	contract.TransactionHash = common.HexToHash(
		bitqueryInfo.Data.SmartContractCreation.SmartContractCalls[0].Transaction.Hash,
	)

	contract.BlockNumber = big.NewInt(
		int64(bitqueryInfo.Data.SmartContractCreation.SmartContractCalls[0].Block.Height),
	)

	tx, _, err := helpers.GetTransactionByHash(c.ctx, chainId, c.ethClient, contract.TransactionHash)
	if err != nil {
		zap.L().Error(
			ErrFailedGetTransactionByHash.Error(),
			zap.String("contract_address", addr.Hex()),
			zap.String("tx_hash", contract.TransactionHash.Hex()),
			zap.Error(err),
		)
		return err
	}

	contract.CreationBytecode = tx.Data()
	contract.CreationBytecodeSize = len(contract.CreationBytecode)

	txReceipt, err := helpers.GetReceiptByHash(c.ctx, chainId, c.ethClient, contract.TransactionHash)
	if err != nil {
		zap.L().Error(
			ErrFailedGetTransactionReceiptByHash.Error(),
			zap.String("contract_address", addr.Hex()),
			zap.String("tx_hash", contract.TransactionHash.Hex()),
			zap.Error(err),
		)
		return err
	}

	contract.BlockHash = txReceipt.BlockHash
	contract.ReceiptStatus = txReceipt.Status

	return nil
}

func (c *Decoder) GetDeployedBytecode(chainId *big.Int, addr common.Address, atBlock *big.Int) (bool, []byte, error) {
	bytecode, err := helpers.GetBytecode(c.ctx, chainId, c.ethClient, addr, atBlock)
	if err != nil {
		return false, nil, err
	}

	return len(bytecode) > 0, bytecode, nil
}

// TODO: Add receipt information including log parsing
func (c *Decoder) buildContractResponse(contract *types.Contract) (*ContractResponse, error) {
	if contract == nil {
		return nil, errors.New("contract is nil")
	}

	abiDecoder, err := abis.NewDecoder(c.ctx, c.readerManager, contract.ABI)
	if err != nil {
		return nil, fmt.Errorf("failed to create abi decoder: %s", err)
	}

	return &ContractResponse{
		BlockHash:       contract.BlockHash,
		TransactionHash: contract.TransactionHash,
		// TODO: Add receipt information including log parsing,
		Address: contract.Address,
		Abi:     abiDecoder,
	}, nil
}
