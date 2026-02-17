package graph

import (
	"testing"

	"github.com/rtmx-ai/rtmx-go/internal/database"
)

func createTestDB() *database.Database {
	db := database.NewDatabase()

	// Create a simple dependency graph:
	// A -> B -> D
	// A -> C -> D
	// E (isolated)

	reqs := []*database.Requirement{
		{ReqID: "A", Category: "TEST", Status: database.StatusMissing, Priority: database.PriorityHigh,
			Dependencies: database.NewStringSet(), Blocks: database.NewStringSet("B", "C")},
		{ReqID: "B", Category: "TEST", Status: database.StatusMissing, Priority: database.PriorityMedium,
			Dependencies: database.NewStringSet("A"), Blocks: database.NewStringSet("D")},
		{ReqID: "C", Category: "TEST", Status: database.StatusComplete, Priority: database.PriorityMedium,
			Dependencies: database.NewStringSet("A"), Blocks: database.NewStringSet("D")},
		{ReqID: "D", Category: "TEST", Status: database.StatusMissing, Priority: database.PriorityLow,
			Dependencies: database.NewStringSet("B", "C"), Blocks: database.NewStringSet()},
		{ReqID: "E", Category: "TEST", Status: database.StatusMissing, Priority: database.PriorityLow,
			Dependencies: database.NewStringSet(), Blocks: database.NewStringSet()},
	}

	for _, req := range reqs {
		req.Extra = make(map[string]string)
		_ = db.Add(req)
	}

	return db
}

func createCyclicDB() *database.Database {
	db := database.NewDatabase()

	// Create a cycle: A -> B -> C -> A
	reqs := []*database.Requirement{
		{ReqID: "A", Category: "TEST", Status: database.StatusMissing, Priority: database.PriorityHigh,
			Dependencies: database.NewStringSet("C"), Blocks: database.NewStringSet("B")},
		{ReqID: "B", Category: "TEST", Status: database.StatusMissing, Priority: database.PriorityMedium,
			Dependencies: database.NewStringSet("A"), Blocks: database.NewStringSet("C")},
		{ReqID: "C", Category: "TEST", Status: database.StatusMissing, Priority: database.PriorityMedium,
			Dependencies: database.NewStringSet("B"), Blocks: database.NewStringSet("A")},
	}

	for _, req := range reqs {
		req.Extra = make(map[string]string)
		_ = db.Add(req)
	}

	return db
}

func TestGraphBasics(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	if g.NodeCount() != 5 {
		t.Errorf("NodeCount = %d, want 5", g.NodeCount())
	}

	if g.EdgeCount() != 4 {
		t.Errorf("EdgeCount = %d, want 4", g.EdgeCount())
	}
}

func TestDependencies(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	deps := g.Dependencies("D")
	if len(deps) != 2 {
		t.Errorf("D should have 2 dependencies, got %d", len(deps))
	}

	deps = g.Dependencies("A")
	if len(deps) != 0 {
		t.Errorf("A should have 0 dependencies, got %d", len(deps))
	}
}

func TestDependents(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	deps := g.Dependents("A")
	if len(deps) != 2 {
		t.Errorf("A should have 2 dependents, got %d", len(deps))
	}

	deps = g.Dependents("D")
	if len(deps) != 0 {
		t.Errorf("D should have 0 dependents, got %d", len(deps))
	}
}

func TestTransitiveDependencies(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	deps := g.TransitiveDependencies("D")
	if len(deps) != 3 { // A, B, C
		t.Errorf("D should have 3 transitive dependencies, got %d: %v", len(deps), deps)
	}

	deps = g.TransitiveDependencies("A")
	if len(deps) != 0 {
		t.Errorf("A should have 0 transitive dependencies, got %d", len(deps))
	}
}

func TestTransitiveDependents(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	deps := g.TransitiveDependents("A")
	if len(deps) != 3 { // B, C, D
		t.Errorf("A should have 3 transitive dependents, got %d: %v", len(deps), deps)
	}

	deps = g.TransitiveDependents("D")
	if len(deps) != 0 {
		t.Errorf("D should have 0 transitive dependents, got %d", len(deps))
	}
}

func TestRootsAndLeaves(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	roots := g.Roots()
	if len(roots) != 2 { // A and E
		t.Errorf("Should have 2 roots (A and E), got %d: %v", len(roots), roots)
	}

	leaves := g.Leaves()
	if len(leaves) != 2 { // D and E
		t.Errorf("Should have 2 leaves (D and E), got %d: %v", len(leaves), leaves)
	}
}

func TestIsBlocked(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	// A has no dependencies, not blocked
	if g.IsBlocked("A") {
		t.Error("A should not be blocked")
	}

	// B depends on A which is incomplete
	if !g.IsBlocked("B") {
		t.Error("B should be blocked")
	}

	// D depends on B (incomplete) and C (complete), so blocked
	if !g.IsBlocked("D") {
		t.Error("D should be blocked")
	}
}

func TestFindCyclesNone(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	cycles := g.FindCycles()
	if len(cycles) != 0 {
		t.Errorf("Should have no cycles, got %d: %v", len(cycles), cycles)
	}

	if g.HasCycles() {
		t.Error("HasCycles should return false")
	}
}

func TestFindCyclesPresent(t *testing.T) {
	db := createCyclicDB()
	g := NewGraph(db)

	cycles := g.FindCycles()
	if len(cycles) != 1 {
		t.Errorf("Should have 1 cycle, got %d", len(cycles))
	}

	if !g.HasCycles() {
		t.Error("HasCycles should return true")
	}

	if len(cycles) > 0 && len(cycles[0]) != 3 {
		t.Errorf("Cycle should have 3 members, got %d", len(cycles[0]))
	}
}

func TestTopologicalSort(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	order := g.TopologicalSort()
	if order == nil {
		t.Fatal("TopologicalSort should not return nil for acyclic graph")
	}

	if len(order) != 5 {
		t.Errorf("TopologicalSort should return 5 nodes, got %d", len(order))
	}

	// Verify order is valid (dependencies before dependents)
	position := make(map[string]int)
	for i, node := range order {
		position[node] = i
	}

	for _, node := range order {
		for _, dep := range g.Dependencies(node) {
			if position[dep] > position[node] {
				t.Errorf("Invalid order: %s appears before its dependency %s", node, dep)
			}
		}
	}
}

func TestTopologicalSortCycle(t *testing.T) {
	db := createCyclicDB()
	g := NewGraph(db)

	order := g.TopologicalSort()
	if order != nil {
		t.Error("TopologicalSort should return nil for cyclic graph")
	}
}

func TestLayers(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	layers := g.Layers()
	if len(layers) < 2 {
		t.Errorf("Should have at least 2 layers, got %d", len(layers))
	}

	// Layer 0 should contain roots
	layer0 := layers[0]
	for _, node := range layer0 {
		if len(g.Dependencies(node)) != 0 {
			t.Errorf("Layer 0 node %s should have no dependencies", node)
		}
	}
}

func TestCriticalPath(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	path := g.CriticalPath()

	// A blocks B which blocks D, and A blocks C
	// So A should be on the critical path (blocks 2 incomplete: B and D)
	if len(path) == 0 {
		t.Error("Critical path should not be empty")
	}
}

func TestBlockingAnalysis(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	analysis := g.BlockingAnalysis()

	// A should block B and D (through B)
	if analysis["A"] < 1 {
		t.Errorf("A should block at least 1 requirement, got %d", analysis["A"])
	}
}

func TestUnblockedIncomplete(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	unblocked := g.UnblockedIncomplete()

	// A and E are incomplete and not blocked
	if len(unblocked) != 2 {
		t.Errorf("Should have 2 unblocked incomplete requirements, got %d: %v", len(unblocked), unblocked)
	}
}

func TestNextWorkable(t *testing.T) {
	db := createTestDB()
	g := NewGraph(db)

	workable := g.NextWorkable()

	// Should be same as UnblockedIncomplete
	if len(workable) != 2 {
		t.Errorf("Should have 2 workable requirements, got %d", len(workable))
	}
}

func TestLoadRealGraph(t *testing.T) {
	paths := []string{
		".rtmx/database.csv",
		"../../.rtmx/database.csv",
	}

	var db *database.Database
	var err error
	for _, path := range paths {
		db, err = database.Load(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Skipf("Skipping real graph test: %v", err)
	}

	g := NewGraph(db)

	t.Logf("Graph stats: %d nodes, %d edges", g.NodeCount(), g.EdgeCount())
	t.Logf("Roots: %d, Leaves: %d", len(g.Roots()), len(g.Leaves()))

	cycles := g.FindCycles()
	t.Logf("Cycles: %d", len(cycles))

	if g.HasCycles() {
		for i, cycle := range cycles {
			t.Logf("Cycle %d: %v", i+1, cycle)
		}
	}

	unblocked := g.UnblockedIncomplete()
	t.Logf("Unblocked incomplete: %d", len(unblocked))
	for _, id := range unblocked {
		t.Logf("  - %s", id)
	}
}
