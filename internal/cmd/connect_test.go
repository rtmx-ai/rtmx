package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/auth"
	"github.com/spf13/cobra"
)

func TestConnectUnsupportedProvider(t *testing.T) {
	cmd := &cobra.Command{}
	err := runConnect(cmd, []string{"jira"})
	if err == nil || !strings.Contains(err.Error(), "unsupported provider") {
		t.Fatalf("expected unsupported provider error, got %v", err)
	}
}

func TestConnectRequiresLogin(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// Ensure no tokens file exists.
	cmd := &cobra.Command{}
	err := runConnect(cmd, []string{"asana"})
	if err == nil || !strings.Contains(err.Error(), "not signed in") {
		t.Fatalf("expected not signed in error, got %v", err)
	}
}

func TestConnectionsList(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	tokenDir := filepath.Join(home, ".rtmx", "auth")
	if err := os.MkdirAll(tokenDir, 0o700); err != nil {
		t.Fatal(err)
	}
	tokens := &auth.TokenSet{AccessToken: "sess-token", Mode: "managed"}
	raw, _ := json.Marshal(tokens)
	if err := os.WriteFile(filepath.Join(tokenDir, "tokens.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/connections" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sess-token" {
			t.Fatalf("auth = %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"connections": []map[string]string{
				{"id": "c1", "provider": "asana", "status": "active", "display_label": "Asana (Ada)"},
			},
		})
	}))
	defer server.Close()
	t.Setenv("RTMX_SYNC_HTTP_URL", server.URL)

	cmd := &cobra.Command{}
	var buf strings.Builder
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	if err := runConnectionsList(cmd, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "asana") || !strings.Contains(out, "Asana (Ada)") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestConnectStartsAsanaOAuth(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	tokenDir := filepath.Join(home, ".rtmx", "auth")
	if err := os.MkdirAll(tokenDir, 0o700); err != nil {
		t.Fatal(err)
	}
	tokens := &auth.TokenSet{AccessToken: "sess-token", Mode: "managed"}
	raw, _ := json.Marshal(tokens)
	if err := os.WriteFile(filepath.Join(tokenDir, "tokens.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/connections/asana/start" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "redirect_uri") {
			t.Fatalf("body = %s", body)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"authorization_url": "https://app.asana.com/-/oauth_authorize?client_id=x",
			"state":             "st",
		})
	}))
	defer server.Close()
	t.Setenv("RTMX_SYNC_HTTP_URL", server.URL)

	// Avoid opening a real browser in CI.
	oldBrowser := connectBrowser
	connectBrowser = func(url string) error {
		if !strings.Contains(url, "asana.com") {
			t.Fatalf("url = %s", url)
		}
		return nil
	}
	t.Cleanup(func() { connectBrowser = oldBrowser })

	cmd := &cobra.Command{}
	var buf strings.Builder
	cmd.SetOut(&buf)
	if err := runConnect(cmd, []string{"asana"}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "authorize Asana") {
		t.Fatalf("output = %q", buf.String())
	}
}
