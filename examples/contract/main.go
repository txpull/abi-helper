// Package main is the entry point of the bytecode application.
// It initializes the required context, logger, configuration, Ethereum client, Badger database,
// and contract reader. It also loads and processes the fixtures for Ethereum blocks and transactions.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"github.com/txpull/unpack/clients"
	"github.com/txpull/unpack/contracts"
	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/fixtures"
	"github.com/txpull/unpack/helpers"
	"github.com/txpull/unpack/logger"
	"go.uber.org/zap"
)

// main is the entry point of the application.
// It initializes a context, a development logger, the configuration settings using a config file,
// an Ethereum client, a Badger database, and a contract reader.
// Then, it loads the Ethereum fixtures and processes each transaction in each block.
// For each transaction that corresponds to a contract creation, it processes the contract creation transaction.
func main() {
	// INITIALIZATION OF THE PREREQUISITES

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// NOTE: Production code should never-ever use Development logger as it's not as efficient.
	// Commands themselves can use collored output for clearner visibility into the logs.
	// Production will print json-based log structure to the stdout
	logger, err := logger.GetDevelopmentLogger()
	if err != nil {
		panic(fmt.Errorf("failure to construct logger: %s", err))
	}
	defer logger.Sync()

	// Basically what we want to achieve is ue zap.L() throughout project without
	// passing by reference logger everywhere.
	zap.ReplaceGlobals(logger)

	cfgFile := flag.String("config", "../../.unpack.yaml", "Path to the configuration file")
	flag.Parse()

	if err := helpers.InitConfig(*cfgFile); err != nil {
		logger.Error("failed to initialize configuration", zap.Error(err))
		os.Exit(1)
	}

	client, err := clients.NewEthClients(
		viper.GetString("eth.archive_node.url"),
		viper.GetUint16("eth.archive_node.concurrent_clients_number"),
	)
	if err != nil {
		logger.Error("failed to initialize eth client", zap.Error(err))
		os.Exit(1)
	}

	bdb, err := db.NewBadgerDB(
		db.WithContext(ctx),
		db.WithDbPath(viper.GetString("database.badger.path")),
	)
	if err != nil {
		logger.Error("failed to initialize badger database", zap.Error(err))
		os.Exit(1)
	}

	// What we are going to do now is use fixtures to load contract, transaction, block and receipt data
	// Why not when we have them already?
	// We can use them to test our contract and make sure it works as expected.

	fixturesPath, err := getFixturesPath()
	if err != nil {
		zap.L().Error("failed to get fixtures path", zap.Error(err))
		os.Exit(1)
	}

	zap.L().Info("Initializing fixtures...", zap.String("path", fixturesPath))

	// New reader will automatically load fixtures from the path. This may take a bit of time.
	// It depends on your machine and the amount of data you have in the fixtures.
	fixturesReader, err := fixtures.NewEthReader(ctx, fixturesPath)
	if err != nil {
		zap.L().Error("failed to construct fixtures", zap.Error(err))
		os.Exit(1)
	}

	// END INITIALIZATION OF THE PREREQUISITES

	decoder := contracts.NewContractDecoder(
		contracts.WithContext(ctx),
		contracts.WithEthClient(client),
		contracts.WithDB(bdb),
	)

	for _, block := range fixturesReader.GetBlocks() {
		for _, tx := range block.Transactions() {
			//zap.L().Info("Processing transaction...", zap.String("hash", tx.Hash().Hex()))
			receipt, receiptOk := fixturesReader.GetReceiptFromTxHash(tx.Hash())
			if !receiptOk {
				zap.L().Error("failed to get receipt for transaction", zap.String("hash", tx.Hash().Hex()))
				continue
			}

			if tx.To() == nil && receipt.Status == 1 && receipt.ContractAddress != helpers.ZeroAddress {
				// Just to make it more readable
				zap.L().Info("-----------------------------------")

				zap.L().Info("Processing contract creation transaction...",
					zap.String("hash", tx.Hash().Hex()),
					zap.String("contract_address", receipt.ContractAddress.Hex()),
				)

				// We have a contract creation transaction. Let's process it.
				result, err := decoder.ProcessContractCreationTx(
					block,
					tx,
					receipt,
					receipt.ContractAddress,
				)
				if err != nil {
					zap.L().Error("failed to process contract", zap.Error(err))
					continue
				}

				_ = result

				// We can use this to determine if the contract is a proxy contract or not.
				// If it is a proxy contract, we can use the proxy contract reader to process it.
				// If it is not a proxy contract, we can use the regular contract reader to process it.
				// We can also use this to determine if the contract is a contract factory or not.
				// If it is a contract factory, we can use the contract factory reader to process it.
				// If it is not a contract factory, we can use the regular contract reader to process it.

				zap.L().Info(
					"Contract processed successfully",
					zap.Int64("block_number", block.Number().Int64()),
					zap.String("tx_hash", tx.Hash().Hex()),
					//zap.Reflect("result", result),
				)

			}

		}
	}
}

// getFixturesPath returns the path to the Ethereum fixtures.
// It retrieves the path from the configuration settings. If the path is not defined in the configuration,
// it assumes the fixtures are in the "../../data/fixtures" directory relative to the current working directory.
// Returns an error if it fails to get the current working directory.
func getFixturesPath() (string, error) {
	fixturesPath := viper.GetString("eth.generator.fixtures_path")

	if fixturesPath == "" {
		currentDir, err := os.Getwd()
		if err != nil {
			return "", err
		}
		fixturesPath = filepath.Join(currentDir, "..", "..", "data", "fixtures")
	}

	return fixturesPath, nil
}
