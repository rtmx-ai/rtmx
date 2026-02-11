package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestHealthRealCommand(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	os.Chdir(projectRoot)
	defer os.Chdir(oldWd)

	rootCmd := createHealthTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"health"})

	err := rootCmd.Execute()
	// Health may return ExitError for warnings - that's OK
	if err != nil {
		var exitErr *ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("health command failed unexpectedly: %v", err)
		}
	}

	output := buf.String()
	expectedPhrases := []string{
		"Health Check",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Expected output to contain %q, got:\n%s", phrase, output)
		}
	}
}

func TestHealthJSONOutput(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	os.Chdir(projectRoot)
	defer os.Chdir(oldWd)

	rootCmd := createHealthTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"health", "--json"})

	err := rootCmd.Execute()
	// Health may return ExitError for warnings - that's OK
	if err != nil {
		var exitErr *ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("health --json failed unexpectedly: %v", err)
		}
	}

	output := buf.String()

	// Verify it's valid JSON
	var result HealthResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &result); err != nil {
		t.Errorf("Expected valid JSON output, got parse error: %v\nOutput: %s", err, output)
	}

	// Verify required fields
	if result.Status == "" {
		t.Error("Expected status field in JSON output")
	}
}

// createHealthTestCmd creates a root command with real health command for testing
func createHealthTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	var jsonOutput bool

	healthCmd := &cobra.Command{
		Use:   "health",
		Short: "Run health check",
		RunE: func(cmd *cobra.Command, args []string) error {
			healthJSON = jsonOutput
			return runHealth(cmd, args)
		},
	}
	healthCmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	root.AddCommand(healthCmd)

	return root
}
