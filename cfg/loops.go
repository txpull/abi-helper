package cfg

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/txpull/bytecode/opcodes"
)

type Loop struct {
	Header         *Node
	Exit           *Node
	LoopNodes      []*Node
	LoopBounds     []int
	LoopInvariants []string
}

func (cfg *CFG) detectLoops() {
	visited := make(map[*Node]bool)
	for _, node := range cfg.nodes {
		if visited[node] {
			continue
		}
		cfg.detectLoop(node, visited)
	}
}

func (cfg *CFG) detectLoop(node *Node, visited map[*Node]bool) {
	loopBounds := make([]int, 0)
	loopInvariants := make([]string, 0)
	exitNode := node

	// Iterate through the graph using an iterative deepening depth-first search (IDDFS)
	for depth := 1; depth <= len(cfg.nodes); depth++ {
		if cfg.dfs(node, visited, loopBounds, loopInvariants, depth, &exitNode) {
			break
		}
	}

	if len(loopBounds) > 0 {
		loop := &Loop{
			Header:         node,
			Exit:           exitNode,
			LoopNodes:      node.getLoopNodes(exitNode),
			LoopBounds:     loopBounds,
			LoopInvariants: loopInvariants,
		}
		cfg.loops = append(cfg.loops, loop)
	}
}

func (cfg *CFG) dfs(node *Node, visited map[*Node]bool, loopBounds []int, loopInvariants []string, depth int, exitNode **Node) bool {
	visited[node] = true

	// Determine loop characteristics (e.g., loop bounds, loop invariants)
	instructions := cfg.instructions[node.Offset : node.Offset+1]
	for _, instruction := range instructions {
		// Extract loop bounds and invariants based on specific conditions or patterns in the bytecode
		if instruction.OpCode == opcodes.JUMP {
			// Example: Extract loop bounds by analyzing the jump destination
			jumpOffset := int(common.BytesToHash(instruction.Args).Big().Int64())
			loopBounds = append(loopBounds, jumpOffset)
		} else if instruction.OpCode == opcodes.ADD {
			// Example: Extract loop invariants based on specific instructions
			// Assuming the loop invariant is the value being added
			// Adjust the condition and extraction logic based on the desired loop invariants
			invariantValue := int(common.BytesToHash(instruction.Args).Big().Int64())
			loopInvariants = append(loopInvariants, fmt.Sprintf("%d", invariantValue))
		}
	}

	// Base case: Reached the desired depth or already visited this node at a shallower depth
	if depth <= 0 || visited[node] {
		return visited[node]
	}

	if node.Next != nil && !visited[node.Next] {
		if *exitNode == nil || node.Next.Offset > (*exitNode).Offset {
			*exitNode = node.Next
		}
		if cfg.dfs(node.Next, visited, loopBounds, loopInvariants, depth-1, exitNode) {
			return true
		}
	}

	if node.Branch != nil && !visited[node.Branch] {
		if *exitNode == nil || node.Branch.Offset > (*exitNode).Offset {
			*exitNode = node.Branch
		}
		if cfg.dfs(node.Branch, visited, loopBounds, loopInvariants, depth-1, exitNode) {
			return true
		}
	}

	if node.Function != nil && !visited[node.Function] {
		if *exitNode == nil || node.Function.Offset > (*exitNode).Offset {
			*exitNode = node.Function
		}
		if cfg.dfs(node.Function, visited, loopBounds, loopInvariants, depth-1, exitNode) {
			return true
		}
	}

	return false
}

func (node *Node) getLoopNodes(exitNode *Node) []*Node {
	loopNodes := make([]*Node, 0)
	loopNodes = append(loopNodes, node)

	// Traverse the graph from the header node until the exit node is reached
	currNode := node
	for currNode != exitNode {
		if currNode.Next != nil {
			loopNodes = append(loopNodes, currNode.Next)
			currNode = currNode.Next
		} else if currNode.Branch != nil {
			loopNodes = append(loopNodes, currNode.Branch)
			currNode = currNode.Branch
		} else if currNode.Function != nil {
			loopNodes = append(loopNodes, currNode.Function)
			currNode = currNode.Function
		}
	}

	return loopNodes
}
