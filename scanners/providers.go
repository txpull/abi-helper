// Package providers implements contract scanner providers for BscScan and EtherScan APIs.
// It provides the ability to scan a contract using different providers and retrieve contract information.
// Each provider implements the Provider interface and can be used interchangeably.
package providers

// Provider interface defines the contract for a contract scanner provider.
type Provider interface {
	// ScanContract scans the contract identified by the given contract address.
	// It returns the contract information or an error if the scan fails.
	ScanContract(contractAddress string) (*Result, error)
}
