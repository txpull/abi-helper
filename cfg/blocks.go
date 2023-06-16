package cfg

type Block struct {
	Nodes []*Node
}

func (cfg *CFG) createBlocks() {
	block := &Block{}

	for _, node := range cfg.nodes {
		block.Nodes = append(block.Nodes, node)

		if node.Branch != nil || node.Function != nil || node == cfg.exit {
			cfg.blocks = append(cfg.blocks, block)
			block = &Block{}
		}
	}

	// Catch the case where the last instruction is not an exit point
	if len(block.Nodes) > 0 {
		cfg.blocks = append(cfg.blocks, block)
	}
}

func (cfg *CFG) FindAllPaths() {
	cfg.paths = cfg.FindPaths(cfg.entry, cfg.exit)
}

func (cfg *CFG) FindPaths(start, end *Node) [][]*Node {
	if start == end {
		return [][]*Node{{end}}
	}

	paths := [][]*Node{}

	for _, node := range cfg.Successors(start) {
		for _, path := range cfg.FindPaths(node, end) {
			paths = append(paths, append([]*Node{start}, path...))
		}
	}

	return paths
}

func (cfg *CFG) Successors(node *Node) []*Node {
	successors := []*Node{}

	if node.Next != nil {
		successors = append(successors, node.Next)
	}

	if node.Branch != nil {
		successors = append(successors, node.Branch)
	}

	if node.Function != nil {
		successors = append(successors, node.Function)
	}

	return successors
}
