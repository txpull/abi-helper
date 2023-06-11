package fixtures

import (
	"context"
	"os"
	"path/filepath"

	"github.com/stretchr/testify/assert"
)

func GetTestFixturesPath(tAssert *assert.Assertions) string {
	currentDir, err := os.Getwd()
	tAssert.NoError(err)

	return filepath.Join(
		currentDir,
		"..", "data", "fixtures",
	)
}

func GetEthReaderForTest(tAssert *assert.Assertions) *EthReader {
	fixturesPath := GetTestFixturesPath(tAssert)

	ethReader, err := NewEthReader(context.TODO(), fixturesPath)
	tAssert.NoError(err, "failed to create EthReader")
	return ethReader
}
