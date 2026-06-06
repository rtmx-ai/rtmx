package adapters

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestSlackAdapter validates the Slack adapter functionality including
// notification sending, channel routing, status updates, auth, and error handling.
func TestSlackAdapter(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-010")

	t.Run("name_returns_slack", func(t *testing.T) {
		adapter := newTestSlackAdapter(t, nil)
		if adapter.Name() != "slack" {
			t.Errorf("Name() = %q, want %q", adapter.Name(), "slack")
		}
	})

	t.Run("is_configured_with_channels", func(t *testing.T) {
		adapter := newTestSlackAdapter(t, nil)
		if !adapter.IsConfigured() {
			t.Error("IsConfigured() = false, want true")
		}
	})

	t.Run("is_not_configured_without_channels", func(t *testing.T) {
		cfg := &config.SlackAdapterConfig{
			Enabled:  true,
			TokenEnv: "TEST_SLACK_TOKEN",
			Channels: map[string]string{},
		}
		t.Setenv("TEST_SLACK_TOKEN", "xoxb-test-token")
		adapter, err := NewSlackAdapter(cfg)
		if err != nil {
			t.Fatalf("NewSlackAdapter: %v", err)
		}
		if adapter.IsConfigured() {
			t.Error("IsConfigured() = true, want false when no channels configured")
		}
	})

	t.Run("disabled_adapter_returns_error", func(t *testing.T) {
		cfg := &config.SlackAdapterConfig{Enabled: false}
		_, err := NewSlackAdapter(cfg)
		if err == nil {
			t.Error("expected error when adapter is disabled")
		}
	})

	t.Run("missing_token_returns_error", func(t *testing.T) {
		cfg := &config.SlackAdapterConfig{
			Enabled:  true,
			TokenEnv: "TEST_SLACK_TOKEN_MISSING",
		}
		t.Setenv("TEST_SLACK_TOKEN_MISSING", "")
		_, err := NewSlackAdapter(cfg)
		if err == nil {
			t.Error("expected error when token is missing")
		}
	})

	t.Run("send_notification_posts_to_correct_channel", func(t *testing.T) {
		var receivedChannel, receivedText string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/chat.postMessage" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.Header.Get("Authorization") != "Bearer xoxb-test-token" {
				t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
			}
			var payload slackPostMessageRequest
			_ = json.NewDecoder(r.Body).Decode(&payload)
			receivedChannel = payload.Channel
			receivedText = payload.Text
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(slackPostMessageResponse{OK: true})
		}))
		defer server.Close()

		adapter := newTestSlackAdapter(t, nil)
		adapter.SetAPIURL(server.URL)

		err := adapter.SendNotification("status", "hello world")
		if err != nil {
			t.Fatalf("SendNotification: %v", err)
		}
		if receivedChannel != "#rtmx-status" {
			t.Errorf("channel = %q, want %q", receivedChannel, "#rtmx-status")
		}
		if receivedText != "hello world" {
			t.Errorf("text = %q, want %q", receivedText, "hello world")
		}
	})

	t.Run("send_notification_routes_release_event", func(t *testing.T) {
		var receivedChannel string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var payload slackPostMessageRequest
			_ = json.NewDecoder(r.Body).Decode(&payload)
			receivedChannel = payload.Channel
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(slackPostMessageResponse{OK: true})
		}))
		defer server.Close()

		adapter := newTestSlackAdapter(t, nil)
		adapter.SetAPIURL(server.URL)

		err := adapter.SendNotification("release", "v1.0.0 released")
		if err != nil {
			t.Fatalf("SendNotification: %v", err)
		}
		if receivedChannel != "#rtmx-releases" {
			t.Errorf("channel = %q, want %q", receivedChannel, "#rtmx-releases")
		}
	})

	t.Run("send_notification_unknown_event_returns_error", func(t *testing.T) {
		adapter := newTestSlackAdapter(t, nil)
		err := adapter.SendNotification("unknown-event", "message")
		if err == nil {
			t.Error("expected error for unmapped event")
		}
	})

	t.Run("send_notification_api_error_response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(slackPostMessageResponse{OK: false, Error: "channel_not_found"})
		}))
		defer server.Close()

		adapter := newTestSlackAdapter(t, nil)
		adapter.SetAPIURL(server.URL)

		err := adapter.SendNotification("status", "hello")
		if err == nil {
			t.Error("expected error when Slack API returns error")
		}
	})

	t.Run("send_notification_http_error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		adapter := newTestSlackAdapter(t, nil)
		adapter.SetAPIURL(server.URL)

		err := adapter.SendNotification("status", "hello")
		if err == nil {
			t.Error("expected error on HTTP 500")
		}
	})

	t.Run("send_status_update", func(t *testing.T) {
		var receivedText string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var payload slackPostMessageRequest
			_ = json.NewDecoder(r.Body).Decode(&payload)
			receivedText = payload.Text
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(slackPostMessageResponse{OK: true})
		}))
		defer server.Close()

		adapter := newTestSlackAdapter(t, nil)
		adapter.SetAPIURL(server.URL)

		db := newTestSlackDB()
		err := adapter.SendStatusUpdate(db)
		if err != nil {
			t.Fatalf("SendStatusUpdate: %v", err)
		}
		// DB has 2 complete, 1 partial, 1 missing = 4 total, 50%
		if receivedText == "" {
			t.Error("expected non-empty status message")
		}
		// Verify message contains key stats
		for _, substr := range []string{"2/4", "50%", "1 partial", "1 missing"} {
			if !contains(receivedText, substr) {
				t.Errorf("status message %q should contain %q", receivedText, substr)
			}
		}
	})

	t.Run("send_status_update_no_status_channel", func(t *testing.T) {
		cfg := &config.SlackAdapterConfig{
			Enabled:  true,
			TokenEnv: "TEST_SLACK_TOKEN",
			Channels: map[string]string{"release": "#releases"},
		}
		t.Setenv("TEST_SLACK_TOKEN", "xoxb-test-token")
		adapter, _ := NewSlackAdapter(cfg)
		db := newTestSlackDB()
		err := adapter.SendStatusUpdate(db)
		if err == nil {
			t.Error("expected error when no status channel configured")
		}
	})

	t.Run("default_token_env", func(t *testing.T) {
		cfg := &config.SlackAdapterConfig{
			Enabled:  true,
			TokenEnv: "", // should default to SLACK_BOT_TOKEN
			Channels: map[string]string{"status": "#test"},
		}
		t.Setenv("SLACK_BOT_TOKEN", "xoxb-default")
		adapter, err := NewSlackAdapter(cfg)
		if err != nil {
			t.Fatalf("NewSlackAdapter: %v", err)
		}
		if adapter.token != "xoxb-default" {
			t.Errorf("token = %q, want %q", adapter.token, "xoxb-default")
		}
	})

	t.Run("custom_env_getter", func(t *testing.T) {
		cfg := &config.SlackAdapterConfig{
			Enabled:  true,
			TokenEnv: "MY_TOKEN",
			Channels: map[string]string{"status": "#test"},
		}
		adapter, err := NewSlackAdapter(cfg, WithEnvGetter(func(key string) string {
			if key == "MY_TOKEN" {
				return "custom-token"
			}
			return ""
		}))
		if err != nil {
			t.Fatalf("NewSlackAdapter: %v", err)
		}
		if adapter.token != "custom-token" {
			t.Errorf("token = %q, want %q", adapter.token, "custom-token")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func newTestSlackAdapter(t *testing.T, channels map[string]string) *SlackAdapter {
	t.Helper()
	if channels == nil {
		channels = map[string]string{
			"status":  "#rtmx-status",
			"release": "#rtmx-releases",
		}
	}
	cfg := &config.SlackAdapterConfig{
		Enabled:  true,
		TokenEnv: "TEST_SLACK_TOKEN",
		Channels: channels,
	}
	t.Setenv("TEST_SLACK_TOKEN", "xoxb-test-token")
	adapter, err := NewSlackAdapter(cfg)
	if err != nil {
		t.Fatalf("NewSlackAdapter: %v", err)
	}
	return adapter
}

func newTestSlackDB() *database.Database {
	db := database.NewDatabase()
	reqs := []*database.Requirement{
		{ReqID: "REQ-TEST-001", Category: "TEST", Status: database.StatusComplete, Dependencies: make(database.StringSet), Blocks: make(database.StringSet)},
		{ReqID: "REQ-TEST-002", Category: "TEST", Status: database.StatusComplete, Dependencies: make(database.StringSet), Blocks: make(database.StringSet)},
		{ReqID: "REQ-TEST-003", Category: "TEST", Status: database.StatusPartial, Dependencies: make(database.StringSet), Blocks: make(database.StringSet)},
		{ReqID: "REQ-TEST-004", Category: "TEST", Status: database.StatusMissing, Dependencies: make(database.StringSet), Blocks: make(database.StringSet)},
	}
	for _, r := range reqs {
		_ = db.Add(r)
	}
	return db
}
