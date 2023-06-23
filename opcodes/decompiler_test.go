package opcodes

import (
	"context"
	"os"
	"testing"

	"github.com/txpull/unpack/clients"
	"github.com/txpull/unpack/fixtures"
	"github.com/txpull/unpack/helpers"

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

			bytes, err := helpers.GetBytecode(context.TODO(), clients, receipt.ContractAddress, nil)
			tAssert.NoError(err)
			tAssert.GreaterOrEqual(len(bytes), 1)

			decompiler := NewDecompiler(context.TODO(), bytes)
			tAssert.NoError(err)
			tAssert.NotNil(decompiler)
			tAssert.IsType(&Decompiler{}, decompiler)

			bSize := decompiler.GetBytecodeSize()
			tAssert.Equal(uint64(len(bytes)), bSize)

			err = decompiler.Decompile()
			tAssert.NoError(err)

			tAssert.GreaterOrEqual(len(decompiler.GetInstructions()), 1)
		}
	}
}

func TestDecompiler_Decompile(t *testing.T) {
	// Create a new Decompiler instance and set the bytecode.
	bytecode := []byte{
		byte(PUSH1), byte(0x01), // PUSH1 0x01
		byte(PUSH1), byte(0x02), // PUSH1 0x02
		byte(ADD), // ADD
	}

	decompiler := &Decompiler{
		bytecode:     bytecode,
		bytecodeSize: uint64(len(bytecode)),
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

func TestDecompiler_MatchInstruction(t *testing.T) {
	// Create a new Decompiler instance and set the instructions.
	decompiler := &Decompiler{
		instructions: []Instruction{
			{
				Offset: 0,
				OpCode: PUSH1,
				Args:   []byte{0x01},
			},
			{
				Offset: 1,
				OpCode: JUMP,
			},
			{
				Offset: 2,
				OpCode: ADD,
			},
		},
	}

	// Test matching of PUSH1 instruction
	instruction := Instruction{
		Offset: 0,
		OpCode: PUSH1,
		Args:   []byte{0x01},
	}
	matched := decompiler.MatchInstruction(instruction)
	assert.True(t, matched, "Expected PUSH1 instruction to match")

	// Test matching of JUMP instruction
	instruction = Instruction{
		Offset: 1,
		OpCode: JUMP,
	}
	matched = decompiler.MatchInstruction(instruction)
	assert.True(t, matched, "Expected JUMP instruction to match")

	// Test non-matching instruction
	instruction = Instruction{
		Offset: 2,
		OpCode: SSTORE,
	}
	matched = decompiler.MatchInstruction(instruction)
	assert.False(t, matched, "Expected SSTORE instruction to not match")
}

func TestDecompiler_GetInstructions(t *testing.T) {
	// Create a new Decompiler instance and set the instructions.
	decompiler := &Decompiler{
		instructions: []Instruction{
			{
				Offset: 0,
				OpCode: PUSH1,
				Args:   []byte{0x01},
			},
			{
				Offset: 1,
				OpCode: ADD,
			},
			{
				Offset: 2,
				OpCode: PUSH2,
				Args:   []byte{0x02, 0x03},
			},
		},
	}

	// Get the decompiled instructions
	decompiledInstructions := decompiler.GetInstructions()

	// Define the expected decompilation output
	expectedInstructions := []Instruction{
		{
			Offset: 0,
			OpCode: PUSH1,
			Args:   []byte{0x01},
		},
		{
			Offset: 1,
			OpCode: ADD,
		},
		{
			Offset: 2,
			OpCode: PUSH2,
			Args:   []byte{0x02, 0x03},
		},
	}

	// Compare the expected instructions with the decompiled instructions
	assert.Equal(t, expectedInstructions, decompiledInstructions, "Decompiled instructions should match the expected output")
}

func TestDecompiler_EdgeCases(t *testing.T) {
	// Edge case: Empty instruction set
	decompiler := &Decompiler{
		instructions: []Instruction{},
	}
	assert.Empty(t, decompiler.GetInstructions(), "Decompiled instructions should be empty")

	// Edge case: Invalid instruction opcode
	decompiler = &Decompiler{
		instructions: []Instruction{
			{
				Offset: 0,
				OpCode: OpCode(255), // Invalid opcode
				Args:   nil,
			},
		},
	}

	assert.NotEmpty(t, decompiler.GetInstructions(), "Decompiled instructions should be empty")

	// Edge case: Unusual instruction sequence
	decompiler = &Decompiler{
		instructions: []Instruction{
			{
				Offset: 0,
				OpCode: PUSH1,
				Args:   []byte{0x01},
			},
			{
				Offset: 1,
				OpCode: PUSH2,
				Args:   []byte{0x02, 0x03},
			},
			{
				Offset: 2,
				OpCode: ADD,
			},
		},
	}

	// Define the expected decompilation output
	expectedInstructions := []Instruction{
		{
			Offset: 0,
			OpCode: PUSH1,
			Args:   []byte{0x01},
		},
		{
			Offset: 1,
			OpCode: PUSH2,
			Args:   []byte{0x02, 0x03},
		},
		{
			Offset: 2,
			OpCode: ADD,
		},
		{
			Offset: 3,
			OpCode: JUMP,
		},
	}

	// Compare the expected instructions with the decompiled instructions
	assert.NotEqual(t, expectedInstructions, decompiler.GetInstructions(), "Decompiled instructions should not match the expected output")
}

func BenchmarkDecompiler_Performance(b *testing.B) {
	// Create a large instruction set for benchmarking
	instructions := make([]Instruction, 1000000)
	for i := 0; i < len(instructions); i++ {
		instructions[i] = Instruction{
			Offset: i,
			OpCode: PUSH1,
			Args:   []byte{0x01},
		}
	}

	// Create a byte slice representing the bytecode
	bytecode := make([]byte, len(instructions)*2)
	for i, instruction := range instructions {
		bytecode[i*2] = byte(instruction.OpCode)
		bytecode[i*2+1] = instruction.Args[0]
	}

	// Run the benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		decompiler := &Decompiler{
			bytecode:     bytecode,
			instructions: instructions,
		}
		decompiler.Decompile()
	}
}
