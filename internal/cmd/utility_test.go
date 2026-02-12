package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestConfigCommand verifies the config command works
// REQ-GO-023: Go CLI config command shall display and validate config
func TestConfigCommand(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	os.Chdir(projectRoot)
	defer os.Chdir(oldWd)

	rootCmd := createConfigTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"config"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("config command failed: %v", err)
	}

	output := buf.String()
	expectedElements := []string{
		"RTMX Configuration",
		"Paths:",
		"Database:",
		"Schema:",
		"Phases:",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(output, elem) {
			t.Errorf("Expected config output to contain %q, got:\n%s", elem, output)
		}
	}
}

// TestMakefileCommand verifies the makefile command works
// REQ-GO-024: Go CLI makefile command shall generate Makefile targets
func TestMakefileCommand(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	os.Chdir(projectRoot)
	defer os.Chdir(oldWd)

	rootCmd := createMakefileTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"makefile"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("makefile command failed: %v", err)
	}

	output := buf.String()
	expectedElements := []string{
		"# RTMX Makefile Targets",
		".PHONY:",
		"rtm:",
		"@rtmx status",
		"backlog:",
		"health:",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(output, elem) {
			t.Errorf("Expected makefile output to contain %q, got:\n%s", elem, output)
		}
	}
}

// TestDocsCommand verifies the docs command works
// REQ-GO-025: Go CLI docs command shall generate schema documentation
func TestDocsCommand(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	os.Chdir(projectRoot)
	defer os.Chdir(oldWd)

	rootCmd := createDocsTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"docs", "schema"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("docs schema command failed: %v", err)
	}

	output := buf.String()
	expectedElements := []string{
		"# RTMX Database Schema",
		"## Core Schema Fields",
		"req_id",
		"status",
		"priority",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(output, elem) {
			t.Errorf("Expected docs output to contain %q, got:\n%s", elem, output)
		}
	}
}

// TestReconcileCommand verifies the reconcile command works
// REQ-GO-021: Go CLI reconcile command shall fix dependency issues
func TestReconcileCommand(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	os.Chdir(projectRoot)
	defer os.Chdir(oldWd)

	rootCmd := createReconcileTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"reconcile"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("reconcile command failed: %v", err)
	}

	output := buf.String()
	expectedElements := []string{
		"Dependency Reconciliation",
		"reciprocity",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(output, elem) {
			t.Errorf("Expected reconcile output to contain %q, got:\n%s", elem, output)
		}
	}
}

// TestDiffCommand verifies the diff command works
// REQ-GO-022: Go CLI diff command shall compare RTM databases
func TestDiffCommand(t *testing.T) {
	cwd, _ := os.Getwd()
	projectRoot := findProjectRootDir(cwd)
	if projectRoot == "" {
		t.Skip("Could not find project root with .rtmx")
	}

	oldWd, _ := os.Getwd()
	os.Chdir(projectRoot)
	defer os.Chdir(oldWd)

	// Compare database with itself
	dbPath := ".rtmx/database.csv"

	rootCmd := createDiffTestCmd()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"diff", dbPath})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("diff command failed: %v", err)
	}

	output := buf.String()
	expectedElements := []string{
		"RTM Database Comparison",
		"Statistics:",
		"Baseline",
		"Current",
		"Result:",
		"STABLE",
	}

	for _, elem := range expectedElements {
		if !strings.Contains(output, elem) {
			t.Errorf("Expected diff output to contain %q, got:\n%s", elem, output)
		}
	}
}

// Helper functions for creating test commands

func createConfigTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	var validate bool
	var format string

	cmd := &cobra.Command{
		Use: "config",
		RunE: func(cmd *cobra.Command, args []string) error {
			configValidate = validate
			configFormat = format
			return runConfig(cmd, args)
		},
	}
	cmd.Flags().BoolVar(&validate, "validate", false, "validate config")
	cmd.Flags().StringVar(&format, "format", "terminal", "output format")
	root.AddCommand(cmd)

	return root
}

func createMakefileTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	var output string

	cmd := &cobra.Command{
		Use: "makefile",
		RunE: func(cmd *cobra.Command, args []string) error {
			makefileOutput = output
			return runMakefile(cmd, args)
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file")
	root.AddCommand(cmd)

	return root
}

func createDocsTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	docs := &cobra.Command{
		Use: "docs",
	}

	schema := &cobra.Command{
		Use:  "schema",
		RunE: runDocsSchema,
	}

	config := &cobra.Command{
		Use:  "config",
		RunE: runDocsConfig,
	}

	docs.AddCommand(schema)
	docs.AddCommand(config)
	root.AddCommand(docs)

	return root
}

func createReconcileTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	var execute bool

	cmd := &cobra.Command{
		Use: "reconcile",
		RunE: func(cmd *cobra.Command, args []string) error {
			reconcileExecute = execute
			return runReconcile(cmd, args)
		},
	}
	cmd.Flags().BoolVar(&execute, "execute", false, "execute fixes")
	root.AddCommand(cmd)

	return root
}

func createDiffTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	var format, output string

	cmd := &cobra.Command{
		Use:  "diff",
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			diffFormat = format
			diffOutput = output
			return runDiff(cmd, args)
		},
	}
	cmd.Flags().StringVar(&format, "format", "terminal", "output format")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file")
	root.AddCommand(cmd)

	return root
}
