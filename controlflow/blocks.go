package controlflow

type Block struct {
	Nodes []*Node
}

func (cfg *CfgDecoder) createBlocks() {
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
