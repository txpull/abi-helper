# bytecode

**WARNING: Work In Progress!**

Set of tools to help decoding bytecode, transactions, blocks, receipts and logs for Ethereum based chains.


## Supported Chains

- [x] BSC (Binance Smart Chain)
- [ ] Ethereum
- [ ] Polygon

## Features

- [x] Automatic fixtures (sample data) download direcly from (bsc|eth) mainnet used for tests.
- [x] Opcode discovery from any contract bytecode that is deployed on the (bsc|eth) mainnets.
- [x] Reading signature database.
- [x] Ability to get opcode for any transaction that contains appropriate data.
- [x] Ability to potentially get transaction method id and arguments from contracts without abi.
- [x] 3rd party contract code scanners (BscScan)
- [x] Commands and utilities that can download and read verified contracts from bscscan.

## TODO

- [ ] Extract compiler information from transaction contract creation data if available.
- [ ] Extract contract deployment constructor information.
- [ ] Extract contract swarm ipfs/bzz information.
- [ ] Extract contract abi from (3rd-party, metadata, reverse engineering bytecode)
- [ ] Extend signatures to download new signatures from 4byte.dictionary and other services including parsing abis and writing signatures from abis.


## BUGS

Just a list for me to fix it while developing without opening tickets

- [ ] Argument decoding works to a point, should be fixed


## Installation

To install the bytecode, follow these steps:

1. Clone the repository: `git clone https://github.com/txpull/bytecode.git`
2. Navigate to the project directory: `cd bytecode`
3. Build the binary: `make install`

## Usage

The `bytecode` provides the following commands:

### Command: txbyte

This is the base command for the application.

Usage: `txbyte txbyte`

### Command: version

Displays the current version of the application.

Usage: `txbyte version`

### Command: generate-eth

Generates Ethereum-based fixtures and writes them into (block|transactions|receipt).gob files.

Usage: `txbyte fixtures generate-eth --eth.node.url <node-url> --eth.node.concurrent_clients_number <num> --eth.generator.start_block_number <start-block> --eth.generator.end_block_number <end-block>`

Replace the following parameters:
- `<node-url>`: Ethereum-based node full URL (example: https://node-url:port).
- `<num>`: Number of concurrent node clients to spawn.
- `<start-block>`: Start block number for generating fixtures.
- `<end-block>`: End block number for generating fixtures.

### Command: syncer

Commands related to syncing data from third-party sources.

Usage: `txbyte syncer <subcommand>`

Replace `<subcommand>` with one of the following:
- `bscscan`: Process verified contracts from bscscan.
- `fourbyte`: Download, process, and store signatures from 4byte.directory.

### Command: fixtures

Commands related to obtaining unit test data.

Usage: `txbyte fixtures <subcommand>`

Replace `<subcommand>` with one of the following:
- `generate-eth`: Generate Ethereum-based fixtures and write them into (block|transactions|receipt).gob files.

## Configuration

The Application Name can be configured using a config file (default: `.txbyte.yaml`) and environment variables. The following configuration options are available:

- `config`: Path to the config file (default: `$HOME/.txbyte.yaml`).
- `bsc.crawler.bscscan_path`: Path for storing BscScan data (default: `data/bscscan`).
- `bscscan.api.url`: BscScan API URL.
- `bscscan.api.key`: BscScan API key.
- `eth.generator.fixtures_path`: Path for storing Ethereum fixtures (default: `data/fixtures`).
- `eth.node.url`: Ethereum-based node full URL.
- `eth.node.concurrent_clients_number`: Number of concurrent node clients to spawn.
- `eth.generator.start_block_number`: Start block number for generating fixtures.
- `eth.generator.end_block_number`: End block number for generating fixtures.

Refer to the config file or environment variables for configuring the Application Name.


## Testing

### Public Node URLS

In order to run tests successfully you will need to have node (not archive node) access urls to the
ethereum and/or bsc network. If you don't have your own node, you can find free nodes at:

- Binance Smart Chain: https://chainlist.org/chain/56
- Ethereum: https://chainlist.org/chain/1


### Running Tests

In order to run tests, get coverage of the tested code, you can simply run:

```sh
make test
```

### Benchmarks

```
goos: linux
goarch: amd64
pkg: github/txpull/bytecode/optcodes
cpu: AMD Ryzen Threadripper 3960X 24-Core Processor 
BenchmarkDecompiler_Performance
BenchmarkDecompiler_Performance-48    	       7	 184002805 ns/op	288450341 B/op	       8 allocs/op
PASS
```

## License

The Application Name is licensed under the MIT License. See the `LICENSE` file for more details.