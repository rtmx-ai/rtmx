// Package adapters provides integrations with external services like GitHub and Jira.
package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// GitHubConfig holds GitHub adapter configuration
type GitHubConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Repo     string `yaml:"repo"`
	TokenEnv string `yaml:"token_env"`
	Labels   struct {
		Requirement string `yaml:"requirement"`
	} `yaml:"labels"`
}

// GitHubAdapter syncs requirements with GitHub Issues
type GitHubAdapter struct {
	config GitHubConfig
	client *http.Client
	token  string
}

// ExternalItem represents an item from an external service
type ExternalItem struct {
	ExternalID    string
	Title         string
	Description   string
	Status        string
	Labels        []string
	URL           string
	CreatedAt     *time.Time
	UpdatedAt     *time.Time
	Assignee      string
	Priority      string
	RequirementID string
}

// GitHubIssue represents a GitHub issue from the API
type GitHubIssue struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	State     string    `json:"state"`
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Labels    []struct {
		Name string `json:"name"`
	} `json:"labels"`
	Assignee *struct {
		Login string `json:"login"`
	} `json:"assignee"`
}

// NewGitHubAdapter creates a new GitHub adapter
func NewGitHubAdapter(config GitHubConfig) (*GitHubAdapter, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("GitHub adapter is not enabled")
	}

	tokenEnv := config.TokenEnv
	if tokenEnv == "" {
		tokenEnv = "GITHUB_TOKEN"
	}

	token := os.Getenv(tokenEnv)
	if token == "" {
		return nil, fmt.Errorf("GitHub token not found. Set %s environment variable", tokenEnv)
	}

	return &GitHubAdapter{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
		token:  token,
	}, nil
}

// Name returns the adapter name
func (g *GitHubAdapter) Name() string {
	return "github"
}

// IsConfigured checks if the adapter is properly configured
func (g *GitHubAdapter) IsConfigured() bool {
	return g.config.Enabled && g.config.Repo != "" && g.token != ""
}

// TestConnection tests the connection to GitHub
func (g *GitHubAdapter) TestConnection(ctx context.Context) (bool, string) {
	url := fmt.Sprintf("https://api.github.com/repos/%s", g.config.Repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Sprintf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "token "+g.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, fmt.Sprintf("Connection failed: HTTP %d", resp.StatusCode)
	}

	var repo struct {
		FullName string `json:"full_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return false, fmt.Sprintf("Failed to parse response: %v", err)
	}

	return true, fmt.Sprintf("Connected to %s", repo.FullName)
}

// FetchIssues fetches issues from GitHub
func (g *GitHubAdapter) FetchIssues(ctx context.Context, state string) ([]ExternalItem, error) {
	if state == "" {
		state = "all"
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/issues?state=%s&per_page=100", g.config.Repo, state)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+g.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: HTTP %d", resp.StatusCode)
	}

	var issues []GitHubIssue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	items := make([]ExternalItem, 0, len(issues))
	for _, issue := range issues {
		items = append(items, g.issueToItem(issue))
	}

	return items, nil
}

// issueToItem converts a GitHub issue to an ExternalItem
func (g *GitHubAdapter) issueToItem(issue GitHubIssue) ExternalItem {
	// Extract requirement ID from body
	reqID := ""
	if issue.Body != "" {
		// Look for patterns like "RTMX: REQ-XX-NNN" or "[REQ-XX-NNN]"
		re := regexp.MustCompile(`(?:RTMX:|REQ-)\s*(REQ-[A-Z]+-\d+)`)
		if matches := re.FindStringSubmatch(issue.Body); len(matches) > 1 {
			reqID = matches[1]
		}
	}

	// Extract labels
	labels := make([]string, len(issue.Labels))
	for i, label := range issue.Labels {
		labels[i] = label.Name
	}

	// Extract assignee
	assignee := ""
	if issue.Assignee != nil {
		assignee = issue.Assignee.Login
	}

	createdAt := issue.CreatedAt
	updatedAt := issue.UpdatedAt

	return ExternalItem{
		ExternalID:    fmt.Sprintf("%d", issue.Number),
		Title:         issue.Title,
		Description:   issue.Body,
		Status:        issue.State,
		Labels:        labels,
		URL:           issue.HTMLURL,
		CreatedAt:     &createdAt,
		UpdatedAt:     &updatedAt,
		Assignee:      assignee,
		Priority:      g.extractPriority(labels),
		RequirementID: reqID,
	}
}

// extractPriority extracts priority from issue labels
func (g *GitHubAdapter) extractPriority(labels []string) string {
	priorityMap := map[string]string{
		"priority:critical": "CRITICAL",
		"priority:high":     "HIGH",
		"priority:medium":   "MEDIUM",
		"priority:low":      "LOW",
		"p0":                "CRITICAL",
		"p1":                "HIGH",
		"p2":                "MEDIUM",
		"p3":                "LOW",
	}

	for _, label := range labels {
		if priority, ok := priorityMap[strings.ToLower(label)]; ok {
			return priority
		}
	}

	return ""
}

// CreateIssue creates a new GitHub issue
func (g *GitHubAdapter) CreateIssue(ctx context.Context, title, body string, labels []string) (*ExternalItem, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues", g.config.Repo)

	payload := map[string]interface{}{
		"title":  title,
		"body":   body,
		"labels": labels,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+g.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return nil, fmt.Errorf("API error: HTTP %d", resp.StatusCode)
	}

	var issue GitHubIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	item := g.issueToItem(issue)
	return &item, nil
}

// MapStatusToRTMX maps GitHub issue state to RTMX status
func (g *GitHubAdapter) MapStatusToRTMX(state string) string {
	switch strings.ToLower(state) {
	case "closed":
		return "COMPLETE"
	case "open":
		return "MISSING"
	default:
		return "MISSING"
	}
}

// MapStatusToGitHub maps RTMX status to GitHub issue state
func (g *GitHubAdapter) MapStatusToGitHub(status string) string {
	switch strings.ToUpper(status) {
	case "COMPLETE":
		return "closed"
	default:
		return "open"
	}
}
