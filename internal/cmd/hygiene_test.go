package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestHygieneChecksDetectFindings(t *testing.T) {
	db := healthDBHeader +
		"REQ-001,CAT,Sub,Base requirement,Target,,,Unit Test,NOT_STARTED,HIGH,1,,1.0,,,team,seed,,,.rtmx/requirements/REQ-001.md,\n"
	dir := setupHealthTestProject(t, db)
	writeRequirementFile(t, dir, ".rtmx/requirements/REQ-001.md", "- [ ] Requirement language is reviewed\n")

	oldWd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(oldWd) }()

	root := createHygieneTestCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"hygiene"})

	if err := root.Execute(); err != nil {
		t.Fatalf("hygiene command failed: %v", err)
	}

	out := buf.String()
	for _, expected := range []string{
		"RTM Hygiene Check",
		"effort_bounds",
		"generic_owner",
		"missing_test_mapping",
		"missing_external_id",
		"generic_acceptance_criteria",
	} {
		if !strings.Contains(out, expected) {
			t.Fatalf("expected output to contain %q, got:\n%s", expected, out)
		}
	}
}

func TestHygieneJSONOutput(t *testing.T) {
	db := healthDBHeader +
		"REQ-001,CAT,Sub,Base requirement,Target,tests/test.py,test_req,Unit Test,NOT_STARTED,HIGH,1,,0.25,,,owner,seed,,,.rtmx/requirements/REQ-001.md,EXT-1\n"
	dir := setupHealthTestProject(t, db)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(oldWd) }()

	root := createHygieneTestCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"hygiene", "--json"})

	if err := root.Execute(); err != nil {
		t.Fatalf("hygiene --json failed: %v", err)
	}

	var result HygieneResult
	if err := json.Unmarshal(bytes.TrimSpace(buf.Bytes()), &result); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, buf.String())
	}
	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}
	if len(result.Findings) != 0 {
		t.Fatalf("expected no findings, got %+v", result.Findings)
	}
}

func TestHygieneStrictReturnsExitError(t *testing.T) {
	db := healthDBHeader +
		"REQ-001,CAT,Sub,Base requirement,Target,,,Unit Test,NOT_STARTED,HIGH,1,,1.0,,,team,seed,,,.rtmx/requirements/REQ-001.md,\n"
	dir := setupHealthTestProject(t, db)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer func() { _ = os.Chdir(oldWd) }()

	root := createHygieneTestCmd()
	root.SetOut(new(bytes.Buffer))
	root.SetArgs([]string{"hygiene", "--strict"})

	err := root.Execute()
	var exitErr *ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("expected ExitError, got %v", err)
	}
	if exitErr.Code != 1 {
		t.Fatalf("expected exit code 1, got %d", exitErr.Code)
	}
}

func createHygieneTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	var jsonOutput bool
	var strict bool
	var minEffort = 0.25
	var maxEffort = 0.5
	cmd := &cobra.Command{
		Use:     "hygiene",
		Aliases: []string{"hygeine"},
		RunE: func(cmd *cobra.Command, args []string) error {
			hygieneJSON = jsonOutput
			hygieneStrict = strict
			hygieneMinEffort = minEffort
			hygieneMaxEffort = maxEffort
			return runHygiene(cmd, args)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	cmd.Flags().BoolVar(&strict, "strict", false, "exit non-zero when findings are present")
	cmd.Flags().Float64Var(&minEffort, "min-effort", 0.25, "minimum actionable effort in weeks")
	cmd.Flags().Float64Var(&maxEffort, "max-effort", 0.5, "maximum actionable effort in weeks")
	root.AddCommand(cmd)
	return root
}

func writeRequirementFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}
