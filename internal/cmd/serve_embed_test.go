package cmd

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/dashboard"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestDashboardSPAEmbed validates the embedded SPA framework.
// REQ-DASH-001: Embedded SPA with htmx, Alpine.js, and Tailwind via embed.FS.
func TestDashboardSPAEmbed(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-001")

	t.Run("embed_fs_contains_static_assets", func(t *testing.T) {
		assets := dashboard.Assets()
		files := []string{
			"static/app.js",
			"static/styles.css",
			"templates/layout.html",
		}
		for _, f := range files {
			_, err := fs.Stat(assets, f)
			if err != nil {
				t.Errorf("embedded asset %q not found: %v", f, err)
			}
		}
	})

	t.Run("embed_fs_contains_partials", func(t *testing.T) {
		assets := dashboard.Assets()
		partials := []string{
			"templates/partials/status.html",
			"templates/partials/requirements.html",
			"templates/partials/detail.html",
			"templates/partials/graph.html",
			"templates/partials/kanban.html",
			"templates/partials/releases.html",
			"templates/partials/health.html",
			"templates/partials/agents.html",
		}
		for _, p := range partials {
			_, err := fs.Stat(assets, p)
			if err != nil {
				t.Errorf("embedded partial %q not found: %v", p, err)
			}
		}
	})

	t.Run("static_fs_serves_files", func(t *testing.T) {
		staticFS := dashboard.StaticFS()
		f, err := staticFS.Open("app.js")
		if err != nil {
			t.Fatalf("StaticFS cannot open app.js: %v", err)
		}
		_ = f.Close()
	})

	t.Run("layout_renders_html", func(t *testing.T) {
		var buf strings.Builder
		err := dashboard.RenderLayout(&buf, dashboard.LayoutData{
			Title:      "Test",
			ActivePage: "status",
			Content:    "<p>Hello</p>",
		})
		if err != nil {
			t.Fatalf("RenderLayout error: %v", err)
		}
		html := buf.String()
		if !strings.Contains(html, "<!DOCTYPE html>") {
			t.Error("layout should contain DOCTYPE")
		}
		if !strings.Contains(html, "htmx.org") {
			t.Error("layout should include htmx")
		}
		if !strings.Contains(html, "alpinejs") {
			t.Error("layout should include Alpine.js")
		}
		if !strings.Contains(html, "tailwind") {
			t.Error("layout should include Tailwind CSS")
		}
		if !strings.Contains(html, "<p>Hello</p>") {
			t.Error("layout should render content")
		}
		if !strings.Contains(html, "Test - RTMX Dashboard") {
			t.Error("layout should include title")
		}
	})

	t.Run("layout_contains_nav_links", func(t *testing.T) {
		var buf strings.Builder
		_ = dashboard.RenderLayout(&buf, dashboard.LayoutData{Title: "Test", ActivePage: "status"})
		html := buf.String()
		for _, link := range []string{"Status", "Requirements", "Graph", "Kanban", "Releases", "Health", "Agents"} {
			if !strings.Contains(html, link) {
				t.Errorf("nav should contain %q link", link)
			}
		}
	})

	t.Run("shell_route_serves_html", func(t *testing.T) {
		db := testDBForDashboard(t)
		cfg := &config.Config{}
		mux := NewDashboardMuxWithPath(db, cfg, "")

		req := httptest.NewRequest("GET", "/app", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("GET /app status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "<!DOCTYPE html>") {
			t.Error("/app should return full HTML page")
		}
		if !strings.Contains(body, "RTMX") {
			t.Error("/app should contain RTMX branding")
		}
	})

	t.Run("static_route_serves_css", func(t *testing.T) {
		db := testDBForDashboard(t)
		cfg := &config.Config{}
		mux := NewDashboardMuxWithPath(db, cfg, "")

		req := httptest.NewRequest("GET", "/static/styles.css", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("GET /static/styles.css status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, ".nav-link") {
			t.Error("styles.css should contain nav styles")
		}
	})

	t.Run("static_route_serves_js", func(t *testing.T) {
		db := testDBForDashboard(t)
		cfg := &config.Config{}
		mux := NewDashboardMuxWithPath(db, cfg, "")

		req := httptest.NewRequest("GET", "/static/app.js", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("GET /static/app.js status = %d, want 200", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "rtmxApp") {
			t.Error("app.js should contain rtmxApp function")
		}
	})

	t.Run("partial_status_returns_html", func(t *testing.T) {
		db := testDBForDashboard(t)
		cfg := &config.Config{}
		mux := NewDashboardMuxWithPath(db, cfg, "")

		req := httptest.NewRequest("GET", "/partials/status", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("GET /partials/status status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "Project Status") {
			t.Error("status partial should contain heading")
		}
		if !strings.Contains(body, "Complete") {
			t.Error("status partial should show completion stats")
		}
	})

	t.Run("client_side_routing_pages", func(t *testing.T) {
		db := testDBForDashboard(t)
		cfg := &config.Config{}
		mux := NewDashboardMuxWithPath(db, cfg, "")

		pages := []string{"/requirements", "/graph", "/kanban", "/releases", "/health", "/agents"}
		for _, page := range pages {
			req := httptest.NewRequest("GET", page, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("GET %s status = %d, want 200", page, rec.Code)
			}
			if !strings.Contains(rec.Body.String(), "<!DOCTYPE html>") {
				t.Errorf("GET %s should return full HTML shell", page)
			}
		}
	})

	t.Run("api_and_dashboard_coexist", func(t *testing.T) {
		db := testDBForDashboard(t)
		cfg := &config.Config{}
		mux := NewDashboardMuxWithPath(db, cfg, "")

		// API endpoint still works
		req := httptest.NewRequest("GET", "/api/health", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("API health check failed: %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"ok"`) {
			t.Error("API health should return ok")
		}

		// Dashboard shell also works
		req2 := httptest.NewRequest("GET", "/app", nil)
		rec2 := httptest.NewRecorder()
		mux.ServeHTTP(rec2, req2)
		if rec2.Code != http.StatusOK {
			t.Errorf("Dashboard shell failed: %d", rec2.Code)
		}
	})
}

func testDBForDashboard(t *testing.T) *database.Database {
	t.Helper()
	db := database.NewDatabase()
	reqs := []*database.Requirement{
		{ReqID: "REQ-CLI-001", Category: "CLI", RequirementText: "Build CLI framework", Status: database.StatusComplete, Priority: database.PriorityP0, Phase: 1, EffortWeeks: 1.0},
		{ReqID: "REQ-CLI-002", Category: "CLI", RequirementText: "Status command", Status: database.StatusComplete, Priority: database.PriorityHigh, Phase: 1, EffortWeeks: 0.5},
		{ReqID: "REQ-MCP-001", Category: "MCP", RequirementText: "MCP server", Status: database.StatusPartial, Priority: database.PriorityP0, Phase: 2, EffortWeeks: 2.0},
		{ReqID: "REQ-MCP-002", Category: "MCP", RequirementText: "MCP tools", Status: database.StatusMissing, Priority: database.PriorityHigh, Phase: 2, EffortWeeks: 1.0},
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
