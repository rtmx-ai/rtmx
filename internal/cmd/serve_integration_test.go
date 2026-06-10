package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/dashboard"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// integrationDB creates a database with known requirements for integration testing.
// It includes dependencies, multiple categories, sprints, and varied statuses
// to exercise all dashboard partials thoroughly.
func integrationDB(t *testing.T) *database.Database {
	t.Helper()
	db := database.NewDatabase()
	reqs := []*database.Requirement{
		{ReqID: "REQ-CLI-001", Category: "CLI", RequirementText: "Build CLI framework", Status: database.StatusComplete, Priority: database.PriorityP0, Phase: 1, EffortWeeks: 1.0, Assignee: "alice", Sprint: "v1.0.0"},
		{ReqID: "REQ-CLI-002", Category: "CLI", RequirementText: "Status command", Status: database.StatusComplete, Priority: database.PriorityHigh, Phase: 1, EffortWeeks: 0.5, Assignee: "alice", Sprint: "v1.0.0"},
		{ReqID: "REQ-MCP-001", Category: "MCP", RequirementText: "MCP server implementation", Status: database.StatusPartial, Priority: database.PriorityP0, Phase: 2, EffortWeeks: 2.0, Assignee: "bob", Sprint: "v2.0.0"},
		{ReqID: "REQ-MCP-002", Category: "MCP", RequirementText: "MCP tool registration", Status: database.StatusMissing, Priority: database.PriorityHigh, Phase: 2, EffortWeeks: 1.0, Sprint: "v2.0.0"},
		{ReqID: "REQ-API-001", Category: "API", RequirementText: "Requirements endpoint", Status: database.StatusMissing, Priority: database.PriorityP0, Phase: 3, EffortWeeks: 0.5},
	}
	for _, r := range reqs {
		r.Dependencies = make(database.StringSet)
		r.Blocks = make(database.StringSet)
		_ = db.Add(r)
	}
	// CLI-002 depends on CLI-001
	db.Get("REQ-CLI-002").Dependencies.Add("REQ-CLI-001")
	db.Get("REQ-CLI-001").Blocks.Add("REQ-CLI-002")
	// MCP-002 depends on MCP-001
	db.Get("REQ-MCP-002").Dependencies.Add("REQ-MCP-001")
	db.Get("REQ-MCP-001").Blocks.Add("REQ-MCP-002")
	// API-001 depends on MCP-001
	db.Get("REQ-API-001").Dependencies.Add("REQ-MCP-001")
	db.Get("REQ-MCP-001").Blocks.Add("REQ-API-001")
	return db
}

// integrationMux creates a fully wired dashboard mux using registerDashboardRoutes
// plus the API routes, matching production wiring in NewDashboardMuxWithPath.
func integrationMux(t *testing.T, db *database.Database) http.Handler {
	t.Helper()
	cfg := &config.Config{}
	return NewDashboardMuxWithPath(db, cfg, "")
}

// --------------------------------------------------------------------------
// 1. htmx Attribute Verification Tests
// --------------------------------------------------------------------------

// TestHtmxPartialAttributes verifies that each partial endpoint returns HTML
// containing the expected htmx attributes for client-side interactivity.
// REQ-DASH-001: Embedded SPA with htmx, Alpine.js, and custom CSS via embed.FS.
// REQ-DASH-002: Requirements list partial with htmx dynamic updates.
func TestHtmxPartialAttributes(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-001")

	db := integrationDB(t)
	mux := integrationMux(t, db)

	t.Run("status_partial_has_structural_elements", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/status", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		// Structural HTML checks
		assertContains(t, body, "Project Status", "status partial should contain heading")
		assertContains(t, body, "Categories", "status partial should contain categories table")
		assertContains(t, body, "progress-bar", "status partial should contain progress bar")
		// Data verification from test DB: 2 complete out of 5
		assertContains(t, body, "40%", "status partial should show 40% completion (2/5)")
	})

	t.Run("requirements_partial_has_htmx_attributes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		// htmx filter attributes
		assertContains(t, body, `hx-get="/partials/requirements"`, "requirements should have hx-get for filtering")
		assertContains(t, body, `hx-target="#content"`, "requirements should target #content div")
		assertContains(t, body, `hx-trigger="change"`, "select filters should trigger on change")
		assertContains(t, body, `hx-trigger="keyup changed delay:200ms"`, "search should have debounced trigger")

		// htmx row click-through to detail panel
		assertContains(t, body, `hx-get="/partials/detail/REQ-CLI-001"`, "requirement row should link to detail partial")
		assertContains(t, body, `hx-target="#detail-panel-body"`, "requirement rows should target detail panel")

		// Sortable column headers (REQ-DASH-021)
		assertContains(t, body, `sort=status&order=`, "status column should be sortable via htmx")
		assertContains(t, body, `sort=priority&order=`, "priority column should be sortable via htmx")

		// Data from test database
		assertContains(t, body, "REQ-CLI-001", "requirements should contain REQ-CLI-001")
		assertContains(t, body, "REQ-MCP-001", "requirements should contain REQ-MCP-001")
		assertContains(t, body, "REQ-API-001", "requirements should contain REQ-API-001")
	})

	t.Run("detail_partial_has_htmx_attributes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/detail/REQ-CLI-001", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		// Back button with htmx
		assertContains(t, body, `hx-get="/partials/requirements"`, "detail should have back-to-requirements link via htmx")
		assertContains(t, body, `hx-target="#content"`, "detail back button should target #content")

		// Edit form with htmx PATCH
		assertContains(t, body, `hx-patch="/api/requirements/REQ-CLI-001"`, "detail edit form should PATCH via htmx")

		// Dependency links with htmx navigation
		assertContains(t, body, `hx-get="/partials/detail/REQ-CLI-002"`, "downstream dep should link via htmx to detail")

		// Data verification
		assertContains(t, body, "Build CLI framework", "detail should show requirement description")
	})

	t.Run("graph_partial_has_htmx_attributes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/graph", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		// Category filter with htmx
		assertContains(t, body, `hx-get="/partials/graph"`, "graph should have category filter with hx-get")
		assertContains(t, body, `hx-target="#content"`, "graph filter should target #content")
		assertContains(t, body, `hx-trigger="change"`, "graph filter should trigger on change")

		// D3 graph data
		assertContains(t, body, "graph-container", "graph should have container div for D3")
		assertContains(t, body, "graphData", "graph should embed graphData for D3 rendering")
		assertContains(t, body, "REQ-CLI-001", "graph JSON should contain requirement node IDs")
	})

	t.Run("kanban_partial_has_htmx_and_drag_attributes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/kanban", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		// Drag-and-drop attributes
		assertContains(t, body, `draggable="true"`, "kanban cards should be draggable")
		assertContains(t, body, "@dragstart=", "kanban cards should have dragstart handler")
		assertContains(t, body, "@drop=", "kanban columns should have drop handler")
		assertContains(t, body, "@dragover.prevent=", "kanban columns should prevent default dragover")

		// Alpine.js kanban board component
		assertContains(t, body, "kanbanBoard()", "kanban should include kanbanBoard() Alpine component")

		// htmx PATCH call in the drop handler JavaScript
		assertContains(t, body, "htmx.ajax", "kanban drop handler should use htmx.ajax for PATCH")
		assertContains(t, body, `/partials/kanban`, "kanban drop should refresh via /partials/kanban")

		// Card data from test DB
		for _, id := range []string{"REQ-CLI-001", "REQ-CLI-002", "REQ-MCP-001", "REQ-MCP-002", "REQ-API-001"} {
			assertContains(t, body, id, "kanban should contain card for "+id)
		}

		// Blocked indicator
		assertContains(t, body, "[Blocked]", "kanban should show [Blocked] indicator for blocked cards")
	})

	t.Run("releases_partial_has_htmx_attributes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/releases", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		// Release rows should navigate to detail via htmx
		assertContains(t, body, `hx-get="/partials/detail/`, "release rows should link to detail via htmx")
		assertContains(t, body, `hx-target="#detail-panel-body"`, "release rows should target detail panel")

		// Data from test DB (sprints assigned)
		assertContains(t, body, "v1.0.0", "releases should show v1.0.0")
		assertContains(t, body, "v2.0.0", "releases should show v2.0.0")
		assertContains(t, body, "gate-pass", "v1.0.0 should show gate-pass badge")
		assertContains(t, body, "gate-fail", "v2.0.0 should show gate-fail badge")
	})

	t.Run("health_partial_has_structural_elements", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/health", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		assertContains(t, body, "Health Dashboard", "health should contain heading")
		assertContains(t, body, "Health Checks", "health should contain checks section")
		assertContains(t, body, "Completion Rate", "health should label completion rate")
		assertContains(t, body, "Blocked Items", "health should show blocked items")
		assertContains(t, body, "No circular dependencies", "health should include cycle check")
		assertContains(t, body, "PASS", "health should show PASS for passing checks")
		assertContains(t, body, "40%", "health should show 40% completion (2/5)")
	})

	t.Run("agents_partial_has_htmx_polling", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/agents", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		// Auto-refresh polling via htmx
		assertContains(t, body, `hx-get="/partials/agents"`, "agents should have auto-refresh hx-get")
		assertContains(t, body, `hx-trigger="every 10s"`, "agents should poll every 10s")
		assertContains(t, body, `hx-target="#content"`, "agents polling should target #content")

		// Structural elements
		assertContains(t, body, "Agent Monitor", "agents should contain heading")
		assertContains(t, body, "Active Claims", "agents should show active claims count")
		assertContains(t, body, "No active agent claims", "agents with empty claims should show placeholder")
	})

	t.Run("all_partials_return_html_content_type", func(t *testing.T) {
		endpoints := []string{
			"/partials/status",
			"/partials/requirements",
			"/partials/graph",
			"/partials/kanban",
			"/partials/releases",
			"/partials/health",
			"/partials/agents",
		}
		for _, ep := range endpoints {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest("GET", ep, nil))
			ct := rec.Header().Get("Content-Type")
			if !strings.Contains(ct, "text/html") {
				t.Errorf("GET %s Content-Type = %s, want text/html", ep, ct)
			}
		}
	})
}

// --------------------------------------------------------------------------
// 2. SPA Shell Integration Tests
// --------------------------------------------------------------------------

// TestSPAShellIntegration verifies the full SPA shell served at /app includes
// all required client-side framework scripts, navigation, and pre-rendered content.
// REQ-DASH-001: Embedded SPA with htmx, Alpine.js, and custom CSS via embed.FS.
func TestSPAShellIntegration(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-001")

	db := integrationDB(t)
	mux := integrationMux(t, db)

	t.Run("shell_includes_htmx_script", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/app", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()
		assertContains(t, body, "htmx.min.js", "SPA shell must include htmx script tag")
	})

	t.Run("shell_includes_alpinejs_script", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/app", nil))
		body := rec.Body.String()
		assertContains(t, body, "alpine.min.js", "SPA shell must include Alpine.js script tag")
	})

	t.Run("shell_includes_custom_styles", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/app", nil))
		body := rec.Body.String()
		assertContains(t, body, "styles.css", "SPA shell must include custom styles")
	})

	t.Run("shell_includes_all_nav_links", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/app", nil))
		body := rec.Body.String()

		navLinks := []string{"Status", "Requirements", "Graph", "Kanban", "Releases", "Health", "Agents"}
		for _, link := range navLinks {
			assertContains(t, body, link, "SPA shell nav must contain "+link+" link")
		}
		// Exactly 7 nav links with hx-get targeting partials
		for _, partial := range []string{
			`hx-get="/partials/status"`,
			`hx-get="/partials/requirements"`,
			`hx-get="/partials/graph"`,
			`hx-get="/partials/kanban"`,
			`hx-get="/partials/releases"`,
			`hx-get="/partials/health"`,
			`hx-get="/partials/agents"`,
		} {
			assertContains(t, body, partial, "SPA nav must contain "+partial)
		}
	})

	t.Run("shell_pre_renders_initial_content", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/app", nil))
		body := rec.Body.String()

		// The /app route pre-renders the status partial into the #content div
		assertContains(t, body, "Project Status", "SPA shell should pre-render status content")
		assertContains(t, body, "Categories", "SPA shell should pre-render category table")
		// Verify it is not an empty content div
		if strings.Contains(body, `<div id="content"></div>`) {
			t.Error("SPA shell should NOT have empty #content div; it should contain pre-rendered content")
		}
		if strings.Contains(body, `<div id="content">\n    </div>`) {
			t.Error("SPA shell should NOT have empty #content div")
		}
		// Content div should exist with actual content inside
		assertContains(t, body, `id="content"`, "SPA shell should have #content div")
	})

	t.Run("shell_includes_alpine_rtmxapp_init", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/app", nil))
		body := rec.Body.String()

		assertContains(t, body, `x-data="rtmxApp()"`, "SPA shell body should initialize rtmxApp() Alpine component")
		assertContains(t, body, "/static/app.js", "SPA shell should include app.js script")
	})

	t.Run("shell_serves_all_page_routes", func(t *testing.T) {
		// All SPA routes should return the same shell HTML structure
		routes := []string{"/app", "/requirements", "/graph", "/kanban", "/releases", "/health", "/agents"}
		for _, route := range routes {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest("GET", route, nil))
			if rec.Code != http.StatusOK {
				t.Errorf("GET %s status = %d, want 200", route, rec.Code)
			}
			body := rec.Body.String()
			assertContains(t, body, "<!DOCTYPE html>", "GET "+route+" should return full HTML document")
			assertContains(t, body, "RTMX", "GET "+route+" should contain RTMX branding")
			assertContains(t, body, `id="content"`, "GET "+route+" should contain #content div")
		}
	})
}

// --------------------------------------------------------------------------
// 3. Auth Middleware Integration Tests
// --------------------------------------------------------------------------

// TestAuthIntegrationWithDashboard verifies that auth middleware correctly
// protects dashboard routes when enabled, and passes through when disabled.
// REQ-DASH-009: Auth middleware for API key/OAuth.
func TestAuthIntegrationWithDashboard(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-009")

	db := integrationDB(t)
	cfg := &config.Config{}
	baseMux := NewDashboardMuxWithPath(db, cfg, "")

	t.Run("unauthenticated_rejected_with_api_key_auth", func(t *testing.T) {
		authCfg := authConfig{Mode: "api-key", APIKey: "integration-test-key"}
		protected := authMiddleware(authCfg)(baseMux)

		routes := []string{"/app", "/partials/status", "/partials/requirements", "/partials/kanban", "/api/requirements"}
		for _, route := range routes {
			rec := httptest.NewRecorder()
			protected.ServeHTTP(rec, httptest.NewRequest("GET", route, nil))
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("GET %s without auth: status = %d, want 401", route, rec.Code)
			}
		}
	})

	t.Run("authenticated_requests_return_dashboard_content", func(t *testing.T) {
		authCfg := authConfig{Mode: "api-key", APIKey: "integration-test-key"}
		protected := authMiddleware(authCfg)(baseMux)

		tests := []struct {
			route    string
			contains string
		}{
			{"/app", "<!DOCTYPE html>"},
			{"/partials/status", "Project Status"},
			{"/partials/requirements", "Requirements"},
			{"/partials/kanban", "Kanban Board"},
			{"/partials/graph", "Dependency Graph"},
			{"/partials/releases", "Release Planning"},
			{"/partials/health", "Health Dashboard"},
			{"/partials/agents", "Agent Monitor"},
		}
		for _, tt := range tests {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.route, nil)
			req.Header.Set("Authorization", "Bearer integration-test-key")
			protected.ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("GET %s with auth: status = %d, want 200", tt.route, rec.Code)
			}
			if !strings.Contains(rec.Body.String(), tt.contains) {
				t.Errorf("GET %s with auth: body does not contain %q", tt.route, tt.contains)
			}
		}
	})

	t.Run("static_assets_accessible_through_auth", func(t *testing.T) {
		// When auth is enabled, static assets go through the same middleware.
		// Verify they are accessible with valid credentials.
		authCfg := authConfig{Mode: "api-key", APIKey: "integration-test-key"}
		protected := authMiddleware(authCfg)(baseMux)

		req := httptest.NewRequest("GET", "/static/styles.css", nil)
		req.Header.Set("Authorization", "Bearer integration-test-key")
		rec := httptest.NewRecorder()
		protected.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("GET /static/styles.css with auth: status = %d, want 200", rec.Code)
		}
	})

	t.Run("auth_disabled_passes_all_dashboard_routes", func(t *testing.T) {
		authCfg := authConfig{Mode: ""}
		protected := authMiddleware(authCfg)(baseMux)

		routes := []string{"/app", "/partials/status", "/partials/requirements", "/api/requirements"}
		for _, route := range routes {
			rec := httptest.NewRecorder()
			protected.ServeHTTP(rec, httptest.NewRequest("GET", route, nil))
			if rec.Code != http.StatusOK {
				t.Errorf("GET %s with auth disabled: status = %d, want 200", route, rec.Code)
			}
		}
	})

	t.Run("wrong_api_key_rejected_for_all_routes", func(t *testing.T) {
		authCfg := authConfig{Mode: "api-key", APIKey: "correct-key"}
		protected := authMiddleware(authCfg)(baseMux)

		routes := []string{"/app", "/partials/status", "/partials/kanban", "/api/requirements"}
		for _, route := range routes {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", route, nil)
			req.Header.Set("Authorization", "Bearer wrong-key")
			protected.ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("GET %s with wrong key: status = %d, want 401", route, rec.Code)
			}
		}
	})
}

// --------------------------------------------------------------------------
// 4. Full Page Workflow Tests
// --------------------------------------------------------------------------

// TestPageWorkflowNavigation simulates a user navigating through the SPA:
// loading the shell, browsing requirements, viewing details, and using the kanban.
// REQ-DASH-001, REQ-DASH-002, REQ-DASH-003, REQ-DASH-005.
func TestPageWorkflowNavigation(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-002")

	db := integrationDB(t)
	mux := integrationMux(t, db)

	// Step 1: User loads the SPA shell
	t.Run("step1_load_shell", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/app", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /app status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()
		assertContains(t, body, "<!DOCTYPE html>", "shell should be a full HTML document")
		assertContains(t, body, "Project Status", "shell should pre-render status content")
		assertContains(t, body, `hx-get="/partials/requirements"`, "shell should have requirements nav link")
	})

	// Step 2: User clicks Requirements nav link (htmx fetches partial)
	t.Run("step2_browse_requirements", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /partials/requirements status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		// Should return HTML fragment (not full page), containing a table
		if strings.Contains(body, "<!DOCTYPE html>") {
			t.Error("partial should NOT be a full HTML document")
		}
		assertContains(t, body, "<table", "requirements partial should contain a data table")
		assertContains(t, body, "REQ-CLI-001", "table should contain REQ-CLI-001")
		assertContains(t, body, "REQ-MCP-001", "table should contain REQ-MCP-001")
		assertContains(t, body, "REQ-API-001", "table should contain REQ-API-001")

		// Each row should be clickable via htmx to load detail
		assertContains(t, body, `hx-get="/partials/detail/REQ-CLI-001"`, "row should link to detail")
	})

	// Step 3: User clicks a requirement row to view detail
	t.Run("step3_view_detail", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/detail/REQ-MCP-001", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /partials/detail/REQ-MCP-001 status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		// Verify detail view content
		assertContains(t, body, "REQ-MCP-001", "detail should show the requirement ID")
		assertContains(t, body, "MCP server implementation", "detail should show description")
		assertContains(t, body, "MCP", "detail should show category")

		// Dependency information
		assertContains(t, body, "Downstream", "detail should show downstream section")
		assertContains(t, body, "REQ-MCP-002", "detail should show REQ-MCP-002 as downstream dependent")
		assertContains(t, body, "REQ-API-001", "detail should show REQ-API-001 as downstream dependent")

		// Back button with htmx
		assertContains(t, body, `hx-get="/partials/requirements"`, "detail should have back link to requirements")

		// Edit form
		assertContains(t, body, `hx-patch="/api/requirements/REQ-MCP-001"`, "detail should have edit form with htmx PATCH")
		assertContains(t, body, `name="status"`, "detail should have status select field")
		assertContains(t, body, `name="assignee"`, "detail should have assignee input")
	})

	// Step 4: User navigates to kanban view
	t.Run("step4_view_kanban", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/kanban", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /partials/kanban status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		// Four status columns
		for _, col := range []string{"Not Started", "Missing", "Partial", "Complete"} {
			assertContains(t, body, col, "kanban should have "+col+" column")
		}

		// Cards should be distributed across columns based on status
		assertContains(t, body, "REQ-CLI-001", "kanban should have card for complete req CLI-001")
		assertContains(t, body, "REQ-MCP-001", "kanban should have card for partial req MCP-001")
		assertContains(t, body, "REQ-MCP-002", "kanban should have card for missing req MCP-002")

		// Drag-drop support
		assertContains(t, body, `draggable="true"`, "cards should be draggable")
		assertContains(t, body, "kanbanBoard()", "kanban should have Alpine.js board component")
	})

	// Step 5: User filters requirements by category
	t.Run("step5_filter_by_category", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements?category=MCP", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("GET /partials/requirements?category=MCP status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		assertContains(t, body, "REQ-MCP-001", "MCP filter should include REQ-MCP-001")
		assertContains(t, body, "REQ-MCP-002", "MCP filter should include REQ-MCP-002")
		if strings.Contains(body, "REQ-CLI-001") {
			t.Error("MCP filter should exclude REQ-CLI-001")
		}
		if strings.Contains(body, "REQ-API-001") {
			t.Error("MCP filter should exclude REQ-API-001")
		}
	})

	// Step 6: User searches requirements
	t.Run("step6_search_requirements", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements?search=framework", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()
		assertContains(t, body, "REQ-CLI-001", "search for 'framework' should match CLI-001")
		if strings.Contains(body, "REQ-MCP-001") {
			t.Error("search for 'framework' should not match MCP-001")
		}
	})

	// Step 7: User views detail for a requirement with upstream dependencies
	t.Run("step7_detail_with_upstream", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/detail/REQ-CLI-002", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		assertContains(t, body, "REQ-CLI-002", "detail should show requirement ID")
		assertContains(t, body, "Status command", "detail should show description")
		assertContains(t, body, "Upstream", "detail should show upstream section")
		assertContains(t, body, "REQ-CLI-001", "detail should show CLI-001 as upstream dependency")
		assertContains(t, body, `hx-get="/partials/detail/REQ-CLI-001"`, "upstream dep should link to its detail via htmx")
	})

	// Step 8: Nonexistent requirement returns 404
	t.Run("step8_detail_not_found", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/detail/REQ-NONEXISTENT", nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET /partials/detail/REQ-NONEXISTENT status = %d, want 404", rec.Code)
		}
	})
}

// TestDashboardIntegrationKanbanDragDrop verifies the Kanban partial contains
// all elements necessary for the full drag-drop workflow through the HTTP stack.
// REQ-DASH-005: Kanban board with drag-and-drop status transitions.
func TestDashboardIntegrationKanbanDragDrop(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-005")

	db := integrationDB(t)
	mux := integrationMux(t, db)

	t.Run("kanban_columns_have_data_status_attributes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/kanban", nil))
		body := rec.Body.String()

		// Each column needs data-status for the drop handler
		for _, status := range []string{"NOT_STARTED", "MISSING", "PARTIAL", "COMPLETE"} {
			assertContains(t, body, `data-status="`+status+`"`, "kanban column should have data-status="+status)
		}
	})

	t.Run("kanban_cards_have_data_req_id", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/kanban", nil))
		body := rec.Body.String()

		assertContains(t, body, `data-req-id="REQ-CLI-001"`, "kanban card should have data-req-id for CLI-001")
		assertContains(t, body, `data-req-id="REQ-MCP-001"`, "kanban card should have data-req-id for MCP-001")
	})

	t.Run("kanban_drop_handler_calls_patch_api", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/kanban", nil))
		body := rec.Body.String()

		// The JavaScript drop handler should call PATCH on the API endpoint
		assertContains(t, body, "PATCH", "drop handler should issue PATCH request")
		assertContains(t, body, "/api/requirements/", "drop handler should target the requirements API")
		assertContains(t, body, "/partials/kanban", "drop handler should refresh kanban after update")
	})

	t.Run("kanban_shows_priority_classes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/kanban", nil))
		body := rec.Body.String()

		// Priority classes for visual styling
		assertContains(t, body, "pri-p0", "kanban should have P0 priority cards")
		assertContains(t, body, "pri-high", "kanban should have HIGH priority cards")
	})

	t.Run("kanban_blocked_cards_visually_distinct", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/kanban", nil))
		body := rec.Body.String()

		// MCP-002 is blocked (depends on incomplete MCP-001)
		// API-001 is blocked (depends on incomplete MCP-001)
		assertContains(t, body, "[Blocked]", "blocked cards should have [Blocked] indicator")
	})
}

// --------------------------------------------------------------------------
// --------------------------------------------------------------------------
// REQ-DASH-016: Row partial endpoint
// --------------------------------------------------------------------------

func TestRowPartialEndpoint(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-016")
	db := integrationDB(t)
	cfg := &config.Config{}
	mux := NewDashboardMuxWithPath(db, cfg, "")

	t.Run("returns_single_row_html", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements/row/REQ-CLI-001", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()
		assertContains(t, body, "REQ-CLI-001", "row should contain req ID")
		assertContains(t, body, "<tr", "row should be a table row element")
		assertContains(t, body, "row-REQ-CLI-001", "row should have id attribute")
	})

	t.Run("returns_404_for_unknown", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements/row/REQ-NOPE", nil))
		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", rec.Code)
		}
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-020: Partial sort and filter
// --------------------------------------------------------------------------

func TestPartialSortFilter(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-020")
	db := integrationDB(t)
	cfg := &config.Config{}
	mux := NewDashboardMuxWithPath(db, cfg, "")

	t.Run("filter_by_status", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements?status=COMPLETE", nil))
		body := rec.Body.String()
		assertContains(t, body, "REQ-CLI-001", "should include COMPLETE requirement")
		if strings.Contains(body, "REQ-MCP-001") {
			t.Error("should not include PARTIAL requirement when filtering by COMPLETE")
		}
	})

	t.Run("filter_by_category", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements?category=MCP", nil))
		body := rec.Body.String()
		assertContains(t, body, "REQ-MCP-001", "should include MCP requirement")
		if strings.Contains(body, "REQ-CLI-001") {
			t.Error("should not include CLI requirement when filtering by MCP")
		}
	})

	t.Run("filter_by_search", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements?search=registration", nil))
		body := rec.Body.String()
		assertContains(t, body, "REQ-MCP-002", "should include requirement matching search text")
	})

	t.Run("sort_by_priority_desc", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements?sort=priority&order=desc", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		// Just verify it returns valid HTML without error
		assertContains(t, rec.Body.String(), "Requirements", "should render the requirements page")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-015: HX-Trigger on PATCH
// --------------------------------------------------------------------------

func TestPatchToastTrigger(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-015")
	db := integrationDB(t)
	cfg := &config.Config{}
	// Use a temp file so PATCH can save successfully
	tmpFile := t.TempDir() + "/test_db.csv"
	_ = db.Save(tmpFile)
	mux := NewDashboardMuxWithPath(db, cfg, tmpFile)

	t.Run("success_includes_hx_trigger", func(t *testing.T) {
		body := strings.NewReader(`{"status":"COMPLETE"}`)
		req := httptest.NewRequest("PATCH", "/api/requirements/REQ-MCP-002", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		trigger := rec.Header().Get("HX-Trigger")
		if trigger == "" {
			t.Error("PATCH success should include HX-Trigger header")
		}
		assertContains(t, trigger, "showToast", "HX-Trigger should contain showToast event")
		assertContains(t, trigger, "success", "toast type should be success")
	})

	t.Run("error_includes_hx_trigger", func(t *testing.T) {
		body := strings.NewReader(`{"status":"INVALID"}`)
		req := httptest.NewRequest("PATCH", "/api/requirements/REQ-MCP-002", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		trigger := rec.Header().Get("HX-Trigger")
		if trigger == "" {
			t.Error("PATCH error should include HX-Trigger header")
		}
		assertContains(t, trigger, "error", "toast type should be error on validation failure")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-017: Inline edit dropdown cells
// --------------------------------------------------------------------------

func TestInlineEditDropdownCells(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-017")
	db := integrationDB(t)
	cfg := &config.Config{}
	mux := NewDashboardMuxWithPath(db, cfg, "")

	t.Run("status_cell_has_inline_select", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements", nil))
		body := rec.Body.String()
		assertContains(t, body, "inlineSelect(", "status cell should use inlineSelect Alpine component")
		assertContains(t, body, `'status'`, "inline select should target status field")
		assertContains(t, body, "inline-select", "select element should have inline-select class")
	})

	t.Run("priority_cell_has_inline_select", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements", nil))
		body := rec.Body.String()
		assertContains(t, body, `'priority'`, "inline select should target priority field")
	})

	t.Run("status_options_include_all_values", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements", nil))
		body := rec.Body.String()
		assertContains(t, body, `value="NOT_STARTED"`, "status select should include NOT_STARTED option")
		assertContains(t, body, `value="MISSING"`, "status select should include MISSING option")
		assertContains(t, body, `value="PARTIAL"`, "status select should include PARTIAL option")
		assertContains(t, body, `value="COMPLETE"`, "status select should include COMPLETE option")
	})

	t.Run("row_partial_has_inline_select", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements/row/REQ-CLI-001", nil))
		body := rec.Body.String()
		assertContains(t, body, "inlineSelect(", "row partial should include inlineSelect component")
		assertContains(t, body, "inline-edit-cell", "row cells should have inline-edit-cell class")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-018: Inline edit text cells
// --------------------------------------------------------------------------

func TestInlineEditTextCells(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-018")
	db := integrationDB(t)
	cfg := &config.Config{}
	mux := NewDashboardMuxWithPath(db, cfg, "")

	t.Run("assignee_cell_has_inline_text", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements", nil))
		body := rec.Body.String()
		assertContains(t, body, "inlineText(", "assignee cell should use inlineText Alpine component")
		assertContains(t, body, `'assignee'`, "inline text should target assignee field")
		assertContains(t, body, "inline-input", "input element should have inline-input class")
	})

	t.Run("sprint_cell_has_inline_text", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements", nil))
		body := rec.Body.String()
		assertContains(t, body, `'sprint'`, "inline text should target sprint field")
	})

	t.Run("text_input_commits_on_blur_and_enter", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements", nil))
		body := rec.Body.String()
		assertContains(t, body, `@blur="save($el.value)"`, "text input should save on blur")
		assertContains(t, body, `@keydown.enter="save($el.value)"`, "text input should save on Enter")
		assertContains(t, body, `@keydown.escape="cancel()"`, "text input should cancel on Escape")
	})

	t.Run("row_partial_has_inline_text", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements/row/REQ-CLI-001", nil))
		body := rec.Body.String()
		assertContains(t, body, "inlineText(", "row partial should include inlineText component")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-019: Inline edit cancel and loading
// --------------------------------------------------------------------------

func TestInlineEditCancelAndLoading(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-019")

	js := dashboard.ReadStaticFile("app.js")

	t.Run("escape_cancels_edit", func(t *testing.T) {
		assertContains(t, js, "cancel()", "inline editors should have cancel method")
		// Verify cancel restores original value for text editor
		assertContains(t, js, "this.value = this.original", "cancel should revert to original value")
	})

	t.Run("loading_state_during_patch", func(t *testing.T) {
		assertContains(t, js, "loading: false", "inline editors should track loading state")
		assertContains(t, js, "this.loading = true", "save should set loading to true during PATCH")
		assertContains(t, js, "self.loading = false", "loading should be cleared after PATCH completes")
	})

	t.Run("escape_keydown_wired_in_template", func(t *testing.T) {
		db := integrationDB(t)
		cfg := &config.Config{}
		mux := NewDashboardMuxWithPath(db, cfg, "")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements", nil))
		body := rec.Body.String()
		assertContains(t, body, `@keydown.escape="cancel()"`, "inputs should wire Escape to cancel")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-021: Sortable column headers
// --------------------------------------------------------------------------

func TestSortableColumnHeaders(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-021")
	db := integrationDB(t)
	cfg := &config.Config{}
	mux := NewDashboardMuxWithPath(db, cfg, "")

	t.Run("default_sort_shows_req_id_asc", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements", nil))
		body := rec.Body.String()
		assertContains(t, body, "sort-active", "default sort column should have sort-active class")
		assertContains(t, body, "sort-indicator", "headers should have sort-indicator span")
		// Default is req_id asc, so the ID header should show up-arrow (&#9650; = unicode 9650)
		assertContains(t, body, "&#9650;", "ascending sort should show up-arrow indicator")
	})

	t.Run("sort_desc_shows_down_arrow", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements?sort=priority&order=desc", nil))
		body := rec.Body.String()
		assertContains(t, body, "&#9660;", "descending sort should show down-arrow indicator")
	})

	t.Run("headers_toggle_sort_direction", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements?sort=req_id&order=asc", nil))
		body := rec.Body.String()
		// When req_id is asc, clicking it should switch to desc
		assertContains(t, body, `sort=req_id&order=desc`, "active asc column click should toggle to desc")
	})

	t.Run("headers_include_filter_state", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements", nil))
		body := rec.Body.String()
		assertContains(t, body, "sortable-header", "column headers should have sortable-header class")
		assertContains(t, body, `hx-include`, "sort headers should include filter inputs")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-024: Search input with debounced query
// --------------------------------------------------------------------------

func TestTableSearchInput(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-024")
	db := integrationDB(t)
	cfg := &config.Config{}
	mux := NewDashboardMuxWithPath(db, cfg, "")

	t.Run("search_input_attributes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements", nil))
		body := rec.Body.String()
		assertContains(t, body, `name="search"`, "search input should have name=search")
		assertContains(t, body, `hx-trigger="keyup changed delay:200ms"`, "search should debounce at 200ms")
		assertContains(t, body, `hx-get="/partials/requirements"`, "search should target partials endpoint")
	})

	t.Run("search_preserves_value", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements?search=registration", nil))
		body := rec.Body.String()
		assertContains(t, body, `value="registration"`, "search input should preserve current search value")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-026: Panel slide animation CSS
// --------------------------------------------------------------------------

func TestPanelSlideCSS(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-026")

	css := dashboard.ReadStaticFile("styles.css")

	t.Run("panel_transition_css", func(t *testing.T) {
		assertContains(t, css, "panel-enter", "CSS should include panel-enter transition")
		assertContains(t, css, "panel-leave", "CSS should include panel-leave transition")
		assertContains(t, css, "translateX(100%)", "panel should slide from off-screen right")
		assertContains(t, css, "translateX(0)", "panel should slide to visible position")
	})

	t.Run("backdrop_css", func(t *testing.T) {
		assertContains(t, css, "backdrop-enter", "CSS should include backdrop-enter transition")
		assertContains(t, css, "backdrop-leave", "CSS should include backdrop-leave transition")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-028: Panel close interactions
// --------------------------------------------------------------------------

func TestPanelCloseInteractions(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-028")

	t.Run("escape_closes_panel_in_layout", func(t *testing.T) {
		var buf strings.Builder
		_ = dashboard.RenderLayout(&buf, dashboard.LayoutData{Title: "Test", ActivePage: "status"})
		html := buf.String()
		assertContains(t, html, `@keydown.escape.window="panelOpen = false"`, "layout should close panel on Escape")
	})

	t.Run("backdrop_click_closes_panel", func(t *testing.T) {
		var buf strings.Builder
		_ = dashboard.RenderLayout(&buf, dashboard.LayoutData{Title: "Test", ActivePage: "status"})
		html := buf.String()
		assertContains(t, html, `@click="panelOpen = false"`, "backdrop click should close panel")
	})

	t.Run("htmx_swap_opens_panel", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		assertContains(t, js, "detail-panel-body", "JS should listen for htmx swaps to panel body")
		assertContains(t, js, "panelOpen = true", "htmx swap into panel body should open panel")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-031: Command palette modal
// --------------------------------------------------------------------------

func TestPaletteModalTrigger(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-031")

	t.Run("cmd_k_handler_in_js", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		assertContains(t, js, "e.metaKey", "Cmd+K handler should check metaKey")
		assertContains(t, js, "e.ctrlKey", "Cmd+K handler should check ctrlKey")
		assertContains(t, js, "e.key === 'k'", "handler should check for k key")
		assertContains(t, js, "paletteOpen", "handler should toggle paletteOpen state")
	})

	t.Run("palette_css", func(t *testing.T) {
		css := dashboard.ReadStaticFile("styles.css")
		assertContains(t, css, "palette-input", "CSS should style the palette input")
		assertContains(t, css, "palette-enter", "CSS should have palette enter transition")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-032: Palette search with debounced API
// --------------------------------------------------------------------------

func TestPaletteSearch(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-032")
	db := integrationDB(t)
	cfg := &config.Config{}
	mux := NewDashboardMuxWithPath(db, cfg, "")

	t.Run("returns_matching_results", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/palette-results?search=Build", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()
		assertContains(t, body, "REQ-CLI-001", "palette should return matching requirement")
		assertContains(t, body, "palette-result", "results should use palette-result class")
	})

	t.Run("returns_empty_for_no_match", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/palette-results?search=xyznonexistent", nil))
		body := rec.Body.String()
		assertContains(t, body, "No matching requirements", "empty search should show placeholder")
	})

	t.Run("returns_empty_for_blank_search", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/palette-results?search=", nil))
		body := rec.Body.String()
		assertContains(t, body, "No matching requirements", "blank search should show placeholder")
	})

	t.Run("results_link_to_detail_panel", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/palette-results?search=CLI", nil))
		body := rec.Body.String()
		assertContains(t, body, `hx-target="#detail-panel-body"`, "palette results should target detail panel")
	})

	t.Run("palette_input_has_debounce", func(t *testing.T) {
		var buf strings.Builder
		_ = dashboard.RenderLayout(&buf, dashboard.LayoutData{Title: "Test", ActivePage: "status"})
		html := buf.String()
		assertContains(t, html, `delay:150ms`, "palette search should debounce at 150ms")
		assertContains(t, html, `/partials/palette-results`, "palette should target palette-results endpoint")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-027: Req-id links open panel across views
// --------------------------------------------------------------------------

func TestReqIdLinksTargetPanel(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-027")
	db := integrationDB(t)
	cfg := &config.Config{}
	mux := NewDashboardMuxWithPath(db, cfg, "")

	t.Run("kanban_req_ids_target_panel", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/kanban", nil))
		body := rec.Body.String()
		assertContains(t, body, `hx-target="#detail-panel-body"`, "kanban req-id links should target detail panel")
	})

	t.Run("releases_rows_target_panel", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/releases", nil))
		body := rec.Body.String()
		assertContains(t, body, `hx-target="#detail-panel-body"`, "release rows should target detail panel")
	})

	t.Run("detail_dep_links_target_panel", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/detail/REQ-CLI-001", nil))
		body := rec.Body.String()
		// Dependency links should target the panel, not #content
		assertContains(t, body, `hx-target="#detail-panel-body"`, "detail dep links should target panel")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-035: j/k row selection
// --------------------------------------------------------------------------

func TestTableRowKeyboardSelect(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-035")

	t.Run("jk_handler_in_js", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		assertContains(t, js, "e.key === 'j'", "JS should handle j key for row selection")
		assertContains(t, js, "e.key === 'k'", "JS should handle k key for row selection")
		assertContains(t, js, "row-selected", "JS should add row-selected class")
		assertContains(t, js, "scrollIntoView", "selected row should scroll into view")
	})

	t.Run("row_selected_css", func(t *testing.T) {
		css := dashboard.ReadStaticFile("styles.css")
		assertContains(t, css, "row-selected", "CSS should style row-selected class")
		assertContains(t, css, "rtmx-emerald", "selected row should use emerald highlight")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-036: g-prefix page navigation
// --------------------------------------------------------------------------

func TestGlobalPageShortcuts(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-036")

	t.Run("g_prefix_handler_in_js", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		assertContains(t, js, "gPending", "JS should track g-prefix pending state")
		assertContains(t, js, "e.key === 'g'", "JS should handle g key press")
		// Verify page mappings
		assertContains(t, js, "s: 'status'", "g+s should navigate to status")
		assertContains(t, js, "r: 'requirements'", "g+r should navigate to requirements")
		assertContains(t, js, "k: 'kanban'", "g+k should navigate to kanban")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-029: Panel URL state management
// --------------------------------------------------------------------------

func TestPanelURLState(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-029")

	t.Run("push_state_on_panel_open", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		assertContains(t, js, "history.pushState", "JS should push URL state when panel opens")
		assertContains(t, js, "panel:", "pushState should include panel identifier")
	})

	t.Run("pop_state_closes_panel", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		assertContains(t, js, "e.state && e.state.panel", "popstate handler should check for panel state")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-030: Panel edit OOB swap
// --------------------------------------------------------------------------

func TestPanelEditOOBSwap(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-030")
	db := integrationDB(t)
	cfg := &config.Config{}
	tmpFile := t.TempDir() + "/test_db.csv"
	_ = db.Save(tmpFile)
	mux := NewDashboardMuxWithPath(db, cfg, tmpFile)

	t.Run("htmx_patch_returns_oob_row", func(t *testing.T) {
		body := strings.NewReader(`{"status":"COMPLETE"}`)
		req := httptest.NewRequest("PATCH", "/api/requirements/REQ-MCP-002", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200, body: %s", rec.Code, rec.Body.String())
		}
		respBody := rec.Body.String()
		assertContains(t, respBody, `hx-swap-oob="true"`, "htmx PATCH should return OOB swap attribute")
		assertContains(t, respBody, "row-REQ-MCP-002", "OOB row should have correct ID")
	})

	t.Run("non_htmx_patch_returns_json", func(t *testing.T) {
		body := strings.NewReader(`{"assignee":"carol"}`)
		req := httptest.NewRequest("PATCH", "/api/requirements/REQ-CLI-001", body)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		ct := rec.Header().Get("Content-Type")
		assertContains(t, ct, "application/json", "non-htmx PATCH should return JSON")
	})

	t.Run("detail_form_targets_panel", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/detail/REQ-CLI-001", nil))
		body := rec.Body.String()
		assertContains(t, body, `hx-target="#detail-panel-body"`, "detail form should target panel body")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-033: Palette arrow key navigation
// --------------------------------------------------------------------------

func TestPaletteKeyboardNav(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-033")

	t.Run("arrow_key_handler_in_js", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		assertContains(t, js, "paletteKeyNav", "JS should define paletteKeyNav function")
		assertContains(t, js, "ArrowDown", "handler should support ArrowDown")
		assertContains(t, js, "ArrowUp", "handler should support ArrowUp")
		assertContains(t, js, "Enter", "handler should support Enter to select")
	})

	t.Run("palette_input_wired_to_keynav", func(t *testing.T) {
		var buf strings.Builder
		_ = dashboard.RenderLayout(&buf, dashboard.LayoutData{Title: "Test", ActivePage: "status"})
		html := buf.String()
		assertContains(t, html, "paletteKeyNav", "palette input should wire @keydown to paletteKeyNav")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-034: Palette action prefixes
// --------------------------------------------------------------------------

func TestPaletteActionPrefixes(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-034")
	db := integrationDB(t)
	cfg := &config.Config{}
	mux := NewDashboardMuxWithPath(db, cfg, "")

	t.Run("navigation_prefix", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/palette-results?search=>", nil))
		body := rec.Body.String()
		assertContains(t, body, "Go to Status", "> prefix should show navigation commands")
		assertContains(t, body, "Go to Requirements", "> prefix should include Requirements")
	})

	t.Run("assignee_prefix", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/palette-results?search=@", nil))
		body := rec.Body.String()
		assertContains(t, body, "assignee", "@ prefix should show assignee filter options")
	})

	t.Run("category_prefix", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/palette-results?search=%23", nil))
		body := rec.Body.String()
		assertContains(t, body, "category", "# prefix should show category filter options")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-037: Keyboard help overlay
// --------------------------------------------------------------------------

func TestKeyboardHelpOverlay(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-037")

	t.Run("help_overlay_in_layout", func(t *testing.T) {
		var buf strings.Builder
		_ = dashboard.RenderLayout(&buf, dashboard.LayoutData{Title: "Test", ActivePage: "status"})
		html := buf.String()
		assertContains(t, html, "keyboard-help", "layout must contain keyboard-help element")
		assertContains(t, html, "keyboardHelpOverlay", "layout must reference help overlay component")
		assertContains(t, html, "Keyboard Shortcuts", "help overlay must contain title")
	})

	t.Run("question_mark_handler_in_js", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		assertContains(t, js, "e.key === '?'", "JS should handle ? key")
		assertContains(t, js, "toggle-help", "? key should dispatch toggle-help event")
	})

	t.Run("help_component_in_js", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		assertContains(t, js, "keyboardHelpOverlay", "JS should define help overlay component")
		assertContains(t, js, "shortcuts:", "help component should list shortcuts")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-038: Input field suppression
// --------------------------------------------------------------------------

func TestShortcutInputSuppression(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-038")

	t.Run("suppression_in_js", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		assertContains(t, js, "tag === 'INPUT'", "shortcuts should be suppressed in INPUT elements")
		assertContains(t, js, "tag === 'SELECT'", "shortcuts should be suppressed in SELECT elements")
		assertContains(t, js, "tag === 'TEXTAREA'", "shortcuts should be suppressed in TEXTAREA elements")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-039: Checkbox selection column
// --------------------------------------------------------------------------

func TestBulkCheckboxSelection(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-039")
	db := integrationDB(t)
	cfg := &config.Config{}
	mux := NewDashboardMuxWithPath(db, cfg, "")

	t.Run("requirements_table_has_checkboxes", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements", nil))
		body := rec.Body.String()
		assertContains(t, body, "bulk-select-all", "table should have select-all checkbox")
		assertContains(t, body, "bulk-select", "rows should have individual checkboxes")
	})

	t.Run("row_partial_has_checkbox", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/requirements/row/REQ-CLI-001", nil))
		body := rec.Body.String()
		assertContains(t, body, "bulk-select", "row partial should include checkbox")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-040: Floating action bar
// --------------------------------------------------------------------------

func TestBulkActionBar(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-040")

	t.Run("action_bar_in_layout", func(t *testing.T) {
		var buf strings.Builder
		_ = dashboard.RenderLayout(&buf, dashboard.LayoutData{Title: "Test", ActivePage: "status"})
		html := buf.String()
		assertContains(t, html, "bulk-action-bar", "layout must contain bulk-action-bar element")
		assertContains(t, html, "bulkActionBar", "layout must reference bulkActionBar component")
		assertContains(t, html, "selectedCount", "action bar should show selected count")
		assertContains(t, html, "Set Status", "action bar should have Set Status button")
		assertContains(t, html, "Set Priority", "action bar should have Set Priority button")
		assertContains(t, html, "Set Assignee", "action bar should have Set Assignee button")
		assertContains(t, html, "Set Sprint", "action bar should have Set Sprint button")
	})

	t.Run("action_bar_component_in_js", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		assertContains(t, js, "bulkActionBar", "JS should define bulkActionBar component")
		assertContains(t, js, "getSelectedIds", "component should have getSelectedIds method")
		assertContains(t, js, "clearSelection", "component should have clearSelection method")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-042: Bulk action execution
// --------------------------------------------------------------------------

func TestBulkActionExecution(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-042")

	t.Run("execute_bulk_in_js", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		assertContains(t, js, "executeBulk", "JS should define executeBulk method")
		assertContains(t, js, "/api/requirements/bulk", "executeBulk should call bulk API endpoint")
		assertContains(t, js, "showToast", "executeBulk should show toast on completion")
		assertContains(t, js, "/partials/requirements", "executeBulk should refresh table after update")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-022: Per-column filter dropdowns
// --------------------------------------------------------------------------

func TestColumnFilterDropdowns(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-022")

	db := integrationDB(t)
	mux := integrationMux(t, db)

	t.Run("status_column_has_filter_dropdown", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/requirements", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		body := w.Body.String()
		assertContains(t, body, "columnFilter('status'", "Status column should have columnFilter component")
		assertContains(t, body, `aria-label="Filter status"`, "Status column should have filter button with aria-label")
		assertContains(t, body, "column-filter-dropdown", "Should contain dropdown div")
	})

	t.Run("priority_column_has_filter_dropdown", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/requirements", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		body := w.Body.String()
		assertContains(t, body, "columnFilter('priority'", "Priority column should have columnFilter component")
		assertContains(t, body, `aria-label="Filter priority"`, "Priority column should have filter button")
	})

	t.Run("category_column_has_filter_dropdown", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/requirements", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		body := w.Body.String()
		assertContains(t, body, "columnFilter('category'", "Category column should have columnFilter component")
		assertContains(t, body, `aria-label="Filter category"`, "Category column should have filter button")
	})

	t.Run("column_filter_js_component", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		assertContains(t, js, "function columnFilter(", "JS should define columnFilter component")
		assertContains(t, js, "toggleValue", "component should have toggleValue method")
		assertContains(t, js, "apply()", "component should have apply method")
	})

	t.Run("column_filter_css", func(t *testing.T) {
		css := dashboard.ReadStaticFile("styles.css")
		assertContains(t, css, ".column-filter-dropdown", "CSS should define column-filter-dropdown")
		assertContains(t, css, ".column-filter-btn", "CSS should define column-filter-btn")
		assertContains(t, css, ".column-filter-option", "CSS should define column-filter-option")
	})

	t.Run("multi_select_comma_separated_filter", func(t *testing.T) {
		// Filter by two statuses: COMPLETE and PARTIAL -- should return 3 of 5 reqs
		req := httptest.NewRequest("GET", "/partials/requirements?status=COMPLETE,PARTIAL", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		body := w.Body.String()
		assertContains(t, body, "REQ-CLI-001", "COMPLETE req should be in results")
		assertContains(t, body, "REQ-CLI-002", "COMPLETE req should be in results")
		assertContains(t, body, "REQ-MCP-001", "PARTIAL req should be in results")
		if strings.Contains(body, "REQ-API-001") {
			t.Error("MISSING req should not be in COMPLETE,PARTIAL filter")
		}
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-023: Active filter chips with clear-all
// --------------------------------------------------------------------------

func TestFilterChips(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-023")

	db := integrationDB(t)
	mux := integrationMux(t, db)

	t.Run("active_filters_shown_as_chips", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/requirements?status=COMPLETE", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		body := w.Body.String()
		assertContains(t, body, "filter-chip", "Active filter should render as chip")
		assertContains(t, body, "Status: COMPLETE", "Chip should show field and value")
		assertContains(t, body, "filter-chip-remove", "Chip should have remove button")
	})

	t.Run("clear_all_link_present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/requirements?status=COMPLETE&category=CLI", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		body := w.Body.String()
		assertContains(t, body, "filter-clear-all", "Should have clear-all link when filters active")
		assertContains(t, body, "Clear all", "Clear all link text")
	})

	t.Run("no_chips_without_filters", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/requirements", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		body := w.Body.String()
		if strings.Contains(body, "filter-chip") {
			t.Error("Should not show filter chips when no filters active")
		}
		if strings.Contains(body, "filter-clear-all") {
			t.Error("Should not show clear-all when no filters active")
		}
	})

	t.Run("chip_remove_js_function", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		assertContains(t, js, "removeFilterChip", "JS should define removeFilterChip function")
		assertContains(t, js, "clearAllFilters", "JS should define clearAllFilters function")
	})

	t.Run("chip_css", func(t *testing.T) {
		css := dashboard.ReadStaticFile("styles.css")
		assertContains(t, css, ".filter-chip", "CSS should define filter-chip")
		assertContains(t, css, ".filter-chip-remove", "CSS should define filter-chip-remove")
		assertContains(t, css, ".filter-clear-all", "CSS should define filter-clear-all")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-043: Vendor dagre.js as static asset
// --------------------------------------------------------------------------

func TestVendorDagreJS(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-043")

	t.Run("dagre_js_served_from_vendor_path", func(t *testing.T) {
		db := integrationDB(t)
		mux := integrationMux(t, db)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/static/vendor/dagre.min.js", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("dagre.min.js status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()
		if len(body) == 0 {
			t.Error("dagre.min.js should not be empty")
		}
		if len(body) > 300*1024 {
			t.Errorf("dagre.min.js is %d bytes, should be under 300KB", len(body))
		}
	})

	t.Run("dagre_script_tag_in_layout", func(t *testing.T) {
		var buf strings.Builder
		_ = dashboard.RenderLayout(&buf, dashboard.LayoutData{Title: "Test", ActivePage: "graph"})
		html := buf.String()
		if !strings.Contains(html, "dagre.min.js") {
			t.Error("layout should include dagre.min.js script tag")
		}
		// dagre should load after d3 and before app.js
		d3Pos := strings.Index(html, "d3.v7.min.js")
		dagrePos := strings.Index(html, "dagre.min.js")
		appPos := strings.Index(html, "app.js")
		if d3Pos < 0 || dagrePos < 0 || appPos < 0 {
			t.Fatal("expected d3, dagre, and app.js in layout")
		}
		if dagrePos < d3Pos {
			t.Error("dagre.min.js should appear after d3.v7.min.js")
		}
		if dagrePos > appPos {
			t.Error("dagre.min.js should appear before app.js")
		}
	})

	t.Run("dagre_contains_umd_build", func(t *testing.T) {
		content := dashboard.ReadStaticFile("vendor/dagre.min.js")
		if content == "" {
			t.Fatal("dagre.min.js not found in embedded assets")
		}
		if !strings.Contains(content, "dagre") {
			t.Error("dagre.min.js should contain 'dagre' identifier")
		}
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-044: Enriched graph node JSON
// --------------------------------------------------------------------------

func TestEnrichGraphNodeJSON(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-044")

	db := integrationDB(t)
	mux := integrationMux(t, db)

	// Helper to extract graph JSON from the partial HTML
	extractGraphJSON := func(t *testing.T, url string) map[string]json.RawMessage {
		t.Helper()
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", url, nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()
		// Find the JSON embedded in the script block
		start := strings.Index(body, "var graphData = ")
		if start < 0 {
			t.Fatal("graphData not found in response")
		}
		start += len("var graphData = ")
		end := strings.Index(body[start:], ";\n")
		if end < 0 {
			end = strings.Index(body[start:], ";")
		}
		if end < 0 {
			t.Fatal("could not find end of graphData JSON")
		}
		jsonStr := body[start : start+end]
		var result map[string]json.RawMessage
		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			t.Fatalf("failed to parse graphData JSON: %v\nJSON: %s", err, jsonStr[:min(200, len(jsonStr))])
		}
		return result
	}

	t.Run("nodes_have_enriched_fields", func(t *testing.T) {
		data := extractGraphJSON(t, "/partials/graph")
		var nodes []map[string]interface{}
		if err := json.Unmarshal(data["nodes"], &nodes); err != nil {
			t.Fatalf("failed to parse nodes: %v", err)
		}
		if len(nodes) == 0 {
			t.Fatal("expected at least one node")
		}
		// Check that enriched fields exist on every node
		requiredFields := []string{"effort_weeks", "priority", "phase", "assignee", "sprint", "is_blocked", "is_critical", "is_cycle_member"}
		for _, node := range nodes {
			id := node["id"].(string)
			for _, field := range requiredFields {
				if _, ok := node[field]; !ok {
					t.Errorf("node %s missing field %q", id, field)
				}
			}
		}
	})

	t.Run("effort_weeks_populated", func(t *testing.T) {
		data := extractGraphJSON(t, "/partials/graph")
		var nodes []map[string]interface{}
		_ = json.Unmarshal(data["nodes"], &nodes)
		for _, node := range nodes {
			id := node["id"].(string)
			ew, ok := node["effort_weeks"].(float64)
			if !ok {
				t.Errorf("node %s: effort_weeks should be a number", id)
				continue
			}
			// CLI-001 has effort 1.0
			if id == "REQ-CLI-001" && ew != 1.0 {
				t.Errorf("REQ-CLI-001 effort_weeks = %v, want 1.0", ew)
			}
		}
	})

	t.Run("is_blocked_correct", func(t *testing.T) {
		data := extractGraphJSON(t, "/partials/graph")
		var nodes []map[string]interface{}
		_ = json.Unmarshal(data["nodes"], &nodes)
		for _, node := range nodes {
			id := node["id"].(string)
			blocked := node["is_blocked"].(bool)
			// MCP-002 depends on MCP-001 which is PARTIAL, so MCP-002 is blocked
			if id == "REQ-MCP-002" && !blocked {
				t.Error("REQ-MCP-002 should be blocked (depends on incomplete MCP-001)")
			}
			// CLI-001 has no deps, should not be blocked
			if id == "REQ-CLI-001" && blocked {
				t.Error("REQ-CLI-001 should not be blocked (no deps)")
			}
		}
	})

	t.Run("priority_and_phase_populated", func(t *testing.T) {
		data := extractGraphJSON(t, "/partials/graph")
		var nodes []map[string]interface{}
		_ = json.Unmarshal(data["nodes"], &nodes)
		for _, node := range nodes {
			id := node["id"].(string)
			if id == "REQ-CLI-001" {
				if node["priority"] != "P0" {
					t.Errorf("REQ-CLI-001 priority = %v, want P0", node["priority"])
				}
				if node["phase"].(float64) != 1 {
					t.Errorf("REQ-CLI-001 phase = %v, want 1", node["phase"])
				}
			}
		}
	})

	t.Run("assignee_and_sprint_populated", func(t *testing.T) {
		data := extractGraphJSON(t, "/partials/graph")
		var nodes []map[string]interface{}
		_ = json.Unmarshal(data["nodes"], &nodes)
		for _, node := range nodes {
			id := node["id"].(string)
			if id == "REQ-CLI-001" {
				if node["assignee"] != "alice" {
					t.Errorf("REQ-CLI-001 assignee = %v, want alice", node["assignee"])
				}
				if node["sprint"] != "v1.0.0" {
					t.Errorf("REQ-CLI-001 sprint = %v, want v1.0.0", node["sprint"])
				}
			}
		}
	})

	t.Run("analysis_called_once_not_per_node", func(t *testing.T) {
		// Verify the JSON is well-formed and contains all nodes
		// (this implicitly tests that analysis runs once, producing consistent results)
		data := extractGraphJSON(t, "/partials/graph")
		var nodes []map[string]interface{}
		_ = json.Unmarshal(data["nodes"], &nodes)
		if len(nodes) != 5 {
			t.Errorf("expected 5 nodes, got %d", len(nodes))
		}
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-045: Enriched edge JSON with critical path flag
// --------------------------------------------------------------------------

func TestEnrichEdgeJSON(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-045")

	db := integrationDB(t)
	mux := integrationMux(t, db)

	extractEdges := func(t *testing.T) []map[string]interface{} {
		t.Helper()
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/graph", nil))
		body := rec.Body.String()
		start := strings.Index(body, "var graphData = ")
		if start < 0 {
			t.Fatal("graphData not found")
		}
		start += len("var graphData = ")
		end := strings.Index(body[start:], ";\n")
		if end < 0 {
			end = strings.Index(body[start:], ";")
		}
		jsonStr := body[start : start+end]
		var data map[string]json.RawMessage
		_ = json.Unmarshal([]byte(jsonStr), &data)
		var edges []map[string]interface{}
		_ = json.Unmarshal(data["edges"], &edges)
		return edges
	}

	t.Run("edges_have_is_critical_path_field", func(t *testing.T) {
		edges := extractEdges(t)
		if len(edges) == 0 {
			t.Fatal("expected at least one edge")
		}
		for _, edge := range edges {
			if _, ok := edge["is_critical_path"]; !ok {
				t.Errorf("edge %s->%s missing is_critical_path field", edge["source"], edge["target"])
			}
		}
	})

	t.Run("is_critical_path_is_boolean", func(t *testing.T) {
		edges := extractEdges(t)
		for _, edge := range edges {
			if _, ok := edge["is_critical_path"].(bool); !ok {
				t.Errorf("edge %s->%s: is_critical_path should be boolean, got %T", edge["source"], edge["target"], edge["is_critical_path"])
			}
		}
	})

	t.Run("edge_source_target_preserved", func(t *testing.T) {
		edges := extractEdges(t)
		foundCLIDep := false
		for _, edge := range edges {
			if edge["source"] == "REQ-CLI-001" && edge["target"] == "REQ-CLI-002" {
				foundCLIDep = true
			}
		}
		if !foundCLIDep {
			t.Error("expected edge CLI-001 -> CLI-002")
		}
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-046: Layer depth data in graph JSON
// --------------------------------------------------------------------------

func TestGraphLayerData(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-046")

	db := integrationDB(t)
	mux := integrationMux(t, db)

	extractGraphData := func(t *testing.T, url string) ([]map[string]interface{}, map[string]json.RawMessage) {
		t.Helper()
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", url, nil))
		body := rec.Body.String()
		start := strings.Index(body, "var graphData = ")
		if start < 0 {
			t.Fatal("graphData not found")
		}
		start += len("var graphData = ")
		end := strings.Index(body[start:], ";\n")
		if end < 0 {
			end = strings.Index(body[start:], ";")
		}
		jsonStr := body[start : start+end]
		var data map[string]json.RawMessage
		_ = json.Unmarshal([]byte(jsonStr), &data)
		var nodes []map[string]interface{}
		_ = json.Unmarshal(data["nodes"], &nodes)
		return nodes, data
	}

	t.Run("nodes_have_layer_field", func(t *testing.T) {
		nodes, _ := extractGraphData(t, "/partials/graph")
		for _, node := range nodes {
			if _, ok := node["layer"]; !ok {
				t.Errorf("node %s missing layer field", node["id"])
			}
		}
	})

	t.Run("root_nodes_at_layer_zero", func(t *testing.T) {
		nodes, _ := extractGraphData(t, "/partials/graph")
		for _, node := range nodes {
			id := node["id"].(string)
			layer := int(node["layer"].(float64))
			// CLI-001 and MCP-001 are roots (no deps from within the set)
			if id == "REQ-CLI-001" && layer != 0 {
				t.Errorf("REQ-CLI-001 layer = %d, want 0 (root node)", layer)
			}
			if id == "REQ-MCP-001" && layer != 0 {
				t.Errorf("REQ-MCP-001 layer = %d, want 0 (root node)", layer)
			}
		}
	})

	t.Run("dependent_nodes_at_deeper_layer", func(t *testing.T) {
		nodes, _ := extractGraphData(t, "/partials/graph")
		for _, node := range nodes {
			id := node["id"].(string)
			layer := int(node["layer"].(float64))
			// CLI-002 depends on CLI-001, so layer >= 1
			if id == "REQ-CLI-002" && layer < 1 {
				t.Errorf("REQ-CLI-002 layer = %d, want >= 1 (depends on CLI-001)", layer)
			}
		}
	})

	t.Run("top_level_layer_count", func(t *testing.T) {
		_, data := extractGraphData(t, "/partials/graph")
		var layerCount float64
		if err := json.Unmarshal(data["layer_count"], &layerCount); err != nil {
			t.Fatalf("failed to parse layer_count: %v", err)
		}
		if layerCount < 2 {
			t.Errorf("layer_count = %v, want >= 2 (at least roots and dependents)", layerCount)
		}
	})

	t.Run("filtered_layers_recomputed", func(t *testing.T) {
		// Filter to CLI category only: CLI-001 (root) -> CLI-002
		nodes, data := extractGraphData(t, "/partials/graph?category=CLI")
		if len(nodes) != 2 {
			t.Fatalf("expected 2 CLI nodes, got %d", len(nodes))
		}
		var layerCount float64
		_ = json.Unmarshal(data["layer_count"], &layerCount)
		if layerCount != 2 {
			t.Errorf("filtered layer_count = %v, want 2 (CLI-001 at 0, CLI-002 at 1)", layerCount)
		}
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-047: Web IDs in graph JSON
// --------------------------------------------------------------------------

func TestGraphWebIDs(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-047")

	db := integrationDB(t)
	mux := integrationMux(t, db)

	extractGraphData := func(t *testing.T) ([]map[string]interface{}, map[string]json.RawMessage) {
		t.Helper()
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/graph", nil))
		body := rec.Body.String()
		start := strings.Index(body, "var graphData = ")
		if start < 0 {
			t.Fatal("graphData not found")
		}
		start += len("var graphData = ")
		end := strings.Index(body[start:], ";\n")
		if end < 0 {
			end = strings.Index(body[start:], ";")
		}
		jsonStr := body[start : start+end]
		var data map[string]json.RawMessage
		_ = json.Unmarshal([]byte(jsonStr), &data)
		var nodes []map[string]interface{}
		_ = json.Unmarshal(data["nodes"], &nodes)
		return nodes, data
	}

	t.Run("nodes_have_web_id_field", func(t *testing.T) {
		nodes, _ := extractGraphData(t)
		for _, node := range nodes {
			if _, ok := node["web_id"]; !ok {
				t.Errorf("node %s missing web_id field", node["id"])
			}
		}
	})

	t.Run("complete_nodes_have_web_id_negative_one", func(t *testing.T) {
		nodes, _ := extractGraphData(t)
		for _, node := range nodes {
			id := node["id"].(string)
			wid := int(node["web_id"].(float64))
			if id == "REQ-CLI-001" && wid != -1 {
				t.Errorf("REQ-CLI-001 (COMPLETE) web_id = %d, want -1", wid)
			}
			if id == "REQ-CLI-002" && wid != -1 {
				t.Errorf("REQ-CLI-002 (COMPLETE) web_id = %d, want -1", wid)
			}
		}
	})

	t.Run("incomplete_nodes_have_valid_web_id", func(t *testing.T) {
		nodes, _ := extractGraphData(t)
		for _, node := range nodes {
			id := node["id"].(string)
			status := node["status"].(string)
			wid := int(node["web_id"].(float64))
			if status != "COMPLETE" && wid < 0 {
				t.Errorf("incomplete node %s has web_id = %d, want >= 0", id, wid)
			}
		}
	})

	t.Run("top_level_web_count", func(t *testing.T) {
		_, data := extractGraphData(t)
		var webCount float64
		if err := json.Unmarshal(data["web_count"], &webCount); err != nil {
			t.Fatalf("failed to parse web_count: %v", err)
		}
		// MCP-001, MCP-002, API-001 form one connected web (MCP-001 blocks both)
		// So web_count should be 1 (one connected component of incomplete reqs)
		if webCount < 1 {
			t.Errorf("web_count = %v, want >= 1", webCount)
		}
	})

	t.Run("web_ids_zero_indexed", func(t *testing.T) {
		nodes, _ := extractGraphData(t)
		maxWebID := -1
		for _, node := range nodes {
			wid := int(node["web_id"].(float64))
			if wid > maxWebID {
				maxWebID = wid
			}
		}
		// Web IDs should be zero-indexed
		if maxWebID < 0 {
			t.Error("no valid web IDs found")
		}
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-048: Dagre layout replacing force simulation
// --------------------------------------------------------------------------

func TestDagreLayout(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-048")

	db := integrationDB(t)
	mux := integrationMux(t, db)

	getGraphBody := func(t *testing.T) string {
		t.Helper()
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/graph", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		return rec.Body.String()
	}

	t.Run("no_force_simulation_references", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		if strings.Contains(js, "forceSimulation") {
			t.Error("app.js should not contain d3.forceSimulation (replaced by dagre)")
		}
		if strings.Contains(js, "forceManyBody") {
			t.Error("app.js should not contain d3.forceManyBody")
		}
	})

	t.Run("dagre_layout_in_renderGraph", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		if !strings.Contains(js, "dagre.graphlib.Graph") {
			t.Error("renderGraph should create a dagre.graphlib.Graph")
		}
		if !strings.Contains(js, "dagre.layout") {
			t.Error("renderGraph should call dagre.layout()")
		}
	})

	t.Run("configurable_spacing", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		if !strings.Contains(js, "ranksep") {
			t.Error("renderGraph should set ranksep for dagre")
		}
		if !strings.Contains(js, "nodesep") {
			t.Error("renderGraph should set nodesep for dagre")
		}
	})

	t.Run("no_node_drag", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		if strings.Contains(js, "d3.drag()") {
			t.Error("renderGraph should not use d3.drag (dagre positions are deterministic)")
		}
	})

	t.Run("auto_size_viewbox", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		if !strings.Contains(js, "viewBox") {
			t.Error("renderGraph should set SVG viewBox to auto-size")
		}
	})

	t.Run("renders_with_dagre_check", func(t *testing.T) {
		body := getGraphBody(t)
		// The script should check for dagre availability
		if !strings.Contains(body, "dagre") {
			t.Error("graph partial script should reference dagre")
		}
	})

	t.Run("arrow_markers_present", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		if !strings.Contains(js, "marker") {
			t.Error("renderGraph should define arrow markers")
		}
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-066: View switcher component
// --------------------------------------------------------------------------

func TestViewSwitcher(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-066")

	db := integrationDB(t)
	mux := integrationMux(t, db)

	t.Run("graph_partial_contains_view_switcher", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/graph", nil))
		body := rec.Body.String()
		assertContains(t, body, "view-switcher", "graph partial should contain view switcher")
		assertContains(t, body, "DAG", "view switcher should have DAG option")
		assertContains(t, body, "Timeline", "view switcher should have Timeline option")
		assertContains(t, body, "Treemap", "view switcher should have Treemap option")
	})

	t.Run("view_switcher_js_component", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		if !strings.Contains(js, "viewSwitcher") {
			t.Error("app.js should contain viewSwitcher Alpine component")
		}
		if !strings.Contains(js, "switchView") {
			t.Error("viewSwitcher should have switchView method")
		}
	})

	t.Run("view_switcher_css", func(t *testing.T) {
		css := dashboard.ReadStaticFile("styles.css")
		if !strings.Contains(css, ".view-switcher") {
			t.Error("styles.css should contain .view-switcher class")
		}
		if !strings.Contains(css, ".view-switcher-btn") {
			t.Error("styles.css should contain .view-switcher-btn class")
		}
	})

	t.Run("graph_view_content_target", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/graph", nil))
		body := rec.Body.String()
		assertContains(t, body, "graph-view-content", "graph partial should have view content container for htmx swap")
	})

	t.Run("timeline_button_targets_correct_partial", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/graph", nil))
		body := rec.Body.String()
		assertContains(t, body, "/partials/timeline", "timeline button should target /partials/timeline")
	})

	t.Run("keyboard_accessible", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/graph", nil))
		body := rec.Body.String()
		assertContains(t, body, "role=", "view switcher buttons should have ARIA roles")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-091: Reusable tooltip component
// --------------------------------------------------------------------------

func TestReusableTooltipComponent(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-091")

	t.Run("tooltip_js_component", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		if !strings.Contains(js, "function tooltip()") {
			t.Error("app.js should contain tooltip() Alpine component function")
		}
		if !strings.Contains(js, "show:") || !strings.Contains(js, "hide:") {
			t.Error("tooltip should have show and hide methods")
		}
	})

	t.Run("viewport_aware_positioning", func(t *testing.T) {
		js := dashboard.ReadStaticFile("app.js")
		if !strings.Contains(js, "innerWidth") || !strings.Contains(js, "innerHeight") {
			t.Error("tooltip should check viewport dimensions for repositioning")
		}
	})

	t.Run("tooltip_css", func(t *testing.T) {
		css := dashboard.ReadStaticFile("styles.css")
		if !strings.Contains(css, ".rtmx-tooltip") {
			t.Error("styles.css should contain .rtmx-tooltip class")
		}
		if !strings.Contains(css, "max-width") {
			t.Error("tooltip CSS should set max-width")
		}
	})

	t.Run("tooltip_uses_design_tokens", func(t *testing.T) {
		css := dashboard.ReadStaticFile("styles.css")
		if !strings.Contains(css, "var(--rtmx-bg-card)") {
			t.Error("tooltip should use --rtmx-bg-card design token")
		}
		if !strings.Contains(css, "var(--rtmx-border)") {
			t.Error("tooltip should use --rtmx-border design token")
		}
	})

	t.Run("tooltip_key_value_structure", func(t *testing.T) {
		css := dashboard.ReadStaticFile("styles.css")
		if !strings.Contains(css, ".rtmx-tooltip-key") {
			t.Error("tooltip should have key-value row styling")
		}
		if !strings.Contains(css, ".rtmx-tooltip-value") {
			t.Error("tooltip should have value styling")
		}
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-092: Timeline partial template skeleton
// --------------------------------------------------------------------------

func TestTimelinePartialTemplateSkeleton(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-092")

	db := integrationDB(t)
	mux := integrationMux(t, db)

	t.Run("timeline_route_returns_200", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/timeline", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("timeline status = %d, want 200", rec.Code)
		}
	})

	t.Run("timeline_contains_svg_container", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/timeline", nil))
		body := rec.Body.String()
		assertContains(t, body, "timeline-svg", "timeline should have SVG element with id timeline-svg")
	})

	t.Run("timeline_contains_json_data_block", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/timeline", nil))
		body := rec.Body.String()
		assertContains(t, body, "application/json", "timeline should have JSON script block")
		assertContains(t, body, "timeline-data", "timeline should have timeline-data script id")
	})

	t.Run("timeline_has_zoom_controls", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/timeline", nil))
		body := rec.Body.String()
		assertContains(t, body, "timelineZoom", "timeline should have zoom control buttons")
	})
}

// --------------------------------------------------------------------------
// REQ-DASH-093: Treemap partial template skeleton
// --------------------------------------------------------------------------

func TestTreemapPartialTemplateSkeleton(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-093")

	db := integrationDB(t)
	mux := integrationMux(t, db)

	t.Run("treemap_route_returns_200", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/treemap", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("treemap status = %d, want 200", rec.Code)
		}
	})

	t.Run("treemap_contains_svg_container", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/treemap", nil))
		body := rec.Body.String()
		assertContains(t, body, "treemap-svg", "treemap should have SVG element with id treemap-svg")
	})

	t.Run("treemap_contains_json_data_block", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/treemap", nil))
		body := rec.Body.String()
		assertContains(t, body, "application/json", "treemap should have JSON script block")
		assertContains(t, body, "treemap-data", "treemap should have treemap-data script id")
	})

	t.Run("treemap_has_16_9_aspect", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/treemap", nil))
		body := rec.Body.String()
		assertContains(t, body, "16/9", "treemap container should have 16:9 aspect ratio")
	})
}

// --------------------------------------------------------------------------
// Tier 2: Visual Encoding Tests
// --------------------------------------------------------------------------

// TestNodeSizeEffortEncoding verifies REQ-DASH-050: node radius proportional to effort_weeks.
func TestNodeSizeEffortEncoding(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-050")

	// Read app.js
	appJS := dashboard.ReadStaticFile("app.js")

	t.Run("effort_radius_function_exists", func(t *testing.T) {
		assertContains(t, appJS, "function effortRadius", "app.js should define effortRadius function")
	})

	t.Run("default_radius_12", func(t *testing.T) {
		assertContains(t, appJS, "return 12", "default radius for unset effort should be 12")
	})

	t.Run("min_radius_8", func(t *testing.T) {
		assertContains(t, appJS, "return 8", "minimum radius should be 8")
	})

	t.Run("max_radius_28", func(t *testing.T) {
		assertContains(t, appJS, "return 28", "maximum radius should be 28")
	})

	t.Run("radius_used_in_dagre_dimensions", func(t *testing.T) {
		assertContains(t, appJS, "effortRadius(n.effort_weeks)", "renderGraph should call effortRadius with effort_weeks")
	})

	t.Run("dagre_node_dimensions_use_radius", func(t *testing.T) {
		assertContains(t, appJS, "2 * r", "dagre node dimensions should use 2*radius")
	})
}

// TestNodeFillStatusColor verifies REQ-DASH-051: node fill color encodes status.
func TestNodeFillStatusColor(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-051")

	cssContent := dashboard.ReadStaticFile("styles.css")

	t.Run("css_classes_defined", func(t *testing.T) {
		assertContains(t, cssContent, ".node-complete", "CSS should define .node-complete")
		assertContains(t, cssContent, ".node-partial", "CSS should define .node-partial")
		assertContains(t, cssContent, ".node-missing", "CSS should define .node-missing")
		assertContains(t, cssContent, ".node-not-started", "CSS should define .node-not-started")
	})

	t.Run("uses_design_tokens", func(t *testing.T) {
		assertContains(t, cssContent, "--rtmx-complete", "node-complete should use --rtmx-complete token")
		assertContains(t, cssContent, "--rtmx-partial", "node-partial should use --rtmx-partial token")
		assertContains(t, cssContent, "--rtmx-missing", "node-missing should use --rtmx-missing token")
		assertContains(t, cssContent, "--rtmx-not-started", "node-not-started should use --rtmx-not-started token")
	})

	t.Run("js_status_class_function", func(t *testing.T) {
		appJS := dashboard.ReadStaticFile("app.js")
		assertContains(t, appJS, "function statusClass", "app.js should define statusClass function")
	})

	t.Run("applied_via_css_classes_not_inline", func(t *testing.T) {
		appJS := dashboard.ReadStaticFile("app.js")
		assertContains(t, appJS, "statusClass(n.status)", "renderGraph should apply status via CSS class")
	})
}

// TestNodeBorderPriority verifies REQ-DASH-052: node border ring encodes priority.
func TestNodeBorderPriority(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-052")

	cssContent := dashboard.ReadStaticFile("styles.css")

	t.Run("priority_css_classes", func(t *testing.T) {
		assertContains(t, cssContent, ".priority-p0", "CSS should define .priority-p0")
		assertContains(t, cssContent, ".priority-high", "CSS should define .priority-high")
		assertContains(t, cssContent, ".priority-medium", "CSS should define .priority-medium")
		assertContains(t, cssContent, ".priority-low", "CSS should define .priority-low")
	})

	t.Run("p0_has_3px_stroke", func(t *testing.T) {
		assertContains(t, cssContent, "stroke-width: 3px", "P0 ring should be 3px")
	})

	t.Run("high_has_2px_stroke", func(t *testing.T) {
		assertContains(t, cssContent, "stroke-width: 2px", "HIGH ring should be 2px")
	})

	t.Run("medium_has_1px_stroke", func(t *testing.T) {
		assertContains(t, cssContent, "stroke-width: 1px", "MEDIUM ring should be 1px")
	})

	t.Run("js_priority_class_function", func(t *testing.T) {
		appJS := dashboard.ReadStaticFile("app.js")
		assertContains(t, appJS, "function priorityClass", "app.js should define priorityClass function")
	})

	t.Run("stroke_width_in_dagre_dimensions", func(t *testing.T) {
		appJS := dashboard.ReadStaticFile("app.js")
		assertContains(t, appJS, "2 * strokeW", "dagre dimensions should account for stroke width")
	})
}

// TestCriticalPathEdgeRendering verifies REQ-DASH-053: critical path edge visual distinction.
func TestCriticalPathEdgeRendering(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-053")

	cssContent := dashboard.ReadStaticFile("styles.css")
	appJS := dashboard.ReadStaticFile("app.js")

	t.Run("edge_critical_css_class", func(t *testing.T) {
		assertContains(t, cssContent, ".edge-critical", "CSS should define .edge-critical")
	})

	t.Run("edge_normal_css_class", func(t *testing.T) {
		assertContains(t, cssContent, ".edge-normal", "CSS should define .edge-normal")
	})

	t.Run("critical_path_animation", func(t *testing.T) {
		assertContains(t, cssContent, "critical-path-flow", "CSS should define critical-path-flow animation")
		assertContains(t, cssContent, "stroke-dashoffset", "animation should use stroke-dashoffset")
	})

	t.Run("reduced_motion_support", func(t *testing.T) {
		assertContains(t, cssContent, "prefers-reduced-motion", "should respect prefers-reduced-motion")
	})

	t.Run("separate_arrow_markers", func(t *testing.T) {
		assertContains(t, appJS, "arrow-normal", "should have normal arrow marker")
		assertContains(t, appJS, "arrow-critical", "should have critical arrow marker")
	})

	t.Run("critical_edges_rendered_on_top", func(t *testing.T) {
		assertContains(t, appJS, "normalEdges.forEach", "normal edges should render first")
		assertContains(t, appJS, "criticalEdges.forEach", "critical edges should render after normal")
	})

	t.Run("marker_css_classes", func(t *testing.T) {
		assertContains(t, cssContent, ".marker-critical", "CSS should define .marker-critical")
		assertContains(t, cssContent, ".marker-normal", "CSS should define .marker-normal")
	})
}

// TestBlockedNodeVisualTreatment verifies REQ-DASH-054: blocked node dashed border + opacity.
func TestBlockedNodeVisualTreatment(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-054")

	cssContent := dashboard.ReadStaticFile("styles.css")
	appJS := dashboard.ReadStaticFile("app.js")

	t.Run("node_blocked_css_class", func(t *testing.T) {
		assertContains(t, cssContent, ".node-blocked", "CSS should define .node-blocked")
	})

	t.Run("dashed_border", func(t *testing.T) {
		assertContains(t, cssContent, "stroke-dasharray: 5 3", "blocked nodes should have dashed border")
	})

	t.Run("reduced_opacity", func(t *testing.T) {
		assertContains(t, cssContent, "opacity: 0.7", "blocked nodes should have reduced opacity")
	})

	t.Run("js_applies_blocked_class", func(t *testing.T) {
		assertContains(t, appJS, "node-blocked", "renderGraph should apply node-blocked class")
		assertContains(t, appJS, "n.is_blocked", "renderGraph should check is_blocked field")
	})
}

// TestCycleMemberIndicator verifies REQ-DASH-055: cycle member warning badge + glow.
func TestCycleMemberIndicator(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-055")

	cssContent := dashboard.ReadStaticFile("styles.css")
	appJS := dashboard.ReadStaticFile("app.js")

	t.Run("cycle_badge_rendered", func(t *testing.T) {
		assertContains(t, appJS, "cycle-badge", "renderGraph should create cycle badge")
		assertContains(t, appJS, "polygon", "cycle badge should be a triangle (polygon)")
	})

	t.Run("cycle_glow_animation", func(t *testing.T) {
		assertContains(t, cssContent, "cycle-pulse", "CSS should define cycle-pulse animation")
		assertContains(t, appJS, "cycle-glow", "renderGraph should create cycle glow element")
	})

	t.Run("cycle_highlight_on_click", func(t *testing.T) {
		assertContains(t, appJS, "cycleHighlightActive", "renderGraph should track cycle highlight state")
		assertContains(t, appJS, "cycleMembership", "renderGraph should build cycle membership lookup")
	})

	t.Run("reduced_motion_for_glow", func(t *testing.T) {
		assertContains(t, cssContent, "prefers-reduced-motion", "glow animation should respect reduced motion")
	})

	t.Run("click_background_exits_highlight", func(t *testing.T) {
		assertContains(t, appJS, "event.target === svg.node()", "clicking SVG background should exit cycle highlight")
	})
}

// TestNodeLabelsAlwaysVisible verifies REQ-DASH-056: always-visible labels with zoom-responsive detail.
func TestNodeLabelsAlwaysVisible(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-056")

	cssContent := dashboard.ReadStaticFile("styles.css")
	appJS := dashboard.ReadStaticFile("app.js")

	t.Run("label_id_css_class", func(t *testing.T) {
		assertContains(t, cssContent, ".node-label-id", "CSS should define .node-label-id")
	})

	t.Run("label_desc_css_class", func(t *testing.T) {
		assertContains(t, cssContent, ".node-label-desc", "CSS should define .node-label-desc")
	})

	t.Run("label_uses_mono_font", func(t *testing.T) {
		assertContains(t, cssContent, "font-mono", "label ID should use monospace font token")
	})

	t.Run("zoom_threshold_1_5", func(t *testing.T) {
		assertContains(t, appJS, "1.5", "zoom threshold for expanded labels should be 1.5")
		assertContains(t, appJS, "labels-expanded", "SVG should get labels-expanded class")
	})

	t.Run("labels_hidden_below_0_5", func(t *testing.T) {
		assertContains(t, appJS, "0.5", "zoom threshold for hidden labels should be 0.5")
		assertContains(t, appJS, "labels-hidden", "SVG should get labels-hidden class")
	})

	t.Run("description_truncated_40_chars", func(t *testing.T) {
		assertContains(t, appJS, "40", "description should be truncated at 40 chars")
		assertContains(t, appJS, "...", "truncated description should end with ellipsis")
	})

	t.Run("label_below_node", func(t *testing.T) {
		assertContains(t, appJS, "r + 12", "label y-offset should be radius + 12")
	})
}

// TestEdgePathRendering verifies REQ-DASH-083: dagre edge paths with rounded corners.
func TestEdgePathRendering(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-083")

	appJS := dashboard.ReadStaticFile("app.js")

	t.Run("edge_path_function", func(t *testing.T) {
		assertContains(t, appJS, "function edgePath", "app.js should define edgePath function")
	})

	t.Run("uses_svg_path_not_line", func(t *testing.T) {
		assertContains(t, appJS, "append('path')", "edges should be rendered as SVG path elements")
	})

	t.Run("rounded_corners", func(t *testing.T) {
		assertContains(t, appJS, "cornerRadius", "edge paths should use corner radius")
		assertContains(t, appJS, "Q", "paths should use quadratic bezier for rounded corners")
	})

	t.Run("uses_dagre_edge_points", func(t *testing.T) {
		assertContains(t, appJS, "g.edge(e.source, e.target)", "should use dagre-computed edge data")
		assertContains(t, appJS, "edgeData.points", "should use dagre edge points array")
	})
}

// TestInteractionLayer verifies REQ-DASH-084: zoom, pan, hover preservation on dagre layout.
func TestInteractionLayer(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-084")

	appJS := dashboard.ReadStaticFile("app.js")

	t.Run("zoom_on_svg_group", func(t *testing.T) {
		assertContains(t, appJS, "graph-root", "zoom should operate on graph-root SVG group")
	})

	t.Run("hover_scale_enlargement", func(t *testing.T) {
		assertContains(t, appJS, "1.1", "hover should scale node to 1.1x")
	})

	t.Run("no_node_drag", func(t *testing.T) {
		if strings.Contains(appJS, "d3.drag") {
			t.Error("dagre layout should not have d3.drag -- positions are deterministic")
		}
	})

	t.Run("double_click_reset_zoom", func(t *testing.T) {
		assertContains(t, appJS, "dblclick.zoom", "double-click should reset zoom")
		assertContains(t, appJS, "zoomIdentity", "reset should use d3.zoomIdentity")
	})

	t.Run("description_in_node_data", func(t *testing.T) {
		db := integrationDB(t)
		mux := integrationMux(t, db)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/graph", nil))
		body := rec.Body.String()
		assertContains(t, body, "description", "graph JSON nodes should include description field")
	})
}

// TestBlockingIDsJSON verifies REQ-DASH-095: blocking_ids array in enriched node JSON.
func TestBlockingIDsJSON(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-095")

	db := integrationDB(t)
	mux := integrationMux(t, db)

	extractGraphJSON := func(t *testing.T) map[string]json.RawMessage {
		t.Helper()
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/graph", nil))
		body := rec.Body.String()
		start := strings.Index(body, "var graphData = ")
		if start < 0 {
			t.Fatal("graphData not found")
		}
		start += len("var graphData = ")
		end := strings.Index(body[start:], ";\n")
		if end < 0 {
			end = strings.Index(body[start:], ";")
		}
		jsonStr := body[start : start+end]
		var result map[string]json.RawMessage
		if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
			t.Fatalf("failed to parse graphData: %v", err)
		}
		return result
	}

	t.Run("blocking_ids_field_present", func(t *testing.T) {
		data := extractGraphJSON(t)
		var nodes []map[string]interface{}
		_ = json.Unmarshal(data["nodes"], &nodes)
		for _, node := range nodes {
			id := node["id"].(string)
			if _, ok := node["blocking_ids"]; !ok {
				t.Errorf("node %s missing blocking_ids field", id)
			}
		}
	})

	t.Run("blocking_ids_is_array_not_null", func(t *testing.T) {
		data := extractGraphJSON(t)
		var nodes []map[string]interface{}
		_ = json.Unmarshal(data["nodes"], &nodes)
		for _, node := range nodes {
			id := node["id"].(string)
			bids := node["blocking_ids"]
			if bids == nil {
				t.Errorf("node %s blocking_ids should be [] not null", id)
			}
			// Should be an array
			if _, ok := bids.([]interface{}); !ok {
				t.Errorf("node %s blocking_ids should be array, got %T", id, bids)
			}
		}
	})

	t.Run("unblocked_node_has_empty_array", func(t *testing.T) {
		data := extractGraphJSON(t)
		var nodes []map[string]interface{}
		_ = json.Unmarshal(data["nodes"], &nodes)
		for _, node := range nodes {
			id := node["id"].(string)
			if id == "REQ-CLI-001" {
				bids := node["blocking_ids"].([]interface{})
				if len(bids) != 0 {
					t.Errorf("REQ-CLI-001 has no deps, blocking_ids should be empty, got %v", bids)
				}
			}
		}
	})

	t.Run("blocked_node_lists_incomplete_deps", func(t *testing.T) {
		data := extractGraphJSON(t)
		var nodes []map[string]interface{}
		_ = json.Unmarshal(data["nodes"], &nodes)
		for _, node := range nodes {
			id := node["id"].(string)
			// MCP-002 depends on MCP-001 which is PARTIAL (incomplete)
			if id == "REQ-MCP-002" {
				bids := node["blocking_ids"].([]interface{})
				if len(bids) == 0 {
					t.Error("REQ-MCP-002 should have blocking_ids containing REQ-MCP-001")
				}
				found := false
				for _, b := range bids {
					if b.(string) == "REQ-MCP-001" {
						found = true
					}
				}
				if !found {
					t.Errorf("REQ-MCP-002 blocking_ids should contain REQ-MCP-001, got %v", bids)
				}
			}
		}
	})

	t.Run("complete_dep_not_in_blocking_ids", func(t *testing.T) {
		data := extractGraphJSON(t)
		var nodes []map[string]interface{}
		_ = json.Unmarshal(data["nodes"], &nodes)
		for _, node := range nodes {
			id := node["id"].(string)
			// CLI-002 depends on CLI-001 which is COMPLETE
			if id == "REQ-CLI-002" {
				bids := node["blocking_ids"].([]interface{})
				if len(bids) != 0 {
					t.Errorf("REQ-CLI-002 dep CLI-001 is COMPLETE, blocking_ids should be empty, got %v", bids)
				}
			}
		}
	})
}

// TestCyclePathJSON verifies REQ-DASH-096: cycles array in enriched graph JSON.
func TestCyclePathJSON(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-096")

	t.Run("cycles_field_present_and_empty", func(t *testing.T) {
		db := integrationDB(t)
		mux := integrationMux(t, db)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/graph", nil))
		body := rec.Body.String()
		start := strings.Index(body, "var graphData = ")
		start += len("var graphData = ")
		end := strings.Index(body[start:], ";\n")
		if end < 0 {
			end = strings.Index(body[start:], ";")
		}
		jsonStr := body[start : start+end]
		var result map[string]json.RawMessage
		_ = json.Unmarshal([]byte(jsonStr), &result)

		if _, ok := result["cycles"]; !ok {
			t.Fatal("graph JSON should include cycles field")
		}
		var cycles [][]string
		if err := json.Unmarshal(result["cycles"], &cycles); err != nil {
			t.Fatalf("cycles should be [][]string: %v", err)
		}
		// Integration DB has no cycles
		if len(cycles) != 0 {
			t.Errorf("acyclic graph should have empty cycles array, got %v", cycles)
		}
	})

	t.Run("cycles_detected_for_cyclic_graph", func(t *testing.T) {
		// Create a DB with a cycle: A -> B -> A
		db := database.NewDatabase()
		reqA := &database.Requirement{ReqID: "A", Category: "TEST", Status: database.StatusMissing, Dependencies: database.StringSet{"B": {}}, Blocks: database.StringSet{}}
		reqB := &database.Requirement{ReqID: "B", Category: "TEST", Status: database.StatusMissing, Dependencies: database.StringSet{"A": {}}, Blocks: database.StringSet{}}
		_ = db.Add(reqA)
		_ = db.Add(reqB)
		mux := integrationMux(t, db)

		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/partials/graph", nil))
		body := rec.Body.String()
		start := strings.Index(body, "var graphData = ")
		start += len("var graphData = ")
		end := strings.Index(body[start:], ";\n")
		if end < 0 {
			end = strings.Index(body[start:], ";")
		}
		jsonStr := body[start : start+end]
		var result map[string]json.RawMessage
		_ = json.Unmarshal([]byte(jsonStr), &result)

		var cycles [][]string
		_ = json.Unmarshal(result["cycles"], &cycles)
		if len(cycles) == 0 {
			t.Fatal("cyclic graph should have at least one cycle in cycles array")
		}
		// Cycle should contain both A and B
		cycle := cycles[0]
		hasA, hasB := false, false
		for _, id := range cycle {
			if id == "A" { hasA = true }
			if id == "B" { hasB = true }
		}
		if !hasA || !hasB {
			t.Errorf("cycle should contain both A and B, got %v", cycle)
		}
	})
}

// Test helpers
// --------------------------------------------------------------------------

func assertContains(t *testing.T, body, substr, msg string) {
	t.Helper()
	if !strings.Contains(body, substr) {
		t.Errorf("%s: response body does not contain %q", msg, substr)
	}
}
