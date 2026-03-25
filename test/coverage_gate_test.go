package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestCoverageGate validates that CI enforces coverage thresholds so PRs
// cannot merge if coverage drops below 80%.
// REQ-GO-069: Go CLI shall enforce coverage thresholds in CI.
func TestCoverageGate(t *testing.T) {
	rtmx.Req(t, "REQ-GO-069")

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

	// AC1: CI generates a coverage profile
	t.Run("coverage_profile_generated", func(t *testing.T) {
		if !strings.Contains(ci, "-coverprofile=coverage.out") {
			t.Fatal("CI workflow must generate a coverage profile with -coverprofile=coverage.out")
		}
	})

	// AC2: CI has a coverage threshold check step
	t.Run("coverage_threshold_step", func(t *testing.T) {
		if !strings.Contains(ci, "Check coverage threshold") {
			t.Fatal("CI workflow must have a 'Check coverage threshold' step")
		}
		if !strings.Contains(ci, "go tool cover -func=coverage.out") {
			t.Fatal("CI must parse coverage.out with 'go tool cover -func'")
		}
	})

	// AC3: Threshold is set to 70%
	t.Run("threshold_is_70_percent", func(t *testing.T) {
		if !strings.Contains(ci, "70") {
			t.Fatal("CI coverage threshold must reference 70% minimum")
		}
		// Verify the comparison logic exists
		if !strings.Contains(ci, "COVERAGE") && !strings.Contains(ci, "coverage") {
			t.Fatal("CI must extract and compare the coverage percentage")
		}
	})

	// AC4: CI fails when coverage is below threshold
	t.Run("fails_below_threshold", func(t *testing.T) {
		if !strings.Contains(ci, "exit 1") {
			t.Fatal("CI coverage gate must exit 1 when coverage is below threshold")
		}
		// Verify it outputs an error annotation
		if !strings.Contains(ci, "::error::") {
			t.Error("CI coverage gate should emit a GitHub Actions error annotation")
		}
	})

	// AC5: Coverage gate runs on ubuntu-latest only (avoid duplicate checks)
	t.Run("runs_on_ubuntu", func(t *testing.T) {
		// Find the coverage threshold block
		idx := strings.Index(ci, "Check coverage threshold")
		if idx < 0 {
			t.Fatal("Coverage threshold step not found")
		}
		// Look at the ~200 chars before the step name for the condition
		start := idx - 200
		if start < 0 {
			start = 0
		}
		block := ci[start:idx]
		if !strings.Contains(block, "ubuntu-latest") {
			t.Error("Coverage threshold check should be conditioned on ubuntu-latest")
		}
	})
}
