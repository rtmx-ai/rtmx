package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestBacklogOmitsRationaleIncludesACGaps(t *testing.T) {
	rtmx.Req(t, "REQ-DATA-006")
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "database.csv")
	db := database.NewDatabase()
	if err := db.Add(&database.Requirement{
		ReqID:           "REQ-AUTH-001",
		Category:        "AUTH",
		RequirementText: "Users shall authenticate.",
		Status:          database.StatusMissing,
		Priority:        database.PriorityHigh,
	}); err != nil {
		t.Fatal(err)
	}
	if err := db.Save(dbPath); err != nil {
		t.Fatal(err)
	}

	docPath := filepath.Join(dir, "doc.json")
	doc := `{
  "schema_version": "requirement-document/v0",
  "requirements": [{
    "req_id": "REQ-AUTH-001",
    "requirement_text": "Users shall authenticate.",
    "status": "MISSING",
    "priority": "HIGH",
    "narrative": "Bulk rationale agents must not see",
    "acs": [
      {"ac_id": "AC-1", "statement": "Reject missing token.", "required": true},
      {"ac_id": "AC-2", "statement": "Accept bearer.", "required": false}
    ],
    "test_bindings": [
      {"binding_id": "TB-1", "ac_id": "AC-1", "test_function": "TestRoomRejectsMissingToken"}
    ]
  }]
}`
	if err := os.WriteFile(docPath, []byte(doc), 0o644); err != nil {
		t.Fatal(err)
	}

	srv := NewServer(dbPath, &config.Config{}, WithACDocument(docPath))
	if srv.acDocument == nil {
		t.Fatal("expected AC document loaded")
	}
	loaded, err := database.Load(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(srv.toolBacklog(loaded, toolFilter{}))
	if err != nil {
		t.Fatal(err)
	}
	text := string(raw)
	if strings.Contains(text, "Bulk rationale") || strings.Contains(text, "narrative") {
		t.Fatalf("payload leaked rationale: %s", text)
	}
	if !strings.Contains(text, `"ac_id":"AC-1"`) || !strings.Contains(text, "TB-1") || !strings.Contains(text, "ac_gap_count") {
		t.Fatalf("missing AC gap surface: %s", text)
	}
}
