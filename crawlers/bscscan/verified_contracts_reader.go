package bscscan

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/txpull/bytecode/scanners"
	"github.com/txpull/bytecode/utils"
	"go.uber.org/zap"
)

// VerifiedContractsReader encapsulates the context, the path to the .gob file,
// and a map of contracts data.
type VerifiedContractsReader struct {
	ctx      context.Context
	filePath string
	data     map[string]*scanners.Result
}

// NewVerifiedContractsReader creates and returns a new VerifiedContractsReader instance.
//
// Example:
//
//	reader := bscscan.NewVerifiedContractsReader(ctx, "path/to/output.gob")
func NewVerifiedContractsReader(ctx context.Context, filePath string) *VerifiedContractsReader {
	return &VerifiedContractsReader{
		ctx:      ctx,
		filePath: filePath,
		data:     make(map[string]*scanners.Result),
	}
}

// Read loads data from the .gob file into the data map of the reader.
// Each item in the map represents a verified contract.
//
// Example:
//
//	err := reader.Read()
//	if err != nil {
//	    log.Fatal(err)
//	}
func (r *VerifiedContractsReader) Read() error {
	var results map[string][]byte
	if err := utils.ReadGob(r.filePath, &results); err != nil {
		zap.L().Error(
			"failed to read data from gob file",
			zap.String("file_path", r.filePath),
			zap.Error(err),
		)
		return err
	}

	for contractAddress, res := range results {
		contract := &scanners.Result{}

		if err := contract.UnmarshalBytes(res); err != nil {
			zap.L().Error(
				"failed to unmarshal contract data",
				zap.Error(err),
			)
			return err
		}

		// Add contract to the data map, using contract address as the key
		r.data[contractAddress] = contract
	}

	return nil
}

// GetContractByAddress retrieves a contract by its address from the data map.
//
// Example:
//
//	contract, ok := reader.GetContractByAddress("address")
//	if !ok {
//	    fmt.Println("Contract not found")
//	}
func (r *VerifiedContractsReader) GetContractByAddress(address string) (*scanners.Result, bool) {
	contract, ok := r.data[address]
	return contract, ok
}

// GetContracts retrieves all contracts from the data map.
//
// Example:
//
//	contracts := reader.GetContracts()
//	for address, contract := range contracts {
//	    fmt.Println("Address:", address)
//	    fmt.Println("Contract Name:", contract.Name)
//	    fmt.Println()
//	}
func (r *VerifiedContractsReader) GetContracts() map[string]*scanners.Result {
	return r.data
}

// GetAbiByContractAddress retrieves a contract ABI by its address from the data map.
// It returns the ABI if found, otherwise it returns an error.
//
// Example:
//
//	address := "0x1234567890abcdef"
//	reader := NewVerifiedContractsReader("path/to/contracts.gob")
//	err := reader.Read()
//	if err != nil {
//		log.Fatal("Failed to read contracts:", err)
//	}
//	abi, err := reader.GetAbiByContractAddress(address)
//	if err != nil {
//		log.Fatal("Failed to get ABI:", err)
//	}
//	fmt.Println("Contract ABI:", abi)
func (r *VerifiedContractsReader) GetAbiByContractAddress(address string) (*abi.ABI, error) {
	contract, ok := r.data[address]
	if !ok {
		return nil, fmt.Errorf("contract not found for address: %s", address)
	}
	return contract.UnmarshalABI()
}
