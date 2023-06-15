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

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	bscscan_crawler "github.com/txpull/bytecode/crawlers/bscscan"
	"github.com/txpull/bytecode/scanners"
	"go.uber.org/zap"
)

type Contract struct {
	Txhash          string
	ContractAddress string
	ContractName    string
}

var bscscanCmd = &cobra.Command{
	Use:   "bscscan",
	Short: "Process verified contracts from bscscan",
	RunE: func(cmd *cobra.Command, args []string) error {
		bscscanPath := viper.GetString("bsc.crawler.bscscan_path")

		if bscscanPath == "" {
			currentDir, err := os.Getwd()
			if err != nil {
				return err
			}
			bscscanPath = path.Join(currentDir, "data", "bscscan")
		}

		bscscanVerifiedCsvPath := path.Join(bscscanPath, "verified-contracts.csv")
		bscscanVerifiedOoutputPath := path.Join(bscscanPath, "verified-contracts.gob")

		zap.L().Info(
			"Starting to process bsc scan verified contracts...",
			zap.String("bscscan-path", bscscanPath),
			zap.String("bscscan-csv-path", bscscanVerifiedCsvPath),
		)

		// NewBscScanProvider creates a new instance of BscScanProvider with the provided API key and API URL.
		bp := scanners.NewBscScanProvider(
			viper.GetString("bscscan.api.url"),
			viper.GetString("bscscan.api.key"),
		)

		crawler := bscscan_crawler.New(cmd.Context(), bp, bscscanVerifiedCsvPath, bscscanVerifiedOoutputPath)

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
