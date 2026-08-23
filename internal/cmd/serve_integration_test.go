package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/config"
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
// REQ-DASH-001: Embedded SPA with htmx, Alpine.js, and Tailwind via embed.FS.
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
		assertContains(t, body, `hx-trigger="keyup changed delay:300ms"`, "search should have debounced trigger")

		// htmx row click-through to detail
		assertContains(t, body, `hx-get="/partials/detail/REQ-CLI-001"`, "requirement row should link to detail partial")
		assertContains(t, body, `hx-push-url=`, "requirement rows should push URL for browser history")

		// Sortable column headers
		assertContains(t, body, `hx-get="/partials/requirements?sort=status"`, "status column should be sortable via htmx")
		assertContains(t, body, `hx-get="/partials/requirements?sort=priority"`, "priority column should be sortable via htmx")

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
		assertContains(t, body, `hx-target="#content"`, "release rows should target #content")

		// Data from test DB (sprints assigned)
		assertContains(t, body, "v1.0.0", "releases should show v1.0.0")
		assertContains(t, body, "v2.0.0", "releases should show v2.0.0")
		assertContains(t, body, "Gate: PASS", "v1.0.0 should show Gate: PASS")
		assertContains(t, body, "Gate: FAIL", "v2.0.0 should show Gate: FAIL")
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
// REQ-DASH-001: Embedded SPA with htmx, Alpine.js, and Tailwind via embed.FS.
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
		assertContains(t, body, "htmx.org", "SPA shell must include htmx script tag")
	})

	t.Run("shell_includes_alpinejs_script", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/app", nil))
		body := rec.Body.String()
		assertContains(t, body, "alpinejs", "SPA shell must include Alpine.js script tag")
	})

	t.Run("shell_includes_tailwind_css", func(t *testing.T) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest("GET", "/app", nil))
		body := rec.Body.String()
		assertContains(t, body, "tailwind", "SPA shell must include Tailwind CSS")
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
// Test helpers
// --------------------------------------------------------------------------

func assertContains(t *testing.T, body, substr, msg string) {
	t.Helper()
	if !strings.Contains(body, substr) {
		t.Errorf("%s: response body does not contain %q", msg, substr)
	}
}
