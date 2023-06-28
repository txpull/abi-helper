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
	"github.com/txpull/unpack/clients"
	"github.com/txpull/unpack/crawlers/fourbyte"
	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/db/models"
	"github.com/txpull/unpack/options"
	"github.com/txpull/unpack/scanners"
	"go.uber.org/zap"
)

var fourbyteCmd = &cobra.Command{
	Use:   "fourbyte",
	Short: "Download, process and store signatures from 4byte.directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		rdb, err := clients.NewRedis(cmd.Context(), options.G().Database.Redis)
		if err != nil {
			return fmt.Errorf("failure to initialize redis client: %s", err)
		}

		var fourbOpts []fourbyte.WriterOption

		provider := scanners.NewFourByteProvider(
			scanners.WithCtx(cmd.Context()),
			scanners.WithURL(viper.GetString("syncers.fourbyte.url")), // Replace with your URL
			scanners.WithMaxRetries(3),
		)

		opts := append(fourbOpts,
			fourbyte.WithCtx(cmd.Context()),
			fourbyte.WithProvider(provider),
			fourbyte.WithRedis(rdb),
			fourbyte.WithCooldown(100*time.Millisecond),
			fourbyte.WithChainID(big.NewInt(viper.GetInt64("syncers.fourbyte.chain_id"))),
		)

		// If ClickHouse is enabled, we are going to write signatures into it
		if viper.GetBool("syncers.fourbyte.write_to_clickhouse") {
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

			opts = append(opts, fourbyte.WithClickHouseDb(cdb))
		}

		crawler := fourbyte.NewFourByteWriter(opts...)

		if err := crawler.Crawl(); err != nil {
			return err
		}

		zap.L().Info("Successfully processed 4byte.dictionary signatures")

		return nil
	},
}
