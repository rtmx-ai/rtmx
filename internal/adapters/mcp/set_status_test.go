package mcp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestMCPSetStatus validates the set_status mutation tool: it sets valid
// statuses with provenance, requires agent identity, rejects unknown reqs, and
// REFUSES to set COMPLETE (completion is verify-driven).
// REQ-MCP-011: MCP set_status tool (status writeback, not COMPLETE).
func TestMCPSetStatus(t *testing.T) {
	rtmx.Req(t, "REQ-MCP-011")

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, ".rtmx", "database.csv")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestDB(t, dbPath)
	writeTestConfig(t, filepath.Join(tmpDir, "rtmx.yaml"))
	cfg, err := config.LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	srv := NewServer(dbPath, cfg)

	load := func() *database.Database {
		db, err := database.Load(dbPath)
		if err != nil {
			t.Fatalf("load db: %v", err)
		}
		return db
	}

	// Happy path: PARTIAL -> NOT_STARTED, persisted.
	res, rpcErr := srv.toolSetStatus(load(), map[string]interface{}{
		"req_id": "REQ-TEST-002", "status": "NOT_STARTED", "agent_id": "tester", "reason": "reopened",
	})
	if rpcErr != nil {
		t.Fatalf("unexpected rpc error: %v", rpcErr)
	}
	if tr, ok := res.(toolResult); !ok || tr.IsError {
		t.Fatalf("expected success result, got %+v", res)
	}
	if got := load().Get("REQ-TEST-002").Status; got != database.StatusNotStarted {
		t.Errorf("status not persisted: want NOT_STARTED, got %s", got)
	}

	// COMPLETE is refused (verify owns closure).
	res, _ = srv.toolSetStatus(load(), map[string]interface{}{
		"req_id": "REQ-TEST-002", "status": "COMPLETE", "agent_id": "tester",
	})
	if tr, ok := res.(toolResult); !ok || !tr.IsError {
		t.Error("set_status must refuse COMPLETE")
	}

	// Missing agent_id is rejected.
	res, _ = srv.toolSetStatus(load(), map[string]interface{}{"req_id": "REQ-TEST-002", "status": "MISSING"})
	if tr, ok := res.(toolResult); !ok || !tr.IsError {
		t.Error("set_status must require agent_id")
	}

	// Unknown requirement is an error.
	res, _ = srv.toolSetStatus(load(), map[string]interface{}{
		"req_id": "REQ-NOPE-999", "status": "MISSING", "agent_id": "tester",
	})
	if tr, ok := res.(toolResult); !ok || !tr.IsError {
		t.Error("set_status must error on an unknown requirement")
	}
}
