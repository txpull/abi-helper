// Package signatures provides functionality for working with signature data.
package signatures

import (
	"context"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

// Signatures represents a collection of signatures and provides methods for interacting with them.
type SignaturesReader struct {
	db *badger.DB
}

// Close closes the underlying database connection.
func (s *SignaturesReader) Close() error {
	return s.db.Close()
}

// LookupByHex retrieves a signature based on the provided hex value.
// It returns a pointer to a Signature struct if found, a boolean indicating if the signature is found, and an error if the signature is not found or an error occurs.
func (s *SignaturesReader) LookupByHex(hexSignature string) (*Signature, bool, error) {
	var signature *Signature

	// Make sure we don't look it by a hex string with a leading "0x".
	hexSignature = strings.TrimLeft(hexSignature, "0x")

	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(hexSignature))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil // Key not found means the signature does not exist.
			}
			return err
		}

		value, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		if err := signature.UnmarshalBytes(value); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, false, err
	}

	if signature == nil {
		return nil, false, nil
	}

	return signature, true, nil
}

// NewDb opens a connection to the BadgerDB specified by dbLocation and returns a new instance of SignaturesReader.
// It verifies the connection by performing a database operation and returns an error if the connection fails.
func NewDb(ctx context.Context, dbLocation string) (*SignaturesReader, error) {
	// Open the BadgerDB located in the dbLocation directory.
	// It will be created if it doesn't exist.
	db, err := badger.Open(badger.DefaultOptions(dbLocation))
	if err != nil {
		return nil, err
	}

	// Create a new instance of SignaturesReader struct.
	signatures := &SignaturesReader{
		db: db,
	}

	// Perform a database operation to verify the connection.
	err = db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("ping"))
		return err
	})

	if err != nil {
		db.Close()
		return nil, err
	}

	return signatures, nil
}
