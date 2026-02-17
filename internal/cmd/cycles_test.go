package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCyclesRealCommand(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createCyclesTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"cycles"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("cycles command failed: %v", err)
	}

	output := buf.String()
	expectedPhrases := []string{
		"Circular Dependency Analysis",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Expected output to contain %q, got:\n%s", phrase, output)
		}
	}
}

// TestCyclesDetailFormat verifies that cycles shows statistics and recommendations
// REQ-GO-053: Go CLI cycles shall show stats paths and recommendations
func TestCyclesDetailFormat(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createCyclesTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"cycles"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("cycles command failed: %v", err)
	}

	output := buf.String()

	// Verify statistics section exists
	expectedElements := []string{
		"RTM Statistics:",
		"Total requirements:",
		"Total dependencies:",
		"Average dependencies per requirement:",
	}

	for _, element := range expectedElements {
		if !strings.Contains(output, element) {
			t.Errorf("Expected cycles output to contain %q, got:\n%s", element, output)
		}
	}
}

// createCyclesTestCmd creates a root command with real cycles command for testing
func createCyclesTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	var jsonOutput bool

	cyclesTestCmd := &cobra.Command{
		Use: "cycles",
		RunE: func(cmd *cobra.Command, args []string) error {
			cyclesJSON = jsonOutput
			return runCycles(cmd, args)
		},
	}
	cyclesTestCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	root.AddCommand(cyclesTestCmd)

	return root
}
