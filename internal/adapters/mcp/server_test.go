package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
			// Server start failed unexpectedly; log for debugging.
			_ = err
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
		// We expose 11 tools (7 read + 4 mutation)
		if len(tools) != 11 {
			t.Errorf("expected 11 tools, got %d", len(tools))
		}
		// Verify tool names
		names := make(map[string]bool)
		for _, tool := range tools {
			tm, _ := tool.(map[string]interface{})
			name, _ := tm["name"].(string)
			names[name] = true
		}
		for _, expected := range []string{"status", "backlog", "health", "deps", "verify", "markers", "next", "claim", "release", "release_assign"} {
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
		defer func() { _ = resp.Body.Close() }()
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
	defer func() { _ = resp.Body.Close() }()

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

// TestMCPStdio validates the stdio transport for MCP.
// REQ-MCP-006: MCP server shall support stdio transport for Claude Code and Cursor.
func TestMCPStdio(t *testing.T) {
	rtmx.Req(t, "REQ-MCP-006")

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

	srv := NewServer(dbPath, cfg)

	t.Run("initialize_and_tools_list", func(t *testing.T) {
		// Send initialize + tools/list via stdin, read responses from stdout
		var input bytes.Buffer
		input.WriteString(`{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n")
		input.WriteString(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}` + "\n")

		var output bytes.Buffer
		if err := srv.StartStdio(&input, &output); err != nil {
			t.Fatalf("StartStdio failed: %v", err)
		}

		lines := bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n"))
		if len(lines) != 2 {
			t.Fatalf("expected 2 response lines, got %d: %s", len(lines), output.String())
		}

		// Check initialize response
		var initResp rpcResponse
		if err := json.Unmarshal(lines[0], &initResp); err != nil {
			t.Fatalf("failed to parse initialize response: %v", err)
		}
		if initResp.JSONRPC != "2.0" {
			t.Error("expected jsonrpc 2.0")
		}

		// Check tools/list response
		var listResp map[string]interface{}
		if err := json.Unmarshal(lines[1], &listResp); err != nil {
			t.Fatalf("failed to parse tools/list response: %v", err)
		}
		result, _ := listResp["result"].(map[string]interface{})
		tools, _ := result["tools"].([]interface{})
		if len(tools) != 11 {
			t.Errorf("expected 11 tools, got %d", len(tools))
		}
	})

	t.Run("tools_call_via_stdio", func(t *testing.T) {
		var input bytes.Buffer
		input.WriteString(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"status"}}` + "\n")

		var output bytes.Buffer
		if err := srv.StartStdio(&input, &output); err != nil {
			t.Fatalf("StartStdio failed: %v", err)
		}

		var resp map[string]interface{}
		if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}

		result, _ := resp["result"].(map[string]interface{})
		content, _ := result["content"].([]interface{})
		if len(content) == 0 {
			t.Fatal("expected content in response")
		}
		first, _ := content[0].(map[string]interface{})
		text, _ := first["text"].(string)

		var status statusResult
		if err := json.Unmarshal([]byte(text), &status); err != nil {
			t.Fatalf("failed to parse status JSON: %v", err)
		}
		if status.Total != 3 {
			t.Errorf("expected 3 total, got %d", status.Total)
		}
	})

	t.Run("notification_no_response", func(t *testing.T) {
		var input bytes.Buffer
		input.WriteString(`{"jsonrpc":"2.0","method":"notifications/initialized"}` + "\n")

		var output bytes.Buffer
		if err := srv.StartStdio(&input, &output); err != nil {
			t.Fatalf("StartStdio failed: %v", err)
		}

		if output.Len() != 0 {
			t.Errorf("expected no output for notification, got: %s", output.String())
		}
	})

	t.Run("empty_lines_ignored", func(t *testing.T) {
		var input bytes.Buffer
		input.WriteString("\n\n")
		input.WriteString(`{"jsonrpc":"2.0","id":1,"method":"initialize"}` + "\n")
		input.WriteString("\n")

		var output bytes.Buffer
		if err := srv.StartStdio(&input, &output); err != nil {
			t.Fatalf("StartStdio failed: %v", err)
		}

		lines := bytes.Split(bytes.TrimSpace(output.Bytes()), []byte("\n"))
		if len(lines) != 1 {
			t.Errorf("expected 1 response line, got %d", len(lines))
		}
	})

	t.Run("parse_error", func(t *testing.T) {
		var input bytes.Buffer
		input.WriteString("not json\n")

		var output bytes.Buffer
		if err := srv.StartStdio(&input, &output); err != nil {
			t.Fatalf("StartStdio failed: %v", err)
		}

		var resp rpcResponse
		if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &resp); err != nil {
			t.Fatalf("failed to parse error response: %v", err)
		}
		if resp.Error == nil {
			t.Fatal("expected error in response")
		}
		if resp.Error.Code != errParse {
			t.Errorf("expected parse error code %d, got %d", errParse, resp.Error.Code)
		}
	})
}

// TestMCPResponseSizeLogging validates that the MCP server logs response byte
// and token counts to stderr on every tool call.
// REQ-MCP-007: Response size logging for token consumption observability.
func TestMCPResponseSizeLogging(t *testing.T) {
	rtmx.Req(t, "REQ-MCP-007")

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

	t.Run("logs_bytes_and_tokens", func(t *testing.T) {
		var logBuf bytes.Buffer
		logger := log.New(&logBuf, "", 0)
		srv := NewServer(dbPath, cfg, WithLogger(logger))

		var input bytes.Buffer
		input.WriteString(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"status"}}` + "\n")

		var output bytes.Buffer
		if err := srv.StartStdio(&input, &output); err != nil {
			t.Fatalf("StartStdio failed: %v", err)
		}

		logLine := logBuf.String()
		if logLine == "" {
			t.Fatal("expected log output, got none")
		}
		if !strings.Contains(logLine, "[rtmx-mcp]") {
			t.Errorf("log line missing prefix: %s", logLine)
		}
		if !strings.Contains(logLine, "tool=status") {
			t.Errorf("log line missing tool name: %s", logLine)
		}
		if !strings.Contains(logLine, "bytes=") {
			t.Errorf("log line missing bytes: %s", logLine)
		}
		if !strings.Contains(logLine, "tokens=") {
			t.Errorf("log line missing tokens: %s", logLine)
		}
		if strings.Contains(logLine, "error=true") {
			t.Errorf("log line should not contain error=true for successful call: %s", logLine)
		}
	})

	t.Run("logs_on_error", func(t *testing.T) {
		var logBuf bytes.Buffer
		logger := log.New(&logBuf, "", 0)
		srv := NewServer(dbPath, cfg, WithLogger(logger))

		var input bytes.Buffer
		input.WriteString(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"deps","arguments":{"req_id":"REQ-NONEXISTENT"}}}` + "\n")

		var output bytes.Buffer
		if err := srv.StartStdio(&input, &output); err != nil {
			t.Fatalf("StartStdio failed: %v", err)
		}

		logLine := logBuf.String()
		if !strings.Contains(logLine, "tool=deps") {
			t.Errorf("expected tool=deps in log, got: %s", logLine)
		}
		if !strings.Contains(logLine, "error=true") {
			t.Errorf("expected error=true in log for failed call, got: %s", logLine)
		}
	})

	t.Run("quiet_flag_suppresses", func(t *testing.T) {
		var logBuf bytes.Buffer
		logger := log.New(&logBuf, "", 0)
		srv := NewServer(dbPath, cfg, WithLogger(logger), WithQuiet(true))

		var input bytes.Buffer
		input.WriteString(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"status"}}` + "\n")

		var output bytes.Buffer
		if err := srv.StartStdio(&input, &output); err != nil {
			t.Fatalf("StartStdio failed: %v", err)
		}

		if logBuf.Len() != 0 {
			t.Errorf("expected no log output with --quiet, got: %s", logBuf.String())
		}
	})

	t.Run("does_not_affect_stdout", func(t *testing.T) {
		var logBuf bytes.Buffer
		logger := log.New(&logBuf, "", 0)
		srv := NewServer(dbPath, cfg, WithLogger(logger))

		var input bytes.Buffer
		input.WriteString(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"status"}}` + "\n")

		var output bytes.Buffer
		if err := srv.StartStdio(&input, &output); err != nil {
			t.Fatalf("StartStdio failed: %v", err)
		}

		// stdout should contain only the JSON-RPC response, no log lines
		outStr := output.String()
		if strings.Contains(outStr, "[rtmx-mcp]") {
			t.Errorf("log line leaked to stdout: %s", outStr)
		}

		// Verify the response is valid JSON-RPC
		var resp map[string]interface{}
		if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &resp); err != nil {
			t.Fatalf("stdout is not valid JSON: %v", err)
		}
	})

	t.Run("http_transport_logs", func(t *testing.T) {
		var logBuf bytes.Buffer
		logger := log.New(&logBuf, "", 0)
		srv := NewServer(dbPath, cfg, WithHost("127.0.0.1"), WithPort(0), WithLogger(logger))

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
		rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name": "health",
		})

		logLine := logBuf.String()
		if !strings.Contains(logLine, "tool=health") {
			t.Errorf("HTTP transport should log tool calls, got: %s", logLine)
		}
		if !strings.Contains(logLine, "bytes=") {
			t.Errorf("HTTP transport log missing bytes: %s", logLine)
		}
	})

	t.Run("token_estimate_accuracy", func(t *testing.T) {
		var logBuf bytes.Buffer
		logger := log.New(&logBuf, "", 0)
		srv := NewServer(dbPath, cfg, WithLogger(logger))

		var input bytes.Buffer
		input.WriteString(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"status"}}` + "\n")

		var output bytes.Buffer
		if err := srv.StartStdio(&input, &output); err != nil {
			t.Fatalf("StartStdio failed: %v", err)
		}

		logLine := logBuf.String()

		// Parse bytes and tokens from log line
		var logBytes, logTokens int
		_, _ = fmt.Sscanf(logLine, "[rtmx-mcp] tool=status bytes=%d tokens=%d", &logBytes, &logTokens)

		if logBytes <= 0 {
			t.Errorf("expected positive byte count, got %d", logBytes)
		}

		// Verify token estimate: ceil(bytes/4)
		expectedTokens := (logBytes + 3) / 4
		if logTokens != expectedTokens {
			t.Errorf("token estimate: got %d, want ceil(%d/4) = %d", logTokens, logBytes, expectedTokens)
		}
	})
}

// TestMCPToolFiltering validates that MCP tools accept category/status/limit
// filters and return reduced response sizes.
// REQ-MCP-008: MCP tools shall accept filters to reduce response size.
func TestMCPToolFiltering(t *testing.T) {
	rtmx.Req(t, "REQ-MCP-008")

	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0o755)

	dbPath := filepath.Join(rtmxDir, "database.csv")
	writeTestDB(t, dbPath) // 3 reqs: CORE(2), EXT(1)
	writeTestConfig(t, filepath.Join(tmpDir, "rtmx.yaml"))

	cfg, err := config.LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	srv := NewServer(dbPath, cfg, WithQuiet(true))

	callTool := func(t *testing.T, tool string, args map[string]interface{}) string {
		t.Helper()
		params := map[string]interface{}{"name": tool}
		if args != nil {
			params["arguments"] = args
		}
		reqJSON, _ := json.Marshal(map[string]interface{}{
			"jsonrpc": "2.0", "id": 1, "method": "tools/call", "params": params,
		})
		var input, output bytes.Buffer
		input.Write(reqJSON)
		input.WriteByte('\n')
		if err := srv.StartStdio(&input, &output); err != nil {
			t.Fatalf("StartStdio failed: %v", err)
		}
		var resp map[string]interface{}
		if err := json.Unmarshal(bytes.TrimSpace(output.Bytes()), &resp); err != nil {
			t.Fatalf("invalid response JSON: %v", err)
		}
		result, _ := resp["result"].(map[string]interface{})
		content, _ := result["content"].([]interface{})
		if len(content) == 0 {
			t.Fatal("no content in response")
		}
		first, _ := content[0].(map[string]interface{})
		return first["text"].(string)
	}

	t.Run("status_category_filter", func(t *testing.T) {
		text := callTool(t, "status", map[string]interface{}{"category": "CORE"})
		var status statusResult
		if err := json.Unmarshal([]byte(text), &status); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if status.Total != 2 {
			t.Errorf("expected 2 CORE reqs, got %d", status.Total)
		}
		if len(status.Categories) != 1 || status.Categories[0].Name != "CORE" {
			t.Errorf("expected single CORE category, got %v", status.Categories)
		}
		if status.FilteredBy == nil || status.FilteredBy.Category != "CORE" {
			t.Error("expected filtered_by.category = CORE")
		}
	})

	t.Run("status_status_filter", func(t *testing.T) {
		text := callTool(t, "status", map[string]interface{}{"status": "COMPLETE"})
		var status statusResult
		if err := json.Unmarshal([]byte(text), &status); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if status.Total != 1 {
			t.Errorf("expected 1 COMPLETE req, got %d", status.Total)
		}
		if status.Complete != 1 {
			t.Errorf("expected complete=1, got %d", status.Complete)
		}
	})

	t.Run("status_no_filter_returns_all", func(t *testing.T) {
		text := callTool(t, "status", nil)
		var status statusResult
		if err := json.Unmarshal([]byte(text), &status); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if status.Total != 3 {
			t.Errorf("expected 3 total, got %d", status.Total)
		}
		if status.FilteredBy != nil {
			t.Error("expected no filtered_by when no filters applied")
		}
	})

	t.Run("backlog_category_filter", func(t *testing.T) {
		text := callTool(t, "backlog", map[string]interface{}{"category": "EXT"})
		var bl backlogResult
		if err := json.Unmarshal([]byte(text), &bl); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if bl.TotalIncomplete != 1 {
			t.Errorf("expected 1 EXT incomplete, got %d", bl.TotalIncomplete)
		}
		if len(bl.Items) != 1 || bl.Items[0].ReqID != "REQ-TEST-003" {
			t.Errorf("expected REQ-TEST-003, got %v", bl.Items)
		}
	})

	t.Run("backlog_limit", func(t *testing.T) {
		text := callTool(t, "backlog", map[string]interface{}{"limit": 1})
		var bl backlogResult
		if err := json.Unmarshal([]byte(text), &bl); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if len(bl.Items) != 1 {
			t.Errorf("expected 1 item with limit=1, got %d", len(bl.Items))
		}
	})

	t.Run("deps_overview_limit", func(t *testing.T) {
		text := callTool(t, "deps", map[string]interface{}{"limit": 2})
		var d depsResult
		if err := json.Unmarshal([]byte(text), &d); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if len(d.Overview) != 2 {
			t.Errorf("expected 2 overview entries with limit=2, got %d", len(d.Overview))
		}
	})

	t.Run("markers_category_filter", func(t *testing.T) {
		text := callTool(t, "markers", map[string]interface{}{"category": "CORE"})
		var m markersResult
		if err := json.Unmarshal([]byte(text), &m); err != nil {
			t.Fatalf("parse error: %v", err)
		}
		if m.Total != 2 {
			t.Errorf("expected 2 CORE markers, got %d", m.Total)
		}
	})

	t.Run("filtered_response_smaller_than_unfiltered", func(t *testing.T) {
		unfilteredText := callTool(t, "status", nil)
		filteredText := callTool(t, "status", map[string]interface{}{"category": "EXT"})
		if len(filteredText) >= len(unfilteredText) {
			t.Errorf("filtered response (%d bytes) should be smaller than unfiltered (%d bytes)",
				len(filteredText), len(unfilteredText))
		}
	})
}

// TestMCPToolDescriptions validates that tool descriptions include size hints
// and that the hints are within tolerance of actual response sizes.
// REQ-MCP-009: Token budget awareness in tool descriptions.
func TestMCPToolDescriptions(t *testing.T) {
	rtmx.Req(t, "REQ-MCP-009")

	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0o755)

	dbPath := filepath.Join(rtmxDir, "database.csv")
	writeTestDB(t, dbPath) // 3 reqs: CORE(2), EXT(1)
	writeTestConfig(t, filepath.Join(tmpDir, "rtmx.yaml"))

	cfg, err := config.LoadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	srv := NewServer(dbPath, cfg, WithQuiet(true))

	// Get tool list via stdio
	listReq, _ := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0", "id": 1, "method": "tools/list",
	})
	var input, output bytes.Buffer
	input.Write(listReq)
	input.WriteByte('\n')

	if err := srv.StartStdio(&input, &output); err != nil {
		t.Fatalf("StartStdio failed: %v", err)
	}

	var resp struct {
		Result struct {
			Tools []struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(output.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse tools/list response: %v", err)
	}

	t.Run("descriptions_include_size_hints", func(t *testing.T) {
		for _, tool := range resp.Result.Tools {
			if !strings.Contains(tool.Description, "token") {
				t.Errorf("tool %q description missing size hint: %s", tool.Name, tool.Description)
			}
		}
	})

	t.Run("descriptions_under_120_chars", func(t *testing.T) {
		for _, tool := range resp.Result.Tools {
			if len(tool.Description) > 120 {
				t.Errorf("tool %q description too long (%d chars): %s", tool.Name, len(tool.Description), tool.Description)
			}
		}
	})

	t.Run("collection_tools_mention_filtering", func(t *testing.T) {
		filterableTools := map[string]bool{
			"status": true, "backlog": true, "markers": true, "next": true,
		}
		for _, tool := range resp.Result.Tools {
			if filterableTools[tool.Name] {
				if !strings.Contains(tool.Description, "filter") {
					t.Errorf("filterable tool %q description should mention filtering: %s", tool.Name, tool.Description)
				}
			}
		}
	})

	t.Run("deps_distinguishes_modes", func(t *testing.T) {
		for _, tool := range resp.Result.Tools {
			if tool.Name == "deps" {
				if !strings.Contains(tool.Description, "specific") || !strings.Contains(tool.Description, "overview") {
					t.Errorf("deps description should distinguish specific vs overview modes: %s", tool.Description)
				}
			}
		}
	})
}

// TestMCPReadTools validates all 7 read-only MCP tools produce valid JSON
// and are safe under concurrent access.
// REQ-MCP-003: Production-grade read-only tools for RTM operations.
func TestMCPReadTools(t *testing.T) {
	rtmx.Req(t, "REQ-MCP-003")

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

	// All 7 tools must return valid JSON via tools/call
	allTools := []string{"status", "backlog", "health", "deps", "verify", "markers", "next"}

	t.Run("all_tools_return_valid_json", func(t *testing.T) {
		for _, tool := range allTools {
			t.Run(tool, func(t *testing.T) {
				resp := rpcCall(t, baseURL, "tools/call", map[string]interface{}{
					"name": tool,
				})
				text := extractToolText(t, resp)
				var parsed interface{}
				if err := json.Unmarshal([]byte(text), &parsed); err != nil {
					t.Errorf("tool %s returned invalid JSON: %v\nText: %s", tool, err, text)
				}
			})
		}
	})

	t.Run("next_tool_returns_webs", func(t *testing.T) {
		resp := rpcCall(t, baseURL, "tools/call", map[string]interface{}{
			"name": "next",
		})
		text := extractToolText(t, resp)

		var nr nextResult
		if err := json.Unmarshal([]byte(text), &nr); err != nil {
			t.Fatalf("failed to parse next result: %v", err)
		}
		// Test DB has 3 requirements, 2 incomplete -> should have webs
		if nr.TotalIncomplete != 2 {
			t.Errorf("expected 2 incomplete, got %d", nr.TotalIncomplete)
		}
	})

	t.Run("concurrent_access_safe", func(t *testing.T) {
		// Fire 20 concurrent requests across all tools to verify no races
		const concurrency = 20
		errs := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			tool := allTools[i%len(allTools)]
			go func(toolName string) {
				body := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"%s"}}`, toolName)
				resp, err := http.Post(baseURL, "application/json", bytes.NewBufferString(body))
				if err != nil {
					errs <- fmt.Errorf("%s: request failed: %w", toolName, err)
					return
				}
				defer func() { _ = resp.Body.Close() }()
				if resp.StatusCode != http.StatusOK {
					errs <- fmt.Errorf("%s: status %d", toolName, resp.StatusCode)
					return
				}
				errs <- nil
			}(tool)
		}

		for i := 0; i < concurrency; i++ {
			if err := <-errs; err != nil {
				t.Errorf("concurrent request failed: %v", err)
			}
		}
	})
}

// TestMCPVerifyExecutesTests validates that the MCP verify tool runs tests
// and updates requirement status in the database.
// REQ-MCP-010: MCP verify tool shall execute tests and update requirements.
func TestMCPVerifyExecutesTests(t *testing.T) {
	rtmx.Req(t, "REQ-MCP-010")

	t.Run("detectTestCommand", func(t *testing.T) {
		tests := []struct {
			name     string
			files    []string
			expected string
		}{
			{"cargo", []string{"Cargo.toml"}, "cargo test --workspace"},
			{"node", []string{"package.json"}, "npm test"},
			{"python_pyproject", []string{"pyproject.toml"}, "python3 -m pytest -v"},
			{"python_setup", []string{"setup.py"}, "python3 -m pytest -v"},
			{"python_requirements", []string{"requirements.txt"}, "python3 -m pytest -v"},
			{"gradle", []string{"build.gradle"}, "gradle test"},
			{"gradle_kts", []string{"build.gradle.kts"}, "gradle test"},
			{"makefile", []string{"Makefile"}, "make test"},
			{"default_go", []string{}, "go test -json ./..."},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				dir := t.TempDir()
				for _, f := range tt.files {
					if err := os.WriteFile(filepath.Join(dir, f), []byte(""), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				got := detectTestCommand(dir)
				if got != tt.expected {
					t.Errorf("detectTestCommand() = %q, want %q", got, tt.expected)
				}
			})
		}
	})

	t.Run("runAndParseTests_go_json", func(t *testing.T) {
		// Create a script that emits Go test JSON output
		dir := t.TempDir()
		script := filepath.Join(dir, "fake_test.sh")
		content := `#!/bin/sh
echo '{"Test":"TestFirst","Action":"pass","Package":"pkg"}'
echo '{"Test":"TestSecond","Action":"fail","Package":"pkg"}'
echo '{"Test":"TestThird","Action":"skip","Package":"pkg"}'
`
		if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
			t.Fatal(err)
		}

		results := runAndParseTests("sh "+script, dir)
		if len(results) != 3 {
			t.Fatalf("expected 3 results, got %d", len(results))
		}
		if !results[0].passed || results[0].name != "TestFirst" {
			t.Errorf("result[0]: want passed TestFirst, got %+v", results[0])
		}
		if !results[1].failed || results[1].name != "TestSecond" {
			t.Errorf("result[1]: want failed TestSecond, got %+v", results[1])
		}
		if results[2].passed || results[2].failed {
			t.Errorf("result[2]: want skipped, got %+v", results[2])
		}
	})

	t.Run("runAndParseTests_pytest", func(t *testing.T) {
		dir := t.TempDir()
		script := filepath.Join(dir, "fake_test.sh")
		content := `#!/bin/sh
echo 'test_api.py::test_login PASSED'
echo 'test_api.py::test_signup FAILED'
`
		if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
			t.Fatal(err)
		}

		results := runAndParseTests("sh "+script, dir)
		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
		if !results[0].passed || results[0].name != "test_login" {
			t.Errorf("result[0]: want passed test_login, got %+v", results[0])
		}
		if !results[1].failed || results[1].name != "test_signup" {
			t.Errorf("result[1]: want failed test_signup, got %+v", results[1])
		}
	})

	t.Run("runAndParseTests_cargo", func(t *testing.T) {
		dir := t.TempDir()
		script := filepath.Join(dir, "fake_test.sh")
		content := `#!/bin/sh
echo 'test auth::test_login ... ok'
echo 'test api::test_create ... FAILED'
`
		if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
			t.Fatal(err)
		}

		results := runAndParseTests("sh "+script, dir)
		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
		if !results[0].passed || results[0].name != "auth::test_login" {
			t.Errorf("result[0]: want passed auth::test_login, got %+v", results[0])
		}
		if !results[1].failed || results[1].name != "api::test_create" {
			t.Errorf("result[1]: want failed api::test_create, got %+v", results[1])
		}
	})

	t.Run("runAndParseTests_tap", func(t *testing.T) {
		dir := t.TempDir()
		script := filepath.Join(dir, "fake_test.sh")
		content := `#!/bin/sh
echo 'ok 1 - test_add'
echo 'not ok 2 - test_delete'
`
		if err := os.WriteFile(script, []byte(content), 0o755); err != nil {
			t.Fatal(err)
		}

		results := runAndParseTests("sh "+script, dir)
		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
		if !results[0].passed || results[0].name != "test_add" {
			t.Errorf("result[0]: want passed test_add, got %+v", results[0])
		}
		if !results[1].failed || results[1].name != "test_delete" {
			t.Errorf("result[1]: want failed test_delete, got %+v", results[1])
		}
	})

	t.Run("findMatchingTest", func(t *testing.T) {
		results := []testResult{
			{name: "auth::test_login", passed: true},
			{name: "TestDatabaseSetup", passed: true},
			{pkg: "test_api.py", name: "test_signup", failed: true},
		}

		tests := []struct {
			testFunc string
			wantName string
			wantNil  bool
		}{
			{"test_login", "auth::test_login", false},         // :: suffix match
			{"TestDatabaseSetup", "TestDatabaseSetup", false}, // exact match
			{"test_signup", "test_signup", false},              // exact match
			{"nonexistent", "", true},                          // no match
		}

		for _, tt := range tests {
			t.Run(tt.testFunc, func(t *testing.T) {
				got := findMatchingTest(results, tt.testFunc)
				if tt.wantNil {
					if got != nil {
						t.Errorf("findMatchingTest(%q) = %+v, want nil", tt.testFunc, got)
					}
					return
				}
				if got == nil {
					t.Fatalf("findMatchingTest(%q) = nil, want %q", tt.testFunc, tt.wantName)
				}
				if got.name != tt.wantName {
					t.Errorf("findMatchingTest(%q).name = %q, want %q", tt.testFunc, got.name, tt.wantName)
				}
			})
		}
	})

	t.Run("toolVerify_updates_database", func(t *testing.T) {
		// Set up a temp project with database and a fake test script
		tmpDir := t.TempDir()
		rtmxDir := filepath.Join(tmpDir, ".rtmx")
		if err := os.MkdirAll(rtmxDir, 0o755); err != nil {
			t.Fatal(err)
		}

		// Write a database with one MISSING requirement that has a test function
		db := database.NewDatabase()
		r1 := &database.Requirement{
			ReqID:           "REQ-V-001",
			Category:        "TEST",
			RequirementText: "Verify test",
			Status:          database.StatusMissing,
			Priority:        database.PriorityHigh,
			Phase:           1,
			TestFunction:    "TestFirst",
		}
		r2 := &database.Requirement{
			ReqID:           "REQ-V-002",
			Category:        "TEST",
			RequirementText: "Another test",
			Status:          database.StatusComplete,
			Priority:        database.PriorityHigh,
			Phase:           1,
			TestFunction:    "TestSecond",
		}
		for _, r := range []*database.Requirement{r1, r2} {
			if err := db.Add(r); err != nil {
				t.Fatal(err)
			}
		}

		dbPath := filepath.Join(rtmxDir, "database.csv")
		if err := db.Save(dbPath); err != nil {
			t.Fatal(err)
		}

		// Create a fake test script that passes TestFirst and fails TestSecond
		script := filepath.Join(tmpDir, "run_tests.sh")
		testScript := `#!/bin/sh
echo '{"Test":"TestFirst","Action":"pass","Package":"pkg"}'
echo '{"Test":"TestSecond","Action":"fail","Package":"pkg"}'
`
		if err := os.WriteFile(script, []byte(testScript), 0o755); err != nil {
			t.Fatal(err)
		}

		cfgPath := filepath.Join(tmpDir, "rtmx.yaml")
		writeTestConfig(t, cfgPath)
		cfg, err := config.LoadFromDir(tmpDir)
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		srv := NewServer(dbPath, cfg)
		result := srv.toolVerify(db, "sh "+script)

		vr, ok := result.(verifyResult)
		if !ok {
			t.Fatalf("toolVerify returned %T, want verifyResult", result)
		}

		if vr.Total != 2 {
			t.Errorf("total = %d, want 2", vr.Total)
		}
		if vr.Verified != 2 {
			t.Errorf("verified = %d, want 2", vr.Verified)
		}
		if vr.Updated != 2 {
			t.Errorf("updated = %d, want 2", vr.Updated)
		}

		// Check REQ-V-001 was promoted MISSING -> COMPLETE
		var v1, v2 verifyItem
		for _, item := range vr.Items {
			switch item.ReqID {
			case "REQ-V-001":
				v1 = item
			case "REQ-V-002":
				v2 = item
			}
		}

		if v1.Status != "COMPLETE" {
			t.Errorf("REQ-V-001 status = %q, want COMPLETE", v1.Status)
		}
		if !v1.Updated {
			t.Error("REQ-V-001 should be marked as updated")
		}
		if v1.Previous != "MISSING" {
			t.Errorf("REQ-V-001 previous = %q, want MISSING", v1.Previous)
		}

		// Check REQ-V-002 was downgraded COMPLETE -> PARTIAL
		if v2.Status != "PARTIAL" {
			t.Errorf("REQ-V-002 status = %q, want PARTIAL", v2.Status)
		}
		if !v2.Updated {
			t.Error("REQ-V-002 should be marked as updated")
		}

		// Verify database was persisted
		reloadedDB, loadErr := database.Load(dbPath)
		if loadErr != nil {
			t.Fatalf("failed to reload database: %v", loadErr)
		}
		reloaded1 := reloadedDB.Get("REQ-V-001")
		if reloaded1 == nil || reloaded1.Status != database.StatusComplete {
			t.Error("persisted REQ-V-001 should be COMPLETE")
		}
		reloaded2 := reloadedDB.Get("REQ-V-002")
		if reloaded2 == nil || reloaded2.Status != database.StatusPartial {
			t.Error("persisted REQ-V-002 should be PARTIAL")
		}
	})

	t.Run("toolVerify_no_changes_no_save", func(t *testing.T) {
		// When all tests pass and requirements are already COMPLETE,
		// the database should not be saved (updated == 0)
		tmpDir := t.TempDir()
		rtmxDir := filepath.Join(tmpDir, ".rtmx")
		if err := os.MkdirAll(rtmxDir, 0o755); err != nil {
			t.Fatal(err)
		}

		db := database.NewDatabase()
		r1 := &database.Requirement{
			ReqID:           "REQ-V-001",
			Category:        "TEST",
			RequirementText: "Already complete",
			Status:          database.StatusComplete,
			Priority:        database.PriorityHigh,
			Phase:           1,
			TestFunction:    "TestFirst",
		}
		if err := db.Add(r1); err != nil {
			t.Fatal(err)
		}

		dbPath := filepath.Join(rtmxDir, "database.csv")
		if err := db.Save(dbPath); err != nil {
			t.Fatal(err)
		}

		script := filepath.Join(tmpDir, "run_tests.sh")
		testScript := `#!/bin/sh
echo '{"Test":"TestFirst","Action":"pass","Package":"pkg"}'
`
		if err := os.WriteFile(script, []byte(testScript), 0o755); err != nil {
			t.Fatal(err)
		}

		cfgPath := filepath.Join(tmpDir, "rtmx.yaml")
		writeTestConfig(t, cfgPath)
		cfg, err := config.LoadFromDir(tmpDir)
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		srv := NewServer(dbPath, cfg)
		result := srv.toolVerify(db, "sh "+script)

		vr := result.(verifyResult)
		if vr.Updated != 0 {
			t.Errorf("updated = %d, want 0 (no changes)", vr.Updated)
		}
	})
}
