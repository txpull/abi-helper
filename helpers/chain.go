package helpers

import "math/big"

var ChainMappingInformation = map[int64]string{
	1:  "ethereum",
	56: "bsc",
}

func GetChainName(chainId *big.Int) string {
	return ChainMappingInformation[chainId.Int64()]
}
