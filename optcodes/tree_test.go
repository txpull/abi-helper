package optcodes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTree_GetTree(t *testing.T) {
	// Create a new Decompiler instance and set the instructions.
	decompiler := &Decompiler{
		instructions: []Instruction{
			{
				Offset: 0,
				OpCode: PUSH2,
				Args:   nil,
			},
			{
				Offset: 1,
				OpCode: ADD,
				Args:   nil,
			},
			{
				Offset: 2,
				OpCode: JUMP,
				Args:   nil,
			},
		},
	}

	tree := decompiler.GetTree()

	// Define the expected tree structure
	expectedTree := &TreeNode{
		Instruction: Instruction{
			Offset: 0,
			OpCode: PUSH2,
			Args:   nil,
		},
		Children: []*TreeNode{
			{
				Instruction: Instruction{
					Offset: 1,
					OpCode: ADD,
					Args:   nil,
				},
				Children: []*TreeNode{
					{
						Instruction: Instruction{
							Offset: 2,
							OpCode: JUMP,
							Args:   nil,
						},
						Children: nil,
					},
				},
			},
		},
	}

	// Compare the expected tree with the actual tree
	assert.Equal(t, expectedTree, tree)
}
