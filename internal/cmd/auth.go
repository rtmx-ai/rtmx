package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/rtmx-ai/rtmx/internal/auth"
	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication with OIDC providers",
	Long: `Manage authentication for RTMX sync and collaboration.

Managed Sync (default for 'rtmx login'):
  Zero-config browser OAuth against sync.rtmx.ai with localhost loopback.

Self-managed / air-gapped OIDC:
  Configure rtmx.yaml under rtmx.auth (issuer, client_id) and run
  'rtmx auth login' without --managed.

Subcommands:
  login   Start managed or OIDC login flow
  status  Show current authentication status
  logout  Clear stored tokens`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with managed Sync or a configured OIDC provider",
	Long: `Start a browser login flow.

Managed (recommended):
  rtmx auth login --managed
  rtmx login

Opens the system browser to the managed sync host. A local callback
server receives the session and stores it in ~/.rtmx/auth/tokens.json.

OIDC (self-managed):
  Configure auth.issuer and auth.client_id in rtmx.yaml, then:
  rtmx auth login`,
	RunE: runAuthLogin,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	Long: `Display whether you are currently authenticated, and if so,
whether your tokens are valid or expired.`,
	RunE: runAuthStatus,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear stored authentication tokens",
	Long:  `Remove the stored tokens file at ~/.rtmx/auth/tokens.json`,
	RunE:  runAuthLogout,
}

// Top-level alias: rtmx login → managed login (REQ-GO-082 / MONO-021a).
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Sign in to managed Sync (browser + localhost loopback)",
	Long: `Zero-config managed login against sync.rtmx.ai.

Opens the system browser for GitHub (or Google) OAuth. Completes via
a localhost loopback callback and stores the session credential.

Equivalent to: rtmx auth login --managed`,
	RunE: runManagedLoginCmd,
}

var (
	authManaged   bool
	authProvider  string
	authSyncURL   string
	authTokenPath string
)

// oidcClientFactory allows tests to inject a mock OIDC client.
// In production this is nil and buildOIDCClient is used.
var oidcClientFactory func(cfg *config.AuthConfig) *auth.OIDCClient

// managedLoginFunc allows tests to inject ManagedLogin.
var managedLoginFunc = auth.ManagedLogin

// getEnv abstracts environment lookups for tests.
var getEnv = os.Getenv

func init() {
	authLoginCmd.Flags().BoolVar(&authManaged, "managed", false,
		"Use managed Sync OAuth (sync.rtmx.ai) instead of configured OIDC issuer")
	authLoginCmd.Flags().StringVar(&authProvider, "provider", "github",
		"OAuth provider for managed login: github or google")
	authLoginCmd.Flags().StringVar(&authSyncURL, "sync-http-url", "",
		"Managed sync HTTP base URL (default https://sync.rtmx.ai or $RTMX_SYNC_HTTP_URL)")
	authLoginCmd.Flags().StringVar(&authTokenPath, "token-path", "",
		"Override token store path (default ~/.rtmx/auth/tokens.json)")

	loginCmd.Flags().StringVar(&authProvider, "provider", "github",
		"OAuth provider: github or google")
	loginCmd.Flags().StringVar(&authSyncURL, "sync-http-url", "",
		"Managed sync HTTP base URL (default https://sync.rtmx.ai or $RTMX_SYNC_HTTP_URL)")
	loginCmd.Flags().StringVar(&authTokenPath, "token-path", "",
		"Override token store path (default ~/.rtmx/auth/tokens.json)")

	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)

	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(loginCmd)
}

func buildOIDCClient(cfg *config.AuthConfig, opts ...auth.Option) (*auth.OIDCClient, error) {
	if cfg.Issuer == "" {
		return nil, fmt.Errorf("auth.issuer not configured in rtmx.yaml")
	}
	if cfg.ClientID == "" {
		return nil, fmt.Errorf("auth.client_id not configured in rtmx.yaml")
	}

	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = []string{"openid", "profile", "email"}
	}

	port := cfg.CallbackPort
	if port == 0 {
		port = 8765
	}

	allOpts := []auth.Option{
		auth.WithBrowserOpener(openSystemBrowser),
	}
	allOpts = append(allOpts, opts...)

	return auth.NewOIDCClient(cfg.Issuer, cfg.ClientID, scopes, port, allOpts...), nil
}

func resolveManagedSyncURL() string {
	if authSyncURL != "" {
		return strings.TrimRight(authSyncURL, "/")
	}
	if v := getEnv("RTMX_SYNC_HTTP_URL"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return auth.DefaultManagedSyncURL
}

func resolveTokenPath() string {
	if authTokenPath != "" {
		return authTokenPath
	}
	path, err := auth.DefaultTokenStorePath()
	if err != nil {
		return ""
	}
	return path
}

func runManagedLoginCmd(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	syncURL := resolveManagedSyncURL()
	provider := authProvider
	if provider == "" {
		provider = "github"
	}

	cmd.Printf("Signing in to managed Sync (%s) via %s...\n", syncURL, provider)
	cmd.Println("Opening browser for login...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	result, err := managedLoginFunc(ctx, auth.ManagedLoginConfig{
		SyncURL:    syncURL,
		Provider:   provider,
		TokenPath:  resolveTokenPath(),
		Browser:    openSystemBrowser,
		HTTPClient: nil, // default
	})
	if err != nil {
		return fmt.Errorf("managed login failed: %w", err)
	}

	cmd.Printf("%s Signed in", output.Color("OK", output.Green))
	if result.Tokens != nil {
		identity := result.Tokens.DisplayName
		if identity == "" {
			identity = result.Tokens.Email
		}
		if identity != "" {
			cmd.Printf(" as %s", identity)
		}
		if result.Tokens.Email != "" && result.Tokens.DisplayName != "" &&
			result.Tokens.Email != result.Tokens.DisplayName {
			cmd.Printf(" <%s>", result.Tokens.Email)
		}
	}
	cmd.Println(".")
	return nil
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	if authManaged {
		return runManagedLoginCmd(cmd, args)
	}

	if noColor {
		output.DisableColor()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	var client *auth.OIDCClient
	if oidcClientFactory != nil {
		client = oidcClientFactory(&cfg.RTMX.Auth)
	} else {
		client, err = buildOIDCClient(&cfg.RTMX.Auth)
		if err != nil {
			return err
		}
	}

	cmd.Printf("Authenticating with %s ...\n", cfg.RTMX.Auth.Issuer)
	cmd.Println("Opening browser for login...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	tokens, err := client.Login(ctx)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	cmd.Printf("%s Authentication successful.\n", output.Color("OK", output.Green))
	if tokens.ExpiresAt > 0 {
		expiresAt := time.Unix(tokens.ExpiresAt, 0)
		cmd.Printf("Token expires at: %s\n", expiresAt.Format(time.RFC3339))
	}

	return nil
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	tokens, err := loadAuthTokensForStatus()
	if err != nil || tokens == nil {
		cmd.Printf("Status: %s\n", output.Color("Not authenticated", output.Red))
		cmd.Println("Run 'rtmx login' to authenticate.")
		return nil
	}

	if tokens.IsExpired() {
		cmd.Printf("Status: %s\n", output.Color("Expired", output.Yellow))
		expiresAt := time.Unix(tokens.ExpiresAt, 0)
		cmd.Printf("Expired at: %s\n", expiresAt.Format(time.RFC3339))
		cmd.Println("Run 'rtmx login' to re-authenticate.")
		return nil
	}

	cmd.Printf("Status: %s\n", output.Color("Authenticated", output.Green))
	if tokens.Mode == "managed" {
		cmd.Printf("Mode: managed\n")
		if tokens.SyncURL != "" {
			cmd.Printf("Sync: %s\n", tokens.SyncURL)
		}
		if tokens.Provider != "" {
			cmd.Printf("Provider: %s\n", tokens.Provider)
		}
	}
	identity := tokens.DisplayName
	if identity == "" {
		identity = tokens.Email
	}
	if identity != "" {
		cmd.Printf("Identity: %s\n", identity)
	}
	if tokens.Email != "" && tokens.DisplayName != "" && tokens.Email != tokens.DisplayName {
		cmd.Printf("Email: %s\n", tokens.Email)
	}
	if tokens.ExpiresAt > 0 {
		expiresAt := time.Unix(tokens.ExpiresAt, 0)
		cmd.Printf("Expires at: %s\n", expiresAt.Format(time.RFC3339))
	}

	return nil
}

// loadAuthTokensForStatus prefers an injected OIDC client (tests), else the
// default token store — no issuer required (managed login path).
func loadAuthTokensForStatus() (*auth.TokenSet, error) {
	if oidcClientFactory != nil {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		cfg, err := config.LoadFromDir(cwd)
		if err != nil {
			return nil, err
		}
		client := oidcClientFactory(&cfg.RTMX.Auth)
		return client.LoadTokens()
	}
	return auth.LoadStoredTokens(resolveTokenPath())
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	if oidcClientFactory != nil {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		cfg, err := config.LoadFromDir(cwd)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		client := oidcClientFactory(&cfg.RTMX.Auth)
		if err := client.ClearTokens(); err != nil {
			return fmt.Errorf("failed to clear tokens: %w", err)
		}
	} else {
		if err := auth.ClearStoredTokens(resolveTokenPath()); err != nil {
			return fmt.Errorf("failed to clear tokens: %w", err)
		}
	}

	cmd.Println("Logged out. Tokens cleared.")
	return nil
}

// openSystemBrowser opens a URL in the default system browser.
func openSystemBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		return fmt.Errorf("unsupported platform %s; visit: %s", runtime.GOOS, url)
	}

	return exec.Command(cmd, args...).Start()
}
