package abi

import "errors"

var (
	// ErrAbiNotFound is returned when the ABI is not found in the database.
	ErrAbiNotFound = errors.New("abi not found")

	// ErrNoVerifiedContractsReader is returned when the verified contracts reader is not set.
	ErrNoVerifiedContractsReader = errors.New("verified contracts reader is not set.")
)
