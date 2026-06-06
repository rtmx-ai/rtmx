package adapters

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestWebhookAdapter validates webhook delivery, retry, HMAC signature, and event filtering.
func TestWebhookAdapter(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-012")

	t.Run("name_returns_webhook", func(t *testing.T) {
		adapter := newTestWebhookAdapter(t, "http://example.com", nil)
		if adapter.Name() != "webhook" {
			t.Errorf("Name() = %q, want %q", adapter.Name(), "webhook")
		}
	})

	t.Run("is_configured", func(t *testing.T) {
		adapter := newTestWebhookAdapter(t, "http://example.com", nil)
		if !adapter.IsConfigured() {
			t.Error("IsConfigured() = false, want true")
		}
	})

	t.Run("disabled_adapter_returns_error", func(t *testing.T) {
		cfg := &config.WebhookAdapterConfig{Enabled: false, URL: "http://example.com"}
		_, err := NewWebhookAdapter(cfg)
		if err == nil {
			t.Error("expected error when adapter is disabled")
		}
	})

	t.Run("missing_url_returns_error", func(t *testing.T) {
		cfg := &config.WebhookAdapterConfig{Enabled: true, URL: ""}
		_, err := NewWebhookAdapter(cfg)
		if err == nil {
			t.Error("expected error when URL is empty")
		}
	})

	t.Run("send_delivers_payload", func(t *testing.T) {
		var receivedBody []byte
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedBody, _ = io.ReadAll(r.Body)
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", r.Header.Get("Content-Type"))
			}
			if r.Header.Get("User-Agent") != "RTMX-Webhook/1.0" {
				t.Errorf("User-Agent = %q, want RTMX-Webhook/1.0", r.Header.Get("User-Agent"))
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		adapter := newTestWebhookAdapter(t, server.URL, nil)
		err := adapter.Send("status.updated", map[string]string{"req_id": "REQ-TEST-001"})
		if err != nil {
			t.Fatalf("Send: %v", err)
		}

		var envelope webhookPayload
		if err := json.Unmarshal(receivedBody, &envelope); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if envelope.Event != "status.updated" {
			t.Errorf("event = %q, want %q", envelope.Event, "status.updated")
		}
		if envelope.Timestamp == "" {
			t.Error("timestamp should not be empty")
		}
	})

	t.Run("send_with_hmac_signature", func(t *testing.T) {
		secret := "webhook-secret-key"
		var receivedSig string
		var receivedBody []byte

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedSig = r.Header.Get("X-RTMX-Signature")
			receivedBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		adapter := newTestWebhookAdapterWithSecret(t, server.URL, secret, nil)
		err := adapter.Send("test.event", map[string]string{"key": "value"})
		if err != nil {
			t.Fatalf("Send: %v", err)
		}

		// Verify signature
		if receivedSig == "" {
			t.Fatal("expected X-RTMX-Signature header")
		}

		// Compute expected signature
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(receivedBody)
		expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		if receivedSig != expected {
			t.Errorf("signature = %q, want %q", receivedSig, expected)
		}
	})

	t.Run("send_without_secret_has_no_signature", func(t *testing.T) {
		var receivedSig string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedSig = r.Header.Get("X-RTMX-Signature")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		adapter := newTestWebhookAdapter(t, server.URL, nil)
		_ = adapter.Send("test.event", "data")
		if receivedSig != "" {
			t.Errorf("signature = %q, want empty when no secret", receivedSig)
		}
	})

	t.Run("event_filtering_allows_configured_events", func(t *testing.T) {
		var called bool
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		events := []string{"status.updated", "release.created"}
		adapter := newTestWebhookAdapter(t, server.URL, events)

		err := adapter.Send("status.updated", "data")
		if err != nil {
			t.Fatalf("Send: %v", err)
		}
		if !called {
			t.Error("expected webhook to be called for allowed event")
		}
	})

	t.Run("event_filtering_skips_unconfigured_events", func(t *testing.T) {
		var called bool
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		events := []string{"status.updated"}
		adapter := newTestWebhookAdapter(t, server.URL, events)

		err := adapter.Send("unknown.event", "data")
		if err != nil {
			t.Fatalf("Send: %v", err)
		}
		if called {
			t.Error("expected webhook NOT to be called for unconfigured event")
		}
	})

	t.Run("empty_events_list_allows_all", func(t *testing.T) {
		var called bool
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		adapter := newTestWebhookAdapter(t, server.URL, nil)
		_ = adapter.Send("any.event", "data")
		if !called {
			t.Error("expected webhook to be called when events list is empty (allow all)")
		}
	})

	t.Run("retry_on_server_error", func(t *testing.T) {
		var attempts int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			n := atomic.AddInt32(&attempts, 1)
			if n < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := &config.WebhookAdapterConfig{
			Enabled:    true,
			URL:        server.URL,
			Events:     nil,
			MaxRetries: 3,
		}
		adapter, err := NewWebhookAdapter(cfg)
		if err != nil {
			t.Fatalf("NewWebhookAdapter: %v", err)
		}

		err = adapter.Send("test.event", "data")
		if err != nil {
			t.Fatalf("Send: %v", err)
		}
		if atomic.LoadInt32(&attempts) != 3 {
			t.Errorf("attempts = %d, want 3", atomic.LoadInt32(&attempts))
		}
	})

	t.Run("retry_exhausted_returns_error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		cfg := &config.WebhookAdapterConfig{
			Enabled:    true,
			URL:        server.URL,
			MaxRetries: 1,
		}
		adapter, _ := NewWebhookAdapter(cfg)

		err := adapter.Send("test.event", "data")
		if err == nil {
			t.Error("expected error after retries exhausted")
		}
	})

	t.Run("zero_retries_single_attempt", func(t *testing.T) {
		var attempts int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddInt32(&attempts, 1)
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		cfg := &config.WebhookAdapterConfig{
			Enabled:    true,
			URL:        server.URL,
			MaxRetries: 0,
		}
		adapter, _ := NewWebhookAdapter(cfg)
		_ = adapter.Send("test.event", "data")
		if atomic.LoadInt32(&attempts) != 1 {
			t.Errorf("attempts = %d, want 1", atomic.LoadInt32(&attempts))
		}
	})

	t.Run("custom_http_client", func(t *testing.T) {
		var called bool
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := &config.WebhookAdapterConfig{
			Enabled: true,
			URL:     server.URL,
		}
		adapter, err := NewWebhookAdapter(cfg, WithHTTPClient(server.Client()))
		if err != nil {
			t.Fatalf("NewWebhookAdapter: %v", err)
		}
		_ = adapter.Send("test", "data")
		if !called {
			t.Error("expected custom HTTP client to be used")
		}
	})
}

func newTestWebhookAdapter(t *testing.T, url string, events []string) *WebhookAdapter {
	t.Helper()
	cfg := &config.WebhookAdapterConfig{
		Enabled:    true,
		URL:        url,
		Events:     events,
		MaxRetries: 0,
	}
	adapter, err := NewWebhookAdapter(cfg)
	if err != nil {
		t.Fatalf("NewWebhookAdapter: %v", err)
	}
	return adapter
}

func newTestWebhookAdapterWithSecret(t *testing.T, url, secret string, events []string) *WebhookAdapter {
	t.Helper()
	cfg := &config.WebhookAdapterConfig{
		Enabled:    true,
		URL:        url,
		SecretEnv:  "TEST_WEBHOOK_SECRET",
		Events:     events,
		MaxRetries: 0,
	}
	adapter, err := NewWebhookAdapter(cfg, WithEnvGetter(func(key string) string {
		if key == "TEST_WEBHOOK_SECRET" {
			return secret
		}
		return ""
	}))
	if err != nil {
		t.Fatalf("NewWebhookAdapter: %v", err)
	}
	return adapter
}
