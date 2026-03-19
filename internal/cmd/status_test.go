package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx-go/pkg/rtmx"
	"github.com/spf13/cobra"
)

func TestStatusRealCommand(t *testing.T) {
	// Find project root with .rtmx directory
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	// Run the real status command
	rootCmd := createStatusTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"status"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("status command failed: %v", err)
	}

	output := buf.String()

	// Verify output contains expected elements
	expectedPhrases := []string{
		"RTM Status Check",
		"Requirements:",
		"Phase Status",
		"complete",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Expected output to contain %q, got:\n%s", phrase, output)
		}
	}
}

// TestStatusPhaseNames verifies that phase names from config are displayed
// REQ-GO-049: Go CLI shall display phase names from config in status output
func TestStatusPhaseNames(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createStatusTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"status"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("status command failed: %v", err)
	}

	output := buf.String()

	// Verify phase names from config are shown
	expectedPhases := []string{
		"Phase 1 (Foundation)",
		"Phase 2 (Core Data Model)",
	}

	for _, phrase := range expectedPhases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Expected output to contain %q, got:\n%s", phrase, output)
		}
	}
}

// TestStatusCategoryListFormat verifies that status -v shows Python-style category list
// REQ-GO-050: Go CLI status -v shall match Python category list format
func TestStatusCategoryListFormat(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createStatusTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"status", "-v"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("status -v failed: %v", err)
	}

	output := buf.String()

	// Verify Python-style category list format
	expectedPhrases := []string{
		"Requirements by Category:",
		"complete",
		"partial",
		"missing",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Expected output to contain %q, got:\n%s", phrase, output)
		}
	}

	// Verify it does NOT contain progress bars (old format)
	if strings.Contains(output, "[██") || strings.Contains(output, "[░░") {
		// Progress bars in category section would indicate old format
		// Note: overall progress bar is OK, just not per-category
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "complete") && strings.Contains(line, "missing") {
				// This is a category line - should not have progress bar
				if strings.Contains(line, "[██") || strings.Contains(line, "[░░") {
					t.Errorf("Category line should not contain progress bar: %s", line)
				}
			}
		}
	}
}

func TestStatusVerbosityLevels(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	// Test -vv shows phase and category breakdown
	rootCmd := createStatusTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"status", "-vv"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("status -vv failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Phase and Category") {
		t.Errorf("Expected -vv to show Phase and Category breakdown, got:\n%s", output)
	}
}

// findProjectRootDir looks for the project root with .rtmx directory
func findProjectRootDir(start string) string {
	dir := start
	for i := 0; i < 10; i++ {
		rtmxDir := filepath.Join(dir, ".rtmx")
		if info, err := os.Stat(rtmxDir); err == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// TestStatusJSON verifies that --json produces valid JSON with correct structure.
// REQ-PAR-001: JSON output flag for status command
func TestStatusJSON(t *testing.T) {
	rtmx.Req(t, "REQ-PAR-001", rtmx.Scope("unit"), rtmx.Technique("nominal"))

	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createStatusTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"status", "--json"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("status --json failed: %v", err)
	}

	output := buf.String()

	// Must be valid JSON
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\nOutput:\n%s", err, output)
	}

	// Verify required top-level fields
	requiredFields := []string{"total", "complete", "partial", "missing", "completion_pct", "phases", "categories"}
	for _, field := range requiredFields {
		if _, ok := result[field]; !ok {
			t.Errorf("JSON output missing required field %q", field)
		}
	}

	// Verify phases is an array with correct structure
	phases, ok := result["phases"].([]interface{})
	if !ok {
		t.Fatalf("phases is not an array")
	}
	if len(phases) > 0 {
		phase := phases[0].(map[string]interface{})
		for _, field := range []string{"phase", "name", "total", "complete", "pct"} {
			if _, ok := phase[field]; !ok {
				t.Errorf("phase entry missing required field %q", field)
			}
		}
	}

	// Verify categories is an array with correct structure
	categories, ok := result["categories"].([]interface{})
	if !ok {
		t.Fatalf("categories is not an array")
	}
	if len(categories) > 0 {
		cat := categories[0].(map[string]interface{})
		for _, field := range []string{"name", "total", "complete", "pct"} {
			if _, ok := cat[field]; !ok {
				t.Errorf("category entry missing required field %q", field)
			}
		}
	}

	// Verify numeric values are reasonable
	total := result["total"].(float64)
	complete := result["complete"].(float64)
	if total <= 0 {
		t.Errorf("total should be positive, got %v", total)
	}
	if complete < 0 || complete > total {
		t.Errorf("complete should be between 0 and total, got %v", complete)
	}
}

// TestStatusJSONSuppressesNonJSON verifies that --json suppresses headers and progress bars.
// REQ-PAR-001: JSON output should contain no non-JSON text
func TestStatusJSONSuppressesNonJSON(t *testing.T) {
	rtmx.Req(t, "REQ-PAR-001", rtmx.Scope("unit"), rtmx.Technique("nominal"))

	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createStatusTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"status", "--json"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("status --json failed: %v", err)
	}

	output := strings.TrimSpace(buf.String())

	// Must not contain human-readable elements
	if strings.Contains(output, "RTM Status Check") {
		t.Error("JSON output should not contain 'RTM Status Check' header")
	}
	if strings.Contains(output, "Requirements:") {
		t.Error("JSON output should not contain 'Requirements:' text")
	}

	// Must start with { and end with }
	if !strings.HasPrefix(output, "{") || !strings.HasSuffix(output, "}") {
		t.Errorf("JSON output should be a JSON object, got:\n%s", output)
	}
}

// TestStatusFailUnder verifies that --fail-under exits with error when below threshold.
// REQ-PAR-002: Fail-under threshold for status command
func TestStatusFailUnder(t *testing.T) {
	rtmx.Req(t, "REQ-PAR-002", rtmx.Scope("unit"), rtmx.Technique("nominal"))

	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createStatusTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	// Use 100% threshold - should always fail since project is incomplete
	rootCmd.SetArgs([]string{"status", "--fail-under", "100"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when completion is below --fail-under threshold")
	}

	// Verify it's an ExitError with code 1
	exitErr, ok := err.(*ExitError)
	if !ok {
		t.Fatalf("expected *ExitError, got %T: %v", err, err)
	}
	if exitErr.Code != 1 {
		t.Errorf("expected exit code 1, got %d", exitErr.Code)
	}
}

// TestStatusFailUnderPassing verifies that --fail-under passes when above threshold.
// REQ-PAR-002: Fail-under threshold for status command (passing case)
func TestStatusFailUnderPassing(t *testing.T) {
	rtmx.Req(t, "REQ-PAR-002", rtmx.Scope("unit"), rtmx.Technique("nominal"))

	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createStatusTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	// Use 0% threshold - should always pass
	rootCmd.SetArgs([]string{"status", "--fail-under", "0"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("expected no error when completion is above --fail-under threshold, got: %v", err)
	}
}

// TestStatusJSONWithFailUnder verifies that --json and --fail-under work together.
// REQ-PAR-001, REQ-PAR-002: JSON output with fail-under threshold
func TestStatusJSONWithFailUnder(t *testing.T) {
	rtmx.Req(t, "REQ-PAR-001", rtmx.Scope("unit"), rtmx.Technique("nominal"))

	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createStatusTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"status", "--json", "--fail-under", "100"})

	err := rootCmd.Execute()
	// Should fail because of threshold
	if err == nil {
		t.Fatal("expected error when below threshold")
	}

	// Output should still be valid JSON
	output := strings.TrimSpace(buf.String())
	var result map[string]interface{}
	if jsonErr := json.Unmarshal([]byte(output), &result); jsonErr != nil {
		t.Fatalf("output should be valid JSON even when fail-under triggers: %v\nOutput:\n%s", jsonErr, output)
	}

	// JSON should include fail_under info
	if _, ok := result["fail_under"]; !ok {
		t.Error("JSON output should include fail_under field when --fail-under is used")
	}
	if _, ok := result["threshold_passed"]; !ok {
		t.Error("JSON output should include threshold_passed field when --fail-under is used")
	}
}

// createStatusTestCmd creates a root command with real status command for testing
func createStatusTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Create fresh status command with local flags
	var verbosity int
	var jsonOutput bool
	var failUnder float64
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show RTM completion status",
		RunE: func(cmd *cobra.Command, args []string) error {
			statusVerbosity = verbosity
			statusJSON = jsonOutput
			statusFailUnder = failUnder
			return runStatus(cmd, args)
		},
	}
	statusCmd.Flags().CountVarP(&verbosity, "verbose", "v", "increase verbosity")
	statusCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	statusCmd.Flags().Float64Var(&failUnder, "fail-under", 0, "fail if completion below threshold")
	root.AddCommand(statusCmd)

	return root
}
