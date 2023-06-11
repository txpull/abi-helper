package fixtures_test

import (
	"testing"

	"github.com/txpull/abi-helper/fixtures"

	"github.com/stretchr/testify/assert"
)

func TestEthReader_Discover(t *testing.T) {
	tAssert := assert.New(t)

	ethReader := fixtures.GetEthReaderForTest(tAssert)

	err := ethReader.Discover()
	tAssert.NoError(err, "discover returned an error")
}

func TestEthReader_GetBlocks(t *testing.T) {
	tAssert := assert.New(t)

	ethReader := fixtures.GetEthReaderForTest(tAssert)

	err := ethReader.Discover()
	tAssert.NoError(err, "discover returned an error")

	blocks := ethReader.GetBlocks()
	tAssert.GreaterOrEqual(len(blocks), 1)
}

func TestEthReader_GetTransactions(t *testing.T) {
	tAssert := assert.New(t)

	ethReader := fixtures.GetEthReaderForTest(tAssert)

	err := ethReader.Discover()
	tAssert.NoError(err, "discover returned an error")

	blocks := ethReader.GetBlocks()
	tAssert.GreaterOrEqual(len(blocks), 1)
}

func TestEthReader_GetReceipts(t *testing.T) {
	tAssert := assert.New(t)

	ethReader := fixtures.GetEthReaderForTest(tAssert)

	err := ethReader.Discover()
	tAssert.NoError(err, "discover returned an error")

	txns := ethReader.GetTransactions()
	receipts := ethReader.GetReceipts()
	tAssert.Equal(len(txns), len(receipts))
}

func TestEthReader_GetReceiptFromTxHash(t *testing.T) {
	tAssert := assert.New(t)

	ethReader := fixtures.GetEthReaderForTest(tAssert)

	err := ethReader.Discover()
	tAssert.NoError(err, "discover returned an error")

	for _, txn := range ethReader.GetTransactions() {
		receipt, found := ethReader.GetReceiptFromTxHash(txn.Hash())
		tAssert.True(found)
		tAssert.Equal(txn.Hash().Hex(), receipt.TxHash.Hex())
	}
}
