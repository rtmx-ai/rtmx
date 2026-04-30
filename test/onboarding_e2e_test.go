package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// buildBinary builds the rtmx binary and returns the path. Skips the test on failure.
func buildBinary(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rtmx-onboard-bin")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	binaryPath := filepath.Join(tmpDir, binaryName())
	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/rtmx")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\n%s", err, output)
	}
	return binaryPath
}

// runRtmx runs the rtmx binary in the given directory with the given args.
func runRtmx(t *testing.T, binary, dir string, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(binary, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// TestInitThenSetup verifies that running init then setup produces a single
// coherent .rtmx/ structure with no legacy docs/ directory.
func TestInitThenSetup(t *testing.T) {
	// Traces to REQ-E2E-005 via test_function in database.
	// Marker omitted: reqIDPattern does not accept alphanumeric category (E2E).

	binary := buildBinary(t)
	tmpDir, err := os.MkdirTemp("", "rtmx-init-setup")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Initialize git so setup's git detection works
	gitInit := exec.Command("git", "init")
	gitInit.Dir = tmpDir
	if out, err := gitInit.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}

	// Step 1: rtmx init (creates modern .rtmx/ structure)
	out, err := runRtmx(t, binary, tmpDir, "init")
	if err != nil {
		t.Fatalf("rtmx init failed: %v\n%s", err, out)
	}

	// Verify init created .rtmx/
	if _, err := os.Stat(filepath.Join(tmpDir, ".rtmx", "database.csv")); err != nil {
		t.Fatal("init should create .rtmx/database.csv")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, ".rtmx", "config.yaml")); err != nil {
		t.Fatal("init should create .rtmx/config.yaml")
	}

	// Step 2: rtmx setup (should respect existing .rtmx/ structure)
	out, err = runRtmx(t, binary, tmpDir, "setup", "--skip-agents", "--skip-makefile")
	if err != nil {
		t.Fatalf("rtmx setup failed: %v\n%s", err, out)
	}

	// Assert: NO docs/ directory created
	if _, err := os.Stat(filepath.Join(tmpDir, "docs")); err == nil {
		t.Error("setup after init should NOT create docs/ directory")
	}

	// Assert: NO root rtmx.yaml created
	if _, err := os.Stat(filepath.Join(tmpDir, "rtmx.yaml")); err == nil {
		t.Error("setup after init should NOT create root rtmx.yaml")
	}

	// Assert: .rtmx/database.csv is the only database
	if _, err := os.Stat(filepath.Join(tmpDir, ".rtmx", "database.csv")); err != nil {
		t.Fatal(".rtmx/database.csv should still exist")
	}

	// Assert: .rtmx/requirements/ is the only req tree
	if _, err := os.Stat(filepath.Join(tmpDir, ".rtmx", "requirements")); err != nil {
		t.Fatal(".rtmx/requirements/ should still exist")
	}

	// Assert: rtmx status succeeds
	out, err = runRtmx(t, binary, tmpDir, "status")
	if err != nil {
		t.Fatalf("rtmx status should succeed after init+setup: %v\n%s", err, out)
	}
}

// TestSetupAlone verifies that running setup without prior init creates the
// modern .rtmx/ structure by default.
func TestSetupAlone(t *testing.T) {
	// Traces to REQ-E2E-005 via test_function in database.
	// Marker omitted: reqIDPattern does not accept alphanumeric category (E2E).

	binary := buildBinary(t)
	tmpDir, err := os.MkdirTemp("", "rtmx-setup-alone")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Initialize git
	gitInit := exec.Command("git", "init")
	gitInit.Dir = tmpDir
	if out, err := gitInit.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}

	// Run setup directly (no prior init)
	out, err := runRtmx(t, binary, tmpDir, "setup", "--skip-agents", "--skip-makefile")
	if err != nil {
		t.Fatalf("rtmx setup failed: %v\n%s", err, out)
	}

	// Assert: .rtmx/ structure created
	if _, err := os.Stat(filepath.Join(tmpDir, ".rtmx", "database.csv")); err != nil {
		t.Error("setup alone should create .rtmx/database.csv")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, ".rtmx", "config.yaml")); err != nil {
		t.Error("setup alone should create .rtmx/config.yaml")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, ".rtmx", "requirements")); err != nil {
		t.Error("setup alone should create .rtmx/requirements/")
	}

	// Assert: NO docs/ directory
	if _, err := os.Stat(filepath.Join(tmpDir, "docs")); err == nil {
		t.Error("setup alone should NOT create docs/ directory")
	}

	// Assert: NO root rtmx.yaml
	if _, err := os.Stat(filepath.Join(tmpDir, "rtmx.yaml")); err == nil {
		t.Error("setup alone should NOT create root rtmx.yaml")
	}

	// Assert: rtmx status succeeds
	out, err = runRtmx(t, binary, tmpDir, "status")
	if err != nil {
		t.Fatalf("rtmx status should succeed: %v\n%s", err, out)
	}
}

// TestSetupOnExistingProject verifies that running setup on a project with
// an existing populated .rtmx/ structure preserves all content.
func TestSetupOnExistingProject(t *testing.T) {
	// Traces to REQ-E2E-005 via test_function in database.
	// Marker omitted: reqIDPattern does not accept alphanumeric category (E2E).

	binary := buildBinary(t)
	tmpDir, err := os.MkdirTemp("", "rtmx-setup-existing")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Initialize git
	gitInit := exec.Command("git", "init")
	gitInit.Dir = tmpDir
	if out, err := gitInit.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}

	// Create pre-existing .rtmx/ with populated database
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(filepath.Join(rtmxDir, "requirements", "CORE"), 0755)

	dbContent := `req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file
REQ-CORE-001,CORE,AUTH,User authentication shall use OAuth2,OAuth2 flow works,tests/test_auth.py,test_oauth2,Integration Test,COMPLETE,P0,1,Core auth,2,,,,,,2026-01-01,2026-01-15,.rtmx/requirements/CORE/REQ-CORE-001.md
REQ-CORE-002,CORE,API,API shall return JSON responses,All endpoints return JSON,tests/test_api.py,test_json_response,Unit Test,MISSING,HIGH,2,API format,1,REQ-CORE-001,,,,,,,.rtmx/requirements/CORE/REQ-CORE-002.md
`
	if err := os.WriteFile(filepath.Join(rtmxDir, "database.csv"), []byte(dbContent), 0644); err != nil {
		t.Fatal(err)
	}

	configContent := `rtmx:
  database: .rtmx/database.csv
  requirements_dir: .rtmx/requirements
  schema: core
`
	if err := os.WriteFile(filepath.Join(rtmxDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	reqContent := "# REQ-CORE-001: User Authentication\n\nOAuth2 flow.\n"
	if err := os.WriteFile(filepath.Join(rtmxDir, "requirements", "CORE", "REQ-CORE-001.md"), []byte(reqContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Run setup on existing project
	out, err := runRtmx(t, binary, tmpDir, "setup", "--skip-agents", "--skip-makefile")
	if err != nil {
		t.Fatalf("rtmx setup on existing project failed: %v\n%s", err, out)
	}

	// Assert: existing database content preserved
	db, err := os.ReadFile(filepath.Join(rtmxDir, "database.csv"))
	if err != nil {
		t.Fatal("database.csv should still exist")
	}
	if !strings.Contains(string(db), "REQ-CORE-001") {
		t.Error("existing database content should be preserved")
	}
	if !strings.Contains(string(db), "REQ-CORE-002") {
		t.Error("all existing requirements should be preserved")
	}

	// Assert: existing requirement file preserved
	req, err := os.ReadFile(filepath.Join(rtmxDir, "requirements", "CORE", "REQ-CORE-001.md"))
	if err != nil {
		t.Fatal("existing requirement file should be preserved")
	}
	if !strings.Contains(string(req), "OAuth2") {
		t.Error("requirement file content should be preserved")
	}

	// Assert: no docs/ directory created
	if _, err := os.Stat(filepath.Join(tmpDir, "docs")); err == nil {
		t.Error("setup on existing .rtmx/ should NOT create docs/")
	}

	// Assert: rtmx status reports correct counts
	out, err = runRtmx(t, binary, tmpDir, "status")
	if err != nil {
		t.Fatalf("rtmx status should succeed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "1 complete") {
		t.Errorf("status should report 1 complete requirement, got:\n%s", out)
	}
}

// TestSetupLegacyMode verifies that setup detects and respects an existing
// legacy docs/ layout created by rtmx init --legacy.
func TestSetupLegacyMode(t *testing.T) {
	// Traces to REQ-E2E-005 via test_function in database.
	// Marker omitted: reqIDPattern does not accept alphanumeric category (E2E).

	binary := buildBinary(t)
	tmpDir, err := os.MkdirTemp("", "rtmx-setup-legacy")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Initialize git
	gitInit := exec.Command("git", "init")
	gitInit.Dir = tmpDir
	if out, err := gitInit.CombinedOutput(); err != nil {
		t.Fatalf("git init failed: %v\n%s", err, out)
	}

	// Step 1: rtmx init --legacy
	out, err := runRtmx(t, binary, tmpDir, "init", "--legacy")
	if err != nil {
		t.Fatalf("rtmx init --legacy failed: %v\n%s", err, out)
	}

	// Verify legacy structure
	if _, err := os.Stat(filepath.Join(tmpDir, "docs", "rtm_database.csv")); err != nil {
		t.Fatal("init --legacy should create docs/rtm_database.csv")
	}

	// Step 2: rtmx setup
	out, err = runRtmx(t, binary, tmpDir, "setup", "--skip-agents", "--skip-makefile")
	if err != nil {
		t.Fatalf("rtmx setup after legacy init failed: %v\n%s", err, out)
	}

	// Assert: setup detects docs/ layout and uses it
	if _, err := os.Stat(filepath.Join(tmpDir, "docs", "rtm_database.csv")); err != nil {
		t.Error("legacy database should still exist")
	}

	// Assert: no .rtmx/ directory created (legacy mode stays legacy)
	if _, err := os.Stat(filepath.Join(tmpDir, ".rtmx", "database.csv")); err == nil {
		t.Error("setup on legacy project should NOT create .rtmx/database.csv")
	}

	// Assert: rtmx status succeeds
	out, err = runRtmx(t, binary, tmpDir, "status")
	if err != nil {
		t.Fatalf("rtmx status should succeed: %v\n%s", err, out)
	}
}

// TestDogfoodSelf runs rtmx setup --dry-run against the rtmx repo itself
// to verify it detects the existing .rtmx/ structure and would not create docs/.
func TestDogfoodSelf(t *testing.T) {
	// Traces to REQ-E2E-005 via test_function in database.
	// Marker omitted: reqIDPattern does not accept alphanumeric category (E2E).

	binary := buildBinary(t)

	// Find project root (this test runs from test/ subdirectory)
	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, ".rtmx", "database.csv")); err != nil {
		projectRoot = wd
		if _, err := os.Stat(filepath.Join(projectRoot, ".rtmx", "database.csv")); err != nil {
			t.Skip("Cannot find rtmx repo .rtmx/ directory")
		}
	}

	// Run setup --dry-run against the rtmx repo itself
	out, err := runRtmx(t, binary, projectRoot, "setup", "--dry-run", "--skip-agents", "--skip-makefile")
	if err != nil {
		t.Fatalf("rtmx setup --dry-run on self failed: %v\n%s", err, out)
	}

	// Assert: detects existing .rtmx/ structure
	if !strings.Contains(out, "RTMX config: Found") {
		t.Errorf("should detect existing config, got:\n%s", out)
	}
	if !strings.Contains(out, "RTM database: Found") {
		t.Errorf("should detect existing database, got:\n%s", out)
	}

	// Assert: would NOT create docs/ directory (should skip since .rtmx/ exists)
	if strings.Contains(out, "docs/rtm_database.csv") {
		t.Errorf("should NOT reference docs/ paths when .rtmx/ exists, got:\n%s", out)
	}

	// Assert: references .rtmx/ paths
	if strings.Contains(out, "[CREATE]") && strings.Contains(out, "docs/") {
		t.Errorf("should not propose creating docs/ structure, got:\n%s", out)
	}
}
