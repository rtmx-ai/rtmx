package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestDefaultConfig(t *testing.T) {
	rtmx.Req(t, "REQ-GO-008")
	cfg := DefaultConfig()

	if cfg.RTMX.Database != ".rtmx/database.csv" {
		t.Errorf("Default database = %q, want .rtmx/database.csv", cfg.RTMX.Database)
	}

	if cfg.RTMX.Schema != "core" {
		t.Errorf("Default schema = %q, want core", cfg.RTMX.Schema)
	}

	if !cfg.RTMX.Pytest.RegisterMarkers {
		t.Error("Default register_markers should be true")
	}

	if cfg.RTMX.Pytest.MarkerPrefix != "req" {
		t.Errorf("Default marker_prefix = %q, want req", cfg.RTMX.Pytest.MarkerPrefix)
	}

	if !cfg.RTMX.Agents.Claude.Enabled {
		t.Error("Default Claude agent should be enabled")
	}
}

func TestLoadConfig(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
rtmx:
  database: custom/database.csv
  schema: phoenix
  phases:
    1: Foundation
    2: Integration
    3: Testing
  adapters:
    github:
      enabled: true
      repo: rtmx-ai/rtmx
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.RTMX.Database != "custom/database.csv" {
		t.Errorf("Database = %q, want custom/database.csv", cfg.RTMX.Database)
	}

	if cfg.RTMX.Schema != "phoenix" {
		t.Errorf("Schema = %q, want phoenix", cfg.RTMX.Schema)
	}

	if !cfg.RTMX.Adapters.GitHub.Enabled {
		t.Error("GitHub adapter should be enabled")
	}

	if cfg.RTMX.Adapters.GitHub.Repo != "rtmx-ai/rtmx" {
		t.Errorf("GitHub repo = %q, want rtmx-ai/rtmx", cfg.RTMX.Adapters.GitHub.Repo)
	}

	// Verify phases
	if desc := cfg.RTMX.Phases[1]; desc != "Foundation" {
		t.Errorf("Phase 1 = %q, want Foundation", desc)
	}
}

func TestFindConfig(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0755); err != nil {
		t.Fatalf("Failed to create .rtmx dir: %v", err)
	}

	configPath := filepath.Join(rtmxDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("rtmx:\n  schema: test"), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	// Test finding config
	found, err := FindConfig(tmpDir)
	if err != nil {
		t.Fatalf("FindConfig failed: %v", err)
	}

	if found != configPath {
		t.Errorf("FindConfig = %q, want %q", found, configPath)
	}

	// Test finding from subdirectory
	subDir := filepath.Join(tmpDir, "src", "pkg")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	found, err = FindConfig(subDir)
	if err != nil {
		t.Fatalf("FindConfig from subdir failed: %v", err)
	}

	if found != configPath {
		t.Errorf("FindConfig from subdir = %q, want %q", found, configPath)
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".rtmx", "config.yaml")

	cfg := DefaultConfig()
	cfg.RTMX.Database = "custom.csv"

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("Config file not created: %v", err)
	}

	// Load and verify
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load after save failed: %v", err)
	}

	if loaded.RTMX.Database != "custom.csv" {
		t.Errorf("Loaded database = %q, want custom.csv", loaded.RTMX.Database)
	}
}

func TestDatabasePath(t *testing.T) {
	cfg := DefaultConfig()

	// Test relative path - use a temp directory for cross-platform compatibility
	tmpDir := t.TempDir()
	path := cfg.DatabasePath(tmpDir)
	expected := filepath.Join(tmpDir, ".rtmx", "database.csv")
	if path != expected {
		t.Errorf("DatabasePath = %q, want %q", path, expected)
	}

	// Test absolute path
	absPath := filepath.Join(tmpDir, "absolute", "path", "db.csv")
	cfg.RTMX.Database = absPath
	path = cfg.DatabasePath(tmpDir)
	if path != absPath {
		t.Errorf("Absolute DatabasePath = %q, want %s", path, absPath)
	}
}

func TestPhaseDescription(t *testing.T) {
	cfg := DefaultConfig()

	// Test existing phase
	desc := cfg.PhaseDescription(1)
	if desc != "Foundation" {
		t.Errorf("Phase 1 description = %q, want Foundation", desc)
	}

	// Test non-existing phase
	desc = cfg.PhaseDescription(99)
	if desc != "Phase 99" {
		t.Errorf("Phase 99 description = %q, want 'Phase 99'", desc)
	}
}

func TestLoadRealConfig(t *testing.T) {
	// Try to load the real config from rtmx-go project
	paths := []string{
		".rtmx/config.yaml",
		"../../.rtmx/config.yaml",
	}

	var cfg *Config
	var err error
	for _, path := range paths {
		cfg, err = Load(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Skipf("Skipping real config test: %v", err)
	}

	// Just verify it loaded without error and has expected structure
	if cfg.RTMX.Database == "" {
		t.Error("Real config should have database path")
	}

	t.Logf("Loaded real config: database=%s, schema=%s",
		cfg.RTMX.Database, cfg.RTMX.Schema)
}

func TestVersionPolicyIncrementLevel(t *testing.T) {
	policy := &VersionPolicyConfig{
		Categories: map[string]string{
			"CLI":  "minor",
			"DATA": "major",
			"BENCH": "none",
		},
		Subcategories: map[string]string{
			"CLI/Foundation": "major",
			"PLAN/Display":   "patch",
		},
		Default: "patch",
	}

	tests := []struct {
		cat, sub string
		want     string
	}{
		{"CLI", "Commands", "minor"},       // category match
		{"CLI", "Foundation", "major"},     // subcategory override
		{"DATA", "Config", "major"},        // category match
		{"BENCH", "Observability", "none"}, // category match
		{"PLAN", "Display", "patch"},       // subcategory override
		{"PLAN", "Release", "patch"},       // default (no category match)
		{"UNKNOWN", "", "patch"},           // default
	}

	for _, tt := range tests {
		name := tt.cat
		if tt.sub != "" {
			name += "/" + tt.sub
		}
		t.Run(name, func(t *testing.T) {
			got := policy.IncrementLevel(tt.cat, tt.sub)
			if got != tt.want {
				t.Errorf("IncrementLevel(%q, %q) = %q, want %q", tt.cat, tt.sub, got, tt.want)
			}
		})
	}
}

func TestVersionPolicyNil(t *testing.T) {
	var policy *VersionPolicyConfig
	if got := policy.IncrementLevel("CLI", "Commands"); got != "patch" {
		t.Errorf("nil policy IncrementLevel = %q, want patch", got)
	}
	if !policy.IsBackwardCompatible() {
		t.Error("nil policy should be backward compatible")
	}
	if policy.IsEnabled() {
		t.Error("nil policy should not be enabled")
	}
}

func TestVersionPolicyBackwardCompatible(t *testing.T) {
	yes := true
	no := false

	tests := []struct {
		name string
		bc   *bool
		want bool
	}{
		{"nil (default true)", nil, true},
		{"explicit true", &yes, true},
		{"explicit false", &no, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &VersionPolicyConfig{BackwardCompatible: tt.bc}
			if got := policy.IsBackwardCompatible(); got != tt.want {
				t.Errorf("IsBackwardCompatible = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVersionPolicyIsEnabled(t *testing.T) {
	tests := []struct {
		enforcement string
		want        bool
	}{
		{"", false},
		{"off", false},
		{"warn", true},
		{"enforce", true},
	}

	for _, tt := range tests {
		t.Run(tt.enforcement, func(t *testing.T) {
			policy := &VersionPolicyConfig{Enforcement: tt.enforcement}
			if got := policy.IsEnabled(); got != tt.want {
				t.Errorf("IsEnabled(%q) = %v, want %v", tt.enforcement, got, tt.want)
			}
		})
	}
}

func TestVersionPolicyFromYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rtmx-vp-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	configContent := `rtmx:
  database: test.csv
  schema: core
  version_policy:
    enforcement: warn
    backward_compatible: true
    default: patch
    categories:
      CLI: minor
      DATA: major
    subcategories:
      CLI/Foundation: major
`
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0755)
	configPath := filepath.Join(rtmxDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	vp := &cfg.RTMX.VersionPolicy
	if !vp.IsEnabled() {
		t.Error("version_policy with enforcement=warn should be enabled")
	}
	if got := vp.IncrementLevel("CLI", "Commands"); got != "minor" {
		t.Errorf("CLI/Commands = %q, want minor", got)
	}
	if got := vp.IncrementLevel("CLI", "Foundation"); got != "major" {
		t.Errorf("CLI/Foundation = %q, want major", got)
	}
	if got := vp.IncrementLevel("DATA", ""); got != "major" {
		t.Errorf("DATA = %q, want major", got)
	}
	if got := vp.IncrementLevel("UNKNOWN", ""); got != "patch" {
		t.Errorf("UNKNOWN = %q, want patch", got)
	}
}
