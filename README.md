# bytecode

**WARNING: Work In Progress!**

Set of tools to help decoding bytecode, transactions, blocks, receipts and logs for Ethereum based chains.


## Supported Chains

- [x] BSC (Binance Smart Chain)
- [] Ethereum
- [] Polygon

## Features

- [x] Automatic fixtures (sample data) download direcly from (bsc|eth) mainnet used for tests.
- [x] Optcode discovery from any contract bytecode that is deployed on the (bsc|eth) mainnets.
- [x] Reading signature database.
- [x] Ability to get optcode for any transaction that contains appropriate data.
- [x] Ability to potentially get transaction method id and arguments from contracts without abi.
- [x] 3rd party contract code scanners (BscScan)

## TODO

- [] Extract compiler information from transaction contract creation data if available.
- [] Extract contract deployment constructor information.
- [] Extract contract swarm ipfs/bzz information.
- [] Extract contract abi from (3rd-party, metadata, reverse engineering bytecode)
- [] Extend signatures to download new signatures from 4byte.dictionary and other services including parsing abis and writing signatures from abis.


## BUGS

Just a list for me to fix it while developing without opening tickets

- [] Argument decoding works to a point, should be fixed

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