package clients

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/ethclient"
)

// Error messages
var (
	ErrClientURLNotSet         = errors.New("configuration client URL not set")
	ErrConcurrentClientsNotSet = errors.New("configuration amount of concurrent clients is not set")
)

// EthClient is a struct that represents a load-balanced Ethereum client.
type EthClient struct {
	clients []*ethclient.Client
	next    uint32
}

// Len returns the number of underlying Ethereum clients.
func (c *EthClient) Len() uint16 {
	return uint16(len(c.clients))
}

// GetNetworkID retrieves the network ID from one of the underlying Ethereum clients.
func (c *EthClient) GetNetworkID(ctx context.Context) (*big.Int, error) {
	toReturn, err := c.GetClient().NetworkID(ctx)
	if err != nil {
		return nil, err
	}
	return toReturn, nil
}

// GetClient returns the next Ethereum client in a round-robin fashion.
func (c *EthClient) GetClient() *ethclient.Client {
	n := atomic.AddUint32(&c.next, 1)
	return c.clients[(int(n)-1)%len(c.clients)]
}

// Close closes all the underlying Ethereum clients.
func (c *EthClient) Close() {
	for _, client := range c.clients {
		client.Close()
	}
}

// NewEthClient creates a new load-balanced EthClient instance with multiple Ethereum clients.
func NewEthClient(url string, concurrentClients uint16) (*EthClient, error) {
	if url == "" {
		return nil, ErrClientURLNotSet
	}

	if concurrentClients == 0 {
		return nil, ErrConcurrentClientsNotSet
	}

	var wg sync.WaitGroup
	clients := make([]*ethclient.Client, 0, concurrentClients)
	mutex := sync.Mutex{}

	var ethClientErr error

	// In case we have a large number of load balanced clients, we load them concurrently
	for i := uint16(0); i < concurrentClients; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Dial an Ethereum client using the provided URL
			client, err := ethclient.Dial(url)
			if err != nil {
				ethClientErr = err
				return
			}

			mutex.Lock()
			clients = append(clients, client)
			mutex.Unlock()
		}()
	}

	wg.Wait()

	if ethClientErr != nil {
		return nil, ethClientErr
	}

	return &EthClient{
		clients: clients,
		next:    0,
	}, nil
}
