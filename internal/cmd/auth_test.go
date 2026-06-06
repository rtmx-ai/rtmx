package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rtmx-ai/rtmx/internal/auth"
	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

// createAuthTestCmd creates an isolated root command with auth subcommands
// wired in, so tests do not share global flag state.
func createAuthTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	authParent := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication with OIDC providers",
	}

	loginCmd := &cobra.Command{
		Use:  "login",
		RunE: runAuthLogin,
	}

	statusCmd := &cobra.Command{
		Use:  "status",
		RunE: runAuthStatus,
	}

	logoutCmd := &cobra.Command{
		Use:  "logout",
		RunE: runAuthLogout,
	}

	authParent.AddCommand(loginCmd, statusCmd, logoutCmd)
	root.AddCommand(authParent)
	return root
}

// writeMinimalConfig writes an rtmx.yaml with the given auth section.
func writeMinimalConfig(t *testing.T, dir string, authCfg map[string]interface{}) {
	t.Helper()
	cfg := map[string]interface{}{
		"rtmx": map[string]interface{}{
			"database": "database.csv",
			"auth":     authCfg,
		},
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	// config.FindConfig looks for rtmx.yaml
	if err := os.WriteFile(filepath.Join(dir, "rtmx.yaml"), data, 0644); err != nil {
		t.Fatal(err)
	}
}

// setupAuthTestDir creates a temp directory with an rtmx.yaml and returns
// a cleanup function that restores the original working directory.
func setupAuthTestDir(t *testing.T, authCfg map[string]interface{}) (string, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "rtmx-auth-test-*")
	if err != nil {
		t.Fatal(err)
	}
	writeMinimalConfig(t, tmpDir, authCfg)

	oldWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}

	return tmpDir, func() {
		_ = os.Chdir(oldWd)
		_ = os.RemoveAll(tmpDir)
	}
}

// --- Tests ---

func TestAuthLoginMissingIssuer(t *testing.T) {
	rtmx.Req(t, "REQ-GO-078")
	output.DisableColor()
	defer output.EnableColor()

	_, cleanup := setupAuthTestDir(t, map[string]interface{}{
		"client_id": "some-client",
	})
	defer cleanup()

	// Ensure the production path is used (no factory override).
	old := oidcClientFactory
	oidcClientFactory = nil
	defer func() { oidcClientFactory = old }()

	cmd := createAuthTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"auth", "login"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when issuer is not configured")
	}
	if !strings.Contains(err.Error(), "auth.issuer not configured") {
		t.Errorf("expected issuer-not-configured error, got: %v", err)
	}
}

func TestAuthLoginMissingClientID(t *testing.T) {
	rtmx.Req(t, "REQ-GO-078")
	output.DisableColor()
	defer output.EnableColor()

	_, cleanup := setupAuthTestDir(t, map[string]interface{}{
		"issuer": "https://auth.example.com",
	})
	defer cleanup()

	old := oidcClientFactory
	oidcClientFactory = nil
	defer func() { oidcClientFactory = old }()

	cmd := createAuthTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"auth", "login"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when client_id is not configured")
	}
	if !strings.Contains(err.Error(), "auth.client_id not configured") {
		t.Errorf("expected client_id-not-configured error, got: %v", err)
	}
}

func TestAuthLoginMissingBothIssuerAndClientID(t *testing.T) {
	rtmx.Req(t, "REQ-GO-078")
	output.DisableColor()
	defer output.EnableColor()

	_, cleanup := setupAuthTestDir(t, map[string]interface{}{})
	defer cleanup()

	old := oidcClientFactory
	oidcClientFactory = nil
	defer func() { oidcClientFactory = old }()

	cmd := createAuthTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"auth", "login"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when both issuer and client_id are missing")
	}
	// First validation check is issuer
	if !strings.Contains(err.Error(), "auth.issuer not configured") {
		t.Errorf("expected issuer-not-configured error, got: %v", err)
	}
}

func TestAuthStatusNoToken(t *testing.T) {
	rtmx.Req(t, "REQ-GO-078")
	output.DisableColor()
	defer output.EnableColor()

	tmpDir, cleanup := setupAuthTestDir(t, map[string]interface{}{
		"issuer":    "https://auth.example.com",
		"client_id": "test-client",
	})
	defer cleanup()

	// Use factory to inject a client with a token store in the temp dir.
	tokenPath := filepath.Join(tmpDir, "tokens.json")
	old := oidcClientFactory
	oidcClientFactory = func(cfg *config.AuthConfig) *auth.OIDCClient {
		return auth.NewOIDCClient(
			cfg.Issuer, cfg.ClientID,
			[]string{"openid"}, 8765,
			auth.WithTokenStorePath(tokenPath),
		)
	}
	defer func() { oidcClientFactory = old }()

	cmd := createAuthTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"auth", "status"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("auth status should not error when no token exists, got: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Not authenticated") {
		t.Errorf("expected 'Not authenticated' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "rtmx auth login") {
		t.Errorf("expected login hint in output, got:\n%s", out)
	}
}

func TestAuthStatusWithValidToken(t *testing.T) {
	rtmx.Req(t, "REQ-GO-078")
	output.DisableColor()
	defer output.EnableColor()

	tmpDir, cleanup := setupAuthTestDir(t, map[string]interface{}{
		"issuer":    "https://auth.example.com",
		"client_id": "test-client",
	})
	defer cleanup()

	// Write a valid (non-expired) token file.
	tokenPath := filepath.Join(tmpDir, "tokens.json")
	futureExpiry := time.Now().Add(1 * time.Hour).Unix()
	tokens := auth.TokenSet{
		AccessToken: "test-access-token",
		TokenType:   "Bearer",
		ExpiresAt:   futureExpiry,
	}
	data, _ := json.Marshal(&tokens)
	if err := os.MkdirAll(filepath.Dir(tokenPath), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		t.Fatal(err)
	}

	old := oidcClientFactory
	oidcClientFactory = func(cfg *config.AuthConfig) *auth.OIDCClient {
		return auth.NewOIDCClient(
			cfg.Issuer, cfg.ClientID,
			[]string{"openid"}, 8765,
			auth.WithTokenStorePath(tokenPath),
		)
	}
	defer func() { oidcClientFactory = old }()

	cmd := createAuthTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"auth", "status"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("auth status with valid token should not error, got: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Authenticated") {
		t.Errorf("expected 'Authenticated' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Expires at:") {
		t.Errorf("expected 'Expires at:' in output, got:\n%s", out)
	}
}

func TestAuthStatusWithExpiredToken(t *testing.T) {
	rtmx.Req(t, "REQ-GO-078")
	output.DisableColor()
	defer output.EnableColor()

	tmpDir, cleanup := setupAuthTestDir(t, map[string]interface{}{
		"issuer":    "https://auth.example.com",
		"client_id": "test-client",
	})
	defer cleanup()

	tokenPath := filepath.Join(tmpDir, "tokens.json")
	pastExpiry := time.Now().Add(-1 * time.Hour).Unix()
	tokens := auth.TokenSet{
		AccessToken: "expired-token",
		TokenType:   "Bearer",
		ExpiresAt:   pastExpiry,
	}
	data, _ := json.Marshal(&tokens)
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		t.Fatal(err)
	}

	old := oidcClientFactory
	oidcClientFactory = func(cfg *config.AuthConfig) *auth.OIDCClient {
		return auth.NewOIDCClient(
			cfg.Issuer, cfg.ClientID,
			[]string{"openid"}, 8765,
			auth.WithTokenStorePath(tokenPath),
		)
	}
	defer func() { oidcClientFactory = old }()

	cmd := createAuthTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"auth", "status"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("auth status with expired token should not error, got: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Expired") {
		t.Errorf("expected 'Expired' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "rtmx auth login") {
		t.Errorf("expected re-auth hint in output, got:\n%s", out)
	}
}

func TestAuthLogoutClearsToken(t *testing.T) {
	rtmx.Req(t, "REQ-GO-078")
	output.DisableColor()
	defer output.EnableColor()

	tmpDir, cleanup := setupAuthTestDir(t, map[string]interface{}{
		"issuer":    "https://auth.example.com",
		"client_id": "test-client",
	})
	defer cleanup()

	// Write a token file, then verify logout removes it.
	tokenPath := filepath.Join(tmpDir, "tokens.json")
	tokens := auth.TokenSet{
		AccessToken: "to-be-cleared",
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(1 * time.Hour).Unix(),
	}
	data, _ := json.Marshal(&tokens)
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		t.Fatal(err)
	}

	old := oidcClientFactory
	oidcClientFactory = func(cfg *config.AuthConfig) *auth.OIDCClient {
		return auth.NewOIDCClient(
			cfg.Issuer, cfg.ClientID,
			[]string{"openid"}, 8765,
			auth.WithTokenStorePath(tokenPath),
		)
	}
	defer func() { oidcClientFactory = old }()

	cmd := createAuthTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"auth", "logout"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("auth logout failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Logged out") {
		t.Errorf("expected 'Logged out' in output, got:\n%s", out)
	}
	if !strings.Contains(out, "Tokens cleared") {
		t.Errorf("expected 'Tokens cleared' in output, got:\n%s", out)
	}

	// Verify token file was removed.
	if _, err := os.Stat(tokenPath); !os.IsNotExist(err) {
		t.Errorf("expected token file to be removed after logout")
	}
}

func TestAuthLogoutNoExistingToken(t *testing.T) {
	rtmx.Req(t, "REQ-GO-078")
	output.DisableColor()
	defer output.EnableColor()

	tmpDir, cleanup := setupAuthTestDir(t, map[string]interface{}{
		"issuer":    "https://auth.example.com",
		"client_id": "test-client",
	})
	defer cleanup()

	tokenPath := filepath.Join(tmpDir, "tokens.json")

	old := oidcClientFactory
	oidcClientFactory = func(cfg *config.AuthConfig) *auth.OIDCClient {
		return auth.NewOIDCClient(
			cfg.Issuer, cfg.ClientID,
			[]string{"openid"}, 8765,
			auth.WithTokenStorePath(tokenPath),
		)
	}
	defer func() { oidcClientFactory = old }()

	cmd := createAuthTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"auth", "logout"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("auth logout with no existing token should not error, got: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Logged out") {
		t.Errorf("expected 'Logged out' in output, got:\n%s", out)
	}
}

func TestAuthLoginOIDCClientFactoryInjection(t *testing.T) {
	rtmx.Req(t, "REQ-GO-078")
	output.DisableColor()
	defer output.EnableColor()

	tmpDir, cleanup := setupAuthTestDir(t, map[string]interface{}{
		"issuer":    "https://auth.example.com",
		"client_id": "test-client",
	})
	defer cleanup()

	tokenPath := filepath.Join(tmpDir, "tokens.json")
	var factoryCalled bool
	var receivedIssuer string
	var receivedClientID string

	old := oidcClientFactory
	oidcClientFactory = func(cfg *config.AuthConfig) *auth.OIDCClient {
		factoryCalled = true
		receivedIssuer = cfg.Issuer
		receivedClientID = cfg.ClientID
		return auth.NewOIDCClient(
			cfg.Issuer, cfg.ClientID,
			[]string{"openid"}, 0, // port 0 = let OS pick
			auth.WithTokenStorePath(tokenPath),
			auth.WithBrowserOpener(func(url string) error {
				// Simulate browser open by directly hitting the callback
				// with a known code -- but we cannot complete the flow without
				// a real token endpoint. Instead, just return an error to
				// short-circuit the login. The point of this test is to
				// verify the factory injection seam works.
				return context.DeadlineExceeded
			}),
		)
	}
	defer func() { oidcClientFactory = old }()

	cmd := createAuthTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"auth", "login"})

	// The login will fail because the mock browser opener errors out
	// or the OIDC discovery will fail. Either way, we verify the factory
	// was called with the right config.
	_ = cmd.Execute()

	if !factoryCalled {
		t.Fatal("oidcClientFactory was not called")
	}
	if receivedIssuer != "https://auth.example.com" {
		t.Errorf("expected issuer 'https://auth.example.com', got %q", receivedIssuer)
	}
	if receivedClientID != "test-client" {
		t.Errorf("expected clientID 'test-client', got %q", receivedClientID)
	}
}

func TestAuthStatusMissingConfig(t *testing.T) {
	rtmx.Req(t, "REQ-GO-078")
	output.DisableColor()
	defer output.EnableColor()

	// Use a dir with no auth config (empty issuer/client_id).
	_, cleanup := setupAuthTestDir(t, map[string]interface{}{})
	defer cleanup()

	old := oidcClientFactory
	oidcClientFactory = nil
	defer func() { oidcClientFactory = old }()

	cmd := createAuthTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"auth", "status"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when issuer is not configured for status")
	}
	if !strings.Contains(err.Error(), "auth.issuer not configured") {
		t.Errorf("expected issuer-not-configured error, got: %v", err)
	}
}

func TestAuthLogoutMissingConfig(t *testing.T) {
	rtmx.Req(t, "REQ-GO-078")
	output.DisableColor()
	defer output.EnableColor()

	_, cleanup := setupAuthTestDir(t, map[string]interface{}{})
	defer cleanup()

	old := oidcClientFactory
	oidcClientFactory = nil
	defer func() { oidcClientFactory = old }()

	cmd := createAuthTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"auth", "logout"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error when issuer is not configured for logout")
	}
	if !strings.Contains(err.Error(), "auth.issuer not configured") {
		t.Errorf("expected issuer-not-configured error, got: %v", err)
	}
}

func TestAuthStatusTokenWithoutExpiry(t *testing.T) {
	rtmx.Req(t, "REQ-GO-078")
	output.DisableColor()
	defer output.EnableColor()

	tmpDir, cleanup := setupAuthTestDir(t, map[string]interface{}{
		"issuer":    "https://auth.example.com",
		"client_id": "test-client",
	})
	defer cleanup()

	// Token with no ExpiresAt (never expires).
	tokenPath := filepath.Join(tmpDir, "tokens.json")
	tokens := auth.TokenSet{
		AccessToken: "no-expiry-token",
		TokenType:   "Bearer",
		ExpiresAt:   0,
	}
	data, _ := json.Marshal(&tokens)
	if err := os.WriteFile(tokenPath, data, 0600); err != nil {
		t.Fatal(err)
	}

	old := oidcClientFactory
	oidcClientFactory = func(cfg *config.AuthConfig) *auth.OIDCClient {
		return auth.NewOIDCClient(
			cfg.Issuer, cfg.ClientID,
			[]string{"openid"}, 8765,
			auth.WithTokenStorePath(tokenPath),
		)
	}
	defer func() { oidcClientFactory = old }()

	cmd := createAuthTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"auth", "status"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("auth status with no-expiry token should not error, got: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Authenticated") {
		t.Errorf("expected 'Authenticated' in output, got:\n%s", out)
	}
	// Should NOT show "Expires at:" when ExpiresAt is 0
	if strings.Contains(out, "Expires at:") {
		t.Errorf("should not show expiry when ExpiresAt is 0, got:\n%s", out)
	}
}

func TestBuildOIDCClientDefaults(t *testing.T) {
	rtmx.Req(t, "REQ-GO-078")

	cfg := &config.AuthConfig{
		Issuer:   "https://auth.example.com",
		ClientID: "my-client",
	}

	client, err := buildOIDCClient(cfg)
	if err != nil {
		t.Fatalf("buildOIDCClient failed: %v", err)
	}

	if client.Issuer != "https://auth.example.com" {
		t.Errorf("expected issuer 'https://auth.example.com', got %q", client.Issuer)
	}
	if client.ClientID != "my-client" {
		t.Errorf("expected clientID 'my-client', got %q", client.ClientID)
	}
	if client.CallbackPort != 8765 {
		t.Errorf("expected default callback port 8765, got %d", client.CallbackPort)
	}
	if len(client.Scopes) != 3 {
		t.Errorf("expected 3 default scopes, got %d", len(client.Scopes))
	}
}

func TestBuildOIDCClientCustomPort(t *testing.T) {
	rtmx.Req(t, "REQ-GO-078")

	cfg := &config.AuthConfig{
		Issuer:       "https://auth.example.com",
		ClientID:     "my-client",
		CallbackPort: 9999,
		Scopes:       []string{"openid", "custom"},
	}

	client, err := buildOIDCClient(cfg)
	if err != nil {
		t.Fatalf("buildOIDCClient failed: %v", err)
	}

	if client.CallbackPort != 9999 {
		t.Errorf("expected callback port 9999, got %d", client.CallbackPort)
	}
	if len(client.Scopes) != 2 {
		t.Errorf("expected 2 scopes, got %d", len(client.Scopes))
	}
}

func TestBuildOIDCClientMissingIssuer(t *testing.T) {
	rtmx.Req(t, "REQ-GO-078")

	cfg := &config.AuthConfig{
		ClientID: "my-client",
	}

	_, err := buildOIDCClient(cfg)
	if err == nil {
		t.Fatal("expected error for missing issuer")
	}
	if !strings.Contains(err.Error(), "auth.issuer not configured") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestBuildOIDCClientMissingClientID(t *testing.T) {
	rtmx.Req(t, "REQ-GO-078")

	cfg := &config.AuthConfig{
		Issuer: "https://auth.example.com",
	}

	_, err := buildOIDCClient(cfg)
	if err == nil {
		t.Fatal("expected error for missing client_id")
	}
	if !strings.Contains(err.Error(), "auth.client_id not configured") {
		t.Errorf("unexpected error: %v", err)
	}
}
