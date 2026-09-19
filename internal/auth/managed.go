// Managed Sync browser login with localhost loopback handoff (REQ-GO-082 / MONO-021a).
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const DefaultManagedSyncURL = "https://sync.rtmx.ai"

// ManagedLoginConfig configures zero-config managed OAuth login.
type ManagedLoginConfig struct {
	SyncURL      string
	Provider     string // github (default) or google
	CallbackPort int    // 0 = ephemeral
	Browser      BrowserOpener
	HTTPClient   HTTPClient
	TokenPath    string
}

// ManagedLoginResult is returned after a successful loopback handoff.
type ManagedLoginResult struct {
	Tokens *TokenSet
}

// ManagedLogin opens the system browser to the managed sync OAuth start
// endpoint, waits for the website callback to redirect to localhost with a
// session_token, and stores the credential.
func ManagedLogin(ctx context.Context, cfg ManagedLoginConfig) (*ManagedLoginResult, error) {
	syncURL := strings.TrimRight(cfg.SyncURL, "/")
	if syncURL == "" {
		syncURL = DefaultManagedSyncURL
	}
	provider := cfg.Provider
	if provider == "" {
		provider = "github"
	}
	browser := cfg.Browser
	if browser == nil {
		return nil, fmt.Errorf("browser opener is required")
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	tokenPath := cfg.TokenPath
	if tokenPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("home directory: %w", err)
		}
		tokenPath = filepath.Join(home, ".rtmx", "auth", "tokens.json")
	}

	port := cfg.CallbackPort
	listenAddr := "127.0.0.1:0"
	if port > 0 {
		listenAddr = fmt.Sprintf("127.0.0.1:%d", port)
	}
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, fmt.Errorf("start loopback listener: %w", err)
	}
	defer func() { _ = listener.Close() }()
	actualPort := listener.Addr().(*net.TCPAddr).Port
	cliRedirect := fmt.Sprintf("http://127.0.0.1:%d/callback", actualPort)

	resultCh := make(chan managedHandoff, 1)
	server := &http.Server{
		Handler: managedCallbackHandler(resultCh),
	}
	go func() { _ = server.Serve(listener) }()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	startURL := fmt.Sprintf(
		"%s/auth/login/%s?cli_redirect=%s&next=cli",
		syncURL,
		url.PathEscape(provider),
		url.QueryEscape(cliRedirect),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, startURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("start managed login: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("managed login start failed (%d): %s", resp.StatusCode, truncate(string(body), 200))
	}
	var start struct {
		AuthorizationURL string `json:"authorization_url"`
	}
	if err := json.Unmarshal(body, &start); err != nil {
		return nil, fmt.Errorf("parse login start: %w", err)
	}
	if start.AuthorizationURL == "" {
		return nil, fmt.Errorf("login start missing authorization_url")
	}

	if err := browser(start.AuthorizationURL); err != nil {
		return nil, fmt.Errorf("open browser: %w", err)
	}

	var handoff managedHandoff
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case handoff = <-resultCh:
	}
	if handoff.err != nil {
		return nil, handoff.err
	}
	if handoff.sessionToken == "" {
		return nil, fmt.Errorf("loopback callback missing session_token")
	}

	tokens := &TokenSet{
		AccessToken:  handoff.sessionToken,
		TokenType:    "Bearer",
		Email:        handoff.email,
		DisplayName:  handoff.displayName,
		SyncURL:      syncURL,
		Mode:         "managed",
		Provider:     provider,
		ExpiresAt:    0, // managed sessions are opaque; refresh via re-login
	}
	client := &OIDCClient{TokenStorePath: tokenPath}
	if err := client.SaveTokens(tokens); err != nil {
		return nil, err
	}
	return &ManagedLoginResult{Tokens: tokens}, nil
}

type managedHandoff struct {
	sessionToken string
	email        string
	displayName  string
	err          error
}

func managedCallbackHandler(ch chan<- managedHandoff) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/callback" {
			http.NotFound(w, r)
			return
		}
		q := r.URL.Query()
		if errMsg := q.Get("error"); errMsg != "" {
			msg := errMsg
			if d := q.Get("error_description"); d != "" {
				msg = d
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte("<html><body><h1>Sign-in failed</h1><p>You can close this window.</p></body></html>"))
			ch <- managedHandoff{err: fmt.Errorf("oauth error: %s", msg)}
			return
		}
		token := q.Get("session_token")
		if token == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("missing session_token"))
			ch <- managedHandoff{err: fmt.Errorf("missing session_token in callback")}
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<html><body><h1>Signed in to RTMX</h1><p>You can close this window and return to the CLI.</p></body></html>"))
		ch <- managedHandoff{
			sessionToken: token,
			email:        q.Get("email"),
			displayName:  q.Get("display_name"),
		}
	})
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// DefaultTokenStorePath returns ~/.rtmx/auth/tokens.json.
func DefaultTokenStorePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".rtmx", "auth", "tokens.json"), nil
}

// LoadStoredTokens loads tokens from the default or given path.
func LoadStoredTokens(path string) (*TokenSet, error) {
	if path == "" {
		var err error
		path, err = DefaultTokenStorePath()
		if err != nil {
			return nil, err
		}
	}
	c := &OIDCClient{TokenStorePath: path}
	return c.LoadTokens()
}

// ClearStoredTokens removes the token store at path (default if empty).
func ClearStoredTokens(path string) error {
	if path == "" {
		var err error
		path, err = DefaultTokenStorePath()
		if err != nil {
			return err
		}
	}
	c := &OIDCClient{TokenStorePath: path}
	return c.ClearTokens()
}
