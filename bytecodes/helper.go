package bytecodes

import (
	"context"
	"github/txpull/abi-helper/clients"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

func GetBytecode(ctx context.Context, client *clients.EthClient, addr common.Address, blockNumber *big.Int) ([]byte, error) {
	return client.GetClient().CodeAt(ctx, addr, blockNumber) // nil represents the latest block
}
