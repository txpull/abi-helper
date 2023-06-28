/*
Copyright © 2023 TxPull <code@txpull.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package syncers_cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/txpull/unpack/clients"
	bscscan_crawler "github.com/txpull/unpack/crawlers/bscscan"
	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/db/models"
	"github.com/txpull/unpack/options"
	"github.com/txpull/unpack/scanners"
	"go.uber.org/zap"
)

var bscscanCmd = &cobra.Command{
	Use:   "bscscan",
	Short: "Process verified contracts from bscscan",
	RunE: func(cmd *cobra.Command, args []string) error {
		bscscanPath := viper.GetString("syncers.bscscan.verified_contracts_path")

		if bscscanPath == "" {
			currentDir, err := os.Getwd()
			if err != nil {
				return err
			}
			bscscanPath = path.Join(currentDir, "data", "bscscan")
		}

		bscscanVerifiedCsvPath := path.Join(bscscanPath, "verified-contracts.csv")

		zap.L().Info(
			"Starting to process bsc scan verified contracts...",
			zap.String("bscscan-path", bscscanPath),
			zap.String("bscscan-csv-path", bscscanVerifiedCsvPath),
		)

		rdb, err := clients.NewRedis(cmd.Context(), options.G().Database.Redis)
		if err != nil {
			return fmt.Errorf("failure to initialize redis client: %s", err)
		}

		// NewBscScanProvider creates a new instance of BscScanProvider with the provided API key and API URL.
		scanner := scanners.NewBscScanProvider(viper.GetString("bscscan.api.url"), viper.GetString("bscscan.api.key"))

		client, err := clients.NewEthClient(cmd.Context(), options.G().Networks.Binance.ArchiveNode)
		if err != nil {
			return fmt.Errorf("failure to initialize eth client: %s", err)
		}

		chainId, err := client.GetNetworkID(cmd.Context())
		if err != nil {
			return fmt.Errorf("failure to get network id: %s", err)
		}

		opts := []bscscan_crawler.Option{
			bscscan_crawler.WithRequestLimit(8),
			bscscan_crawler.WithDataPath(bscscanVerifiedCsvPath),
			bscscan_crawler.WithMaxRetry(5),
			bscscan_crawler.WithBackoffFactor(2),
			bscscan_crawler.WithScanner(scanner),
			bscscan_crawler.WithRedis(rdb),
			bscscan_crawler.WithEthClient(client),
			bscscan_crawler.WithChainID(chainId),
		}

		if viper.GetBool("syncers.bscscan.write_to_clickhouse") {
			cdb, err := db.NewClickHouse(cmd.Context(), options.G().Database.Clickhouse)
			if err != nil {
				return err
			}

			if err := db.ExecClickHouseQuery(cmd.Context(), cdb, "SET allow_experimental_object_type = ?;", "1"); err != nil {
				return fmt.Errorf("failure to set allow_experimental_object_type: %s", err)
			}

			if err := models.CreateContractsTable(cmd.Context(), cdb); err != nil {
				return fmt.Errorf("failure to create (if does not exist) contracts table: %s", err)
			}

			if err := models.CreateMethodsTable(cmd.Context(), cdb); err != nil {
				return fmt.Errorf("failure to create (if does not exist) methods table: %s", err)
			}

			if err := models.CreateMethodMapperTable(cmd.Context(), cdb); err != nil {
				return fmt.Errorf("failure to create (if does not exist) methods_mapper table: %s", err)
			}

			if err := models.CreateEventsTable(cmd.Context(), cdb); err != nil {
				return fmt.Errorf("failure to create (if does not exist) events table: %s", err)
			}

			if err := models.CreateEventMapperTable(cmd.Context(), cdb); err != nil {
				return fmt.Errorf("failure to create (if does not exist) events_mapper table: %s", err)
			}

			opts = append(opts, bscscan_crawler.WithClickHouseDb(cdb))
		}

		crawler := bscscan_crawler.NewVerifiedContractsWritter(
			cmd.Context(),
			opts...,
		)

		contracts, err := crawler.GatherVerifiedContracts()
		if err != nil {
			return err
		}

		if err := crawler.ProcessVerifiedContracts(contracts); err != nil {
			return err
		}

		zap.L().Info("Successfully processed verified contracts")

		return nil
	},
}
