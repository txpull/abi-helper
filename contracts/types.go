package contracts

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/txpull/unpack/abis"
	"github.com/txpull/unpack/opcodes"
)

type ContractResponse struct {
	// BlockHash represents the hash of the block where the contract was created.
	BlockHash common.Hash `json:"block_hash"`

	// TransactionHash represents the hash of the contract creation transaction.
	TransactionHash common.Hash `json:"transaction_hash"`

	// ReceiptStatus represents the status of the transaction receipt.
	ReceiptStatus uint64 `json:"receipt_status"`

	// Address represents the address of the created contract.
	Address common.Address `json:"contract_address"`

	// RuntimeBytecode represents the bytecode of the contract at runtime.
	RuntimeBytecode []byte `json:"runtime_bytecode"`

	// RuntimeBytecodeSize represents the size of the runtime bytecode.
	RuntimeBytecodeSize uint64 `json:"runtime_bytecode_size"`

	// RuntimeOpCodes represents the decompiled version of the runtime bytecode.
	RuntimeOpCodes []opcodes.OpCode `json:"runtime_opcodes"`

	// ContractBytecode represents the bytecode of the contract.
	ContractBytecode []byte `json:"contract_bytecode"`

	// ContractBytecodeSize represents the size of the contract bytecode.
	ContractBytecodeSize uint64 `json:"contract_bytecode_size"`

	// ContractOpCodes represents the decompiled version of the contract bytecode.
	ContractOpCodes []opcodes.OpCode `json:"contract_opcodes"`

	// Abi represents the decoded version of the contract bytecode.
	Abi *abis.Decoder `json:"contract_abi"`

	// ContractSourceCode represents the source code of the contract.
	ContractSourceCode string `json:"contract_source_code"`
}
