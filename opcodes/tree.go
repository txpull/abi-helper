package opcodes

import (
	"fmt"
	"strings"
)

// TreeNode represents a node in the execution flow tree.
// Each node represents an instruction, and has zero or more children.
// These children represent subsequent instructions in the execution flow.
type TreeNode struct {
	Instruction Instruction
	Children    []*TreeNode
}

// GetTree returns the root of the tree representation of the execution flow.
// The root of the tree is not a real instruction; its children are the real
// start of the execution flow.
// If there are no instructions, it returns nil.
// Example:
//
//	tree := d.GetTree()
func (d *Decompiler) GetTree() *TreeNode {
	if len(d.instructions) == 0 {
		return nil
	}

	root := &TreeNode{}
	stack := []*TreeNode{root}

	d.buildExecutionTree(stack)

	return root.Children[0]
}

// buildExecutionTree builds the execution flow tree from the given stack of nodes.
// This is a private method used by GetTree().
func (d *Decompiler) buildExecutionTree(stack []*TreeNode) {
	for _, instruction := range d.instructions {
		node := &TreeNode{Instruction: instruction}

		if len(stack) > 0 {
			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, node)
		}

		stack = append(stack, node)

		if instruction.OpCode.IsJump() {
			stack = stack[:len(stack)-1]
		}
	}
}

// PrintTree prints the tree representation of the execution flow.
// Each node is indented according to its depth in the tree.
// Example:
//
//	d.PrintTree()
func (d *Decompiler) PrintTree() {
	tree := d.GetTree()
	d.printExecutionTree(tree, 0)
}

// PrintInstructionTree prints the execution flow tree for a specific instruction.
// The root of the tree is the given instruction, and the children are the subsequent instructions.
// Each node is indented according to its depth in the tree.
// Example:
//
//	d.PrintInstructionTree(instruction)
func (d *Decompiler) PrintInstructionTree(instruction Instruction) {
	tree := d.getInstructionTree(instruction)
	d.printExecutionTree(tree, 0)
}

// GetInstructionTreeFormatted returns a string representation of the execution flow tree for a specific instruction.
// The root of the tree is the given instruction, and the children are the subsequent instructions.
// Each node is indented according to its depth in the tree.
// The indent string is prepended to each line.
// Example:
//
//	s := d.GetInstructionTreeFormatted(instruction, "  ")
//	fmt.Println(s)
func (d *Decompiler) GetInstructionTreeFormatted(instruction Instruction, indent string) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s0x%04x %s\n", indent, instruction.Offset, instruction.OpCode.String()))

	childIndent := indent + "   "
	for _, child := range d.GetChildrenByOffset(instruction.Offset) {
		builder.WriteString(d.GetInstructionTreeFormatted(child, childIndent))
	}

	return builder.String()
}

// GetChildrenByOffset returns the child instructions of a given offset.
// These are the instructions that directly follow the instruction at the given offset.
// If no such instructions exist, it returns an empty slice.
// Example:
//
//	children := d.GetChildrenByOffset(offset)
func (d *Decompiler) GetChildrenByOffset(offset int) []Instruction {
	var children []Instruction
	for _, instr := range d.instructions {
		if instr.Offset > offset {
			break
		}
		if instr.Offset == offset+1 {
			children = append(children, instr)
		}
	}
	return children
}

// printExecutionTree prints a given node and its children in the execution flow tree.
// This is a private method used by PrintTree() and PrintInstructionTree().
func (d *Decompiler) printExecutionTree(node *TreeNode, indent int) {
	fmt.Printf("%sOffset: 0x%04x, OpCode: %s, Args: %x\n", strings.Repeat(" ", indent*2), node.Instruction.Offset, node.Instruction.OpCode.String(), node.Instruction.Args)
	for _, child := range node.Children {
		d.printExecutionTree(child, indent+1)
	}
}

// getInstructionTree returns the root of the tree representation of the execution flow for a specific instruction.
// The root of the tree is the given instruction, and the children are the subsequent instructions.
// This is a private method used by PrintInstructionTree() and GetInstructionTreeFormatted().
func (d *Decompiler) getInstructionTree(instruction Instruction) *TreeNode {
	root := &TreeNode{Instruction: instruction}
	offset := instruction.Offset
	visited := make(map[int]bool)

	d.buildInstructionTree(root, offset, visited)

	return root
}

// buildInstructionTree builds the execution flow tree for a specific instruction from the given node.
// This is a private method used by getInstructionTree().
func (d *Decompiler) buildInstructionTree(node *TreeNode, offset int, visited map[int]bool) {
	for _, instr := range d.instructions {
		if instr.Offset == offset && !visited[offset] {
			child := &TreeNode{Instruction: instr}
			node.Children = append(node.Children, child)
			visited[offset] = true
			d.buildInstructionTree(child, offset+1, visited)
		}
	}
}
