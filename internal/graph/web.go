package graph

import (
	"sort"
)

// Web represents a connected component of incomplete requirements in the
// dependency graph. Requirements within a web are connected by dependency
// edges (in either direction) and can be worked on together. Requirements
// in different webs are fully independent and can be parallelized.
type Web struct {
	// IDs is the sorted list of requirement IDs in this web.
	IDs []string

	// Unblocked contains IDs that have no incomplete dependencies
	// and can be started immediately.
	Unblocked []string

	// Blocked contains IDs that have at least one incomplete dependency.
	Blocked []string

	// TotalEffort is the sum of effort_weeks across all requirements.
	TotalEffort float64
}

// DetectWebs computes independent work webs from the dependency graph.
// Each web is a connected component in the undirected dependency graph,
// restricted to incomplete requirements. Complete requirements are excluded
// because they no longer need work.
//
// The algorithm:
//  1. Build an undirected adjacency list of incomplete requirements.
//  2. BFS/DFS from each unvisited node to find connected components.
//  3. For each component, classify members as blocked or unblocked.
//  4. Sort webs by total effort descending (largest first).
func (g *Graph) DetectWebs() []Web {
	// Collect incomplete requirement IDs
	incomplete := make(map[string]bool)
	for _, req := range g.db.All() {
		if req.IsIncomplete() {
			incomplete[req.ReqID] = true
		}
	}

	if len(incomplete) == 0 {
		return nil
	}

	// Build undirected adjacency among incomplete nodes only
	adj := make(map[string][]string)
	for id := range incomplete {
		adj[id] = nil // ensure every node has an entry
	}
	for id := range incomplete {
		for _, dep := range g.dependencies[id] {
			if incomplete[dep] {
				adj[id] = append(adj[id], dep)
				adj[dep] = append(adj[dep], id)
			}
		}
	}

	// BFS to find connected components
	visited := make(map[string]bool)
	var webs []Web

	for id := range incomplete {
		if visited[id] {
			continue
		}

		// BFS from this node
		var component []string
		queue := []string{id}
		visited[id] = true

		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]
			component = append(component, cur)

			for _, neighbor := range adj[cur] {
				if !visited[neighbor] {
					visited[neighbor] = true
					queue = append(queue, neighbor)
				}
			}
		}

		sort.Strings(component)

		// Classify blocked vs unblocked and sum effort
		web := Web{IDs: component}
		for _, rid := range component {
			req := g.db.Get(rid)
			if req != nil {
				web.TotalEffort += req.EffortWeeks
			}
			if g.IsBlocked(rid) {
				web.Blocked = append(web.Blocked, rid)
			} else {
				web.Unblocked = append(web.Unblocked, rid)
			}
		}

		webs = append(webs, web)
	}

	// Sort webs: largest effort first
	sort.Slice(webs, func(i, j int) bool {
		return webs[i].TotalEffort > webs[j].TotalEffort
	})

	return webs
}
