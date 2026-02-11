package graph

// TopologicalSort returns requirements in topological order (Kahn's algorithm).
// Returns nil if the graph contains cycles.
func (g *Graph) TopologicalSort() []string {
	// Count incoming edges for each node
	inDegree := make(map[string]int)
	for _, req := range g.db.All() {
		inDegree[req.ReqID] = 0
	}
	for _, deps := range g.dependencies {
		for _, dep := range deps {
			inDegree[dep]++ // dep has an incoming edge from the node that depends on it
		}
	}

	// Wait, this is backwards. Let me reconsider.
	// In a dependency graph:
	// - A depends on B means there's an edge A -> B
	// - We want B to come before A in the order
	// - So we need to count how many things depend on each node

	// Reset and recalculate properly
	inDegree = make(map[string]int)
	for _, req := range g.db.All() {
		inDegree[req.ReqID] = len(g.dependencies[req.ReqID])
	}

	// Start with nodes that have no dependencies
	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}

	var result []string
	for len(queue) > 0 {
		// Dequeue
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		// Reduce in-degree for dependents
		for _, dependent := range g.dependents[node] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// If we couldn't visit all nodes, there's a cycle
	if len(result) != g.db.Len() {
		return nil
	}

	return result
}

// ExecutionOrder returns requirements in the order they should be completed.
// Dependencies come before dependents.
func (g *Graph) ExecutionOrder() []string {
	return g.TopologicalSort()
}

// Layers returns requirements grouped by their depth in the dependency graph.
// Layer 0 contains roots (no dependencies), layer 1 contains their dependents, etc.
func (g *Graph) Layers() [][]string {
	// Calculate depth for each node
	depth := make(map[string]int)
	maxDepth := 0

	// Start with roots at depth 0
	for _, root := range g.Roots() {
		depth[root] = 0
	}

	// BFS to calculate depths
	queue := g.Roots()
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		for _, dependent := range g.dependents[node] {
			newDepth := depth[node] + 1
			if d, exists := depth[dependent]; !exists || newDepth > d {
				depth[dependent] = newDepth
				if newDepth > maxDepth {
					maxDepth = newDepth
				}
				queue = append(queue, dependent)
			}
		}
	}

	// Group by depth
	layers := make([][]string, maxDepth+1)
	for i := range layers {
		layers[i] = make([]string, 0)
	}

	for node, d := range depth {
		layers[d] = append(layers[d], node)
	}

	return layers
}
