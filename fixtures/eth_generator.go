package fixtures

import (
	"context"
	"github/txpull/abi-helper/clients"
	"math/big"
	"os"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"go.uber.org/zap"
)

// EthGeneratorConfig holds the configuration for EthGenerator.
type EthGeneratorConfig struct {
	ClientUrl               string
	ConcurrentClientsNumber uint16
	StartBlockNumber        uint64
	EndBlockNumber          uint64
	FixtureDataPath         string
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
func (e *EthGenerator) Generate() error {

	// Reset data...
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

		transactions := block.Transactions()

		for _, tx := range transactions {
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
			txBytes, err := rlp.EncodeToBytes(block)
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

			// Encode the transaction receipt into RLP format
			receiptBytes, err := rlp.EncodeToBytes(receipt)
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

	if _, err := os.Stat(blocksPath); err == nil {
		if err := os.Remove(blocksPath); err != nil {
			return err
		}
	}

	if err := writeGob(blocksPath, e.blocks); err != nil {
		zap.L().Error(
			"failed to write RLP encoded blocks",
			zap.Error(err),
		)
		return err
	}

	txPath := filepath.Join(e.config.FixtureDataPath, "transactions.gob")

	if _, err := os.Stat(txPath); err == nil {
		if err := os.Remove(txPath); err != nil {
			return err
		}
	}

	if err := writeGob(txPath, e.transactions); err != nil {
		zap.L().Error(
			"failed to write RLP encoded block transactions",
			zap.Error(err),
		)
		return err
	}

	receiptPath := filepath.Join(e.config.FixtureDataPath, "receipts.gob")

	if _, err := os.Stat(receiptPath); err == nil {
		if err := os.Remove(receiptPath); err != nil {
			return err
		}
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

// NewEthGenerator creates a new instance of EthGenerator.
func NewEthGenerator(ctx context.Context, config EthGeneratorConfig) (Generator, error) {
	generator := EthGenerator{
		ctx:          ctx,
		config:       config,
		transactions: make(map[common.Hash][]byte),
		receipts:     make(map[common.Hash][]byte),
	}

	clients, err := clients.NewEthClients(config.ClientUrl, config.ConcurrentClientsNumber)
	if err != nil {
		return nil, err
	}
	generator.clients = clients

	return Generator(&generator), nil
}
