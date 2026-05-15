// Package mcp provides an MCP (Model Context Protocol) server that exposes
// RTMX operations as tools for AI agent integration via JSON-RPC 2.0 over HTTP.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
	"github.com/rtmx-ai/rtmx/internal/orchestration"
)

// Server is an MCP server that exposes RTMX tools via JSON-RPC 2.0 over HTTP.
type Server struct {
	host     string
	port     int
	dbPath   string
	cfg      *config.Config
	claims   *orchestration.ClaimStore
	mu       sync.RWMutex
	server   *http.Server
	listener net.Listener
}

// Option configures the Server.
type Option func(*Server)

// WithHost sets the bind address.
func WithHost(host string) Option {
	return func(s *Server) { s.host = host }
}

// WithPort sets the listen port.
func WithPort(port int) Option {
	return func(s *Server) { s.port = port }
}

// NewServer creates a new MCP server for the given project directory.
func NewServer(dbPath string, cfg *config.Config, opts ...Option) *Server {
	s := &Server{
		host:   "localhost",
		port:   3000,
		dbPath: dbPath,
		cfg:    cfg,
	}
	for _, opt := range opts {
		opt(s)
	}
	// Initialize claims store alongside database
	claimsDir := filepath.Join(filepath.Dir(dbPath), "claims")
	s.claims, _ = orchestration.NewClaimStore(claimsDir)
	return s
}

// Addr returns the listen address once the server is started.
// Returns empty string if not yet started.
func (s *Server) Addr() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return ""
}

// Start starts the MCP server. It binds to the configured address and serves
// JSON-RPC 2.0 requests. The server can be stopped with Shutdown.
func (s *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.handleRPC)

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	s.mu.Lock()
	s.listener = ln
	s.server = &http.Server{Handler: mux}
	s.mu.Unlock()

	return s.server.Serve(ln)
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.RLock()
	srv := s.server
	s.mu.RUnlock()
	if srv == nil {
		return nil
	}
	return srv.Shutdown(ctx)
}

// ----- JSON-RPC 2.0 types -----

// rpcRequest is a JSON-RPC 2.0 request.
type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// rpcResponse is a JSON-RPC 2.0 response.
type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

// rpcError is a JSON-RPC 2.0 error object.
type rpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Standard JSON-RPC error codes.
const (
	errParse      = -32700
	errInvalidReq = -32600
	errNoMethod   = -32601
	errInternal   = -32603
)

// ----- MCP protocol types -----

// toolDef describes an MCP tool for the tools/list response.
type toolDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

// toolCallParams contains the params for tools/call.
type toolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// toolResult is what tools/call returns.
type toolResult struct {
	Content []toolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// toolContent is a single piece of content in a tool result.
type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// StartStdio runs the MCP server over stdin/stdout using JSON-RPC 2.0
// newline-delimited messages. This is the standard transport for local
// MCP servers integrated with Claude Code, Cursor, and similar tools.
func (s *Server) StartStdio(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB max message
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		resp := s.processRPC(line)
		if resp == nil {
			continue // notification; no response
		}
		resp = append(resp, '\n')
		if _, err := out.Write(resp); err != nil {
			return fmt.Errorf("write error: %w", err)
		}
	}
	return scanner.Err()
}

// processRPC handles a single JSON-RPC 2.0 request and returns the response bytes.
func (s *Server) processRPC(input []byte) []byte {
	var req rpcRequest
	if err := json.Unmarshal(input, &req); err != nil {
		return marshalResponse(rpcResponse{
			JSONRPC: "2.0",
			Error:   &rpcError{Code: errParse, Message: "parse error"},
		})
	}

	if req.JSONRPC != "2.0" {
		return marshalResponse(rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &rpcError{Code: errInvalidReq, Message: "invalid jsonrpc version"},
		})
	}

	var result interface{}
	var rpcErr *rpcError

	switch req.Method {
	case "initialize":
		result = s.handleInitialize()
	case "tools/list":
		result = s.handleToolsList()
	case "tools/call":
		result, rpcErr = s.handleToolsCall(req.Params)
	case "notifications/initialized":
		// Client acknowledgment; no response needed per MCP spec.
		// Return empty to signal no-op to callers.
		return nil
	default:
		rpcErr = &rpcError{Code: errNoMethod, Message: fmt.Sprintf("method not found: %s", req.Method)}
	}

	resp := rpcResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
	}
	if rpcErr != nil {
		resp.Error = rpcErr
	} else {
		resp.Result = result
	}

	return marshalResponse(resp)
}

func marshalResponse(resp rpcResponse) []byte {
	data, _ := json.Marshal(resp)
	return data
}

// ----- HTTP Handler -----

func (s *Server) handleRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	input, err := io.ReadAll(r.Body)
	if err != nil {
		writeRPCError(w, nil, errParse, "read error")
		return
	}

	resp := s.processRPC(input)
	if resp == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(resp)
}

func writeRPCError(w http.ResponseWriter, id json.RawMessage, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(marshalResponse(rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: msg},
	}))
}

// ----- MCP method handlers -----

func (s *Server) handleInitialize() interface{} {
	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
		"serverInfo": map[string]interface{}{
			"name":    "rtmx",
			"version": "1.0.0",
		},
	}
}

func (s *Server) handleToolsList() interface{} {
	emptyObj := map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}

	tools := []toolDef{
		{
			Name:        "status",
			Description: "Show RTM completion status with counts and percentages",
			InputSchema: emptyObj,
		},
		{
			Name:        "backlog",
			Description: "Show prioritized incomplete requirements",
			InputSchema: emptyObj,
		},
		{
			Name:        "health",
			Description: "Run health checks on the RTM database",
			InputSchema: emptyObj,
		},
		{
			Name:        "deps",
			Description: "Show dependency information for a requirement",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"req_id": map[string]interface{}{
						"type":        "string",
						"description": "Requirement ID (e.g. REQ-GO-001). Omit for overview.",
					},
				},
			},
		},
		{
			Name:        "verify",
			Description: "Run tests and verify requirements. Executes the test command, maps results to requirements, and updates the RTM database. Returns updated status for each requirement.",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "Test command to run (e.g. 'npm test', 'pytest -v', 'go test ./...'). If omitted, auto-detects from project files.",
					},
				},
			},
		},
		{
			Name:        "markers",
			Description: "Show requirement markers found in test files",
			InputSchema: emptyObj,
		},
		{
			Name:        "next",
			Description: "Show independent work webs and highest-priority unblocked requirement",
			InputSchema: emptyObj,
		},
		{
			Name:        "claim",
			Description: "Claim a requirement for an agent (mutation: requires agent_id)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"req_id":   map[string]interface{}{"type": "string", "description": "Requirement ID to claim"},
					"agent_id": map[string]interface{}{"type": "string", "description": "Agent identity claiming the requirement"},
				},
				"required": []string{"req_id", "agent_id"},
			},
		},
		{
			Name:        "release",
			Description: "Release a claimed requirement (mutation: requires agent_id)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"req_id":   map[string]interface{}{"type": "string", "description": "Requirement ID to release"},
					"agent_id": map[string]interface{}{"type": "string", "description": "Agent identity releasing the requirement"},
				},
				"required": []string{"req_id", "agent_id"},
			},
		},
		{
			Name:        "release_assign",
			Description: "Assign requirements to a target version (mutation: requires agent_id)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"version":  map[string]interface{}{"type": "string", "description": "Target version (e.g., v0.4.0)"},
					"req_ids":  map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Requirement IDs to assign"},
					"agent_id": map[string]interface{}{"type": "string", "description": "Agent identity performing the assignment"},
				},
				"required": []string{"version", "req_ids", "agent_id"},
			},
		},
	}

	return map[string]interface{}{"tools": tools}
}

func (s *Server) handleToolsCall(params json.RawMessage) (interface{}, *rpcError) {
	var call toolCallParams
	if err := json.Unmarshal(params, &call); err != nil {
		return nil, &rpcError{Code: errInvalidReq, Message: "invalid params"}
	}

	db, err := database.Load(s.dbPath)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to load database: %v", err)), nil
	}

	var data interface{}

	switch call.Name {
	case "status":
		data = s.toolStatus(db)
	case "backlog":
		data = s.toolBacklog(db)
	case "health":
		data = s.toolHealth(db)
	case "deps":
		reqID, _ := call.Arguments["req_id"].(string)
		data, err = s.toolDeps(db, reqID)
		if err != nil {
			return errorResult(err.Error()), nil
		}
	case "verify":
		command, _ := call.Arguments["command"].(string)
		data = s.toolVerify(db, command)
	case "markers":
		data = s.toolMarkers(db)
	case "next":
		data = s.toolNext(db)
	case "claim":
		return s.toolClaim(call.Arguments)
	case "release":
		return s.toolRelease(call.Arguments)
	case "release_assign":
		return s.toolReleaseAssign(db, call.Arguments)
	default:
		return nil, &rpcError{Code: errNoMethod, Message: fmt.Sprintf("unknown tool: %s", call.Name)}
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return errorResult(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}

	return toolResult{
		Content: []toolContent{
			{Type: "text", Text: string(jsonBytes)},
		},
	}, nil
}

func errorResult(msg string) toolResult {
	return toolResult{
		Content: []toolContent{{Type: "text", Text: msg}},
		IsError: true,
	}
}

// ----- Tool implementations -----

// statusResult is the JSON output for the status tool.
type statusResult struct {
	Total         int                    `json:"total"`
	Complete      int                    `json:"complete"`
	Partial       int                    `json:"partial"`
	Missing       int                    `json:"missing"`
	CompletionPct float64                `json:"completion_pct"`
	Categories    []statusCategoryResult `json:"categories"`
}

type statusCategoryResult struct {
	Name     string  `json:"name"`
	Total    int     `json:"total"`
	Complete int     `json:"complete"`
	Pct      float64 `json:"pct"`
}

func (s *Server) toolStatus(db *database.Database) interface{} {
	counts := db.StatusCounts()
	pct := math.Round(db.CompletionPercentage()*10) / 10

	result := statusResult{
		Total:         db.Len(),
		Complete:      counts[database.StatusComplete],
		Partial:       counts[database.StatusPartial],
		Missing:       counts[database.StatusMissing] + counts[database.StatusNotStarted],
		CompletionPct: pct,
		Categories:    make([]statusCategoryResult, 0),
	}

	categories := db.Categories()
	byCategory := db.ByCategory()
	for _, cat := range categories {
		reqs := byCategory[cat]
		catComplete := 0
		for _, r := range reqs {
			if r.Status == database.StatusComplete {
				catComplete++
			}
		}
		catPct := 0.0
		if len(reqs) > 0 {
			catPct = float64(catComplete) / float64(len(reqs)) * 100
		}
		catPct = math.Round(catPct*10) / 10

		result.Categories = append(result.Categories, statusCategoryResult{
			Name:     cat,
			Total:    len(reqs),
			Complete: catComplete,
			Pct:      catPct,
		})
	}

	return result
}

// backlogItem is a single item in the backlog tool output.
type backlogItem struct {
	ReqID       string  `json:"req_id"`
	Description string  `json:"description"`
	Priority    string  `json:"priority"`
	Status      string  `json:"status"`
	Effort      float64 `json:"effort_weeks"`
	Blocked     bool    `json:"blocked"`
	Blocks      int     `json:"blocks"`
}

// backlogResult is the JSON output for the backlog tool.
type backlogResult struct {
	TotalIncomplete int           `json:"total_incomplete"`
	Items           []backlogItem `json:"items"`
}

func (s *Server) toolBacklog(db *database.Database) interface{} {
	incomplete := db.Backlog()

	items := make([]backlogItem, 0, len(incomplete))
	for _, req := range incomplete {
		blocked := req.IsBlocked(db)
		blocksCount := 0
		for _, r := range db.All() {
			if r.Dependencies.Contains(req.ReqID) && r.IsIncomplete() {
				blocksCount++
			}
		}

		items = append(items, backlogItem{
			ReqID:       req.ReqID,
			Description: req.RequirementText,
			Priority:    string(req.Priority),
			Status:      string(req.Status),
			Effort:      req.EffortWeeks,
			Blocked:     blocked,
			Blocks:      blocksCount,
		})
	}

	return backlogResult{
		TotalIncomplete: len(incomplete),
		Items:           items,
	}
}

func (s *Server) toolHealth(db *database.Database) interface{} {
	counts := db.StatusCounts()
	total := db.Len()
	complete := counts[database.StatusComplete]
	partial := counts[database.StatusPartial]
	missing := counts[database.StatusMissing] + counts[database.StatusNotStarted]
	pct := db.CompletionPercentage()

	type checkEntry struct {
		Name    string `json:"name"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}

	checks := make([]checkEntry, 0)

	// Check 1: database loaded
	checks = append(checks, checkEntry{
		Name:    "rtm_loads",
		Status:  "PASS",
		Message: fmt.Sprintf("RTM database loaded: %d requirements", total),
	})

	// Check 2: orphaned dependencies
	orphanCount := 0
	for _, req := range db.All() {
		for dep := range req.Dependencies {
			if len(dep) > 0 && dep[0] != '@' && !db.Exists(dep) {
				orphanCount++
			}
		}
	}
	if orphanCount > 0 {
		checks = append(checks, checkEntry{
			Name:    "orphaned_deps",
			Status:  "FAIL",
			Message: fmt.Sprintf("Orphaned dependencies: %d errors", orphanCount),
		})
	} else {
		checks = append(checks, checkEntry{
			Name:    "orphaned_deps",
			Status:  "PASS",
			Message: "No orphaned dependencies",
		})
	}

	// Check 3: test coverage
	withTests := 0
	for _, req := range db.All() {
		if req.HasTest() {
			withTests++
		}
	}
	testPct := 0.0
	if total > 0 {
		testPct = float64(withTests) / float64(total) * 100
	}
	if testPct >= 80 {
		checks = append(checks, checkEntry{
			Name:    "test_coverage",
			Status:  "PASS",
			Message: fmt.Sprintf("Test coverage: %.1f%%", testPct),
		})
	} else {
		checks = append(checks, checkEntry{
			Name:    "test_coverage",
			Status:  "WARN",
			Message: fmt.Sprintf("Test coverage: %.1f%% (%d without tests)", testPct, total-withTests),
		})
	}

	// Determine overall status
	overallStatus := "HEALTHY"
	for _, c := range checks {
		if c.Status == "FAIL" {
			overallStatus = "UNHEALTHY"
			break
		}
		if c.Status == "WARN" {
			overallStatus = "WARNING"
		}
	}

	result := struct {
		Status string       `json:"status"`
		Checks []checkEntry `json:"checks"`
		Stats  struct {
			Total      int     `json:"total"`
			Complete   int     `json:"complete"`
			Partial    int     `json:"partial"`
			Missing    int     `json:"missing"`
			Completion float64 `json:"completion_percent"`
		} `json:"stats"`
	}{
		Status: overallStatus,
		Checks: checks,
	}
	result.Stats.Total = total
	result.Stats.Complete = complete
	result.Stats.Partial = partial
	result.Stats.Missing = missing
	result.Stats.Completion = math.Round(pct*10) / 10

	return result
}

// depsResult is the JSON output for the deps tool.
type depsResult struct {
	ReqID        string     `json:"req_id,omitempty"`
	Description  string     `json:"description,omitempty"`
	Status       string     `json:"status,omitempty"`
	Dependencies []depEntry `json:"dependencies,omitempty"`
	Dependents   []depEntry `json:"dependents,omitempty"`
	Overview     []depEntry `json:"overview,omitempty"`
}

type depEntry struct {
	ReqID       string `json:"req_id"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	DepsCount   int    `json:"deps_count,omitempty"`
	BlocksCount int    `json:"blocks_count,omitempty"`
}

func (s *Server) toolDeps(db *database.Database, reqID string) (interface{}, error) {
	g := graph.NewGraph(db)

	if reqID != "" {
		req := db.Get(reqID)
		if req == nil {
			return nil, fmt.Errorf("requirement %s not found", reqID)
		}

		deps := g.Dependencies(reqID)
		dependents := g.Dependents(reqID)

		depEntries := make([]depEntry, 0, len(deps))
		for _, d := range deps {
			r := db.Get(d)
			desc := ""
			status := ""
			if r != nil {
				desc = r.RequirementText
				status = string(r.Status)
			}
			depEntries = append(depEntries, depEntry{
				ReqID:       d,
				Description: desc,
				Status:      status,
			})
		}

		depntEntries := make([]depEntry, 0, len(dependents))
		for _, d := range dependents {
			r := db.Get(d)
			desc := ""
			status := ""
			if r != nil {
				desc = r.RequirementText
				status = string(r.Status)
			}
			depntEntries = append(depntEntries, depEntry{
				ReqID:       d,
				Description: desc,
				Status:      status,
			})
		}

		return depsResult{
			ReqID:        reqID,
			Description:  req.RequirementText,
			Status:       string(req.Status),
			Dependencies: depEntries,
			Dependents:   depntEntries,
		}, nil
	}

	// Overview mode
	type info struct {
		id    string
		deps  int
		blks  int
		desc  string
	}

	var entries []info
	for _, r := range db.All() {
		d := len(g.Dependencies(r.ReqID))
		b := len(g.Dependents(r.ReqID))
		entries = append(entries, info{r.ReqID, d, b, r.RequirementText})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].blks > entries[j].blks
	})

	overview := make([]depEntry, 0, len(entries))
	for _, e := range entries {
		overview = append(overview, depEntry{
			ReqID:       e.id,
			Description: e.desc,
			DepsCount:   e.deps,
			BlocksCount: e.blks,
		})
	}

	return depsResult{Overview: overview}, nil
}

// verifyItem is a single entry in the verify tool output.
type verifyItem struct {
	ReqID      string `json:"req_id"`
	Status     string `json:"status"`
	Previous   string `json:"previous_status,omitempty"`
	HasTest    bool   `json:"has_test"`
	TestFunc   string `json:"test_function,omitempty"`
	TestPassed bool   `json:"test_passed,omitempty"`
	Updated    bool   `json:"updated,omitempty"`
}

// verifyResult is the JSON output for the verify tool.
type verifyResult struct {
	Total     int          `json:"total"`
	Complete  int          `json:"complete"`
	Verified  int          `json:"verified"`
	Updated   int          `json:"updated"`
	Command   string       `json:"command,omitempty"`
	Items     []verifyItem `json:"items"`
}

// testResult holds the outcome of a single parsed test.
type testResult struct {
	pkg    string
	name   string
	passed bool
	failed bool
}

func (s *Server) toolVerify(db *database.Database, command string) interface{} {
	// Determine test command
	if command == "" {
		command = detectTestCommand(filepath.Dir(s.dbPath))
	}

	// Run the test command and parse output
	testResults := runAndParseTests(command, filepath.Dir(filepath.Dir(s.dbPath)))

	// Map test results to requirements and update database
	items := make([]verifyItem, 0)
	updated := 0
	verified := 0
	complete := 0

	for _, req := range db.All() {
		item := verifyItem{
			ReqID:    req.ReqID,
			HasTest:  req.HasTest(),
			TestFunc: req.TestFunction,
		}

		if req.TestFunction != "" {
			// Try to find a matching test result
			if tr := findMatchingTest(testResults, req.TestFunction); tr != nil {
				verified++
				item.TestPassed = tr.passed

				prevStatus := req.Status
				if tr.passed {
					req.Status = database.StatusComplete
				} else if tr.failed && req.Status == database.StatusComplete {
					req.Status = database.StatusPartial
				}

				if req.Status != prevStatus {
					item.Previous = string(prevStatus)
					item.Updated = true
					if req.Status == database.StatusComplete || req.Status == database.StatusPartial {
						req.SetStartedDate()
					}
					if req.Status == database.StatusComplete {
						req.SetCompletedDate()
					}
					updated++
				}
			}
		}

		item.Status = string(req.Status)
		if req.Status == database.StatusComplete {
			complete++
		}
		items = append(items, item)
	}

	// Save the database if anything changed
	if updated > 0 {
		_ = db.Save(s.dbPath)
	}

	return verifyResult{
		Total:    db.Len(),
		Complete: complete,
		Verified: verified,
		Updated:  updated,
		Command:  command,
		Items:    items,
	}
}

// detectTestCommand determines the test command from project files.
func detectTestCommand(dir string) string {
	exists := func(name string) bool {
		_, err := os.Stat(filepath.Join(dir, name))
		return err == nil
	}
	switch {
	case exists("Cargo.toml"):
		return "cargo test --workspace"
	case exists("package.json"):
		return "npm test"
	case exists("pyproject.toml"), exists("setup.py"), exists("requirements.txt"):
		return "python3 -m pytest -v"
	case exists("build.gradle"), exists("build.gradle.kts"):
		return "gradle test"
	case exists("Makefile"):
		return "make test"
	default:
		return "go test -json ./..."
	}
}

// runAndParseTests executes a test command and parses the output into test results.
func runAndParseTests(command string, workDir string) []testResult {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return nil
	}

	testCmd := exec.Command(parts[0], parts[1:]...)
	testCmd.Dir = workDir

	stdout, err := testCmd.StdoutPipe()
	if err != nil {
		return nil
	}
	testCmd.Stderr = nil // discard stderr

	if err := testCmd.Start(); err != nil {
		return nil
	}

	var results []testResult
	scanner := bufio.NewScanner(stdout)

	// Patterns for multi-framework test output parsing
	goTestEvent := regexp.MustCompile(`"Test"\s*:\s*"([^"]+)"`)
	goTestAction := regexp.MustCompile(`"Action"\s*:\s*"(pass|fail|skip)"`)
	cargoPattern := regexp.MustCompile(`^test\s+(\S+)\s+\.\.\.\s+(ok|FAILED|ignored)`)
	pytestPattern := regexp.MustCompile(`^(\S+\.py)::(\S+)\s+(PASSED|FAILED|SKIPPED)`)
	nodePattern := regexp.MustCompile(`^\s*(ok|not ok)\s+\d+\s+[-—]\s*(.+)`)

	for scanner.Scan() {
		line := scanner.Text()

		// Go test JSON format
		if strings.Contains(line, `"Test"`) && strings.Contains(line, `"Action"`) {
			testMatch := goTestEvent.FindStringSubmatch(line)
			actionMatch := goTestAction.FindStringSubmatch(line)
			if len(testMatch) > 1 && len(actionMatch) > 1 {
				tr := testResult{name: testMatch[1]}
				switch actionMatch[1] {
				case "pass":
					tr.passed = true
				case "fail":
					tr.failed = true
				}
				results = append(results, tr)
			}
			continue
		}

		// Cargo test format
		if m := cargoPattern.FindStringSubmatch(line); len(m) > 2 {
			tr := testResult{name: m[1]}
			switch m[2] {
			case "ok":
				tr.passed = true
			case "FAILED":
				tr.failed = true
			}
			results = append(results, tr)
			continue
		}

		// pytest verbose format
		if m := pytestPattern.FindStringSubmatch(line); len(m) > 3 {
			tr := testResult{pkg: m[1], name: m[2]}
			switch m[3] {
			case "PASSED":
				tr.passed = true
			case "FAILED":
				tr.failed = true
			}
			results = append(results, tr)
			continue
		}

		// Node.js TAP-like format
		if m := nodePattern.FindStringSubmatch(line); len(m) > 2 {
			tr := testResult{name: strings.TrimSpace(m[2])}
			tr.passed = m[1] == "ok"
			tr.failed = m[1] == "not ok"
			results = append(results, tr)
			continue
		}
	}

	_ = testCmd.Wait()
	return results
}

// findMatchingTest finds a test result matching a database test_function,
// supporting suffix matching at path separator boundaries.
func findMatchingTest(results []testResult, testFunc string) *testResult {
	for i := range results {
		name := results[i].name
		if name == testFunc {
			return &results[i]
		}
		// Suffix match at :: boundary (Rust module paths)
		if strings.HasSuffix(name, "::"+testFunc) {
			return &results[i]
		}
		// Suffix match at . boundary (Python package paths)
		if strings.HasSuffix(name, "."+testFunc) {
			return &results[i]
		}
		// Suffix match at / boundary (Go package paths)
		if strings.HasSuffix(name, "/"+testFunc) {
			return &results[i]
		}
		// Partial name match (test_database matches test_database_setup)
		if strings.Contains(name, testFunc) {
			return &results[i]
		}
	}
	return nil
}

// markerEntry is a single marker in the markers tool output.
type markerEntry struct {
	ReqID    string `json:"req_id"`
	Status   string `json:"status"`
	HasTest  bool   `json:"has_test"`
	TestFunc string `json:"test_function,omitempty"`
}

// markersResult is the JSON output for the markers tool.
type markersResult struct {
	Total      int           `json:"total"`
	WithTests  int           `json:"with_tests"`
	Missing    int           `json:"missing"`
	Items      []markerEntry `json:"items"`
}

func (s *Server) toolMarkers(db *database.Database) interface{} {
	items := make([]markerEntry, 0)
	withTests := 0

	for _, req := range db.All() {
		hasTest := req.HasTest()
		if hasTest {
			withTests++
		}

		testFunc := req.TestFunction
		if testFunc == "" {
			testFunc = ""
		}

		items = append(items, markerEntry{
			ReqID:    req.ReqID,
			Status:   string(req.Status),
			HasTest:  hasTest,
			TestFunc: testFunc,
		})
	}

	return markersResult{
		Total:     db.Len(),
		WithTests: withTests,
		Missing:   db.Len() - withTests,
		Items:     items,
	}
}

// nextWebResult is a single web in the next tool output.
type nextWebResult struct {
	ID          int      `json:"id"`
	Size        int      `json:"size"`
	Unblocked   int      `json:"unblocked"`
	Blocked     int      `json:"blocked"`
	EffortWeeks float64  `json:"effort_weeks"`
	Members     []string `json:"members"`
	TopItem     string   `json:"top_item,omitempty"`
	TopPriority string   `json:"top_priority,omitempty"`
}

// nextResult is the JSON output for the next tool.
type nextResult struct {
	TotalWebs        int             `json:"total_webs"`
	TotalIncomplete  int             `json:"total_incomplete"`
	TotalEffortWeeks float64         `json:"total_effort_weeks"`
	Webs             []nextWebResult `json:"webs"`
}

func (s *Server) toolNext(db *database.Database) interface{} {
	g := graph.NewGraph(db)
	webs := g.DetectWebs()

	result := nextResult{
		TotalWebs: len(webs),
		Webs:      make([]nextWebResult, 0, len(webs)),
	}

	for i, web := range webs {
		result.TotalIncomplete += len(web.IDs)
		result.TotalEffortWeeks += web.TotalEffort

		wr := nextWebResult{
			ID:          i + 1,
			Size:        len(web.IDs),
			Unblocked:   len(web.Unblocked),
			Blocked:     len(web.Blocked),
			EffortWeeks: web.TotalEffort,
			Members:     web.IDs,
		}

		// Find top-priority unblocked item
		if len(web.Unblocked) > 0 {
			var bestReq *database.Requirement
			for _, id := range web.Unblocked {
				r := db.Get(id)
				if r == nil {
					continue
				}
				if bestReq == nil || r.Priority.Weight() < bestReq.Priority.Weight() {
					bestReq = r
				}
			}
			if bestReq != nil {
				wr.TopItem = bestReq.ReqID
				wr.TopPriority = string(bestReq.Priority)
			}
		}

		result.Webs = append(result.Webs, wr)
	}

	return result
}

// ----- Mutation tools -----

func (s *Server) toolClaim(args map[string]interface{}) (interface{}, *rpcError) {
	reqID, _ := args["req_id"].(string)
	agentID, _ := args["agent_id"].(string)

	if reqID == "" || agentID == "" {
		return errorResult("req_id and agent_id are required"), nil
	}
	if s.claims == nil {
		return errorResult("claims store not initialized"), nil
	}

	claim, err := s.claims.Claim(reqID, agentID)
	if err != nil {
		return errorResult(err.Error()), nil
	}

	data, _ := json.Marshal(claim)
	return toolResult{
		Content: []toolContent{{Type: "text", Text: string(data)}},
	}, nil
}

func (s *Server) toolRelease(args map[string]interface{}) (interface{}, *rpcError) {
	reqID, _ := args["req_id"].(string)
	agentID, _ := args["agent_id"].(string)

	if reqID == "" || agentID == "" {
		return errorResult("req_id and agent_id are required"), nil
	}
	if s.claims == nil {
		return errorResult("claims store not initialized"), nil
	}

	if err := s.claims.Release(reqID, agentID); err != nil {
		return errorResult(err.Error()), nil
	}

	result := map[string]interface{}{
		"released": reqID,
		"agent_id": agentID,
	}
	data, _ := json.Marshal(result)
	return toolResult{
		Content: []toolContent{{Type: "text", Text: string(data)}},
	}, nil
}

func (s *Server) toolReleaseAssign(db *database.Database, args map[string]interface{}) (interface{}, *rpcError) {
	version, _ := args["version"].(string)
	agentID, _ := args["agent_id"].(string)
	reqIDsRaw, _ := args["req_ids"].([]interface{})

	if version == "" || agentID == "" || len(reqIDsRaw) == 0 {
		return errorResult("version, agent_id, and req_ids are required"), nil
	}

	var assigned []string
	var errs []string

	for _, raw := range reqIDsRaw {
		id, ok := raw.(string)
		if !ok {
			continue
		}
		req := db.Get(id)
		if req == nil {
			errs = append(errs, fmt.Sprintf("%s: not found", id))
			continue
		}
		req.SetTargetVersion(version)
		assigned = append(assigned, id)
	}

	// Save the database
	if len(assigned) > 0 {
		if err := db.Save(s.dbPath); err != nil {
			return errorResult(fmt.Sprintf("failed to save database: %v", err)), nil
		}
	}

	result := map[string]interface{}{
		"version":  version,
		"assigned": assigned,
		"agent_id": agentID,
	}
	if len(errs) > 0 {
		result["errors"] = errs
	}
	data, _ := json.Marshal(result)
	return toolResult{
		Content: []toolContent{{Type: "text", Text: string(data)}},
	}, nil
}

