package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// mockHTTPClient records requests and returns canned responses.
type mockHTTPClient struct {
	responses map[string]*http.Response
	requests  []*http.Request
}

func newMockHTTPClient() *mockHTTPClient {
	return &mockHTTPClient{
		responses: make(map[string]*http.Response),
	}
}

func (m *mockHTTPClient) addResponse(urlPrefix string, statusCode int, body string) {
	m.responses[urlPrefix] = &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.requests = append(m.requests, req)
	for prefix, resp := range m.responses {
		if strings.Contains(req.URL.String(), prefix) {
			return resp, nil
		}
	}
	return &http.Response{
		StatusCode: 404,
		Body:       io.NopCloser(strings.NewReader("not found")),
		Header:     make(http.Header),
	}, nil
}

// --- Discovery Tests ---

func TestDiscover(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	tests := []struct {
		name       string
		status     int
		body       string
		wantErr    bool
		errContain string
	}{
		{
			name:   "successful discovery",
			status: 200,
			body: `{
				"authorization_endpoint": "https://idp.example.com/authorize",
				"token_endpoint": "https://idp.example.com/token",
				"userinfo_endpoint": "https://idp.example.com/userinfo",
				"issuer": "https://idp.example.com"
			}`,
			wantErr: false,
		},
		{
			name:       "non-200 status",
			status:     500,
			body:       "server error",
			wantErr:    true,
			errContain: "status 500",
		},
		{
			name:       "missing required endpoints",
			status:     200,
			body:       `{"issuer": "https://idp.example.com"}`,
			wantErr:    true,
			errContain: "missing required endpoints",
		},
		{
			name:       "invalid JSON",
			status:     200,
			body:       "not json",
			wantErr:    true,
			errContain: "failed to parse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := newMockHTTPClient()
			mock.addResponse(".well-known/openid-configuration", tt.status, tt.body)

			client := NewOIDCClient("https://idp.example.com", "test-client", nil, 8765,
				WithHTTPClient(mock))

			disc, err := client.Discover(context.Background())
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContain != "" && !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errContain)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if disc.AuthorizationEndpoint != "https://idp.example.com/authorize" {
				t.Errorf("unexpected auth endpoint: %s", disc.AuthorizationEndpoint)
			}
			if disc.TokenEndpoint != "https://idp.example.com/token" {
				t.Errorf("unexpected token endpoint: %s", disc.TokenEndpoint)
			}
		})
	}
}

func TestDiscoverCachesResult(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	mock := newMockHTTPClient()
	mock.addResponse(".well-known/openid-configuration", 200, `{
		"authorization_endpoint": "https://idp.example.com/authorize",
		"token_endpoint": "https://idp.example.com/token",
		"issuer": "https://idp.example.com"
	}`)

	client := NewOIDCClient("https://idp.example.com", "test-client", nil, 8765,
		WithHTTPClient(mock))

	_, _ = client.Discover(context.Background())
	_, _ = client.Discover(context.Background())

	if len(mock.requests) != 1 {
		t.Errorf("expected 1 request (cached), got %d", len(mock.requests))
	}
}

// --- PKCE Tests ---

func TestGenerateCodeVerifier(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	verifier, err := generateCodeVerifier()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Base64-URL encoded 32 bytes = 43 characters
	if len(verifier) != 43 {
		t.Errorf("expected verifier length 43, got %d", len(verifier))
	}

	// Verify uniqueness
	v2, _ := generateCodeVerifier()
	if verifier == v2 {
		t.Error("two verifiers should not be equal")
	}
}

func TestComputeCodeChallenge(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	verifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := computeCodeChallenge(verifier)

	// The challenge should be non-empty and different from the verifier.
	if challenge == "" {
		t.Error("challenge should not be empty")
	}
	if challenge == verifier {
		t.Error("challenge should differ from verifier")
	}

	// Challenge should be URL-safe base64
	if strings.ContainsAny(challenge, "+/=") {
		t.Errorf("challenge contains non-URL-safe characters: %s", challenge)
	}
}

func TestBuildAuthURL(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	authURL, err := buildAuthURL(
		"https://idp.example.com/authorize",
		"my-client",
		"http://127.0.0.1:8765/callback",
		[]string{"openid", "profile", "email"},
		"test-state",
		"test-nonce",
		"test-challenge",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	u, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("invalid URL: %v", err)
	}

	q := u.Query()
	checks := map[string]string{
		"response_type":         "code",
		"client_id":             "my-client",
		"redirect_uri":          "http://127.0.0.1:8765/callback",
		"scope":                 "openid profile email",
		"state":                 "test-state",
		"nonce":                 "test-nonce",
		"code_challenge":        "test-challenge",
		"code_challenge_method": "S256",
	}
	for key, want := range checks {
		got := q.Get(key)
		if got != want {
			t.Errorf("param %s: got %q, want %q", key, got, want)
		}
	}
}

// --- Token Store Tests ---

func TestTokenStoreSaveLoadRoundTrip(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "auth", "tokens.json")

	client := NewOIDCClient("https://idp.example.com", "test", nil, 8765,
		WithTokenStorePath(storePath))

	original := &TokenSet{
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-456",
		IDToken:      "id-token-789",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		ExpiresAt:    time.Now().Add(time.Hour).Unix(),
		Scope:        "openid profile email",
	}

	// Save
	if err := client.SaveTokens(original); err != nil {
		t.Fatalf("SaveTokens failed: %v", err)
	}

	// Verify file permissions
	info, err := os.Stat(storePath)
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("expected 0600 permissions, got %o", info.Mode().Perm())
	}

	// Load
	loaded, err := client.LoadTokens()
	if err != nil {
		t.Fatalf("LoadTokens failed: %v", err)
	}

	if loaded.AccessToken != original.AccessToken {
		t.Errorf("AccessToken mismatch: got %q, want %q", loaded.AccessToken, original.AccessToken)
	}
	if loaded.RefreshToken != original.RefreshToken {
		t.Errorf("RefreshToken mismatch: got %q, want %q", loaded.RefreshToken, original.RefreshToken)
	}
	if loaded.IDToken != original.IDToken {
		t.Errorf("IDToken mismatch: got %q, want %q", loaded.IDToken, original.IDToken)
	}
	if loaded.ExpiresAt != original.ExpiresAt {
		t.Errorf("ExpiresAt mismatch: got %d, want %d", loaded.ExpiresAt, original.ExpiresAt)
	}
}

func TestTokenStoreLoadNotFound(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "nonexistent", "tokens.json")

	client := NewOIDCClient("https://idp.example.com", "test", nil, 8765,
		WithTokenStorePath(storePath))

	_, err := client.LoadTokens()
	if err == nil {
		t.Fatal("expected error for nonexistent token store")
	}
}

func TestTokenStoreClear(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "tokens.json")

	client := NewOIDCClient("https://idp.example.com", "test", nil, 8765,
		WithTokenStorePath(storePath))

	// Save a token
	_ = client.SaveTokens(&TokenSet{AccessToken: "test"})

	// Clear
	if err := client.ClearTokens(); err != nil {
		t.Fatalf("ClearTokens failed: %v", err)
	}

	// Verify deleted
	if _, err := os.Stat(storePath); !os.IsNotExist(err) {
		t.Error("token file should not exist after clear")
	}

	// Clearing again should not error
	if err := client.ClearTokens(); err != nil {
		t.Errorf("ClearTokens on missing file should not error: %v", err)
	}
}

// --- Token Expiry Tests ---

func TestTokenSetIsExpired(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	tests := []struct {
		name      string
		token     TokenSet
		wantExpired bool
	}{
		{
			name:        "no expiry set",
			token:       TokenSet{AccessToken: "test"},
			wantExpired: false,
		},
		{
			name:        "future expiry",
			token:       TokenSet{AccessToken: "test", ExpiresAt: time.Now().Add(time.Hour).Unix()},
			wantExpired: false,
		},
		{
			name:        "past expiry",
			token:       TokenSet{AccessToken: "test", ExpiresAt: time.Now().Add(-time.Hour).Unix()},
			wantExpired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.IsExpired(); got != tt.wantExpired {
				t.Errorf("IsExpired() = %v, want %v", got, tt.wantExpired)
			}
		})
	}
}

// --- Login Flow Test (with mock OIDC provider) ---

func TestOIDCLogin(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	// Start a mock OIDC provider.
	var mockProviderURL string
	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/.well-known/openid-configuration"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{
				"authorization_endpoint": "%s/authorize",
				"token_endpoint": "%s/token",
				"userinfo_endpoint": "%s/userinfo",
				"issuer": "%s"
			}`, mockProviderURL, mockProviderURL, mockProviderURL, mockProviderURL)

		case r.URL.Path == "/token":
			// Exchange code for tokens.
			_ = r.ParseForm()
			code := r.FormValue("code")
			verifier := r.FormValue("code_verifier")

			if code != "test-auth-code" {
				http.Error(w, `{"error":"invalid_grant"}`, 400)
				return
			}
			if verifier == "" {
				http.Error(w, `{"error":"invalid_request","error_description":"missing code_verifier"}`, 400)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"access_token": "mock-access-token",
				"refresh_token": "mock-refresh-token",
				"id_token": "mock-id-token",
				"token_type": "Bearer",
				"expires_in": 3600,
				"scope": "openid profile email"
			}`))

		default:
			http.NotFound(w, r)
		}
	}))
	defer mockProvider.Close()
	mockProviderURL = mockProvider.URL

	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "tokens.json")

	// We need a mock HTTP client that routes discovery and token requests to our test server.
	mockHTTP := &routingHTTPClient{
		server:    mockProvider,
		transport: http.DefaultTransport,
	}

	// The browser opener will simulate the user completing auth by
	// hitting the local callback server with a code.
	browserOpened := false
	mockBrowser := func(authURL string) error {
		browserOpened = true

		// Parse the auth URL to get state param.
		u, err := url.Parse(authURL)
		if err != nil {
			return err
		}
		state := u.Query().Get("state")
		redirectURI := u.Query().Get("redirect_uri")

		// Simulate the provider redirecting back with an auth code.
		callbackURL := fmt.Sprintf("%s?code=test-auth-code&state=%s", redirectURI, state)
		resp, err := http.Get(callbackURL)
		if err != nil {
			return err
		}
		resp.Body.Close()
		return nil
	}

	client := NewOIDCClient(mockProvider.URL, "test-client-id",
		[]string{"openid", "profile", "email"}, 0, // port 0 = ephemeral
		WithHTTPClient(mockHTTP),
		WithBrowserOpener(mockBrowser),
		WithTokenStorePath(storePath),
	)

	// Use port 0 for the callback, but we need to pick a free port.
	// The client uses CallbackPort. Let's set it to 0 which won't work
	// because the redirectURI needs to match. Let's pick a real free port.
	client.CallbackPort = findFreePort(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tokens, err := client.Login(ctx)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if !browserOpened {
		t.Error("browser was not opened")
	}

	if tokens.AccessToken != "mock-access-token" {
		t.Errorf("unexpected access token: %s", tokens.AccessToken)
	}
	if tokens.RefreshToken != "mock-refresh-token" {
		t.Errorf("unexpected refresh token: %s", tokens.RefreshToken)
	}
	if tokens.ExpiresAt == 0 {
		t.Error("ExpiresAt should be set from ExpiresIn")
	}

	// Verify tokens were stored.
	loaded, err := client.LoadTokens()
	if err != nil {
		t.Fatalf("failed to load stored tokens: %v", err)
	}
	if loaded.AccessToken != "mock-access-token" {
		t.Errorf("stored access token mismatch: %s", loaded.AccessToken)
	}
}

// --- Refresh Token Test ---

func TestRefreshToken(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/.well-known/openid-configuration"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{
				"authorization_endpoint": "%s/authorize",
				"token_endpoint": "%s/token",
				"issuer": "%s"
			}`, mockServerURL(r), mockServerURL(r), mockServerURL(r))

		case r.URL.Path == "/token":
			_ = r.ParseForm()
			if r.FormValue("grant_type") != "refresh_token" {
				http.Error(w, `{"error":"invalid_grant_type"}`, 400)
				return
			}
			if r.FormValue("refresh_token") != "stored-refresh-token" {
				http.Error(w, `{"error":"invalid_grant"}`, 400)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"access_token": "new-access-token",
				"token_type": "Bearer",
				"expires_in": 7200
			}`))

		default:
			http.NotFound(w, r)
		}
	}))
	defer mockProvider.Close()

	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "tokens.json")

	client := NewOIDCClient(mockProvider.URL, "test-client", nil, 8765,
		WithHTTPClient(&routingHTTPClient{server: mockProvider, transport: http.DefaultTransport}),
		WithTokenStorePath(storePath),
	)

	// Pre-store expired tokens with a refresh token.
	expired := &TokenSet{
		AccessToken:  "old-access-token",
		RefreshToken: "stored-refresh-token",
		ExpiresAt:    time.Now().Add(-time.Hour).Unix(),
	}
	if err := client.SaveTokens(expired); err != nil {
		t.Fatalf("failed to save initial tokens: %v", err)
	}

	ctx := context.Background()
	tokens, err := client.RefreshToken(ctx)
	if err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}

	if tokens.AccessToken != "new-access-token" {
		t.Errorf("unexpected access token: %s", tokens.AccessToken)
	}
	// Refresh token should be preserved from stored tokens.
	if tokens.RefreshToken != "stored-refresh-token" {
		t.Errorf("refresh token not preserved: %s", tokens.RefreshToken)
	}
	if tokens.ExpiresAt == 0 {
		t.Error("ExpiresAt should be set")
	}
}

func TestRefreshTokenNoRefreshToken(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "tokens.json")

	client := NewOIDCClient("https://idp.example.com", "test", nil, 8765,
		WithTokenStorePath(storePath),
		WithHTTPClient(newMockHTTPClient()),
	)

	// Store tokens without a refresh token.
	_ = client.SaveTokens(&TokenSet{
		AccessToken: "access-only",
		ExpiresAt:   time.Now().Add(-time.Hour).Unix(),
	})

	_, err := client.RefreshToken(context.Background())
	if err == nil {
		t.Fatal("expected error when no refresh token available")
	}
	if !strings.Contains(err.Error(), "no refresh token") {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- GetAccessToken Tests ---

func TestGetAccessTokenValid(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "tokens.json")

	client := NewOIDCClient("https://idp.example.com", "test", nil, 8765,
		WithTokenStorePath(storePath),
		WithHTTPClient(newMockHTTPClient()),
	)

	_ = client.SaveTokens(&TokenSet{
		AccessToken: "valid-token",
		ExpiresAt:   time.Now().Add(time.Hour).Unix(),
	})

	token, err := client.GetAccessToken(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "valid-token" {
		t.Errorf("unexpected token: %s", token)
	}
}

func TestGetAccessTokenNotAuthenticated(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "nonexistent", "tokens.json")

	client := NewOIDCClient("https://idp.example.com", "test", nil, 8765,
		WithTokenStorePath(storePath),
		WithHTTPClient(newMockHTTPClient()),
	)

	_, err := client.GetAccessToken(context.Background())
	if err == nil {
		t.Fatal("expected error when not authenticated")
	}
	if !strings.Contains(err.Error(), "not authenticated") {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- Callback Handler Tests ---

func TestCallbackHandlerSuccess(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	ch := make(chan callbackResult, 1)
	handler := callbackHandler("expected-state", ch)

	req := httptest.NewRequest("GET", "/callback?code=auth-code&state=expected-state", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	select {
	case result := <-ch:
		if result.err != nil {
			t.Errorf("unexpected error: %v", result.err)
		}
		if result.code != "auth-code" {
			t.Errorf("unexpected code: %s", result.code)
		}
	default:
		t.Fatal("no result received")
	}

	if w.Code != 200 {
		t.Errorf("unexpected status: %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Successful") {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}

func TestCallbackHandlerStateMismatch(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	ch := make(chan callbackResult, 1)
	handler := callbackHandler("expected-state", ch)

	req := httptest.NewRequest("GET", "/callback?code=auth-code&state=wrong-state", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	select {
	case result := <-ch:
		if result.err == nil {
			t.Fatal("expected error for state mismatch")
		}
		if !strings.Contains(result.err.Error(), "state mismatch") {
			t.Errorf("unexpected error: %v", result.err)
		}
	default:
		t.Fatal("no result received")
	}
}

func TestCallbackHandlerMissingCode(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	ch := make(chan callbackResult, 1)
	handler := callbackHandler("expected-state", ch)

	req := httptest.NewRequest("GET", "/callback?state=expected-state", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	select {
	case result := <-ch:
		if result.err == nil {
			t.Fatal("expected error for missing code")
		}
		if !strings.Contains(result.err.Error(), "no authorization code") {
			t.Errorf("unexpected error: %v", result.err)
		}
	default:
		t.Fatal("no result received")
	}
}

func TestCallbackHandlerProviderError(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	ch := make(chan callbackResult, 1)
	handler := callbackHandler("expected-state", ch)

	req := httptest.NewRequest("GET", "/callback?error=access_denied&error_description=user+denied", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	select {
	case result := <-ch:
		if result.err == nil {
			t.Fatal("expected error for provider error")
		}
		if !strings.Contains(result.err.Error(), "access_denied") {
			t.Errorf("unexpected error: %v", result.err)
		}
	default:
		t.Fatal("no result received")
	}
}

func TestCallbackHandlerWrongPath(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	ch := make(chan callbackResult, 1)
	handler := callbackHandler("state", ch)

	req := httptest.NewRequest("GET", "/wrong", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("expected 404 for wrong path, got %d", w.Code)
	}

	select {
	case <-ch:
		t.Fatal("should not receive result for wrong path")
	default:
		// Expected.
	}
}

// --- Helper: routingHTTPClient ---

// routingHTTPClient routes requests to the test server regardless of URL.
type routingHTTPClient struct {
	server    *httptest.Server
	transport http.RoundTripper
}

func (c *routingHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Rewrite the request URL to point to our test server, preserving the path.
	serverURL, _ := url.Parse(c.server.URL)
	req.URL.Scheme = serverURL.Scheme
	req.URL.Host = serverURL.Host
	return c.server.Client().Do(req)
}

// mockServerURL extracts the server URL from the request for use in discovery docs.
func mockServerURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// findFreePort finds an available TCP port for tests.
func findFreePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()
	return port
}

// --- Token exchange test ---

func TestExchangeCodeSuccess(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			http.NotFound(w, r)
			return
		}

		_ = r.ParseForm()

		// Verify all required PKCE params.
		if r.FormValue("grant_type") != "authorization_code" {
			t.Errorf("wrong grant_type: %s", r.FormValue("grant_type"))
		}
		if r.FormValue("code_verifier") == "" {
			t.Error("missing code_verifier")
		}
		if r.FormValue("redirect_uri") == "" {
			t.Error("missing redirect_uri")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"access_token": "exchanged-access",
			"token_type": "Bearer",
			"expires_in": 3600
		}`))
	}))
	defer mockProvider.Close()

	client := NewOIDCClient("https://idp.example.com", "test-client", nil, 8765,
		WithHTTPClient(&routingHTTPClient{server: mockProvider, transport: http.DefaultTransport}),
	)

	tokens, err := client.exchangeCode(context.Background(), mockProvider.URL+"/token", "test-code", "test-verifier")
	if err != nil {
		t.Fatalf("exchangeCode failed: %v", err)
	}

	if tokens.AccessToken != "exchanged-access" {
		t.Errorf("unexpected access token: %s", tokens.AccessToken)
	}
}

func TestExchangeCodeFailure(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	mockProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"invalid_grant"}`, 400)
	}))
	defer mockProvider.Close()

	client := NewOIDCClient("https://idp.example.com", "test-client", nil, 8765,
		WithHTTPClient(&routingHTTPClient{server: mockProvider, transport: http.DefaultTransport}),
	)

	_, err := client.exchangeCode(context.Background(), mockProvider.URL+"/token", "bad-code", "verifier")
	if err == nil {
		t.Fatal("expected error for failed exchange")
	}
	if !strings.Contains(err.Error(), "status 400") {
		t.Errorf("unexpected error: %v", err)
	}
}

// --- GenerateRandomString Test ---

func TestGenerateRandomString(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	s1, err := generateRandomString(32)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s2, err := generateRandomString(32)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if s1 == s2 {
		t.Error("two random strings should not be equal")
	}

	if len(s1) == 0 {
		t.Error("random string should not be empty")
	}
}

// --- RedirectURI Test ---

func TestRedirectURI(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	client := NewOIDCClient("https://idp.example.com", "test", nil, 9999)
	got := client.redirectURI()
	want := "http://127.0.0.1:9999/callback"
	if got != want {
		t.Errorf("redirectURI() = %q, want %q", got, want)
	}
}

// --- TokenSet JSON roundtrip ---

func TestTokenSetJSONRoundTrip(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	original := TokenSet{
		AccessToken:  "at",
		RefreshToken: "rt",
		IDToken:      "idt",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		ExpiresAt:    1234567890,
		Scope:        "openid",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded TokenSet
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if decoded != original {
		t.Errorf("roundtrip mismatch:\ngot  %+v\nwant %+v", decoded, original)
	}
}

// --- Options Test ---

func TestNewOIDCClientDefaults(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	client := NewOIDCClient("https://idp.example.com", "client-id",
		[]string{"openid"}, 8765)

	if client.Issuer != "https://idp.example.com" {
		t.Errorf("unexpected issuer: %s", client.Issuer)
	}
	if client.ClientID != "client-id" {
		t.Errorf("unexpected client ID: %s", client.ClientID)
	}
	if len(client.Scopes) != 1 || client.Scopes[0] != "openid" {
		t.Errorf("unexpected scopes: %v", client.Scopes)
	}
	if client.CallbackPort != 8765 {
		t.Errorf("unexpected callback port: %d", client.CallbackPort)
	}
	if client.TokenStorePath == "" {
		t.Error("TokenStorePath should be set to default")
	}
}

func TestWithOptions(t *testing.T) {
	rtmx.Req(t, "REQ-GO-040")

	mock := newMockHTTPClient()
	opened := false
	opener := func(u string) error {
		opened = true
		return nil
	}

	client := NewOIDCClient("https://idp.example.com", "test", nil, 8765,
		WithHTTPClient(mock),
		WithBrowserOpener(opener),
		WithTokenStorePath("/tmp/test-tokens.json"),
	)

	if client.HTTPClient != mock {
		t.Error("HTTPClient not set by option")
	}
	_ = client.OpenBrowser("test")
	if !opened {
		t.Error("BrowserOpener not set by option")
	}
	if client.TokenStorePath != "/tmp/test-tokens.json" {
		t.Errorf("TokenStorePath not set by option: %s", client.TokenStorePath)
	}
}

// --- Fuzz test for token parsing ---

// Ensure TokenSet JSON parsing does not panic on arbitrary input.
func FuzzTokenSetParse(f *testing.F) {
	f.Add([]byte(`{"access_token":"x","expires_in":3600}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`invalid`))
	f.Add([]byte(`{"expires_at":-1}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var ts TokenSet
		_ = json.Unmarshal(data, &ts)
		// Must not panic. Result validity is not checked.
		_ = ts.IsExpired()
	})
}

// writeTestTokens is a helper that writes token data for testing.
func writeTestTokens(t *testing.T, path string, tokens *TokenSet) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	data, _ := json.Marshal(tokens)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("failed to write tokens: %v", err)
	}
}

