package contract

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/txpull/bytecode/abi"
	"github.com/txpull/bytecode/controlflow"
	"github.com/txpull/bytecode/opcodes"
	"github.com/txpull/bytecode/signatures"
	"github.com/txpull/bytecode/solidity"
)

// ContractCreationTxResult is a structure that encapsulates the result of processing
// an Ethereum contract creation transaction. It includes various data such as
// block number, transaction hash, contract address, runtime bytecode, etc. extracted
// from the transaction. It also includes decompiled versions of runtime and contract bytecodes,
// decoded contract ABI, decoded source code, signatures, and a flag indicating if the contract is verified.
type ContractCreationTxResult struct {
	// BlockNumber represents the block number where the contract creation transaction was included.
	BlockNumber *big.Int `json:"block_number"`

	// TransactionHash represents the hash of the contract creation transaction.
	TransactionHash common.Hash `json:"transaction_hash"`

	// ReceiptStatus represents the status of the transaction receipt.
	ReceiptStatus uint64 `json:"receipt_status"`

	// ContractAddress represents the address of the created contract.
	ContractAddress common.Address `json:"contract_address"`

	// RuntimeBytecode represents the bytecode of the contract at runtime.
	RuntimeBytecode []byte `json:"runtime_bytecode"`

	// RuntimeBytecodeSize represents the size of the runtime bytecode.
	RuntimeBytecodeSize uint64 `json:"runtime_bytecode_size"`

	// RuntimeOpCodes represents the decompiled version of the runtime bytecode.
	RuntimeOpCodes *opcodes.Decompiler `json:"-"`

	// Bytecode represents the bytecode of the contract.
	Bytecode []byte `json:"bytecode"`

	// BytecodeSize represents the size of the contract bytecode.
	BytecodeSize uint64 `json:"bytecode_size"`

	// OpCodes represents the decompiled version of the contract bytecode.
	OpCodes *opcodes.Decompiler `json:"-"`

	// ControlFlowGraph represents the control flow graph of the contract.
	ControlFlowGraph *controlflow.Decoder `json:"-"`

	// ABI represents the decoded contract ABI.
	ABI *abi.Decoder `json:"-"`

	// SourceCode represents the decoded source code of the contract.
	SourceCode *solidity.Decoder `json:"-"`

	// Signatures represents the signatures used in the contract.
	Signatures []*signatures.Signature `json:"signatures"`

	// ContractVerified indicates whether the contract has been verified or not.
	ContractVerified bool `json:"contract_verified"`
}
