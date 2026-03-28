package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

// newTestMoveCmd creates a fresh move command for testing.
func newTestMoveCmd() *cobra.Command {
	var to, id, branch string
	var dryRun, pr bool

	cmd := &cobra.Command{
		Use:  "move",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			moveTo = to
			moveID = id
			moveDryRun = dryRun
			moveBranch = branch
			movePR = pr
			return runMove(cmd, args)
		},
	}
	cmd.Flags().StringVar(&to, "to", "", "target repo path")
	cmd.Flags().StringVar(&id, "id", "", "override target requirement ID")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview changes without writing")
	cmd.Flags().StringVar(&branch, "branch", "", "create branch in target repo")
	cmd.Flags().BoolVar(&pr, "pr", false, "create pull request after writing")
	_ = cmd.MarkFlagRequired("to")
	return cmd
}

// newTestCloneCmd creates a fresh clone command for testing.
func newTestCloneCmd() *cobra.Command {
	var to, id, branch string
	var dryRun, pr bool

	cmd := &cobra.Command{
		Use:  "clone",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			moveTo = to
			moveID = id
			moveDryRun = dryRun
			moveBranch = branch
			movePR = pr
			return runClone(cmd, args)
		},
	}
	cmd.Flags().StringVar(&to, "to", "", "target repo path")
	cmd.Flags().StringVar(&id, "id", "", "override target requirement ID")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview changes without writing")
	cmd.Flags().StringVar(&branch, "branch", "", "create branch in target repo")
	cmd.Flags().BoolVar(&pr, "pr", false, "create pull request after writing")
	_ = cmd.MarkFlagRequired("to")
	return cmd
}

// newTestMoveRootCmd creates a root command with move and clone subcommands.
func newTestMoveRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newTestMoveCmd())
	root.AddCommand(newTestCloneCmd())
	return root
}

func TestMoveCommandHelp(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
	)

	root := newTestMoveRootCmd()
	output, err := executeCommand(root, "move", "--help")
	if err != nil {
		t.Fatalf("move --help failed: %v", err)
	}

	expectedPhrases := []string{
		"move",
		"--to",
		"--dry-run",
		"--id",
	}
	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("expected help to contain %q, got: %s", phrase, output)
		}
	}
}

func TestCloneCommandHelp(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
	)

	root := newTestMoveRootCmd()
	output, err := executeCommand(root, "clone", "--help")
	if err != nil {
		t.Fatalf("clone --help failed: %v", err)
	}

	expectedPhrases := []string{
		"clone",
		"--to",
		"--dry-run",
		"--id",
	}
	for _, phrase := range expectedPhrases {
		if !strings.Contains(output, phrase) {
			t.Errorf("expected help to contain %q, got: %s", phrase, output)
		}
	}
}

func TestMoveCommandEndToEnd(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
	)

	srcRows := [][]string{
		makeRow("REQ-MV-001", "CORE", "Requirement to move", "PARTIAL", "HIGH", 2, 1.0, ""),
	}
	srcDir := createTestProject(t, srcRows)
	dstDir := createTestProject(t, nil)

	origDir, _ := os.Getwd()
	_ = os.Chdir(srcDir)
	defer func() { _ = os.Chdir(origDir) }()

	root := newTestMoveRootCmd()
	output, err := executeCommand(root, "move", "REQ-MV-001", "--to", dstDir)
	if err != nil {
		t.Fatalf("move command failed: %v\noutput: %s", err, output)
	}

	if !strings.Contains(output, "Moved") {
		t.Errorf("expected output to contain 'Moved', got: %s", output)
	}
	if !strings.Contains(output, "external_id") {
		t.Errorf("expected output to contain 'external_id', got: %s", output)
	}

	// Verify destination database has the requirement
	dstDBPath := filepath.Join(dstDir, ".rtmx", "database.csv")
	dstDB, err := database.Load(dstDBPath)
	if err != nil {
		t.Fatalf("failed to load destination database: %v", err)
	}
	if dstDB.Get("REQ-MV-001") == nil {
		t.Error("expected requirement to exist in destination after move")
	}

	// Verify source still has reference with external_id
	srcDBPath := filepath.Join(srcDir, ".rtmx", "database.csv")
	srcDB, err := database.Load(srcDBPath)
	if err != nil {
		t.Fatalf("failed to load source database: %v", err)
	}
	srcReq := srcDB.Get("REQ-MV-001")
	if srcReq == nil {
		t.Fatal("expected source requirement to still exist after move")
	}
	if srcReq.ExternalID == "" {
		t.Error("expected source external_id to be set after move")
	}
}

func TestCloneCommandEndToEnd(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
	)

	srcRows := [][]string{
		makeRow("REQ-CL-001", "CORE", "Requirement to clone", "COMPLETE", "P0", 1, 0.5, ""),
	}
	srcDir := createTestProject(t, srcRows)
	dstDir := createTestProject(t, nil)

	origDir, _ := os.Getwd()
	_ = os.Chdir(srcDir)
	defer func() { _ = os.Chdir(origDir) }()

	root := newTestMoveRootCmd()
	output, err := executeCommand(root, "clone", "REQ-CL-001", "--to", dstDir)
	if err != nil {
		t.Fatalf("clone command failed: %v\noutput: %s", err, output)
	}

	if !strings.Contains(output, "Cloned") {
		t.Errorf("expected output to contain 'Cloned', got: %s", output)
	}

	// Verify both databases
	dstDBPath := filepath.Join(dstDir, ".rtmx", "database.csv")
	dstDB, err := database.Load(dstDBPath)
	if err != nil {
		t.Fatalf("failed to load destination database: %v", err)
	}
	dstReq := dstDB.Get("REQ-CL-001")
	if dstReq == nil {
		t.Fatal("expected requirement to exist in destination after clone")
	}
	if dstReq.ExternalID == "" {
		t.Error("expected destination external_id to be set")
	}

	srcDBPath := filepath.Join(srcDir, ".rtmx", "database.csv")
	srcDB, err := database.Load(srcDBPath)
	if err != nil {
		t.Fatalf("failed to load source database: %v", err)
	}
	srcReq := srcDB.Get("REQ-CL-001")
	if srcReq == nil {
		t.Fatal("expected source requirement to still exist after clone")
	}
	if srcReq.ExternalID == "" {
		t.Error("expected source external_id to be set after clone")
	}
}

func TestMoveCommandDryRun(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
	)

	srcRows := [][]string{
		makeRow("REQ-DR-001", "CORE", "Dry run test", "MISSING", "MEDIUM", 1, 1.0, ""),
	}
	srcDir := createTestProject(t, srcRows)
	dstDir := createTestProject(t, nil)

	origDir, _ := os.Getwd()
	_ = os.Chdir(srcDir)
	defer func() { _ = os.Chdir(origDir) }()

	root := newTestMoveRootCmd()
	output, err := executeCommand(root, "move", "REQ-DR-001", "--to", dstDir, "--dry-run")
	if err != nil {
		t.Fatalf("move --dry-run failed: %v\noutput: %s", err, output)
	}

	if !strings.Contains(output, "dry-run") {
		t.Errorf("expected output to contain 'dry-run', got: %s", output)
	}

	// Verify destination is unchanged
	dstDBPath := filepath.Join(dstDir, ".rtmx", "database.csv")
	dstDB, err := database.Load(dstDBPath)
	if err != nil {
		t.Fatalf("failed to load destination database: %v", err)
	}
	if dstDB.Get("REQ-DR-001") != nil {
		t.Error("expected destination to be unchanged in dry-run")
	}
}

func TestMoveCommandWithIDOverride(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
	)

	srcRows := [][]string{
		makeRow("REQ-ID-001", "CORE", "ID override test", "MISSING", "MEDIUM", 1, 1.0, ""),
	}
	srcDir := createTestProject(t, srcRows)
	dstDir := createTestProject(t, nil)

	origDir, _ := os.Getwd()
	_ = os.Chdir(srcDir)
	defer func() { _ = os.Chdir(origDir) }()

	root := newTestMoveRootCmd()
	output, err := executeCommand(root, "move", "REQ-ID-001", "--to", dstDir, "--id", "REQ-NEW-999")
	if err != nil {
		t.Fatalf("move --id failed: %v\noutput: %s", err, output)
	}

	dstDBPath := filepath.Join(dstDir, ".rtmx", "database.csv")
	dstDB, err := database.Load(dstDBPath)
	if err != nil {
		t.Fatalf("failed to load destination database: %v", err)
	}
	if dstDB.Get("REQ-NEW-999") == nil {
		t.Error("expected requirement with overridden ID to exist in destination")
	}
	if dstDB.Get("REQ-ID-001") != nil {
		t.Error("expected original ID to not exist in destination")
	}
}

func TestMoveCommandErrorNotRtmxEnabled(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("integration"),
		rtmx.Technique("error"),
	)

	srcRows := [][]string{
		makeRow("REQ-ERR-001", "CORE", "Error test", "MISSING", "MEDIUM", 1, 1.0, ""),
	}
	srcDir := createTestProject(t, srcRows)
	dstDir := t.TempDir() // not rtmx-enabled

	origDir, _ := os.Getwd()
	_ = os.Chdir(srcDir)
	defer func() { _ = os.Chdir(origDir) }()

	root := newTestMoveRootCmd()
	_, err := executeCommand(root, "move", "REQ-ERR-001", "--to", dstDir)
	if err == nil {
		t.Fatal("expected error when target is not rtmx-enabled")
	}
}

func TestMoveCommandErrorReqNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("integration"),
		rtmx.Technique("error"),
	)

	srcDir := createTestProject(t, nil)
	dstDir := createTestProject(t, nil)

	origDir, _ := os.Getwd()
	_ = os.Chdir(srcDir)
	defer func() { _ = os.Chdir(origDir) }()

	root := newTestMoveRootCmd()
	_, err := executeCommand(root, "move", "REQ-NONEXISTENT-001", "--to", dstDir)
	if err == nil {
		t.Fatal("expected error when requirement does not exist")
	}
}

func TestMoveCommandMissingToFlag(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("unit"),
		rtmx.Technique("error"),
	)

	root := newTestMoveRootCmd()
	_, err := executeCommand(root, "move", "REQ-X-001")
	if err == nil {
		t.Fatal("expected error when --to flag is missing")
	}
}

// createGitTestProject creates a test project that is also a git repo.
func createGitTestProject(t *testing.T, csvRows [][]string) string {
	t.Helper()
	dir := createTestProject(t, csvRows)

	// Initialize git repo
	cmds := [][]string{
		{"git", "init"},
		{"git", "add", "-A"},
		{"git", "commit", "-m", "initial"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@test.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@test.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git command %v failed: %v\n%s", args, err, out)
		}
	}
	return dir
}

func TestMoveBranchCreatesGitBranch(t *testing.T) {
	rtmx.Req(t, "REQ-GO-076",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
	)

	srcRows := [][]string{
		makeRow("REQ-BR-001", "CORE", "Branch test", "PARTIAL", "HIGH", 2, 1.0, ""),
	}
	srcDir := createTestProject(t, srcRows)
	dstDir := createGitTestProject(t, nil)

	origDir, _ := os.Getwd()
	_ = os.Chdir(srcDir)
	defer func() { _ = os.Chdir(origDir) }()

	root := newTestMoveRootCmd()
	output, err := executeCommand(root, "move", "REQ-BR-001", "--to", dstDir, "--branch", "rtmx-sync/test-branch")
	if err != nil {
		t.Fatalf("move --branch failed: %v\noutput: %s", err, output)
	}

	if !strings.Contains(output, "Branch: rtmx-sync/test-branch") {
		t.Errorf("expected output to mention branch, got: %s", output)
	}

	// Verify the branch was actually created in the target repo
	cmd := exec.Command("git", "branch", "--list", "rtmx-sync/test-branch")
	cmd.Dir = dstDir
	branchOut, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to list branches: %v", err)
	}
	if !strings.Contains(string(branchOut), "rtmx-sync/test-branch") {
		t.Errorf("expected branch to exist in target repo, got: %s", branchOut)
	}
}

func TestCloneBranchCreatesGitBranch(t *testing.T) {
	rtmx.Req(t, "REQ-GO-076",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
	)

	srcRows := [][]string{
		makeRow("REQ-CBR-001", "CORE", "Clone branch test", "COMPLETE", "P0", 1, 0.5, ""),
	}
	srcDir := createTestProject(t, srcRows)
	dstDir := createGitTestProject(t, nil)

	origDir, _ := os.Getwd()
	_ = os.Chdir(srcDir)
	defer func() { _ = os.Chdir(origDir) }()

	root := newTestMoveRootCmd()
	output, err := executeCommand(root, "clone", "REQ-CBR-001", "--to", dstDir, "--branch", "rtmx-sync/clone-branch")
	if err != nil {
		t.Fatalf("clone --branch failed: %v\noutput: %s", err, output)
	}

	if !strings.Contains(output, "Branch: rtmx-sync/clone-branch") {
		t.Errorf("expected output to mention branch, got: %s", output)
	}

	// Verify the branch was actually created
	cmd := exec.Command("git", "branch", "--list", "rtmx-sync/clone-branch")
	cmd.Dir = dstDir
	branchOut, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to list branches: %v", err)
	}
	if !strings.Contains(string(branchOut), "rtmx-sync/clone-branch") {
		t.Errorf("expected branch to exist in target repo, got: %s", branchOut)
	}
}

func TestMovePRFlagAccepted(t *testing.T) {
	rtmx.Req(t, "REQ-GO-076",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
	)

	root := newTestMoveRootCmd()
	output, err := executeCommand(root, "move", "--help")
	if err != nil {
		t.Fatalf("move --help failed: %v", err)
	}

	expectedFlags := []string{"--branch", "--pr"}
	for _, flag := range expectedFlags {
		if !strings.Contains(output, flag) {
			t.Errorf("expected help to contain %q, got: %s", flag, output)
		}
	}
}

func TestClonePRFlagAccepted(t *testing.T) {
	rtmx.Req(t, "REQ-GO-076",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
	)

	root := newTestMoveRootCmd()
	output, err := executeCommand(root, "clone", "--help")
	if err != nil {
		t.Fatalf("clone --help failed: %v", err)
	}

	expectedFlags := []string{"--branch", "--pr"}
	for _, flag := range expectedFlags {
		if !strings.Contains(output, flag) {
			t.Errorf("expected help to contain %q, got: %s", flag, output)
		}
	}
}

func TestMoveDryRunWithBranch(t *testing.T) {
	rtmx.Req(t, "REQ-GO-076",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
	)

	srcRows := [][]string{
		makeRow("REQ-DBR-001", "CORE", "Dry run branch test", "MISSING", "MEDIUM", 1, 1.0, ""),
	}
	srcDir := createTestProject(t, srcRows)
	dstDir := createTestProject(t, nil)

	origDir, _ := os.Getwd()
	_ = os.Chdir(srcDir)
	defer func() { _ = os.Chdir(origDir) }()

	root := newTestMoveRootCmd()
	output, err := executeCommand(root, "move", "REQ-DBR-001", "--to", dstDir, "--branch", "rtmx-sync/dry-branch", "--dry-run")
	if err != nil {
		t.Fatalf("move --dry-run --branch failed: %v\noutput: %s", err, output)
	}

	if !strings.Contains(output, "dry-run") {
		t.Errorf("expected output to contain 'dry-run', got: %s", output)
	}
	if !strings.Contains(output, "rtmx-sync/dry-branch") {
		t.Errorf("expected output to mention branch name, got: %s", output)
	}
	if !strings.Contains(output, "would be created") {
		t.Errorf("expected output to say branch would be created, got: %s", output)
	}
}

func TestCloneDryRunWithBranchAndPR(t *testing.T) {
	rtmx.Req(t, "REQ-GO-076",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
	)

	srcRows := [][]string{
		makeRow("REQ-DBRP-001", "CORE", "Dry run branch+PR test", "MISSING", "MEDIUM", 1, 1.0, ""),
	}
	srcDir := createTestProject(t, srcRows)
	dstDir := createTestProject(t, nil)

	origDir, _ := os.Getwd()
	_ = os.Chdir(srcDir)
	defer func() { _ = os.Chdir(origDir) }()

	root := newTestMoveRootCmd()
	output, err := executeCommand(root, "clone", "REQ-DBRP-001", "--to", dstDir, "--branch", "rtmx-sync/dry-pr", "--pr", "--dry-run")
	if err != nil {
		t.Fatalf("clone --dry-run --branch --pr failed: %v\noutput: %s", err, output)
	}

	if !strings.Contains(output, "dry-run") {
		t.Errorf("expected output to contain 'dry-run', got: %s", output)
	}
	if !strings.Contains(output, "rtmx-sync/dry-pr") {
		t.Errorf("expected output to mention branch name, got: %s", output)
	}
	if !strings.Contains(output, "PR: would be created") {
		t.Errorf("expected output to say PR would be created, got: %s", output)
	}
}
