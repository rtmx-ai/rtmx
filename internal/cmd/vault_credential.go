package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rtmx-ai/rtmx/internal/auth"
)

// managedCredentialGrant is the POST /connections/credential response
// (connections-v1 credential_grant). ViaSharedOrgID is set when Sync issued
// the token for an org-shared connection (REQ-MONO-049c).
type managedCredentialGrant struct {
	ConnectionID   string `json:"connection_id"`
	Provider       string `json:"provider"`
	AccessToken    string `json:"access_token"`
	ViaSharedOrgID string `json:"via_shared_org_id,omitempty"`
}

// silentManagedRestore renews a managed session without interactive login
// when possible: refresh_token first (REQ-MONO-057c), then prompt=none for
// Google/Microsoft (REQ-MONO-053c).
var silentManagedRestore = func(tok *auth.TokenSet) error {
	if tok == nil || tok.Mode != "managed" {
		return auth.ErrSilentUnsupported
	}
	syncURL := tok.SyncURL
	if syncURL == "" {
		syncURL = managedSyncHTTPBase()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	if strings.TrimSpace(tok.RefreshToken) != "" {
		if _, err := auth.RefreshManagedSession(ctx, syncURL, tok.RefreshToken, ""); err == nil {
			return nil
		}
		// Fall through to silent OAuth when refresh is rejected/expired.
	}
	provider := strings.ToLower(strings.TrimSpace(tok.Provider))
	if provider != "google" && provider != "microsoft" {
		return auth.ErrSilentUnsupported
	}
	_, err := managedLoginFunc(ctx, auth.ManagedLoginConfig{
		SyncURL:  syncURL,
		Provider: provider,
		Silent:   true,
		Browser:  openSystemBrowser,
	})
	return err
}

// fetchManagedConnectionCredential requests a vault access token from managed
// Sync (POST /connections/credential). Works for the connection owner and for
// org owner/admin of a shared connection (REQ-MONO-046b / REQ-MONO-049b/c).
func fetchManagedConnectionCredential(provider string) (string, error) {
	return fetchManagedConnectionCredentialOnce(provider, true)
}

func fetchManagedConnectionCredentialOnce(provider string, allowSilentRetry bool) (string, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return "", fmt.Errorf("provider is required")
	}
	tok, err := auth.LoadStoredTokens("")
	if err != nil || tok == nil || strings.TrimSpace(tok.AccessToken) == "" {
		return "", fmt.Errorf("not signed in; run: rtmx login")
	}
	base := managedSyncHTTPBase()
	body := fmt.Sprintf(`{"provider":%q}`, provider)
	req, err := http.NewRequest(http.MethodPost, base+"/connections/credential", strings.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not reach %s: %w", base, err)
	}
	raw, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		if allowSilentRetry {
			if restoreErr := silentManagedRestore(tok); restoreErr == nil {
				return fetchManagedConnectionCredentialOnce(provider, false)
			} else if !errors.Is(restoreErr, auth.ErrSilentUnsupported) {
				return "", fmt.Errorf("session expired; silent restore failed: %w (try: rtmx login --silent --provider google)", restoreErr)
			}
		}
		return "", fmt.Errorf("session expired; run: rtmx login (or rtmx login --silent --provider google)")
	}
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("no vaulted %s connection; run: rtmx connect %s (or ask an org admin to share one)", provider, provider)
	}
	if resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("not allowed to fetch vaulted %s credential (org owner/admin required for shared connections)", provider)
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("could not fetch %s vault credential (%d): %s", provider, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var grant managedCredentialGrant
	if err := json.Unmarshal(raw, &grant); err != nil {
		return "", fmt.Errorf("invalid credential response: %w", err)
	}
	if strings.TrimSpace(grant.AccessToken) == "" {
		return "", fmt.Errorf("empty access_token in credential response")
	}
	return grant.AccessToken, nil
}

// githubEnvWithVault returns an env getter that falls back to a one-shot vault
// fetch when GITHUB_TOKEN and RTMX_GITHUB_CONNECTION_TOKEN are unset.
// The vault path uses the same Sync credential endpoint for owner and
// shared-org admin grants (REQ-MONO-049c).
func githubEnvWithVault(base func(string) string) func(string) string {
	if base == nil {
		base = os.Getenv
	}
	var cached string
	var fetched bool
	return func(key string) string {
		if v := base(key); v != "" {
			return v
		}
		if key != "GITHUB_TOKEN" && key != "RTMX_GITHUB_CONNECTION_TOKEN" {
			return ""
		}
		if !fetched {
			fetched = true
			tok, err := fetchManagedConnectionCredential("github")
			if err == nil {
				cached = tok
			}
		}
		if key == "RTMX_GITHUB_CONNECTION_TOKEN" || key == "GITHUB_TOKEN" {
			return cached
		}
		return ""
	}
}
