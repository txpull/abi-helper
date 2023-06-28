// Package options provides a way to manage global options settings.
package options

import "time"

// Options is a struct that holds the global options settings.
type Options struct {
	Networks Networks `mapstructure:"networks"`
	Fixtures Fixtures `mapstructure:"fixtures"`
	Clients  Clients  `mapstructure:"clients"`
	Database Database `mapstructure:"database"`
	Syncers  Syncers  `mapstructure:"syncers"`
}

// Networks is a struct that holds the network nodes settings.
type Networks struct {
	Ethereum Nodes `mapstructure:"ethereum"`
	Binance  Nodes `mapstructure:"binance"`
}

// GetNode returns the node settings for a given network and node.
// Useful for the configuration and quick access to the nodes settings.
func (o *Options) GetNode(network string, node string) Node {
	switch network {
	case "ethereum":
		switch node {
		case "full":
			return o.Networks.Ethereum.FullNode
		case "archive":
			return o.Networks.Ethereum.ArchiveNode
		}
	case "binance":
		switch node {
		case "full":
			return o.Networks.Binance.FullNode
		case "archive":
			return o.Networks.Binance.ArchiveNode
		}
	}

	return Node{}
}

// Nodes is a struct that holds the full and archive nodes settings.
type Nodes struct {
	FullNode    Node `mapstructure:"full"`
	ArchiveNode Node `mapstructure:"archive"`
}

// Node is a struct that holds the URL and the number of concurrent clients for a node.
type Node struct {
	URL                     string `mapstructure:"url"`
	ConcurrentClientsNumber int    `mapstructure:"concurrent_clients_number"`
}

// Fixtures is a struct that holds the generator settings.
type Fixtures struct {
	Network          string `mapstructure:"network"`
	NodeType         string `mapstructure:"node_type"`
	FixturesPath     string `mapstructure:"fixtures_path"`
	StartBlockNumber uint64 `mapstructure:"start_block_number"`
	EndBlockNumber   uint64 `mapstructure:"end_block_number"`
}

// Clients is a struct that holds the Bscscan and Bitquery client settings.
type Clients struct {
	Bscscan  BscscanClient  `mapstructure:"bscscan"`
	Bitquery BitqueryClient `mapstructure:"bitquery"`
}

// BscscanClient is a struct that holds the API settings for the Bscscan client.
type BscscanClient struct {
	API ClientInfo `mapstructure:"api"`
}

// BitqueryClient is a struct that holds the API settings for the Bitquery client.
type BitqueryClient struct {
	API ClientInfo `mapstructure:"api"`
}

// ClientInfo is a struct that holds the URL and key for a client's API.
type ClientInfo struct {
	URL string `mapstructure:"url"`
	Key string `mapstructure:"key"`
}

// Database is a struct that holds the Redis and Clickhouse database settings.
type Database struct {
	Redis      Redis      `mapstructure:"redis"`
	Clickhouse ClickHouse `mapstructure:"clickhouse"`
}

// Redis is a struct that holds the settings for a Redis database.
type Redis struct {
	Addr            string        `mapstructure:"addr"`
	Password        string        `mapstructure:"password"`
	DB              int           `mapstructure:"db"`
	MaxRetries      int           `mapstructure:"max_retries"`
	MinRetryBackoff time.Duration `mapstructure:"min_retry_backoff_ms"`
	MaxRetryBackoff time.Duration `mapstructure:"max_retry_backoff_ms"`
}

// Clickhouse is a struct that holds the settings for a Clickhouse database.
type ClickHouse struct {
	DebugEnabled     bool          `mapstructure:"debug_enabled"`
	Hosts            []string      `mapstructure:"hosts"`
	Database         string        `mapstructure:"database"`
	Username         string        `mapstructure:"username"`
	Password         string        `mapstructure:"password"`
	MaxExecutionTime int           `mapstructure:"max_execution_time"`
	DialTimeout      time.Duration `mapstructure:"dial_timeout"`
	MaxOpenConns     int           `mapstructure:"max_open_conns"`
	MaxIdleConns     int           `mapstructure:"max_idle_conns"`
	MaxConnLifetime  time.Duration `mapstructure:"max_conn_lifetime_m"`
}

// Syncers is a struct that holds the settings for different syncers.
type Syncers struct {
	Fourbyte FourbyteSyncer `mapstructure:"fourbyte"`
	Bscscan  BscscanSyncer  `mapstructure:"bscscan"`
	Sourcify SourcifySyncer `mapstructure:"sourcify"`
}

// FourbyteSyncer is a struct that holds the settings for a Fourbyte syncer.
type FourbyteSyncer struct {
	URL               string `mapstructure:"url"`
	WriteToClickhouse bool   `mapstructure:"write_to_clickhouse"`
	ChainID           int    `mapstructure:"chain_id"`
}

// BscscanSyncer is a struct that holds the settings for a Bscscan syncer.
type BscscanSyncer struct {
	VerifiedContractsPath string `mapstructure:"verified_contracts_path"`
	WriteToClickhouse     bool   `mapstructure:"write_to_clickhouse"`
}

// SourcifySyncer is a struct that holds the settings for a Sourcify syncer.
type SourcifySyncer struct {
	URL               string `mapstructure:"url"`
	MaxRetries        int    `mapstructure:"max_retries"`
	RetryDelay        int    `mapstructure:"retry_dely"`
	RateLimitS        int    `mapstructure:"rate_limit_s"`
	WriteToClickhouse bool   `mapstructure:"write_to_clickhouse"`
	ChainIDs          []int  `mapstructure:"chain_ids"`
}
