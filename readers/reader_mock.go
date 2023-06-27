package readers

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/txpull/unpack/types"
)

// MockReader implements the Reader interface for testing purposes.
type MockReader struct{}

func (r *MockReader) GetContractByAddress(chainId *big.Int, address common.Address) (*types.Contract, error) {
	// Mock implementation
	return nil, nil
}

func (r *MockReader) GetMethodBySignature(chainId *big.Int, signature string) (*types.Method, error) {
	// Mock implementation
	return nil, nil
}

func (r *MockReader) GetEventByHash(chainId *big.Int, hash common.Hash) (*types.Event, error) {
	// Mock implementation
	return nil, nil
}

func (r *MockReader) String() string {
	return "mock"
}
