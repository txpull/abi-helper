package db

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ClickHouse struct {
	ctx  context.Context
	conn driver.Conn
}

type ClickHouseOption func(*ClickHouse, *clickhouse.Options)

func WithCtx(ctx context.Context) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		c.ctx = ctx
	}
}

func WithHost(host string) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.Addr = []string{host}
	}
}

func WithDatabase(database string) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.Auth.Database = database
	}
}

func WithUsername(username string) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.Auth.Username = username
	}
}

func WithPassword(password string) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.Auth.Password = password
	}
}

func WithDialTimeout(timeout time.Duration) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.DialTimeout = timeout
	}
}

func WithMaxOpenConns(maxConns int) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.MaxOpenConns = maxConns
	}
}

func WithMaxIdleConns(maxIdleConns int) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.MaxIdleConns = maxIdleConns
	}
}

func WithConnMaxLifetime(lifetime time.Duration) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.ConnMaxLifetime = lifetime
	}
}

func WithDebug(debug bool) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.Debug = debug
	}
}

func (c *ClickHouse) CreateDatabase(database string) error {
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", database)
	return c.conn.Exec(c.ctx, query)
}

func (c *ClickHouse) DB() driver.Conn {
	return c.conn
}

func NewClickHouse(opts ...ClickHouseOption) (*ClickHouse, error) {
	options := &clickhouse.Options{
		Debugf: func(format string, v ...any) {
			fmt.Printf(format, v)
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		// Set default values
		DialTimeout:          time.Second * 30,
		MaxOpenConns:         5,
		MaxIdleConns:         5,
		ConnMaxLifetime:      time.Duration(10) * time.Minute,
		ConnOpenStrategy:     clickhouse.ConnOpenInOrder,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10240,
		ClientInfo: clickhouse.ClientInfo{ // optional, please see Client info section in the README.md
			Products: []struct {
				Name    string
				Version string
			}{
				{Name: "my-app", Version: "0.1"},
			},
		},
	}

	// Apply any specified options
	ch := &ClickHouse{}

	for _, opt := range opts {
		opt(ch, options)
	}

	conn, err := clickhouse.Open(options)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(ch.ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, err
	}

	ch.conn = conn

	return ch, nil
}
