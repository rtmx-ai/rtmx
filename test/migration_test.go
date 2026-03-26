package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestTrunkMigration validates that the rtmx main branch has been
// migrated to the Go implementation.
// REQ-MIG-002: Main branch migration to Go
func TestTrunkMigration(t *testing.T) {
	rtmx.Req(t, "REQ-MIG-002")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	// AC1: Go codebase is on main (go.mod exists with correct module)
	t.Run("go_module_exists", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "go.mod"))
		if err != nil {
			t.Fatal("go.mod must exist on main branch")
		}
		if !strings.Contains(string(content), "module github.com/rtmx-ai/rtmx") {
			t.Error("go.mod module must be github.com/rtmx-ai/rtmx")
		}
	})

	// AC2: Go CLI binary builds
	t.Run("go_cmd_exists", func(t *testing.T) {
		if _, err := os.Stat(filepath.Join(projectRoot, "cmd", "rtmx", "main.go")); err != nil {
			t.Fatal("cmd/rtmx/main.go must exist")
		}
	})

	// AC3: RTM database exists with Go requirements
	t.Run("rtm_database_exists", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".rtmx", "database.csv"))
		if err != nil {
			t.Fatal(".rtmx/database.csv must exist")
		}
		csv := string(content)
		if !strings.Contains(csv, "REQ-GO-") {
			t.Error("database must contain Go requirements (REQ-GO-*)")
		}
		if !strings.Contains(csv, "REQ-MIG-") {
			t.Error("database must contain migration requirements (REQ-MIG-*)")
		}
	})

	// AC4: Python CLI is not on main
	t.Run("no_python_cli", func(t *testing.T) {
		if _, err := os.Stat(filepath.Join(projectRoot, "src", "rtmx", "cli", "main.py")); err == nil {
			t.Error("Python CLI (src/rtmx/cli/main.py) must not exist on main after migration")
		}
	})

	// AC5: Legacy Python branch preserved
	t.Run("deprecation_manifest", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "deprecation.json"))
		if err != nil {
			t.Fatal("deprecation.json must exist")
		}
		if !strings.Contains(string(content), "rtmx") {
			t.Error("deprecation manifest must reference rtmx")
		}
	})

	// AC6: Migration tooling available
	t.Run("migrate_command_exists", func(t *testing.T) {
		if _, err := os.Stat(filepath.Join(projectRoot, "internal", "cmd", "migrate.go")); err != nil {
			t.Fatal("internal/cmd/migrate.go must exist")
		}
	})
}
