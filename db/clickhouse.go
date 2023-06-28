package db

import (
	"context"
	"crypto/tls"
	"errors"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/txpull/unpack/options"
	"go.uber.org/zap"
)

// ClickHouse is a structure that encapsulates a ClickHouse database connection.
// It contains a context, options for the connection, and the connection itself.
type ClickHouse struct {
	ctx  context.Context
	opts options.ClickHouse
	conn driver.Conn
}

// DB is a method that returns the ClickHouse database connection encapsulated in the ClickHouse structure.
func (c *ClickHouse) DB() driver.Conn {
	return c.conn
}

// ValidateOptions checks the validity of the options used to create a ClickHouse connection.
// It checks if the required fields (Hosts, Database, Username) are set and if the numeric fields
// (MaxExecutionTime, DialTimeout, MaxOpenConns, MaxIdleConns, MaxConnLifetime) are greater than zero.
// If any of the checks fail, it returns an error with a message indicating which field is invalid.
// If all checks pass, it returns nil.
func (c *ClickHouse) ValidateOptions() error {
	if len(c.opts.Hosts) == 0 {
		return errors.New("at least one host must be set")
	}
	if c.opts.Database == "" {
		return errors.New("database must be set")
	}
	if c.opts.Username == "" {
		return errors.New("username must be set")
	}
	if c.opts.MaxExecutionTime <= 0 {
		return errors.New("max execution time must be greater than 0")
	}
	if c.opts.DialTimeout <= 0 {
		return errors.New("dial timeout must be greater than 0")
	}
	if c.opts.MaxOpenConns <= 0 {
		return errors.New("max open connections must be greater than 0")
	}
	if c.opts.MaxIdleConns < 0 {
		return errors.New("max idle connections must be greater than or equal to 0")
	}
	if c.opts.MaxConnLifetime <= 0 {
		return errors.New("max connection lifetime must be greater than 0")
	}
	return nil
}

// NewClickHouse is a function that creates a new ClickHouse database connection using the provided context and options.
// It configures the connection with the options, opens the connection, and checks its validity by pinging the database.
// If the connection is successfully opened and is valid, it returns a ClickHouse structure that encapsulates the connection.
// If the connection fails to open or is not valid, it returns an error.
func NewClickHouse(ctx context.Context, opts options.ClickHouse) (*ClickHouse, error) {
	options := &clickhouse.Options{
		Debug: opts.DebugEnabled,
		Settings: clickhouse.Settings{
			"max_execution_time": opts.MaxExecutionTime,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Addr: opts.Hosts,
		Auth: clickhouse.Auth{
			Username: opts.Username,
			Password: opts.Password,
			Database: opts.Database,
		},
		DialTimeout:          time.Second * opts.DialTimeout,
		MaxOpenConns:         opts.MaxOpenConns,
		MaxIdleConns:         opts.MaxIdleConns,
		ConnMaxLifetime:      opts.MaxConnLifetime * time.Minute,
		ConnOpenStrategy:     clickhouse.ConnOpenInOrder,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10240,
		Protocol:             clickhouse.Native,
		// TODO: Add support for TLS
		TLS: &tls.Config{InsecureSkipVerify: true},
	}

	client := &ClickHouse{ctx: ctx, opts: opts}

	if err := client.ValidateOptions(); err != nil {
		return nil, err
	}

	conn, err := clickhouse.Open(options)
	if err != nil {
		return nil, err
	}

	if err := conn.Ping(client.ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok && err.Error() != "EOF" {
			zap.L().Error(
				"Clickhouse raised exception",
				zap.Int32("code", exception.Code),
				zap.String("message", exception.Message),
				zap.String("stacktrace", exception.StackTrace),
			)

			return nil, err
		}
	}

	client.conn = conn

	return client, nil
}
