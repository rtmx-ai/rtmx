package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitCommand(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-init-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(oldWd) }()

	// Create fresh command
	rootCmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"init"})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	output := buf.String()

	// Verify output
	if !strings.Contains(output, "RTM initialized successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify created files
	expectedFiles := []string{
		".rtmx/database.csv",
		".rtmx/config.yaml",
		".rtmx/.gitignore",
		".rtmx/requirements/EXAMPLE/REQ-EX-001.md",
	}

	for _, f := range expectedFiles {
		path := filepath.Join(tmpDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file not created: %s", f)
		}
	}

	// Verify directory
	expectedDirs := []string{
		".rtmx",
		".rtmx/requirements",
		".rtmx/cache",
	}

	for _, d := range expectedDirs {
		path := filepath.Join(tmpDir, d)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			t.Errorf("Expected directory not created: %s", d)
		} else if !info.IsDir() {
			t.Errorf("Expected %s to be a directory", d)
		}
	}
}

func TestInitLegacyCommand(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-init-legacy-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(oldWd) }()

	// Create fresh command
	rootCmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"init", "--legacy"})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("init --legacy command failed: %v", err)
	}

	output := buf.String()

	// Verify output
	if !strings.Contains(output, "RTM initialized successfully") {
		t.Errorf("Expected success message, got: %s", output)
	}

	// Verify created files
	expectedFiles := []string{
		"docs/rtm_database.csv",
		"rtmx.yaml",
		"docs/requirements/EXAMPLE/REQ-EX-001.md",
	}

	for _, f := range expectedFiles {
		path := filepath.Join(tmpDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file not created: %s", f)
		}
	}
}

func TestInitDatabaseContent(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-init-content-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(oldWd) }()

	// Create fresh command
	rootCmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"init"})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// Read and verify database content
	dbPath := filepath.Join(tmpDir, ".rtmx", "database.csv")
	content, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("Failed to read database: %v", err)
	}

	dbContent := string(content)
	if !strings.Contains(dbContent, "REQ-EX-001") {
		t.Error("Database should contain sample requirement REQ-EX-001")
	}
	if !strings.Contains(dbContent, "req_id,category") {
		t.Error("Database should contain CSV header")
	}
}

func TestInitConfigContent(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "rtmx-init-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(oldWd) }()

	// Create fresh command
	rootCmd := newTestRootCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"init"})

	err = rootCmd.Execute()
	if err != nil {
		t.Fatalf("init command failed: %v", err)
	}

	// Read and verify config content
	configPath := filepath.Join(tmpDir, ".rtmx", "config.yaml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	configContent := string(content)
	if !strings.Contains(configContent, "database: .rtmx/database.csv") {
		t.Error("Config should reference .rtmx/database.csv")
	}
	if !strings.Contains(configContent, "rtmx:") {
		t.Error("Config should have rtmx section")
	}
}
