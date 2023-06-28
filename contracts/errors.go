package contracts

import "errors"

// ErrMissingBlock, ErrMissingTransaction, and ErrMissingReceipt are error types
// which are returned when trying to process a contract creation transaction without
// providing the required block, transaction, or receipt, respectively.
var (
	// ErrMissingBlock is an error that occurs when a block is not provided
	// while trying to decode a contract creation transaction.
	ErrMissingBlock = errors.New("you need to provide block in order to decode contract creation tx")

	// ErrMissingTransaction is an error that occurs when a transaction is not provided
	// while trying to decode a contract creation transaction.
	ErrMissingTransaction = errors.New("you need to provide transaction in order to decode contract creation tx")

	// ErrMissingReceipt is an error that occurs when a transaction receipt is not provided
	// while trying to decode a contract creation transaction.
	ErrMissingReceipt = errors.New("you need to provide transaction receipt in order to decode contract creation tx")

	// ErrFailedGetTransactionByHash is an error that occurs when a transaction
	// cannot be retrieved from the blockchain.
	ErrFailedGetTransactionByHash = errors.New("failed to get transaction by hash")

	// ErrFailedGetTransactionReceiptByHash is an error that occurs when a transaction receipt
	// cannot be retrieved from the blockchain.
	ErrFailedGetTransactionReceiptByHash = errors.New("failed to get transaction receipt by hash")
)
