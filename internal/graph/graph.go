// Package graph provides dependency graph algorithms for RTMX.
package graph

import (
	"github.com/rtmx-ai/rtmx-go/internal/database"
)

// Graph represents a dependency graph of requirements.
type Graph struct {
	db *database.Database

	// Adjacency lists
	dependencies map[string][]string // req -> what it depends on
	dependents   map[string][]string // req -> what depends on it
}

// NewGraph creates a new dependency graph from a database.
func NewGraph(db *database.Database) *Graph {
	g := &Graph{
		db:           db,
		dependencies: make(map[string][]string),
		dependents:   make(map[string][]string),
	}

	// Build adjacency lists
	for _, req := range db.All() {
		g.dependencies[req.ReqID] = make([]string, 0)
		for dep := range req.Dependencies {
			// Only include local dependencies (skip cross-repo)
			if db.Exists(dep) {
				g.dependencies[req.ReqID] = append(g.dependencies[req.ReqID], dep)
				g.dependents[dep] = append(g.dependents[dep], req.ReqID)
			}
		}
	}

	return g
}

// NodeCount returns the number of nodes in the graph.
func (g *Graph) NodeCount() int {
	return g.db.Len()
}

// EdgeCount returns the number of edges in the graph.
func (g *Graph) EdgeCount() int {
	count := 0
	for _, deps := range g.dependencies {
		count += len(deps)
	}
	return count
}

// Dependencies returns the direct dependencies of a requirement.
func (g *Graph) Dependencies(reqID string) []string {
	return g.dependencies[reqID]
}

// Dependents returns requirements that directly depend on this one.
func (g *Graph) Dependents(reqID string) []string {
	return g.dependents[reqID]
}

// TransitiveDependencies returns all requirements that this one depends on (directly or indirectly).
func (g *Graph) TransitiveDependencies(reqID string) []string {
	visited := make(map[string]bool)
	var result []string

	var dfs func(id string)
	dfs = func(id string) {
		for _, dep := range g.dependencies[id] {
			if !visited[dep] {
				visited[dep] = true
				result = append(result, dep)
				dfs(dep)
			}
		}
	}

	dfs(reqID)
	return result
}

// TransitiveDependents returns all requirements that depend on this one (directly or indirectly).
func (g *Graph) TransitiveDependents(reqID string) []string {
	visited := make(map[string]bool)
	var result []string

	var dfs func(id string)
	dfs = func(id string) {
		for _, dep := range g.dependents[id] {
			if !visited[dep] {
				visited[dep] = true
				result = append(result, dep)
				dfs(dep)
			}
		}
	}

	dfs(reqID)
	return result
}

// IsBlocked returns true if any dependency of this requirement is incomplete.
func (g *Graph) IsBlocked(reqID string) bool {
	for _, dep := range g.dependencies[reqID] {
		if req := g.db.Get(dep); req != nil && req.IsIncomplete() {
			return true
		}
	}
	return false
}

// BlockingDependencies returns incomplete dependencies of a requirement.
func (g *Graph) BlockingDependencies(reqID string) []string {
	var blocking []string
	for _, dep := range g.dependencies[reqID] {
		if req := g.db.Get(dep); req != nil && req.IsIncomplete() {
			blocking = append(blocking, dep)
		}
	}
	return blocking
}

// Roots returns requirements with no dependencies.
func (g *Graph) Roots() []string {
	var roots []string
	for _, req := range g.db.All() {
		if len(g.dependencies[req.ReqID]) == 0 {
			roots = append(roots, req.ReqID)
		}
	}
	return roots
}

// Leaves returns requirements with no dependents.
func (g *Graph) Leaves() []string {
	var leaves []string
	for _, req := range g.db.All() {
		if len(g.dependents[req.ReqID]) == 0 {
			leaves = append(leaves, req.ReqID)
		}
	}
	return leaves
}

// Statistics returns graph statistics.
func (g *Graph) Statistics() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["nodes"] = g.NodeCount()
	stats["edges"] = g.EdgeCount()
	stats["roots"] = len(g.Roots())
	stats["leaves"] = len(g.Leaves())

	// Average dependencies
	totalDeps := 0
	for _, deps := range g.dependencies {
		totalDeps += len(deps)
	}
	if g.NodeCount() > 0 {
		stats["avg_dependencies"] = float64(totalDeps) / float64(g.NodeCount())
	} else {
		stats["avg_dependencies"] = 0.0
	}

	return stats
}
