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
package main

import (
	"fmt"
	"os"

	"github.com/txpull/unpack/cmd"
	"github.com/txpull/unpack/logger"

	"go.uber.org/zap"
)

var (
	// Current software version flag obtained from go build -ldflags
	Version string
)

func main() {
	// NOTE: Production code should never-ever use Development logger as it's not as efficient.
	// Commands themselves can use collored output for clearner visibility into the logs.
	// Production will print json-based log structure to the stdout
	logger, err := logger.GetDevelopmentLogger()
	if err != nil {
		panic(fmt.Errorf("failure to construct logger: %s", err))
	}
	defer logger.Sync()

	// Basically what we want to achieve is ue zap.L() throughout project without
	// passing by reference logger everywhere.
	zap.ReplaceGlobals(logger)

	// Sets the version variable for future consumption
	cmd.SetVersion(Version)

	if err := cmd.Execute(); err != nil {
		logger.Error("failure to execute root command", zap.Error(err))
		os.Exit(1)
	}
}
