package cmd

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestSlackSlashCommands validates Slack slash command handler including
// /rtmx status, /rtmx backlog, /rtmx req commands, and signature verification.
func TestSlackSlashCommands(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-011")

	signingSecret := "test-signing-secret"
	db := testDBForDashboard(t)
	handler := handleSlackSlashCommand(db, signingSecret)

	t.Run("status_command", func(t *testing.T) {
		body := "text=status&command=%2Frtmx"
		resp := postSlackCommand(t, handler, body, signingSecret)
		if resp.ResponseType != "in_channel" {
			t.Errorf("response_type = %q, want %q", resp.ResponseType, "in_channel")
		}
		// Should contain header block
		if len(resp.Blocks) < 2 {
			t.Fatalf("blocks len = %d, want >= 2", len(resp.Blocks))
		}
		if resp.Blocks[0].Type != "header" {
			t.Errorf("first block type = %q, want header", resp.Blocks[0].Type)
		}
		if resp.Blocks[0].Text.Text != "RTMX Project Status" {
			t.Errorf("header = %q, want %q", resp.Blocks[0].Text.Text, "RTMX Project Status")
		}
		// Should contain stats in fields
		foundTotal := false
		for _, field := range resp.Blocks[1].Fields {
			if strings.Contains(field.Text, "*Total:* 5") {
				foundTotal = true
			}
		}
		if !foundTotal {
			t.Error("status should show total of 5 requirements")
		}
	})

	t.Run("empty_text_defaults_to_status", func(t *testing.T) {
		body := "text=&command=%2Frtmx"
		resp := postSlackCommand(t, handler, body, signingSecret)
		if len(resp.Blocks) < 1 {
			t.Fatal("expected blocks in response")
		}
		if resp.Blocks[0].Text == nil || resp.Blocks[0].Text.Text != "RTMX Project Status" {
			t.Error("empty text should default to status command")
		}
	})

	t.Run("backlog_command", func(t *testing.T) {
		body := "text=backlog&command=%2Frtmx"
		resp := postSlackCommand(t, handler, body, signingSecret)
		if resp.ResponseType != "in_channel" {
			t.Errorf("response_type = %q, want %q", resp.ResponseType, "in_channel")
		}
		if len(resp.Blocks) < 2 {
			t.Fatalf("blocks len = %d, want >= 2", len(resp.Blocks))
		}
		if resp.Blocks[0].Text.Text != "RTMX Backlog" {
			t.Errorf("header = %q, want %q", resp.Blocks[0].Text.Text, "RTMX Backlog")
		}
		// Should have items after the header and summary
		if len(resp.Blocks) < 3 {
			t.Error("backlog should contain at least one requirement item")
		}
	})

	t.Run("req_detail_command", func(t *testing.T) {
		body := "text=req+REQ-CLI-001&command=%2Frtmx"
		resp := postSlackCommand(t, handler, body, signingSecret)
		if resp.ResponseType != "in_channel" {
			t.Errorf("response_type = %q, want %q", resp.ResponseType, "in_channel")
		}
		if len(resp.Blocks) < 3 {
			t.Fatalf("blocks len = %d, want >= 3", len(resp.Blocks))
		}
		if resp.Blocks[0].Text.Text != "REQ-CLI-001" {
			t.Errorf("header = %q, want %q", resp.Blocks[0].Text.Text, "REQ-CLI-001")
		}
		// Description block
		if !strings.Contains(resp.Blocks[1].Text.Text, "Build CLI framework") {
			t.Error("req detail should contain requirement description")
		}
		// Fields block with status, priority, etc.
		foundStatus := false
		for _, field := range resp.Blocks[2].Fields {
			if strings.Contains(field.Text, "*Status:*") {
				foundStatus = true
			}
		}
		if !foundStatus {
			t.Error("req detail should show status field")
		}
	})

	t.Run("req_detail_shows_dependencies", func(t *testing.T) {
		body := "text=req+REQ-CLI-001&command=%2Frtmx"
		resp := postSlackCommand(t, handler, body, signingSecret)
		// Should have dependencies block (4th block)
		if len(resp.Blocks) < 4 {
			t.Fatalf("blocks len = %d, want >= 4", len(resp.Blocks))
		}
		foundBlocks := false
		for _, field := range resp.Blocks[3].Fields {
			if strings.Contains(field.Text, "*Blocks:*") && strings.Contains(field.Text, "REQ-CLI-002") {
				foundBlocks = true
			}
		}
		if !foundBlocks {
			t.Error("req detail should show blocks relationship to REQ-CLI-002")
		}
	})

	t.Run("req_detail_not_found", func(t *testing.T) {
		body := "text=req+REQ-NONEXISTENT&command=%2Frtmx"
		resp := postSlackCommand(t, handler, body, signingSecret)
		if resp.ResponseType != "ephemeral" {
			t.Errorf("response_type = %q, want ephemeral for errors", resp.ResponseType)
		}
		if !strings.Contains(resp.Blocks[0].Text.Text, "not found") {
			t.Error("should report requirement not found")
		}
	})

	t.Run("req_missing_id", func(t *testing.T) {
		body := "text=req&command=%2Frtmx"
		resp := postSlackCommand(t, handler, body, signingSecret)
		if resp.ResponseType != "ephemeral" {
			t.Errorf("response_type = %q, want ephemeral for usage error", resp.ResponseType)
		}
		if !strings.Contains(resp.Blocks[0].Text.Text, "Usage") {
			t.Error("should show usage message")
		}
	})

	t.Run("unknown_command", func(t *testing.T) {
		body := "text=foobar&command=%2Frtmx"
		resp := postSlackCommand(t, handler, body, signingSecret)
		if resp.ResponseType != "ephemeral" {
			t.Errorf("response_type = %q, want ephemeral for unknown command", resp.ResponseType)
		}
		if !strings.Contains(resp.Blocks[0].Text.Text, "Unknown command") {
			t.Error("should report unknown command")
		}
	})

	t.Run("valid_signature_accepted", func(t *testing.T) {
		body := "text=status&command=%2Frtmx"
		timestamp := fmt.Sprintf("%d", time.Now().Unix())
		sig := computeSlackSignature(timestamp, body, signingSecret)

		req := httptest.NewRequest("POST", "/slack/commands", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("X-Slack-Request-Timestamp", timestamp)
		req.Header.Set("X-Slack-Signature", sig)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", w.Code)
		}
	})

	t.Run("invalid_signature_rejected", func(t *testing.T) {
		body := "text=status&command=%2Frtmx"
		timestamp := fmt.Sprintf("%d", time.Now().Unix())

		req := httptest.NewRequest("POST", "/slack/commands", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("X-Slack-Request-Timestamp", timestamp)
		req.Header.Set("X-Slack-Signature", "v0=invalid")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", w.Code)
		}
	})

	t.Run("missing_signature_rejected", func(t *testing.T) {
		body := "text=status&command=%2Frtmx"

		req := httptest.NewRequest("POST", "/slack/commands", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401", w.Code)
		}
	})

	t.Run("stale_timestamp_rejected", func(t *testing.T) {
		body := "text=status&command=%2Frtmx"
		// 10 minutes ago, beyond the 5-minute window
		timestamp := fmt.Sprintf("%d", time.Now().Unix()-600)
		sig := computeSlackSignature(timestamp, body, signingSecret)

		req := httptest.NewRequest("POST", "/slack/commands", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("X-Slack-Request-Timestamp", timestamp)
		req.Header.Set("X-Slack-Signature", sig)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("status = %d, want 401 for stale timestamp", w.Code)
		}
	})

	t.Run("method_not_allowed_for_get", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/slack/commands", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want 405", w.Code)
		}
	})

	t.Run("no_signing_secret_skips_verification", func(t *testing.T) {
		noSecretHandler := handleSlackSlashCommand(db, "")
		body := "text=status&command=%2Frtmx"
		req := httptest.NewRequest("POST", "/slack/commands", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		noSecretHandler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("status = %d, want 200 when no signing secret", w.Code)
		}
	})

	t.Run("response_is_valid_json", func(t *testing.T) {
		body := "text=status&command=%2Frtmx"
		timestamp := fmt.Sprintf("%d", time.Now().Unix())
		sig := computeSlackSignature(timestamp, body, signingSecret)

		req := httptest.NewRequest("POST", "/slack/commands", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("X-Slack-Request-Timestamp", timestamp)
		req.Header.Set("X-Slack-Signature", sig)

		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if ct := w.Header().Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}
		var resp slackResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("response is not valid JSON: %v", err)
		}
	})
}

// postSlackCommand sends a signed Slack slash command and decodes the response.
func postSlackCommand(t *testing.T, handler http.HandlerFunc, body, signingSecret string) slackResponse {
	t.Helper()

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	sig := computeSlackSignature(timestamp, body, signingSecret)

	req := httptest.NewRequest("POST", "/slack/commands", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("X-Slack-Signature", sig)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body = %s", w.Code, w.Body.String())
	}

	var resp slackResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v\nbody: %s", err, w.Body.String())
	}
	return resp
}

// computeSlackSignature computes the expected Slack v0 signature.
func computeSlackSignature(timestamp, body, secret string) string {
	baseString := fmt.Sprintf("v0:%s:%s", timestamp, body)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(baseString))
	return "v0=" + hex.EncodeToString(mac.Sum(nil))
}
