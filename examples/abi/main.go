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

	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/viper"
	"github.com/txpull/unpack/abi"
	"github.com/txpull/unpack/clients"
	"github.com/txpull/unpack/crawlers/bscscan"
	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/fixtures"
	"github.com/txpull/unpack/helpers"
	"github.com/txpull/unpack/logger"
	"go.uber.org/zap"
)

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

	// We are going to use BadgerDB as our database.
	// It's a key-value database that is very fast and efficient.
	// It's also very easy to use.
	// It stores all of our data that we can later on access at will.
	bdb, err := db.NewBadgerDB(
		db.WithContext(ctx),
		db.WithDbPath(viper.GetString("database.badger.path")),
	)
	if err != nil {
		logger.Error("failed to initialize badger database", zap.Error(err))
		os.Exit(1)
	}

	vcReader := bscscan.NewVerifiedContractsReader(ctx, bdb)

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

	// Not needed to be initialized for abi.NewDecoder, this is just example of how to use it.
	// Discover will load all of the contracts into the memory.
	// Can be seriously fast or slow depending on the amount of data you have in the database and
	// disk type you have. NVME is recommended.
	if err := vcReader.Discover(); err != nil {
		zap.L().Error("failed to discover database verified contracts", zap.Error(err))
		os.Exit(1)
	}

	zap.L().Info(
		"Verified contracts information",
		zap.Int("total", len(vcReader.GetContracts())),
	)

	abiDecoder := abi.NewDecoder(
		abi.WithContext(ctx),
		abi.WithBadgerDb(bdb),
		abi.WithEthClient(client),
		abi.WithVerifiedContractsReader(vcReader),
	)

	// Get known contract addresses from the database that we have in verfied contracts.
	contractInfo, err := abiDecoder.Decode(common.HexToAddress("0xFf1A42eEAf4FDFCF2f05292B01dca319EA3D2d26"))
	if err != nil {
		zap.L().Error("failed to get known contract address", zap.Error(err))
		os.Exit(1)
	}

	zap.L().Info("Contract info", zap.Int("contract_methods_length", len(contractInfo.Methods)))

	// Discover some contracts from the fixtures and attempt to get ABIs for them.
	for _, block := range fixturesReader.GetBlocks() {
		for _, tx := range block.Transactions() {
			_ = tx
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
