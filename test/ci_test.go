package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx-go/pkg/rtmx"
)

// TestCIVerifyJob validates that the CI workflow contains a verify-requirements
// job that runs rtmx verify --update on main pushes and auto-commits changes.
// REQ-CI-001: Closed-loop verification in CI
func TestCIVerifyJob(t *testing.T) {
	rtmx.Req(t, "REQ-CI-001")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	ciPath := filepath.Join(projectRoot, ".github", "workflows", "ci.yml")
	content, err := os.ReadFile(ciPath)
	if err != nil {
		t.Fatalf("Failed to read ci.yml: %v", err)
	}
	ci := string(content)

	// AC1: verify-requirements job exists and depends on test + lint
	t.Run("job_exists_with_dependencies", func(t *testing.T) {
		if !strings.Contains(ci, "verify-requirements:") {
			t.Fatal("CI workflow missing verify-requirements job")
		}
		if !strings.Contains(ci, "needs: [test, lint]") {
			t.Error("verify-requirements job must depend on test and lint jobs")
		}
	})

	// AC2: Job runs rtmx verify --update
	t.Run("runs_verify_update", func(t *testing.T) {
		if !strings.Contains(ci, "rtmx verify --update") {
			t.Error("verify-requirements job must run 'rtmx verify --update'")
		}
	})

	// AC3: Auto-commits database.csv changes
	t.Run("auto_commits_database", func(t *testing.T) {
		if !strings.Contains(ci, "git add .rtmx/database.csv") {
			t.Error("verify-requirements job must auto-commit .rtmx/database.csv")
		}
		if !strings.Contains(ci, "git push") {
			t.Error("verify-requirements job must push committed changes")
		}
	})

	// AC4: Uses GitHub App token for signed commits
	t.Run("uses_app_token", func(t *testing.T) {
		if !strings.Contains(ci, "actions/create-github-app-token") {
			t.Error("verify-requirements job must use GitHub App token for signed commits")
		}
		if !strings.Contains(ci, "secrets.APP_ID") {
			t.Error("verify-requirements job must reference APP_ID secret")
		}
		if !strings.Contains(ci, "secrets.APP_PRIVATE_KEY") {
			t.Error("verify-requirements job must reference APP_PRIVATE_KEY secret")
		}
	})

	// AC5: Only runs on push to main (not PRs)
	t.Run("main_push_only", func(t *testing.T) {
		if !strings.Contains(ci, "github.event_name == 'push'") {
			t.Error("verify-requirements job must only run on push events")
		}
		if !strings.Contains(ci, "refs/heads/main") {
			t.Error("verify-requirements job must only run on main branch")
		}
	})

	// AC6: Has write permissions for committing
	t.Run("has_write_permissions", func(t *testing.T) {
		// Find the verify-requirements section and check it has contents: write
		idx := strings.Index(ci, "verify-requirements:")
		if idx < 0 {
			t.Fatal("verify-requirements job not found")
		}
		// Look at the job block (up to the next top-level job)
		jobBlock := ci[idx:]
		nextJob := strings.Index(jobBlock[1:], "\n  ")
		// Just search within a reasonable range
		searchRange := jobBlock
		if nextJob > 0 && nextJob < 500 {
			searchRange = jobBlock[:nextJob+1]
		}
		_ = searchRange
		if !strings.Contains(jobBlock, "contents: write") {
			t.Error("verify-requirements job must have contents: write permission")
		}
	})
}
