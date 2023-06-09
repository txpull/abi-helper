package fixtures

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/txpull/unpack/helpers"
)

func ReadBlockFixtures() ([]*types.Block, error) {
	var toReturn []*types.Block
	if err := helpers.ReadGob("../tests/fixtures/blocks.gob", &toReturn); err != nil {
		return nil, err
	}
	return toReturn, nil
}

func ReadTransactionFixtures() ([]*types.Transaction, error) {
	var toReturn []*types.Transaction
	if err := helpers.ReadGob("../tests/fixtures/transactions.gob", &toReturn); err != nil {
		return nil, err
	}
	return toReturn, nil
}

func ReadReceiptFixtures() (map[common.Hash]*types.Receipt, error) {
	var toReturn map[common.Hash]*types.Receipt
	if err := helpers.ReadGob("../tests/fixtures/receipts.gob", &toReturn); err != nil {
		return nil, err
	}
	return toReturn, nil
}
