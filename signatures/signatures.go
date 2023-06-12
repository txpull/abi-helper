// Package signatures provides functionality for working with signature data.
package signatures

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	_ "github.com/mattn/go-sqlite3" // Import SQLite3 driver
)

// Signatures represents a collection of signatures and provides methods for interacting with them.
type Signatures struct {
	db *sql.DB
}

// Close closes the underlying database connection.
func (s *Signatures) Close() error {
	return s.db.Close()
}

// LookupByHex retrieves a signature based on the provided hex value.
// It returns a pointer to a Signature struct if found, bool as signature found or not and/or an error if the signature is not found or an error occurs.
func (s *Signatures) LookupByHex(hexSignature string) (*Signature, bool, error) {
	if !strings.HasPrefix(hexSignature, "0x") {
		hexSignature = "0x" + hexSignature
	}

	query := "SELECT id, text_signature, hex_signature, bytes, created_at FROM signatures WHERE hex_signature = ?"
	row := s.db.QueryRow(query, hexSignature)

	signature := &Signature{}
	err := row.Scan(
		&signature.ID,
		&signature.TextSignature,
		&signature.HexSignature,
		&signature.Bytes,
		&signature.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, false, SignatureNotFound{HexSignature: hexSignature}
	}

	if err != nil {
		return nil, false, err
	}

	return signature, true, nil
}

// NewDb opens a connection to the SQLite3 database specified by dbLocation and returns a new instance of Signatures.
// It verifies the connection by pinging the database and returns an error if the connection fails.
func NewDb(ctx context.Context, dbLocation string) (*Signatures, error) {
	// Load the SQLite3 database file
	db, err := sql.Open("sqlite3", dbLocation)
	if err != nil {
		return nil, err
	}

	// Ping the database to ensure connection
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Create a new instance of Signatures struct
	signatures := &Signatures{
		db: db,
	}

	return signatures, nil
}
