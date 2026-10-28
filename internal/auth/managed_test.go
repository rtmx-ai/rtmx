package auth

import (
	"context"
	"encoding/json"
	"errors"
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
					q.Set("refresh_token", "rtmx_rt_test")
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
		if res.Tokens.RefreshToken != "rtmx_rt_test" {
			t.Fatalf("refresh = %q", res.Tokens.RefreshToken)
		}
		loaded, err := LoadStoredTokens(tokenPath)
		if err != nil {
			t.Fatal(err)
		}
		if loaded.AccessToken != "sess-123" {
			t.Fatalf("stored token = %q", loaded.AccessToken)
		}
		if loaded.RefreshToken != "rtmx_rt_test" {
			t.Fatalf("stored refresh = %q", loaded.RefreshToken)
		}
	case err := <-errCh:
		t.Fatal(err)
	case <-ctx.Done():
		t.Fatal("timeout")
	}
}

func TestRefreshManagedSession(t *testing.T) {
	tokenPath := filepath.Join(t.TempDir(), "tokens.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/auth/refresh" || r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"session_token":"sess-new","refresh_token":"rtmx_rt_new","email":"a@b.c","display_name":"A"}`))
	}))
	defer srv.Close()

	_ = (&OIDCClient{TokenStorePath: tokenPath}).SaveTokens(&TokenSet{
		AccessToken:  "sess-old",
		RefreshToken: "rtmx_rt_old",
		Mode:         "managed",
		Provider:     "google",
		SyncURL:      srv.URL,
	})

	got, err := RefreshManagedSession(context.Background(), srv.URL, "rtmx_rt_old", tokenPath)
	if err != nil {
		t.Fatal(err)
	}
	if got.AccessToken != "sess-new" || got.RefreshToken != "rtmx_rt_new" {
		t.Fatalf("got %+v", got)
	}
	loaded, err := LoadStoredTokens(tokenPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.AccessToken != "sess-new" || loaded.Provider != "google" {
		t.Fatalf("stored %+v", loaded)
	}
}

func TestManagedLoginSilentAddsPromptNone(t *testing.T) {
	t.Parallel()

	var gotPrompt, gotRedirect string
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/login/google", func(w http.ResponseWriter, r *http.Request) {
		gotPrompt = r.URL.Query().Get("prompt")
		gotRedirect = r.URL.Query().Get("cli_redirect")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"authorization_url": "http://idp.test/auth",
			"state":             "abc",
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	tokenPath := filepath.Join(t.TempDir(), "tokens.json")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan *ManagedLoginResult, 1)
	errCh := make(chan error, 1)
	go func() {
		res, err := ManagedLogin(ctx, ManagedLoginConfig{
			SyncURL:    srv.URL,
			Provider:   "google",
			Silent:     true,
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
					q.Set("session_token", "sess-silent")
					q.Set("email", "dev@rtmx.ai")
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
		if gotPrompt != "none" {
			t.Fatalf("prompt = %q", gotPrompt)
		}
		if res.Tokens.AccessToken != "sess-silent" {
			t.Fatalf("token = %q", res.Tokens.AccessToken)
		}
	case err := <-errCh:
		t.Fatal(err)
	case <-ctx.Done():
		t.Fatal("timeout")
	}
}

func TestManagedLoginSilentConflict(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/login/github", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"detail": map[string]any{"error": "login_required", "silent_supported": false},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	_, err := ManagedLogin(context.Background(), ManagedLoginConfig{
		SyncURL:    srv.URL,
		Provider:   "github",
		Silent:     true,
		TokenPath:  filepath.Join(t.TempDir(), "tokens.json"),
		HTTPClient: srv.Client(),
		Browser:    func(string) error { return nil },
	})
	if err == nil || !errors.Is(err, ErrSilentUnsupported) {
		t.Fatalf("err = %v", err)
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
