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
	// ErrEmptyBytecode is an error that indicates the absence of bytecode.
	ErrEmptyBytecode = errors.New("bytecode is not set or empty bytecode provided")
)

// Decompiler decompiles bytecode into opcode instructions.
// The bytecode to be decompiled can be set using the SetBytecode() method.
// Decompiling is performed by the Decompile() method. The results can be
// obtained using GetInstructions(), String(), and other related methods.
type Decompiler struct {
	ctx          context.Context
	bytecode     []byte
	bytecodeSize uint64
	instructions []Instruction
}

// Instruction represents an opcode instruction. It contains the offset for the
// instruction, the opcode itself and the arguments for the opcode, if any.
type Instruction struct {
	Offset int
	OpCode OpCode
	Args   []byte
}

// NewDecompiler creates a new Decompiler instance with the provided context and bytecode.
// The bytecode is not automatically decompiled; Decompile() must be called before any
// information can be retrieved from the decompiler.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	if err := d.Decompile(); err != nil {
//	  log.Fatal(err)
//	}
func NewDecompiler(ctx context.Context, b []byte) *Decompiler {
	return &Decompiler{
		ctx:          ctx,
		bytecode:     b,
		bytecodeSize: uint64(len(b)),
		instructions: []Instruction{},
	}
}

// SetBytecode sets the bytecode that the decompiler should work on. It also
// updates the bytecode size.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.SetBytecode(anotherBytecode)
func (d *Decompiler) SetBytecode(b []byte) {
	d.bytecode = b
	d.bytecodeSize = uint64(len(b))
}

// GetBytecode returns the bytecode that the decompiler is working on.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	fmt.Println(d.GetBytecode())
func (d *Decompiler) GetBytecode() []byte {
	return d.bytecode
}

// GetBytecodeSize returns the size of the bytecode that the decompiler is working on.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	fmt.Println(d.GetBytecodeSize())
func (d *Decompiler) GetBytecodeSize() uint64 {
	return d.bytecodeSize
}

// Decompile decompiles the bytecode into opcode instructions. This must be
// called before any information can be retrieved from the decompiler. It
// returns an error if the bytecode is empty.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	if err := d.Decompile(); err != nil {
//	  log.Fatal(err)
//	}
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
// The function signature should be the Keccak-256 hash of the function signature.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	fmt.Println(d.MatchFunctionSignature("0xa9059cbb"))
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

// GetInstructionsByOpCode returns all instructions related to the specified opcode.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	instructions := d.GetInstructionsByOpCode(opcodes.PUSH1)
func (d *Decompiler) GetInstructionsByOpCode(op OpCode) []Instruction {
	var instructions []Instruction
	for _, instruction := range d.instructions {
		if instruction.OpCode == op {
			instructions = append(instructions, instruction)
		}
	}
	return instructions
}

// GetInstructions returns the decompiled opcode instructions.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	instructions := d.GetInstructions()
func (d *Decompiler) GetInstructions() []Instruction {
	return d.instructions
}

// String returns a string representation of the decompiled bytecode as a sequence of bytes.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	fmt.Println(d)
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
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	instructions := d.GetInstructions()
//	if len(instructions) > 0 {
//	  fmt.Println(d.IsOpCode(instructions[0], opcodes.PUSH1))
//	}
func (d *Decompiler) IsOpCode(instruction Instruction, op OpCode) bool {
	return instruction.OpCode == op
}

// OpCodeFound checks if the specified opcode is found in any of the decompiled instructions.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	fmt.Println(d.OpCodeFound(opcodes.PUSH1))
func (d *Decompiler) OpCodeFound(op OpCode) bool {
	for _, instruction := range d.instructions {
		if instruction.OpCode == op {
			return true
		}
	}
	return false
}

// GetOpCodesAsString returns the opcode part of the decompiled instructions as a slice of strings.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	fmt.Println(d.GetOpCodesAsString())
func (d *Decompiler) GetOpCodesAsString() []string {
	var opCodes []string
	for _, instruction := range d.instructions {
		opCodes = append(opCodes, instruction.OpCode.String())
	}
	return opCodes
}

// IsSelfDestruct checks if the decompiled instructions contain the SELFDESTRUCT opcode.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	fmt.Println(d.IsSelfDestruct())
func (d *Decompiler) IsSelfDestruct() bool {
	return d.OpCodeFound(SELFDESTRUCT)
}

// IsControlFlowInstruction checks if the given opcode represents a control flow instruction.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	fmt.Println(d.IsControlFlowInstruction(opcodes.JUMP))
func (d *Decompiler) IsControlFlowInstruction(op OpCode) bool {
	return op == JUMP || op == JUMPI || op == JUMPDEST || op == RETURN || op == REVERT || op == SELFDESTRUCT
}

// MatchInstruction checks if the given instruction exists in the decompiled instructions.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	instructions := d.GetInstructions()
//	if len(instructions) > 0 {
//	  fmt.Println(d.MatchInstruction(instructions[0]))
//	}
func (d *Decompiler) MatchInstruction(instruction Instruction) bool {
	for _, inst := range d.instructions {
		if inst.Offset == instruction.Offset && inst.OpCode == instruction.OpCode && bytes.Equal(inst.Args, instruction.Args) {
			return true
		}
	}
	return false
}

// CountOpCodeOccurrences counts the number of occurrences of a specific opcode in the decompiled instructions.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	count := d.CountOpCodeOccurrences(opcodes.PUSH1)
func (d *Decompiler) CountOpCodeOccurrences(op OpCode) int {
	count := 0
	for _, instruction := range d.instructions {
		if instruction.OpCode == op {
			count++
		}
	}
	return count
}

// GetInstructionByOffset retrieves the instruction at a specific offset in the decompiled instructions.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	instruction := d.GetInstructionByOffset(10)
func (d *Decompiler) GetInstructionByOffset(offset int) (Instruction, error) {
	for _, instruction := range d.instructions {
		if instruction.Offset == offset {
			return instruction, nil
		}
	}
	return Instruction{}, errors.New("instruction not found at offset")
}

// GetInstructionByIndex retrieves the instruction at a specific index in the decompiled instructions.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	instruction := d.GetInstructionByIndex(0)
func (d *Decompiler) GetInstructionByIndex(index int) (Instruction, error) {
	if index < 0 || index >= len(d.instructions) {
		return Instruction{}, errors.New("invalid instruction index")
	}
	return d.instructions[index], nil
}

// GetInstructionCount returns the total count of instructions in the decompiled instructions.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	count := d.GetInstructionCount()
func (d *Decompiler) GetInstructionCount() int {
	return len(d.instructions)
}

// ClearInstructions clears all the decompiled instructions.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	d.ClearInstructions()
func (d *Decompiler) ClearInstructions() {
	d.instructions = []Instruction{}
}

// GetInstructionByOpCodeAndArguments retrieves instructions that match a specific opcode and arguments.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	instructions := d.GetInstructionByOpCodeAndArguments(opcodes.CALL, []byte{0xab, 0xcd})
func (d *Decompiler) GetInstructionByOpCodeAndArguments(op OpCode, args []byte) []Instruction {
	var matchingInstructions []Instruction
	for _, instruction := range d.instructions {
		if instruction.OpCode == op && bytes.Equal(instruction.Args, args) {
			matchingInstructions = append(matchingInstructions, instruction)
		}
	}
	return matchingInstructions
}

// ReplaceInstruction replaces an instruction at a specific index or offset with a new instruction.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	newInstruction := opcodes.Instruction{
//		Offset: 10,
//		OpCode: opcodes.PUSH1,
//		Args:   []byte{0x01},
//	}
//	d.ReplaceInstruction(10, newInstruction)
func (d *Decompiler) ReplaceInstruction(indexOrOffset int, newInstruction Instruction) error {
	if indexOrOffset < 0 || indexOrOffset >= len(d.instructions) {
		return errors.New("invalid instruction index or offset")
	}
	d.instructions[indexOrOffset] = newInstruction
	return nil
}

// InsertInstruction inserts a new instruction at a specific index or offset in the decompiled instructions.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	newInstruction := opcodes.Instruction{
//		Offset: 10,
//		OpCode: opcodes.PUSH1,
//		Args:   []byte{0x01},
//	}
//	d.InsertInstruction(10, newInstruction)
func (d *Decompiler) InsertInstruction(indexOrOffset int, newInstruction Instruction) error {
	if indexOrOffset < 0 || indexOrOffset > len(d.instructions) {
		return errors.New("invalid instruction index or offset")
	}
	d.instructions = append(d.instructions[:indexOrOffset], append([]Instruction{newInstruction}, d.instructions[indexOrOffset:]...)...)
	return nil
}

// RemoveInstruction removes an instruction at a specific index or offset from the decompiled instructions.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	d.RemoveInstruction(10)
func (d *Decompiler) RemoveInstruction(indexOrOffset int) error {
	if indexOrOffset < 0 || indexOrOffset >= len(d.instructions) {
		return errors.New("invalid instruction index or offset")
	}
	d.instructions = append(d.instructions[:indexOrOffset], d.instructions[indexOrOffset+1:]...)
	return nil
}

// GetJumpDestinations returns a list of jump destinations (JUMPDEST opcodes) in the decompiled instructions.
// Example:
//
//	d := opcodes.NewDecompiler(context.Background(), bytecode)
//	d.Decompile()
//	jumpDestinations := d.GetJumpDestinations()
func (d *Decompiler) GetJumpDestinations() []Instruction {
	var jumpDestinations []Instruction
	for _, instruction := range d.instructions {
		if instruction.OpCode == JUMPDEST {
			jumpDestinations = append(jumpDestinations, instruction)
		}
	}
	return jumpDestinations
}
