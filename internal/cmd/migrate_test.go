package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// standardCSVHeader is the 21-column header for database.csv.
const standardCSVHeader = "req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file,external_id\n"

func TestMigrateCommand(t *testing.T) {
	rtmx.Req(t, "REQ-GO-046")

	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()

	t.Run("check_modern_project_all_pass", func(t *testing.T) {
		tmpDir := setupModernProject(t)
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		buf := new(bytes.Buffer)
		cmd := migrateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		// Reset flags
		migrateFix = false
		migrateCheck = false

		err := cmd.RunE(cmd, nil)
		if err != nil {
			t.Fatalf("migrate check failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "[PASS] Config file") {
			t.Errorf("expected config PASS, got: %s", out)
		}
		if !strings.Contains(out, "[PASS] Database file") {
			t.Errorf("expected database PASS, got: %s", out)
		}
		if !strings.Contains(out, "[PASS] Database schema") {
			t.Errorf("expected schema PASS, got: %s", out)
		}
		if !strings.Contains(out, "[PASS] Requirements directory") {
			t.Errorf("expected requirements dir PASS, got: %s", out)
		}
		if !strings.Contains(out, "6 passed, 0 warnings, 0 failed") {
			t.Errorf("expected all 6 passed summary, got: %s", out)
		}
	})

	t.Run("check_legacy_project_warns", func(t *testing.T) {
		tmpDir := setupLegacyProject(t)
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		buf := new(bytes.Buffer)
		cmd := migrateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		migrateFix = false
		migrateCheck = false

		err := cmd.RunE(cmd, nil)
		// Should not error since legacy issues are WARN, not FAIL
		if err != nil {
			t.Fatalf("migrate check failed unexpectedly: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "[WARN]") {
			t.Errorf("expected WARN in output, got: %s", out)
		}
		if !strings.Contains(out, "legacy") || !strings.Contains(out, "docs/rtm_database.csv") {
			t.Errorf("expected legacy database warning, got: %s", out)
		}
	})

	t.Run("check_missing_schema_columns_fails", func(t *testing.T) {
		tmpDir := setupBadSchemaProject(t)
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		buf := new(bytes.Buffer)
		cmd := migrateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		migrateFix = false
		migrateCheck = false

		err := cmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error for schema failure")
		}

		out := buf.String()
		if !strings.Contains(out, "[FAIL] Database schema") {
			t.Errorf("expected schema FAIL, got: %s", out)
		}
		if !strings.Contains(out, "expected 21 columns") {
			t.Errorf("expected column count error, got: %s", out)
		}
	})

	t.Run("check_no_project_found", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "rtmx-migrate-empty")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer func() { _ = os.RemoveAll(tmpDir) }()

		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		buf := new(bytes.Buffer)
		cmd := migrateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		migrateFix = false
		migrateCheck = false

		err = cmd.RunE(cmd, nil)
		if err == nil {
			t.Fatal("expected error for empty project")
		}

		out := buf.String()
		if !strings.Contains(out, "[FAIL]") {
			t.Errorf("expected FAIL in output, got: %s", out)
		}
	})

	t.Run("fix_moves_legacy_database", func(t *testing.T) {
		tmpDir := setupLegacyProject(t)
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		buf := new(bytes.Buffer)
		cmd := migrateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		migrateFix = true
		migrateCheck = false

		err := cmd.RunE(cmd, nil)
		if err != nil {
			t.Fatalf("migrate fix failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "moved docs/rtm_database.csv -> .rtmx/database.csv") {
			t.Errorf("expected move action, got: %s", out)
		}

		// Verify file was moved
		modernPath := filepath.Join(tmpDir, ".rtmx", "database.csv")
		if _, err := os.Stat(modernPath); err != nil {
			t.Errorf("expected modern database to exist after fix: %v", err)
		}

		// Verify backup was created
		legacyDir := filepath.Join(tmpDir, "docs")
		entries, _ := os.ReadDir(legacyDir)
		foundBackup := false
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "rtm_database.csv.bak.") {
				foundBackup = true
				break
			}
		}
		if !foundBackup {
			t.Error("expected backup file to be created")
		}

		// Verify legacy file was removed (moved)
		legacyPath := filepath.Join(tmpDir, "docs", "rtm_database.csv")
		if _, err := os.Stat(legacyPath); err == nil {
			t.Error("expected legacy database to be removed after fix")
		}
	})

	t.Run("fix_updates_python_hooks", func(t *testing.T) {
		tmpDir := setupPythonHooksProject(t)
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		buf := new(bytes.Buffer)
		cmd := migrateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		migrateFix = true
		migrateCheck = false

		err := cmd.RunE(cmd, nil)
		if err != nil {
			t.Fatalf("migrate fix failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "updated pre-commit") {
			t.Errorf("expected hook update action, got: %s", out)
		}

		// Verify hook was updated
		hookPath := filepath.Join(tmpDir, ".githooks", "pre-commit")
		data, err := os.ReadFile(hookPath)
		if err != nil {
			t.Fatalf("failed to read hook: %v", err)
		}

		content := string(data)
		if strings.Contains(content, "python") {
			t.Error("expected python references to be removed from hook")
		}
		if !strings.Contains(content, "rtmx") {
			t.Error("expected rtmx reference in updated hook")
		}
	})

	t.Run("fix_creates_requirements_dir", func(t *testing.T) {
		tmpDir := setupNoReqDirProject(t)
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		buf := new(bytes.Buffer)
		cmd := migrateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		migrateFix = true
		migrateCheck = false

		err := cmd.RunE(cmd, nil)
		if err != nil {
			t.Fatalf("migrate fix failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "created .rtmx/requirements/") {
			t.Errorf("expected requirements dir creation action, got: %s", out)
		}

		// Verify directory was created
		reqDir := filepath.Join(tmpDir, ".rtmx", "requirements")
		info, err := os.Stat(reqDir)
		if err != nil {
			t.Errorf("expected requirements dir to exist: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected requirements path to be a directory")
		}
	})

	t.Run("default_mode_is_check", func(t *testing.T) {
		tmpDir := setupModernProject(t)
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatalf("failed to chdir: %v", err)
		}

		buf := new(bytes.Buffer)
		cmd := migrateCmd
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		// Neither --fix nor --check set
		migrateFix = false
		migrateCheck = false

		err := cmd.RunE(cmd, nil)
		if err != nil {
			t.Fatalf("migrate default mode failed: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "Migration Compatibility Report") {
			t.Errorf("expected report header, got: %s", out)
		}
	})
}

// setupModernProject creates a temp directory with a modern RTMX project layout.
func setupModernProject(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rtmx-migrate-modern")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Create .rtmx directory with config and database
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0755); err != nil {
		t.Fatalf("failed to create .rtmx dir: %v", err)
	}

	// Create config
	configContent := "rtmx:\n  database: .rtmx/database.csv\n  requirements_dir: .rtmx/requirements\n  schema: core\n"
	if err := os.WriteFile(filepath.Join(rtmxDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create database with 21 columns
	dbRow := "REQ-TEST-001,CLI,Foundation,Test requirement,target,test.go,TestFunc,Unit Test,MISSING,HIGH,1,notes,1,,,,,,,.rtmx/requirements/CLI/REQ-TEST-001.md,\n"
	if err := os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(standardCSVHeader+dbRow), 0644); err != nil {
		t.Fatalf("failed to write database: %v", err)
	}

	// Create requirements directory
	reqDir := filepath.Join(rtmxDir, "requirements")
	if err := os.MkdirAll(reqDir, 0755); err != nil {
		t.Fatalf("failed to create requirements dir: %v", err)
	}

	return tmpDir
}

// setupLegacyProject creates a temp directory with legacy Python project layout.
func setupLegacyProject(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rtmx-migrate-legacy")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Create .rtmx directory with config but NO database at modern path
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0755); err != nil {
		t.Fatalf("failed to create .rtmx dir: %v", err)
	}

	configContent := "rtmx:\n  database: .rtmx/database.csv\n  requirements_dir: .rtmx/requirements\n  schema: core\n"
	if err := os.WriteFile(filepath.Join(rtmxDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create requirements directory
	if err := os.MkdirAll(filepath.Join(rtmxDir, "requirements"), 0755); err != nil {
		t.Fatalf("failed to create requirements dir: %v", err)
	}

	// Create legacy database at docs/rtm_database.csv
	docsDir := filepath.Join(tmpDir, "docs")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("failed to create docs dir: %v", err)
	}

	dbRow := "REQ-TEST-001,CLI,Foundation,Test requirement,target,test.go,TestFunc,Unit Test,MISSING,HIGH,1,notes,1,,,,,,,.rtmx/requirements/CLI/REQ-TEST-001.md,\n"
	if err := os.WriteFile(filepath.Join(docsDir, "rtm_database.csv"), []byte(standardCSVHeader+dbRow), 0644); err != nil {
		t.Fatalf("failed to write legacy database: %v", err)
	}

	return tmpDir
}

// setupBadSchemaProject creates a project with wrong number of CSV columns.
func setupBadSchemaProject(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rtmx-migrate-badschema")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0755); err != nil {
		t.Fatalf("failed to create .rtmx dir: %v", err)
	}

	configContent := "rtmx:\n  database: .rtmx/database.csv\n  requirements_dir: .rtmx/requirements\n  schema: core\n"
	if err := os.WriteFile(filepath.Join(rtmxDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create database with only 10 columns (wrong schema)
	badHeader := "req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority\n"
	badRow := "REQ-TEST-001,CLI,Foundation,Test,target,test.go,TestFunc,Unit Test,MISSING,HIGH\n"
	if err := os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(badHeader+badRow), 0644); err != nil {
		t.Fatalf("failed to write database: %v", err)
	}

	// Create requirements directory
	if err := os.MkdirAll(filepath.Join(rtmxDir, "requirements"), 0755); err != nil {
		t.Fatalf("failed to create requirements dir: %v", err)
	}

	return tmpDir
}

// setupPythonHooksProject creates a project with Python-referencing git hooks.
func setupPythonHooksProject(t *testing.T) string {
	t.Helper()
	tmpDir := setupModernProject(t)

	// Create .githooks directory with Python-referencing pre-commit
	hooksDir := filepath.Join(tmpDir, ".githooks")
	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		t.Fatalf("failed to create .githooks dir: %v", err)
	}

	hookContent := "#!/bin/bash\npython -m rtmx verify\npytest tests/\n"
	if err := os.WriteFile(filepath.Join(hooksDir, "pre-commit"), []byte(hookContent), 0755); err != nil {
		t.Fatalf("failed to write hook: %v", err)
	}

	return tmpDir
}

// setupNoReqDirProject creates a project without a requirements directory.
func setupNoReqDirProject(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rtmx-migrate-noreqdir")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0755); err != nil {
		t.Fatalf("failed to create .rtmx dir: %v", err)
	}

	configContent := "rtmx:\n  database: .rtmx/database.csv\n  requirements_dir: .rtmx/requirements\n  schema: core\n"
	if err := os.WriteFile(filepath.Join(rtmxDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Create database with 21 columns
	dbRow := "REQ-TEST-001,CLI,Foundation,Test requirement,target,test.go,TestFunc,Unit Test,MISSING,HIGH,1,notes,1,,,,,,,.rtmx/requirements/CLI/REQ-TEST-001.md,\n"
	if err := os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(standardCSVHeader+dbRow), 0644); err != nil {
		t.Fatalf("failed to write database: %v", err)
	}

	// Deliberately do NOT create .rtmx/requirements/

	return tmpDir
}
