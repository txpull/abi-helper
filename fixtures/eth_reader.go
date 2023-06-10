package fixtures

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type EthReader struct {
	ctx          context.Context
	fixturesPath string
	blocks       map[common.Hash]*types.Block
	transactions map[common.Hash]*types.Transaction
	receipts     map[common.Hash]*types.Receipt
}

func (e *EthReader) Discover() error {
	return nil
}

func (e *EthReader) GetBlocks() map[common.Hash]*types.Block {
	return e.blocks
}

func (e *EthReader) GetTransactions() map[common.Hash]*types.Transaction {
	return e.transactions
}

func (e *EthReader) GetReceipts() map[common.Hash]*types.Receipt {
	return e.receipts
}

func NewEthReader(ctx context.Context, fixturesPath string) (*EthReader, error) {
	toReturn := &EthReader{ctx: ctx, fixturesPath: fixturesPath}

	if err := toReturn.Discover(); err != nil {
		return nil, err
	}

	return toReturn, nil
}
