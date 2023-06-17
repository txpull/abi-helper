/*
Package cfg provides functionality for creating and analyzing Control Flow Graphs (CFGs) for Ethereum smart contracts.

Control Flow Graphs represent the flow of execution within a program by modeling the relationships between different instructions or blocks of code.

The CFG struct represents a Control Flow Graph and contains information about its nodes, blocks, entry point, exit point, instructions, and loops.

The Node struct represents a node in the Control Flow Graph and contains information about its offset, next node, branch node, function node, dominators, and post-dominators.

Example usage:

	cfg, err := NewCFG(bytecode)
	if err != nil {
		// Handle error
	}

	cfg.Detect()

	cfg.Print() // Print the Control Flow Graph

	cfg.PerformDominatorAnalysis() // Perform dominator analysis on the CFG

	cfg.RemoveDeadCode() // Remove dead code from the CFG

	cfg.Print() // Print the updated Control Flow Graph

	cfg.Reverse() // Reverse the Control Flow Graph

	cfg.Print() // Print the reversed Control Flow Graph
*/
package cfg

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/txpull/bytecode/opcodes"
)

// CFG represents a Control Flow Graph.
type CFG struct {
	nodes        map[int]*Node         // Nodes in the CFG
	blocks       []*Block              // Blocks in the CFG
	entry        *Node                 // Entry node of the CFG
	exit         *Node                 // Exit node of the CFG
	instructions []opcodes.Instruction // Instructions of the CFG
	loops        []*Loop               // Loops in the CFG
}

// Node represents a node in the Control Flow Graph.
type Node struct {
	Offset         int            // Offset of the node
	Next           *Node          // Next node in the execution flow
	Branch         *Node          // Branch node in the execution flow
	Function       *Node          // Function node in the execution flow
	Dominators     map[*Node]bool // Dominators of the node
	PostDominators map[*Node]bool // Post-dominators of the node
}

// NewCFG creates a new Control Flow Graph from the given bytecode.
func NewCFG(bytecode []byte) (*CFG, error) {
	decompiler := opcodes.NewDecompiler(context.Background(), bytecode)
	if err := decompiler.Decompile(); err != nil {
		return nil, fmt.Errorf("error decompiling bytecode: %w", err)
	}

	instructions := decompiler.GetInstructions()

	cfg := &CFG{
		nodes:        make(map[int]*Node),
		entry:        nil,
		exit:         nil,
		instructions: instructions,
	}

	cfg.createNodes()
	cfg.connectNodes()

	cfg.entry = cfg.nodes[instructions[0].Offset]

	cfg.createBlocks()

	return cfg, nil
}

// Detect initiates the loop detection process on the Control Flow Graph.
func (cfg *CFG) Detect() {
	cfg.DetectLoops()
}

// createNodes creates the nodes of the Control Flow Graph.
func (cfg *CFG) createNodes() {
	for _, instruction := range cfg.instructions {
		cfg.addNode(instruction.Offset)
	}
}

// addNode adds a node to the Control Flow Graph.
func (cfg *CFG) addNode(offset int) {
	if _, ok := cfg.nodes[offset]; !ok {
		node := &Node{
			Offset:         offset,
			Dominators:     make(map[*Node]bool),
			PostDominators: make(map[*Node]bool),
		}
		cfg.nodes[offset] = node
	}
}

// connectNodes connects the nodes of the Control Flow Graph.
func (cfg *CFG) connectNodes() {
	for i, instruction := range cfg.instructions {
		node := cfg.nodes[instruction.Offset]

		if i < len(cfg.instructions)-1 {
			node.Next = cfg.nodes[cfg.instructions[i+1].Offset]
		}

		switch instruction.OpCode {
		case opcodes.JUMP:
			jumpOffset := int(common.BytesToHash(instruction.Args).Big().Int64())
			if jumpOffset >= 0 && jumpOffset < len(cfg.nodes) {
				node.Branch = cfg.nodes[jumpOffset]
			}
		case opcodes.JUMPI:
			jumpOffset := int(common.BytesToHash(instruction.Args).Big().Int64())
			if jumpOffset >= 0 && jumpOffset < len(cfg.nodes) {
				node.Branch = cfg.nodes[jumpOffset]
			}
			if i < len(cfg.instructions)-1 {
				node.Next = cfg.nodes[cfg.instructions[i+1].Offset]
			}
		case opcodes.CALL, opcodes.CALLCODE, opcodes.DELEGATECALL, opcodes.STATICCALL:
			if i < len(cfg.instructions)-1 {
				node.Function = cfg.nodes[cfg.instructions[i+1].Offset]
			}
		case opcodes.RETURN, opcodes.REVERT, opcodes.SELFDESTRUCT:
			cfg.exit = node
		}
	}
}

// Reverse returns a new Control Flow Graph with reversed edges.
func (cfg *CFG) Reverse() *CFG {
	reversedCFG := &CFG{
		nodes:        make(map[int]*Node),
		entry:        cfg.exit,
		exit:         cfg.entry,
		instructions: cfg.instructions,
	}

	for _, node := range cfg.nodes {
		reversedCFG.addNode(node.Offset)
	}

	for _, node := range cfg.nodes {
		if node.Next != nil {
			reversedCFG.connectReversedNodes(node.Next.Offset, node.Offset)
		}
		if node.Branch != nil {
			reversedCFG.connectReversedNodes(node.Branch.Offset, node.Offset)
		}
		if node.Function != nil {
			reversedCFG.connectReversedNodes(node.Function.Offset, node.Offset)
		}
	}

	return reversedCFG
}

// connectReversedNodes connects reversed nodes in the Control Flow Graph.
func (cfg *CFG) connectReversedNodes(from, to int) {
	fromNode := cfg.nodes[from]
	toNode := cfg.nodes[to]

	if fromNode.Next == nil {
		fromNode.Next = toNode
	} else if fromNode.Branch == nil {
		fromNode.Branch = toNode
	} else if fromNode.Function == nil {
		fromNode.Function = toNode
	}
}

// Print prints the Control Flow Graph.
func (cfg *CFG) Print() {
	fmt.Println("Control Flow Graph:")
	fmt.Println("Entry:", cfg.entry.Offset)

	for _, node := range cfg.nodes {
		fmt.Printf("Node: %d\n", node.Offset)
		if node.Next != nil {
			fmt.Printf("  Next: %d\n", node.Next.Offset)
		}
		if node.Branch != nil {
			fmt.Printf("  Branch: %d\n", node.Branch.Offset)
		}
		if node.Function != nil {
			fmt.Printf("  Function: %d\n", node.Function.Offset)
		}
	}

	fmt.Println("Exit:", cfg.exit.Offset)
}

// PerformDominatorAnalysis performs dominator analysis on the Control Flow Graph.
func (cfg *CFG) PerformDominatorAnalysis() {
	// Initialize dominators of each node to include itself
	for _, node := range cfg.nodes {
		node.Dominators[node] = true
	}

	// Perform dominator analysis on the original CFG
	cfg.computeDominators(cfg.entry)

	// Perform dominator analysis on the reversed CFG
	reversedCFG := cfg.Reverse()
	reversedCFG.computeDominators(reversedCFG.entry)

	// Copy the dominators and post-dominators from the reversed CFG
	for _, node := range cfg.nodes {
		reversedNode := reversedCFG.nodes[node.Offset]
		node.Dominators = reversedNode.Dominators
		node.PostDominators = reversedNode.PostDominators
	}
}

// computeDominators computes the dominators of a node in the Control Flow Graph.
func (cfg *CFG) computeDominators(node *Node) {
	changed := true
	for changed {
		changed = false

		for _, n := range cfg.nodes {
			if n != node && len(n.Dominators) > 0 {
				newDominators := make(map[*Node]bool)
				for dom := range n.Dominators {
					if dom.Dominators[node] {
						newDominators[dom] = true
					}
				}

				if len(newDominators) != len(n.Dominators)+1 {
					newDominators[node] = true
					n.Dominators = newDominators
					changed = true
				}
			}
		}
	}

	for _, n := range cfg.nodes {
		if len(n.Dominators) == 0 {
			n.Dominators[node] = true
		}
	}
}

// removeDeadCode removes dead code from the Control Flow Graph.
func (cfg *CFG) RemoveDeadCode() {
	reachableNodes := cfg.findReachableNodes()
	for offset := range cfg.nodes {
		if _, ok := reachableNodes[offset]; !ok {
			delete(cfg.nodes, offset)
		}
	}
}

// findReachableNodes finds the reachable nodes in the Control Flow Graph.
func (cfg *CFG) findReachableNodes() map[int]*Node {
	reachableNodes := make(map[int]*Node)
	cfg.markReachable(cfg.entry, reachableNodes)
	return reachableNodes
}

// markReachable marks the reachable nodes in the Control Flow Graph.
func (cfg *CFG) markReachable(node *Node, reachableNodes map[int]*Node) {
	if _, ok := reachableNodes[node.Offset]; ok {
		// Node has already been visited
		return
	}

	reachableNodes[node.Offset] = node

	if node.Next != nil {
		cfg.markReachable(node.Next, reachableNodes)
	}
	if node.Branch != nil {
		cfg.markReachable(node.Branch, reachableNodes)
	}
	if node.Function != nil {
		cfg.markReachable(node.Function, reachableNodes)
	}
}
