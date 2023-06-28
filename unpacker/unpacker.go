package unpacker

import (
	"context"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/txpull/sourcify-go"
	"github.com/txpull/unpack/abis"
	"github.com/txpull/unpack/clients"
	"github.com/txpull/unpack/contracts"
	"github.com/txpull/unpack/readers"
	"github.com/txpull/unpack/scanners"
)

type Options struct {
}

type Unpacker struct {
	ctx             context.Context
	sourcifyClient  *sourcify.Client
	reader          *readers.Manager
	bitquery        *scanners.BitQueryProvider
	ethClient       *clients.EthClient
	bscscan         *scanners.BscScanProvider
	contractDecoder *contracts.Decoder
}

type UnpackerOption func(*Unpacker)

func WithBitQuery(bq *scanners.BitQueryProvider) UnpackerOption {
	return func(c *Unpacker) {
		c.bitquery = bq
	}
}

func WithEthClient(client *clients.EthClient) UnpackerOption {
	return func(w *Unpacker) {
		w.ethClient = client
	}
}

func WithReaderManager(client *readers.Manager) UnpackerOption {
	return func(w *Unpacker) {
		w.reader = client
	}
}

func WithBscScan(client *scanners.BscScanProvider) UnpackerOption {
	return func(w *Unpacker) {
		w.bscscan = client
	}
}

func WithSourcify(client *sourcify.Client) UnpackerOption {
	return func(w *Unpacker) {
		w.sourcifyClient = client
	}
}

func NewUnpacker(ctx context.Context, opts ...UnpackerOption) (*Unpacker, error) {
	unpacker := &Unpacker{
		ctx: ctx,
	}

	for _, opt := range opts {
		opt(unpacker)
	}

	if unpacker.ethClient == nil {
		return nil, errors.New("eth client is required")
	}

	// We can look into redis, clickhouse or both, but we need at least one of them.
	if unpacker.reader == nil {
		return nil, errors.New("reader manager is required")
	}

	if unpacker.bitquery == nil {
		return nil, errors.New("bitquery client is required")
	}

	// Setup the decoders for future use
	if err := unpacker.setupDecoders(); err != nil {
		return nil, err
	}

	return unpacker, nil
}

func (u *Unpacker) UnpackContract(chainId *big.Int, addr common.Address, abi *abis.Decoder) (*contracts.ContractResponse, bool, error) {
	contract, complete, err := u.contractDecoder.DecodeByAddress(chainId, addr, abi)
	if err != nil {
		return nil, false, err
	}

	return contract, complete, nil
}

func (u *Unpacker) UnpackTransaction(chainId *big.Int, txHash common.Hash) error {
	return nil
}

func (u *Unpacker) UnpackReceipt(blockNumber uint64) error {
	return nil
}

func (u *Unpacker) UnpackLogs(blockNumber uint64) error {
	return nil
}

func (u *Unpacker) UnpackTrace(blockNumber uint64) error {
	return nil
}

func (u *Unpacker) setupDecoders() error {
	decoder, err := contracts.NewDecoder(
		u.ctx,
		contracts.WithReaderManager(u.reader),
		contracts.WithSourcify(u.sourcifyClient),
		contracts.WithBitQuery(u.bitquery),
		contracts.WithEthClient(u.ethClient),
		contracts.WithBscScan(u.bscscan),
	)
	if err != nil {
		return err
	}
	u.contractDecoder = decoder

	return nil
}
