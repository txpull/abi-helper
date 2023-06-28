package clients

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/txpull/unpack/options"
)

// Error messages
var (
	ErrClientURLNotSet         error = errors.New("configuration client URL not set")
	ErrConcurrentClientsNotSet error = errors.New("configuration amount of concurrent clients is not set")
)

// EthClient represents a load-balanced Ethereum client.
// It maintains a list of underlying Ethereum clients and provides methods to interact with the Ethereum network.
type EthClient struct {
	ctx     context.Context
	opts    options.Node
	clients map[uint64][]*ethclient.Client
	next    uint32
}

// Len returns the number of underlying Ethereum clients.
func (c *EthClient) Len() int {
	return len(c.clients)
}

// GetNetworkID retrieves the network ID from one of the underlying Ethereum clients.
func (c *EthClient) GetNetworkID(ctx context.Context, chainId *big.Int) (*big.Int, error) {
	toReturn, err := c.GetClient(chainId).NetworkID(ctx)
	if err != nil {
		return nil, err
	}
	return toReturn, nil
}

// GetClient returns the next Ethereum client in a round-robin fashion.
func (c *EthClient) GetClient(chainId *big.Int) *ethclient.Client {
	n := atomic.AddUint32(&c.next, 1)
	return c.clients[chainId.Uint64()][(int(n)-1)%len(c.clients)]
}

// Close closes all the underlying Ethereum clients.
func (c *EthClient) Close() {
	for _, clients := range c.clients {
		for _, client := range clients {
			client.Close()
		}
	}
}

// ValidateOptions checks the validity of the options used to create an EthClient.
// It returns an error if any of the options are invalid.
// Specifically, it checks if the URL for the Ethereum client is set and if the number of concurrent clients is specified.
func (c *EthClient) ValidateOptions() error {
	if c.opts.URL == "" {
		return ErrClientURLNotSet
	}

	if c.opts.ConcurrentClientsNumber == 0 {
		return ErrConcurrentClientsNotSet
	}

	return nil
}

// NewEthClient creates a new EthClient with the given context and options.
// It concurrently dials the specified number of Ethereum clients and returns an EthClient that load balances requests among them.
// If any error occurs during the dialing of the Ethereum clients, it is returned.
func NewEthClient(ctx context.Context, opts options.Node) (*EthClient, error) {
	var wg sync.WaitGroup
	clients := map[uint64][]*ethclient.Client{}
	mutex := sync.Mutex{}

	errCh := make(chan error, opts.ConcurrentClientsNumber)

	for i := 0; i < opts.ConcurrentClientsNumber; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			client, err := ethclient.DialContext(ctx, opts.URL)
			if err != nil {
				errCh <- err
				return
			}

			networkId, err := client.NetworkID(ctx)
			if err != nil {
				errCh <- err
				return
			}

			mutex.Lock()
			if _, ok := clients[networkId.Uint64()]; !ok {
				clients[networkId.Uint64()] = []*ethclient.Client{}
			}

			clients[networkId.Uint64()] = append(clients[networkId.Uint64()], client)
			mutex.Unlock()
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}

	return &EthClient{
		ctx:     ctx,
		opts:    opts,
		clients: clients,
		next:    0,
	}, nil
}
