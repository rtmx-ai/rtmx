package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx-go/internal/database"
)

// TestFixtures validates the test fixtures functionality.
// REQ-GO-063: Go CLI shall provide test fixtures package
func TestFixtures(t *testing.T) {
	// Test that we can create a test database
	db := NewTestDatabase(t)
	if db == nil {
		t.Fatal("NewTestDatabase returned nil")
	}
	if db.Len() != 0 {
		t.Errorf("Expected empty database, got %d requirements", db.Len())
	}
}

func TestNewTestRequirement(t *testing.T) {
	req := NewTestRequirement("REQ-TEST-001")

	if req.ReqID != "REQ-TEST-001" {
		t.Errorf("Expected ID REQ-TEST-001, got %s", req.ReqID)
	}

	if req.Status != database.StatusMissing {
		t.Errorf("Expected status MISSING, got %s", req.Status)
	}

	if req.Priority != database.PriorityMedium {
		t.Errorf("Expected priority MEDIUM, got %s", req.Priority)
	}
}

func TestNewTestRequirementWithOptions(t *testing.T) {
	req := NewTestRequirement("REQ-TEST-002",
		WithStatus(database.StatusComplete),
		WithPriority(database.PriorityHigh),
		WithPhase(5),
		WithCategory("CUSTOM"),
		WithText("Custom requirement text"),
		WithDependencies("REQ-DEP-001", "REQ-DEP-002"),
		WithEffort(2.5),
	)

	if req.Status != database.StatusComplete {
		t.Errorf("Expected status COMPLETE, got %s", req.Status)
	}

	if req.Priority != database.PriorityHigh {
		t.Errorf("Expected priority HIGH, got %s", req.Priority)
	}

	if req.Phase != 5 {
		t.Errorf("Expected phase 5, got %d", req.Phase)
	}

	if req.Category != "CUSTOM" {
		t.Errorf("Expected category CUSTOM, got %s", req.Category)
	}

	if req.RequirementText != "Custom requirement text" {
		t.Errorf("Expected custom text, got %s", req.RequirementText)
	}

	if req.Dependencies.Len() != 2 {
		t.Errorf("Expected 2 dependencies, got %d", req.Dependencies.Len())
	}

	if req.EffortWeeks != 2.5 {
		t.Errorf("Expected effort 2.5, got %f", req.EffortWeeks)
	}
}

func TestNewTestDatabaseWithRequirements(t *testing.T) {
	req1 := NewTestRequirement("REQ-001")
	req2 := NewTestRequirement("REQ-002")

	db := NewTestDatabase(t, WithRequirements(req1, req2))

	if db.Len() != 2 {
		t.Errorf("Expected 2 requirements, got %d", db.Len())
	}
}

func TestNewTestConfig(t *testing.T) {
	cfg := NewTestConfig(t)

	if cfg == nil {
		t.Fatal("NewTestConfig returned nil")
	}
}

func TestNewTestConfigWithOptions(t *testing.T) {
	cfg := NewTestConfig(t,
		WithDatabasePath("custom/path/database.csv"),
		WithPhaseDescription(1, "Foundation"),
		WithGitHubAdapter("owner/repo", "MY_TOKEN"),
	)

	if cfg.RTMX.Database != "custom/path/database.csv" {
		t.Errorf("Expected custom database path, got %s", cfg.RTMX.Database)
	}

	if cfg.RTMX.Phases[1] != "Foundation" {
		t.Errorf("Expected phase 1 description 'Foundation', got %s", cfg.RTMX.Phases[1])
	}

	if !cfg.RTMX.Adapters.GitHub.Enabled {
		t.Error("Expected GitHub adapter to be enabled")
	}

	if cfg.RTMX.Adapters.GitHub.Repo != "owner/repo" {
		t.Errorf("Expected repo 'owner/repo', got %s", cfg.RTMX.Adapters.GitHub.Repo)
	}
}

func TestTempProject(t *testing.T) {
	dir, cleanup := TempProject(t)
	defer cleanup()

	// Verify directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Temp directory does not exist")
	}

	// Verify .rtmx directory exists
	rtmxDir := filepath.Join(dir, ".rtmx")
	if _, err := os.Stat(rtmxDir); os.IsNotExist(err) {
		t.Error(".rtmx directory does not exist")
	}
}

func TestTempProjectCleanup(t *testing.T) {
	dir, cleanup := TempProject(t)
	cleanup()

	// Verify directory is removed
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("Temp directory was not cleaned up")
	}
}

func TestTempProjectWithDatabase(t *testing.T) {
	db := NewTestDatabase(t, WithRequirement(NewTestRequirement("REQ-001")))

	dir, cleanup := TempProjectWithDatabase(t, db)
	defer cleanup()

	// Verify database file exists
	dbPath := filepath.Join(dir, ".rtmx", "database.csv")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file does not exist")
	}

	// Verify we can load the database
	loadedDB, err := database.Load(dbPath)
	if err != nil {
		t.Fatalf("Failed to load database: %v", err)
	}

	if loadedDB.Len() != 1 {
		t.Errorf("Expected 1 requirement, got %d", loadedDB.Len())
	}
}

func TestSampleRequirements(t *testing.T) {
	reqs := SampleRequirements()

	if len(reqs) != 5 {
		t.Errorf("Expected 5 sample requirements, got %d", len(reqs))
	}

	// Verify variety of statuses
	statusCounts := make(map[database.Status]int)
	for _, req := range reqs {
		statusCounts[req.Status]++
	}

	if statusCounts[database.StatusComplete] != 2 {
		t.Errorf("Expected 2 complete requirements, got %d", statusCounts[database.StatusComplete])
	}

	if statusCounts[database.StatusPartial] != 1 {
		t.Errorf("Expected 1 partial requirement, got %d", statusCounts[database.StatusPartial])
	}

	if statusCounts[database.StatusMissing] != 2 {
		t.Errorf("Expected 2 missing requirements, got %d", statusCounts[database.StatusMissing])
	}
}

func TestSampleDatabaseWithCycle(t *testing.T) {
	db := SampleDatabaseWithCycle(t)

	if db.Len() != 3 {
		t.Errorf("Expected 3 requirements, got %d", db.Len())
	}

	// These requirements form a cycle: A -> C -> B -> A
	reqs := db.All()
	var hasCycle bool
	for _, req := range reqs {
		if len(req.Dependencies) > 0 {
			hasCycle = true
			break
		}
	}

	if !hasCycle {
		t.Error("Expected requirements with dependencies")
	}
}

func TestSampleDatabaseNoCycle(t *testing.T) {
	db := SampleDatabaseNoCycle(t)

	if db.Len() != 5 {
		t.Errorf("Expected 5 requirements, got %d", db.Len())
	}
}

func TestWithBlocks(t *testing.T) {
	req := NewTestRequirement("REQ-001", WithBlocks("REQ-002", "REQ-003"))

	if req.Blocks.Len() != 2 {
		t.Errorf("Expected 2 blocks, got %d", req.Blocks.Len())
	}

	if !req.Blocks.Contains("REQ-002") || !req.Blocks.Contains("REQ-003") {
		t.Errorf("Unexpected blocks: %v", req.Blocks)
	}
}
