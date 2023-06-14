package opcodes

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"strings"

	"github.com/ethereum/go-ethereum/common"
)

var (
	ErrEmptyBytecode = errors.New("bytecode is not set or empty bytecode provided")
)

// Decompiler decompiles bytecode into optcode instructions.
type Decompiler struct {
	ctx          context.Context
	bytecode     []byte
	bytecodeSize uint64
	instructions []Instruction
}

// Instruction represents an optcode instruction.
type Instruction struct {
	Offset int
	OpCode OpCode
	Args   []byte
}

func (d *Decompiler) SetBytecode(b []byte) {
	d.bytecode = b
	d.bytecodeSize = uint64(len(b))
}

func (d *Decompiler) GetBytecode() []byte {
	return d.bytecode
}

func (d *Decompiler) GetBytecodeSize() uint64 {
	return d.bytecodeSize
}

func (d *Decompiler) Decompile() error {
	if d.bytecodeSize < 1 {
		return ErrEmptyBytecode
	}

	offset := 0
	for offset < len(d.bytecode) {
		op := OpCode(d.bytecode[offset])
		instruction := Instruction{
			Offset: offset,
			OpCode: op,
			Args:   []byte{},
		}

		// Check if the opcode is a PUSH instruction.
		if op.IsPush() {
			argSize := int(op) - int(PUSH1) + 1
			if offset+argSize >= len(d.bytecode) {
				break
			}
			instruction.Args = d.bytecode[offset+1 : offset+argSize+1]
			offset += argSize
		}

		d.instructions = append(d.instructions, instruction)
		offset++
	}
	return nil
}

// MatchFunctionSignature checks if a given function signature matches any of the decompiled instructions.
func (d *Decompiler) MatchFunctionSignature(signature string) bool {
	// Remove "0x" prefix if present
	signature = strings.TrimPrefix(signature, "0x")

	for _, instruction := range d.instructions {
		if instruction.OpCode == CALL && len(instruction.Args) >= 4 {
			functionSig := common.Bytes2Hex(instruction.Args[:4])
			if functionSig == signature {
				return true
			}
		}
	}
	return false
}

// GetCallInstructions returns all instructions related to the CALL opcode.
func (d *Decompiler) GetInstructionsByOpCode(op OpCode) []Instruction {
	var callInstructions []Instruction
	for _, instruction := range d.instructions {
		if instruction.OpCode == op {
			callInstructions = append(callInstructions, instruction)
		}
	}
	return callInstructions
}

// GetInstructions returns the decompiled optcode instructions.
func (d *Decompiler) GetInstructions() []Instruction {
	return d.instructions
}

// String returns a string representation of the decompiled bytecode as a sequence of bytes.
func (d *Decompiler) String() string {
	var buf bytes.Buffer

	for _, instr := range d.instructions {
		offset := fmt.Sprintf("0x%04x", instr.Offset)
		opCode := instr.OpCode.String()

		buf.WriteString(offset + " " + opCode)

		if len(instr.Args) > 0 {
			buf.WriteString(" " + common.Bytes2Hex(instr.Args))
		}

		buf.WriteString("\n")
	}

	return buf.String()
}

// IsOpCode checks if the given instruction represents the specified opcode.
func (d *Decompiler) IsOpCode(instruction Instruction, op OpCode) bool {
	return instruction.OpCode == op
}

// OpCodeFound checks if the specified opcode is found in any of the decompiled instructions.
func (d *Decompiler) OpCodeFound(op OpCode) bool {
	for _, instruction := range d.instructions {
		if instruction.OpCode == op {
			return true
		}
	}
	return false
}

// IsSelfDestruct checks if the decompiled instructions contain the SELFDESTRUCT opcode.
func (d *Decompiler) IsSelfDestruct() bool {
	return d.OpCodeFound(SELFDESTRUCT)
}

// isControlFlowInstruction checks if the given opcode represents a control flow instruction.
func (d *Decompiler) IsControlFlowInstruction(op OpCode) bool {
	return op == JUMP || op == JUMPI || op == JUMPDEST || op == RETURN || op == REVERT || op == SELFDESTRUCT
}

func (d *Decompiler) MatchInstruction(instruction Instruction) bool {
	for _, inst := range d.instructions {
		if inst.Offset == instruction.Offset && inst.OpCode == instruction.OpCode && bytes.Equal(inst.Args, instruction.Args) {
			return true
		}
	}
	return false
}

// NewDecompiler creates a new Decompiler instance.
// @TODO: Add automatic decompile in the future from NewDecompiler()
func NewDecompiler(ctx context.Context, b []byte) (*Decompiler, error) {
	return &Decompiler{
		ctx:          ctx,
		bytecode:     b,
		bytecodeSize: uint64(len(b)),
		instructions: []Instruction{},
	}, nil
}
