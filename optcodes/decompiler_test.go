package optcodes

import (
	"context"
	"github/txpull/abi-helper/clients"
	"github/txpull/abi-helper/fixtures"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

func TestOptcode_DiscoverAndDecompile(t *testing.T) {
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
		}
	}
}

func TestDecompiler_Decompile(t *testing.T) {
	// Create a new Decompiler instance and set the bytecode.
	decompiler := &Decompiler{
		bytecode: []byte{
			byte(PUSH1), byte(0x01), // PUSH1 0x01
			byte(PUSH1), byte(0x02), // PUSH1 0x02
			byte(ADD), // ADD
		},
	}

	err := decompiler.Decompile()
	assert.NoError(t, err)

	instructions := decompiler.GetInstructions()
	assert.Len(t, instructions, 3)

	assert.Equal(t, OpCode(PUSH1), instructions[0].OpCode)
	assert.Equal(t, OpCode(PUSH1), instructions[1].OpCode)
	assert.Equal(t, OpCode(ADD), instructions[2].OpCode)
}

func TestDecompiler_MatchFunctionSignature(t *testing.T) {
	// Create a new Decompiler instance and set the instructions.
	decompiler := &Decompiler{
		instructions: []Instruction{
			{
				OpCode: CALL,
				Args:   []byte{0x01, 0x02, 0x03, 0x04},
			},
			{
				OpCode: CALL,
				Args:   []byte{0x05, 0x06, 0x07, 0x08},
			},
		},
	}

	signature := "0x01020304"
	assert.True(t, decompiler.MatchFunctionSignature(signature))

	signature = "0x05060708"
	assert.True(t, decompiler.MatchFunctionSignature(signature))

	signature = "0x01020305" // Signature not present
	assert.False(t, decompiler.MatchFunctionSignature(signature))
}
