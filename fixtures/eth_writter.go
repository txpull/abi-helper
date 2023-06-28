package fixtures

import (
	"context"
	"math/big"
	"os"
	"path/filepath"

	"github.com/txpull/unpack/clients"
	"github.com/txpull/unpack/helpers"
	"github.com/txpull/unpack/options"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"go.uber.org/zap"
)

// EthWriter is a structure that encapsulates the context, options, and clients needed to generate Ethereum fixtures.
// It also maintains the blocks, transactions, and receipts that are generated.
type EthWriter struct {
	ctx          context.Context
	opts         options.Fixtures
	clients      *clients.EthClient
	blocks       [][]byte
	transactions map[common.Hash][]byte
	receipts     map[common.Hash][]byte
}

// Generate is a method that generates Ethereum fixtures.
// It retrieves blocks within the specified range from the Ethereum blockchain, encodes them into RLP format, and stores them.
// It also retrieves, encodes, and stores the transactions and receipts associated with these blocks.
// If any error occurs during this process, it is logged and returned.
func (e *EthWriter) Generate() error {
	// Clean up previously generated data
	e.blocks = [][]byte{}
	e.transactions = make(map[common.Hash][]byte)
	e.receipts = make(map[common.Hash][]byte)

	for blockNumber := e.opts.StartBlockNumber; blockNumber <= e.opts.EndBlockNumber; blockNumber++ {
		// Retrieve the block by number
		block, err := e.clients.GetClient(big.NewInt(56)).BlockByNumber(e.ctx, big.NewInt(int64(blockNumber)))
		if err != nil {
			zap.L().Error(
				"failed to retrieve block",
				zap.Uint64("block_number", blockNumber),
				zap.Error(err),
			)
			return err
		}

		// Encode the block into RLP format
		blockBytes, err := rlp.EncodeToBytes(block)
		if err != nil {
			zap.L().Error(
				"failed to RLP encode block",
				zap.Uint64("block_number", blockNumber),
				zap.Error(err),
			)
			return err
		}
		e.blocks = append(e.blocks, blockBytes)

		for _, tx := range block.Transactions() {
			// Retrieve the transaction receipt
			receipt, err := e.clients.GetClient(big.NewInt(56)).TransactionReceipt(e.ctx, tx.Hash())
			if err != nil {
				zap.L().Error(
					"failed to retrieve transaction receipt",
					zap.Uint64("block_number", blockNumber),
					zap.String("tx_hash", tx.Hash().Hex()),
					zap.Error(err),
				)
				continue
			}

			// Encode the transaction into RLP format
			txBytes, err := rlp.EncodeToBytes(tx)
			if err != nil {
				zap.L().Error(
					"failed to RLP encode transaction",
					zap.Uint64("block_number", blockNumber),
					zap.String("tx_hash", tx.Hash().Hex()),
					zap.Error(err),
				)
				return err
			}
			e.transactions[tx.Hash()] = txBytes

			receiptBytes, err := receipt.MarshalJSON()
			if err != nil {
				zap.L().Error(
					"failed to RLP encode transaction receipt",
					zap.Uint64("block_number", blockNumber),
					zap.String("tx_hash", tx.Hash().Hex()),
					zap.Error(err),
				)
				return err
			}
			e.receipts[tx.Hash()] = receiptBytes
		}
		zap.L().Info("Successfully generated block", zap.Int64("number", block.Number().Int64()))
	}
	return nil
}

// Write is a method that writes the generated Ethereum fixtures to files.
// It writes the encoded blocks, transactions, and receipts to separate files.
// If any error occurs during this process, it is logged and returned.
func (e *EthWriter) Write() error {
	blocksPath := filepath.Join(e.opts.FixturesPath, "blocks.gob")
	if err := removeFileIfExists(blocksPath); err != nil {
		return err
	}

	if err := helpers.WriteGob(blocksPath, e.blocks); err != nil {
		zap.L().Error(
			"failed to write RLP encoded blocks",
			zap.Error(err),
		)
		return err
	}

	txPath := filepath.Join(e.opts.FixturesPath, "transactions.gob")
	if err := removeFileIfExists(txPath); err != nil {
		return err
	}

	if err := helpers.WriteGob(txPath, e.transactions); err != nil {
		zap.L().Error(
			"failed to write RLP encoded block transactions",
			zap.Error(err),
		)
		return err
	}

	receiptPath := filepath.Join(e.opts.FixturesPath, "receipts.gob")
	if err := removeFileIfExists(receiptPath); err != nil {
		return err
	}

	if err := helpers.WriteGob(receiptPath, e.receipts); err != nil {
		zap.L().Error(
			"failed to write RLP encoded transaction receipts",
			zap.Error(err),
		)
		return err
	}

	zap.L().Info("Successfully wrote fixtures")

	return nil
}

// removeFileIfExists is a helper function that removes the file at the given path if it exists.
func removeFileIfExists(path string) error {
	if _, err := os.Stat(path); err == nil {
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}

// NewEthWriter is a function that creates a new instance of EthWriter.
// It takes a context and options as parameters, creates a new Ethereum client, and returns an EthWriter that uses this client.
// If any error occurs during the creation of the Ethereum client, it is returned.
func NewEthWriter(ctx context.Context, opts options.Fixtures) (*EthWriter, error) {
	generator := &EthWriter{
		ctx:          ctx,
		opts:         opts,
		transactions: make(map[common.Hash][]byte),
		receipts:     make(map[common.Hash][]byte),
	}

	clients, err := clients.NewEthClient(ctx, options.G().GetNode(opts.Network, opts.NodeType))
	if err != nil {
		return nil, err
	}
	generator.clients = clients

	return generator, nil
}
