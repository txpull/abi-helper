package transactions

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/txpull/bytecode/opcodes"
	"github.com/txpull/bytecode/signatures"
	"go.uber.org/zap"
)

type MethodArgument struct {
	Name  string
	Index int64
	Type  string
	Value interface{}
}

// Decompiler decompiles Ethereum compatible transactions into EVM opcode instructions.
type Decompiler struct {
	ctx        context.Context
	tx         *types.Transaction
	decompiler *opcodes.Decompiler
	signatures *signatures.SignaturesReader
}

// NewDecompiler creates a new instance of the Decompiler struct.
// It takes a context and a types.Transaction object as input parameters.
// It returns a pointer to the created Decompiler object and an error.
func NewDecompiler(ctx context.Context, tx *types.Transaction, s *signatures.SignaturesReader) (*Decompiler, error) {
	opcode := opcodes.NewDecompiler(ctx, tx.Data())

	return &Decompiler{
		ctx:        ctx,
		tx:         tx,
		decompiler: opcode,
		signatures: s,
	}, nil
}

// GetMethodIdBytes returns the method ID bytes extracted from the transaction data.
// If the transaction data is empty, it returns an empty byte slice.
// Otherwise, it returns the first 4 bytes of the transaction data, which represent the method ID.
func (d *Decompiler) GetMethodIdBytes() []byte {
	if len(d.tx.Data()) == 0 {
		return []byte{}
	}

	return d.tx.Data()[:4]
}

// GetMethodIdHex returns the hexadecimal representation of the method ID
// extracted from the transaction data.
//
// If the transaction data is empty, an empty string is returned.
// The leading "0x" prefix is stripped from the returned string.
//
// The method ID is obtained by taking the first 4 bytes of the transaction data.
//
// Returns:
// - string: The hexadecimal representation of the method ID without the "0x" prefix.
func (d *Decompiler) GetMethodIdHex() string {
	if len(d.tx.Data()) == 0 {
		return ""
	}

	// Extract the first 4 bytes of the transaction data
	methodID := d.tx.Data()[:4]

	// Convert the method ID bytes to hexadecimal representation
	return common.Bytes2Hex(methodID)
}

// GetMethodArgsBytes returns the method arguments bytes extracted from the transaction data.
// If the transaction data is empty, it returns an empty byte slice.
// Otherwise, it returns the bytes of the transaction data starting from the fifth byte (index 4),
// which represent the method arguments.
func (d *Decompiler) GetMethodArgsBytes() []byte {
	if len(d.tx.Data()) == 0 {
		return []byte{}
	}

	return d.tx.Data()[4:]
}

func (d *Decompiler) GetMethodArgsFromSignature(s *signatures.Signature) map[int64]string {
	return extractArgumentTypes(s.Text)
}

func (d *Decompiler) GetDataSlice() []string {
	data := d.tx.Data()

	if len(data) == 0 {
		return nil
	}

	var toReturn []string

	for i := 4; i+32 <= len(data); i += 32 {
		value := data[i : i+32]
		toReturn = append(toReturn, common.Bytes2Hex(value))
	}

	return toReturn
}

// DiscoverSignature looks up the signature based on the method ID extracted from the transaction data.
// It returns a *signatures.Signature if found, or an error if the signature is not found or an error occurs.
func (d *Decompiler) DiscoverSignature() (*signatures.Signature, bool, error) {
	methodID := d.GetMethodIdHex()
	if methodID == "" {
		return nil, false, ErrEmptyMethodId
	}

	signature, found, err := d.signatures.LookupByHex(methodID)
	if err != nil {
		return nil, false, err
	}

	if signature == nil {
		return nil, false, ErrSignatureNotFound{Hex: methodID}
	}

	return signature, found, nil
}

func (d *Decompiler) DiscoverSignatureArguments(signature *signatures.Signature) ([]MethodArgument, error) {
	if signature == nil {
		return nil, ErrArgSignatureIsRequired
	}

	argTypes := d.GetMethodArgsFromSignature(signature)

	if len(argTypes) < 1 {
		return nil, nil
	}

	toReturn := []MethodArgument{}

	index := int64(0)
	for argIndex, argType := range argTypes {
		argValue := extractValue(argType, argIndex, d.GetMethodArgsBytes())
		toReturn = append(toReturn, MethodArgument{
			Index: index,
			Type:  argType,
			Value: argValue,
		})
		index++
	}

	fmt.Printf("Arguments: %+v \n", toReturn)
	fmt.Println()

	return toReturn, nil
}

// Decompile decompiles the transaction's bytecode into EVM opcode instructions.
// If there is an error during the decompilation process, it logs the error and returns it.
// Otherwise, it returns nil.
func (d *Decompiler) Decompile() error {
	if err := d.decompiler.Decompile(); err != nil {
		zap.L().Error(
			"failed to decode transaction opcodes",
			zap.String("tx_hash", d.tx.Hash().Hex()),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// GetInstructions returns the decompiled EVM opcode instructions.
func (d *Decompiler) GetInstructions() []opcodes.Instruction {
	return d.decompiler.GetInstructions()
}

// GetOptDecompiler returns the underlying optcodes.Decompiler instance.
func (d *Decompiler) GetOptDecompiler() *opcodes.Decompiler {
	return d.decompiler
}
