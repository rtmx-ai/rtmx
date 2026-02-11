package graph

// Tarjan's Strongly Connected Components algorithm
// Used for cycle detection in dependency graphs

// tarjanState holds the state for Tarjan's algorithm
type tarjanState struct {
	graph   *Graph
	index   int
	stack   []string
	onStack map[string]bool
	indices map[string]int
	lowlink map[string]int
	sccs    [][]string
}

// FindCycles finds all cycles in the graph using Tarjan's SCC algorithm.
// Returns a list of strongly connected components with more than one node.
func (g *Graph) FindCycles() [][]string {
	state := &tarjanState{
		graph:   g,
		index:   0,
		stack:   make([]string, 0),
		onStack: make(map[string]bool),
		indices: make(map[string]int),
		lowlink: make(map[string]int),
		sccs:    make([][]string, 0),
	}

	// Visit all nodes
	for _, req := range g.db.All() {
		if _, visited := state.indices[req.ReqID]; !visited {
			state.strongConnect(req.ReqID)
		}
	}

	// Filter to only cycles (SCCs with more than one node)
	var cycles [][]string
	for _, scc := range state.sccs {
		if len(scc) > 1 {
			cycles = append(cycles, scc)
		}
	}

	return cycles
}

// strongConnect is the recursive part of Tarjan's algorithm
func (s *tarjanState) strongConnect(v string) {
	// Set the depth index for v
	s.indices[v] = s.index
	s.lowlink[v] = s.index
	s.index++
	s.stack = append(s.stack, v)
	s.onStack[v] = true

	// Consider successors of v (dependencies)
	for _, w := range s.graph.dependencies[v] {
		if _, visited := s.indices[w]; !visited {
			// Successor w has not yet been visited; recurse on it
			s.strongConnect(w)
			s.lowlink[v] = min(s.lowlink[v], s.lowlink[w])
		} else if s.onStack[w] {
			// Successor w is in stack and hence in the current SCC
			s.lowlink[v] = min(s.lowlink[v], s.indices[w])
		}
	}

	// If v is a root node, pop the stack and generate an SCC
	if s.lowlink[v] == s.indices[v] {
		var scc []string
		for {
			w := s.stack[len(s.stack)-1]
			s.stack = s.stack[:len(s.stack)-1]
			s.onStack[w] = false
			scc = append(scc, w)
			if w == v {
				break
			}
		}
		s.sccs = append(s.sccs, scc)
	}
}

// HasCycles returns true if the graph contains any cycles.
func (g *Graph) HasCycles() bool {
	return len(g.FindCycles()) > 0
}

// FindCyclePath returns a path through a cycle starting and ending at the same node.
// Given a list of cycle members, finds the actual path.
func (g *Graph) FindCyclePath(cycleMembers []string) []string {
	if len(cycleMembers) == 0 {
		return nil
	}

	// Create a set for quick lookup
	memberSet := make(map[string]bool)
	for _, m := range cycleMembers {
		memberSet[m] = true
	}

	// Start from the first member and find a path back to it
	start := cycleMembers[0]
	visited := make(map[string]bool)
	var path []string

	var dfs func(current string) bool
	dfs = func(current string) bool {
		path = append(path, current)
		visited[current] = true

		for _, dep := range g.dependencies[current] {
			if !memberSet[dep] {
				continue // Skip non-cycle members
			}
			if dep == start && len(path) > 1 {
				// Found cycle back to start
				path = append(path, start)
				return true
			}
			if !visited[dep] {
				if dfs(dep) {
					return true
				}
			}
		}

		path = path[:len(path)-1]
		return false
	}

	if dfs(start) {
		return path
	}
	return cycleMembers // Fallback to just the members
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
