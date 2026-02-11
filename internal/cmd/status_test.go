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
	os.Chdir(projectRoot)
	defer os.Chdir(oldWd)

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

func TestStatusVerbosityLevels(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	os.Chdir(projectRoot)
	defer os.Chdir(oldWd)

	// Test -vv shows category breakdown
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
