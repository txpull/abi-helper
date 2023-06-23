package utils

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// InitConfig initializes the configuration settings. It attempts to read a
// configuration file located either at the path specified by the `cfgFile`
// parameter, or within the user's home directory with the name ".unpack" and
// extension ".yaml".
//
// cfgFile: The path to the configuration file. If this parameter is an empty
// string, InitConfig will attempt to search for a ".unpack.yaml" file in the
// user's home directory.
//
// This function will also read in environment variables that match the keys in
// the configuration file, with the environment variables taking precedence.
//
// If an error occurs while reading the configuration file or finding the user's
// home directory, this function will return the error. If no error occurs, the
// function will return nil.
//
// Example usage:
//
//	err := InitConfig("/path/to/config.yaml")
//	if err != nil {
//	   log.Fatal(err)
//	}
func InitConfig(cfgFile string) error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".unpack" (with .yaml extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".unpack")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}
