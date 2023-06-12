package fixtures

import (
	"context"
	"math/big"
	"os"
	"path/filepath"

	"github.com/txpull/bytecode/clients"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"go.uber.org/zap"
)

// EthGeneratorConfig holds the configuration for EthGenerator.
type EthGeneratorConfig struct {
	ClientURL               string // URL of the Ethereum client.
	ConcurrentClientsNumber uint16 // Number of concurrent Ethereum clients.
	StartBlockNumber        uint64 // Starting block number for generating fixtures.
	EndBlockNumber          uint64 // Ending block number for generating fixtures.
	FixtureDataPath         string // Path to the directory where fixtures will be stored.
}

// EthGenerator is responsible for generating Ethereum fixtures.
type EthGenerator struct {
	ctx          context.Context
	config       EthGeneratorConfig
	clients      *clients.EthClient
	blocks       [][]byte
	transactions map[common.Hash][]byte
	receipts     map[common.Hash][]byte
}

// Generate generates the Ethereum fixtures.
// It retrieves blocks from the blockchain within the specified range and encodes them into RLP format.
// Transactions and receipts associated with the blocks are also encoded and stored.
func (e *EthGenerator) Generate() error {
	// Clean up previously generated data
	e.blocks = [][]byte{}
	e.transactions = make(map[common.Hash][]byte)
	e.receipts = make(map[common.Hash][]byte)

	for blockNumber := e.config.StartBlockNumber; blockNumber <= e.config.EndBlockNumber; blockNumber++ {
		// Retrieve the block by number
		block, err := e.clients.GetClient().BlockByNumber(e.ctx, big.NewInt(int64(blockNumber)))
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
			receipt, err := e.clients.GetClient().TransactionReceipt(e.ctx, tx.Hash())
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

// Write writes the generated fixtures to files.
func (e *EthGenerator) Write() error {
	blocksPath := filepath.Join(e.config.FixtureDataPath, "blocks.gob")
	if err := removeFileIfExists(blocksPath); err != nil {
		return err
	}

	if err := writeGob(blocksPath, e.blocks); err != nil {
		zap.L().Error(
			"failed to write RLP encoded blocks",
			zap.Error(err),
		)
		return err
	}

	txPath := filepath.Join(e.config.FixtureDataPath, "transactions.gob")
	if err := removeFileIfExists(txPath); err != nil {
		return err
	}

	if err := writeGob(txPath, e.transactions); err != nil {
		zap.L().Error(
			"failed to write RLP encoded block transactions",
			zap.Error(err),
		)
		return err
	}

	receiptPath := filepath.Join(e.config.FixtureDataPath, "receipts.gob")
	if err := removeFileIfExists(receiptPath); err != nil {
		return err
	}

	if err := writeGob(receiptPath, e.receipts); err != nil {
		zap.L().Error(
			"failed to write RLP encoded transaction receipts",
			zap.Error(err),
		)
		return err
	}

	zap.L().Info("Successfully wrote fixtures")

	return nil
}

// removeFileIfExists removes the file at the given path if it exists.
func removeFileIfExists(path string) error {
	if _, err := os.Stat(path); err == nil {
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}

// NewEthGenerator creates a new instance of EthGenerator.
func NewEthGenerator(ctx context.Context, config EthGeneratorConfig) (*EthGenerator, error) {
	generator := &EthGenerator{
		ctx:          ctx,
		config:       config,
		transactions: make(map[common.Hash][]byte),
		receipts:     make(map[common.Hash][]byte),
	}

	clients, err := clients.NewEthClients(config.ClientURL, config.ConcurrentClientsNumber)
	if err != nil {
		return nil, err
	}
	generator.clients = clients

	return generator, nil
}
