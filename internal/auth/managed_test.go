package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
	"time"
)

func TestManagedLoginCapturesSession(t *testing.T) {
	t.Parallel()

	var gotRedirect string
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/login/github", func(w http.ResponseWriter, r *http.Request) {
		gotRedirect = r.URL.Query().Get("cli_redirect")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"authorization_url": "http://idp.test/auth",
			"state":             "abc",
			"cli_redirect":      gotRedirect,
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "tokens.json")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan *ManagedLoginResult, 1)
	errCh := make(chan error, 1)
	go func() {
		res, err := ManagedLogin(ctx, ManagedLoginConfig{
			SyncURL:    srv.URL,
			TokenPath:  tokenPath,
			HTTPClient: srv.Client(),
			Browser: func(string) error {
				go func() {
					time.Sleep(30 * time.Millisecond)
					u, err := url.Parse(gotRedirect)
					if err != nil {
						errCh <- err
						return
					}
					q := u.Query()
					q.Set("session_token", "sess-123")
					q.Set("email", "dev@rtmx.ai")
					q.Set("display_name", "Dev")
					u.RawQuery = q.Encode()
					resp, err := http.Get(u.String())
					if err != nil {
						errCh <- err
						return
					}
					_ = resp.Body.Close()
				}()
				return nil
			},
		})
		if err != nil {
			errCh <- err
			return
		}
		done <- res
	}()

	select {
	case res := <-done:
		if res.Tokens.AccessToken != "sess-123" {
			t.Fatalf("token = %q", res.Tokens.AccessToken)
		}
		if res.Tokens.Email != "dev@rtmx.ai" {
			t.Fatalf("email = %q", res.Tokens.Email)
		}
		if res.Tokens.Mode != "managed" {
			t.Fatalf("mode = %q", res.Tokens.Mode)
		}
		loaded, err := LoadStoredTokens(tokenPath)
		if err != nil {
			t.Fatal(err)
		}
		if loaded.AccessToken != "sess-123" {
			t.Fatalf("stored token = %q", loaded.AccessToken)
		}
	case err := <-errCh:
		t.Fatal(err)
	case <-ctx.Done():
		t.Fatal("timeout")
	}
}

func TestManagedCallbackHandlerRequiresToken(t *testing.T) {
	t.Parallel()
	ch := make(chan managedHandoff, 1)
	h := managedCallbackHandler(ch)
	req := httptest.NewRequest(http.MethodGet, "/callback", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", rr.Code)
	}
	got := <-ch
	if got.err == nil {
		t.Fatal("expected error")
	}
}
