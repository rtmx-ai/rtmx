package sync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// setupTestProject creates a temp directory with .rtmx/database.csv containing the given requirements.
// Returns the project dir path.
func setupTestProject(t *testing.T, reqs []*database.Requirement) string {
	t.Helper()
	dir := t.TempDir()
	rtmxDir := filepath.Join(dir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write config
	configContent := "rtmx:\n  database: .rtmx/database.csv\n  project_name: test-project\n"
	if err := os.WriteFile(filepath.Join(rtmxDir, "config.yaml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create requirements dir
	reqDir := filepath.Join(rtmxDir, "requirements")
	if err := os.MkdirAll(reqDir, 0755); err != nil {
		t.Fatal(err)
	}

	db := database.NewDatabase()
	for _, req := range reqs {
		if err := db.Add(req); err != nil {
			t.Fatal(err)
		}
	}
	dbPath := filepath.Join(rtmxDir, "database.csv")
	if err := db.Save(dbPath); err != nil {
		t.Fatal(err)
	}

	return dir
}

// loadTestDB loads a database from a test project directory, failing the test on error.
func loadTestDB(t *testing.T, dir string) *database.Database {
	t.Helper()
	dbPath := filepath.Join(dir, ".rtmx", "database.csv")
	db, err := database.Load(dbPath)
	if err != nil {
		t.Fatalf("failed to load database from %s: %v", dbPath, err)
	}
	return db
}

func TestMoveRequirement(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
	)

	srcReq := database.NewRequirement("REQ-SRC-001")
	srcReq.Category = "CORE"
	srcReq.Subcategory = "Feature"
	srcReq.RequirementText = "Source requirement to move"
	srcReq.Status = database.StatusPartial
	srcReq.Priority = database.PriorityHigh
	srcReq.Phase = 3

	srcDir := setupTestProject(t, []*database.Requirement{srcReq})
	dstDir := setupTestProject(t, nil)

	srcDB := loadTestDB(t, srcDir)
	dstDB := loadTestDB(t, dstDir)

	opts := CrossRepoOptions{
		SrcDir: srcDir,
		DstDir: dstDir,
	}
	result, err := MoveRequirement(srcDB, dstDB, "REQ-SRC-001", opts)
	if err != nil {
		t.Fatalf("MoveRequirement failed: %v", err)
	}

	// Verify the requirement was transferred to destination
	dstReq := dstDB.Get("REQ-SRC-001")
	if dstReq == nil {
		t.Fatal("expected requirement to exist in destination database")
	}
	if dstReq.RequirementText != "Source requirement to move" {
		t.Errorf("expected requirement text to match, got %q", dstReq.RequirementText)
	}
	if dstReq.Category != "CORE" {
		t.Errorf("expected category CORE, got %q", dstReq.Category)
	}
	if dstReq.Status != database.StatusPartial {
		t.Errorf("expected status PARTIAL, got %q", dstReq.Status)
	}

	// Verify bidirectional external_id links
	srcReqAfter := srcDB.Get("REQ-SRC-001")
	if srcReqAfter == nil {
		t.Fatal("expected source requirement to still exist (move keeps reference)")
	}
	if srcReqAfter.ExternalID == "" {
		t.Error("expected source external_id to be set after move")
	}
	if dstReq.ExternalID == "" {
		t.Error("expected destination external_id to be set after move")
	}

	// Verify the result summary
	if result.MovedID == "" {
		t.Error("expected result to contain moved ID")
	}
	if result.SourceExternalID == "" {
		t.Error("expected result to contain source external_id")
	}
	if result.TargetExternalID == "" {
		t.Error("expected result to contain target external_id")
	}
}

func TestMoveRequirementWithIDOverride(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
	)

	srcReq := database.NewRequirement("REQ-SRC-001")
	srcReq.Category = "CORE"
	srcReq.RequirementText = "Requirement to move with ID override"

	srcDir := setupTestProject(t, []*database.Requirement{srcReq})
	dstDir := setupTestProject(t, nil)

	srcDB := loadTestDB(t, srcDir)
	dstDB := loadTestDB(t, dstDir)

	opts := CrossRepoOptions{
		SrcDir:     srcDir,
		DstDir:     dstDir,
		TargetID:   "REQ-DST-999",
	}
	_, err := MoveRequirement(srcDB, dstDB, "REQ-SRC-001", opts)
	if err != nil {
		t.Fatalf("MoveRequirement with ID override failed: %v", err)
	}

	// Should exist in destination under the overridden ID
	dstReq := dstDB.Get("REQ-DST-999")
	if dstReq == nil {
		t.Fatal("expected requirement to exist with overridden ID REQ-DST-999")
	}
	if dstReq.RequirementText != "Requirement to move with ID override" {
		t.Errorf("expected requirement text to match, got %q", dstReq.RequirementText)
	}

	// Original ID should NOT exist in destination
	if dstDB.Get("REQ-SRC-001") != nil {
		t.Error("expected original ID to not exist in destination")
	}
}

func TestCloneRequirement(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
	)

	srcReq := database.NewRequirement("REQ-SRC-001")
	srcReq.Category = "CORE"
	srcReq.Subcategory = "Feature"
	srcReq.RequirementText = "Source requirement to clone"
	srcReq.Status = database.StatusComplete
	srcReq.Priority = database.PriorityHigh
	srcReq.Phase = 2

	srcDir := setupTestProject(t, []*database.Requirement{srcReq})
	dstDir := setupTestProject(t, nil)

	srcDB := loadTestDB(t, srcDir)
	dstDB := loadTestDB(t, dstDir)

	opts := CrossRepoOptions{
		SrcDir: srcDir,
		DstDir: dstDir,
	}
	result, err := CloneRequirement(srcDB, dstDB, "REQ-SRC-001", opts)
	if err != nil {
		t.Fatalf("CloneRequirement failed: %v", err)
	}

	// Verify original is preserved in source
	srcReqAfter := srcDB.Get("REQ-SRC-001")
	if srcReqAfter == nil {
		t.Fatal("expected source requirement to still exist after clone")
	}
	if srcReqAfter.Status != database.StatusComplete {
		t.Errorf("expected source status to remain COMPLETE, got %q", srcReqAfter.Status)
	}

	// Verify copy exists in destination
	dstReq := dstDB.Get("REQ-SRC-001")
	if dstReq == nil {
		t.Fatal("expected requirement to exist in destination after clone")
	}
	if dstReq.RequirementText != "Source requirement to clone" {
		t.Errorf("expected requirement text to match, got %q", dstReq.RequirementText)
	}

	// Verify bidirectional external_id links
	if srcReqAfter.ExternalID == "" {
		t.Error("expected source external_id to be set after clone")
	}
	if dstReq.ExternalID == "" {
		t.Error("expected destination external_id to be set after clone")
	}

	// Verify result
	if result.ClonedID == "" {
		t.Error("expected result to contain cloned ID")
	}
}

func TestCloneRequirementPreservesOriginal(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
	)

	srcReq := database.NewRequirement("REQ-SRC-001")
	srcReq.Category = "CORE"
	srcReq.RequirementText = "Original stays in source"
	srcReq.Status = database.StatusPartial
	srcReq.TestModule = "src_test.go"
	srcReq.TestFunction = "TestSrc"

	srcDir := setupTestProject(t, []*database.Requirement{srcReq})
	dstDir := setupTestProject(t, nil)

	srcDB := loadTestDB(t, srcDir)
	dstDB := loadTestDB(t, dstDir)

	opts := CrossRepoOptions{
		SrcDir: srcDir,
		DstDir: dstDir,
	}
	_, err := CloneRequirement(srcDB, dstDB, "REQ-SRC-001", opts)
	if err != nil {
		t.Fatalf("CloneRequirement failed: %v", err)
	}

	// Source must preserve all original fields
	srcAfter := srcDB.Get("REQ-SRC-001")
	if srcAfter.Status != database.StatusPartial {
		t.Errorf("expected source status PARTIAL, got %q", srcAfter.Status)
	}
	if srcAfter.TestModule != "src_test.go" {
		t.Errorf("expected source test_module preserved, got %q", srcAfter.TestModule)
	}
	if srcAfter.TestFunction != "TestSrc" {
		t.Errorf("expected source test_function preserved, got %q", srcAfter.TestFunction)
	}
}

func TestDryRunMove(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
	)

	srcReq := database.NewRequirement("REQ-SRC-001")
	srcReq.Category = "CORE"
	srcReq.RequirementText = "Should not be moved in dry-run"

	srcDir := setupTestProject(t, []*database.Requirement{srcReq})
	dstDir := setupTestProject(t, nil)

	srcDB := loadTestDB(t, srcDir)
	dstDB := loadTestDB(t, dstDir)

	opts := CrossRepoOptions{
		SrcDir: srcDir,
		DstDir: dstDir,
		DryRun: true,
	}
	result, err := MoveRequirement(srcDB, dstDB, "REQ-SRC-001", opts)
	if err != nil {
		t.Fatalf("MoveRequirement dry-run failed: %v", err)
	}

	if !result.DryRun {
		t.Error("expected result to indicate dry-run")
	}

	// Databases should NOT be modified in dry-run
	if dstDB.Get("REQ-SRC-001") != nil {
		t.Error("expected destination database to be unchanged in dry-run")
	}

	srcReqAfter := srcDB.Get("REQ-SRC-001")
	if srcReqAfter.ExternalID != "" {
		t.Error("expected source external_id to be unchanged in dry-run")
	}
}

func TestDryRunClone(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
	)

	srcReq := database.NewRequirement("REQ-SRC-001")
	srcReq.Category = "CORE"
	srcReq.RequirementText = "Should not be cloned in dry-run"

	srcDir := setupTestProject(t, []*database.Requirement{srcReq})
	dstDir := setupTestProject(t, nil)

	srcDB := loadTestDB(t, srcDir)
	dstDB := loadTestDB(t, dstDir)

	opts := CrossRepoOptions{
		SrcDir: srcDir,
		DstDir: dstDir,
		DryRun: true,
	}
	result, err := CloneRequirement(srcDB, dstDB, "REQ-SRC-001", opts)
	if err != nil {
		t.Fatalf("CloneRequirement dry-run failed: %v", err)
	}

	if !result.DryRun {
		t.Error("expected result to indicate dry-run")
	}

	if dstDB.Get("REQ-SRC-001") != nil {
		t.Error("expected destination database to be unchanged in dry-run")
	}
}

func TestMoveErrorNotRtmxEnabled(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("integration"),
		rtmx.Technique("error"),
	)

	srcReq := database.NewRequirement("REQ-SRC-001")
	srcReq.Category = "CORE"
	srcReq.RequirementText = "Test error for non-rtmx target"

	srcDir := setupTestProject(t, []*database.Requirement{srcReq})
	dstDir := t.TempDir() // empty dir, not rtmx-enabled

	srcDB := loadTestDB(t, srcDir)
	dstDB := database.NewDatabase()

	opts := CrossRepoOptions{
		SrcDir: srcDir,
		DstDir: dstDir,
	}
	_, err := MoveRequirement(srcDB, dstDB, "REQ-SRC-001", opts)
	if err == nil {
		t.Fatal("expected error when target is not rtmx-enabled")
	}
}

func TestMoveErrorRequirementNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("integration"),
		rtmx.Technique("error"),
	)

	srcDir := setupTestProject(t, nil) // empty database
	dstDir := setupTestProject(t, nil)

	srcDB := loadTestDB(t, srcDir)
	dstDB := loadTestDB(t, dstDir)

	opts := CrossRepoOptions{
		SrcDir: srcDir,
		DstDir: dstDir,
	}
	_, err := MoveRequirement(srcDB, dstDB, "REQ-NONEXISTENT-001", opts)
	if err == nil {
		t.Fatal("expected error when requirement does not exist")
	}
}

func TestCloneErrorRequirementNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("integration"),
		rtmx.Technique("error"),
	)

	srcDir := setupTestProject(t, nil)
	dstDir := setupTestProject(t, nil)

	srcDB := loadTestDB(t, srcDir)
	dstDB := loadTestDB(t, dstDir)

	opts := CrossRepoOptions{
		SrcDir: srcDir,
		DstDir: dstDir,
	}
	_, err := CloneRequirement(srcDB, dstDB, "REQ-NONEXISTENT-001", opts)
	if err == nil {
		t.Fatal("expected error when requirement does not exist")
	}
}

func TestMoveRequirementWithSpecFile(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
	)

	srcReq := database.NewRequirement("REQ-SRC-001")
	srcReq.Category = "CORE"
	srcReq.RequirementText = "Requirement with spec file"
	srcReq.RequirementFile = "requirements/CORE/REQ-SRC-001.md"

	srcDir := setupTestProject(t, []*database.Requirement{srcReq})

	// Create the spec file in source
	specDir := filepath.Join(srcDir, ".rtmx", "requirements", "CORE")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	specContent := "# REQ-SRC-001\n\nThis is the spec file content.\n"
	if err := os.WriteFile(filepath.Join(specDir, "REQ-SRC-001.md"), []byte(specContent), 0644); err != nil {
		t.Fatal(err)
	}

	dstDir := setupTestProject(t, nil)

	srcDB := loadTestDB(t, srcDir)
	dstDB := loadTestDB(t, dstDir)

	opts := CrossRepoOptions{
		SrcDir: srcDir,
		DstDir: dstDir,
	}
	_, err := MoveRequirement(srcDB, dstDB, "REQ-SRC-001", opts)
	if err != nil {
		t.Fatalf("MoveRequirement with spec file failed: %v", err)
	}

	// Verify spec file was copied to destination
	dstSpecPath := filepath.Join(dstDir, ".rtmx", "requirements", "CORE", "REQ-SRC-001.md")
	content, err := os.ReadFile(dstSpecPath)
	if err != nil {
		t.Fatalf("expected spec file to exist in destination: %v", err)
	}
	if string(content) != specContent {
		t.Errorf("expected spec content to match, got %q", string(content))
	}
}

func TestCloneWithIDOverride(t *testing.T) {
	rtmx.Req(t, "REQ-GO-075",
		rtmx.Scope("integration"),
		rtmx.Technique("nominal"),
	)

	srcReq := database.NewRequirement("REQ-SRC-001")
	srcReq.Category = "CORE"
	srcReq.RequirementText = "Clone with ID override"

	srcDir := setupTestProject(t, []*database.Requirement{srcReq})
	dstDir := setupTestProject(t, nil)

	srcDB := loadTestDB(t, srcDir)
	dstDB := loadTestDB(t, dstDir)

	opts := CrossRepoOptions{
		SrcDir:   srcDir,
		DstDir:   dstDir,
		TargetID: "REQ-CLONE-001",
	}
	_, err := CloneRequirement(srcDB, dstDB, "REQ-SRC-001", opts)
	if err != nil {
		t.Fatalf("CloneRequirement with ID override failed: %v", err)
	}

	if dstDB.Get("REQ-CLONE-001") == nil {
		t.Fatal("expected cloned requirement to exist with overridden ID")
	}
	if dstDB.Get("REQ-SRC-001") != nil {
		t.Error("expected original ID to not exist in destination when ID is overridden")
	}
}
