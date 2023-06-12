package transactions

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyMethodId          = errors.New("empty method id while discovering signature")
	ErrArgSignatureIsRequired = errors.New("signature is required to decode arguments")
)

// SignatureNotFound is a custom error indicating that a signature was not found.
type ErrSignatureNotFound struct {
	Hex string
}

// Error returns the error message for the SignatureNotFound error.
func (e ErrSignatureNotFound) Error() string {
	return fmt.Sprintf("signature not found: %s", e.Hex)
}
