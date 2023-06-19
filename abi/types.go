package abi

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/txpull/bytecode/signatures"
)

type Abi struct {
	*abi.ABI
}

func (a *Abi) GetMethodsAsSignatures() ([]signatures.Signature, error) {
	var toReturn []signatures.Signature

	return toReturn, nil
}
