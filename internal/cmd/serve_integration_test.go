package cmd

import (
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

// Test helpers
// --------------------------------------------------------------------------

func assertContains(t *testing.T, body, substr, msg string) {
	t.Helper()
	if !strings.Contains(body, substr) {
		t.Errorf("%s: response body does not contain %q", msg, substr)
	}
}
