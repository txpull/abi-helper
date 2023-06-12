// Package signatures provides functionality for working with signature data.
package signatures

type SignatureNotFound struct {
	HexSignature string
}

func (e SignatureNotFound) Error() string {
	return "Signature not found: " + e.HexSignature
}
