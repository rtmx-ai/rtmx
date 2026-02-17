package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

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

// createStatusTestCmd creates a root command with real status command for testing
func createStatusTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Create fresh status command with local flags
	var verbosity int
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show RTM completion status",
		RunE: func(cmd *cobra.Command, args []string) error {
			statusVerbosity = verbosity
			return runStatus(cmd, args)
		},
	}
	statusCmd.Flags().CountVarP(&verbosity, "verbose", "v", "increase verbosity")
	root.AddCommand(statusCmd)

	return root
}
