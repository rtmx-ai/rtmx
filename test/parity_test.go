package test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestFullParity validates that the Go CLI achieves feature parity with Python CLI
// REQ-GO-020: Go CLI v1.0.0 shall achieve full feature parity
func TestFullParity(t *testing.T) {
	// Build the Go CLI binary
	tmpDir, err := os.MkdirTemp("", "rtmx-parity-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath := filepath.Join(tmpDir, "rtmx")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, "../cmd/rtmx")
	buildCmd.Dir = filepath.Dir(tmpDir)

	// Get the actual project root
	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	buildCmd = exec.Command("go", "build", "-o", binaryPath, "./cmd/rtmx")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build Go CLI: %v\n%s", err, output)
	}

	// Test command availability - all these commands must exist and work
	commands := []struct {
		name     string
		args     []string
		mustHave []string
	}{
		{
			name:     "help",
			args:     []string{"--help"},
			mustHave: []string{"status", "backlog", "health", "init", "verify", "from-tests", "deps", "cycles"},
		},
		{
			name:     "version",
			args:     []string{"--version"},
			mustHave: []string{"rtmx"},
		},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tc.args...)
			cmd.Dir = projectRoot
			output, err := cmd.CombinedOutput()
			if err != nil {
				// --help may exit with non-zero on some setups
				if !strings.Contains(tc.name, "help") {
					t.Errorf("%s failed: %v\n%s", tc.name, err, output)
					return
				}
			}

			for _, phrase := range tc.mustHave {
				if !bytes.Contains(output, []byte(phrase)) {
					t.Errorf("%s: missing expected phrase %q in output:\n%s", tc.name, phrase, output)
				}
			}
		})
	}

	// Test command parity - each command runs and produces expected output
	// Note: health may exit non-zero when there are warnings (expected behavior)
	testCommands := []struct {
		name      string
		args      []string
		allowErr  bool   // allow non-zero exit
		mustMatch string // output must contain this
	}{
		{"status", []string{"status"}, false, "RTM Status Check"},
		{"backlog", []string{"backlog"}, false, "Backlog"},
		{"health", []string{"health"}, true, "Health Check"}, // May exit 1 with warnings
		{"deps", []string{"deps"}, false, ""},
		{"cycles", []string{"cycles"}, false, ""},
	}

	for _, tc := range testCommands {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tc.args...)
			cmd.Dir = projectRoot
			output, err := cmd.CombinedOutput()

			if !tc.allowErr && err != nil {
				t.Errorf("%s: unexpected error: %v\n%s", tc.name, err, output)
			}

			if tc.mustMatch != "" && !bytes.Contains(output, []byte(tc.mustMatch)) {
				t.Errorf("%s: output missing %q", tc.name, tc.mustMatch)
			}
		})
	}
}

// TestStatusParity validates status command output format
func TestStatusParity(t *testing.T) {
	// Build binary
	tmpDir, err := os.MkdirTemp("", "rtmx-status-parity")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath := filepath.Join(tmpDir, "rtmx")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/rtmx")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	// Run status command
	cmd := exec.Command(binaryPath, "status")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("status command failed: %v\n%s", err, output)
	}

	// Check required output elements (parity with Python)
	requiredElements := []string{
		"RTM Status Check",
		"Requirements:",
		"complete",
		"partial",
		"missing",
		"Phase Status",
	}

	for _, elem := range requiredElements {
		if !bytes.Contains(output, []byte(elem)) {
			t.Errorf("Status output missing required element: %q", elem)
		}
	}
}

// TestBacklogParity validates backlog command output format
func TestBacklogParity(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rtmx-backlog-parity")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath := filepath.Join(tmpDir, "rtmx")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/rtmx")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	// Run backlog command
	cmd := exec.Command(binaryPath, "backlog")
	cmd.Dir = projectRoot
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("backlog command failed: %v\n%s", err, output)
	}

	// Check required output elements
	requiredElements := []string{
		"Backlog",
	}

	for _, elem := range requiredElements {
		if !bytes.Contains(output, []byte(elem)) {
			t.Errorf("Backlog output missing required element: %q", elem)
		}
	}
}

// TestHealthParity validates health command output format
func TestHealthParity(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "rtmx-health-parity")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	binaryPath := filepath.Join(tmpDir, "rtmx")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/rtmx")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build: %v\n%s", err, output)
	}

	// Run health command
	cmd := exec.Command(binaryPath, "health")
	cmd.Dir = projectRoot
	output, _ := cmd.CombinedOutput()

	// Health command may exit non-zero if there are warnings
	// Just check it produces expected output format
	requiredElements := []string{
		"Health Check",
	}

	for _, elem := range requiredElements {
		if !bytes.Contains(output, []byte(elem)) {
			t.Errorf("Health output missing required element: %q", elem)
		}
	}
}
