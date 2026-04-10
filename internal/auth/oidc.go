// Package auth provides generic OIDC authentication with PKCE for the RTMX CLI.
//
// It works with any standards-compliant OIDC provider (Zitadel, Okta, Auth0, etc.)
// by discovering endpoints via .well-known/openid-configuration.
package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// HTTPClient abstracts HTTP operations for testability.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// BrowserOpener abstracts opening a URL in the system browser.
type BrowserOpener func(url string) error

// OIDCClient implements the OIDC Authorization Code flow with PKCE.
type OIDCClient struct {
	Issuer       string
	ClientID     string
	Scopes       []string
	CallbackPort int

	// Injected dependencies for testability.
	HTTPClient    HTTPClient
	OpenBrowser   BrowserOpener
	TokenStorePath string

	// discoveredEndpoints caches the OIDC discovery result.
	discovery *oidcDiscovery
}

// oidcDiscovery holds endpoints from .well-known/openid-configuration.
type oidcDiscovery struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	UserinfoEndpoint      string `json:"userinfo_endpoint"`
	EndSessionEndpoint    string `json:"end_session_endpoint"`
	Issuer                string `json:"issuer"`
}

// TokenSet holds the tokens returned by the OIDC provider.
type TokenSet struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	ExpiresAt    int64  `json:"expires_at,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// IsExpired returns true if the access token has expired.
func (t *TokenSet) IsExpired() bool {
	if t.ExpiresAt == 0 {
		return false
	}
	return time.Now().Unix() >= t.ExpiresAt
}

// Option configures an OIDCClient.
type Option func(*OIDCClient)

// WithHTTPClient sets the HTTP client for the OIDC client.
func WithHTTPClient(c HTTPClient) Option {
	return func(o *OIDCClient) {
		o.HTTPClient = c
	}
}

// WithBrowserOpener sets the browser opener function.
func WithBrowserOpener(fn BrowserOpener) Option {
	return func(o *OIDCClient) {
		o.OpenBrowser = fn
	}
}

// WithTokenStorePath overrides the default token store path.
func WithTokenStorePath(path string) Option {
	return func(o *OIDCClient) {
		o.TokenStorePath = path
	}
}

// NewOIDCClient creates a new OIDC client with the given configuration.
func NewOIDCClient(issuer, clientID string, scopes []string, callbackPort int, opts ...Option) *OIDCClient {
	c := &OIDCClient{
		Issuer:       issuer,
		ClientID:     clientID,
		Scopes:       scopes,
		CallbackPort: callbackPort,
		HTTPClient:   http.DefaultClient,
		OpenBrowser:  defaultOpenBrowser,
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.TokenStorePath == "" {
		home, _ := os.UserHomeDir()
		c.TokenStorePath = filepath.Join(home, ".rtmx", "auth", "tokens.json")
	}

	return c
}

// Discover fetches the OIDC provider configuration from the well-known endpoint.
func (c *OIDCClient) Discover(ctx context.Context) (*oidcDiscovery, error) {
	if c.discovery != nil {
		return c.discovery, nil
	}

	discoveryURL := strings.TrimRight(c.Issuer, "/") + "/.well-known/openid-configuration"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create discovery request: %w", err)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OIDC discovery: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OIDC discovery returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read discovery response: %w", err)
	}

	var disc oidcDiscovery
	if err := json.Unmarshal(body, &disc); err != nil {
		return nil, fmt.Errorf("failed to parse discovery response: %w", err)
	}

	if disc.AuthorizationEndpoint == "" || disc.TokenEndpoint == "" {
		return nil, fmt.Errorf("OIDC discovery missing required endpoints")
	}

	c.discovery = &disc
	return &disc, nil
}

// Login performs the full OIDC Authorization Code flow with PKCE.
//
// It opens the browser to the authorization URL, starts a local HTTP
// server to receive the callback, and exchanges the authorization code
// for tokens. The tokens are stored in the token store.
func (c *OIDCClient) Login(ctx context.Context) (*TokenSet, error) {
	disc, err := c.Discover(ctx)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	// Generate PKCE verifier and challenge.
	verifier, err := generateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}
	challenge := computeCodeChallenge(verifier)

	// Generate state and nonce for CSRF protection.
	state, err := generateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}
	nonce, err := generateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Build authorization URL.
	authURL, err := buildAuthURL(disc.AuthorizationEndpoint, c.ClientID, c.redirectURI(), c.Scopes, state, nonce, challenge)
	if err != nil {
		return nil, fmt.Errorf("failed to build auth URL: %w", err)
	}

	// Start local callback server.
	codeCh := make(chan callbackResult, 1)
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", c.CallbackPort))
	if err != nil {
		return nil, fmt.Errorf("failed to start callback server on port %d: %w", c.CallbackPort, err)
	}

	srv := &http.Server{
		Handler: callbackHandler(state, codeCh),
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = srv.Serve(listener)
	}()

	// Open browser.
	if err := c.OpenBrowser(authURL); err != nil {
		_ = srv.Close()
		return nil, fmt.Errorf("failed to open browser: %w", err)
	}

	// Wait for callback or context cancellation.
	var result callbackResult
	select {
	case result = <-codeCh:
	case <-ctx.Done():
		_ = srv.Close()
		wg.Wait()
		return nil, ctx.Err()
	}

	// Shutdown server.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	wg.Wait()

	if result.err != nil {
		return nil, result.err
	}

	// Exchange code for tokens.
	tokens, err := c.exchangeCode(ctx, disc.TokenEndpoint, result.code, verifier)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	// Set expiry timestamp if expires_in is provided.
	if tokens.ExpiresIn > 0 && tokens.ExpiresAt == 0 {
		tokens.ExpiresAt = time.Now().Unix() + int64(tokens.ExpiresIn)
	}

	// Store tokens.
	if err := c.SaveTokens(tokens); err != nil {
		return nil, fmt.Errorf("failed to save tokens: %w", err)
	}

	return tokens, nil
}

// RefreshToken uses the refresh token to get a new access token.
func (c *OIDCClient) RefreshToken(ctx context.Context) (*TokenSet, error) {
	stored, err := c.LoadTokens()
	if err != nil {
		return nil, fmt.Errorf("no stored tokens: %w", err)
	}

	if stored.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	disc, err := c.Discover(ctx)
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {c.ClientID},
		"refresh_token": {stored.RefreshToken},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, disc.TokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("refresh request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokens TokenSet
	if err := json.Unmarshal(body, &tokens); err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}

	// Preserve refresh token if the provider did not return a new one.
	if tokens.RefreshToken == "" {
		tokens.RefreshToken = stored.RefreshToken
	}

	// Set expiry timestamp.
	if tokens.ExpiresIn > 0 && tokens.ExpiresAt == 0 {
		tokens.ExpiresAt = time.Now().Unix() + int64(tokens.ExpiresIn)
	}

	if err := c.SaveTokens(&tokens); err != nil {
		return nil, fmt.Errorf("failed to save refreshed tokens: %w", err)
	}

	return &tokens, nil
}

// GetAccessToken returns a valid access token, refreshing if needed.
func (c *OIDCClient) GetAccessToken(ctx context.Context) (string, error) {
	tokens, err := c.LoadTokens()
	if err != nil {
		return "", fmt.Errorf("not authenticated: %w", err)
	}

	if !tokens.IsExpired() {
		return tokens.AccessToken, nil
	}

	// Try to refresh.
	refreshed, err := c.RefreshToken(ctx)
	if err != nil {
		return "", fmt.Errorf("token expired and refresh failed: %w", err)
	}

	return refreshed.AccessToken, nil
}

// SaveTokens writes the token set to the token store file.
func (c *OIDCClient) SaveTokens(tokens *TokenSet) error {
	dir := filepath.Dir(c.TokenStorePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create token store directory: %w", err)
	}

	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tokens: %w", err)
	}

	if err := os.WriteFile(c.TokenStorePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write token store: %w", err)
	}

	return nil
}

// LoadTokens reads the token set from the token store file.
func (c *OIDCClient) LoadTokens() (*TokenSet, error) {
	data, err := os.ReadFile(c.TokenStorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read token store: %w", err)
	}

	var tokens TokenSet
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, fmt.Errorf("failed to parse token store: %w", err)
	}

	return &tokens, nil
}

// ClearTokens removes the stored tokens.
func (c *OIDCClient) ClearTokens() error {
	err := os.Remove(c.TokenStorePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove token store: %w", err)
	}
	return nil
}

// redirectURI returns the local callback URL.
func (c *OIDCClient) redirectURI() string {
	return fmt.Sprintf("http://127.0.0.1:%d/callback", c.CallbackPort)
}

// callbackResult holds the result from the OAuth callback.
type callbackResult struct {
	code string
	err  error
}

// callbackHandler returns an HTTP handler that receives the OAuth callback.
func callbackHandler(expectedState string, resultCh chan<- callbackResult) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only handle the callback path.
		if r.URL.Path != "/callback" {
			http.NotFound(w, r)
			return
		}

		// Check for error response from provider.
		if errCode := r.URL.Query().Get("error"); errCode != "" {
			desc := r.URL.Query().Get("error_description")
			msg := fmt.Sprintf("authorization error: %s", errCode)
			if desc != "" {
				msg += ": " + desc
			}
			w.Header().Set("Content-Type", "text/plain")
			_, _ = fmt.Fprintf(w, "Authentication Failed\n\n%s\n\nYou may close this window.", msg)
			resultCh <- callbackResult{err: fmt.Errorf("%s", msg)}
			return
		}

		// Validate state.
		state := r.URL.Query().Get("state")
		if state != expectedState {
			w.Header().Set("Content-Type", "text/plain")
			_, _ = fmt.Fprint(w, "Authentication Failed\n\nState mismatch.\n\nYou may close this window.")
			resultCh <- callbackResult{err: fmt.Errorf("state mismatch: expected %q, got %q", expectedState, state)}
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			w.Header().Set("Content-Type", "text/plain")
			_, _ = fmt.Fprint(w, "Authentication Failed\n\nNo authorization code received.\n\nYou may close this window.")
			resultCh <- callbackResult{err: fmt.Errorf("no authorization code in callback")}
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		_, _ = fmt.Fprint(w, "Authentication Successful\n\nYou may close this window and return to the CLI.")
		resultCh <- callbackResult{code: code}
	})
}

// exchangeCode exchanges an authorization code for tokens.
func (c *OIDCClient) exchangeCode(ctx context.Context, tokenEndpoint, code, verifier string) (*TokenSet, error) {
	data := url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {c.ClientID},
		"code":          {code},
		"redirect_uri":  {c.redirectURI()},
		"code_verifier": {verifier},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokens TokenSet
	if err := json.Unmarshal(body, &tokens); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &tokens, nil
}

// buildAuthURL constructs the OIDC authorization URL.
func buildAuthURL(endpoint, clientID, redirectURI string, scopes []string, state, nonce, codeChallenge string) (string, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid authorization endpoint: %w", err)
	}

	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", strings.Join(scopes, " "))
	q.Set("state", state)
	q.Set("nonce", nonce)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// generateCodeVerifier creates a cryptographically random PKCE code verifier.
// Per RFC 7636, the verifier is 43-128 characters from [A-Z][a-z][0-9]-._~
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// computeCodeChallenge computes the S256 code challenge from a verifier.
func computeCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// generateRandomString creates a URL-safe random string of the given byte length.
func generateRandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// defaultOpenBrowser is a no-op in production; overridden for real usage.
// Real browser opening is injected via WithBrowserOpener.
func defaultOpenBrowser(url string) error {
	return fmt.Errorf("browser opener not configured; visit: %s", url)
}
