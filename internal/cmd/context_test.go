package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx-go/pkg/rtmx"
)

// createTestProject creates a temp directory with a minimal RTMX project.
func createTestProject(t *testing.T, csvRows [][]string) string {
	t.Helper()
	tmpDir := t.TempDir()

	// Create .rtmx directory
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write config.yaml
	configContent := "rtmx:\n  database: .rtmx/database.csv\n  project_name: test-project\n"
	if err := os.WriteFile(filepath.Join(rtmxDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Write database CSV
	dbPath := filepath.Join(rtmxDir, "database.csv")
	f, err := os.Create(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	header := []string{
		"req_id", "category", "subcategory", "requirement_text", "target_value",
		"test_module", "test_function", "validation_method", "status", "priority",
		"phase", "notes", "effort_weeks", "dependencies", "blocks",
		"assignee", "sprint", "started_date", "completed_date", "requirement_file",
		"external_id",
	}
	if err := w.Write(header); err != nil {
		t.Fatal(err)
	}
	for _, row := range csvRows {
		if err := w.Write(row); err != nil {
			t.Fatal(err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		t.Fatal(err)
	}

	return tmpDir
}

// makeRow creates a CSV row with specified values.
func makeRow(reqID, category, text, status, priority string, phase int, effort float64, deps string) []string {
	phaseStr := ""
	if phase > 0 {
		phaseStr = fmt.Sprintf("%d", phase)
	}
	effortStr := ""
	if effort > 0 {
		effortStr = fmt.Sprintf("%.1f", effort)
	}
	return []string{
		reqID, category, "", text, "",
		"", "", "", status, priority,
		phaseStr, "", effortStr, deps, "",
		"", "", "", "", "",
		"",
	}
}

func TestContextCommandHelp(t *testing.T) {
	rtmx.Req(t, "REQ-AGENT-002",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	cmd := newTestRootCmd()
	output, err := executeCommand(cmd, "context", "--help")

	if err != nil {
		t.Fatalf("context --help failed: %v", err)
	}

	expectedPhrases := []string{
		"context",
		"--format",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Expected help to contain %q, got: %s", phrase, output)
		}
	}
}

func TestContextCommandDefaultFormat(t *testing.T) {
	rtmx.Req(t, "REQ-AGENT-002",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	rows := [][]string{
		makeRow("REQ-A-001", "CORE", "First requirement", "COMPLETE", "HIGH", 1, 1.0, ""),
		makeRow("REQ-A-002", "CORE", "Second requirement", "MISSING", "P0", 1, 1.0, ""),
		makeRow("REQ-A-003", "ADAPT", "Third requirement", "MISSING", "LOW", 2, 1.0, "REQ-A-002"),
		makeRow("REQ-A-004", "CLI", "Fourth requirement", "PARTIAL", "MEDIUM", 2, 1.0, ""),
	}
	tmpDir := createTestProject(t, rows)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	output, err := executeCommand(cmd, "context")

	if err != nil {
		t.Fatalf("context command failed: %v", err)
	}

	// Should contain completion percentage
	if !strings.Contains(output, "%") {
		t.Errorf("Expected output to contain completion %%, got: %s", output)
	}

	// Should contain RTM context header
	if !strings.Contains(output, "RTM") {
		t.Errorf("Expected output to contain 'RTM', got: %s", output)
	}
}

func TestContextCommandClaudeFormat(t *testing.T) {
	rtmx.Req(t, "REQ-AGENT-002",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	rows := [][]string{
		makeRow("REQ-A-001", "CORE", "First requirement", "COMPLETE", "HIGH", 1, 1.0, ""),
		makeRow("REQ-A-002", "CORE", "Second requirement", "MISSING", "P0", 1, 1.0, ""),
		makeRow("REQ-A-003", "ADAPT", "Third requirement", "MISSING", "LOW", 2, 1.0, "REQ-A-002"),
		makeRow("REQ-A-004", "CLI", "Fourth requirement", "MISSING", "MEDIUM", 2, 1.0, ""),
	}
	tmpDir := createTestProject(t, rows)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	output, err := executeCommand(cmd, "context", "--format", "claude")

	if err != nil {
		t.Fatalf("context --format claude failed: %v", err)
	}

	// Claude format should contain completion %
	if !strings.Contains(output, "%") {
		t.Errorf("Expected output to contain completion %%, got: %s", output)
	}

	// Should contain blockers section
	if !strings.Contains(strings.ToLower(output), "blocker") {
		t.Errorf("Expected output to contain blockers section, got: %s", output)
	}
}

func TestContextCommandPlainFormat(t *testing.T) {
	rtmx.Req(t, "REQ-AGENT-002",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	rows := [][]string{
		makeRow("REQ-A-001", "CORE", "First requirement", "COMPLETE", "HIGH", 1, 1.0, ""),
		makeRow("REQ-A-002", "CORE", "Second requirement", "MISSING", "P0", 1, 1.0, ""),
	}
	tmpDir := createTestProject(t, rows)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	output, err := executeCommand(cmd, "context", "--format", "plain")

	if err != nil {
		t.Fatalf("context --format plain failed: %v", err)
	}

	if !strings.Contains(output, "%") {
		t.Errorf("Expected output to contain completion %%, got: %s", output)
	}
}

func TestContextCommandTokenEfficiency(t *testing.T) {
	rtmx.Req(t, "REQ-AGENT-002",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	// Create a project with many requirements
	var rows [][]string
	for i := 0; i < 50; i++ {
		status := "MISSING"
		if i < 20 {
			status = "COMPLETE"
		}
		priority := "MEDIUM"
		if i%5 == 0 {
			priority = "P0"
		}
		reqID := fmt.Sprintf("REQ-T-%03d", i)
		rows = append(rows, makeRow(
			reqID, "CORE", "Requirement text for "+reqID, status, priority, 1, 1.0, "",
		))
	}
	tmpDir := createTestProject(t, rows)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	output, err := executeCommand(cmd, "context", "--format", "claude")

	if err != nil {
		t.Fatalf("context command failed: %v", err)
	}

	// Token estimation: ~4 chars per token on average.
	// 500 tokens * 4 chars = 2000 chars. Be generous and allow up to 3000.
	if len(output) > 3000 {
		t.Errorf("Context output too long for token efficiency: %d chars (target < 2000 for ~500 tokens), output:\n%s", len(output), output)
	}
}

func TestContextBlockersAndQuickWins(t *testing.T) {
	rtmx.Req(t, "REQ-AGENT-002",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	rows := [][]string{
		// Incomplete requirement that blocks others
		makeRow("REQ-B-001", "CORE", "Foundation", "MISSING", "P0", 1, 1.0, ""),
		// Blocked by REQ-B-001
		makeRow("REQ-B-002", "CORE", "Feature A", "MISSING", "HIGH", 1, 1.0, "REQ-B-001"),
		// Quick win: high priority, low effort, not blocked
		makeRow("REQ-B-003", "CLI", "Quick CLI fix", "MISSING", "P0", 1, 0.5, ""),
		// Low priority, high effort - less attractive quick win
		makeRow("REQ-B-004", "ADAPT", "Big refactor", "MISSING", "LOW", 2, 3.0, ""),
	}
	tmpDir := createTestProject(t, rows)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	output, err := executeCommand(cmd, "context", "--format", "claude")

	if err != nil {
		t.Fatalf("context command failed: %v", err)
	}

	lowerOutput := strings.ToLower(output)
	if !strings.Contains(lowerOutput, "blocker") {
		t.Errorf("Expected output to mention blockers, got: %s", output)
	}

	if !strings.Contains(lowerOutput, "quick win") {
		t.Errorf("Expected output to mention quick wins, got: %s", output)
	}
}

func TestContextEmptyDatabase(t *testing.T) {
	rtmx.Req(t, "REQ-AGENT-002",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	tmpDir := createTestProject(t, [][]string{})

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	output, err := executeCommand(cmd, "context")

	if err != nil {
		t.Fatalf("context command failed: %v", err)
	}

	if !strings.Contains(output, "0") || !strings.Contains(output, "%") {
		t.Errorf("Expected output to show 0%% for empty database, got: %s", output)
	}
}

func TestContextInvalidFormat(t *testing.T) {
	rtmx.Req(t, "REQ-AGENT-002",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	rows := [][]string{
		makeRow("REQ-C-001", "CORE", "Test req", "COMPLETE", "HIGH", 1, 1.0, ""),
	}
	tmpDir := createTestProject(t, rows)

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	cmd := newTestRootCmd()
	_, err := executeCommand(cmd, "context", "--format", "invalid")

	if err == nil {
		t.Error("Expected error for invalid format, got nil")
	}
}
