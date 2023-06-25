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
	"math/big"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	sourcify_go "github.com/txpull/sourcify-go"
	"github.com/txpull/unpack/clients"
	"github.com/txpull/unpack/crawlers/sourcify"
	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/db/models"
	"github.com/txpull/unpack/scanners"
	"go.uber.org/zap"
)

var sourcifyCmd = &cobra.Command{
	Use:   "sourcify",
	Short: "Download, process and store contracts from sourcify.dev",
	RunE: func(cmd *cobra.Command, args []string) error {
		rdb, err := clients.NewRedis(
			clients.RedisWithAddr(viper.GetString("database.redis.addr")),
			clients.RedisWithPassword(viper.GetString("database.redis.password")),
			clients.RedisWithDB(viper.GetInt("database.redis.db")),
		)
		if err != nil {
			return fmt.Errorf("failure to initialize redis client: %s", err)
		}

		var sourcifyOpts []sourcify.WriterOption

		provider := sourcify_go.NewClient(
			sourcify_go.WithBaseURL(viper.GetString("syncers.sourcify.url")),
			sourcify_go.WithRetryOptions(
				sourcify_go.WithMaxRetries(viper.GetInt("syncers.sourcify.max_retries")),
				sourcify_go.WithDelay(viper.GetDuration("syncers.sourcify.retry_dely")*time.Second),
			),
			sourcify_go.WithRateLimit(viper.GetInt("syncers.sourcify.rate_limit_s"), 1*time.Second),
		)

		bitquery := scanners.NewBitQueryProvider(
			viper.GetString("bitquery.api.url"),
			viper.GetString("bitquery.api.key"),
		)

		client, err := clients.NewEthClients(
			viper.GetString("nodes.eth.archive.url"),
			viper.GetUint16("nodes.eth.archive.concurrent_clients_number"),
		)
		if err != nil {
			return err
		}

		bscscan := scanners.NewBscScanProvider(viper.GetString("bscscan.api.url"), viper.GetString("bscscan.api.key"))

		opts := append(sourcifyOpts,
			sourcify.WithCtx(cmd.Context()),
			sourcify.WithSourcify(provider),
			sourcify.WithRedis(rdb),
			sourcify.WithBitQuery(bitquery),
			sourcify.WithEthClient(client),
			sourcify.WithBscScan(bscscan),
		)

		// If ClickHouse is enabled, we are going to write signatures into it
		if viper.GetBool("syncers.sourcify.write_to_clickhouse") {
			cdb, err := db.NewClickHouse(
				db.WithCtx(cmd.Context()),
				db.WithDebug(false),
				db.WithHost(viper.GetString("database.clickhouse.host")),
				db.WithDatabase(viper.GetString("database.clickhouse.database")),
				db.WithUsername(viper.GetString("database.clickhouse.username")),
				db.WithPassword(viper.GetString("database.clickhouse.password")),
			)
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

			opts = append(opts, sourcify.WithClickHouseDb(cdb))
		}

		writer := sourcify.NewSourcifyWriter(opts...)

		// This may take a very, very long time to finish as there are bunch of contracts
		// and it depends on how many you wish to retreive from the sourcify API
		for _, chainId := range viper.GetStringSlice("syncers.sourcify.chain_ids") {
			parsedChainId, ok := big.NewInt(0).SetString(chainId, 10)
			if !ok {
				return fmt.Errorf("failure to parse chain id: %v", chainId)
			}

			contracts, err := writer.GetContractListByChainID(parsedChainId)
			if err != nil {
				return fmt.Errorf("failure to get contract list by chain: %s", err)
			}

			zap.L().Info(
				"Discovered sourcify contracts information",
				zap.Int("full_contracts", len(contracts.Full)),
				zap.Int("partial_contracts", len(contracts.Partial)),
				zap.String("chain_id", chainId),
			)

			if len(contracts.Full) > 0 {
				if err := writer.ProcessContractsByType(parsedChainId, contracts, sourcify_go.MethodMatchTypeFull); err != nil {
					zap.L().Error("Failure to process sourcify full contracts", zap.Error(err))
				}
			}

			if len(contracts.Partial) > 0 {
				if err := writer.ProcessContractsByType(parsedChainId, contracts, sourcify_go.MethodMatchTypePartial); err != nil {
					zap.L().Error("Failure to process sourcify full contracts", zap.Error(err))
				}
			}

			zap.L().Info("Successfully processed sourcify contracts", zap.String("chain_id", chainId))
		}

		return nil
	},
}
