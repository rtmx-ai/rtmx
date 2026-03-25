package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoadFromDir tests LoadFromDir with and without config files.
func TestLoadFromDir(t *testing.T) {
	// LoadFromDir with no config file should return defaults
	tmpDir := t.TempDir()
	cfg, err := LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadFromDir failed: %v", err)
	}
	if cfg.RTMX.Database != ".rtmx/database.csv" {
		t.Errorf("Default database = %q, want .rtmx/database.csv", cfg.RTMX.Database)
	}

	// LoadFromDir with config file should load it
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0755)
	configContent := "rtmx:\n  database: custom.csv\n  schema: phoenix\n"
	_ = os.WriteFile(filepath.Join(rtmxDir, "config.yaml"), []byte(configContent), 0644)

	cfg, err = LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("LoadFromDir with config failed: %v", err)
	}
	if cfg.RTMX.Database != "custom.csv" {
		t.Errorf("Database = %q, want custom.csv", cfg.RTMX.Database)
	}
}

// TestRequirementsPath tests RequirementsPath with relative and absolute paths.
func TestRequirementsPath(t *testing.T) {
	cfg := DefaultConfig()

	// Relative path
	tmpDir := t.TempDir()
	path := cfg.RequirementsPath(tmpDir)
	expected := filepath.Join(tmpDir, ".rtmx", "requirements")
	if path != expected {
		t.Errorf("RequirementsPath = %q, want %q", path, expected)
	}

	// Absolute path
	absPath := filepath.Join(tmpDir, "abs", "requirements")
	cfg.RTMX.RequirementsDir = absPath
	path = cfg.RequirementsPath(tmpDir)
	if path != absPath {
		t.Errorf("Absolute RequirementsPath = %q, want %q", path, absPath)
	}
}

// TestLoadInvalidYAML tests Load with invalid YAML content.
func TestLoadInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "bad.yaml")
	_ = os.WriteFile(path, []byte("{{{{invalid yaml"), 0644)

	_, err := Load(path)
	if err == nil {
		t.Error("Load with invalid YAML should fail")
	}
}

// TestLoadNonExistent tests Load with non-existent file.
func TestLoadNonExistent(t *testing.T) {
	_, err := Load("/nonexistent/config.yaml")
	if err == nil {
		t.Error("Load with non-existent file should fail")
	}
}

// TestFindConfigNotFound tests FindConfig when no config exists.
func TestFindConfigNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	_, err := FindConfig(tmpDir)
	if err == nil {
		t.Error("FindConfig should fail when no config exists")
	}
}

// TestFindConfigRtmxYaml tests FindConfig finding rtmx.yaml format.
func TestFindConfigRtmxYaml(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "rtmx.yaml")
	_ = os.WriteFile(configPath, []byte("rtmx:\n  schema: test\n"), 0644)

	found, err := FindConfig(tmpDir)
	if err != nil {
		t.Fatalf("FindConfig failed: %v", err)
	}
	if found != configPath {
		t.Errorf("FindConfig = %q, want %q", found, configPath)
	}
}

// TestSaveCreatesDirectory tests that Save creates the directory.
func TestSaveCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "deep", "nested", "config.yaml")

	cfg := DefaultConfig()
	err := cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Save should create the file")
	}

	// Load it back
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load after save failed: %v", err)
	}
	if loaded.RTMX.Schema != "core" {
		t.Errorf("Loaded schema = %q, want core", loaded.RTMX.Schema)
	}
}

// TestPhaseDescriptionAll tests PhaseDescription for all default phases.
func TestPhaseDescriptionAll(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		phase    int
		expected string
	}{
		{1, "Foundation"},
		{2, "Core Features"},
		{3, "Integration"},
		{99, "Phase 99"},
		{0, "Phase 0"},
	}

	for _, tt := range tests {
		got := cfg.PhaseDescription(tt.phase)
		if got != tt.expected {
			t.Errorf("PhaseDescription(%d) = %q, want %q", tt.phase, got, tt.expected)
		}
	}
}
