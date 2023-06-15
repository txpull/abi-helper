package transactions

import (
	"context"
	"log"
	"math/big"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/txpull/bytecode/clients"
	"github.com/txpull/bytecode/fixtures"
	"github.com/txpull/bytecode/signatures"
)

func TestTransaction_DiscoverAndDecompile(t *testing.T) {
	tAssert := assert.New(t)
	ctx := context.TODO()

	ethReader := fixtures.GetEthReaderForTest(tAssert)

	// We are going to use actual real network to fetch necessary information to complete test.
	// You can find free rpc urls at:
	// BSC - https://chainlist.org/chain/56
	// ETH - https://chainlist.org/chain/1
	// Necessary for now are: bytecode,
	clients, err := clients.NewEthClients(os.Getenv("TEST_ETH_NODE_URL"), 1)
	tAssert.NoError(err)
	tAssert.NotNil(clients)

	signaturesDb, err := signatures.NewDb(ctx, os.Getenv("TEST_SIGNATURES_DB_PATH"))
	tAssert.NoError(err)
	defer signaturesDb.Close()

	i := 0
	txns := ethReader.GetTransactions()
	// Deliberately looping through the transactions instead of just looping through
	// receipts in search for non 0x0 ContractAddress
	for _, tx := range txns {
		if i == 1 {
			break
		}

		log.Printf("Processing: %s \n", tx.Hash().Hex())
		if tx.To() != nil {
			receipt, found := ethReader.GetReceiptFromTxHash(tx.Hash())
			tAssert.True(found)
			tAssert.NotNil(receipt)
			tAssert.IsType(&types.Receipt{}, receipt)

			decompiler, err := NewDecompiler(ctx, tx, signaturesDb)
			tAssert.NoError(err)
			tAssert.NotNil(decompiler)
			tAssert.IsType(&Decompiler{}, decompiler)

			methodId := decompiler.GetMethodIdBytes()
			tAssert.GreaterOrEqual(len(methodId), 0)

			methodArgs := decompiler.GetMethodArgsBytes()
			tAssert.GreaterOrEqual(len(methodArgs), 0)

			if len(methodId) > 0 {
				methodIdHex := decompiler.GetMethodIdHex()
				tAssert.Equal(methodIdHex, common.Bytes2Hex(methodId))

				signature, found, err := decompiler.DiscoverSignature()
				if err == nil && found {
					tAssert.IsType(&signatures.Signature{}, signature)
					t.Logf("Signature: %s", signature.TextSignature)

					args := decompiler.GetMethodArgsFromSignature(signature)
					for _, arg := range args {
						t.Logf("Signature Arg: %s \n", arg)
					}

					decompiler.DiscoverSignatureArguments(signature)

				}
			}

			if len(tx.Data()) > 0 {
				err = decompiler.Decompile()
				tAssert.NoError(err)

				if len(decompiler.GetInstructions()) > 0 {
					opcode := decompiler.GetOptDecompiler()
					tAssert.Greater(len(opcode.String()), 1)
				}

				for _, slice := range decompiler.GetDataSlice() {
					t.Logf("%v \n", slice)
				}
			}
		}

		i = i + 1
	}
}

func TestDecompiler_GetMethodIdBytes(t *testing.T) {
	transaction := createTestTransaction()
	decompiler := createTestDecompiler(transaction)

	methodIDBytes := decompiler.GetMethodIdBytes()

	assert.Equal(t, []byte{0x12, 0x34, 0x56, 0x78}, methodIDBytes)
}

func TestDecompiler_GetMethodIdHex(t *testing.T) {
	transaction := createTestTransaction()
	decompiler := createTestDecompiler(transaction)

	methodIDHex := decompiler.GetMethodIdHex()

	assert.Equal(t, "12345678", methodIDHex)
}

func TestDecompiler_GetMethodArgsBytes(t *testing.T) {
	transaction := createTestTransaction()
	decompiler := createTestDecompiler(transaction)

	methodArgsBytes := decompiler.GetMethodArgsBytes()

	assert.Equal(t, []byte{0x9a, 0xbc, 0xde, 0xf0}, methodArgsBytes)
}

func TestDecompiler_Decompile(t *testing.T) {
	transaction := createTestTransaction()
	decompiler := createTestDecompiler(transaction)

	err := decompiler.Decompile()

	assert.NoError(t, err)
}

func TestDecompiler_GetInstructions(t *testing.T) {
	transaction := createTestTransaction()
	decompiler := createTestDecompiler(transaction)

	instructions := decompiler.GetInstructions()
	assert.NotNil(t, instructions)
}

func TestDecompiler_GetOptDecompiler(t *testing.T) {
	transaction := createTestTransaction()
	decompiler := createTestDecompiler(transaction)

	optDecompiler := decompiler.GetOptDecompiler()
	assert.NotNil(t, optDecompiler)
}

// Helper function to create a test transaction
func createTestTransaction() *types.Transaction {
	data := []byte{0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0}
	return types.NewTransaction(big.NewInt(1).Uint64(), common.HexToAddress("0x1234567890abcdef"), big.NewInt(1000000), 21000, big.NewInt(1), data)
}

// Helper function to create a test Decompiler instance
func createTestDecompiler(tx *types.Transaction) *Decompiler {
	ctx := context.TODO()
	decompiler, _ := NewDecompiler(ctx, tx, nil)
	return decompiler
}
