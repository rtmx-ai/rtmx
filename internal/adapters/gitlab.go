package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
)

// GitLabAdapter syncs requirements with GitLab Issues via the REST API v4.
type GitLabAdapter struct {
	config    *config.GitLabAdapterConfig
	client    HTTPClient
	getEnv    func(string) string
	token     string
	serverURL string // base URL without trailing slash
}

// GitLabIssue represents a GitLab issue from the API.
type GitLabIssue struct {
	IID         int       `json:"iid"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	State       string    `json:"state"`
	WebURL      string    `json:"web_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Labels      []string  `json:"labels"`
	Assignee    *struct {
		Username string `json:"username"`
	} `json:"assignee"`
	Milestone *GitLabMilestone `json:"milestone"`
}

// GitLabMilestone represents a GitLab milestone.
type GitLabMilestone struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

// GitLabPipeline represents a GitLab pipeline from the API.
type GitLabPipeline struct {
	ID     int    `json:"id"`
	Status string `json:"status"`
	WebURL string `json:"web_url"`
}

// NewGitLabAdapter creates a new GitLab adapter.
// Options can be provided to inject custom dependencies for testing.
func NewGitLabAdapter(cfg *config.GitLabAdapterConfig, opts ...AdapterOption) (*GitLabAdapter, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("GitLab adapter is not enabled")
	}

	if cfg.Project == "" {
		return nil, fmt.Errorf("GitLab project is required")
	}

	options := applyOptions(opts)

	tokenEnv := cfg.TokenEnv
	if tokenEnv == "" {
		tokenEnv = "GITLAB_TOKEN"
	}

	token := options.getEnv(tokenEnv)
	if token == "" {
		return nil, fmt.Errorf("GitLab token not found. Set %s environment variable", tokenEnv)
	}

	serverURL := cfg.Server
	if serverURL == "" {
		serverURL = "https://gitlab.com"
	}
	serverURL = strings.TrimRight(serverURL, "/")

	return &GitLabAdapter{
		config:    cfg,
		client:    options.httpClient,
		getEnv:    options.getEnv,
		token:     token,
		serverURL: serverURL,
	}, nil
}

// Name returns the adapter name.
func (g *GitLabAdapter) Name() string {
	return "gitlab"
}

// IsConfigured checks if the adapter is properly configured.
func (g *GitLabAdapter) IsConfigured() bool {
	return g.config.Enabled && g.config.Project != "" && g.token != ""
}

// apiURL returns the full API v4 URL for the given path segments.
// The project ID/path is URL-encoded automatically.
func (g *GitLabAdapter) apiURL(pathSegments ...string) string {
	encodedProject := url.PathEscape(g.config.Project)
	base := fmt.Sprintf("%s/api/v4/projects/%s", g.serverURL, encodedProject)
	if len(pathSegments) > 0 {
		base += "/" + strings.Join(pathSegments, "/")
	}
	return base
}

// TestConnection tests the connection to GitLab.
func (g *GitLabAdapter) TestConnection() (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqURL := g.apiURL()

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return false, fmt.Sprintf("Failed to create request: %v", err)
	}

	req.Header.Set("PRIVATE-TOKEN", g.token)

	resp, err := g.client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Connection failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return false, fmt.Sprintf("Connection failed: HTTP %d", resp.StatusCode)
	}

	var project struct {
		PathWithNamespace string `json:"path_with_namespace"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return false, fmt.Sprintf("Failed to parse response: %v", err)
	}

	return true, fmt.Sprintf("Connected to %s", project.PathWithNamespace)
}

// FetchItems fetches issues from GitLab.
func (g *GitLabAdapter) FetchItems(query map[string]interface{}) ([]ExternalItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	state := "all"
	if query != nil {
		if s, ok := query["state"].(string); ok {
			state = s
		}
	}

	reqURL := fmt.Sprintf("%s?state=%s&per_page=100", g.apiURL("issues"), state)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", g.token)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: HTTP %d", resp.StatusCode)
	}

	var issues []GitLabIssue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	items := make([]ExternalItem, 0, len(issues))
	for _, issue := range issues {
		items = append(items, g.issueToItem(issue))
	}

	return items, nil
}

// GetItem gets a single issue by its IID.
func (g *GitLabAdapter) GetItem(externalID string) (*ExternalItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqURL := g.apiURL("issues", externalID)

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", g.token)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: HTTP %d", resp.StatusCode)
	}

	var issue GitLabIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	item := g.issueToItem(issue)
	return &item, nil
}

// CreateItem creates a new GitLab issue from a requirement.
func (g *GitLabAdapter) CreateItem(req *database.Requirement) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqURL := g.apiURL("issues")

	desc := req.RequirementText
	if req.Notes != "" {
		desc += "\n\n## Notes\n" + req.Notes
	}
	desc += fmt.Sprintf("\n\n---\nRTMX: %s", req.ReqID)

	title := fmt.Sprintf("[%s] %s", req.ReqID, truncateStr(req.RequirementText, 80))

	payload := map[string]interface{}{
		"title":       title,
		"description": desc,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("PRIVATE-TOKEN", g.token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 201 {
		return "", fmt.Errorf("API error: HTTP %d", resp.StatusCode)
	}

	var issue GitLabIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return fmt.Sprintf("%d", issue.IID), nil
}

// UpdateItem updates an existing GitLab issue.
func (g *GitLabAdapter) UpdateItem(externalID string, req *database.Requirement) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqURL := g.apiURL("issues", externalID)

	desc := req.RequirementText
	if req.Notes != "" {
		desc += "\n\n## Notes\n" + req.Notes
	}
	desc += fmt.Sprintf("\n\n---\nRTMX: %s", req.ReqID)

	title := fmt.Sprintf("[%s] %s", req.ReqID, truncateStr(req.RequirementText, 80))

	stateEvent := g.MapStatusFromRTMX(req.Status)
	payload := map[string]interface{}{
		"title":       title,
		"description": desc,
	}
	// GitLab uses state_event (close/reopen) rather than setting state directly.
	if stateEvent != "" {
		payload["state_event"] = stateEvent
	}

	payloadBytes, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", reqURL, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return false
	}

	httpReq.Header.Set("PRIVATE-TOKEN", g.token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(httpReq)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == 200
}

// MapStatusToRTMX maps GitLab issue state and labels to RTMX status.
// It first checks the configurable StatusMapping, then falls back to defaults.
func (g *GitLabAdapter) MapStatusToRTMX(externalStatus string) database.Status {
	lower := strings.ToLower(externalStatus)

	// Check user-defined status mapping first.
	if g.config.StatusMapping != nil {
		if mapped, ok := g.config.StatusMapping[lower]; ok {
			switch strings.ToUpper(mapped) {
			case "COMPLETE":
				return database.StatusComplete
			case "PARTIAL":
				return database.StatusPartial
			case "NOT_STARTED":
				return database.StatusNotStarted
			default:
				return database.StatusMissing
			}
		}
	}

	// Default mapping.
	switch lower {
	case "closed":
		return database.StatusComplete
	case "opened":
		return database.StatusMissing
	default:
		return database.StatusMissing
	}
}

// MapStatusFromRTMX maps RTMX status to a GitLab state_event value.
// GitLab uses "close" and "reopen" events rather than direct state assignment.
func (g *GitLabAdapter) MapStatusFromRTMX(status database.Status) string {
	switch status {
	case database.StatusComplete:
		return "close"
	default:
		return "reopen"
	}
}

// GetPipelineStatus returns the CI pipeline status for a given merge request IID.
func (g *GitLabAdapter) GetPipelineStatus(mergeRequestIID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	reqURL := fmt.Sprintf("%s?per_page=1", g.apiURL("merge_requests", mergeRequestIID, "pipelines"))

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("PRIVATE-TOKEN", g.token)

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error: HTTP %d", resp.StatusCode)
	}

	var pipelines []GitLabPipeline
	if err := json.NewDecoder(resp.Body).Decode(&pipelines); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(pipelines) == 0 {
		return "none", nil
	}

	return pipelines[0].Status, nil
}

// issueToItem converts a GitLab issue to an ExternalItem.
func (g *GitLabAdapter) issueToItem(issue GitLabIssue) ExternalItem {
	// Extract requirement ID from description, then title
	reqID := ExtractReqID(issue.Description)
	if reqID == "" {
		reqID = ExtractReqID(issue.Title)
	}

	assignee := ""
	if issue.Assignee != nil {
		assignee = issue.Assignee.Username
	}

	// Map milestone title to version-style string for RTMX.
	version := ""
	if issue.Milestone != nil {
		version = issue.Milestone.Title
	}

	// Note: version from milestone is extracted but ExternalItem does not
	// have a Version field. It is available through MilestoneToVersion()
	// for consumers that need release version mapping.
	_ = version

	return ExternalItem{
		ExternalID:    fmt.Sprintf("%d", issue.IID),
		Title:         issue.Title,
		Description:   issue.Description,
		Status:        issue.State,
		Labels:        issue.Labels,
		URL:           issue.WebURL,
		CreatedAt:     issue.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     issue.UpdatedAt.Format(time.RFC3339),
		Assignee:      assignee,
		Priority:      g.extractPriority(issue.Labels),
		RequirementID: reqID,
	}
}

// extractPriority extracts priority from GitLab issue labels.
func (g *GitLabAdapter) extractPriority(labels []string) string {
	priorityMap := map[string]string{
		"priority::critical": "P0",
		"priority::high":     "HIGH",
		"priority::medium":   "MEDIUM",
		"priority::low":      "LOW",
		"p0":                 "P0",
		"p1":                 "HIGH",
		"p2":                 "MEDIUM",
		"p3":                 "LOW",
	}

	for _, label := range labels {
		if priority, ok := priorityMap[strings.ToLower(label)]; ok {
			return priority
		}
	}

	return ""
}

// MilestoneToVersion maps a GitLab milestone title to an RTMX version string.
// This supports bidirectional sync where milestones correspond to release versions.
func (g *GitLabAdapter) MilestoneToVersion(milestoneTitle string) string {
	// If the milestone title already looks like a version (vX.Y.Z), return as-is.
	versionPattern := regexp.MustCompile(`^v?\d+\.\d+(\.\d+)?$`)
	if versionPattern.MatchString(milestoneTitle) {
		if !strings.HasPrefix(milestoneTitle, "v") {
			return "v" + milestoneTitle
		}
		return milestoneTitle
	}
	// Otherwise return the raw title.
	return milestoneTitle
}

// VersionToMilestone maps an RTMX version string to a GitLab milestone title.
func (g *GitLabAdapter) VersionToMilestone(version string) string {
	// Strip the "v" prefix for milestone titles if present.
	return strings.TrimPrefix(version, "v")
}
