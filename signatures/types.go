// Package signatures provides functionality for working with signature data.
package signatures

import (
	"time"
)

type Signature struct {
	ID            int
	TextSignature string
	HexSignature  string
	Bytes         []byte
	CreatedAt     time.Time
}
