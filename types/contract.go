package types

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/txpull/sourcify-go"
)

type ContractVerificationType int16

const (
	ContractVerificationTypeNone ContractVerificationType = iota
	ContractVerificationTypeSourcify
	ContractVerificationTypeEtherscan
	ContractVerificationTypeBscscan
)

type Contract struct {
	UUID uuid.UUID `json:"uuid"`

	ChainID *big.Int `json:"chain_id"`

	// Address represents the address of the contract.
	Address common.Address `json:"address"`

	// BlockHash represents the hash of the block where the contract was created.
	BlockHash common.Hash `json:"block_hash"`

	// TransactionHash represents the hash of the transaction that created the contract.
	TransactionHash common.Hash `json:"transaction_hash"`

	Name                 string                   `json:"name"`
	Language             string                   `json:"language"`
	CompilerVersion      string                   `json:"compiler_version"`
	OptimizationUsed     string                   `json:"optimization_used"`
	Runs                 string                   `json:"runs"`
	ConstructorArguments string                   `json:"constructor_arguments"`
	EVMVersion           string                   `json:"evm_version"`
	Library              string                   `json:"library"`
	LicenseType          string                   `json:"license_type"`
	Proxy                string                   `json:"is_proxy"`
	SourceCode           string                   `json:"source_code"`
	ConstructorABI       string                   `json:"constructor_abi"`
	ABI                  string                   `json:"abi"`
	MetaData             string                   `json:"metadata"`
	SourceUrls           []string                 `json:"source_urls"`
	VerificationType     ContractVerificationType `json:"verification_type"`
	VerificationStatus   string                   `json:"verification_status"`
}

func (r *Contract) MarshalBytes() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := enc.Encode(r); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (r *Contract) UnmarshalBytes(data []byte) error {
	buffer := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buffer)
	err := dec.Decode(r)
	if err != nil {
		return err
	}

	return nil
}

// Unmarshal the ABI from JSON
func (r *Contract) UnmarshalABI() (*abi.ABI, error) {
	parsedAbi, err := abi.JSON(strings.NewReader(r.ABI))
	if err != nil {
		return nil, err
	}

	return &parsedAbi, nil
}

func NewContractFromSourcify(chainId *big.Int, address common.Address, metadata *sourcify.Metadata, metadataBytes []byte) (*Contract, error) {
	contract := &Contract{
		UUID:             uuid.New(),
		ChainID:          chainId,
		Address:          address,
		Language:         strings.ToLower(metadata.Language),
		CompilerVersion:  metadata.Compiler.Version,
		OptimizationUsed: strconv.FormatBool(metadata.Settings.Optimizer.Enabled),
		Runs:             strconv.Itoa(metadata.Settings.Optimizer.Runs),
		EVMVersion:       metadata.Settings.EvmVersion,
		MetaData:         string(metadataBytes),
	}

	// Search for fucking license...
	for entryPointTarget, contractName := range metadata.Settings.CompilationTarget {
		contract.Name = contractName

		if len(contract.LicenseType) == 0 {
			if metadata.Sources != nil && len(metadata.Sources) > 0 {
				for sourcesTarget, source := range metadata.Sources {
					if entryPointTarget == sourcesTarget {
						contract.LicenseType = source.License
						contract.SourceUrls = source.Urls
					}
				}

			}
		}

	}

	abiJson, err := json.Marshal(metadata.Output.Abi)
	if err != nil {
		return nil, err
	}
	contract.ABI = string(abiJson)

	return contract, nil
}
