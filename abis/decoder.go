package abis

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/txpull/unpack/readers"
)

type Decoder struct {
	ctx    context.Context
	reader *readers.Manager
	abiRaw string
	abi    abi.ABI
}

func NewDecoder(ctx context.Context, r *readers.Manager, abiRaw string) (*Decoder, error) {
	decoder := &Decoder{ctx: ctx, reader: r, abiRaw: abiRaw}

	if err := decoder.DecodeRawAbi(); err != nil {
		return nil, fmt.Errorf("failed to decode raw abi: %w", err)
	}

	return decoder, nil
}

func (d *Decoder) GetABI() abi.ABI {
	return d.abi
}

func (d *Decoder) DecodeRawAbi() error {
	parsedAbi, err := abi.JSON(strings.NewReader(d.abiRaw))
	if err != nil {
		return err
	}

	d.abi = parsedAbi

	return nil
}
