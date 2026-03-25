package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx-go/pkg/rtmx"
)

// TestDeprecationNotice validates the Python CLI deprecation notice artifacts.
// REQ-GO-045: Python CLI shall emit deprecation warnings pointing to Go CLI
func TestDeprecationNotice(t *testing.T) {
	rtmx.Req(t, "REQ-GO-045")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	// AC1: deprecation.json exists and contains required fields
	t.Run("deprecation_json_exists_and_valid", func(t *testing.T) {
		depPath := filepath.Join(projectRoot, "deprecation.json")
		content, err := os.ReadFile(depPath)
		if err != nil {
			t.Fatalf("deprecation.json must exist in repository root: %v", err)
		}

		var manifest map[string]interface{}
		if err := json.Unmarshal(content, &manifest); err != nil {
			t.Fatalf("deprecation.json must be valid JSON: %v", err)
		}

		requiredFields := []string{
			"deprecated_package",
			"replacement_package",
			"deprecation_date",
			"sunset_date",
			"migration_guide",
			"message",
		}
		for _, field := range requiredFields {
			val, ok := manifest[field]
			if !ok {
				t.Errorf("deprecation.json missing required field: %s", field)
				continue
			}
			str, isStr := val.(string)
			if !isStr || str == "" {
				t.Errorf("deprecation.json field %s must be a non-empty string", field)
			}
		}

		// Verify the migration guide URL points to the Go CLI repo
		if guide, ok := manifest["migration_guide"].(string); ok {
			if !strings.Contains(guide, "rtmx-go") {
				t.Error("migration_guide must reference rtmx-go repository")
			}
		}

		// Verify the message mentions deprecation
		if msg, ok := manifest["message"].(string); ok {
			if !strings.Contains(strings.ToLower(msg), "deprecated") {
				t.Error("message must mention deprecation")
			}
		}
	})

	// AC2: README.md contains migration section
	t.Run("readme_migration_section", func(t *testing.T) {
		readmePath := filepath.Join(projectRoot, "README.md")
		content, err := os.ReadFile(readmePath)
		if err != nil {
			t.Fatalf("README.md must exist: %v", err)
		}
		readme := string(content)

		if !strings.Contains(readme, "Migrating from Python") {
			t.Error("README.md must contain 'Migrating from Python' section")
		}
		if !strings.Contains(readme, "deprecated") || !strings.Contains(readme, "Deprecation") {
			t.Error("README.md migration section must mention deprecation")
		}
		if !strings.Contains(readme, "pip uninstall rtmx") {
			t.Error("README.md migration section must include pip uninstall instructions")
		}
		if !strings.Contains(readme, "end-of-life") {
			t.Error("README.md migration section must mention end-of-life date")
		}
	})

	// AC3: install.sh contains Python detection logic
	t.Run("install_sh_python_detection", func(t *testing.T) {
		scriptPath := filepath.Join(projectRoot, "scripts", "install.sh")
		content, err := os.ReadFile(scriptPath)
		if err != nil {
			t.Fatalf("scripts/install.sh must exist: %v", err)
		}
		script := string(content)

		if !strings.Contains(script, "pip show rtmx") && !strings.Contains(script, "pip3 show rtmx") {
			t.Error("install.sh must detect Python rtmx via pip")
		}
		if !strings.Contains(script, "deprecated") {
			t.Error("install.sh must print deprecation notice when Python rtmx is detected")
		}
		if !strings.Contains(script, "migration") || !strings.Contains(script, "Migration") {
			t.Error("install.sh must reference migration guide")
		}
		if !strings.Contains(script, "uninstall rtmx") {
			t.Error("install.sh must suggest removing the Python CLI")
		}
	})
}
