package optcodes

import (
	"fmt"
	"strings"
)

// TreeNode represents a node in the execution flow tree.
type TreeNode struct {
	Instruction Instruction
	Children    []*TreeNode
}

// GetTree returns the tree representation of the execution flow.
func (d *Decompiler) GetTree() *TreeNode {
	root := &TreeNode{}
	d.buildExecutionTree(root, 0)
	return root
}

// buildExecutionTree recursively builds the execution flow tree starting from the given node.
func (d *Decompiler) buildExecutionTree(node *TreeNode, offset int) {
	for i := offset; i < len(d.instructions); i++ {
		instruction := d.instructions[i]
		childNode := &TreeNode{Instruction: instruction}
		node.Children = append(node.Children, childNode)
		if d.IsControlFlowInstruction(instruction.OpCode) {
			d.buildExecutionTree(childNode, instruction.Offset+1)
		}
	}
}

// PrintTree prints the tree representation of the execution flow.
func (d *Decompiler) PrintTree() {
	tree := d.GetTree()
	d.printExecutionTree(tree, 0)
}

// printExecutionTree prints the execution flow tree.
func (d *Decompiler) printExecutionTree(node *TreeNode, indent int) {
	fmt.Printf("%sOffset: 0x%04x, OpCode: %s, Args: %x\n", strings.Repeat(" ", indent*2), node.Instruction.Offset, node.Instruction.OpCode.String(), node.Instruction.Args)
	for _, child := range node.Children {
		d.printExecutionTree(child, indent+1)
	}
}

// PrintInstructionTree prints the execution flow tree for a specific instruction.
func (d *Decompiler) PrintInstructionTree(instruction Instruction) {
	tree := d.getInstructionTree(instruction)
	d.printExecutionTree(tree, 0)
}

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

// getInstructionTree returns the execution flow tree for a specific instruction.
func (d *Decompiler) getInstructionTree(instruction Instruction) *TreeNode {
	root := &TreeNode{Instruction: instruction}
	offset := instruction.Offset
	visited := make(map[int]bool)

	d.buildInstructionTree(root, offset, visited)

	return root
}

// buildInstructionTree recursively builds the execution flow tree for a specific instruction.
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
