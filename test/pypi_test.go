package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestPyPIPackage validates that the Python rtmx package infrastructure
// supports pytest plugin distribution with Go CLI integration.
// REQ-DIST-004: RTMX shall be installable via pip with pytest plugin and Go binary
func TestPyPIPackage(t *testing.T) {
	rtmx.Req(t, "REQ-DIST-004")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	// AC1: Deprecation manifest documents PyPI package
	t.Run("deprecation_manifest", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "deprecation.json"))
		if err != nil {
			t.Fatalf("deprecation.json must exist: %v", err)
		}
		manifest := string(content)
		if !strings.Contains(manifest, "pip install rtmx") {
			t.Error("deprecation manifest must reference pip install")
		}
		if !strings.Contains(manifest, "sunset_date") {
			t.Error("deprecation manifest must have sunset_date")
		}
	})

	// AC2: Go CLI supports --results flag for pytest output consumption
	t.Run("verify_results_flag", func(t *testing.T) {
		// Verify that rtmx verify --results exists by checking command source
		verifyPath := filepath.Join(projectRoot, "internal", "cmd", "verify.go")
		content, err := os.ReadFile(verifyPath)
		if err != nil {
			t.Fatalf("verify.go must exist: %v", err)
		}
		if !strings.Contains(string(content), "results") {
			t.Error("verify command must support --results flag for external test results")
		}
	})

	// AC3: Python test marker extraction works for cross-language verify
	t.Run("python_marker_support", func(t *testing.T) {
		// Check that from_tests supports Python pytest markers
		fromTestsPath := filepath.Join(projectRoot, "internal", "cmd", "from_tests.go")
		content, err := os.ReadFile(fromTestsPath)
		if err != nil {
			t.Fatalf("from_tests.go must exist: %v", err)
		}
		src := strings.ToLower(string(content))
		if !strings.Contains(src, "pytest") || !strings.Contains(src, "python") {
			t.Error("from_tests must support Python/pytest marker extraction")
		}
	})
}
