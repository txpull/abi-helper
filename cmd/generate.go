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
package cmd

import (
	"fmt"
	"github/txpull/abi-helper/fixtures"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// generateCmd represents the generate command
var generateEthCmd = &cobra.Command{
	Use:   "generate-eth",
	Short: "Generate ethereum based fixtures .gob files",
	RunE: func(cmd *cobra.Command, args []string) error {
		fixturesPath := viper.GetString("eth.generator.fixtures_path")

		if fixturesPath == "" {
			currentDir, err := os.Getwd()
			if err != nil {
				fmt.Println("Failed to get current directory:", err)
				return err
			}
			fixturesPath = filepath.Join(currentDir, "data", "fixtures")
		}

		zap.L().Info(
			"Starting to generate ethereum based fixtures...",
			zap.String("client-url", viper.GetString("eth.node.url")),
			zap.String("fixtures-output-path", fixturesPath),
		)

		config := fixtures.EthGeneratorConfig{
			ClientUrl:               viper.GetString("eth.node.url"),
			ConcurrentClientsNumber: viper.GetUint16("eth.node.concurrent_clients_number"),
			StartBlockNumber:        viper.GetUint64("eth.generator.start_block_number"),
			EndBlockNumber:          viper.GetUint64("eth.generator.end_block_number"),
			FixtureDataPath:         fixturesPath,
		}

		generator, err := fixtures.NewEthGenerator(cmd.Context(), config)
		if err != nil {
			return err
		}

		// Generate all of the data. Basically, fetch data from the blockchain itself.
		if err := generator.Generate(); err != nil {
			return err
		}

		if err := generator.Write(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	fixturesCmd.AddCommand(generateEthCmd)

	generateEthCmd.PersistentFlags().StringP("eth.node.url", "u", "", "Ethereum based node full url (example: https://node-url:port)")
	generateEthCmd.MarkFlagRequired("eth.node.url")
	viper.BindPFlag("eth.node.url", generateEthCmd.PersistentFlags().Lookup("eth.node.url"))

	generateEthCmd.PersistentFlags().Uint16P("eth.node.concurrent_clients_number", "c", 1, "How many concurrent node clients to spawn")
	generateEthCmd.MarkFlagRequired("eth.node.concurrent_clients_number")
	viper.BindPFlag("eth.node.concurrent_clients_number", generateEthCmd.PersistentFlags().Lookup("eth.node.concurrent_clients_number"))

	generateEthCmd.PersistentFlags().Uint64P("eth.generator.start_block_number", "s", 0, "How many concurrent node clients to spawn")
	generateEthCmd.MarkFlagRequired("eth.generator.start_block_numberr")
	viper.BindPFlag("eth.generator.start_block_number", generateEthCmd.PersistentFlags().Lookup("eth.generator.start_block_number"))

	generateEthCmd.PersistentFlags().Uint64P("eth.generator.end_block_number", "e", 0, "How many concurrent node clients to spawn")
	generateEthCmd.MarkFlagRequired("eth.generator.end_block_numberr")
	viper.BindPFlag("eth.generator.end_block_number", generateEthCmd.PersistentFlags().Lookup("eth.generator.end_block_number"))
}
