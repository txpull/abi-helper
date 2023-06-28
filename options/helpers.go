package options

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// globalOptions is a global variable of type Options.
// It is used to hold the options settings that are populated by the New function.
// The settings can be accessed from anywhere in the code using the G function.
var globalOptions Options

// G is a global function that returns a pointer to the globalOptions variable.
// It can be used to easily access the options settings from anywhere in the code.
func G() *Options {
	return &globalOptions
}

// New creates a new Options struct and populates it with the settings from a configuration file.
// The configuration file is specified by the optFile parameter. If optFile is an empty string,
// the function looks for a configuration file named "unpack.toml" in the ".unpack" directory
// within the user's home directory.
//
// The function also reads in environment variables that match the keys in the configuration file.
//
// If a configuration file is found, the function reads it and unmarshals its contents into the
// globalOptions variable, which is of type Options.
//
// The function returns a pointer to the populated Options struct and an error if any occurred
// during the process. If the configuration file cannot be read or the contents cannot be unmarshaled
// into the Options struct, the function returns an error.
func New(optFile string) (*Options, error) {
	if optFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(optFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Make sure that default options directory is /home/{user}/.unpack/
		home = filepath.Join(home, ".unpack")
		viper.AddConfigPath(home)

		// Search config in home directory with name "unpack" (with .toml extension).
		viper.SetConfigType("toml")
		viper.SetConfigName("unpack")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Unmarshal options into globally accessible struct
	err := viper.Unmarshal(&globalOptions)
	if err != nil {
		return nil, fmt.Errorf("unable to decode options into struct: %v", err)
	}

	return &globalOptions, nil
}
