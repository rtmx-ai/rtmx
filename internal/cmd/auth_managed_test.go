package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/auth"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

func createManagedLoginTestCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	login := &cobra.Command{
		Use:  "login",
		RunE: runManagedLoginCmd,
	}
	login.Flags().StringVar(&authProvider, "provider", "github", "")
	login.Flags().StringVar(&authSyncURL, "sync-http-url", "", "")
	login.Flags().StringVar(&authTokenPath, "token-path", "", "")

	authParent := &cobra.Command{Use: "auth"}
	authLogin := &cobra.Command{
		Use:  "login",
		RunE: runAuthLogin,
	}
	authLogin.Flags().BoolVar(&authManaged, "managed", false, "")
	authLogin.Flags().StringVar(&authProvider, "provider", "github", "")
	authLogin.Flags().StringVar(&authSyncURL, "sync-http-url", "", "")
	authLogin.Flags().StringVar(&authTokenPath, "token-path", "", "")

	statusCmd := &cobra.Command{Use: "status", RunE: runAuthStatus}
	logoutCmd := &cobra.Command{Use: "logout", RunE: runAuthLogout}

	authParent.AddCommand(authLogin, statusCmd, logoutCmd)
	root.AddCommand(login, authParent)
	return root
}

func TestManagedLoginCommandStoresSession(t *testing.T) {
	rtmx.Req(t, "REQ-GO-082")
	output.DisableColor()
	defer output.EnableColor()

	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "tokens.json")

	oldFn := managedLoginFunc
	managedLoginFunc = func(_ context.Context, cfg auth.ManagedLoginConfig) (*auth.ManagedLoginResult, error) {
		if cfg.TokenPath != tokenPath {
			t.Errorf("token path = %q", cfg.TokenPath)
		}
		tokens := &auth.TokenSet{
			AccessToken: "cli-sess",
			Mode:        "managed",
			Email:       "a@rtmx.ai",
			DisplayName: "Ada",
			SyncURL:     cfg.SyncURL,
			Provider:    cfg.Provider,
		}
		if err := os.WriteFile(tokenPath, mustJSON(tokens), 0600); err != nil {
			return nil, err
		}
		return &auth.ManagedLoginResult{Tokens: tokens}, nil
	}
	defer func() { managedLoginFunc = oldFn }()

	cmd := createManagedLoginTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{
		"login",
		"--token-path", tokenPath,
		"--sync-http-url", "https://sync.example.test",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("login: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Signed in") {
		t.Fatalf("output: %s", out)
	}
	if !strings.Contains(out, "Ada") {
		t.Fatalf("expected identity in output: %s", out)
	}

	loaded, err := auth.LoadStoredTokens(tokenPath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.AccessToken != "cli-sess" {
		t.Fatalf("token = %q", loaded.AccessToken)
	}
	if loaded.Mode != "managed" {
		t.Fatalf("mode = %q", loaded.Mode)
	}
}

func TestAuthLoginManagedFlag(t *testing.T) {
	rtmx.Req(t, "REQ-GO-082")
	output.DisableColor()
	defer output.EnableColor()

	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "tokens.json")

	old := managedLoginFunc
	managedLoginFunc = func(_ context.Context, cfg auth.ManagedLoginConfig) (*auth.ManagedLoginResult, error) {
		if cfg.Provider != "github" {
			t.Errorf("provider = %q", cfg.Provider)
		}
		tokens := &auth.TokenSet{
			AccessToken: "x",
			Mode:        "managed",
			Email:       "u@rtmx.ai",
			DisplayName: "User",
			SyncURL:     cfg.SyncURL,
		}
		if err := os.WriteFile(tokenPath, mustJSON(tokens), 0600); err != nil {
			return nil, err
		}
		return &auth.ManagedLoginResult{Tokens: tokens}, nil
	}
	defer func() { managedLoginFunc = old }()

	cmd := createManagedLoginTestCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{
		"auth", "login", "--managed",
		"--token-path", tokenPath,
		"--sync-http-url", "https://sync.example.test",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth login --managed: %v", err)
	}
	if !strings.Contains(buf.String(), "Signed in") {
		t.Fatalf("output: %s", buf.String())
	}
}

func TestAuthStatusManagedIdentity(t *testing.T) {
	rtmx.Req(t, "REQ-GO-082")
	output.DisableColor()
	defer output.EnableColor()

	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "tokens.json")
	tokens := auth.TokenSet{
		AccessToken: "sess",
		TokenType:   "Bearer",
		Mode:        "managed",
		Email:       "dev@rtmx.ai",
		DisplayName: "Dev",
		SyncURL:     "https://sync.rtmx.ai",
		Provider:    "github",
	}
	if err := os.WriteFile(tokenPath, mustJSON(&tokens), 0600); err != nil {
		t.Fatal(err)
	}

	oldFactory := oidcClientFactory
	oidcClientFactory = nil
	defer func() { oidcClientFactory = oldFactory }()

	cmd := createManagedLoginTestCmd()
	// Flag registration resets package vars to defaults; set after create.
	authTokenPath = tokenPath

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"auth", "status"})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Authenticated") {
		t.Fatalf("output: %s", out)
	}
	if !strings.Contains(out, "Dev") {
		t.Fatalf("expected identity: %s", out)
	}
	if !strings.Contains(out, "managed") {
		t.Fatalf("expected mode: %s", out)
	}
}

func TestAuthLogoutManaged(t *testing.T) {
	rtmx.Req(t, "REQ-GO-082")
	output.DisableColor()
	defer output.EnableColor()

	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "tokens.json")
	if err := os.WriteFile(tokenPath, mustJSON(&auth.TokenSet{AccessToken: "x", Mode: "managed"}), 0600); err != nil {
		t.Fatal(err)
	}

	oldFactory := oidcClientFactory
	oidcClientFactory = nil
	defer func() { oidcClientFactory = oldFactory }()

	cmd := createManagedLoginTestCmd()
	authTokenPath = tokenPath

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"auth", "logout"})

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(tokenPath); !os.IsNotExist(err) {
		t.Fatalf("token file should be removed, err=%v", err)
	}
}

func TestAuthStatusWithoutIssuerShowsNotAuthenticated(t *testing.T) {
	rtmx.Req(t, "REQ-GO-082")
	output.DisableColor()
	defer output.EnableColor()

	dir := t.TempDir()
	tokenPath := filepath.Join(dir, "missing-tokens.json")

	oldFactory := oidcClientFactory
	oidcClientFactory = nil
	defer func() { oidcClientFactory = oldFactory }()

	cmd := createManagedLoginTestCmd()
	authTokenPath = tokenPath

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"auth", "status"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("status without issuer should succeed: %v", err)
	}
	if !strings.Contains(buf.String(), "Not authenticated") {
		t.Fatalf("output: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "rtmx login") {
		t.Fatalf("expected login hint: %s", buf.String())
	}
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
