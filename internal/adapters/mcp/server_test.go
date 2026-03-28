package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestMCPServer validates the MCP server implementation.
// REQ-GO-039: Go CLI shall implement MCP server for AI agent integration.
func TestMCPServer(t *testing.T) {
	rtmx.Req(t, "REQ-GO-039")

	// Create temp project with a small RTM database
	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0o755); err != nil {
		t.Fatalf("failed to create .rtmx dir: %v", err)
	}

	dbPath := filepath.Join(rtmxDir, "database.csv")
	writeTestDB(t, dbPath)

	cfgPath := filepath.Join(tmpDir, "rtmx.yaml")
	writeTestConfig(t, cfgPath)

	cfg, err := config.LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Start server on random port
	srv := NewServer(dbPath, cfg, WithHost("127.0.0.1"), WithPort(0))

	// Use port 0 so the OS picks an available port
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			// Server shut down -- expected during test cleanup
		}
	}()

	// Wait for server to be ready
	deadline := time.Now().Add(2 * time.Second)
	for srv.Addr() == "" && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	if srv.Addr() == "" {
		t.Fatal("server did not start in time")
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	baseURL := fmt.Sprintf("http://%s/mcp", srv.Addr())

	t.Run("initialize", func(t *testing.T) {
		resp := rpcCall(t, baseURL, "initialize", nil)
		result, ok := resp["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected result object, got %T", resp["result"])
		}
		if result["protocolVersion"] != "2024-11-05" {
			t.Errorf("unexpected protocol version: %v", result["protocolVersion"])
		}
		info, _ := result["serverInfo"].(map[string]interface{})
		if info["name"] != "rtmx" {
			t.Errorf("expected server name 'rtmx', got %v", info["name"])
		}
	})

	t.Run("tools_list", func(t *testing.T) {
		resp := rpcCall(t, baseURL, "tools/list", nil)
		result, ok := resp["result"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected result object, got %T", resp["result"])
		}
		tools, ok := result["tools"].([]interface{})
		if !ok {
			t.Fatalf("expected tools array, got %T", result["tools"])
		}
		// We expose 6 tools
		if len(tools) != 6 {
			t.Errorf("expected 6 tools, got %d", len(tools))
		}
		// Verify tool names
		names := make(map[string]bool)
		for _, tool := range tools {
			tm, _ := tool.(map[string]interface{})
			name, _ := tm["name"].(string)
			names[name] = true
		}
		for _, expected := range []string{"status", "backlog", "health", "deps", "verify", "markers"} {
			if !names[expected] {
				t.Errorf("missing tool: %s", expected)
			}
		}
	})

	t.Run("tool_status", func(t *testing.T) {
		resp := rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name": "status",
		})
		text := extractToolText(t, resp)

		var status statusResult
		if err := json.Unmarshal([]byte(text), &status); err != nil {
			t.Fatalf("failed to parse status JSON: %v", err)
		}
		if status.Total != 3 {
			t.Errorf("expected 3 total, got %d", status.Total)
		}
		if status.Complete != 1 {
			t.Errorf("expected 1 complete, got %d", status.Complete)
		}
		if status.CompletionPct <= 0 {
			t.Error("expected positive completion percentage")
		}
	})

	t.Run("tool_backlog", func(t *testing.T) {
		resp := rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name": "backlog",
		})
		text := extractToolText(t, resp)

		var bl backlogResult
		if err := json.Unmarshal([]byte(text), &bl); err != nil {
			t.Fatalf("failed to parse backlog JSON: %v", err)
		}
		if bl.TotalIncomplete != 2 {
			t.Errorf("expected 2 incomplete, got %d", bl.TotalIncomplete)
		}
	})

	t.Run("tool_health", func(t *testing.T) {
		resp := rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name": "health",
		})
		text := extractToolText(t, resp)

		var h map[string]interface{}
		if err := json.Unmarshal([]byte(text), &h); err != nil {
			t.Fatalf("failed to parse health JSON: %v", err)
		}
		status, _ := h["status"].(string)
		if status == "" {
			t.Error("expected non-empty health status")
		}
		checks, _ := h["checks"].([]interface{})
		if len(checks) == 0 {
			t.Error("expected at least one health check")
		}
	})

	t.Run("tool_deps_overview", func(t *testing.T) {
		resp := rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name": "deps",
		})
		text := extractToolText(t, resp)

		var d depsResult
		if err := json.Unmarshal([]byte(text), &d); err != nil {
			t.Fatalf("failed to parse deps JSON: %v", err)
		}
		if len(d.Overview) != 3 {
			t.Errorf("expected 3 overview entries, got %d", len(d.Overview))
		}
	})

	t.Run("tool_deps_specific", func(t *testing.T) {
		resp := rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name":      "deps",
			"arguments": map[string]interface{}{"req_id": "REQ-TEST-001"},
		})
		text := extractToolText(t, resp)

		var d depsResult
		if err := json.Unmarshal([]byte(text), &d); err != nil {
			t.Fatalf("failed to parse deps JSON: %v", err)
		}
		if d.ReqID != "REQ-TEST-001" {
			t.Errorf("expected req_id REQ-TEST-001, got %s", d.ReqID)
		}
	})

	t.Run("tool_deps_not_found", func(t *testing.T) {
		resp := rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name":      "deps",
			"arguments": map[string]interface{}{"req_id": "REQ-NONEXISTENT"},
		})
		// Should be an error result, not an RPC error
		result := resp["result"]
		rm, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("expected result map, got %T", result)
		}
		isError, _ := rm["isError"].(bool)
		if !isError {
			t.Error("expected isError=true for nonexistent requirement")
		}
	})

	t.Run("tool_verify", func(t *testing.T) {
		resp := rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name": "verify",
		})
		text := extractToolText(t, resp)

		var v verifyResult
		if err := json.Unmarshal([]byte(text), &v); err != nil {
			t.Fatalf("failed to parse verify JSON: %v", err)
		}
		if v.Total != 3 {
			t.Errorf("expected 3 total, got %d", v.Total)
		}
	})

	t.Run("tool_markers", func(t *testing.T) {
		resp := rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name": "markers",
		})
		text := extractToolText(t, resp)

		var m markersResult
		if err := json.Unmarshal([]byte(text), &m); err != nil {
			t.Fatalf("failed to parse markers JSON: %v", err)
		}
		if m.Total != 3 {
			t.Errorf("expected 3 total, got %d", m.Total)
		}
		if m.WithTests+m.Missing != m.Total {
			t.Errorf("with_tests + missing should equal total")
		}
	})

	t.Run("unknown_tool", func(t *testing.T) {
		resp := rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name": "nonexistent",
		})
		errObj, ok := resp["error"].(map[string]interface{})
		if !ok {
			t.Fatal("expected RPC error for unknown tool")
		}
		code, _ := errObj["code"].(float64)
		if int(code) != errNoMethod {
			t.Errorf("expected error code %d, got %v", errNoMethod, code)
		}
	})

	t.Run("unknown_method", func(t *testing.T) {
		resp := rpcCall(t, baseURL, "unknown/method", nil)
		errObj, ok := resp["error"].(map[string]interface{})
		if !ok {
			t.Fatal("expected RPC error for unknown method")
		}
		code, _ := errObj["code"].(float64)
		if int(code) != errNoMethod {
			t.Errorf("expected error code %d, got %v", errNoMethod, code)
		}
	})

	t.Run("method_not_allowed", func(t *testing.T) {
		resp, err := http.Get(baseURL)
		if err != nil {
			t.Fatalf("GET request failed: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected 405, got %d", resp.StatusCode)
		}
	})
}

// TestMCPServerPort0 verifies that port 0 works (OS-assigned port).
func TestMCPServerPort0(t *testing.T) {
	rtmx.Req(t, "REQ-GO-039")

	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0o755)

	dbPath := filepath.Join(rtmxDir, "database.csv")
	writeTestDB(t, dbPath)

	cfgPath := filepath.Join(tmpDir, "rtmx.yaml")
	writeTestConfig(t, cfgPath)

	cfg, _ := config.LoadFromDir(tmpDir)

	srv := NewServer(dbPath, cfg, WithHost("127.0.0.1"), WithPort(0))

	go func() { _ = srv.Start() }()

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

	// Verify it answers
	baseURL := fmt.Sprintf("http://%s/mcp", srv.Addr())
	resp := rpcCall(t, baseURL, "initialize", nil)
	if resp["result"] == nil {
		t.Error("expected initialize result")
	}
}

// ----- helpers -----

func writeTestDB(t *testing.T, path string) {
	t.Helper()

	db := database.NewDatabase()

	r1 := &database.Requirement{
		ReqID:           "REQ-TEST-001",
		Category:        "CORE",
		RequirementText: "First requirement",
		Status:          database.StatusComplete,
		Priority:        database.PriorityHigh,
		Phase:           1,
		TestFunction:    "TestFirst",
	}
	r2 := &database.Requirement{
		ReqID:           "REQ-TEST-002",
		Category:        "CORE",
		RequirementText: "Second requirement",
		Status:          database.StatusPartial,
		Priority:        database.PriorityHigh,
		Phase:           1,
		Dependencies:    database.NewStringSet("REQ-TEST-001"),
	}
	r3 := &database.Requirement{
		ReqID:           "REQ-TEST-003",
		Category:        "EXT",
		RequirementText: "Third requirement",
		Status:          database.StatusMissing,
		Priority:        database.PriorityMedium,
		Phase:           2,
	}

	for _, r := range []*database.Requirement{r1, r2, r3} {
		if err := db.Add(r); err != nil {
			t.Fatalf("failed to add requirement: %v", err)
		}
	}

	if err := db.Save(path); err != nil {
		t.Fatalf("failed to save test database: %v", err)
	}
}

func writeTestConfig(t *testing.T, path string) {
	t.Helper()
	content := `rtmx:
  database: .rtmx/database.csv
  phases:
    1: "Foundation"
    2: "Extensions"
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
}

func rpcCall(t *testing.T, url, method string, params interface{}) map[string]interface{} {
	t.Helper()

	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
	}
	if params != nil {
		body["params"] = params
	}

	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST %s failed: %v", url, err)
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	return result
}

func extractToolText(t *testing.T, resp map[string]interface{}) string {
	t.Helper()

	result, ok := resp["result"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result object, got %T: %v", resp["result"], resp)
	}

	content, ok := result["content"].([]interface{})
	if !ok || len(content) == 0 {
		t.Fatalf("expected content array with entries, got %v", result["content"])
	}

	first, ok := content[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected content entry to be object, got %T", content[0])
	}

	text, ok := first["text"].(string)
	if !ok {
		t.Fatalf("expected text string, got %T", first["text"])
	}

	return text
}
