package contracts

import (
	"context"

	"github.com/txpull/sourcify-go"
	"github.com/txpull/unpack/clients"
	"github.com/txpull/unpack/scanners"
)

type Writer struct {
	// ctx holds the context to be used by the Decoder. It can be customized via WithCtx Option.
	ctx context.Context

	sourcify  *sourcify.Client
	bitquery  *scanners.BitQueryProvider
	ethClient *clients.EthClient
	bscscan   *scanners.BscScanProvider
}
