package cfg

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/txpull/bytecode/opcodes"
)

type CFG struct {
	nodes        map[int]*Node
	blocks       []*Block
	entry        *Node
	exit         *Node
	instructions []opcodes.Instruction
	loops        []*Loop
	paths        [][]*Node
}

type Node struct {
	Offset         int
	Next           *Node
	Branch         *Node
	Function       *Node
	Dominators     map[*Node]bool
	PostDominators map[*Node]bool
	visited        bool // for DFS
}

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
	cfg.detectLoops()

	return cfg, nil
}

func (cfg *CFG) createNodes() {
	for _, instruction := range cfg.instructions {
		cfg.addNode(instruction.Offset)
	}
}

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

func (cfg *CFG) connectNodes() {
	for i, instruction := range cfg.instructions {
		node := cfg.nodes[instruction.Offset]

		if i < len(cfg.instructions)-1 {
			node.Next = cfg.nodes[cfg.instructions[i+1].Offset]
		}

		switch instruction.OpCode {
		case opcodes.JUMP, opcodes.JUMPI:
			jumpOffset := int(common.BytesToHash(instruction.Args).Big().Int64())
			node.Branch = cfg.nodes[jumpOffset]
		case opcodes.CALL, opcodes.CALLCODE, opcodes.DELEGATECALL, opcodes.STATICCALL:
			node.Function = cfg.nodes[cfg.instructions[i+1].Offset]
		case opcodes.RETURN, opcodes.REVERT, opcodes.SELFDESTRUCT:
			cfg.exit = node
		}
	}
}

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

func (cfg *CFG) removeDeadCode() {
	reachableNodes := cfg.findReachableNodes()
	for offset, _ := range cfg.nodes {
		if _, ok := reachableNodes[offset]; !ok {
			delete(cfg.nodes, offset)
		}
	}
}

func (cfg *CFG) findReachableNodes() map[int]*Node {
	reachableNodes := make(map[int]*Node)
	cfg.markReachable(cfg.entry, reachableNodes)
	return reachableNodes
}

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