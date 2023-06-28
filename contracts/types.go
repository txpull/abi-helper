package contracts

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/txpull/unpack/abis"
	"github.com/txpull/unpack/opcodes"
	"github.com/txpull/unpack/solidity"
	"github.com/txpull/unpack/types"
)

var ContractProcessStatus uint8

const (
	ContractProcessStatusInit uint8 = iota
	ContractProcessStatusPending
	ContractProcessStatusSuccess
	ContractProcessStatusFailed
)

func ContractProcessStatusToString(status uint8) string {
	switch status {
	case ContractProcessStatusInit:
		return "init"
	case ContractProcessStatusPending:
		return "pending"
	case ContractProcessStatusSuccess:
		return "success"
	case ContractProcessStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

type ContractResponse struct {
	// ChainId represents the chain id of the contract.
	ChainId *big.Int `json:"chain_id"`

	// BlockNumber represents the block number where the contract was created.
	BlockNumber *big.Int `json:"block_number"`

	// BlockHash represents the hash of the block where the contract was created.
	BlockHash common.Hash `json:"block_hash"`

	// TransactionHash represents the hash of the contract creation transaction.
	TransactionHash common.Hash `json:"transaction_hash"`

	// ReceiptStatus represents the status of the transaction receipt.
	ReceiptStatus uint64 `json:"receipt_status"`

	// Address represents the address of the created contract.
	Address common.Address `json:"contract_address"`

	// CreationBytecode represents the bytecode of the contract at the time of creation.
	CreationBytecode []byte `json:"runtime_bytecode"`

	// CreationBytecodeSize represents the size of the runtime bytecode at the time of creation.
	CreationBytecodeSize int `json:"runtime_bytecode_size"`

	// CreationOpCodes represents the decompiled version of the bytecode at the time of creation.
	CreationOpCodes []opcodes.OpCode `json:"runtime_opcodes"`

	// ContractBytecode represents the bytecode of the contract.
	ContractBytecode []byte `json:"contract_bytecode"`

	// ContractBytecodeSize represents the size of the contract bytecode.
	ContractBytecodeSize int `json:"contract_bytecode_size"`

	// Name represents the name of the contract.
	Name string `json:"name"`

	// Language represents the language of the contract.
	Language types.ContractLanguage `json:"language"`

	// LicenseType represents the license type of the contract.
	LicenseType string `json:"license_type"`

	// CompilerVersion represents the version of the compiler used to compile the contract.
	CompilerVersion string `json:"compiler_version"`

	// OptimizationUsed represents whether the contract was compiled with optimization.
	OptimizationUsed bool `json:"optimization_used"`

	// Runs represents the number of runs used to compile the contract.
	Runs int `json:"runs"`

	// ConstructorArguments represents the arguments used to create the contract.
	ConstructorArguments string `json:"constructor_arguments"`

	// ConstructorArgumentsSize represents the size of the constructor arguments.
	ConstructorArgumentsSize int `json:"constructor_arguments_size"`

	// Library represents the library used to compile the contract.
	Library string `json:"library"`

	// ProxyDetected represents whether the contract is a proxy contract.
	ProxyDetected bool `json:"proxy_detected"`

	// Implementation represents the implementation address of the proxy contract.
	Implementation common.Address `json:"implementation"`

	// SourceUrls represents the source urls of the contract from swarm,ipfs.
	SourceUrls []string `json:"source_urls"`

	// Abi represents the decoded version of the contract bytecode.
	Abi *abis.Decoder `json:"contract_abi"`

	// ContractSourceCode represents the source code of the contract.
	SourceCode *solidity.Decoder `json:"source_code"`

	// VerificationType represents the type of verification used to verify the contract.
	VerificationType types.ContractVerificationType `json:"verification_type"`

	// VerificationStatus represents the status of the verification.
	VerificationStatus string `json:"verification_status"`

	// ProcessStatus represents the status of the contract process by unpack.
	ProcessStatus uint8 `json:"process_status"`
}

func (c *ContractResponse) SetLanguageFromCompilerVersion(version string) {
	if strings.Contains(version, "vyper") {
		c.Language = types.ContractLanguageVyper
	} else {
		c.Language = types.ContractLanguageSolidity
	}
}
