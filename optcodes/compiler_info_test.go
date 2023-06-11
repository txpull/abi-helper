package optcodes

import (
	"context"

	"github.com/txpull/abi-helper/clients"
	"github.com/txpull/abi-helper/fixtures"

	"os"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestOptcode_DiscoverContractInfo(t *testing.T) {
	tAssert := assert.New(t)

	ethReader := fixtures.GetEthReaderForTest(tAssert)

	// We are going to use actual real network to fetch necessary information to complete test.
	// You can find free rpc urls at:
	// BSC - https://chainlist.org/chain/56
	// ETH - https://chainlist.org/chain/1
	// Necessary for now are: bytecode,
	clients, err := clients.NewEthClients(os.Getenv("TEST_ETH_NODE_URL"), 1)
	tAssert.NoError(err)
	tAssert.NotNil(clients)

	// Deliberately looping through the transactions instead of just looping through
	// receipts in search for non 0x0 ContractAddress
	for _, tx := range ethReader.GetTransactions() {
		if tx.To() == nil {
			receipt, found := ethReader.GetReceiptFromTxHash(tx.Hash())
			tAssert.True(found)
			tAssert.NotNil(receipt)
			tAssert.IsType(&types.Receipt{}, receipt)

			t.Logf("Discovered contract address: %v", receipt.ContractAddress)

			decompiler, err := NewDecompiler(
				context.TODO(),
				clients,
				receipt.ContractAddress,
			)
			tAssert.NoError(err)
			tAssert.NotNil(decompiler)
			tAssert.IsType(&Decompiler{}, decompiler)

			bytecode, err := decompiler.DiscoverContractBytecode()
			tAssert.NoError(err)
			tAssert.IsType([]byte{}, bytecode)

			err = decompiler.Decompile()
			tAssert.NoError(err)

			tAssert.GreaterOrEqual(len(decompiler.GetInstructions()), 1)

			debugInfo, err := decompiler.ExtractDebugInfo()
			tAssert.NoError(err)

			_ = debugInfo
		}
	}
}
