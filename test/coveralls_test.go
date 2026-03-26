package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestCoverallsIntegration validates that the CI workflow integrates Coveralls
// for coverage badge reporting and that the README displays the badge.
// REQ-GO-068: Go CLI shall integrate Coveralls for coverage badges.
func TestCoverallsIntegration(t *testing.T) {
	rtmx.Req(t, "REQ-GO-068")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	// Read CI workflow
	ciPath := filepath.Join(projectRoot, ".github", "workflows", "ci.yml")
	ciContent, err := os.ReadFile(ciPath)
	if err != nil {
		t.Fatalf("Failed to read ci.yml: %v", err)
	}
	ci := string(ciContent)

	// Read README
	readmePath := filepath.Join(projectRoot, "README.md")
	readmeContent, err := os.ReadFile(readmePath)
	if err != nil {
		t.Fatalf("Failed to read README.md: %v", err)
	}
	readme := string(readmeContent)

	// AC1: CI workflow generates a coverage profile with coverprofile flag
	t.Run("coverage_profile_generated", func(t *testing.T) {
		if !strings.Contains(ci, "-coverprofile=") {
			t.Fatal("CI workflow must generate a coverage profile using -coverprofile flag")
		}
		if !strings.Contains(ci, "-covermode=atomic") {
			t.Error("CI workflow should use atomic cover mode for race-safe coverage")
		}
	})

	// AC2: CI workflow uploads coverage to Coveralls
	t.Run("coveralls_upload_step", func(t *testing.T) {
		if !strings.Contains(ci, "coverallsapp/github-action") {
			t.Fatal("CI workflow must use coverallsapp/github-action for Coveralls upload")
		}
		if !strings.Contains(ci, "coverage.out") {
			t.Error("Coveralls step must reference the coverage.out file")
		}
	})

	// AC3: README has a Coveralls coverage badge
	t.Run("readme_coverage_badge", func(t *testing.T) {
		if !strings.Contains(readme, "coveralls.io") {
			t.Fatal("README must contain a Coveralls coverage badge")
		}
		if !strings.Contains(readme, "Coverage Status") {
			t.Error("README badge should use 'Coverage Status' alt text")
		}
	})

	// AC4: Coveralls upload is conditional on ubuntu-latest (avoid duplicate uploads)
	t.Run("coveralls_os_condition", func(t *testing.T) {
		// Find the Coveralls block and verify it has an OS condition
		idx := strings.Index(ci, "coverallsapp/github-action")
		if idx < 0 {
			t.Fatal("coverallsapp/github-action not found in CI workflow")
		}
		// Look at the ~200 chars before the action reference for the condition
		start := idx - 200
		if start < 0 {
			start = 0
		}
		block := ci[start:idx]
		if !strings.Contains(block, "ubuntu-latest") {
			t.Error("Coveralls upload should be conditioned on ubuntu-latest to avoid duplicate uploads")
		}
	})
}
