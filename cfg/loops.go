package cfg

type Loop struct {
	Header *Node
	Nodes  []*Node
}

func (cfg *CFG) detectLoops() {
	cfg.clearVisited()

	for _, node := range cfg.nodes {
		cfg.findBackEdges(node, nil)
	}
}

func (cfg *CFG) clearVisited() {
	for _, node := range cfg.nodes {
		node.visited = false
	}
}

func (cfg *CFG) findBackEdges(node, parent *Node) {
	if node == nil || node.visited {
		return
	}

	node.visited = true

	if parent != nil && cfg.isAncestor(node, parent) {
		cfg.loops = append(cfg.loops, &Loop{Header: node, Nodes: cfg.findLoopNodes(node, parent)})
	}

	cfg.findBackEdges(node.Next, node)
	cfg.findBackEdges(node.Branch, node)
	cfg.findBackEdges(node.Function, node)
}

func (cfg *CFG) isAncestor(node, descendant *Node) bool {
	for n := descendant; n != nil; n = n.Next {
		if n == node {
			return true
		}
	}

	return false
}

func (cfg *CFG) findLoopNodes(header, tail *Node) []*Node {
	nodes := []*Node{tail}

	for n := tail; n != header; n = n.Next {
		nodes = append(nodes, n)
	}

	return nodes
}
