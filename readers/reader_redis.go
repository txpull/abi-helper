package readers

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/txpull/unpack/clients"
	"github.com/txpull/unpack/types"
)

type RedisReader struct {
	ctx    context.Context
	client *clients.Redis
}

func NewRedisReader(ctx context.Context, client *clients.Redis) (Reader, error) {
	return &RedisReader{
		ctx:    ctx,
		client: client,
	}, nil
}

func (r *RedisReader) GetContractByAddress(chainId *big.Int, address common.Address) (*types.Contract, error) {
	redisKey := types.GetContractStorageKey(chainId, address)
	contractBytes, err := r.client.Get(r.ctx, redisKey)
	if err != nil {
		return nil, err
	}

	contract := &types.Contract{}
	if err := contract.UnmarshalBytes(contractBytes); err != nil {
		return nil, err
	}

	return contract, nil
}

func (r *RedisReader) GetMethodBySignature(chainId *big.Int, signature string) (*types.Method, error) {
	redisKey := types.GetMethodStorageKey(chainId, common.Hex2Bytes(signature))
	methodBytes, err := r.client.Get(r.ctx, redisKey)
	if err != nil {
		return nil, err
	}

	method := &types.Method{}
	if err := method.UnmarshalBytes(methodBytes); err != nil {
		return nil, err
	}

	return method, nil
}

func (r *RedisReader) GetEventByHash(chainId *big.Int, hash common.Hash) (*types.Event, error) {
	redisKey := types.GetEventStorageKey(chainId, hash)
	eventBytes, err := r.client.Get(r.ctx, redisKey)
	if err != nil {
		return nil, err
	}

	event := &types.Event{}
	if err := event.UnmarshalBytes(eventBytes); err != nil {
		return nil, err
	}

	return event, nil
}

func (r *RedisReader) String() string {
	return "redis"
}
