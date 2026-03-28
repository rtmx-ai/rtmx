// Package mcp provides an MCP (Model Context Protocol) server that exposes
// RTMX operations as tools for AI agent integration via JSON-RPC 2.0 over HTTP.
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"net/http"
	"sort"
	"sync"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
)

// Server is an MCP server that exposes RTMX tools via JSON-RPC 2.0 over HTTP.
type Server struct {
	host     string
	port     int
	dbPath   string
	cfg      *config.Config
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

// ----- Handler -----

func (s *Server) handleRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req rpcRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeRPCError(w, nil, errParse, "parse error")
		return
	}

	if req.JSONRPC != "2.0" {
		writeRPCError(w, req.ID, errInvalidReq, "invalid jsonrpc version")
		return
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

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func writeRPCError(w http.ResponseWriter, id json.RawMessage, code int, msg string) {
	resp := rpcResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &rpcError{Code: code, Message: msg},
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
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
			Description: "Show verification status for requirements with linked tests",
			InputSchema: emptyObj,
		},
		{
			Name:        "markers",
			Description: "Show requirement markers found in test files",
			InputSchema: emptyObj,
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
		data = s.toolVerify(db)
	case "markers":
		data = s.toolMarkers(db)
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

// healthCheckResult mirrors the CLI health check output.
type healthCheckResult struct {
	Status string `json:"status"`
	Checks []struct {
		Name    string `json:"name"`
		Status  string `json:"status"`
		Message string `json:"message"`
	} `json:"checks"`
	Stats struct {
		Total      int     `json:"total"`
		Complete   int     `json:"complete"`
		Partial    int     `json:"partial"`
		Missing    int     `json:"missing"`
		Completion float64 `json:"completion_percent"`
	} `json:"stats"`
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
	ReqID    string `json:"req_id"`
	Status   string `json:"status"`
	HasTest  bool   `json:"has_test"`
	TestFunc string `json:"test_function,omitempty"`
}

// verifyResult is the JSON output for the verify tool.
type verifyResult struct {
	Total     int          `json:"total"`
	WithTests int          `json:"with_tests"`
	Items     []verifyItem `json:"items"`
}

func (s *Server) toolVerify(db *database.Database) interface{} {
	items := make([]verifyItem, 0)
	withTests := 0

	for _, req := range db.All() {
		hasTest := req.HasTest()
		if hasTest {
			withTests++
		}
		items = append(items, verifyItem{
			ReqID:    req.ReqID,
			Status:   string(req.Status),
			HasTest:  hasTest,
			TestFunc: req.TestFunction,
		})
	}

	return verifyResult{
		Total:     db.Len(),
		WithTests: withTests,
		Items:     items,
	}
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

