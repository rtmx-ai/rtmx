package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/auth"
	"github.com/spf13/cobra"
)

// REQ-MONO-026 / REQ-MONO-038b / REQ-MONO-038c: SaaS connection OAuth.

var connectBrowser = openSystemBrowser

var connectCmd = &cobra.Command{
	Use:   "connect [provider]",
	Short: "Connect a project-management SaaS via OAuth",
	Long: `Authenticate RTMX against a project-management SaaS (Asana, Jira, Monday, GitLab, GitHub Issues).

Opens a browser OAuth flow mediated by the managed sync host.
List connections with: rtmx connections

Forge org affiliation (Prove with GitHub) is separate — use Organizations on rtmx.ai.

Examples:
  rtmx connect asana
  rtmx connect jira
  rtmx connect monday
  rtmx connect gitlab
  rtmx connect github
`,
	Args: cobra.ExactArgs(1),
	RunE: runConnect,
}

var connectionsCmd = &cobra.Command{
	Use:   "connections",
	Short: "List SaaS connections from the managed vault",
	Long:  `List project-management SaaS connections visible to your managed sync session.`,
	RunE:  runConnectionsList,
}

func init() {
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(connectionsCmd)
}

func managedSyncHTTPBase() string {
	if v := strings.TrimSpace(os.Getenv("RTMX_SYNC_HTTP_URL")); v != "" {
		return strings.TrimRight(v, "/")
	}
	return auth.DefaultManagedSyncURL
}

func runConnect(cmd *cobra.Command, args []string) error {
	provider := strings.ToLower(strings.TrimSpace(args[0]))
	switch provider {
	case "asana", "jira", "monday", "gitlab", "github":
	default:
		return fmt.Errorf("unsupported provider %q (supported: asana, jira, monday, gitlab, github)", provider)
	}

	tok, err := auth.LoadStoredTokens("")
	if err != nil || tok == nil || strings.TrimSpace(tok.AccessToken) == "" {
		return fmt.Errorf("not signed in; run: rtmx login")
	}

	base := managedSyncHTTPBase()
	// Browser completes on the marketing site; sync stores tokens via /callback.
	redirectURI := "https://rtmx.ai/account"
	if v := strings.TrimSpace(os.Getenv("RTMX_CONNECT_REDIRECT_URI")); v != "" {
		redirectURI = v
	}
	body := map[string]string{
		"redirect_uri": redirectURI,
		"cli_redirect": "http://127.0.0.1:8765/callback",
	}
	payload, _ := json.Marshal(body)
	startPath := fmt.Sprintf("/connections/%s/start", provider)
	req, err := http.NewRequest(http.MethodPost, base+startPath, strings.NewReader(string(payload)))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not reach %s: %w", base, err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusServiceUnavailable {
		envHint := "RTMX_ASANA_CLIENT_ID"
		switch provider {
		case "jira":
			envHint = "RTMX_JIRA_CLIENT_ID"
		case "monday":
			envHint = "RTMX_MONDAY_CLIENT_ID"
		case "gitlab":
			envHint = "RTMX_GITLAB_CLIENT_ID"
		case "github":
			envHint = "RTMX_GITHUB_ISSUES_CLIENT_ID"
		}
		return fmt.Errorf("%s OAuth is not configured on the sync host (set %s)", provider, envHint)
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("could not start %s OAuth (%d): %s", provider, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var started struct {
		AuthorizationURL string `json:"authorization_url"`
	}
	if err := json.Unmarshal(raw, &started); err != nil {
		return err
	}
	if started.AuthorizationURL == "" {
		return fmt.Errorf("sync host returned no authorization_url")
	}

	cmd.Printf("Open this URL to authorize %s:\n%s\n", provider, started.AuthorizationURL)
	cmd.Println("After approval, list connections with: rtmx connections")
	_ = connectBrowser(started.AuthorizationURL)
	return nil
}

func runConnectionsList(cmd *cobra.Command, args []string) error {
	tok, err := auth.LoadStoredTokens("")
	if err != nil || tok == nil || strings.TrimSpace(tok.AccessToken) == "" {
		return fmt.Errorf("not signed in; run: rtmx login")
	}
	base := managedSyncHTTPBase()
	req, err := http.NewRequest(http.MethodGet, base+"/connections", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+tok.AccessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("could not list connections (%d): %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	var body struct {
		Connections []struct {
			ID           string `json:"id"`
			Provider     string `json:"provider"`
			Status       string `json:"status"`
			DisplayLabel string `json:"display_label"`
		} `json:"connections"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return err
	}
	if len(body.Connections) == 0 {
		cmd.Println("No connections.")
		return nil
	}
	for _, c := range body.Connections {
		label := c.DisplayLabel
		if label == "" {
			label = c.Provider
		}
		cmd.Printf("%s\t%s\t%s\t%s\n", c.ID, c.Provider, c.Status, label)
	}
	return nil
}
