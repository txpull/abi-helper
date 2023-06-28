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
package fixtures_cmd

import (
	"os"
	"path/filepath"

	"github.com/txpull/unpack/fixtures"
	"github.com/txpull/unpack/options"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// generateCmd represents the generate command
var generateEthCmd = &cobra.Command{
	Use:   "generate-eth",
	Short: "Generate ethereum based fixtures and write them into (block|transactions|receipt).gob files",
	RunE: func(cmd *cobra.Command, args []string) error {
		fixturesPath := viper.GetString("eth.generator.fixtures_path")

		if fixturesPath == "" {
			currentDir, err := os.Getwd()
			if err != nil {
				return err
			}
			fixturesPath = filepath.Join(currentDir, "data", "fixtures")
		}

		zap.L().Info(
			"Starting to generate ethereum based fixtures...",
			zap.String("client-url", viper.GetString("eth.node.url")),
			zap.String("fixtures-output-path", fixturesPath),
		)

		writer, err := fixtures.NewEthWriter(cmd.Context(), options.G().Fixtures)
		if err != nil {
			return err
		}

		// Generate all of the data. Basically, fetch data from the blockchain itself
		if err := writer.Generate(); err != nil {
			return err
		}

		// Will remove existing files in fixture data path and replace files with new content
		if err := writer.Write(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	fixturesCmd.AddCommand(generateEthCmd)
}
