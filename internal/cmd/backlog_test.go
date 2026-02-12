package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestBacklogRealCommand(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	os.Chdir(projectRoot)
	defer os.Chdir(oldWd)

	rootCmd := createBacklogTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"backlog"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("backlog command failed: %v", err)
	}

	output := buf.String()
	expectedPhrases := []string{
		"Prioritized Backlog",
		"Total Requirements:",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Expected output to contain %q, got:\n%s", phrase, output)
		}
	}
}

func TestBacklogPhaseFilter(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	os.Chdir(projectRoot)
	defer os.Chdir(oldWd)

	rootCmd := createBacklogTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"backlog", "--phase", "1"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("backlog --phase 1 failed: %v", err)
	}

	output := buf.String()
	// Phase 1 is complete, so backlog should be empty or show no items
	if !strings.Contains(output, "Prioritized Backlog") {
		t.Errorf("Expected backlog header, got:\n%s", output)
	}
}

func TestBacklogViewModes(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	os.Chdir(projectRoot)
	defer os.Chdir(oldWd)

	views := []string{"all", "critical", "quick-wins", "blockers", "list"}

	for _, view := range views {
		t.Run(view, func(t *testing.T) {
			rootCmd := createBacklogTestCmd()
			buf := new(bytes.Buffer)
			rootCmd.SetOut(buf)
			rootCmd.SetArgs([]string{"backlog", "--view", view})

			err := rootCmd.Execute()
			if err != nil {
				t.Fatalf("backlog --view %s failed: %v", view, err)
			}

			output := buf.String()
			if !strings.Contains(output, "Prioritized Backlog") {
				t.Errorf("Expected backlog output for view %s, got:\n%s", view, output)
			}
		})
	}
}

// TestBacklogTableFormat verifies that backlog uses ASCII table format
// REQ-GO-048: Go CLI shall use ASCII tables matching Python tabulate output
func TestBacklogTableFormat(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	os.Chdir(projectRoot)
	defer os.Chdir(oldWd)

	rootCmd := createBacklogTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"backlog"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("backlog command failed: %v", err)
	}

	output := buf.String()

	// Verify ASCII table format markers
	expectedTableElements := []string{
		"+---+",      // Column separator
		"|",          // Row borders
		"+===+",      // Header separator (with = instead of -)
		"Status",     // Column header
		"Requirement", // Column header
		"Description", // Column header
	}

	for _, element := range expectedTableElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected table format element %q, got:\n%s", element, output)
		}
	}

	// Verify sections exist
	expectedSections := []string{
		"Prioritized Backlog",
		"Total Requirements:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Expected section %q, got:\n%s", section, output)
		}
	}
}

// createBacklogTestCmd creates a root command with real backlog command for testing
func createBacklogTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	var view string
	var phase int
	var category string
	var limit int

	backlogCmd := &cobra.Command{
		Use:   "backlog",
		Short: "Show prioritized backlog",
		RunE: func(cmd *cobra.Command, args []string) error {
			backlogView = view
			backlogPhase = phase
			backlogCategory = category
			backlogLimit = limit
			return runBacklog(cmd, args)
		},
	}
	backlogCmd.Flags().StringVar(&view, "view", "all", "view mode")
	backlogCmd.Flags().IntVar(&phase, "phase", 0, "filter by phase")
	backlogCmd.Flags().StringVar(&category, "category", "", "filter by category")
	backlogCmd.Flags().IntVarP(&limit, "limit", "n", 0, "limit results")
	root.AddCommand(backlogCmd)

	return root
}
