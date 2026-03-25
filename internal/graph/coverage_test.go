package graph

import (
	"testing"

	"github.com/rtmx-ai/rtmx-go/internal/database"
)

// TestIsIncomplete tests the IsIncomplete method on the graph.
func TestIsIncomplete(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	// A is MISSING (incomplete)
	if !g.IsIncomplete("A") {
		t.Error("A should be incomplete")
	}

	// C is COMPLETE
	if g.IsIncomplete("C") {
		t.Error("C should not be incomplete")
	}

	// Non-existent node
	if g.IsIncomplete("NONEXISTENT") {
		t.Error("Non-existent node should not be incomplete")
	}
}

// TestBlockingDependencies tests the BlockingDependencies method.
func TestBlockingDependencies(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	// D depends on B (incomplete) and C (complete)
	blocking := g.BlockingDependencies("D")
	if len(blocking) != 1 {
		t.Errorf("D should have 1 blocking dependency, got %d: %v", len(blocking), blocking)
	}
	if len(blocking) > 0 && blocking[0] != "B" {
		t.Errorf("D's blocking dependency should be B, got %s", blocking[0])
	}

	// A has no dependencies, so no blocking
	blocking = g.BlockingDependencies("A")
	if len(blocking) != 0 {
		t.Errorf("A should have 0 blocking dependencies, got %d", len(blocking))
	}
}

// TestStatistics tests the Statistics method.
func TestStatistics(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	stats := g.Statistics()
	if stats["nodes"] != 5 {
		t.Errorf("stats[nodes] = %v, want 5", stats["nodes"])
	}
	if stats["edges"] != 4 {
		t.Errorf("stats[edges] = %v, want 4", stats["edges"])
	}
	if stats["roots"] != 2 {
		t.Errorf("stats[roots] = %v, want 2", stats["roots"])
	}
	if stats["leaves"] != 2 {
		t.Errorf("stats[leaves] = %v, want 2", stats["leaves"])
	}
	avgDeps, ok := stats["avg_dependencies"].(float64)
	if !ok || avgDeps < 0 {
		t.Errorf("stats[avg_dependencies] = %v, want positive float", stats["avg_dependencies"])
	}
}

// TestStatisticsEmptyGraph tests Statistics on an empty graph.
func TestStatisticsEmptyGraph(t *testing.T) {
	db := database.NewDatabase()
	g := NewGraph(db)

	stats := g.Statistics()
	if stats["nodes"] != 0 {
		t.Errorf("stats[nodes] = %v, want 0", stats["nodes"])
	}
	avgDeps := stats["avg_dependencies"].(float64)
	if avgDeps != 0.0 {
		t.Errorf("stats[avg_dependencies] = %v, want 0", avgDeps)
	}
}

// TestBottleneckRequirements tests the BottleneckRequirements method.
func TestBottleneckRequirements(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	// A blocks B and D (transitively), so A is a bottleneck
	bottlenecks := g.BottleneckRequirements(1)
	if len(bottlenecks) == 0 {
		t.Error("Should have at least 1 bottleneck requirement")
	}

	// High threshold - should find fewer
	bottlenecks = g.BottleneckRequirements(100)
	if len(bottlenecks) != 0 {
		t.Errorf("Should have no bottlenecks with min=100, got %d", len(bottlenecks))
	}
}

// TestFindCyclePath tests the FindCyclePath method.
func TestFindCyclePath(t *testing.T) {
	db := createCyclicDB()
	g := NewGraph(db)

	cycles := g.FindCycles()
	if len(cycles) == 0 {
		t.Fatal("Expected at least 1 cycle")
	}

	path := g.FindCyclePath(cycles[0])
	if len(path) == 0 {
		t.Error("FindCyclePath should return non-empty path")
	}

	// Path should start and end with the same node (a cycle)
	if len(path) > 1 && path[0] != path[len(path)-1] {
		t.Errorf("Cycle path should start and end with same node, got %v", path)
	}
}

// TestFindCyclePathEmpty tests FindCyclePath with empty input.
func TestFindCyclePathEmpty(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	path := g.FindCyclePath([]string{})
	if path != nil {
		t.Error("FindCyclePath with empty input should return nil")
	}
}

// TestExecutionOrder tests the ExecutionOrder method.
func TestExecutionOrder(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	order := g.ExecutionOrder()
	if order == nil {
		t.Fatal("ExecutionOrder should not return nil for acyclic graph")
	}
	if len(order) != 5 {
		t.Errorf("ExecutionOrder should return 5 nodes, got %d", len(order))
	}
}

// TestExecutionOrderCyclic tests ExecutionOrder on a cyclic graph.
func TestExecutionOrderCyclic(t *testing.T) {
	db := createCyclicDB()
	g := NewGraph(db)

	order := g.ExecutionOrder()
	if order != nil {
		t.Error("ExecutionOrder should return nil for cyclic graph")
	}
}

// TestCriticalPathAllComplete tests CriticalPath when all requirements are complete.
func TestCriticalPathAllComplete(t *testing.T) {
	db := database.NewDatabase()

	reqs := []*database.Requirement{
		{ReqID: "A", Category: "T", Status: database.StatusComplete, Priority: database.PriorityHigh,
			Dependencies: database.NewStringSet(), Blocks: database.NewStringSet("B"), Extra: make(map[string]string)},
		{ReqID: "B", Category: "T", Status: database.StatusComplete, Priority: database.PriorityMedium,
			Dependencies: database.NewStringSet("A"), Blocks: database.NewStringSet(), Extra: make(map[string]string)},
	}
	for _, req := range reqs {
		_ = db.Add(req)
	}

	g := NewGraph(db)
	path := g.CriticalPath()
	if len(path) != 0 {
		t.Errorf("CriticalPath should be empty when all complete, got %v", path)
	}
}
