package cmd

import (
	"bytes"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// executeCommand executes a cobra command and returns the output and error.
func executeCommand(root *cobra.Command, args ...string) (string, error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	err := root.Execute()
	return buf.String(), err
}

// newTestStatusCmd creates a fresh status command for testing.
func newTestStatusCmd() *cobra.Command {
	var verbosity int

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show RTM completion status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Status command not yet implemented")
			cmd.Printf("Verbosity level: %d\n", verbosity)
			return nil
		},
	}
	cmd.Flags().CountVarP(&verbosity, "verbose", "v", "increase verbosity")
	return cmd
}

// newTestBacklogCmd creates a fresh backlog command for testing.
func newTestBacklogCmd() *cobra.Command {
	var (
		view     string
		phase    int
		category string
		limit    int
	)

	cmd := &cobra.Command{
		Use:   "backlog",
		Short: "Show prioritized backlog",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Println("Backlog command not yet implemented")
			cmd.Printf("View: %s\n", view)
			if phase > 0 {
				cmd.Printf("Phase filter: %d\n", phase)
			}
			if category != "" {
				cmd.Printf("Category filter: %s\n", category)
			}
			if limit > 0 {
				cmd.Printf("Limit: %d\n", limit)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&view, "view", "all", "view mode")
	cmd.Flags().IntVar(&phase, "phase", 0, "filter by phase")
	cmd.Flags().StringVar(&category, "category", "", "filter by category")
	cmd.Flags().IntVarP(&limit, "limit", "n", 0, "limit results")
	return cmd
}

// newTestHealthCmd creates a fresh health command for testing.
func newTestHealthCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "health",
		Short: "Run health check",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				cmd.Println(`{"status": "not_implemented", "errors": [], "warnings": []}`)
			} else {
				cmd.Println("Health command not yet implemented")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	return cmd
}

// newTestVersionCmd creates a fresh version command for testing.
func newTestVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.Printf("rtmx version %s\n", Version)
			cmd.Printf("  commit:  %s\n", Commit)
			cmd.Printf("  built:   %s\n", Date)
			cmd.Printf("  go:      %s\n", runtime.Version())
			cmd.Printf("  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			return nil
		},
	}
}

// newTestInitCmd creates a fresh init command for testing with real behavior.
func newTestInitCmd() *cobra.Command {
	var force bool
	var legacy bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize RTM structure",
		RunE: func(cmd *cobra.Command, args []string) error {
			initForce = force
			initLegacy = legacy
			return runInit(cmd, args)
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "overwrite existing files")
	cmd.Flags().BoolVar(&legacy, "legacy", false, "use legacy docs/ directory structure")
	return cmd
}

// newTestRootCmd creates a fresh root command for testing.
// This avoids state pollution between tests.
func newTestRootCmd() *cobra.Command {
	var cfgFileTest string
	var noColorTest bool

	cmd := &cobra.Command{
		Use:   "rtmx",
		Short: "Requirements Traceability Matrix toolkit",
		Long: `RTMX is a CLI tool for managing requirements traceability in GenAI-driven development.

It provides commands to track requirements, run verification tests, manage dependencies,
and synchronize with external tools like GitHub and Jira.

Documentation: https://rtmx.ai/docs
Source: https://github.com/rtmx-ai/rtmx-go`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&cfgFileTest, "config", "", "config file")
	cmd.PersistentFlags().BoolVar(&noColorTest, "no-color", false, "disable colored output")

	cmd.AddCommand(newTestVersionCmd())
	cmd.AddCommand(newTestStatusCmd())
	cmd.AddCommand(newTestBacklogCmd())
	cmd.AddCommand(newTestHealthCmd())
	cmd.AddCommand(newTestInitCmd())

	return cmd
}

func TestRootCommandHelp(t *testing.T) {
	cmd := newTestRootCmd()
	output, err := executeCommand(cmd, "--help")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPhrases := []string{
		"rtmx",
		"RTMX",
		"--config",
		"--no-color",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("expected output to contain %q, got:\n%s", phrase, output)
		}
	}
}

func TestVersionCommand(t *testing.T) {
	cmd := newTestRootCmd()
	output, err := executeCommand(cmd, "version")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPhrases := []string{
		"rtmx version",
		"commit:",
		"built:",
		"go:",
		"os/arch:",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("expected output to contain %q, got:\n%s", phrase, output)
		}
	}
}

func TestStatusCommand(t *testing.T) {
	cmd := newTestRootCmd()
	output, err := executeCommand(cmd, "status")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Status") || !strings.Contains(output, "not yet implemented") {
		t.Errorf("expected placeholder output, got: %s", output)
	}
}

func TestStatusVerbosity(t *testing.T) {
	tests := []struct {
		args          []string
		wantVerbosity string
	}{
		{[]string{"status"}, "Verbosity level: 0"},
		{[]string{"status", "-v"}, "Verbosity level: 1"},
		{[]string{"status", "-vv"}, "Verbosity level: 2"},
		{[]string{"status", "-vvv"}, "Verbosity level: 3"},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			cmd := newTestRootCmd()
			output, err := executeCommand(cmd, tt.args...)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !strings.Contains(output, tt.wantVerbosity) {
				t.Errorf("expected output to contain %q, got:\n%s", tt.wantVerbosity, output)
			}
		})
	}
}

func TestBacklogCommand(t *testing.T) {
	cmd := newTestRootCmd()
	output, err := executeCommand(cmd, "backlog")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Backlog") || !strings.Contains(output, "not yet implemented") {
		t.Errorf("expected placeholder output, got: %s", output)
	}
}

func TestBacklogFlags(t *testing.T) {
	cmd := newTestRootCmd()
	output, err := executeCommand(cmd, "backlog", "--phase", "5", "--category", "CLI")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Phase filter: 5") {
		t.Errorf("expected phase filter in output, got: %s", output)
	}

	if !strings.Contains(output, "Category filter: CLI") {
		t.Errorf("expected category filter in output, got: %s", output)
	}
}

func TestHealthCommand(t *testing.T) {
	cmd := newTestRootCmd()
	output, err := executeCommand(cmd, "health")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, "Health") || !strings.Contains(output, "not yet implemented") {
		t.Errorf("expected placeholder output, got: %s", output)
	}
}

func TestHealthJSON(t *testing.T) {
	cmd := newTestRootCmd()
	output, err := executeCommand(cmd, "health", "--json")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(output, `"status"`) || !strings.Contains(output, `"errors"`) {
		t.Errorf("expected JSON output, got: %s", output)
	}
}
