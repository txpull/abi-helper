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
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/txpull/bytecode/crawlers/fourbyte"
	"github.com/txpull/bytecode/db"
	"github.com/txpull/bytecode/scanners"
	"go.uber.org/zap"
)

var fourbyteCmd = &cobra.Command{
	Use:   "fourbyte",
	Short: "Download, process and store signatures from 4byte.directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		databasePath := viper.GetString("database.badger.path")

		if databasePath == "" {
			currentDir, err := os.Getwd()
			if err != nil {
				return err
			}
			databasePath = path.Join(currentDir, "data", "db")
		}

		// Open the Badger database located in the databasePath directory.
		// It will be created if it doesn't exist.
		bdb, err := db.NewBadgerDB(db.WithDbPath(databasePath))
		if err != nil {
			return err
		}
		defer bdb.Close()
		go bdb.GarbageCollect()

		var fourbOpts []fourbyte.WriterOption

		provider := scanners.NewFourByteProvider(
			scanners.WithContext(cmd.Context()),
			scanners.WithURL(viper.GetString("syncers.fourbyte.url")), // Replace with your URL
			scanners.WithMaxRetries(3),
			scanners.WithContext(cmd.Context()),
		)

		opts := append(fourbOpts,
			fourbyte.WithContext(cmd.Context()),
			fourbyte.WithProvider(provider),
			fourbyte.WithDB(bdb),
			fourbyte.WithCooldown(100*time.Millisecond),
		)

		// If ClickHouse is enabled, we are going to write signatures into it
		if viper.GetBool("syncers.fourbyte.write_to_clickhouse") {
			cdb, err := db.NewClickHouse(
				db.WithCtx(cmd.Context()),
				db.WithDebug(true),
				db.WithHost(viper.GetString("database.clickhouse.host")),
				db.WithDatabase(viper.GetString("database.clickhouse.database")),
				db.WithUsername(viper.GetString("database.clickhouse.username")),
				db.WithPassword(viper.GetString("database.clickhouse.password")),
			)
			if err != nil {
				return err
			}

			// Will create database if it does not exist and return error if it fails
			if err := cdb.CreateDatabase(viper.GetString("database.clickhouse.database")); err != nil {
				return err
			}

			opts = append(opts, fourbyte.WithPostHook(fourbyte.PostWriteClickHouseHook(cdb)))
		}

		crawler := fourbyte.NewFourByteWriter(opts...)

		if err := crawler.Crawl(); err != nil {
			return err
		}

		zap.L().Info("Successfully processed 4byte.dictionary signatures")

		return nil
	},
}
