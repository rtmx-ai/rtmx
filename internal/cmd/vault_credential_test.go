package cmd

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/auth"
)

func TestFetchManagedConnectionCredential(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	tokenDir := filepath.Join(home, ".rtmx", "auth")
	if err := os.MkdirAll(tokenDir, 0o700); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(&auth.TokenSet{AccessToken: "sess", Mode: "managed"})
	if err := os.WriteFile(filepath.Join(tokenDir, "tokens.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/connections/credential" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer sess" {
			t.Fatalf("auth = %q", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"connection_id": "c1",
			"provider":      "github",
			"access_token":  "vault-gh",
		})
	}))
	defer server.Close()
	t.Setenv("RTMX_SYNC_HTTP_URL", server.URL)

	tok, err := fetchManagedConnectionCredential("github")
	if err != nil {
		t.Fatal(err)
	}
	if tok != "vault-gh" {
		t.Fatalf("tok = %q", tok)
	}
}

func TestFetchManagedConnectionCredentialNotFound(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	tokenDir := filepath.Join(home, ".rtmx", "auth")
	_ = os.MkdirAll(tokenDir, 0o700)
	raw, _ := json.Marshal(&auth.TokenSet{AccessToken: "sess", Mode: "managed"})
	_ = os.WriteFile(filepath.Join(tokenDir, "tokens.json"), raw, 0o600)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"detail":"Connection not found"}`))
	}))
	defer server.Close()
	t.Setenv("RTMX_SYNC_HTTP_URL", server.URL)

	_, err := fetchManagedConnectionCredential("github")
	if err == nil || !strings.Contains(err.Error(), "rtmx connect github") {
		t.Fatalf("err = %v", err)
	}
	if !strings.Contains(err.Error(), "share") {
		t.Fatalf("err should mention org share: %v", err)
	}
}

func TestFetchManagedConnectionCredentialSharedOrg(t *testing.T) {
	// REQ-MONO-049c: shared-org grant (via_shared_org_id) still yields token.
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	tokenDir := filepath.Join(home, ".rtmx", "auth")
	if err := os.MkdirAll(tokenDir, 0o700); err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(&auth.TokenSet{AccessToken: "sess", Mode: "managed"})
	if err := os.WriteFile(filepath.Join(tokenDir, "tokens.json"), raw, 0o600); err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/connections/credential" {
			t.Fatalf("%s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"connection_id":      "c-shared",
			"provider":           "github",
			"access_token":       "vault-shared-gh",
			"via_shared_org_id":  "org-acme",
		})
	}))
	defer server.Close()
	t.Setenv("RTMX_SYNC_HTTP_URL", server.URL)

	tok, err := fetchManagedConnectionCredential("github")
	if err != nil {
		t.Fatal(err)
	}
	if tok != "vault-shared-gh" {
		t.Fatalf("tok = %q", tok)
	}
}

func TestFetchManagedConnectionCredentialUnauthorizedSilentRetry(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	tokenDir := filepath.Join(home, ".rtmx", "auth")
	_ = os.MkdirAll(tokenDir, 0o700)
	raw, _ := json.Marshal(&auth.TokenSet{
		AccessToken: "expired",
		Mode:        "managed",
		Provider:    "google",
	})
	_ = os.WriteFile(filepath.Join(tokenDir, "tokens.json"), raw, 0o600)

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"connection_id": "c1",
			"provider":      "github",
			"access_token":  "vault-after-silent",
		})
	}))
	defer server.Close()
	t.Setenv("RTMX_SYNC_HTTP_URL", server.URL)

	prev := silentManagedRestore
	silentManagedRestore = func(tok *auth.TokenSet) error {
		fresh, _ := json.Marshal(&auth.TokenSet{
			AccessToken: "fresh",
			Mode:        "managed",
			Provider:    "google",
		})
		return os.WriteFile(filepath.Join(tokenDir, "tokens.json"), fresh, 0o600)
	}
	defer func() { silentManagedRestore = prev }()

	tok, err := fetchManagedConnectionCredential("github")
	if err != nil {
		t.Fatal(err)
	}
	if tok != "vault-after-silent" {
		t.Fatalf("tok = %q", tok)
	}
	if calls != 2 {
		t.Fatalf("calls = %d", calls)
	}
}

func TestFetchManagedConnectionCredentialForbidden(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	tokenDir := filepath.Join(home, ".rtmx", "auth")
	_ = os.MkdirAll(tokenDir, 0o700)
	raw, _ := json.Marshal(&auth.TokenSet{AccessToken: "sess", Mode: "managed"})
	_ = os.WriteFile(filepath.Join(tokenDir, "tokens.json"), raw, 0o600)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"detail":"Forbidden"}`))
	}))
	defer server.Close()
	t.Setenv("RTMX_SYNC_HTTP_URL", server.URL)

	_, err := fetchManagedConnectionCredential("github")
	if err == nil || !strings.Contains(err.Error(), "org owner/admin") {
		t.Fatalf("err = %v", err)
	}
}
