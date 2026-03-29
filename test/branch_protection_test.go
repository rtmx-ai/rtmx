package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestBranchProtection validates that the repository's CI configuration and
// security command enforce branch protection for CI-only status changes.
// REQ-INT-003: Git Branch Protection Integration
func TestBranchProtection(t *testing.T) {
	rtmx.Req(t, "REQ-INT-003")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	ciPath := filepath.Join(projectRoot, ".github", "workflows", "ci.yml")
	ciContent, err := os.ReadFile(ciPath)
	if err != nil {
		t.Fatalf("Failed to read ci.yml: %v", err)
	}
	ci := string(ciContent)

	// AC1: verify-requirements job exists in CI workflow.
	t.Run("verify_job_exists", func(t *testing.T) {
		if !strings.Contains(ci, "verify-requirements:") {
			t.Fatal("CI workflow must contain a verify-requirements job")
		}
	})

	// AC2: verify-requirements job depends on test and lint (runs after both pass).
	t.Run("verify_job_depends_on_test_and_lint", func(t *testing.T) {
		if !strings.Contains(ci, "needs: [test, lint]") {
			t.Error("verify-requirements job must have needs: [test, lint]")
		}
	})

	// AC3: verify-requirements job has a file guard that checks only
	// database.csv was modified (prevents unexpected file changes).
	t.Run("verify_job_has_file_guard", func(t *testing.T) {
		// The file guard uses git diff --name-only to ensure only
		// .rtmx/database.csv was modified by the verify step.
		if !strings.Contains(ci, "git diff --name-only") {
			t.Error("verify-requirements job must include a file guard using 'git diff --name-only'")
		}
		if !strings.Contains(ci, ".rtmx/database.csv") {
			t.Error("verify-requirements job file guard must check for .rtmx/database.csv")
		}
	})

	// AC4: verify-requirements job runs rtmx verify --update so only CI
	// can transition requirement statuses on the protected branch.
	t.Run("verify_job_runs_verify_update", func(t *testing.T) {
		if !strings.Contains(ci, "rtmx verify --update") {
			t.Error("verify-requirements job must run 'rtmx verify --update'")
		}
	})

	// AC5: The security command checks branch protection via gh CLI.
	// This validates that `rtmx security` includes the branch_protection check
	// so users can audit their repository's enforcement posture.
	t.Run("security_command_checks_branch_protection", func(t *testing.T) {
		securityPath := filepath.Join(projectRoot, "internal", "cmd", "security.go")
		secContent, err := os.ReadFile(securityPath)
		if err != nil {
			t.Fatalf("Failed to read security.go: %v", err)
		}
		sec := string(secContent)

		// The security command must include the branch protection check function.
		if !strings.Contains(sec, "checkBranchProtection") {
			t.Error("security.go must contain checkBranchProtection function")
		}

		// The check must query the GitHub API for repository rules.
		if !strings.Contains(sec, "repos/%s/rules") {
			t.Error("checkBranchProtection must query GitHub API repos/{owner/repo}/rules endpoint")
		}

		// The check must produce a result with category "repository" and
		// name "branch_protection".
		if !strings.Contains(sec, `"branch_protection"`) {
			t.Error("branch protection check must use name \"branch_protection\"")
		}
		if !strings.Contains(sec, `"repository"`) {
			t.Error("branch protection check must use category \"repository\"")
		}
	})

	// AC6: Only CI (push to main) runs verify-requirements -- PRs do not.
	// This ensures agents/developers cannot trigger status changes from PRs.
	t.Run("verify_only_on_main_push", func(t *testing.T) {
		// Find the verify-requirements job block and confirm it has the
		// conditional that restricts it to push events on main.
		idx := strings.Index(ci, "verify-requirements:")
		if idx < 0 {
			t.Fatal("verify-requirements job not found")
		}
		jobBlock := ci[idx:]

		if !strings.Contains(jobBlock, "github.event_name == 'push'") {
			t.Error("verify-requirements must only run on push events")
		}
		if !strings.Contains(jobBlock, "refs/heads/main") {
			t.Error("verify-requirements must only run on main branch")
		}
	})

	// AC7: The file guard rejects unexpected file modifications -- not just
	// any file, specifically only .rtmx/database.csv is allowed to change.
	t.Run("file_guard_rejects_unexpected_changes", func(t *testing.T) {
		idx := strings.Index(ci, "verify-requirements:")
		if idx < 0 {
			t.Fatal("verify-requirements job not found")
		}
		jobBlock := ci[idx:]

		// The guard must have logic that fails if files other than database.csv changed.
		if !strings.Contains(jobBlock, "Unexpected files modified") {
			t.Error("file guard must report 'Unexpected files modified' when non-database files change")
		}
		if !strings.Contains(jobBlock, "Refusing to commit") {
			t.Error("file guard must refuse to commit when unexpected files are modified")
		}
	})
}
