package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestDepsRealCommand(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createDepsTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"deps"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("deps command failed: %v", err)
	}

	output := buf.String()
	expectedPhrases := []string{
		"Dependencies",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Expected output to contain %q, got:\n%s", phrase, output)
		}
	}
}

// TestDepsTableFormat verifies that deps shows a table with all requirements
// REQ-GO-052: Go CLI deps shall show full requirements table like Python
func TestDepsTableFormat(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createDepsTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"deps"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("deps command failed: %v", err)
	}

	output := buf.String()

	// Verify table format with column headers
	expectedElements := []string{
		"ID",          // Column header
		"Deps",        // Column header
		"Blocks",      // Column header
		"Description", // Column header
		"---",         // Row separator
		"REQ-GO-",     // Should show requirement IDs
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected deps output to contain %q, got:\n%s", element, output)
		}
	}
}

func TestDepsWorkable(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createDepsTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"deps", "--workable"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("deps --workable failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Workable") {
		t.Errorf("Expected workable requirements output, got:\n%s", output)
	}
}

// createDepsTestCmd creates a root command with real deps command for testing
func createDepsTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	var reverse, all, workable bool

	depsTestCmd := &cobra.Command{
		Use:  "deps [req_id]",
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			depsReverse = reverse
			depsAll = all
			depsWorkable = workable
			return runDeps(cmd, args)
		},
	}
	depsTestCmd.Flags().BoolVarP(&reverse, "reverse", "r", false, "show dependents")
	depsTestCmd.Flags().BoolVarP(&all, "all", "a", false, "show transitive")
	depsTestCmd.Flags().BoolVarP(&workable, "workable", "w", false, "show workable")
	root.AddCommand(depsTestCmd)

	return root
}
