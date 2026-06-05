package cmd

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestDashboardAuthMiddleware validates authentication middleware.
// REQ-DASH-009: Auth middleware for API key/OAuth.
func TestDashboardAuthMiddleware(t *testing.T) {
	rtmx.Req(t, "REQ-DASH-009")

	// Simple handler that returns 200 OK
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	t.Run("auth_disabled_passes_all_requests", func(t *testing.T) {
		cfg := authConfig{Mode: ""}
		wrapped := authMiddleware(cfg)(okHandler)

		req := httptest.NewRequest("GET", "/api/requirements", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200 when auth disabled", rec.Code)
		}
	})

	t.Run("api_key_auth_rejects_unauthenticated", func(t *testing.T) {
		cfg := authConfig{Mode: "api-key", APIKey: "test-secret-key"}
		wrapped := authMiddleware(cfg)(okHandler)

		req := httptest.NewRequest("GET", "/api/requirements", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401 without auth header", rec.Code)
		}
	})

	t.Run("api_key_auth_rejects_wrong_key", func(t *testing.T) {
		cfg := authConfig{Mode: "api-key", APIKey: "test-secret-key"}
		wrapped := authMiddleware(cfg)(okHandler)

		req := httptest.NewRequest("GET", "/api/requirements", nil)
		req.Header.Set("Authorization", "Bearer wrong-key")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401 with wrong key", rec.Code)
		}
	})

	t.Run("api_key_auth_accepts_valid_key", func(t *testing.T) {
		cfg := authConfig{Mode: "api-key", APIKey: "test-secret-key"}
		wrapped := authMiddleware(cfg)(okHandler)

		req := httptest.NewRequest("GET", "/api/requirements", nil)
		req.Header.Set("Authorization", "Bearer test-secret-key")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200 with valid key", rec.Code)
		}
	})

	t.Run("api_key_auth_rejects_malformed_header", func(t *testing.T) {
		cfg := authConfig{Mode: "api-key", APIKey: "test-secret-key"}
		wrapped := authMiddleware(cfg)(okHandler)

		req := httptest.NewRequest("GET", "/api/requirements", nil)
		req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401 with non-Bearer auth", rec.Code)
		}
	})

	t.Run("api_key_auth_rejects_empty_bearer", func(t *testing.T) {
		cfg := authConfig{Mode: "api-key", APIKey: "test-secret-key"}
		wrapped := authMiddleware(cfg)(okHandler)

		req := httptest.NewRequest("GET", "/api/requirements", nil)
		req.Header.Set("Authorization", "Bearer ")
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401 with empty bearer token", rec.Code)
		}
	})

	t.Run("unknown_auth_mode_denies", func(t *testing.T) {
		cfg := authConfig{Mode: "unknown"}
		wrapped := authMiddleware(cfg)(okHandler)

		req := httptest.NewRequest("GET", "/api/requirements", nil)
		rec := httptest.NewRecorder()
		wrapped.ServeHTTP(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401 for unknown auth mode", rec.Code)
		}
	})

	t.Run("auth_works_with_different_http_methods", func(t *testing.T) {
		cfg := authConfig{Mode: "api-key", APIKey: "test-secret-key"}
		wrapped := authMiddleware(cfg)(okHandler)

		for _, method := range []string{"GET", "POST", "PATCH", "DELETE"} {
			req := httptest.NewRequest(method, "/api/requirements", nil)
			req.Header.Set("Authorization", "Bearer test-secret-key")
			rec := httptest.NewRecorder()
			wrapped.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("%s status = %d, want 200", method, rec.Code)
			}
		}
	})
}
