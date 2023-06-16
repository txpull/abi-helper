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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	fixtures_cmd "github.com/txpull/bytecode/cmd/fixtures"
	syncers_cmd "github.com/txpull/bytecode/cmd/syncers"
	"github.com/txpull/bytecode/utils"
	"go.uber.org/zap"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "txbyte",
	Short: "A brief description of your application",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

// SetVersion sets version of the abihelper obtained from the go.build -ldflags
func SetVersion(v string) {
	version = v
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if err := utils.InitConfig(cfgFile); err != nil {
		zap.L().Error("failed to initialize configuration", zap.Error(err))
		os.Exit(1)
	}

	zap.L().Info("Using configuration file", zap.String("path", viper.ConfigFileUsed()))
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.txbyte.yaml)")

	// Load syncers subcommands designed to sync data from third party sources
	syncers_cmd.Init(rootCmd)

	// Load fixtures subcommands designed to generate fixtures for testing purposes using mainnet nodes data
	fixtures_cmd.Init(rootCmd)
}
