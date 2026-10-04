package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/adapters"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestImportPersistsMappedStatus(t *testing.T) {
	rtmx.Req(t, "REQ-SYNC-001b")

	req := database.NewRequirement("REQ-TEST-001")
	req.Category = "TEST"
	req.RequirementText = "Linked requirement"
	req.Status = database.StatusMissing
	req.ExternalID = "EXT-1"

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)
	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-1", Title: "Linked", Status: "closed"},
		},
		statusMapping: map[string]database.Status{
			"closed": database.StatusComplete,
		},
	}

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	result := runImport(adapter, cfg, false)
	_ = w.Close()
	os.Stdout = old

	if len(result.Errors) != 0 {
		t.Fatalf("errors: %+v", result.Errors)
	}
	reloaded, err := database.Load(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	got := reloaded.Get("REQ-TEST-001")
	if got == nil || got.Status != database.StatusComplete {
		t.Fatalf("status after reload = %#v", got)
	}
}

func TestImportPersistsDiscoveredLinkage(t *testing.T) {
	rtmx.Req(t, "REQ-SYNC-001b")

	req := database.NewRequirement("REQ-TEST-002")
	req.Category = "TEST"
	req.RequirementText = "Unlinked requirement"
	req.Status = database.StatusMissing

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)
	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-99", Title: "Mentions req", Status: "open", RequirementID: "REQ-TEST-002"},
		},
		statusMapping: map[string]database.Status{"open": database.StatusMissing},
	}

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	result := runImport(adapter, cfg, false)
	_ = w.Close()
	os.Stdout = old

	if len(result.Errors) != 0 {
		t.Fatalf("errors: %+v", result.Errors)
	}
	reloaded, err := database.Load(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	got := reloaded.Get("REQ-TEST-002")
	if got == nil || got.ExternalID != "EXT-99" {
		t.Fatalf("external_id after reload = %#v", got)
	}
}

func TestImportUnchangedLinkedDoesNotRewrite(t *testing.T) {
	rtmx.Req(t, "REQ-SYNC-001b")

	req := database.NewRequirement("REQ-TEST-003")
	req.Category = "TEST"
	req.RequirementText = "Stable"
	req.Status = database.StatusComplete
	req.ExternalID = "EXT-1"

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	before, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	cfg := createTestConfig(dbPath)
	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-1", Title: "Same", Status: "closed"},
		},
		statusMapping: map[string]database.Status{"closed": database.StatusComplete},
	}

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	result := runImport(adapter, cfg, false)
	_ = w.Close()
	os.Stdout = old

	if len(result.Updated) != 0 {
		t.Fatalf("updated = %v, want none", result.Updated)
	}
	if len(result.Skipped) != 1 {
		t.Fatalf("skipped = %v", result.Skipped)
	}
	after, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("database rewritten despite no changes")
	}
}

func TestImportPersistenceFailureIsSyncError(t *testing.T) {
	rtmx.Req(t, "REQ-SYNC-001b")

	req := database.NewRequirement("REQ-TEST-004")
	req.Category = "TEST"
	req.RequirementText = "Will fail save"
	req.Status = database.StatusMissing
	req.ExternalID = "EXT-1"

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "database.csv")
	db := database.NewDatabase()
	_ = db.Add(req)
	if err := db.Save(dbPath); err != nil {
		t.Fatal(err)
	}
	cfg := createTestConfig(dbPath)
	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-1", Title: "Linked", Status: "closed"},
		},
		statusMapping: map[string]database.Status{"closed": database.StatusComplete},
	}

	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	result := runImport(adapter, cfg, false)
	_ = w.Close()
	os.Stdout = old

	if len(result.Errors) == 0 {
		t.Fatal("expected persistence error")
	}
	if len(result.Updated) != 0 {
		t.Fatalf("updated should be cleared on save failure, got %v", result.Updated)
	}
}

func TestImportDryRunDoesNotMutate(t *testing.T) {
	rtmx.Req(t, "REQ-SYNC-001b")

	req := database.NewRequirement("REQ-TEST-005")
	req.Category = "TEST"
	req.RequirementText = "Dry run"
	req.Status = database.StatusMissing
	req.ExternalID = "EXT-1"

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	before, _ := os.ReadFile(dbPath)
	cfg := createTestConfig(dbPath)
	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-1", Title: "Linked", Status: "closed"},
		},
		statusMapping: map[string]database.Status{"closed": database.StatusComplete},
	}

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	_ = runImport(adapter, cfg, true)
	_ = w.Close()
	os.Stdout = old

	after, _ := os.ReadFile(dbPath)
	if string(before) != string(after) {
		t.Fatal("dry-run mutated database on disk")
	}
	reloaded, _ := database.Load(dbPath)
	got := reloaded.Get("REQ-TEST-005")
	if got.Status != database.StatusMissing {
		t.Fatalf("status mutated: %s", got.Status)
	}
}

func TestImportLoadFailureSurfaced(t *testing.T) {
	rtmx.Req(t, "REQ-SYNC-001b")

	dir := t.TempDir()
	dbPath := filepath.Join(dir, "database.csv")
	if err := os.WriteFile(dbPath, []byte("not,a,valid,rtm,csv\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cfg := createTestConfig(dbPath)
	adapter := &mockAdapter{name: "test-service", connected: true}

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	result := runImport(adapter, cfg, false)
	_ = w.Close()
	os.Stdout = old

	if len(result.Errors) == 0 {
		t.Fatal("expected load failure error")
	}
}
