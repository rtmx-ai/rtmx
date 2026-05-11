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

func TestServeCommand(t *testing.T) {
	rtmx.Req(t, "REQ-GO-037")

	t.Run("flags_registered", func(t *testing.T) {
		// Verify flags exist on the real serveCmd
		if serveCmd.Flags().Lookup("port") == nil {
			t.Error("serve should have --port flag")
		}
		if serveCmd.Flags().Lookup("auth") == nil {
			t.Error("serve should have --auth flag")
		}
		if serveCmd.Flags().Lookup("sync-url") == nil {
			t.Error("serve should have --sync-url flag")
		}
	})

	t.Run("dashboard_mux_status_api", func(t *testing.T) {
		db := database.NewDatabase()
		_ = db.Add(&database.Requirement{
			ReqID:           "REQ-TEST-001",
			Category:        "TEST",
			RequirementText: "Test requirement",
			Status:          database.StatusComplete,
		})
		_ = db.Add(&database.Requirement{
			ReqID:           "REQ-TEST-002",
			Category:        "TEST",
			RequirementText: "Another test",
			Status:          database.StatusMissing,
		})

		cfg := &config.Config{}
		mux := NewDashboardMux(db, cfg)

		// Test /api/status endpoint
		req := httptest.NewRequest("GET", "/api/status", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status code = %d, want 200", w.Code)
		}

		body := w.Body.String()
		if !strings.Contains(body, `"total":2`) {
			t.Errorf("expected total:2, got: %s", body)
		}
		if !strings.Contains(body, `"complete":1`) {
			t.Errorf("expected complete:1, got: %s", body)
		}
	})

	t.Run("dashboard_mux_health_api", func(t *testing.T) {
		db := database.NewDatabase()
		cfg := &config.Config{}
		mux := NewDashboardMux(db, cfg)

		req := httptest.NewRequest("GET", "/api/health", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status code = %d, want 200", w.Code)
		}
		if !strings.Contains(w.Body.String(), `"status":"ok"`) {
			t.Errorf("expected ok status, got: %s", w.Body.String())
		}
	})

	t.Run("dashboard_mux_html", func(t *testing.T) {
		db := database.NewDatabase()
		cfg := &config.Config{}
		mux := NewDashboardMux(db, cfg)

		req := httptest.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status code = %d, want 200", w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, "RTMX Dashboard") {
			t.Error("expected dashboard HTML")
		}
		if !strings.Contains(body, "text/html") {
			ct := w.Header().Get("Content-Type")
			if !strings.Contains(ct, "text/html") {
				t.Errorf("expected text/html content type, got %s", ct)
			}
		}
	})
}
