package readers

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/txpull/unpack/db"
	"github.com/txpull/unpack/db/models"
	"github.com/txpull/unpack/types"
)

type ClickHouseReader struct {
	ctx    context.Context
	client *db.ClickHouse
}

func NewClickHouseReader(ctx context.Context, client *db.ClickHouse) (Reader, error) {
	return &ClickHouseReader{
		ctx:    ctx,
		client: client,
	}, nil
}

func (r *ClickHouseReader) GetContractByAddress(chainId *big.Int, address common.Address) (*types.Contract, error) {
	return models.GetContract(r.ctx, r.client, chainId, address)
}

func (r *ClickHouseReader) String() string {
	return "clickhouse"
}
