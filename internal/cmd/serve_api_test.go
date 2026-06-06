package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/orchestration"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func testDB() *database.Database {
	db := database.NewDatabase()
	reqs := []*database.Requirement{
		{ReqID: "REQ-CLI-001", Category: "CLI", Subcategory: "Foundation", RequirementText: "Build CLI framework", Status: database.StatusComplete, Priority: database.PriorityP0, Phase: 1, EffortWeeks: 1.0, Assignee: "alice", Sprint: "v1.0.0", StartedDate: "2026-01-01", CompletedDate: "2026-01-15", TestModule: "internal/cmd/cli_test.go", TestFunction: "TestCLIFramework", ValidationMethod: "Unit Test", RequirementFile: ".rtmx/requirements/CLI/REQ-CLI-001.md"},
		{ReqID: "REQ-CLI-002", Category: "CLI", Subcategory: "Commands", RequirementText: "Add status command", Status: database.StatusComplete, Priority: database.PriorityHigh, Phase: 1, EffortWeeks: 0.5, Assignee: "alice", Sprint: "v1.0.0"},
		{ReqID: "REQ-MCP-001", Category: "MCP", Subcategory: "Server", RequirementText: "MCP server implementation", Status: database.StatusPartial, Priority: database.PriorityP0, Phase: 2, EffortWeeks: 2.0, Assignee: "bob", Sprint: "v1.1.0"},
		{ReqID: "REQ-MCP-002", Category: "MCP", Subcategory: "Tools", RequirementText: "MCP tool registration", Status: database.StatusMissing, Priority: database.PriorityHigh, Phase: 2, EffortWeeks: 1.0},
		{ReqID: "REQ-API-001", Category: "API", Subcategory: "REST", RequirementText: "Requirements list endpoint with pagination", Status: database.StatusMissing, Priority: database.PriorityP0, Phase: 3, EffortWeeks: 0.5, Notes: "keystone endpoint"},
	}
	for _, r := range reqs {
		if r.Dependencies == nil {
			r.Dependencies = make(database.StringSet)
		}
		if r.Blocks == nil {
			r.Blocks = make(database.StringSet)
		}
		_ = db.Add(r)
	}
	// Set up dependency relationships: CLI-002 depends on CLI-001, MCP-002 depends on MCP-001, API-001 depends on MCP-001
	db.Get("REQ-CLI-002").Dependencies.Add("REQ-CLI-001")
	db.Get("REQ-CLI-001").Blocks.Add("REQ-CLI-002")
	db.Get("REQ-MCP-002").Dependencies.Add("REQ-MCP-001")
	db.Get("REQ-MCP-001").Blocks.Add("REQ-MCP-002")
	db.Get("REQ-API-001").Dependencies.Add("REQ-MCP-001")
	db.Get("REQ-MCP-001").Blocks.Add("REQ-API-001")
	return db
}

func apiGet(t *testing.T, mux http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	return w
}

func decodeResponse(t *testing.T, w *httptest.ResponseRecorder) apiRequirementsResponse {
	t.Helper()
	var resp apiRequirementsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v\nbody: %s", err, w.Body.String())
	}
	return resp
}

// TestAPIRequirements validates the GET /api/requirements endpoint.
// REQ-API-001: Requirements list endpoint with filter/sort/paginate.
func TestAPIRequirements(t *testing.T) {
	rtmx.Req(t, "REQ-API-001")

	db := testDB()
	cfg := &config.Config{}
	mux := NewDashboardMux(db, cfg)

	t.Run("returns_all_requirements", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		resp := decodeResponse(t, w)
		if resp.Pagination.Total != 5 {
			t.Errorf("total = %d, want 5", resp.Pagination.Total)
		}
		if len(resp.Requirements) != 5 {
			t.Errorf("len = %d, want 5", len(resp.Requirements))
		}
	})

	t.Run("filter_by_category", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?category=CLI")
		resp := decodeResponse(t, w)
		if resp.Pagination.Total != 2 {
			t.Errorf("total = %d, want 2", resp.Pagination.Total)
		}
		for _, r := range resp.Requirements {
			if r.Category != "CLI" {
				t.Errorf("expected category CLI, got %s", r.Category)
			}
		}
		if resp.Filters.Category == nil || *resp.Filters.Category != "CLI" {
			t.Error("filters_applied should show category=CLI")
		}
	})

	t.Run("filter_by_status", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?status=COMPLETE")
		resp := decodeResponse(t, w)
		if resp.Pagination.Total != 2 {
			t.Errorf("total = %d, want 2", resp.Pagination.Total)
		}
		for _, r := range resp.Requirements {
			if r.Status != "COMPLETE" {
				t.Errorf("expected status COMPLETE, got %s", r.Status)
			}
		}
	})

	t.Run("filter_by_priority", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?priority=P0")
		resp := decodeResponse(t, w)
		if resp.Pagination.Total != 3 {
			t.Errorf("total = %d, want 3", resp.Pagination.Total)
		}
	})

	t.Run("filter_by_assignee", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?assignee=bob")
		resp := decodeResponse(t, w)
		if resp.Pagination.Total != 1 {
			t.Errorf("total = %d, want 1", resp.Pagination.Total)
		}
		if resp.Requirements[0].ReqID != "REQ-MCP-001" {
			t.Errorf("expected REQ-MCP-001, got %s", resp.Requirements[0].ReqID)
		}
	})

	t.Run("filter_by_version", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?version=v1.0.0")
		resp := decodeResponse(t, w)
		if resp.Pagination.Total != 2 {
			t.Errorf("total = %d, want 2", resp.Pagination.Total)
		}
	})

	t.Run("search_by_req_id", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?search=MCP-001")
		resp := decodeResponse(t, w)
		if resp.Pagination.Total != 1 {
			t.Errorf("total = %d, want 1", resp.Pagination.Total)
		}
		if resp.Requirements[0].ReqID != "REQ-MCP-001" {
			t.Errorf("expected REQ-MCP-001, got %s", resp.Requirements[0].ReqID)
		}
	})

	t.Run("search_by_text", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?search=pagination")
		resp := decodeResponse(t, w)
		if resp.Pagination.Total != 1 {
			t.Errorf("total = %d, want 1", resp.Pagination.Total)
		}
	})

	t.Run("search_by_notes", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?search=keystone")
		resp := decodeResponse(t, w)
		if resp.Pagination.Total != 1 {
			t.Errorf("total = %d, want 1", resp.Pagination.Total)
		}
	})

	t.Run("sort_by_priority", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?sort=priority")
		resp := decodeResponse(t, w)
		// P0 reqs first (weight 0), then HIGH (weight 1)
		if resp.Requirements[0].Priority != "P0" {
			t.Errorf("first req should be P0, got %s", resp.Requirements[0].Priority)
		}
	})

	t.Run("sort_by_status_desc", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?sort=status&order=desc")
		resp := decodeResponse(t, w)
		// desc: MISSING (weight 2) first
		if resp.Requirements[0].Status != "MISSING" {
			t.Errorf("first req should be MISSING, got %s", resp.Requirements[0].Status)
		}
	})

	t.Run("sort_by_effort_weeks", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?sort=effort_weeks")
		resp := decodeResponse(t, w)
		if resp.Requirements[0].EffortWeeks > resp.Requirements[len(resp.Requirements)-1].EffortWeeks {
			t.Error("expected ascending effort_weeks order")
		}
	})

	t.Run("pagination_first_page", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?per_page=2&page=1")
		resp := decodeResponse(t, w)
		if len(resp.Requirements) != 2 {
			t.Errorf("len = %d, want 2", len(resp.Requirements))
		}
		if resp.Pagination.Total != 5 {
			t.Errorf("total = %d, want 5", resp.Pagination.Total)
		}
		if resp.Pagination.TotalPages != 3 {
			t.Errorf("total_pages = %d, want 3", resp.Pagination.TotalPages)
		}
		if resp.Pagination.Page != 1 {
			t.Errorf("page = %d, want 1", resp.Pagination.Page)
		}
	})

	t.Run("pagination_last_page_partial", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?per_page=2&page=3")
		resp := decodeResponse(t, w)
		if len(resp.Requirements) != 1 {
			t.Errorf("len = %d, want 1", len(resp.Requirements))
		}
	})

	t.Run("pagination_beyond_range", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?per_page=2&page=10")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		resp := decodeResponse(t, w)
		if len(resp.Requirements) != 0 {
			t.Errorf("len = %d, want 0", len(resp.Requirements))
		}
	})

	t.Run("per_page_capped_at_200", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?per_page=500")
		resp := decodeResponse(t, w)
		if resp.Pagination.PerPage != 200 {
			t.Errorf("per_page = %d, want 200 (capped)", resp.Pagination.PerPage)
		}
	})

	t.Run("combined_filters", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?category=CLI&status=COMPLETE")
		resp := decodeResponse(t, w)
		if resp.Pagination.Total != 2 {
			t.Errorf("total = %d, want 2", resp.Pagination.Total)
		}
	})

	t.Run("empty_result", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?category=NONEXISTENT")
		resp := decodeResponse(t, w)
		if resp.Pagination.Total != 0 {
			t.Errorf("total = %d, want 0", resp.Pagination.Total)
		}
		if len(resp.Requirements) != 0 {
			t.Errorf("len = %d, want 0", len(resp.Requirements))
		}
		if resp.Pagination.TotalPages != 1 {
			t.Errorf("total_pages = %d, want 1", resp.Pagination.TotalPages)
		}
	})

	t.Run("invalid_status_400", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?status=INVALID")
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})

	t.Run("invalid_priority_400", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?priority=INVALID")
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})

	t.Run("invalid_sort_field_400", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?sort=unknown_field")
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})

	t.Run("invalid_order_400", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?order=sideways")
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})

	t.Run("invalid_page_400", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?page=0")
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})

	t.Run("method_not_allowed", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/requirements", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want 405", w.Code)
		}
	})

	t.Run("response_includes_all_fields", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements?category=CLI&per_page=1")
		resp := decodeResponse(t, w)
		r := resp.Requirements[0]
		if r.ReqID == "" || r.Category == "" || r.Status == "" || r.Priority == "" {
			t.Error("missing required fields in response")
		}
		if r.Dependencies == nil || r.Blocks == nil {
			t.Error("dependencies and blocks should be arrays, not null")
		}
	})

	t.Run("json_content_type", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements")
		ct := w.Header().Get("Content-Type")
		if ct != "application/json" {
			t.Errorf("Content-Type = %s, want application/json", ct)
		}
	})

	t.Run("concurrent_requests_safe", func(t *testing.T) {
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				w := apiGet(t, mux, "/api/requirements")
				if w.Code != http.StatusOK {
					t.Errorf("concurrent request failed: %d", w.Code)
				}
				done <- true
			}()
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

// testMuxWithPath creates a mux with a real dbPath for persistence tests.
func testMuxWithPath(t *testing.T, db *database.Database) (http.Handler, string) {
	t.Helper()
	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0o755)
	dbPath := filepath.Join(rtmxDir, "database.csv")
	_ = db.Save(dbPath)
	cfg := &config.Config{}
	return NewDashboardMuxWithPath(db, cfg, dbPath), dbPath
}

// TestAPIRequirementDetail validates GET /api/requirements/:id.
// REQ-API-002: Requirement detail endpoint with dependencies.
func TestAPIRequirementDetail(t *testing.T) {
	rtmx.Req(t, "REQ-API-002")

	db := testDB()
	mux, _ := testMuxWithPath(t, db)

	t.Run("returns_full_detail", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements/REQ-CLI-001")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		var resp apiRequirementDetailResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Requirement.ReqID != "REQ-CLI-001" {
			t.Errorf("req_id = %s, want REQ-CLI-001", resp.Requirement.ReqID)
		}
		if resp.Requirement.TestModule != "internal/cmd/cli_test.go" {
			t.Errorf("test_module = %s, want internal/cmd/cli_test.go", resp.Requirement.TestModule)
		}
	})

	t.Run("upstream_dependencies", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements/REQ-MCP-002")
		var resp apiRequirementDetailResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.DependencyDetail.Upstream) != 1 {
			t.Fatalf("upstream len = %d, want 1", len(resp.DependencyDetail.Upstream))
		}
		if resp.DependencyDetail.Upstream[0].ReqID != "REQ-MCP-001" {
			t.Errorf("upstream[0] = %s, want REQ-MCP-001", resp.DependencyDetail.Upstream[0].ReqID)
		}
	})

	t.Run("downstream_dependencies", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements/REQ-MCP-001")
		var resp apiRequirementDetailResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.DependencyDetail.Downstream) < 2 {
			t.Errorf("downstream len = %d, want >= 2", len(resp.DependencyDetail.Downstream))
		}
	})

	t.Run("transitive_counts", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements/REQ-MCP-001")
		var resp apiRequirementDetailResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.DependencyDetail.TransitiveDownstreamCount < 2 {
			t.Errorf("transitive downstream = %d, want >= 2", resp.DependencyDetail.TransitiveDownstreamCount)
		}
	})

	t.Run("all_upstream_complete", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements/REQ-CLI-002")
		var resp apiRequirementDetailResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if !resp.DependencyDetail.AllUpstreamComplete {
			t.Error("all_upstream_complete should be true (CLI-001 is COMPLETE)")
		}
	})

	t.Run("all_upstream_not_complete", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements/REQ-API-001")
		var resp apiRequirementDetailResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.DependencyDetail.AllUpstreamComplete {
			t.Error("all_upstream_complete should be false (MCP-001 is PARTIAL)")
		}
	})

	t.Run("no_dependencies_empty_arrays", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements/REQ-CLI-001")
		var resp apiRequirementDetailResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.DependencyDetail.Upstream == nil {
			t.Error("upstream should be empty array, not null")
		}
	})

	t.Run("not_found_404", func(t *testing.T) {
		w := apiGet(t, mux, "/api/requirements/REQ-NONEXISTENT")
		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", w.Code)
		}
	})
}

// TestAPIRequirementUpdate validates PATCH /api/requirements/:id.
// REQ-API-003: Requirement update endpoint.
func TestAPIRequirementUpdate(t *testing.T) {
	rtmx.Req(t, "REQ-API-003")

	t.Run("update_status", func(t *testing.T) {
		db := testDB()
		mux, _ := testMuxWithPath(t, db)
		body := strings.NewReader(`{"status":"PARTIAL"}`)
		req := httptest.NewRequest("PATCH", "/api/requirements/REQ-MCP-002", body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
		}
		var resp apiRequirementDetailResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Requirement.Status != "PARTIAL" {
			t.Errorf("status = %s, want PARTIAL", resp.Requirement.Status)
		}
	})

	t.Run("update_assignee", func(t *testing.T) {
		db := testDB()
		mux, _ := testMuxWithPath(t, db)
		body := strings.NewReader(`{"assignee":"charlie"}`)
		req := httptest.NewRequest("PATCH", "/api/requirements/REQ-MCP-002", body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		var resp apiRequirementDetailResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Requirement.Assignee != "charlie" {
			t.Errorf("assignee = %s, want charlie", resp.Requirement.Assignee)
		}
	})

	t.Run("immutable_field_rejected", func(t *testing.T) {
		db := testDB()
		mux, _ := testMuxWithPath(t, db)
		body := strings.NewReader(`{"category":"NEWCAT"}`)
		req := httptest.NewRequest("PATCH", "/api/requirements/REQ-MCP-002", body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})

	t.Run("invalid_status_rejected", func(t *testing.T) {
		db := testDB()
		mux, _ := testMuxWithPath(t, db)
		body := strings.NewReader(`{"status":"INVALID"}`)
		req := httptest.NewRequest("PATCH", "/api/requirements/REQ-MCP-002", body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})

	t.Run("invalid_priority_rejected", func(t *testing.T) {
		db := testDB()
		mux, _ := testMuxWithPath(t, db)
		body := strings.NewReader(`{"priority":"INVALID"}`)
		req := httptest.NewRequest("PATCH", "/api/requirements/REQ-MCP-002", body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want 400", w.Code)
		}
	})

	t.Run("not_found_404", func(t *testing.T) {
		db := testDB()
		mux, _ := testMuxWithPath(t, db)
		body := strings.NewReader(`{"status":"COMPLETE"}`)
		req := httptest.NewRequest("PATCH", "/api/requirements/REQ-NOPE", body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", w.Code)
		}
	})

	t.Run("complete_sets_completed_date", func(t *testing.T) {
		db := testDB()
		mux, _ := testMuxWithPath(t, db)
		body := strings.NewReader(`{"status":"COMPLETE"}`)
		req := httptest.NewRequest("PATCH", "/api/requirements/REQ-MCP-002", body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		var resp apiRequirementDetailResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		today := time.Now().UTC().Format("2006-01-02")
		if resp.Requirement.CompletedDate != today {
			t.Errorf("completed_date = %s, want %s", resp.Requirement.CompletedDate, today)
		}
	})

	t.Run("uncomplete_clears_completed_date", func(t *testing.T) {
		db := testDB()
		mux, _ := testMuxWithPath(t, db)
		// First complete it
		body := strings.NewReader(`{"status":"COMPLETE"}`)
		req := httptest.NewRequest("PATCH", "/api/requirements/REQ-MCP-002", body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		// Then uncomplete
		body = strings.NewReader(`{"status":"PARTIAL"}`)
		req = httptest.NewRequest("PATCH", "/api/requirements/REQ-MCP-002", body)
		w = httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		var resp apiRequirementDetailResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Requirement.CompletedDate != "" {
			t.Errorf("completed_date = %s, want empty", resp.Requirement.CompletedDate)
		}
	})

	t.Run("persists_to_disk", func(t *testing.T) {
		db := testDB()
		mux, dbPath := testMuxWithPath(t, db)
		body := strings.NewReader(`{"status":"PARTIAL"}`)
		req := httptest.NewRequest("PATCH", "/api/requirements/REQ-MCP-002", body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		// Reload and verify
		reloaded, err := database.Load(dbPath)
		if err != nil {
			t.Fatalf("reload failed: %v", err)
		}
		r := reloaded.Get("REQ-MCP-002")
		if r == nil || r.Status != database.StatusPartial {
			t.Error("change not persisted to disk")
		}
	})
}

// TestAPIGraph validates GET /api/graph.
// REQ-API-004: Dependency graph endpoint.
func TestAPIGraph(t *testing.T) {
	rtmx.Req(t, "REQ-API-004")

	db := testDB()
	mux, _ := testMuxWithPath(t, db)

	t.Run("full_graph", func(t *testing.T) {
		w := apiGet(t, mux, "/api/graph")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		var resp apiGraphResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Metadata.TotalNodes != 5 {
			t.Errorf("total_nodes = %d, want 5", resp.Metadata.TotalNodes)
		}
		if resp.Metadata.TotalEdges < 3 {
			t.Errorf("total_edges = %d, want >= 3", resp.Metadata.TotalEdges)
		}
	})

	t.Run("node_blocked_field", func(t *testing.T) {
		w := apiGet(t, mux, "/api/graph")
		var resp apiGraphResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		for _, n := range resp.Nodes {
			if n.ID == "REQ-API-001" && !n.Blocked {
				t.Error("REQ-API-001 should be blocked (depends on incomplete MCP-001)")
			}
			if n.ID == "REQ-CLI-001" && n.Blocked {
				t.Error("REQ-CLI-001 should not be blocked (no deps)")
			}
		}
	})

	t.Run("category_filter", func(t *testing.T) {
		w := apiGet(t, mux, "/api/graph?category=CLI")
		var resp apiGraphResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		for _, n := range resp.Nodes {
			if n.Category != "CLI" {
				t.Errorf("expected only CLI nodes, got %s", n.Category)
			}
		}
	})

	t.Run("root_filter", func(t *testing.T) {
		w := apiGet(t, mux, "/api/graph?root=REQ-MCP-001")
		var resp apiGraphResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Metadata.TotalNodes < 2 {
			t.Errorf("subgraph should include MCP-001 and its dependents, got %d nodes", resp.Metadata.TotalNodes)
		}
	})

	t.Run("root_not_found_404", func(t *testing.T) {
		w := apiGet(t, mux, "/api/graph?root=REQ-NONEXISTENT")
		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", w.Code)
		}
	})

	t.Run("critical_path_present", func(t *testing.T) {
		w := apiGet(t, mux, "/api/graph")
		var resp apiGraphResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Metadata.CriticalPath == nil {
			t.Error("critical_path should not be null")
		}
	})

	t.Run("empty_graph", func(t *testing.T) {
		emptyDB := database.NewDatabase()
		cfg := &config.Config{}
		emptyMux := NewDashboardMux(emptyDB, cfg)
		w := apiGet(t, emptyMux, "/api/graph")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		var resp apiGraphResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.Nodes) != 0 {
			t.Errorf("nodes len = %d, want 0", len(resp.Nodes))
		}
	})
}

// TestAPIBacklog validates GET /api/backlog.
// REQ-API-005: Backlog endpoint with views.
func TestAPIBacklog(t *testing.T) {
	rtmx.Req(t, "REQ-API-005")

	db := testDB()
	mux, _ := testMuxWithPath(t, db)

	t.Run("default_view_all", func(t *testing.T) {
		w := apiGet(t, mux, "/api/backlog")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		var resp apiBacklogResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.View != "all" {
			t.Errorf("view = %s, want all", resp.View)
		}
		if len(resp.Sections) != 3 {
			t.Errorf("sections = %d, want 3 (Critical Path, Quick Wins, Remaining)", len(resp.Sections))
		}
		if resp.Summary.TotalIncomplete != 3 {
			t.Errorf("total_incomplete = %d, want 3", resp.Summary.TotalIncomplete)
		}
	})

	t.Run("view_blockers", func(t *testing.T) {
		w := apiGet(t, mux, "/api/backlog?view=blockers")
		var resp apiBacklogResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.View != "blockers" {
			t.Errorf("view = %s, want blockers", resp.View)
		}
		if len(resp.Sections) != 1 {
			t.Fatalf("sections = %d, want 1", len(resp.Sections))
		}
		// MCP-001 blocks both MCP-002 and API-001
		found := false
		for _, item := range resp.Sections[0].Items {
			if item.ReqID == "REQ-MCP-001" {
				found = true
				if item.TransitiveBlockCount < 2 {
					t.Errorf("MCP-001 transitive_blocks_count = %d, want >= 2", item.TransitiveBlockCount)
				}
			}
		}
		if !found {
			t.Error("REQ-MCP-001 should appear in blockers view")
		}
	})

	t.Run("view_quick_wins", func(t *testing.T) {
		w := apiGet(t, mux, "/api/backlog?view=quick-wins")
		var resp apiBacklogResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.View != "quick-wins" {
			t.Errorf("view = %s, want quick-wins", resp.View)
		}
	})

	t.Run("category_filter", func(t *testing.T) {
		w := apiGet(t, mux, "/api/backlog?category=MCP")
		var resp apiBacklogResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		// Only MCP-001 (PARTIAL) and MCP-002 (MISSING) are incomplete MCP reqs
		if resp.Summary.TotalIncomplete != 2 {
			t.Errorf("total_incomplete = %d, want 2", resp.Summary.TotalIncomplete)
		}
	})

	t.Run("limit_caps_items", func(t *testing.T) {
		w := apiGet(t, mux, "/api/backlog?limit=1")
		var resp apiBacklogResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		totalItems := 0
		for _, s := range resp.Sections {
			totalItems += len(s.Items)
		}
		if totalItems > 1 {
			t.Errorf("total items = %d, want <= 1", totalItems)
		}
	})

	t.Run("summary_counts_correct", func(t *testing.T) {
		w := apiGet(t, mux, "/api/backlog")
		var resp apiBacklogResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Summary.UnblockedCount+resp.Summary.BlockedCount != resp.Summary.TotalIncomplete {
			t.Error("unblocked + blocked should equal total_incomplete")
		}
	})

	t.Run("empty_backlog", func(t *testing.T) {
		allComplete := database.NewDatabase()
		r := &database.Requirement{
			ReqID: "REQ-DONE-001", Category: "DONE", Status: database.StatusComplete,
			Dependencies: make(database.StringSet), Blocks: make(database.StringSet),
		}
		_ = allComplete.Add(r)
		cfg := &config.Config{}
		m := NewDashboardMux(allComplete, cfg)
		w := apiGet(t, m, "/api/backlog")
		var resp apiBacklogResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Summary.TotalIncomplete != 0 {
			t.Errorf("total_incomplete = %d, want 0", resp.Summary.TotalIncomplete)
		}
	})
}

// TestAPIReleases validates GET /api/releases and /api/releases/:version.
// REQ-API-006: Release scope and gate endpoint.
func TestAPIReleases(t *testing.T) {
	rtmx.Req(t, "REQ-API-006")

	db := testDB()
	mux, _ := testMuxWithPath(t, db)

	t.Run("list_versions", func(t *testing.T) {
		w := apiGet(t, mux, "/api/releases")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		var resp apiReleasesResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.Versions) < 2 {
			t.Errorf("versions count = %d, want >= 2 (v1.0.0, v1.1.0, unversioned)", len(resp.Versions))
		}
	})

	t.Run("gate_pass_for_complete_version", func(t *testing.T) {
		w := apiGet(t, mux, "/api/releases")
		var resp apiReleasesResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		for _, v := range resp.Versions {
			if v.Version == "v1.0.0" {
				if v.GateStatus != "PASS" {
					t.Errorf("v1.0.0 gate = %s, want PASS", v.GateStatus)
				}
				if v.CompletionPct != 100.0 {
					t.Errorf("v1.0.0 completion = %f, want 100", v.CompletionPct)
				}
			}
		}
	})

	t.Run("gate_fail_for_incomplete_version", func(t *testing.T) {
		w := apiGet(t, mux, "/api/releases")
		var resp apiReleasesResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		for _, v := range resp.Versions {
			if v.Version == "v1.1.0" {
				if v.GateStatus != "FAIL" {
					t.Errorf("v1.1.0 gate = %s, want FAIL", v.GateStatus)
				}
			}
		}
	})

	t.Run("version_detail", func(t *testing.T) {
		w := apiGet(t, mux, "/api/releases/v1.0.0")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		var resp apiReleaseDetailResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Version != "v1.0.0" {
			t.Errorf("version = %s, want v1.0.0", resp.Version)
		}
		if resp.GateStatus != "PASS" {
			t.Errorf("gate = %s, want PASS", resp.GateStatus)
		}
		if resp.Summary.Total != 2 {
			t.Errorf("total = %d, want 2", resp.Summary.Total)
		}
	})

	t.Run("version_detail_with_failures", func(t *testing.T) {
		w := apiGet(t, mux, "/api/releases/v1.1.0")
		var resp apiReleaseDetailResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.GateStatus != "FAIL" {
			t.Errorf("gate = %s, want FAIL", resp.GateStatus)
		}
		if len(resp.GateFailures) == 0 {
			t.Error("gate_failures should be non-empty")
		}
	})

	t.Run("unknown_version_404", func(t *testing.T) {
		w := apiGet(t, mux, "/api/releases/v9.9.9")
		if w.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", w.Code)
		}
	})

	t.Run("unversioned_group", func(t *testing.T) {
		w := apiGet(t, mux, "/api/releases")
		var resp apiReleasesResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		found := false
		for _, v := range resp.Versions {
			if v.Label == "unversioned" {
				found = true
				if v.Total == 0 {
					t.Error("unversioned group should have items")
				}
			}
		}
		if !found {
			t.Error("unversioned group should be present")
		}
	})
}

// TestAPIAgentClaims validates GET /api/agents/claims.
// REQ-API-007: Agent activity and claims endpoint.
func TestAPIAgentClaims(t *testing.T) {
	rtmx.Req(t, "REQ-API-007")

	t.Run("empty_claims", func(t *testing.T) {
		db := testDB()
		mux, _ := testMuxWithPath(t, db)
		w := apiGet(t, mux, "/api/agents/claims")
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		var resp apiClaimsResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if len(resp.ActiveClaims) != 0 {
			t.Errorf("active_claims len = %d, want 0", len(resp.ActiveClaims))
		}
		if resp.Summary.TotalActive != 0 {
			t.Errorf("total_active = %d, want 0", resp.Summary.TotalActive)
		}
	})

	t.Run("with_active_claims", func(t *testing.T) {
		db := testDB()
		tmpDir := t.TempDir()
		rtmxDir := filepath.Join(tmpDir, ".rtmx")
		_ = os.MkdirAll(rtmxDir, 0o755)
		dbPath := filepath.Join(rtmxDir, "database.csv")
		_ = db.Save(dbPath)
		claimsDir := filepath.Join(rtmxDir, "claims")

		store, err := orchestration.NewClaimStore(claimsDir)
		if err != nil {
			t.Fatalf("NewClaimStore: %v", err)
		}
		_, _ = store.Claim("REQ-MCP-001", "claude-001")
		_, _ = store.Claim("REQ-MCP-002", "claude-002")

		cfg := &config.Config{}
		mux := NewDashboardMuxWithPath(db, cfg, dbPath)

		w := apiGet(t, mux, "/api/agents/claims")
		var resp apiClaimsResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Summary.TotalActive != 2 {
			t.Errorf("total_active = %d, want 2", resp.Summary.TotalActive)
		}
		if len(resp.Summary.Agents) != 2 {
			t.Errorf("agents = %d, want 2", len(resp.Summary.Agents))
		}
		// Verify requirement_text is populated
		for _, c := range resp.ActiveClaims {
			if c.RequirementText == "" {
				t.Errorf("claim for %s missing requirement_text", c.ReqID)
			}
		}
	})

	t.Run("stale_detection", func(t *testing.T) {
		db := testDB()
		tmpDir := t.TempDir()
		rtmxDir := filepath.Join(tmpDir, ".rtmx")
		_ = os.MkdirAll(rtmxDir, 0o755)
		dbPath := filepath.Join(rtmxDir, "database.csv")
		_ = db.Save(dbPath)
		claimsDir := filepath.Join(rtmxDir, "claims")

		store, _ := orchestration.NewClaimStore(claimsDir)
		_, _ = store.Claim("REQ-MCP-001", "claude-001")

		// Use a very short stale timeout for testing
		cfg := &config.Config{}
		mux := handleAPIAgentClaims(db, claimsDir, 1*time.Millisecond)
		time.Sleep(5 * time.Millisecond)

		req := httptest.NewRequest("GET", "/api/agents/claims", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		var resp apiClaimsResponse
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.Summary.StaleCount != 1 {
			t.Errorf("stale_count = %d, want 1", resp.Summary.StaleCount)
		}
		for _, c := range resp.ActiveClaims {
			if !c.Stale {
				t.Error("claim should be marked stale")
			}
		}
		_ = cfg // suppress unused
	})
}

// TestDashboardRequirementsList validates GET /partials/requirements.
// REQ-DASH-002: Requirements list partial with filter, sort, pagination.
func TestDashboardRequirementsList(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-002")

	db := testDBForDashboard(t)
	cfg := &config.Config{}
	mux := NewDashboardMuxWithPath(db, cfg, "")

	t.Run("returns_html_with_all_requirements", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/requirements", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		// All 5 requirements should be present
		for _, id := range []string{"REQ-CLI-001", "REQ-CLI-002", "REQ-MCP-001", "REQ-MCP-002", "REQ-API-001"} {
			if !strings.Contains(body, id) {
				t.Errorf("requirements list should contain %s", id)
			}
		}
	})

	t.Run("column_headers_present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/requirements", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		for _, header := range []string{"ID", "Status", "Priority", "Category", "Description", "Effort"} {
			if !strings.Contains(body, header) {
				t.Errorf("requirements list should contain column header %q", header)
			}
		}
	})

	t.Run("filter_controls_present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/requirements", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		if !strings.Contains(body, "<select") {
			t.Error("requirements list should contain select elements for filtering")
		}
		if !strings.Contains(body, "name=\"status\"") {
			t.Error("requirements list should contain status filter select")
		}
		if !strings.Contains(body, "name=\"category\"") {
			t.Error("requirements list should contain category filter select")
		}
		if !strings.Contains(body, "name=\"search\"") {
			t.Error("requirements list should contain search input")
		}
	})

	t.Run("filter_by_status", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/requirements?status=COMPLETE", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		if !strings.Contains(body, "REQ-CLI-001") {
			t.Error("COMPLETE filter should include REQ-CLI-001")
		}
		if !strings.Contains(body, "REQ-CLI-002") {
			t.Error("COMPLETE filter should include REQ-CLI-002")
		}
		if strings.Contains(body, "REQ-MCP-001") {
			t.Error("COMPLETE filter should exclude REQ-MCP-001 (PARTIAL)")
		}
		if strings.Contains(body, "REQ-API-001") {
			t.Error("COMPLETE filter should exclude REQ-API-001 (MISSING)")
		}
	})

	t.Run("filter_by_category", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/requirements?category=CLI", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		if !strings.Contains(body, "REQ-CLI-001") {
			t.Error("CLI filter should include REQ-CLI-001")
		}
		if strings.Contains(body, "REQ-MCP-001") {
			t.Error("CLI filter should exclude MCP requirements")
		}
	})

	t.Run("filter_by_search", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/requirements?search=server", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		if !strings.Contains(body, "REQ-MCP-001") {
			t.Error("search for 'server' should match REQ-MCP-001 (MCP server)")
		}
		if strings.Contains(body, "REQ-CLI-001") {
			t.Error("search for 'server' should not match REQ-CLI-001")
		}
	})

	t.Run("html_content_type", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/requirements", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		ct := rec.Header().Get("Content-Type")
		if !strings.Contains(ct, "text/html") {
			t.Errorf("Content-Type = %s, want text/html", ct)
		}
	})
}

// TestDashboardRequirementDetail validates GET /partials/detail/{reqID}.
// REQ-DASH-003: Detail partial with inline editing fields.
func TestDashboardRequirementDetail(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-003")

	db := testDBForDashboard(t)
	cfg := &config.Config{}
	mux := NewDashboardMuxWithPath(db, cfg, "")

	t.Run("returns_detail_html", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/detail/REQ-CLI-001", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		if !strings.Contains(body, "REQ-CLI-001") {
			t.Error("detail should contain the requirement ID")
		}
		if !strings.Contains(body, "Build CLI framework") {
			t.Error("detail should contain the requirement description")
		}
	})

	t.Run("shows_metadata_fields", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/detail/REQ-CLI-001", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		for _, field := range []string{"Category", "Priority", "Phase", "Effort"} {
			if !strings.Contains(body, field) {
				t.Errorf("detail should show metadata field %q", field)
			}
		}
	})

	t.Run("shows_dependency_sections", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/detail/REQ-CLI-001", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		if !strings.Contains(body, "Upstream") {
			t.Error("detail should contain Upstream Dependencies section")
		}
		if !strings.Contains(body, "Downstream") {
			t.Error("detail should contain Downstream Dependents section")
		}
	})

	t.Run("shows_downstream_dependents", func(t *testing.T) {
		// REQ-CLI-001 blocks REQ-CLI-002
		req := httptest.NewRequest("GET", "/partials/detail/REQ-CLI-001", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		if !strings.Contains(body, "REQ-CLI-002") {
			t.Error("REQ-CLI-001 detail should show REQ-CLI-002 as downstream dependent")
		}
	})

	t.Run("shows_upstream_dependencies", func(t *testing.T) {
		// REQ-CLI-002 depends on REQ-CLI-001
		req := httptest.NewRequest("GET", "/partials/detail/REQ-CLI-002", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		if !strings.Contains(body, "REQ-CLI-001") {
			t.Error("REQ-CLI-002 detail should show REQ-CLI-001 as upstream dependency")
		}
	})

	t.Run("contains_edit_form_elements", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/detail/REQ-CLI-001", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		if !strings.Contains(body, "name=\"status\"") {
			t.Error("detail should contain status select field")
		}
		if !strings.Contains(body, "name=\"assignee\"") {
			t.Error("detail should contain assignee input field")
		}
		if !strings.Contains(body, "Editable Fields") {
			t.Error("detail should contain Editable Fields section")
		}
	})

	t.Run("not_found_returns_404", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/detail/REQ-NONEXISTENT", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
	})
}

// TestDashboardGraph validates GET /partials/graph.
// REQ-DASH-004: Graph partial with D3 data.
func TestDashboardGraph(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-004")

	db := testDBForDashboard(t)
	cfg := &config.Config{}
	mux := NewDashboardMuxWithPath(db, cfg, "")

	t.Run("returns_graph_html", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/graph", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		if !strings.Contains(body, "graph-container") {
			t.Error("graph partial should contain graph container div")
		}
		if !strings.Contains(body, "Dependency Graph") {
			t.Error("graph partial should contain heading")
		}
	})

	t.Run("includes_node_edge_web_counts", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/graph", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		if !strings.Contains(body, "nodes") {
			t.Error("graph partial should include node count")
		}
		if !strings.Contains(body, "edges") {
			t.Error("graph partial should include edge count")
		}
		if !strings.Contains(body, "webs") {
			t.Error("graph partial should include web count")
		}
	})

	t.Run("includes_graph_json", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/graph", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		// GraphJSON is embedded as inline script data
		if !strings.Contains(body, "graphData") {
			t.Error("graph partial should embed graph data for D3")
		}
		// Should contain node IDs in the JSON
		if !strings.Contains(body, "REQ-CLI-001") {
			t.Error("graph JSON should contain requirement IDs")
		}
	})

	t.Run("category_filter_select_present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/graph", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		if !strings.Contains(body, "name=\"category\"") {
			t.Error("graph partial should contain category filter select")
		}
		if !strings.Contains(body, "All Categories") {
			t.Error("graph partial should have 'All Categories' option")
		}
	})
}

// TestDashboardKanban validates GET /partials/kanban.
// REQ-DASH-005: Kanban partial with drag-drop columns.
func TestDashboardKanban(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-005")

	db := testDBForDashboard(t)
	cfg := &config.Config{}
	mux := NewDashboardMuxWithPath(db, cfg, "")

	t.Run("returns_kanban_html_with_columns", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/kanban", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		// 4 column headers
		for _, label := range []string{"Not Started", "Missing", "Partial", "Complete"} {
			if !strings.Contains(body, label) {
				t.Errorf("kanban should contain column %q", label)
			}
		}
	})

	t.Run("cards_have_draggable_attribute", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/kanban", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		if !strings.Contains(body, `draggable="true"`) {
			t.Error("kanban cards should have draggable attribute")
		}
	})

	t.Run("blocked_indicator_present", func(t *testing.T) {
		// REQ-MCP-002 depends on REQ-MCP-001 (PARTIAL), so MCP-002 is blocked
		req := httptest.NewRequest("GET", "/partials/kanban", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		if !strings.Contains(body, "[Blocked]") {
			t.Error("kanban should show [Blocked] indicator for blocked requirements")
		}
	})

	t.Run("kanban_board_javascript_function", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/kanban", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		if !strings.Contains(body, "kanbanBoard()") {
			t.Error("kanban should include kanbanBoard() JavaScript function")
		}
	})

	t.Run("cards_contain_requirement_ids", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/kanban", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		for _, id := range []string{"REQ-CLI-001", "REQ-CLI-002", "REQ-MCP-001", "REQ-MCP-002", "REQ-API-001"} {
			if !strings.Contains(body, id) {
				t.Errorf("kanban should contain card for %s", id)
			}
		}
	})
}

// TestDashboardReleasePlanning validates GET /partials/releases.
// REQ-DASH-006: Release partial with version cards.
func TestDashboardReleasePlanning(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-006")

	t.Run("empty_state_when_no_versions", func(t *testing.T) {
		db := testDBForDashboard(t)
		cfg := &config.Config{}
		mux := NewDashboardMuxWithPath(db, cfg, "")

		req := httptest.NewRequest("GET", "/partials/releases", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		if !strings.Contains(body, "No versions assigned") {
			t.Error("releases partial with no sprints should show empty state message")
		}
		if !strings.Contains(body, "rtmx release assign") {
			t.Error("empty state should reference rtmx release assign command")
		}
	})

	t.Run("shows_version_cards_when_sprints_set", func(t *testing.T) {
		db := testDBForDashboardWithSprints(t)
		cfg := &config.Config{}
		mux := NewDashboardMuxWithPath(db, cfg, "")

		req := httptest.NewRequest("GET", "/partials/releases", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		body := rec.Body.String()

		if !strings.Contains(body, "v1.0.0") {
			t.Error("releases should show v1.0.0 version card")
		}
		if !strings.Contains(body, "v2.0.0") {
			t.Error("releases should show v2.0.0 version card")
		}
		if !strings.Contains(body, "Release Planning") {
			t.Error("releases partial should contain heading")
		}
	})

	t.Run("version_card_shows_gate_status", func(t *testing.T) {
		db := testDBForDashboardWithSprints(t)
		cfg := &config.Config{}
		mux := NewDashboardMuxWithPath(db, cfg, "")

		req := httptest.NewRequest("GET", "/partials/releases", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		body := rec.Body.String()
		// v1.0.0 has 2 COMPLETE reqs -> Gate: PASS
		if !strings.Contains(body, "Gate: PASS") {
			t.Error("v1.0.0 should show Gate: PASS")
		}
		// v2.0.0 has 1 PARTIAL + 1 MISSING -> Gate: FAIL
		if !strings.Contains(body, "Gate: FAIL") {
			t.Error("v2.0.0 should show Gate: FAIL")
		}
	})

	t.Run("version_card_shows_completion_percentage", func(t *testing.T) {
		db := testDBForDashboardWithSprints(t)
		cfg := &config.Config{}
		mux := NewDashboardMuxWithPath(db, cfg, "")

		req := httptest.NewRequest("GET", "/partials/releases", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		body := rec.Body.String()
		// v1.0.0: 2/2 = 100%
		if !strings.Contains(body, "100%") {
			t.Error("v1.0.0 should show 100% completion")
		}
	})
}

// testDBForDashboardWithSprints creates a database with sprint assignments for release testing.
func testDBForDashboardWithSprints(t *testing.T) *database.Database {
	t.Helper()
	db := database.NewDatabase()
	reqs := []*database.Requirement{
		{ReqID: "REQ-CLI-001", Category: "CLI", RequirementText: "Build CLI framework", Status: database.StatusComplete, Priority: database.PriorityP0, Phase: 1, EffortWeeks: 1.0, Sprint: "v1.0.0"},
		{ReqID: "REQ-CLI-002", Category: "CLI", RequirementText: "Status command", Status: database.StatusComplete, Priority: database.PriorityHigh, Phase: 1, EffortWeeks: 0.5, Sprint: "v1.0.0"},
		{ReqID: "REQ-MCP-001", Category: "MCP", RequirementText: "MCP server", Status: database.StatusPartial, Priority: database.PriorityP0, Phase: 2, EffortWeeks: 2.0, Sprint: "v2.0.0"},
		{ReqID: "REQ-MCP-002", Category: "MCP", RequirementText: "MCP tools", Status: database.StatusMissing, Priority: database.PriorityHigh, Phase: 2, EffortWeeks: 1.0, Sprint: "v2.0.0"},
		{ReqID: "REQ-API-001", Category: "API", RequirementText: "Requirements endpoint", Status: database.StatusMissing, Priority: database.PriorityP0, Phase: 3, EffortWeeks: 0.5},
	}
	for _, r := range reqs {
		r.Dependencies = make(database.StringSet)
		r.Blocks = make(database.StringSet)
		_ = db.Add(r)
	}
	db.Get("REQ-CLI-002").Dependencies.Add("REQ-CLI-001")
	db.Get("REQ-CLI-001").Blocks.Add("REQ-CLI-002")
	db.Get("REQ-MCP-002").Dependencies.Add("REQ-MCP-001")
	db.Get("REQ-MCP-001").Blocks.Add("REQ-MCP-002")
	return db
}

// TestDashboardHealthTrends validates GET /partials/health.
// REQ-DASH-007: Health partial with checks and stats.
func TestDashboardHealthTrends(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-007")

	db := testDBForDashboard(t)
	cfg := &config.Config{}
	mux := NewDashboardMuxWithPath(db, cfg, "")

	t.Run("returns_health_html", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/health", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		if !strings.Contains(body, "Health Dashboard") {
			t.Error("health partial should contain heading")
		}
	})

	t.Run("shows_completion_percentage", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/health", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		// 2 of 5 complete = 40%
		if !strings.Contains(body, "40%") {
			t.Error("health partial should show 40% completion (2 of 5)")
		}
		if !strings.Contains(body, "Completion Rate") {
			t.Error("health partial should label the completion rate")
		}
	})

	t.Run("health_checks_table_present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/health", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		if !strings.Contains(body, "Health Checks") {
			t.Error("health partial should contain Health Checks section")
		}
		if !strings.Contains(body, "Check") {
			t.Error("health partial should contain check table headers")
		}
	})

	t.Run("no_circular_dependencies_check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/health", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		if !strings.Contains(body, "No circular dependencies") {
			t.Error("health partial should contain 'No circular dependencies' check")
		}
		// No cycles in test DB, so it should pass
		if !strings.Contains(body, "PASS") {
			t.Error("health checks should show PASS for no-cycle check")
		}
	})

	t.Run("shows_blocked_count", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/health", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		if !strings.Contains(body, "Blocked Items") {
			t.Error("health partial should show blocked items count")
		}
	})
}
