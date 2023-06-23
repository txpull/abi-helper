package abi

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/txpull/unpack/clients"
	"github.com/txpull/unpack/crawlers/bscscan"
	"github.com/txpull/unpack/db"
)

type Decoder struct {
	ctx       context.Context
	badgerDb  *db.BadgerDB
	ethClient *clients.EthClient
	vcReader  *bscscan.VerifiedContractsReader
}

// WriterOption is a functional option for customizing the Decoder.
type DecoderOption func(*Decoder)

// WithContext sets the context for the Decoder.
//
// Example:
//
//	ctx, cancel := context.WithCancel(context.Background())
//	crawler := abi.NewDecoder(abi.WithContext(ctx))
//	defer cancel()
func WithContext(ctx context.Context) DecoderOption {
	return func(d *Decoder) {
		d.ctx = ctx
	}
}

func WithBadgerDb(bdb *db.BadgerDB) DecoderOption {
	return func(d *Decoder) {
		d.badgerDb = bdb
	}
}

func WithEthClient(client *clients.EthClient) DecoderOption {
	return func(d *Decoder) {
		d.ethClient = client
	}
}

func WithVerifiedContractsReader(reader *bscscan.VerifiedContractsReader) DecoderOption {
	return func(d *Decoder) {
		d.vcReader = reader
	}
}

func NewDecoder(opts ...DecoderOption) *Decoder {
	decoder := &Decoder{
		ctx: context.Background(),
	}

	for _, opt := range opts {
		opt(decoder)
	}

	return decoder
}

func (d *Decoder) Decode(addr common.Address) (*Abi, error) {
	// First we are going to look if we have the ABI in the database.
	// This is going to happen only if we have reader for verified contracts set.
	if d.vcReader != nil {
		toReturn, err := d.getAbiFromVerifiedContracts(addr)
		if err != nil {
			return nil, err
		} else if err == nil && toReturn != nil {
			return toReturn, nil
		}
	}

	// If we don't have one, we are going to attempt retrieve it from sourcify.

	// LETS GET SOURCIFY GOING HERE!

	// If we can't retrieve it from sourcify, we are going to attempt to retrieve it from bscscan.
	// If all of these fails, we are going to attempt decoding the contract bytecode.
	// If all of it fails, well then we are going to return an error.
	return nil, ErrAbiNotFound
}

func (d *Decoder) getAbiFromVerifiedContracts(addr common.Address) (*Abi, error) {
	if d.vcReader == nil {
		return nil, ErrNoVerifiedContractsReader
	}

	contract, found := d.vcReader.GetContractByAddress(addr)
	if !found {
		return nil, nil
	}

	contractAbi, err := contract.UnmarshalABI()
	if err != nil {
		return nil, err
	}

	return &Abi{contractAbi}, nil
}
