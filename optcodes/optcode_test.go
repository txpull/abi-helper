package optcodes_test

import (
	"context"
	"fmt"
	"github/txpull/abi-helper/clients"
	"github/txpull/abi-helper/fixtures"
	"github/txpull/abi-helper/optcodes"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestOptcode_DiscoverAndDecompiler(t *testing.T) {
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

			decompiler, err := optcodes.NewDecompiler(
				context.TODO(),
				clients,
				receipt.ContractAddress,
			)
			tAssert.NoError(err)
			tAssert.NotNil(decompiler)
			tAssert.IsType(&optcodes.Decompiler{}, decompiler)

			bytecode, err := decompiler.DiscoverContractBytecode()
			tAssert.NoError(err)
			tAssert.IsType([]byte{}, bytecode)

			err = decompiler.Decompile()
			tAssert.NoError(err)

			tAssert.GreaterOrEqual(len(decompiler.GetInstructions()), 1)

			for _, instr := range decompiler.GetInstructionsByOpCode(optcodes.CALL) {
				// Print the execution flow tree for the CALL opcode
				decompiler.PrintInstructionTree(instr)
				fmt.Println("---------------------------------")
			}

		}
	}
}
