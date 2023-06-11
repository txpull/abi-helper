# abi-helper

**WARNING: Work In Progress!**

Set of tools to help decoding bytecode, transactions, blocks, receipts and logs for Ethereum based chains.


## Supported Chains

- BSC (Binance Smart Chain)
- Ethereum

## Features

- [x] Automatic fixtures (sample data) download direcly from (bsc|eth) mainnet used for tests.
- [x] Optcode discovery from any contract bytecode that is deployed on the (bsc|eth) mainnets.

## TODO

- [] Bytecode ABI decompiler (Solidity)


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
pkg: github/txpull/abi-helper/optcodes
cpu: AMD Ryzen Threadripper 3960X 24-Core Processor 
BenchmarkDecompiler_Performance
BenchmarkDecompiler_Performance-48    	       7	 184002805 ns/op	288450341 B/op	       8 allocs/op
PASS
```