// Package contract provides a customizable Decoder
// which uses the Option pattern to set configurations.
package contracts

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/txpull/sourcify-go"
	"github.com/txpull/unpack/abis"
	"github.com/txpull/unpack/clients"
	"github.com/txpull/unpack/readers"
	"github.com/txpull/unpack/scanners"
	"github.com/txpull/unpack/types"
	"go.uber.org/zap"
)

// Decoder is a structure that holds a context, a BadgerDB instance and an EthClient instance.
// The context, the BadgerDB instance, and the EthClient instance within the Decoder
// can be customized via the Option functions.
type Decoder struct {
	// ctx holds the context to be used by the Decoder. It can be customized via WithCtx Option.
	ctx context.Context

	sourcifyClient *sourcify.Client
	readerManager  *readers.Manager
	bitquery       *scanners.BitQueryProvider
	ethClient      *clients.EthClient
	bscscan        *scanners.BscScanProvider
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
		w.sourcifyClient = client
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

	return decoder, nil
}

func (c *Decoder) DecodeByAddress(chainId *big.Int, addr common.Address, abi *abis.Decoder) (*ContractResponse, error) {
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
			return nil, err
		}

		return response, nil
	}

	zap.L().Info(
		"Contract not found in any of the database readers, trying to fetch it from the blockchain",
		zap.String("address", addr.Hex()),
		zap.Int64("chain_id", chainId.Int64()),
	)

	return nil, nil
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
		Address:         contract.Address,
		RuntimeBytecode: nil,
		Abi:             abiDecoder,
	}, nil
}
