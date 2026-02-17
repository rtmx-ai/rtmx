// Package testutil provides test utilities and fixtures for RTMX Go CLI testing.
package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
)

// DatabaseOption configures a test database.
type DatabaseOption func(*database.Database)

// NewTestDatabase creates a database for testing with optional configuration.
func NewTestDatabase(t *testing.T, opts ...DatabaseOption) *database.Database {
	t.Helper()

	db := database.NewDatabase()

	for _, opt := range opts {
		opt(db)
	}

	return db
}

// WithRequirement adds a requirement to the database.
func WithRequirement(req *database.Requirement) DatabaseOption {
	return func(db *database.Database) {
		_ = db.Add(req)
	}
}

// WithRequirements adds multiple requirements to the database.
func WithRequirements(reqs ...*database.Requirement) DatabaseOption {
	return func(db *database.Database) {
		for _, req := range reqs {
			_ = db.Add(req)
		}
	}
}

// RequirementOption configures a test requirement.
type RequirementOption func(*database.Requirement)

// NewTestRequirement creates a requirement for testing with optional configuration.
func NewTestRequirement(id string, opts ...RequirementOption) *database.Requirement {
	req := database.NewRequirement(id)
	req.Category = "TEST"
	req.Subcategory = "Unit"
	req.RequirementText = "Test requirement " + id
	req.TargetValue = "Test passes"
	req.TestModule = "test_file.go"
	req.TestFunction = "TestFunction"
	req.ValidationMethod = "Unit Test"
	req.Status = database.StatusMissing
	req.Priority = database.PriorityMedium
	req.Phase = 1
	req.EffortWeeks = 1.0

	for _, opt := range opts {
		opt(req)
	}

	return req
}

// WithStatus sets the requirement status.
func WithStatus(status database.Status) RequirementOption {
	return func(r *database.Requirement) {
		r.Status = status
	}
}

// WithPriority sets the requirement priority.
func WithPriority(priority database.Priority) RequirementOption {
	return func(r *database.Requirement) {
		r.Priority = priority
	}
}

// WithPhase sets the requirement phase.
func WithPhase(phase int) RequirementOption {
	return func(r *database.Requirement) {
		r.Phase = phase
	}
}

// WithCategory sets the requirement category.
func WithCategory(category string) RequirementOption {
	return func(r *database.Requirement) {
		r.Category = category
	}
}

// WithText sets the requirement text.
func WithText(text string) RequirementOption {
	return func(r *database.Requirement) {
		r.RequirementText = text
	}
}

// WithDependencies sets the requirement dependencies.
func WithDependencies(deps ...string) RequirementOption {
	return func(r *database.Requirement) {
		r.Dependencies = database.NewStringSet(deps...)
	}
}

// WithBlocks sets the requirements that this one blocks.
func WithBlocks(blocks ...string) RequirementOption {
	return func(r *database.Requirement) {
		r.Blocks = database.NewStringSet(blocks...)
	}
}

// WithEffort sets the requirement effort in weeks.
func WithEffort(effort float64) RequirementOption {
	return func(r *database.Requirement) {
		r.EffortWeeks = effort
	}
}

// ConfigOption configures a test config.
type ConfigOption func(*config.Config)

// NewTestConfig creates a config for testing with optional configuration.
func NewTestConfig(t *testing.T, opts ...ConfigOption) *config.Config {
	t.Helper()

	cfg := config.DefaultConfig()

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// WithDatabasePath sets the database path in the config.
func WithDatabasePath(path string) ConfigOption {
	return func(c *config.Config) {
		c.RTMX.Database = path
	}
}

// WithPhaseDescription adds a phase description to the config.
func WithPhaseDescription(phase int, description string) ConfigOption {
	return func(c *config.Config) {
		if c.RTMX.Phases == nil {
			c.RTMX.Phases = make(map[int]string)
		}
		c.RTMX.Phases[phase] = description
	}
}

// WithGitHubAdapter enables and configures the GitHub adapter.
func WithGitHubAdapter(repo, tokenEnv string) ConfigOption {
	return func(c *config.Config) {
		c.RTMX.Adapters.GitHub.Enabled = true
		c.RTMX.Adapters.GitHub.Repo = repo
		c.RTMX.Adapters.GitHub.TokenEnv = tokenEnv
	}
}

// TempProject creates a temporary directory with RTMX structure for testing.
// Returns the directory path and a cleanup function.
func TempProject(t *testing.T) (string, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", "rtmx-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	// Create .rtmx directory
	rtmxDir := filepath.Join(dir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0755); err != nil {
		os.RemoveAll(dir)
		t.Fatalf("Failed to create .rtmx directory: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return dir, cleanup
}

// TempProjectWithConfig creates a temp project with a config file.
func TempProjectWithConfig(t *testing.T, cfg *config.Config) (string, func()) {
	t.Helper()

	dir, cleanup := TempProject(t)

	// Write config
	configPath := filepath.Join(dir, ".rtmx", "config.yaml")
	if err := cfg.Save(configPath); err != nil {
		cleanup()
		t.Fatalf("Failed to write config: %v", err)
	}

	return dir, cleanup
}

// TempProjectWithDatabase creates a temp project with a database.
func TempProjectWithDatabase(t *testing.T, db *database.Database) (string, func()) {
	t.Helper()

	dir, cleanup := TempProject(t)

	// Write database
	dbPath := filepath.Join(dir, ".rtmx", "database.csv")
	if err := db.Save(dbPath); err != nil {
		cleanup()
		t.Fatalf("Failed to write database: %v", err)
	}

	return dir, cleanup
}

// TempProjectFull creates a temp project with both config and database.
func TempProjectFull(t *testing.T, cfg *config.Config, db *database.Database) (string, func()) {
	t.Helper()

	dir, cleanup := TempProject(t)

	// Write config
	configPath := filepath.Join(dir, ".rtmx", "config.yaml")
	if err := cfg.Save(configPath); err != nil {
		cleanup()
		t.Fatalf("Failed to write config: %v", err)
	}

	// Write database
	dbPath := filepath.Join(dir, ".rtmx", "database.csv")
	if err := db.Save(dbPath); err != nil {
		cleanup()
		t.Fatalf("Failed to write database: %v", err)
	}

	return dir, cleanup
}

// SampleRequirements returns a set of sample requirements for testing.
func SampleRequirements() []*database.Requirement {
	return []*database.Requirement{
		NewTestRequirement("REQ-TEST-001", WithStatus(database.StatusComplete), WithPhase(1)),
		NewTestRequirement("REQ-TEST-002", WithStatus(database.StatusPartial), WithPhase(1)),
		NewTestRequirement("REQ-TEST-003", WithStatus(database.StatusMissing), WithPhase(2)),
		NewTestRequirement("REQ-TEST-004", WithStatus(database.StatusMissing), WithPhase(2), WithDependencies("REQ-TEST-001")),
		NewTestRequirement("REQ-TEST-005", WithStatus(database.StatusComplete), WithPhase(3), WithDependencies("REQ-TEST-003", "REQ-TEST-004")),
	}
}

// SampleDatabaseWithCycle returns a database with a dependency cycle for testing.
func SampleDatabaseWithCycle(t *testing.T) *database.Database {
	t.Helper()

	return NewTestDatabase(t,
		WithRequirement(NewTestRequirement("REQ-CYCLE-A", WithDependencies("REQ-CYCLE-C"))),
		WithRequirement(NewTestRequirement("REQ-CYCLE-B", WithDependencies("REQ-CYCLE-A"))),
		WithRequirement(NewTestRequirement("REQ-CYCLE-C", WithDependencies("REQ-CYCLE-B"))),
	)
}

// SampleDatabaseNoCycle returns a database without dependency cycles.
func SampleDatabaseNoCycle(t *testing.T) *database.Database {
	t.Helper()

	return NewTestDatabase(t, WithRequirements(SampleRequirements()...))
}
