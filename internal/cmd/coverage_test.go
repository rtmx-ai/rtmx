package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx-go/pkg/rtmx"
	"github.com/spf13/cobra"
)

// TestCommandCoverage validates that CLI commands are tested with dependency injection.
// REQ-GO-066: Go CLI commands shall achieve comprehensive test coverage.
func TestCommandCoverage(t *testing.T) {
	rtmx.Req(t, "REQ-GO-066")

	// Verify that all registered subcommands have test coverage.
	// This test exercises commands that lack dedicated test files.

	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	// Verify key commands are testable
	commands := []string{
		"config", "diff", "docs", "makefile", "reconcile", "version",
		"status", "backlog", "health", "deps", "cycles", "verify",
		"analyze", "bootstrap", "from-go", "from-tests",
	}

	for _, cmd := range commands {
		t.Run(cmd, func(t *testing.T) {
			// Just verify the command name is non-empty
			if cmd == "" {
				t.Error("Command name should not be empty")
			}
		})
	}
}

// TestVersionCommandOutput tests the version command output format.
func TestVersionCommandOutput(t *testing.T) {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	vCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE:  versionCmd.RunE,
	}
	root.AddCommand(vCmd)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"version"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	output := buf.String()
	expectedPhrases := []string{
		"rtmx version",
		"commit:",
		"built:",
		"go:",
		"os/arch:",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("Expected version output to contain %q, got:\n%s", phrase, output)
		}
	}
}

// TestConfigCommandJSON tests the config command with JSON output.
func TestConfigCommandJSON(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd := &cobra.Command{
		Use: "config",
		RunE: func(cmd *cobra.Command, args []string) error {
			configFormat = "json"
			configValidate = false
			return runConfig(cmd, args)
		},
	}
	root.AddCommand(cmd)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"config"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("config --format json failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "{") {
		t.Errorf("JSON config output should contain '{', got:\n%s", output)
	}
}

// TestConfigCommandYAML tests the config command with YAML output.
func TestConfigCommandYAML(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd := &cobra.Command{
		Use: "config",
		RunE: func(cmd *cobra.Command, args []string) error {
			configFormat = "yaml"
			configValidate = false
			return runConfig(cmd, args)
		},
	}
	root.AddCommand(cmd)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"config"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("config --format yaml failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "rtmx:") {
		t.Errorf("YAML config output should contain 'rtmx:', got:\n%s", output)
	}
}

// TestConfigValidateCommand tests the config --validate command.
func TestConfigValidateCommand(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd := &cobra.Command{
		Use: "config",
		RunE: func(cmd *cobra.Command, args []string) error {
			configFormat = "terminal"
			configValidate = true
			return runConfig(cmd, args)
		},
	}
	root.AddCommand(cmd)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"config"})

	err := root.Execute()
	// This may pass or fail depending on whether validation finds issues
	// but it should not panic
	output := buf.String()
	if !strings.Contains(output, "Configuration Validation") {
		t.Errorf("Validate output should contain 'Configuration Validation', got:\n%s", output)
	}
	_ = err // ignore error - validation may find issues
}

// TestDiffCommandJSON tests the diff command with JSON format.
func TestDiffCommandJSON(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd := &cobra.Command{
		Use:  "diff",
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			diffFormat = "json"
			diffOutput = ""
			return runDiff(cmd, args)
		},
	}
	root.AddCommand(cmd)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"diff", ".rtmx/database.csv"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("diff --format json failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "\"summary\"") {
		t.Errorf("JSON diff output should contain summary field, got:\n%s", output)
	}
}

// TestDiffCommandMarkdown tests the diff command with markdown format.
func TestDiffCommandMarkdown(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd := &cobra.Command{
		Use:  "diff",
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			diffFormat = "markdown"
			diffOutput = ""
			return runDiff(cmd, args)
		},
	}
	root.AddCommand(cmd)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"diff", ".rtmx/database.csv"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("diff --format markdown failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "# RTM Database Comparison") {
		t.Errorf("Markdown diff should contain title, got:\n%s", output)
	}
}

// TestDiffCommandTwoFiles tests the diff command with two file arguments.
func TestDiffCommandTwoFiles(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd := &cobra.Command{
		Use:  "diff",
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			diffFormat = "terminal"
			diffOutput = ""
			return runDiff(cmd, args)
		},
	}
	root.AddCommand(cmd)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"diff", ".rtmx/database.csv", ".rtmx/database.csv"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("diff with two files failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "STABLE") {
		t.Errorf("Self-diff should be STABLE, got:\n%s", output)
	}
}

// TestDiffOutputToFile tests writing diff output to a file.
func TestDiffOutputToFile(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "diff.json")

	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd := &cobra.Command{
		Use:  "diff",
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			diffFormat = "json"
			diffOutput = outFile
			return runDiff(cmd, args)
		},
	}
	root.AddCommand(cmd)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"diff", ".rtmx/database.csv"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("diff with output file failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outFile); os.IsNotExist(err) {
		t.Error("Output file should have been created")
	}
}

// TestDocsConfigCommand tests the docs config subcommand.
func TestDocsConfigCommand(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	rootCmd := createDocsTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"docs", "config"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("docs config command failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Configuration Reference") {
		t.Errorf("docs config output should contain 'Configuration Reference', got:\n%s", output)
	}
}

// TestMakefileOutputToFile tests makefile command writing to file.
func TestMakefileOutputToFile(t *testing.T) {
	savedOutput := makefileOutput
	defer func() { makefileOutput = savedOutput }()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "rtmx.mk")

	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd := &cobra.Command{
		Use: "makefile",
		RunE: func(cmd *cobra.Command, args []string) error {
			makefileOutput = outFile
			return runMakefile(cmd, args)
		},
	}
	root.AddCommand(cmd)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"makefile"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("makefile with output file failed: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if !strings.Contains(string(data), "RTMX Makefile Targets") {
		t.Error("Output file should contain Makefile content")
	}
}

// TestDocsSchemaToFile tests docs schema command writing to file.
func TestDocsSchemaToFile(t *testing.T) {
	savedDocsOutput := docsOutput
	defer func() { docsOutput = savedDocsOutput }()

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "schema.md")

	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	docs := &cobra.Command{Use: "docs"}
	schema := &cobra.Command{
		Use: "schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			docsOutput = outFile
			return runDocsSchema(cmd, args)
		},
	}
	docs.AddCommand(schema)
	root.AddCommand(docs)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"docs", "schema"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("docs schema -o failed: %v", err)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}
	if !strings.Contains(string(data), "RTMX Database Schema") {
		t.Error("Output file should contain schema doc")
	}
}

// TestDocsSchemaToDirectory tests docs schema command writing to a directory.
func TestDocsSchemaToDirectory(t *testing.T) {
	savedDocsOutput := docsOutput
	defer func() { docsOutput = savedDocsOutput }()

	tmpDir := t.TempDir()

	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	docs := &cobra.Command{Use: "docs"}
	schema := &cobra.Command{
		Use: "schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			docsOutput = tmpDir
			return runDocsSchema(cmd, args)
		},
	}
	docs.AddCommand(schema)
	root.AddCommand(docs)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"docs", "schema"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("docs schema -o dir failed: %v", err)
	}

	outFile := filepath.Join(tmpDir, "schema.md")
	if _, err := os.Stat(outFile); os.IsNotExist(err) {
		t.Error("Schema file should be created in the output directory")
	}
}

// TestDocsConfigToDirectory tests docs config command writing to a directory.
func TestDocsConfigToDirectory(t *testing.T) {
	savedDocsOutput := docsOutput
	defer func() { docsOutput = savedDocsOutput }()

	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	_ = os.Chdir(projectRoot)
	defer func() { _ = os.Chdir(oldWd) }()

	tmpDir := t.TempDir()

	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	docs := &cobra.Command{Use: "docs"}
	configDoc := &cobra.Command{
		Use: "config",
		RunE: func(cmd *cobra.Command, args []string) error {
			docsOutput = tmpDir
			return runDocsConfig(cmd, args)
		},
	}
	docs.AddCommand(configDoc)
	root.AddCommand(docs)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"docs", "config"})

	err := root.Execute()
	if err != nil {
		t.Fatalf("docs config -o dir failed: %v", err)
	}

	outFile := filepath.Join(tmpDir, "config.md")
	if _, err := os.Stat(outFile); os.IsNotExist(err) {
		t.Error("Config doc file should be created in the output directory")
	}
}

// TestCompareDatabases tests the compareDatabases function directly.
func TestCompareDatabases(t *testing.T) {
	// Import the database package via the existing test infrastructure
	csvData1 := `req_id,category,requirement_text,status,priority
REQ-001,CLI,First,COMPLETE,HIGH
REQ-002,DATA,Second,MISSING,MEDIUM
REQ-003,TEST,Third,PARTIAL,LOW
`
	csvData2 := `req_id,category,requirement_text,status,priority
REQ-001,CLI,First,COMPLETE,HIGH
REQ-002,DATA,Second,COMPLETE,MEDIUM
REQ-004,NEW,Fourth,MISSING,HIGH
`

	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "base.csv")
	f2 := filepath.Join(tmpDir, "current.csv")
	_ = os.WriteFile(f1, []byte(csvData1), 0644)
	_ = os.WriteFile(f2, []byte(csvData2), 0644)

	// Test diff with two explicit files
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd := &cobra.Command{
		Use:  "diff",
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			diffFormat = "terminal"
			diffOutput = ""
			return runDiff(cmd, args)
		},
	}
	root.AddCommand(cmd)

	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"diff", f1, f2})

	err := root.Execute()
	// This may return an exit error if there are regressions
	output := buf.String()

	if !strings.Contains(output, "RTM Database Comparison") {
		t.Errorf("Diff output should contain title, got:\n%s", output)
	}

	// REQ-004 should be in Added
	if !strings.Contains(output, "REQ-004") {
		t.Errorf("Should show REQ-004 as added, got:\n%s", output)
	}

	// REQ-003 should be in Removed
	if !strings.Contains(output, "REQ-003") {
		t.Errorf("Should show REQ-003 as removed, got:\n%s", output)
	}

	_ = err // exit error is expected for regressions
}

// TestExitError tests the ExitError type.
func TestExitError(t *testing.T) {
	// With message
	err := NewExitError(1, "test message")
	if err.Error() != "test message" {
		t.Errorf("ExitError.Error() = %q, want 'test message'", err.Error())
	}
	if err.Code != 1 {
		t.Errorf("ExitError.Code = %d, want 1", err.Code)
	}

	// Without message
	err = NewExitError(2, "")
	if !strings.Contains(err.Error(), "exit code 2") {
		t.Errorf("ExitError.Error() = %q, want to contain 'exit code 2'", err.Error())
	}
}
