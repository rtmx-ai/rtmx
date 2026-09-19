package cmd

import (
	"os"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/adapters"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestBidiPreferRemotePersistsStatus(t *testing.T) {
	rtmx.Req(t, "REQ-SYNC-001c")

	req := database.NewRequirement("REQ-TEST-001")
	req.Category = "TEST"
	req.RequirementText = "Conflict"
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
		statusMapping: map[string]database.Status{"closed": database.StatusComplete},
		updateResult:  true,
	}

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	result := runBidirectional(adapter, cfg, "prefer-remote", false)
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

func TestBidiPreferLocalReportsRemoteFailure(t *testing.T) {
	rtmx.Req(t, "REQ-SYNC-001c")

	req := database.NewRequirement("REQ-TEST-002")
	req.Category = "TEST"
	req.RequirementText = "Local wins fail"
	req.Status = database.StatusComplete
	req.ExternalID = "EXT-1"

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)
	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-1", Title: "Linked", Status: "open"},
		},
		statusMapping: map[string]database.Status{"open": database.StatusMissing},
		updateResult:  false,
	}

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	result := runBidirectional(adapter, cfg, "prefer-local", false)
	_ = w.Close()
	os.Stdout = old

	if len(result.Errors) == 0 {
		t.Fatal("expected remote update failure error")
	}
	if len(result.Updated) != 0 {
		t.Fatalf("updated = %v, want none on failure", result.Updated)
	}
}


func TestBidiRequirementIDMatchPersistsLinkage(t *testing.T) {
	rtmx.Req(t, "REQ-SYNC-001c")

	req := database.NewRequirement("REQ-TEST-003")
	req.Category = "TEST"
	req.RequirementText = "Unlinked"
	req.Status = database.StatusMissing

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)
	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-42", Title: "Mentions", Status: "open", RequirementID: "REQ-TEST-003"},
		},
		statusMapping: map[string]database.Status{"open": database.StatusMissing},
	}

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	result := runBidirectional(adapter, cfg, "prefer-remote", false)
	_ = w.Close()
	os.Stdout = old

	if len(result.Errors) != 0 {
		t.Fatalf("errors: %+v", result.Errors)
	}
	reloaded, err := database.Load(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	got := reloaded.Get("REQ-TEST-003")
	if got == nil || got.ExternalID != "EXT-42" {
		t.Fatalf("external_id = %#v", got)
	}
}

func TestBidiExportCandidateCheckpointsExternalID(t *testing.T) {
	rtmx.Req(t, "REQ-SYNC-001c")

	req := database.NewRequirement("REQ-TEST-004")
	req.Category = "TEST"
	req.RequirementText = "Export me"
	req.Status = database.StatusMissing

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)
	adapter := &mockAdapter{
		name:         "test-service",
		connected:    true,
		items:        nil,
		createResult: "EXT-NEW",
	}

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	result := runBidirectional(adapter, cfg, "prefer-remote", false)
	_ = w.Close()
	os.Stdout = old

	if len(result.Errors) != 0 {
		t.Fatalf("errors: %+v", result.Errors)
	}
	reloaded, err := database.Load(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	got := reloaded.Get("REQ-TEST-004")
	if got == nil || got.ExternalID != "EXT-NEW" {
		t.Fatalf("external_id = %#v", got)
	}
}

func TestBidiAdapterFailureInErrors(t *testing.T) {
	rtmx.Req(t, "REQ-SYNC-001c")

	dbPath := createTestDatabase(t, nil)
	cfg := createTestConfig(dbPath)
	adapter := &mockAdapter{name: "test-service", connected: true, fetchErr: errFetchBoom{}}

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	result := runBidirectional(adapter, cfg, "prefer-remote", false)
	_ = w.Close()
	os.Stdout = old

	if len(result.Errors) == 0 {
		t.Fatal("expected fetch error in SyncResult.Errors")
	}
}

type errFetchBoom struct{}

func (errFetchBoom) Error() string { return "boom" }

func TestBidiSecondRunConverges(t *testing.T) {
	rtmx.Req(t, "REQ-SYNC-001c")

	req := database.NewRequirement("REQ-TEST-005")
	req.Category = "TEST"
	req.RequirementText = "Stable"
	req.Status = database.StatusComplete
	req.ExternalID = "EXT-1"

	dbPath := createTestDatabase(t, []*database.Requirement{req})
	cfg := createTestConfig(dbPath)
	adapter := &mockAdapter{
		name:      "test-service",
		connected: true,
		items: []adapters.ExternalItem{
			{ExternalID: "EXT-1", Title: "Same", Status: "closed"},
		},
		statusMapping: map[string]database.Status{"closed": database.StatusComplete},
		updateResult:  true,
	}

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	_ = runBidirectional(adapter, cfg, "prefer-remote", false)
	second := runBidirectional(adapter, cfg, "prefer-remote", false)
	_ = w.Close()
	os.Stdout = old

	if len(second.Updated) != 0 {
		t.Fatalf("second run updated = %v, want none", second.Updated)
	}
	if len(second.Skipped) != 1 {
		t.Fatalf("second skipped = %v", second.Skipped)
	}
}

func TestBidiDryRunNoWrites(t *testing.T) {
	rtmx.Req(t, "REQ-SYNC-001c")

	req := database.NewRequirement("REQ-TEST-006")
	req.Category = "TEST"
	req.RequirementText = "Dry"
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
		updateResult:  true,
		createResult:  "EXT-SHOULD-NOT",
	}

	old := os.Stdout
	_, w, _ := os.Pipe()
	os.Stdout = w
	_ = runBidirectional(adapter, cfg, "prefer-remote", true)
	_ = w.Close()
	os.Stdout = old

	after, _ := os.ReadFile(dbPath)
	if string(before) != string(after) {
		t.Fatal("dry-run mutated disk")
	}
	if adapter.updateCalls != 0 || adapter.createCalls != 0 {
		t.Fatalf("dry-run must not call adapter writes (update=%d create=%d)", adapter.updateCalls, adapter.createCalls)
	}
}
