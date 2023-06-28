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

// Nodes is a struct that holds the full and archive nodes settings.
type Nodes struct {
	Full    Node `mapstructure:"full"`
	Archive Node `mapstructure:"archive"`
}

// Node is a struct that holds the URL and the number of concurrent clients for a node.
type Node struct {
	URL                     string `mapstructure:"url"`
	ConcurrentClientsNumber int    `mapstructure:"concurrent_clients_number"`
}

// Fixtures is a struct that holds the generator settings.
type Fixtures struct {
	Generator Generator `mapstructure:"generator"`
}

// Generator is a struct that holds the Ethereum and Binance generator settings.
type Generator struct {
	Ethereum EthereumGenerator `mapstructure:"ethereum"`
	Binance  BinanceGenerator  `mapstructure:"binance"`
}

// EthereumGenerator is a struct that holds the start and end block numbers for the Ethereum generator.
type EthereumGenerator struct {
	StartBlockNumber int `mapstructure:"start_block_number"`
	EndBlockNumber   int `mapstructure:"end_block_number"`
}

// BinanceGenerator is a struct that holds the start and end block numbers for the Binance generator.
type BinanceGenerator struct {
	StartBlockNumber int `mapstructure:"start_block_number"`
	EndBlockNumber   int `mapstructure:"end_block_number"`
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
	Clickhouse Clickhouse `mapstructure:"clickhouse"`
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
type Clickhouse struct {
	Host            string `mapstructure:"host"`
	Database        string `mapstructure:"database"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	CertificatePath string `mapstructure:"certificate_path"`
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
