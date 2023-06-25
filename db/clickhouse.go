package db

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// ClickHouse struct represents a ClickHouse database connection.
type ClickHouse struct {
	ctx  context.Context
	conn driver.Conn
}

// ClickHouseOption represents a function that applies a specific configuration to a ClickHouse database connection.
type ClickHouseOption func(*ClickHouse, *clickhouse.Options)

// WithCtx sets the context for the ClickHouse database connection.
func WithCtx(ctx context.Context) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		c.ctx = ctx
	}
}

// WithHost sets the host for the ClickHouse database connection.
func WithHost(host string) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.Addr = []string{host}
	}
}

// WithDatabase sets the database for the ClickHouse database connection.
func WithDatabase(database string) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.Auth.Database = database
	}
}

// WithUsername sets the username for the ClickHouse database connection.
func WithUsername(username string) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.Auth.Username = username
	}
}

// WithPassword sets the password for the ClickHouse database connection.
func WithPassword(password string) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.Auth.Password = password
	}
}

// WithDialTimeout sets the dial timeout for the ClickHouse database connection.
func WithDialTimeout(timeout time.Duration) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.DialTimeout = timeout
	}
}

// WithMaxOpenConns sets the maximum number of open connections for the ClickHouse database connection.
func WithMaxOpenConns(maxConns int) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.MaxOpenConns = maxConns
	}
}

// WithMaxIdleConns sets the maximum number of idle connections for the ClickHouse database connection.
func WithMaxIdleConns(maxIdleConns int) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.MaxIdleConns = maxIdleConns
	}
}

// WithConnMaxLifetime sets the maximum lifetime of a connection for the ClickHouse database connection.
func WithConnMaxLifetime(lifetime time.Duration) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.ConnMaxLifetime = lifetime
	}
}

// WithDebug sets the debug mode for the ClickHouse database connection.
func WithDebug(debug bool) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.Debug = debug
	}
}

// WithTLS sets the TLS configuration for the ClickHouse database connection.
func WithTLS(t *tls.Config) ClickHouseOption {
	return func(c *ClickHouse, o *clickhouse.Options) {
		o.TLS = t
	}
}

// CreateDatabase creates a new database in the ClickHouse database connection.
// It returns an error if the execution of the query fails.
func (c *ClickHouse) CreateDatabase(database string) error {
	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", database)
	return c.conn.Exec(c.ctx, query)
}

// DB returns the ClickHouse database connection.
func (c *ClickHouse) DB() driver.Conn {
	return c.conn
}

// NewClickHouse creates a new ClickHouse database connection with the specifiedoptions.
// It returns a ClickHouse database connection and an error if the connection fails.
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
		MaxOpenConns:         1,
		MaxIdleConns:         1,
		ConnMaxLifetime:      time.Duration(10) * time.Minute,
		ConnOpenStrategy:     clickhouse.ConnOpenInOrder,
		BlockBufferSize:      10,
		MaxCompressionBuffer: 10240,
		Protocol:             clickhouse.Native,
		TLS:                  &tls.Config{InsecureSkipVerify: true},
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
		if exception, ok := err.(*clickhouse.Exception); ok && err.Error() != "EOF" {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
			return nil, err
		}
	}

	ch.conn = conn

	return ch, nil
}
