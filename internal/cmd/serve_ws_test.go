package cmd

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestDashboardWebSocket validates live update infrastructure.
// REQ-DASH-008: WebSocket live updates (polling fallback via htmx).
func TestDashboardWebSocket(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-008")

	db := testDBForDashboard(t)
	cfg := &config.Config{}
	mux := NewDashboardMuxWithPath(db, cfg, "")

	t.Run("agents_partial_has_polling_trigger", func(t *testing.T) {
		// The agents partial uses hx-trigger="every 10s" as polling fallback
		// for live updates until WebSocket is fully implemented.
		req := httptest.NewRequest("GET", "/partials/agents", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		body := rec.Body.String()

		if !strings.Contains(body, `hx-trigger="every 10s"`) {
			t.Error("agents partial should have hx-trigger='every 10s' for polling fallback")
		}
	})

	t.Run("agents_partial_targets_content", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/agents", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		body := rec.Body.String()

		if !strings.Contains(body, `hx-get="/partials/agents"`) {
			t.Error("agents partial should self-refresh via hx-get")
		}
	})

	t.Run("agents_partial_returns_html", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/partials/agents", nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		ct := rec.Header().Get("Content-Type")
		if !strings.Contains(ct, "text/html") {
			t.Errorf("Content-Type = %s, want text/html", ct)
		}

		body := rec.Body.String()
		if !strings.Contains(body, "Agent Monitor") {
			t.Error("agents partial should contain Agent Monitor heading")
		}
	})

	t.Run("status_partial_supports_htmx_refresh", func(t *testing.T) {
		// Verify partials are accessible for htmx-based live refresh
		partials := []string{
			"/partials/status",
			"/partials/requirements",
			"/partials/kanban",
			"/partials/health",
		}
		for _, path := range partials {
			req := httptest.NewRequest("GET", path, nil)
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("GET %s status = %d, want 200", path, rec.Code)
			}
			ct := rec.Header().Get("Content-Type")
			if !strings.Contains(ct, "text/html") {
				t.Errorf("GET %s Content-Type = %s, want text/html", path, ct)
			}
		}
	})

	t.Run("htmx_partial_request_header_accepted", func(t *testing.T) {
		// htmx sends HX-Request header; partials should work with or without it
		req := httptest.NewRequest("GET", "/partials/agents", nil)
		req.Header.Set("HX-Request", "true")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200 with HX-Request header", rec.Code)
		}
	})
}
