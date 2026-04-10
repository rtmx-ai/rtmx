package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestCIScreenshots validates that the CI workflow contains a screenshots job
// that captures terminal output from core commands for documentation.
// REQ-CI-004: CI shall generate terminal screenshots and create PRs to update documentation.
func TestCIScreenshots(t *testing.T) {
	rtmx.Req(t, "REQ-CI-004")

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

	// AC1: screenshots job exists in the workflow
	t.Run("job_exists", func(t *testing.T) {
		if !strings.Contains(ci, "screenshots:") {
			t.Fatal("CI workflow missing screenshots job")
		}
	})

	// AC2: screenshots job depends on test and lint
	t.Run("depends_on_test_and_lint", func(t *testing.T) {
		// Extract the screenshots job block
		idx := strings.Index(ci, "screenshots:")
		if idx < 0 {
			t.Fatal("screenshots job not found")
		}
		jobBlock := ci[idx:]
		// Look within the first 300 chars for the needs line
		end := 300
		if len(jobBlock) < end {
			end = len(jobBlock)
		}
		header := jobBlock[:end]
		if !strings.Contains(header, "needs:") {
			t.Fatal("screenshots job must have a needs declaration")
		}
		if !strings.Contains(header, "test") {
			t.Error("screenshots job must depend on test job")
		}
		if !strings.Contains(header, "lint") {
			t.Error("screenshots job must depend on lint job")
		}
	})

	// AC3: screenshots job only runs on push to main
	t.Run("main_push_only", func(t *testing.T) {
		idx := strings.Index(ci, "screenshots:")
		if idx < 0 {
			t.Fatal("screenshots job not found")
		}
		jobBlock := ci[idx:]
		end := 400
		if len(jobBlock) < end {
			end = len(jobBlock)
		}
		header := jobBlock[:end]
		if !strings.Contains(header, "github.event_name == 'push'") {
			t.Error("screenshots job must only run on push events")
		}
		if !strings.Contains(header, "refs/heads/main") {
			t.Error("screenshots job must only run on main branch")
		}
	})

	// AC4: screenshots job captures output from core commands
	t.Run("captures_core_commands", func(t *testing.T) {
		idx := strings.Index(ci, "screenshots:")
		if idx < 0 {
			t.Fatal("screenshots job not found")
		}
		jobBlock := ci[idx:]

		commands := []string{
			"rtmx status",
			"rtmx backlog",
			"rtmx health",
			"rtmx deps",
			"rtmx markers",
		}
		for _, cmd := range commands {
			if !strings.Contains(jobBlock, cmd) {
				t.Errorf("screenshots job must capture output from '%s'", cmd)
			}
		}
	})

	// AC5: screenshots job writes output to docs/screenshots/ directory
	t.Run("output_directory", func(t *testing.T) {
		idx := strings.Index(ci, "screenshots:")
		if idx < 0 {
			t.Fatal("screenshots job not found")
		}
		jobBlock := ci[idx:]
		if !strings.Contains(jobBlock, "docs/screenshots/") {
			t.Error("screenshots job must write output to docs/screenshots/ directory")
		}
	})

	// AC6: screenshots job uploads artifact
	t.Run("uploads_artifact", func(t *testing.T) {
		idx := strings.Index(ci, "screenshots:")
		if idx < 0 {
			t.Fatal("screenshots job not found")
		}
		jobBlock := ci[idx:]
		if !strings.Contains(jobBlock, "actions/upload-artifact") {
			t.Error("screenshots job must upload screenshots as artifact")
		}
	})

	// AC7: screenshots job builds rtmx binary
	t.Run("builds_binary", func(t *testing.T) {
		idx := strings.Index(ci, "screenshots:")
		if idx < 0 {
			t.Fatal("screenshots job not found")
		}
		jobBlock := ci[idx:]
		if !strings.Contains(jobBlock, "go build") {
			t.Error("screenshots job must build the rtmx binary")
		}
	})

	// AC8: screenshots job uses pinned action versions (REQ-SEC-004)
	t.Run("pinned_action_shas", func(t *testing.T) {
		idx := strings.Index(ci, "screenshots:")
		if idx < 0 {
			t.Fatal("screenshots job not found")
		}
		jobBlock := ci[idx:]

		// Check that actions/checkout uses a pinned version (SHA or tag like v5), not a branch
		checkoutIdx := strings.Index(jobBlock, "actions/checkout@")
		if checkoutIdx < 0 {
			t.Fatal("screenshots job must use actions/checkout")
		}
		afterAt := jobBlock[checkoutIdx+len("actions/checkout@"):]
		shaEnd := strings.IndexAny(afterAt, " \n\r\t")
		if shaEnd < 0 {
			shaEnd = len(afterAt)
		}
		ref := strings.TrimSpace(afterAt[:shaEnd])
		if ref == "main" || ref == "master" || ref == "" {
			t.Errorf("actions/checkout must be pinned to a version or SHA, got: %s", ref)
		}

		// Check that actions/upload-artifact uses a pinned version
		uploadIdx := strings.Index(jobBlock, "actions/upload-artifact@")
		if uploadIdx < 0 {
			t.Fatal("screenshots job must use actions/upload-artifact")
		}
		afterAt = jobBlock[uploadIdx+len("actions/upload-artifact@"):]
		shaEnd = strings.IndexAny(afterAt, " \n\r\t")
		if shaEnd < 0 {
			shaEnd = len(afterAt)
		}
		ref = strings.TrimSpace(afterAt[:shaEnd])
		if ref == "main" || ref == "master" || ref == "" {
			t.Errorf("actions/upload-artifact must be pinned to a version or SHA, got: %s", ref)
		}
	})
}
