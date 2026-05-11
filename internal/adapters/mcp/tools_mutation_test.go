package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/orchestration"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestMCPMutationTools validates the mutation tools (claim, release,
// release_assign) require agent identity and work correctly.
// REQ-MCP-005: MCP mutation tools behind authorization model.
func TestMCPMutationTools(t *testing.T) {
	rtmx.Req(t, "REQ-MCP-005")

	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0o755)

	dbPath := filepath.Join(rtmxDir, "database.csv")
	writeTestDB(t, dbPath)
	writeTestConfig(t, filepath.Join(tmpDir, "rtmx.yaml"))

	cfg, err := config.LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	srv := NewServer(dbPath, cfg, WithHost("127.0.0.1"), WithPort(0))
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			_ = err
		}
	}()

	deadline := time.Now().Add(2 * time.Second)
	for srv.Addr() == "" && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if srv.Addr() == "" {
		t.Fatal("server did not start")
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	baseURL := fmt.Sprintf("http://%s/mcp", srv.Addr())

	t.Run("claim_and_release", func(t *testing.T) {
		// Claim
		resp := rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name":      "claim",
			"arguments": map[string]interface{}{"req_id": "REQ-TEST-002", "agent_id": "claude-1"},
		})
		text := extractToolText(t, resp)

		var claim orchestration.Claim
		if err := json.Unmarshal([]byte(text), &claim); err != nil {
			t.Fatalf("failed to parse claim: %v\nText: %s", err, text)
		}
		if claim.ReqID != "REQ-TEST-002" {
			t.Errorf("claim ReqID = %q, want REQ-TEST-002", claim.ReqID)
		}
		if claim.AgentID != "claude-1" {
			t.Errorf("claim AgentID = %q, want claude-1", claim.AgentID)
		}

		// Release
		resp = rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name":      "release",
			"arguments": map[string]interface{}{"req_id": "REQ-TEST-002", "agent_id": "claude-1"},
		})
		text = extractToolText(t, resp)
		if text == "" {
			t.Error("release should return result text")
		}
	})

	t.Run("claim_double_fails", func(t *testing.T) {
		// Claim first
		rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name":      "claim",
			"arguments": map[string]interface{}{"req_id": "REQ-TEST-003", "agent_id": "claude-1"},
		})

		// Double claim should return error result
		resp := rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name":      "claim",
			"arguments": map[string]interface{}{"req_id": "REQ-TEST-003", "agent_id": "claude-2"},
		})
		result, ok := resp["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected result, got %T", resp["result"])
		}
		isError, _ := result["isError"].(bool)
		if !isError {
			t.Error("double claim should return isError=true")
		}

		// Cleanup
		_ = srv.claims.ForceRelease("REQ-TEST-003")
	})

	t.Run("release_wrong_owner_fails", func(t *testing.T) {
		rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name":      "claim",
			"arguments": map[string]interface{}{"req_id": "REQ-TEST-001", "agent_id": "claude-1"},
		})

		resp := rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name":      "release",
			"arguments": map[string]interface{}{"req_id": "REQ-TEST-001", "agent_id": "claude-2"},
		})
		result, ok := resp["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected result, got %T", resp["result"])
		}
		isError, _ := result["isError"].(bool)
		if !isError {
			t.Error("release by non-owner should return isError=true")
		}

		// Cleanup
		_ = srv.claims.ForceRelease("REQ-TEST-001")
	})

	t.Run("claim_missing_agent_id_fails", func(t *testing.T) {
		resp := rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name":      "claim",
			"arguments": map[string]interface{}{"req_id": "REQ-TEST-001"},
		})
		result, ok := resp["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected result, got %T", resp["result"])
		}
		isError, _ := result["isError"].(bool)
		if !isError {
			t.Error("claim without agent_id should return isError=true")
		}
	})

	t.Run("release_assign_sets_version", func(t *testing.T) {
		resp := rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name": "release_assign",
			"arguments": map[string]interface{}{
				"version":  "v0.5.0",
				"req_ids":  []interface{}{"REQ-TEST-002", "REQ-TEST-003"},
				"agent_id": "claude-1",
			},
		})
		text := extractToolText(t, resp)

		var result map[string]interface{}
		if err := json.Unmarshal([]byte(text), &result); err != nil {
			t.Fatalf("failed to parse result: %v", err)
		}
		assigned, _ := result["assigned"].([]interface{})
		if len(assigned) != 2 {
			t.Errorf("expected 2 assigned, got %d", len(assigned))
		}
		version, _ := result["version"].(string)
		if version != "v0.5.0" {
			t.Errorf("version = %q, want v0.5.0", version)
		}
	})

	t.Run("release_assign_unknown_req", func(t *testing.T) {
		resp := rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name": "release_assign",
			"arguments": map[string]interface{}{
				"version":  "v0.5.0",
				"req_ids":  []interface{}{"REQ-NONEXISTENT"},
				"agent_id": "claude-1",
			},
		})
		text := extractToolText(t, resp)

		var result map[string]interface{}
		if err := json.Unmarshal([]byte(text), &result); err != nil {
			t.Fatalf("failed to parse result: %v", err)
		}
		errs, _ := result["errors"].([]interface{})
		if len(errs) != 1 {
			t.Errorf("expected 1 error for unknown req, got %d", len(errs))
		}
	})
}
