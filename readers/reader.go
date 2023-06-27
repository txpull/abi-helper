// Package readers provides a manager for handling multiple Reader instances.
package readers

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/txpull/unpack/types"
)

type Reader interface {
	GetContractByAddress(chainId *big.Int, address common.Address) (*types.Contract, error)

	GetMethodBySignature(chainId *big.Int, signature string) (*types.Method, error)

	GetEventByHash(chainId *big.Int, hash common.Hash) (*types.Event, error)

	// String returns the name of the Reader.
	String() string
}
