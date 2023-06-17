package controlflow

func (cfg *CfgDecoder) FindAllPaths() [][]*Node {
	paths := make([][]*Node, 0)
	visited := make(map[*Node]bool)

	cfg.dfsWithLoopDetection(cfg.entry, nil, visited, &paths)

	return paths
}

func (cfg *CfgDecoder) dfsWithLoopDetection(node *Node, loopPath []*Node, visited map[*Node]bool, paths *[][]*Node) {
	if visited[node] {
		return
	}

	visited[node] = true
	loopPath = append(loopPath, node)

	if node == cfg.exit {
		exitPath := make([]*Node, len(loopPath))
		copy(exitPath, loopPath)
		*paths = append(*paths, exitPath)
	} else {
		if node.Next != nil {
			cfg.dfsWithLoopDetection(node.Next, loopPath, visited, paths)
		}
		if node.Branch != nil {
			cfg.dfsWithLoopDetection(node.Branch, loopPath, visited, paths)
		}
		if node.Function != nil {
			cfg.dfsWithLoopDetection(node.Function, loopPath, visited, paths)
		}
	}

	if cfg.isLoopHeader(node) {
		loop := cfg.findLoopByHeader(node)
		if loop != nil {
			loopExitPath := make([]*Node, len(loopPath))
			copy(loopExitPath, loopPath)
			*paths = append(*paths, loopExitPath)

			cfg.dfsWithLoopDetection(loop.Exit, append([]*Node{}, loopPath...), visited, paths)
		}
	}

	delete(visited, node)
}

func (cfg *CfgDecoder) isLoopHeader(node *Node) bool {
	for _, loop := range cfg.loops {
		if loop.Header == node {
			return true
		}
	}
	return false
}

func (cfg *CfgDecoder) findLoopByHeader(header *Node) *Loop {
	for _, loop := range cfg.loops {
		if loop.Header == header {
			return loop
		}
	}
	return nil
}
