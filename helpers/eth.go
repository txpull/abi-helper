package helpers

import (
	"context"
	"math/big"

	"github.com/txpull/unpack/clients"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// GetBytecode retrieves the bytecode of a contract at a specific address and block number.
//
// It takes a context.Context object (`ctx`), a clients.EthClient object (`client`),
// a common.Address object (`addr`) representing the contract address,
// and a *big.Int object (`blockNumber`) representing the block number.
//
// If `blockNumber` is nil, the function retrieves the bytecode at the latest block.
//
// It returns the bytecode as a byte slice and an error if the retrieval fails.
func GetBytecode(ctx context.Context, client *clients.EthClient, addr common.Address, blockNumber *big.Int) ([]byte, error) {
	return client.GetClient().CodeAt(ctx, addr, blockNumber)
}

func GetTransactionByHash(ctx context.Context, client *clients.EthClient, hash common.Hash) (*types.Transaction, bool, error) {
	return client.GetClient().TransactionByHash(ctx, hash)
}

func GetReceiptByHash(ctx context.Context, client *clients.EthClient, hash common.Hash) (*types.Receipt, error) {
	return client.GetClient().TransactionReceipt(ctx, hash)
}
