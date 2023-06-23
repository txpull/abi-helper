// Package signatures provides functionality for working with signature data.
//
// It offers an interface to retrieve signatures from a BadgerDB database.
// The main entry point to the package is the SignaturesReader struct.
package signatures

import (
	"strings"

	"github.com/txpull/unpack/db"
)

// SignaturesReader represents a collection of signatures and provides methods for interacting with them.
// It leverages the BadgerDB for storing and retrieving signature data.
type SignaturesReader struct {
	db *db.BadgerDB
}

// Option is a function type that applies configurations to SignaturesReader
type Option func(*SignaturesReader)

// WithBadgerDB is an option to provide an existing BadgerDB instance to the SignaturesReader
func WithBadgerDB(bdb *db.BadgerDB) Option {
	return func(s *SignaturesReader) {
		s.db = bdb
	}
}

// Close cleanly shuts down the underlying BadgerDB database connection.
func (s *SignaturesReader) Close() error {
	return s.db.Close()
}

// LookupByHex retrieves a signature based on the provided hex value.
// It returns a pointer to a Signature struct if found, a boolean indicating if the signature was found,
// and an error if there's an issue with the database operation.
func (s *SignaturesReader) LookupByHex(hexSignature string) (*Signature, bool, error) {
	// Make sure we don't look it by a hex string with a leading "0x".
	hexSignature = strings.TrimLeft(hexSignature, "0x")

	// Use the db.Get() method
	value, err := s.db.Get(hexSignature)
	if err != nil {
		return nil, false, err
	}

	// Unmarshal bytes to signature
	var signature *Signature
	if err := signature.UnmarshalBytes(value); err != nil {
		return nil, false, err
	}

	return signature, true, nil
}

// NewDb returns a new instance of SignaturesReader.
func NewDb(opts ...Option) *SignaturesReader {
	signatures := &SignaturesReader{}

	// Apply all options to signatures
	for _, opt := range opts {
		opt(signatures)
	}

	return signatures
}
