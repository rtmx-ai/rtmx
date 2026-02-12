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

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
)

// GitHubAdapter syncs requirements with GitHub Issues
type GitHubAdapter struct {
	config *config.GitHubAdapterConfig
	client *http.Client
	token  string
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
func NewGitHubAdapter(cfg *config.GitHubAdapterConfig) (*GitHubAdapter, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("GitHub adapter is not enabled")
	}

	tokenEnv := cfg.TokenEnv
	if tokenEnv == "" {
		tokenEnv = "GITHUB_TOKEN"
	}

	token := os.Getenv(tokenEnv)
	if token == "" {
		return nil, fmt.Errorf("GitHub token not found. Set %s environment variable", tokenEnv)
	}

	return &GitHubAdapter{
		config: cfg,
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
func (g *GitHubAdapter) TestConnection() (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

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

// FetchItems fetches issues from GitHub
func (g *GitHubAdapter) FetchItems(query map[string]interface{}) ([]ExternalItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	state := "all"
	if query != nil {
		if s, ok := query["state"].(string); ok {
			state = s
		}
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

// GetItem gets a single issue by number
func (g *GitHubAdapter) GetItem(externalID string) (*ExternalItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s", g.config.Repo, externalID)

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

	var issue GitHubIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	item := g.issueToItem(issue)
	return &item, nil
}

// CreateItem creates a new GitHub issue from a requirement
func (g *GitHubAdapter) CreateItem(req *database.Requirement) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://api.github.com/repos/%s/issues", g.config.Repo)

	// Build description
	desc := req.RequirementText
	if req.Notes != "" {
		desc += "\n\n## Notes\n" + req.Notes
	}
	desc += fmt.Sprintf("\n\n---\nRTMX: %s", req.ReqID)

	title := fmt.Sprintf("[%s] %s", req.ReqID, truncateStr(req.RequirementText, 80))

	payload := map[string]interface{}{
		"title": title,
		"body":  desc,
	}

	// Add labels if configured
	if g.config.Labels.Requirement != "" {
		payload["labels"] = []string{g.config.Labels.Requirement}
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "token "+g.token)
	httpReq.Header.Set("Accept", "application/vnd.github.v3+json")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return "", fmt.Errorf("API error: HTTP %d", resp.StatusCode)
	}

	var issue GitHubIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return fmt.Sprintf("%d", issue.Number), nil
}

// UpdateItem updates an existing GitHub issue
func (g *GitHubAdapter) UpdateItem(externalID string, req *database.Requirement) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s", g.config.Repo, externalID)

	// Build description
	desc := req.RequirementText
	if req.Notes != "" {
		desc += "\n\n## Notes\n" + req.Notes
	}
	desc += fmt.Sprintf("\n\n---\nRTMX: %s", req.ReqID)

	title := fmt.Sprintf("[%s] %s", req.ReqID, truncateStr(req.RequirementText, 80))

	payload := map[string]interface{}{
		"title": title,
		"body":  desc,
		"state": g.MapStatusFromRTMX(req.Status),
	}

	payloadBytes, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx, "PATCH", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return false
	}

	httpReq.Header.Set("Authorization", "token "+g.token)
	httpReq.Header.Set("Accept", "application/vnd.github.v3+json")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

// MapStatusToRTMX maps GitHub issue state to RTMX status
func (g *GitHubAdapter) MapStatusToRTMX(state string) database.Status {
	switch strings.ToLower(state) {
	case "closed":
		return database.StatusComplete
	case "open":
		return database.StatusMissing
	default:
		return database.StatusMissing
	}
}

// MapStatusFromRTMX maps RTMX status to GitHub issue state
func (g *GitHubAdapter) MapStatusFromRTMX(status database.Status) string {
	switch status {
	case database.StatusComplete:
		return "closed"
	default:
		return "open"
	}
}

// issueToItem converts a GitHub issue to an ExternalItem
func (g *GitHubAdapter) issueToItem(issue GitHubIssue) ExternalItem {
	// Extract requirement ID from body
	reqID := ""
	if issue.Body != "" {
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

	return ExternalItem{
		ExternalID:    fmt.Sprintf("%d", issue.Number),
		Title:         issue.Title,
		Description:   issue.Body,
		Status:        issue.State,
		Labels:        labels,
		URL:           issue.HTMLURL,
		CreatedAt:     issue.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     issue.UpdatedAt.Format(time.RFC3339),
		Assignee:      assignee,
		Priority:      g.extractPriority(labels),
		RequirementID: reqID,
	}
}

// extractPriority extracts priority from issue labels
func (g *GitHubAdapter) extractPriority(labels []string) string {
	priorityMap := map[string]string{
		"priority:critical": "P0",
		"priority:high":     "HIGH",
		"priority:medium":   "MEDIUM",
		"priority:low":      "LOW",
		"p0":                "P0",
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

func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
