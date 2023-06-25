package clients

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

type Option func(*Redis)

func RedisWithAddr(addr string) Option {
	return func(r *Redis) {
		r.client.Options().Addr = addr
	}
}

func RedisWithPassword(password string) Option {
	return func(r *Redis) {
		r.client.Options().Password = password
	}
}

func RedisWithDB(db int) Option {
	return func(r *Redis) {
		r.client.Options().DB = db
	}
}

func NewRedis(options ...Option) (*Redis, error) {
	r := &Redis{
		client: redis.NewClient(&redis.Options{}),
	}

	for _, option := range options {
		option(r)
	}

	if resp := r.client.Ping(context.Background()); resp.Err() != nil {
		return nil, resp.Err()
	}

	return r, nil
}

func (r *Redis) Get(ctx context.Context, key string) ([]byte, error) {
	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	return []byte(result), nil
}

func (r *Redis) Write(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	resp, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return resp == 1, nil
}
