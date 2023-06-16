// Package contract provides a customizable ContractReader
// which uses the Option pattern to set configurations.
package contract

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/txpull/bytecode/clients"
	"github.com/txpull/bytecode/db"
)

// ContractReader is a structure that holds a context, a BadgerDB instance and an EthClient instance.
// The context, the BadgerDB instance, and the EthClient instance within the ContractReader
// can be customized via the Option functions.
type ContractReader struct {
	// ctx holds the context to be used by the ContractReader. It can be customized via WithCtx Option.
	ctx context.Context
	// bdb is the BadgerDB instance to be used by the ContractReader. It can be customized via WithDB Option.
	bdb *db.BadgerDB
	// ethClient is the EthClient instance to be used by the ContractReader. It can be customized via WithEthClient Option.
	ethClient *clients.EthClient
}

// Option defines a function type that applies configurations to a ContractReader.
// It is used to customize the context held by the ContractReader.
type Option func(*ContractReader)

// WithContext is an Option that allows to set a custom context to the ContractReader.
// It returns an Option function which, when executed, modifies the context within the ContractReader.
//
// ctx: The context to be set in the ContractReader.
//
// Example usage:
//
//	r := NewContract(WithContext(myContext))
func WithContext(ctx context.Context) Option {
	return func(c *ContractReader) {
		c.ctx = ctx
	}
}

// WithDB is an Option that allows setting a custom BadgerDB instance to the ContractReader.
// It returns an Option function which, when executed, modifies the BadgerDB instance within the ContractReader.
//
// bdb: The BadgerDB instance to be set in the ContractReader.
//
// Example usage:
//
//	r := NewContract(WithDB(myBadgerDB))
func WithDB(bdb *db.BadgerDB) Option {
	return func(c *ContractReader) {
		c.bdb = bdb
	}
}

// WithEthClient is an Option that allows setting a custom EthClient instance to the ContractReader.
// It returns an Option function which, when executed, modifies the EthClient instance within the ContractReader.
//
// client: The EthClient instance to be set in the ContractReader.
//
// Example usage:
//
//	r := NewContract(WithEthClient(myEthClient))
func WithEthClient(client *clients.EthClient) Option {
	return func(c *ContractReader) {
		c.ethClient = client
	}
}

// NewContract creates a new ContractReader and applies any provided Option functions to it.
// By default, the ContractReader is created with a Background context, which can be customized
// using the provided Option functions.
//
// opts: Variadic arguments of type Option to customize the created ContractReader.
//
// Returns a pointer to a ContractReader.
//
// Example usage:
//
//	r := NewContract(WithCtx(myContext))
func NewContract(opts ...Option) *ContractReader {
	reader := &ContractReader{
		ctx: context.Background(),
	}

	// Apply all options to reader
	for _, opt := range opts {
		opt(reader)
	}

	return reader
}

func (c *ContractReader) ProcessContractCreationTx(block *types.Block, tx *types.Transaction, receipt *types.Receipt, contractAddress common.Address) (*struct{}, error) {
	return nil, nil
}
