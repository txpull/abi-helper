// Package bscscan provides a tool for scanning verified smart contracts
// on the Binance Smart Chain (BSC). It uses the BscScan API to fetch and
// process information about these smart contracts.
package bscscan

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/txpull/bytecode/scanners"
	"github.com/txpull/bytecode/utils"
	"go.uber.org/zap"
)

const (
	requestLimit    = 6                      // Maximum number of requests allowed per second
	requestInterval = 100 * time.Millisecond // Time interval between each request
	maxRetry        = 5                      // Maximum number of retries
	backoffFactor   = 2                      // Backoff factor for exponential backoff
)

// CsvContract holds a transaction hash, a contract address, and a contract name
type CsvContract struct {
	Txhash          string
	ContractAddress string
	ContractName    string
}

// BscscanWritter encapsulates the context, scanner, data paths, rate-limiting mechanism,
// and concurrency controls needed to interact with the BscScan API.
type BscscanWritter struct {
	ctx            context.Context
	scanner        *scanners.BscScanProvider
	dataPath       string
	outputDataFile string
	semaphore      chan struct{}
	wg             sync.WaitGroup
	mu             sync.Mutex // Mutex for safe access to results
}

// New creates and returns a new Bscscan instance.
//
// Example:
//
//	ctx := context.Background()
//	scanner := scanners.NewBscScanProvider("your-api-key")
//	bs := bscscan.New(ctx, scanner, "path/to/data.csv", "path/to/output.gob")
func New(ctx context.Context, scanner *scanners.BscScanProvider, dataPath string, outputDataFile string) *BscscanWritter {
	return &BscscanWritter{ctx: ctx, scanner: scanner, dataPath: dataPath, outputDataFile: outputDataFile}
}

// GatherVerifiedContracts reads contract data from a CSV file.
// It skips any records that do not contain valid contract data.
//
// Example:
//
//	contracts, err := bs.GatherVerifiedContracts()
//	if err != nil {
//	    log.Fatal(err)
//	}
func (bs *BscscanWritter) GatherVerifiedContracts() ([]CsvContract, error) {
	file, err := os.Open(bs.dataPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	var contracts []CsvContract

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		if strings.Contains(record[0], "Note:") || strings.Contains(record[0], "Txhash") {
			continue
		}

		contract := CsvContract{
			Txhash:          record[0],
			ContractAddress: record[1],
			ContractName:    record[2],
		}

		contracts = append(contracts, contract)
	}

	return contracts, nil
}

// ProcessVerifiedContracts processes each contract in a slice of CsvContract.
// It scans each contract asynchronously, respecting the API rate limit.
// It retries requests in case of a "Max rate limit reached" error,
// using an exponential backoff strategy. The result for each contract
// is appended to a gob file.
//
// Example:
//
//	err = bs.ProcessVerifiedContracts(contracts)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (bs *BscscanWritter) ProcessVerifiedContracts(contracts []CsvContract) error {
	if len(contracts) < 1 {
		return errors.New("no contracts to process")
	}

	bs.semaphore = make(chan struct{}, requestLimit)

	// Create a map to hold all contractInfo results with contract address as key
	results := make(map[string][]byte)

	for _, contract := range contracts {
		bs.wg.Add(1)
		bs.semaphore <- struct{}{}

		go func(c CsvContract) {
			defer bs.wg.Done()

			for i := 0; i < maxRetry; i++ {
				zap.L().Info("Processing contract", zap.String("contract_address", c.ContractAddress))

				contractInfo, err := bs.scanner.ScanContract(c.ContractAddress)

				if err != nil {
					if strings.Contains(err.Error(), "Max rate limit reached") {
						zap.L().Error(
							"Max rate limit reached, retrying...",
							zap.String("contract_address", c.ContractAddress),
							zap.Int("retry_attempt", i+1),
							zap.Error(err),
						)
						time.Sleep(time.Duration(int64(math.Pow(backoffFactor, float64(i)))) * time.Second)
						continue
					} else {
						zap.L().Error(
							"Failed to get contract information from BSCScan",
							zap.String("contract_address", c.ContractAddress),
							zap.Error(err),
						)
					}
				}

				resBytes, err := contractInfo.MarshalBytes()
				if err != nil {
					zap.L().Error(
						"Failed to marshal contract information to binary",
						zap.String("contract_address", c.ContractAddress),
						zap.Error(err),
					)
					break
				}

				bs.mu.Lock()
				results[c.ContractAddress] = resBytes
				bs.mu.Unlock()

				break
			}

			<-bs.semaphore
		}(contract)

		time.Sleep(requestInterval)
	}

	// Wait for all contracts to finish processing
	bs.wg.Wait()

	// Write all results to a .gob file
	if err := utils.WriteGob(bs.outputDataFile, results); err != nil {
		zap.L().Error(
			"failed to write results to gob file",
			zap.String("file_path", bs.outputDataFile),
			zap.Error(err),
		)
		return err
	}

	return nil
}
