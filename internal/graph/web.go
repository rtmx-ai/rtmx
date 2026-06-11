package graph

import (
	"sort"
	"strings"
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

// WebDep represents a directed dependency between two webs.
type WebDep struct {
	From int // index of upstream web
	To   int // index of downstream web
}

// WebDependencies computes cross-web dependencies by analyzing inter-requirement
// dependencies that span web boundaries, including transitive dependencies
// through complete intermediary requirements. Returns a list of directed edges
// between web indices where From must complete before To. REQ-ORCH-010.
func (g *Graph) WebDependencies(webs []Web) []WebDep {
	if len(webs) < 2 {
		return nil
	}

	// Map each incomplete requirement ID to its web index
	reqToWeb := make(map[string]int)
	for i, web := range webs {
		for _, id := range web.IDs {
			reqToWeb[id] = i
		}
	}

	// For each incomplete req, find all reachable incomplete reqs via dependency
	// edges (traversing through complete intermediaries). If any reachable req
	// is in a different web, that creates a cross-web dependency.
	edgeSet := make(map[[2]int]bool)
	var deps []WebDep

	for _, req := range g.db.All() {
		if !req.IsIncomplete() {
			continue
		}
		toWeb, ok := reqToWeb[req.ReqID]
		if !ok {
			continue
		}

		// BFS through dependency chain, including complete intermediaries
		visited := make(map[string]bool)
		queue := make([]string, 0)
		for dep := range req.Dependencies {
			if !visited[dep] {
				visited[dep] = true
				queue = append(queue, dep)
			}
		}

		for len(queue) > 0 {
			cur := queue[0]
			queue = queue[1:]

			// If cur is incomplete and in a different web, record cross-web dep
			if fromWeb, ok2 := reqToWeb[cur]; ok2 && fromWeb != toWeb {
				key := [2]int{fromWeb, toWeb}
				if !edgeSet[key] {
					edgeSet[key] = true
					deps = append(deps, WebDep{From: fromWeb, To: toWeb})
				}
				continue // don't traverse further past an incomplete node
			}

			// If cur is complete, traverse its dependencies to find
			// transitive cross-web deps through complete intermediaries
			curReq := g.db.Get(cur)
			if curReq != nil && !curReq.IsIncomplete() {
				for dep := range curReq.Dependencies {
					if !visited[dep] {
						visited[dep] = true
						queue = append(queue, dep)
					}
				}
			}
		}
	}

	// Sort for stable output
	sort.Slice(deps, func(i, j int) bool {
		if deps[i].From != deps[j].From {
			return deps[i].From < deps[j].From
		}
		return deps[i].To < deps[j].To
	})

	return deps
}

// MergeOrder computes a safe merge ordering for work webs, respecting
// cross-web dependencies and separating webs with file overlaps.
// Returns web indices in topological order. REQ-ORCH-011.
func (g *Graph) MergeOrder(webs []Web) []int {
	n := len(webs)
	if n == 0 {
		return nil
	}

	// Build adjacency and in-degree from web dependencies
	deps := g.WebDependencies(webs)
	adj := make(map[int][]int)
	inDeg := make(map[int]int)
	for i := 0; i < n; i++ {
		inDeg[i] = 0
	}
	for _, d := range deps {
		adj[d.From] = append(adj[d.From], d.To)
		inDeg[d.To]++
	}

	// Also add ordering edges for overlapping webs (lower index first for stability)
	overlaps := g.DetectOverlaps(webs)
	overlapSet := make(map[[2]int]bool)
	for _, ov := range overlaps {
		key := [2]int{ov.WebA, ov.WebB}
		if !overlapSet[key] {
			overlapSet[key] = true
			// Only add if no existing dep edge in either direction
			fwd := [2]int{ov.WebA, ov.WebB}
			rev := [2]int{ov.WebB, ov.WebA}
			hasFwd := false
			hasRev := false
			for _, d := range deps {
				if d.From == fwd[0] && d.To == fwd[1] {
					hasFwd = true
				}
				if d.From == rev[0] && d.To == rev[1] {
					hasRev = true
				}
			}
			if !hasFwd && !hasRev {
				adj[ov.WebA] = append(adj[ov.WebA], ov.WebB)
				inDeg[ov.WebB]++
			}
		}
	}

	// Kahn's topological sort
	var queue []int
	for i := 0; i < n; i++ {
		if inDeg[i] == 0 {
			queue = append(queue, i)
		}
	}
	sort.Ints(queue) // stable ordering

	var order []int
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		order = append(order, cur)
		for _, next := range adj[cur] {
			inDeg[next]--
			if inDeg[next] == 0 {
				queue = append(queue, next)
				sort.Ints(queue) // keep stable
			}
		}
	}

	// If cycle detected, append remaining nodes
	if len(order) < n {
		inOrder := make(map[int]bool)
		for _, idx := range order {
			inOrder[idx] = true
		}
		for i := 0; i < n; i++ {
			if !inOrder[i] {
				order = append(order, i)
			}
		}
	}

	return order
}

// ParallelGroup represents a set of webs that can execute concurrently.
type ParallelGroup struct {
	WebIndices []int
}

// ParallelGroups partitions webs into parallel execution groups where
// no two webs in the same group have cross-web dependencies or file overlaps.
// Returns minimum groups maximizing parallelism. REQ-ORCH-012.
func (g *Graph) ParallelGroups(webs []Web) []ParallelGroup {
	n := len(webs)
	if n == 0 {
		return nil
	}

	// Build conflict graph: webs that cannot be in the same group
	conflicts := make(map[int]map[int]bool)
	for i := 0; i < n; i++ {
		conflicts[i] = make(map[int]bool)
	}

	// Cross-web dependencies create conflicts
	deps := g.WebDependencies(webs)
	for _, d := range deps {
		conflicts[d.From][d.To] = true
		conflicts[d.To][d.From] = true
	}

	// File overlaps create conflicts
	overlaps := g.DetectOverlaps(webs)
	for _, ov := range overlaps {
		conflicts[ov.WebA][ov.WebB] = true
		conflicts[ov.WebB][ov.WebA] = true
	}

	// Greedy coloring in merge order for stability
	order := g.MergeOrder(webs)
	color := make(map[int]int) // web index -> group index
	maxColor := 0

	for _, idx := range order {
		// Find smallest color not used by any conflicting neighbor
		usedColors := make(map[int]bool)
		for neighbor := range conflicts[idx] {
			if c, ok := color[neighbor]; ok {
				usedColors[c] = true
			}
		}
		c := 0
		for usedColors[c] {
			c++
		}
		color[idx] = c
		if c > maxColor {
			maxColor = c
		}
	}

	// Build groups
	groups := make([]ParallelGroup, maxColor+1)
	for _, idx := range order {
		c := color[idx]
		groups[c].WebIndices = append(groups[c].WebIndices, idx)
	}

	return groups
}

// WebOverlap describes an implicit coupling between two webs
// via shared file surface.
type WebOverlap struct {
	WebA         int      // index of first web
	WebB         int      // index of second web
	SharedFiles  []string // files touched by both webs
}

// DetectOverlaps finds webs that touch the same source files, indicating
// implicit coupling that could cause merge conflicts. File surface is
// derived from the test_module field of each requirement.
func (g *Graph) DetectOverlaps(webs []Web) []WebOverlap {
	if len(webs) < 2 {
		return nil
	}

	// Build file -> web index mapping
	fileToWebs := make(map[string][]int) // file -> list of web indices

	for i, web := range webs {
		seen := make(map[string]bool)
		for _, id := range web.IDs {
			req := g.db.Get(id)
			if req == nil {
				continue
			}
			// Use test_module as file surface indicator
			if req.TestModule != "" {
				// Normalize to directory level for broader overlap detection
				dir := req.TestModule
				if idx := strings.LastIndex(dir, "/"); idx > 0 {
					dir = dir[:idx]
				}
				if !seen[dir] {
					seen[dir] = true
					fileToWebs[dir] = append(fileToWebs[dir], i)
				}
			}
		}
	}

	// Find overlaps: files referenced by multiple webs
	overlapMap := make(map[[2]int][]string) // [webA, webB] -> shared files
	for file, webIdxs := range fileToWebs {
		if len(webIdxs) < 2 {
			continue
		}
		// Report all pairs
		for a := 0; a < len(webIdxs); a++ {
			for b := a + 1; b < len(webIdxs); b++ {
				key := [2]int{webIdxs[a], webIdxs[b]}
				overlapMap[key] = append(overlapMap[key], file)
			}
		}
	}

	var overlaps []WebOverlap
	for key, files := range overlapMap {
		sort.Strings(files)
		overlaps = append(overlaps, WebOverlap{
			WebA:        key[0],
			WebB:        key[1],
			SharedFiles: files,
		})
	}

	// Sort for stable output
	sort.Slice(overlaps, func(i, j int) bool {
		if overlaps[i].WebA != overlaps[j].WebA {
			return overlaps[i].WebA < overlaps[j].WebA
		}
		return overlaps[i].WebB < overlaps[j].WebB
	})

	return overlaps
}
