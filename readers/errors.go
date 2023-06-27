package readers

import "errors"

var (
	// ErrReaderNotFound is returned when a reader is not found
	ErrReaderNotFound = errors.New("reader not found")

	// ErrPriorityReaderNotSet is returned when the priority reader is not set
	ErrPriorityReaderNotSet = errors.New("priority reader not set")

	// ErrRecordNotFound is returned when a record is not found
	ErrRecordNotFound = errors.New("record not found")
)
