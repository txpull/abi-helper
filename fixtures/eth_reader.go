package fixtures

import (
	"bytes"
	"context"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/txpull/bytecode/utils"
)

// EthReader is a structure that provides methods for discovering Ethereum blocks, transactions, and receipts.
type EthReader struct {
	ctx          context.Context // Context for managing the lifecycle of the EthReader.
	fixturesPath string          // Path to the directory containing fixture files.
	blocks       map[common.Hash]*types.Block
	transactions map[common.Hash]*types.Transaction
	receipts     map[common.Hash]*types.Receipt
}

// Discover populates the EthReader with blocks, transactions, and receipts by reading from the fixture files.
func (e *EthReader) Discover() error {
	if err := e.discoverBlocks(); err != nil {
		return err
	}

	if err := e.discoverTransactions(); err != nil {
		return err
	}

	if err := e.discoverReceipts(); err != nil {
		return err
	}

	return nil
}

// GetBlocks returns the map of blocks stored in the EthReader.
func (e *EthReader) GetBlocks() map[common.Hash]*types.Block {
	return e.blocks
}

// GetTransactions returns the map of transactions stored in the EthReader.
func (e *EthReader) GetTransactions() map[common.Hash]*types.Transaction {
	return e.transactions
}

// GetReceipts returns the map of receipts stored in the EthReader.
func (e *EthReader) GetReceipts() map[common.Hash]*types.Receipt {
	return e.receipts
}

// GetReceiptFromTxHash retrieves the receipt associated with the given transaction hash from the EthReader.
// It returns the receipt and a boolean indicating if the receipt was found or not.
func (e *EthReader) GetReceiptFromTxHash(txhash common.Hash) (*types.Receipt, bool) {
	receipt, ok := e.receipts[txhash]
	return receipt, ok
}

// discoverBlocks reads the blocks from the fixture file and populates the blocks map in the EthReader.
func (e *EthReader) discoverBlocks() error {
	var blocksRlp [][]byte

	if err := utils.ReadGob(filepath.Join(e.fixturesPath, "blocks.gob"), &blocksRlp); err != nil {
		return err
	}

	for _, blkRlp := range blocksRlp {
		var blk *types.Block
		if err := rlp.DecodeBytes(blkRlp, &blk); err != nil {
			return err
		}
		e.blocks[blk.Hash()] = blk
	}

	return nil
}

// discoverTransactions reads the transactions from the fixture file and populates the transactions map in the EthReader.
func (e *EthReader) discoverTransactions() error {
	var txnsRlp map[common.Hash][]byte

	if err := utils.ReadGob(filepath.Join(e.fixturesPath, "transactions.gob"), &txnsRlp); err != nil {
		return err
	}

	for _, txnRlp := range txnsRlp {
		var tx *types.Transaction
		if err := rlp.Decode(bytes.NewReader(txnRlp), &tx); err != nil {
			return err
		}
		e.transactions[tx.Hash()] = tx
	}

	return nil
}

// discoverReceipts reads the receipts from the fixture file and populates the receipts map in the EthReader.
func (e *EthReader) discoverReceipts() error {
	var receiptsRlp map[common.Hash][]byte

	if err := utils.ReadGob(filepath.Join(e.fixturesPath, "receipts.gob"), &receiptsRlp); err != nil {
		return err
	}

	for _, receiptRlp := range receiptsRlp {
		r := &types.Receipt{}
		if err := r.UnmarshalJSON(receiptRlp); err != nil {
			return err
		}
		e.receipts[r.TxHash] = r
	}

	return nil
}

// NewEthReader creates a new EthReader instance with the provided context and fixtures path.
func NewEthReader(ctx context.Context, fixturesPath string) (*EthReader, error) {
	ethReader := &EthReader{
		ctx:          ctx,
		fixturesPath: fixturesPath,
		blocks:       make(map[common.Hash]*types.Block),
		transactions: make(map[common.Hash]*types.Transaction),
		receipts:     make(map[common.Hash]*types.Receipt),
	}

	if err := ethReader.Discover(); err != nil {
		return nil, err
	}

	return ethReader, nil
}
