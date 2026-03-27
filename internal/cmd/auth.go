package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
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

Supports any OIDC-compliant identity provider (Zitadel, Okta, Auth0, etc.)
configured in rtmx.yaml under rtmx.auth.

Subcommands:
  login   Start OIDC PKCE login flow
  status  Show current authentication status
  logout  Clear stored tokens`,
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the configured OIDC provider",
	Long: `Start the OIDC Authorization Code flow with PKCE.

Opens the system browser to the identity provider's login page.
A local callback server receives the authorization code and
exchanges it for access/refresh tokens.

Tokens are stored in ~/.rtmx/auth/tokens.json`,
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

// oidcClientFactory allows tests to inject a mock OIDC client.
// In production this is nil and buildOIDCClient is used.
var oidcClientFactory func(cfg *config.AuthConfig) *auth.OIDCClient

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)

	rootCmd.AddCommand(authCmd)
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

func runAuthLogin(cmd *cobra.Command, args []string) error {
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

	tokens, err := client.LoadTokens()
	if err != nil {
		cmd.Printf("Status: %s\n", output.Color("Not authenticated", output.Red))
		cmd.Println("Run 'rtmx auth login' to authenticate.")
		return nil
	}

	if tokens.IsExpired() {
		cmd.Printf("Status: %s\n", output.Color("Expired", output.Yellow))
		expiresAt := time.Unix(tokens.ExpiresAt, 0)
		cmd.Printf("Expired at: %s\n", expiresAt.Format(time.RFC3339))
		cmd.Println("Run 'rtmx auth login' to re-authenticate.")
		return nil
	}

	cmd.Printf("Status: %s\n", output.Color("Authenticated", output.Green))
	if tokens.ExpiresAt > 0 {
		expiresAt := time.Unix(tokens.ExpiresAt, 0)
		cmd.Printf("Expires at: %s\n", expiresAt.Format(time.RFC3339))
	}

	return nil
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
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

	if err := client.ClearTokens(); err != nil {
		return fmt.Errorf("failed to clear tokens: %w", err)
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
