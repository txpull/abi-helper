package fixtures

import (
	"context"
	"github/txpull/abi-helper/clients"
	"math/big"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"go.uber.org/zap"
)

type EthGeneratorConfig struct {
	ClientUrl               string
	ConcurrentClientsNumber uint16
	StartBlockNumber        uint64
	EndBlockNumber          uint64
	FixtureDataPath         string
}

type EthGenerator struct {
	ctx     context.Context
	config  EthGeneratorConfig
	clients *clients.EthClient

	blocks       [][]byte
	transactions map[common.Hash][]byte
	receipts     map[common.Hash][]byte
}

func (e *EthGenerator) Generate() error {
	for blockNumber := e.config.StartBlockNumber; blockNumber <= e.config.EndBlockNumber; blockNumber++ {
		block, err := e.clients.GetClient().BlockByNumber(e.ctx, big.NewInt(int64(blockNumber)))
		if err != nil {
			zap.L().Error(
				"failed to retrieve block",
				zap.Uint64("block_number", blockNumber),
				zap.Error(err),
			)
			return err
		}

		blockBytes, err := rlp.EncodeToBytes(block)
		if err != nil {
			zap.L().Error(
				"failed to rlp encode block",
				zap.Uint64("block_number", blockNumber),
				zap.Error(err),
			)
			return err
		}
		e.blocks = append(e.blocks, blockBytes)

		transactions := block.Transactions()

		for _, tx := range transactions {
			receipt, err := e.clients.GetClient().TransactionReceipt(e.ctx, tx.Hash())
			if err != nil {
				zap.L().Error(
					"failed to retreive transaction receipt",
					zap.Uint64("block_number", blockNumber),
					zap.String("tx_hash", tx.Hash().Hex()),
					zap.Error(err),
				)
				continue
			}

			txBytes, err := rlp.EncodeToBytes(block)
			if err != nil {
				zap.L().Error(
					"failed to rlp encode transaction",
					zap.Uint64("block_number", blockNumber),
					zap.String("tx_hash", tx.Hash().Hex()),
					zap.Error(err),
				)
				return err
			}
			e.transactions[tx.Hash()] = txBytes

			receiptBytes, err := rlp.EncodeToBytes(receipt)
			if err != nil {
				zap.L().Error(
					"failed to rlp encode transaction receipt",
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

func (e *EthGenerator) Write() error {
	blocksPath := filepath.Join(e.config.FixtureDataPath, "blocks.gob")
	if err := writeGob(blocksPath, e.blocks); err != nil {
		zap.L().Error(
			"failed to write rlp encoded blocks",
			zap.Error(err),
		)
		return err
	}

	txPath := filepath.Join(e.config.FixtureDataPath, "transactions.gob")
	if err := writeGob(txPath, e.transactions); err != nil {
		zap.L().Error(
			"failed to write rlp encoded block transactions",
			zap.Error(err),
		)
		return err
	}

	receiptPath := filepath.Join(e.config.FixtureDataPath, "receipts.gob")
	if err := writeGob(receiptPath, e.receipts); err != nil {
		zap.L().Error(
			"failed to write rlp encoded transaction receipts",
			zap.Error(err),
		)
		return err
	}

	zap.L().Info("Successfully wrote fixtures")

	return nil
}

func (e *EthGenerator) Read() error {
	return nil
}

func NewEthGenerator(ctx context.Context, config EthGeneratorConfig) (Generator, error) {
	toReturn := EthGenerator{
		ctx:          ctx,
		config:       config,
		transactions: make(map[common.Hash][]byte),
		receipts:     make(map[common.Hash][]byte),
	}

	clients, err := clients.NewEthClients(config.ClientUrl, config.ConcurrentClientsNumber)
	if err != nil {
		return nil, err
	}
	toReturn.clients = clients

	return Generator(&toReturn), nil
}
