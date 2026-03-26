package graph

import (
	"sort"
)

// CriticalPath returns requirements on the critical path.
// The critical path consists of requirements that block the most other incomplete requirements.
func (g *Graph) CriticalPath() []string {
	// Calculate how many incomplete requirements each incomplete requirement blocks
	blockCount := make(map[string]int)

	for _, req := range g.db.All() {
		if req.IsIncomplete() {
			// Count how many incomplete requirements depend on this one
			count := g.countBlockedIncomplete(req.ReqID)
			blockCount[req.ReqID] = count
		}
	}

	// Find requirements that block the most others
	type reqScore struct {
		id    string
		score int
	}
	var scores []reqScore
	for id, count := range blockCount {
		if count > 0 {
			scores = append(scores, reqScore{id, count})
		}
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Build critical path - start with highest blocker and trace dependencies
	if len(scores) == 0 {
		return nil
	}

	// Return all blockers sorted by impact
	result := make([]string, 0, len(scores))
	for _, s := range scores {
		result = append(result, s.id)
	}
	return result
}

// countBlockedIncomplete counts how many incomplete requirements are transitively blocked by this one
func (g *Graph) countBlockedIncomplete(reqID string) int {
	count := 0
	visited := make(map[string]bool)

	var dfs func(id string)
	dfs = func(id string) {
		for _, dependent := range g.dependents[id] {
			if !visited[dependent] {
				visited[dependent] = true
				if req := g.db.Get(dependent); req != nil && req.IsIncomplete() {
					count++
					dfs(dependent)
				}
			}
		}
	}

	dfs(reqID)
	return count
}

// BlockingAnalysis returns a map of requirement ID to the number of requirements it blocks.
func (g *Graph) BlockingAnalysis() map[string]int {
	analysis := make(map[string]int)

	for _, req := range g.db.All() {
		if req.IsIncomplete() {
			count := g.countBlockedIncomplete(req.ReqID)
			if count > 0 {
				analysis[req.ReqID] = count
			}
		}
	}

	return analysis
}

// BottleneckRequirements returns incomplete requirements that block more than n others.
func (g *Graph) BottleneckRequirements(minBlocked int) []string {
	var bottlenecks []string

	for _, req := range g.db.All() {
		if req.IsIncomplete() {
			count := g.countBlockedIncomplete(req.ReqID)
			if count >= minBlocked {
				bottlenecks = append(bottlenecks, req.ReqID)
			}
		}
	}

	// Sort by block count descending
	sort.Slice(bottlenecks, func(i, j int) bool {
		return g.countBlockedIncomplete(bottlenecks[i]) > g.countBlockedIncomplete(bottlenecks[j])
	})

	return bottlenecks
}

// UnblockedIncomplete returns incomplete requirements that have no blocking dependencies.
func (g *Graph) UnblockedIncomplete() []string {
	var unblocked []string

	for _, req := range g.db.All() {
		if req.IsIncomplete() && !g.IsBlocked(req.ReqID) {
			unblocked = append(unblocked, req.ReqID)
		}
	}

	return unblocked
}

// NextWorkable returns incomplete requirements that can be started immediately.
// These are requirements that have no blocking dependencies.
func (g *Graph) NextWorkable() []string {
	return g.UnblockedIncomplete()
}
