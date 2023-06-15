package bscscan

import (
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/txpull/bytecode/scanners"
)

// TestVerifiedContractsReader_Read tests the Read method of VerifiedContractsReader.
func TestVerifiedContractsReader_Read(t *testing.T) {
	// Prepare a temporary .gob file for testing
	filePath := "/tmp/contracts.gob"
	createTestGobFile(t, filePath)

	ctx := context.Background()
	reader := NewVerifiedContractsReader(ctx, filePath)
	err := reader.Read()
	assert.NoError(t, err, "Read should not return an error")

	// Verify the expected contracts are loaded
	// Define the expected contracts
	expectedContracts := map[string]*scanners.Result{
		"0x1234567890abcdef": {
			Name:                 "Contract1",
			CompilerVersion:      "v0.8.7",
			OptimizationUsed:     "Yes",
			Runs:                 "200",
			ConstructorArguments: "arg1, arg2",
			EVMVersion:           "petersburg",
			Library:              "Lib1",
			LicenseType:          "MIT",
			Proxy:                "Proxy1",
			Implementation:       "Impl1",
			SourceCode:           "contract Contract1 { function test() public { } }",
			ABI:                  `[{"constant":false,"inputs":[],"name":"test","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`,
			SwarmSource:          "swarm-hash-1",
		},
		"0xabcdef1234567890": {
			Name:                 "Contract2",
			CompilerVersion:      "v0.7.3",
			OptimizationUsed:     "No",
			Runs:                 "100",
			ConstructorArguments: "",
			EVMVersion:           "byzantium",
			Library:              "",
			LicenseType:          "Apache",
			Proxy:                "Proxy2",
			Implementation:       "Impl2",
			SourceCode:           "contract Contract2 { function test() public { } }",
			ABI:                  `[{"constant":false,"inputs":[],"name":"test","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`,
			SwarmSource:          "swarm-hash-2",
		},
	}

	// Get the contracts from the reader
	contracts := reader.GetContracts()

	// Check the number of contracts
	assert.Equal(t, len(expectedContracts), len(contracts), "Number of contracts mismatch")

	// Check each contract
	for address, expectedContract := range expectedContracts {
		contract, ok := contracts[address]
		assert.True(t, ok, fmt.Sprintf("Contract not found for address: %s", address))
		assert.Equal(t, expectedContract, contract, fmt.Sprintf("Contract mismatch for address: %s", address))
	}
}

// TestVerifiedContractsReader_GetContractByAddress tests the GetContractByAddress method of VerifiedContractsReader.
func TestVerifiedContractsReader_GetContractByAddress(t *testing.T) {
	// Prepare a temporary .gob file for testing
	filePath := "/tmp/contracts.gob"
	createTestGobFile(t, filePath)

	ctx := context.Background()
	reader := NewVerifiedContractsReader(ctx, filePath)
	err := reader.Read()
	assert.NoError(t, err, "Read should not return an error")

	// Test retrieving an existing contract
	address := "0x1234567890abcdef"
	contract, ok := reader.GetContractByAddress(address)
	assert.True(t, ok, "Contract should exist")
	assert.Equal(t, "Contract1", contract.Name, "Contract name should match")

	// Test retrieving a non-existing contract
	nonExistingAddress := "0xfff"
	_, ok = reader.GetContractByAddress(nonExistingAddress)
	assert.False(t, ok, "Contract should not exist")
}

// TestVerifiedContractsReader_GetContracts tests the GetContracts method of VerifiedContractsReader.
func TestVerifiedContractsReader_GetContracts(t *testing.T) {
	// Prepare a temporary .gob file for testing
	filePath := "/tmp/contracts.gob"
	createTestGobFile(t, filePath)

	ctx := context.Background()
	reader := NewVerifiedContractsReader(ctx, filePath)
	err := reader.Read()
	assert.NoError(t, err, "Read should not return an error")

	contracts := reader.GetContracts()
	assert.Len(t, contracts, 2, "Number of contracts should match")

	// Verify each contract in the result
	for address, contract := range contracts {
		t.Run(fmt.Sprintf("Contract at Address %s", address), func(t *testing.T) {
			assert.NotEmpty(t, contract.Name, "Contract name should not be empty")
			assert.NotEmpty(t, contract.CompilerVersion, "Compiler version should not be empty")
			assert.NotEmpty(t, contract.OptimizationUsed, "Optimization used should not be empty")
			assert.NotEmpty(t, contract.Runs, "Runs should not be empty")
			// Additional assertions for each contract's properties if needed
		})
	}
}

// TestVerifiedContractsReader_GetAbiByContractAddress tests the GetAbiByContractAddress method of VerifiedContractsReader.
func TestVerifiedContractsReader_GetAbiByContractAddress(t *testing.T) {
	// Prepare a temporary .gob file for testing
	filePath := "/tmp/contracts.gob"
	createTestGobFile(t, filePath)
	ctx := context.Background()
	reader := NewVerifiedContractsReader(ctx, filePath)
	err := reader.Read()
	assert.NoError(t, err, "Read should not return an error")

	// Test retrieving the ABI of an existing contract
	address := "0x1234567890abcdef"
	abi, err := reader.GetAbiByContractAddress(address)
	assert.NoError(t, err, "GetAbiByContractAddress should not return an error")
	assert.NotNil(t, abi, "ABI should not be nil")

	// Test retrieving the ABI of a non-existing contract
	nonExistingAddress := "0xfdfff"
	_, err = reader.GetAbiByContractAddress(nonExistingAddress)
	assert.Error(t, err, "GetAbiByContractAddress should return an error")
	assert.EqualError(t, err, fmt.Sprintf("contract not found for address: %s", nonExistingAddress))
}

// Helper function to create a temporary .gob file for testing
func createTestGobFile(t *testing.T, filePath string) {
	contracts := map[string][]byte{
		"0x1234567890abcdef": marshalScannerResult(t, &scanners.Result{
			Name:                 "Contract1",
			CompilerVersion:      "v0.8.7",
			OptimizationUsed:     "Yes",
			Runs:                 "200",
			ConstructorArguments: "arg1, arg2",
			EVMVersion:           "petersburg",
			Library:              "Lib1",
			LicenseType:          "MIT",
			Proxy:                "Proxy1",
			Implementation:       "Impl1",
			SourceCode:           "contract Contract1 { function test() public { } }",
			ABI:                  `[{"constant":false,"inputs":[],"name":"test","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`,
			SwarmSource:          "swarm-hash-1",
		}),
		"0xabcdef1234567890": marshalScannerResult(t, &scanners.Result{
			Name:                 "Contract2",
			CompilerVersion:      "v0.7.3",
			OptimizationUsed:     "No",
			Runs:                 "100",
			ConstructorArguments: "",
			EVMVersion:           "byzantium",
			Library:              "",
			LicenseType:          "Apache",
			Proxy:                "Proxy2",
			Implementation:       "Impl2",
			SourceCode:           "contract Contract2 { function test() public { } }",
			ABI:                  `[{"constant":false,"inputs":[],"name":"test","outputs":[],"payable":false,"stateMutability":"nonpayable","type":"function"}]`,
			SwarmSource:          "swarm-hash-2",
		}),
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	assert.NoError(t, err, "Failed to open or create test gob file")
	defer file.Close()

	encoder := gob.NewEncoder(file)
	err = encoder.Encode(contracts)
	assert.NoError(t, err, "Failed to encode test contracts data")
}

func marshalScannerResult(t *testing.T, result *scanners.Result) []byte {
	binary, err := result.MarshalBytes()
	assert.NoError(t, err, "Failed to marshal binary")

	// Assert the marshaled binary
	assert.NotNil(t, binary, "Marshaled binary is nil")

	return binary
}
