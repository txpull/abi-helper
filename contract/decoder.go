// Package contract provides a customizable ContractDecoder
// which uses the Option pattern to set configurations.
package contract

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/txpull/bytecode/clients"
	"github.com/txpull/bytecode/controlflow"
	"github.com/txpull/bytecode/db"
	"github.com/txpull/bytecode/opcodes"
	"go.uber.org/zap"
)

// ContractDecoder is a structure that holds a context, a BadgerDB instance and an EthClient instance.
// The context, the BadgerDB instance, and the EthClient instance within the ContractDecoder
// can be customized via the Option functions.
type ContractDecoder struct {
	// ctx holds the context to be used by the ContractDecoder. It can be customized via WithCtx Option.
	ctx context.Context
	// bdb is the BadgerDB instance to be used by the ContractDecoder. It can be customized via WithDB Option.
	bdb *db.BadgerDB
	// ethClient is the EthClient instance to be used by the ContractDecoder. It can be customized via WithEthClient Option.
	ethClient *clients.EthClient
}

// Option defines a function type that applies configurations to a ContractDecoder.
// It is used to customize the context held by the ContractDecoder.
type Option func(*ContractDecoder)

// WithContext is an Option that allows to set a custom context to the ContractDecoder.
// It returns an Option function which, when executed, modifies the context within the ContractDecoder.
//
// ctx: The context to be set in the ContractDecoder.
//
// Example usage:
//
//	r := NewContract(WithContext(myContext))
func WithContext(ctx context.Context) Option {
	return func(c *ContractDecoder) {
		c.ctx = ctx
	}
}

// WithDB is an Option that allows setting a custom BadgerDB instance to the ContractDecoder.
// It returns an Option function which, when executed, modifies the BadgerDB instance within the ContractDecoder.
//
// bdb: The BadgerDB instance to be set in the ContractDecoder.
//
// Example usage:
//
//	r := NewContract(WithDB(myBadgerDB))
func WithDB(bdb *db.BadgerDB) Option {
	return func(c *ContractDecoder) {
		c.bdb = bdb
	}
}

// WithEthClient is an Option that allows setting a custom EthClient instance to the ContractDecoder.
// It returns an Option function which, when executed, modifies the EthClient instance within the ContractDecoder.
//
// client: The EthClient instance to be set in the ContractDecoder.
//
// Example usage:
//
//	r := NewContract(WithEthClient(myEthClient))
func WithEthClient(client *clients.EthClient) Option {
	return func(c *ContractDecoder) {
		c.ethClient = client
	}
}

// NewContractDecoder creates a new ContractDecoder and applies any provided Option functions to it.
// By default, the ContractDecoder is created with a Background context, which can be customized
// using the provided Option functions.
//
// opts: Variadic arguments of type Option to customize the created ContractDecoder.
//
// Returns a pointer to a ContractDecoder.
//
// Example usage:
//
//	r := NewContractDecoder(WithCtx(myContext))
func NewContractDecoder(opts ...Option) *ContractDecoder {
	reader := &ContractDecoder{
		ctx: context.Background(),
	}

	// Apply all options to reader
	for _, opt := range opts {
		opt(reader)
	}

	return reader
}

// ProcessContractCreationTx is a method that processes an Ethereum contract creation transaction,
// decomposes it into bytecodes, and returns a ContractCreationTxResult structure.
// If the passed block, transaction or receipt is nil, it returns an error.
// If it fails to decompile the transaction data or to fetch the contract bytecode, it logs the error and returns it.
// If the transaction data and the contract bytecode are successfully decomposed, it returns a ContractCreationTxResult
// structure that contains these bytecodes along with other information from the transaction.
func (c *ContractDecoder) ProcessContractCreationTx(block *types.Block, tx *types.Transaction, receipt *types.Receipt, contractAddress common.Address) (*ContractCreationTxResult, error) {
	// We need block at this stage to process contract creation transaction :lol:
	if tx == nil {
		return nil, ErrMissingBlock
	}

	// We really need transaction at this stage to process contract creation transaction :lol:
	if tx == nil {
		return nil, ErrMissingTransaction
	}

	// We need receipt at this stage to process contract creation transaction :lol:
	if tx == nil {
		return nil, ErrMissingReceipt
	}

	runtimeDecompiler := opcodes.NewDecompiler(c.ctx, tx.Data())

	if err := runtimeDecompiler.Decompile(); err != nil {
		zap.L().Error("failed to decompile transaction", zap.String("tx_hash", tx.Hash().Hex()), zap.Error(err))
		return nil, err
	}

	// Get the bytecode from the transaction
	bytecode, err := c.ethClient.GetClient().CodeAt(c.ctx, contractAddress, block.Number())
	if err != nil {
		zap.L().Error("failed to get contract bytecode", zap.String("tx_hash", tx.Hash().Hex()), zap.Error(err))
		return nil, err
	}

	toReturn := &ContractCreationTxResult{
		BlockNumber:         block.Number(),
		TransactionHash:     tx.Hash(),
		ReceiptStatus:       receipt.Status,
		ContractAddress:     contractAddress,
		RuntimeBytecode:     runtimeDecompiler.GetBytecode(),
		RuntimeBytecodeSize: runtimeDecompiler.GetBytecodeSize(),
		RuntimeOpCodes:      runtimeDecompiler,
	}

	decompiler := opcodes.NewDecompiler(c.ctx, bytecode)
	if len(bytecode) > 0 {
		if err := decompiler.Decompile(); err != nil {
			zap.L().Error("failed to decompile contract bytecode", zap.String("tx_hash", tx.Hash().Hex()), zap.Error(err))
			return nil, err
		}

		toReturn.Bytecode = decompiler.GetBytecode()
		toReturn.BytecodeSize = decompiler.GetBytecodeSize()
		toReturn.OpCodes = decompiler

		graph, err := controlflow.NewDecoder(decompiler)
		if err != nil {
			zap.L().Error("failed to create control flow graph", zap.String("tx_hash", tx.Hash().Hex()), zap.Error(err))
			return nil, err
		}

		toReturn.ControlFlowGraph = graph

		/* 	// Perform path analysis and obtain the paths
		paths := graph.FindAllPaths()

		// Access the paths
		fmt.Println("Number of paths:", len(paths))
		for i, path := range paths {
			fmt.Printf("Path %d: ", i+1)
			for _, node := range path {
				fmt.Printf("%d -> ", node.Offset)
			}
			fmt.Println("Exit")
		} */
	}

	return toReturn, nil
}
