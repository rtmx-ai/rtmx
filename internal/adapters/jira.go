package adapters

import (
	"context"
	"encoding/base64"
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

// JiraAdapter syncs requirements with Jira tickets
type JiraAdapter struct {
	config *config.JiraAdapterConfig
	client *http.Client
	auth   string // base64 encoded email:token
}

// JiraIssue represents a Jira issue from the API
type JiraIssue struct {
	Key    string `json:"key"`
	Fields struct {
		Summary     string `json:"summary"`
		Description string `json:"description"`
		Status      struct {
			Name string `json:"name"`
		} `json:"status"`
		Priority *struct {
			Name string `json:"name"`
		} `json:"priority"`
		Labels   []string `json:"labels"`
		Assignee *struct {
			DisplayName string `json:"displayName"`
		} `json:"assignee"`
		Created string `json:"created"`
		Updated string `json:"updated"`
	} `json:"fields"`
	Self string `json:"self"`
}

// JiraSearchResponse represents a Jira search API response
type JiraSearchResponse struct {
	Issues     []JiraIssue `json:"issues"`
	Total      int         `json:"total"`
	MaxResults int         `json:"maxResults"`
	StartAt    int         `json:"startAt"`
}

// NewJiraAdapter creates a new Jira adapter
func NewJiraAdapter(cfg *config.JiraAdapterConfig) (*JiraAdapter, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("Jira adapter is not enabled")
	}

	tokenEnv := cfg.TokenEnv
	if tokenEnv == "" {
		tokenEnv = "JIRA_API_TOKEN"
	}

	emailEnv := cfg.EmailEnv
	if emailEnv == "" {
		emailEnv = "JIRA_EMAIL"
	}

	token := os.Getenv(tokenEnv)
	email := os.Getenv(emailEnv)

	if token == "" {
		return nil, fmt.Errorf("Jira API token not found. Set %s environment variable", tokenEnv)
	}
	if email == "" {
		return nil, fmt.Errorf("Jira email not found. Set %s environment variable", emailEnv)
	}

	// Create basic auth string
	auth := base64.StdEncoding.EncodeToString([]byte(email + ":" + token))

	return &JiraAdapter{
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
		auth:   auth,
	}, nil
}

// Name returns the adapter name
func (j *JiraAdapter) Name() string {
	return "jira"
}

// IsConfigured checks if the adapter is properly configured
func (j *JiraAdapter) IsConfigured() bool {
	return j.config.Enabled && j.config.Server != "" && j.config.Project != "" && j.auth != ""
}

// TestConnection tests the connection to Jira
func (j *JiraAdapter) TestConnection() (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/rest/api/3/project/%s", strings.TrimSuffix(j.config.Server, "/"), j.config.Project)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Sprintf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Basic "+j.auth)
	req.Header.Set("Accept", "application/json")

	resp, err := j.client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Connection failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false, fmt.Sprintf("Connection failed: HTTP %d", resp.StatusCode)
	}

	var project struct {
		Key  string `json:"key"`
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return false, fmt.Sprintf("Failed to parse response: %v", err)
	}

	return true, fmt.Sprintf("Connected to %s (%s)", project.Name, project.Key)
}

// FetchItems fetches issues from Jira
func (j *JiraAdapter) FetchItems(query map[string]interface{}) ([]ExternalItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Build JQL query
	var jql string
	if query != nil && query["jql"] != nil {
		jql = query["jql"].(string)
	} else {
		jqlParts := []string{fmt.Sprintf("project = %s", j.config.Project)}

		if j.config.IssueType != "" {
			jqlParts = append(jqlParts, fmt.Sprintf("issuetype = '%s'", j.config.IssueType))
		}

		if query != nil {
			if status, ok := query["status"].(string); ok {
				jqlParts = append(jqlParts, fmt.Sprintf("status = '%s'", status))
			}
			if labels, ok := query["labels"].([]string); ok {
				for _, label := range labels {
					jqlParts = append(jqlParts, fmt.Sprintf("labels = '%s'", label))
				}
			}
		}

		jql = strings.Join(jqlParts, " AND ")
	}

	var allItems []ExternalItem
	startAt := 0
	maxResults := 50

	for {
		url := fmt.Sprintf("%s/rest/api/3/search?jql=%s&startAt=%d&maxResults=%d",
			strings.TrimSuffix(j.config.Server, "/"),
			jql,
			startAt,
			maxResults)

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Basic "+j.auth)
		req.Header.Set("Accept", "application/json")

		resp, err := j.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("API error: HTTP %d", resp.StatusCode)
		}

		var searchResp JiraSearchResponse
		if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		for _, issue := range searchResp.Issues {
			allItems = append(allItems, j.issueToItem(issue))
		}

		if len(searchResp.Issues) < maxResults {
			break
		}

		startAt += maxResults
	}

	return allItems, nil
}

// GetItem gets a single issue by key
func (j *JiraAdapter) GetItem(externalID string) (*ExternalItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/rest/api/3/issue/%s", strings.TrimSuffix(j.config.Server, "/"), externalID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+j.auth)
	req.Header.Set("Accept", "application/json")

	resp, err := j.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: HTTP %d", resp.StatusCode)
	}

	var issue JiraIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	item := j.issueToItem(issue)
	return &item, nil
}

// CreateItem creates a new Jira issue from a requirement
func (j *JiraAdapter) CreateItem(req *database.Requirement) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/rest/api/3/issue", strings.TrimSuffix(j.config.Server, "/"))

	// Build description with ADF format (Atlassian Document Format)
	descText := req.RequirementText
	if req.Notes != "" {
		descText += "\n\nNotes:\n" + req.Notes
	}
	descText += fmt.Sprintf("\n\n---\nRTMX: %s", req.ReqID)

	// Simple ADF description
	description := map[string]interface{}{
		"type":    "doc",
		"version": 1,
		"content": []map[string]interface{}{
			{
				"type": "paragraph",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": descText,
					},
				},
			},
		},
	}

	issueType := j.config.IssueType
	if issueType == "" {
		issueType = "Task"
	}

	payload := map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]string{
				"key": j.config.Project,
			},
			"summary":     fmt.Sprintf("[%s] %s", req.ReqID, truncateStr(req.RequirementText, 80)),
			"description": description,
			"issuetype": map[string]string{
				"name": issueType,
			},
		},
	}

	// Add labels if configured
	if len(j.config.Labels) > 0 {
		payload["fields"].(map[string]interface{})["labels"] = j.config.Labels
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Basic "+j.auth)
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := j.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 201 {
		return "", fmt.Errorf("API error: HTTP %d", resp.StatusCode)
	}

	var created struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return created.Key, nil
}

// UpdateItem updates an existing Jira issue
func (j *JiraAdapter) UpdateItem(externalID string, req *database.Requirement) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/rest/api/3/issue/%s", strings.TrimSuffix(j.config.Server, "/"), externalID)

	// Build description with ADF format
	descText := req.RequirementText
	if req.Notes != "" {
		descText += "\n\nNotes:\n" + req.Notes
	}
	descText += fmt.Sprintf("\n\n---\nRTMX: %s", req.ReqID)

	description := map[string]interface{}{
		"type":    "doc",
		"version": 1,
		"content": []map[string]interface{}{
			{
				"type": "paragraph",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": descText,
					},
				},
			},
		},
	}

	payload := map[string]interface{}{
		"fields": map[string]interface{}{
			"summary":     fmt.Sprintf("[%s] %s", req.ReqID, truncateStr(req.RequirementText, 80)),
			"description": description,
		},
	}

	payloadBytes, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return false
	}

	httpReq.Header.Set("Authorization", "Basic "+j.auth)
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := j.client.Do(httpReq)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Note: Jira returns 204 No Content on successful update
	if resp.StatusCode != 204 && resp.StatusCode != 200 {
		return false
	}

	// Transition status if needed
	targetStatus := j.MapStatusFromRTMX(req.Status)
	return j.transitionIssue(ctx, externalID, targetStatus)
}

// transitionIssue transitions an issue to a new status
func (j *JiraAdapter) transitionIssue(ctx context.Context, issueKey string, targetStatus string) bool {
	// Get available transitions
	url := fmt.Sprintf("%s/rest/api/3/issue/%s/transitions", strings.TrimSuffix(j.config.Server, "/"), issueKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Basic "+j.auth)
	req.Header.Set("Accept", "application/json")

	resp, err := j.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return false
	}

	var transResp struct {
		Transitions []struct {
			ID string `json:"id"`
			To struct {
				Name string `json:"name"`
			} `json:"to"`
		} `json:"transitions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&transResp); err != nil {
		return false
	}

	// Find matching transition
	var transitionID string
	for _, t := range transResp.Transitions {
		if strings.EqualFold(t.To.Name, targetStatus) {
			transitionID = t.ID
			break
		}
	}

	if transitionID == "" {
		return true // No transition needed or available
	}

	// Perform transition
	transReq, _ := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(fmt.Sprintf(`{"transition":{"id":"%s"}}`, transitionID)))
	transReq.Header.Set("Authorization", "Basic "+j.auth)
	transReq.Header.Set("Content-Type", "application/json")

	transResp2, err := j.client.Do(transReq)
	if err != nil {
		return false
	}
	defer transResp2.Body.Close()

	return transResp2.StatusCode == 204 || transResp2.StatusCode == 200
}

// MapStatusToRTMX maps Jira status to RTMX status
func (j *JiraAdapter) MapStatusToRTMX(status string) database.Status {
	// Use configured mapping if available
	if j.config.StatusMapping != nil {
		if rtmxStatus, ok := j.config.StatusMapping[status]; ok {
			if parsed, err := database.ParseStatus(rtmxStatus); err == nil {
				return parsed
			}
		}
	}

	// Default mapping
	statusLower := strings.ToLower(status)
	switch {
	case strings.Contains(statusLower, "done") || strings.Contains(statusLower, "closed") || strings.Contains(statusLower, "complete"):
		return database.StatusComplete
	case strings.Contains(statusLower, "progress") || strings.Contains(statusLower, "review"):
		return database.StatusPartial
	default:
		return database.StatusMissing
	}
}

// MapStatusFromRTMX maps RTMX status to Jira status
func (j *JiraAdapter) MapStatusFromRTMX(status database.Status) string {
	// Reverse the status mapping
	if j.config.StatusMapping != nil {
		for jiraStatus, rtmxStatus := range j.config.StatusMapping {
			if parsed, err := database.ParseStatus(rtmxStatus); err == nil && parsed == status {
				return jiraStatus
			}
		}
	}

	// Default mapping
	switch status {
	case database.StatusComplete:
		return "Done"
	case database.StatusPartial:
		return "In Progress"
	default:
		return "Open"
	}
}

// issueToItem converts a Jira issue to an ExternalItem
func (j *JiraAdapter) issueToItem(issue JiraIssue) ExternalItem {
	// Extract requirement ID from description
	reqID := ""
	if issue.Fields.Description != "" {
		re := regexp.MustCompile(`(?:RTMX:|REQ-)\s*(REQ-[A-Z]+-\d+)`)
		if matches := re.FindStringSubmatch(issue.Fields.Description); len(matches) > 1 {
			reqID = matches[1]
		}
	}

	// Extract assignee
	assignee := ""
	if issue.Fields.Assignee != nil {
		assignee = issue.Fields.Assignee.DisplayName
	}

	// Extract priority
	priority := ""
	if issue.Fields.Priority != nil {
		priority = issue.Fields.Priority.Name
	}

	// Build URL
	issueURL := fmt.Sprintf("%s/browse/%s", strings.TrimSuffix(j.config.Server, "/"), issue.Key)

	return ExternalItem{
		ExternalID:    issue.Key,
		Title:         issue.Fields.Summary,
		Description:   issue.Fields.Description,
		Status:        issue.Fields.Status.Name,
		Labels:        issue.Fields.Labels,
		URL:           issueURL,
		CreatedAt:     issue.Fields.Created,
		UpdatedAt:     issue.Fields.Updated,
		Assignee:      assignee,
		Priority:      priority,
		RequirementID: reqID,
	}
}
