package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestStoreRoundTripAndProjection(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-004")
	dir := t.TempDir()
	compDir := filepath.Join(dir, "companions")
	if err := os.MkdirAll(compDir, 0o755); err != nil {
		t.Fatal(err)
	}
	rationale := "Bulk rationale that agents must not see by default."
	if err := os.WriteFile(filepath.Join(compDir, "REQ-AUTH-001.md"), []byte("# REQ-AUTH-001\n\n"+rationale+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	reqd := true
	s := Store{Dir: dir}
	rec := Record{
		ReqID:           "REQ-AUTH-001",
		RequirementText: "Users shall authenticate.",
		Status:          "PARTIAL",
		Priority:        "HIGH",
		RequirementFile: "companions/REQ-AUTH-001.md",
		Narrative:       rationale,
		ACs:             []AC{{ID: "AC-1", Statement: "Reject missing token.", Required: &reqd}},
		TestBindings:    []Binding{{ID: "TB-1", ACID: "AC-1", TestFunction: "TestRoomRejectsMissingToken"}},
	}
	if err := s.Put(rec); err != nil {
		t.Fatal(err)
	}
	loaded, err := s.Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 1 || loaded[0].ReqID != "REQ-AUTH-001" || loaded[0].ACs[0].ID != "AC-1" {
		t.Fatalf("round-trip: %+v", loaded)
	}
	body, err := s.ReadCompanion(loaded[0])
	if err != nil || !strings.Contains(body, rationale) {
		t.Fatalf("companion: %q err=%v", body, err)
	}
	proj := Project(loaded[0])
	raw, err := json.Marshal(proj)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), rationale) || strings.Contains(string(raw), "narrative") {
		t.Fatalf("projection leaked rationale: %s", raw)
	}
	if proj.ACs[0].ID != "AC-1" || len(proj.TestBindings) != 1 {
		t.Fatalf("projection missing AC/binding: %+v", proj)
	}

	// Default CSV path is not created by the store.
	if _, err := os.Stat(filepath.Join(dir, "database.csv")); !os.IsNotExist(err) {
		t.Fatal("store must not write default CSV")
	}
}
