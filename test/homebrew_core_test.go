package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestHomebrewCoreFormula validates that a homebrew-core compatible formula
// exists and meets Homebrew's submission requirements.
// REQ-DIST-006: Homebrew-core formula submission
func TestHomebrewCoreFormula(t *testing.T) {
	rtmx.Req(t, "REQ-DIST-006")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	// AC1: Formula file exists
	t.Run("formula_exists", func(t *testing.T) {
		formulaPath := filepath.Join(projectRoot, "Formula", "rtmx.rb")
		if _, err := os.Stat(formulaPath); err != nil {
			t.Fatalf("Formula/rtmx.rb must exist: %v", err)
		}
	})

	// AC2: Formula builds from source (not prebuilt binary)
	t.Run("builds_from_source", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "Formula", "rtmx.rb"))
		if err != nil {
			t.Fatalf("Formula/rtmx.rb must exist: %v", err)
		}
		formula := string(content)
		if !strings.Contains(formula, "go") && !strings.Contains(formula, "build") {
			t.Error("formula must build from source using Go")
		}
		if !strings.Contains(formula, `depends_on "go"`) {
			t.Error("formula must declare go build dependency")
		}
	})

	// AC3: Formula has test stanza
	t.Run("has_test_stanza", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "Formula", "rtmx.rb"))
		if err != nil {
			t.Fatalf("Formula/rtmx.rb must exist: %v", err)
		}
		formula := string(content)
		if !strings.Contains(formula, "test do") {
			t.Error("formula must have a test stanza")
		}
		if !strings.Contains(formula, "rtmx") {
			t.Error("test stanza must invoke rtmx")
		}
	})

	// AC4: Formula metadata
	t.Run("formula_metadata", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "Formula", "rtmx.rb"))
		if err != nil {
			t.Fatalf("Formula/rtmx.rb must exist: %v", err)
		}
		formula := string(content)
		required := []string{
			"desc",
			"homepage",
			"url",
			"license",
		}
		for _, field := range required {
			if !strings.Contains(formula, field) {
				t.Errorf("formula must have %q field", field)
			}
		}
		if !strings.Contains(formula, "Apache-2.0") {
			t.Error("formula license must be Apache-2.0")
		}
		if !strings.Contains(formula, "rtmx.ai") {
			t.Error("formula homepage must reference rtmx.ai")
		}
	})

	// AC5: Source URL points to GitHub archive (not binary download)
	t.Run("source_url", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "Formula", "rtmx.rb"))
		if err != nil {
			t.Fatalf("Formula/rtmx.rb must exist: %v", err)
		}
		formula := string(content)
		if !strings.Contains(formula, "github.com/rtmx-ai/rtmx") {
			t.Error("formula URL must point to rtmx-ai/rtmx GitHub repository")
		}
		if !strings.Contains(formula, "archive") {
			t.Error("formula URL must use GitHub archive (source tarball)")
		}
	})
}
