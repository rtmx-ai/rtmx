package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
)

// AsanaAdapter syncs requirements with Asana tasks.
type AsanaAdapter struct {
	config  *config.AsanaAdapterConfig
	client  HTTPClient
	getEnv  func(string) string
	token   string
	baseURL string
}

// AsanaTask represents an Asana task from the API.
type AsanaTask struct {
	GID          string        `json:"gid"`
	Name         string        `json:"name"`
	Notes        string        `json:"notes"`
	Completed    bool          `json:"completed"`
	CreatedAt    string        `json:"created_at"`
	ModifiedAt   string        `json:"modified_at"`
	Permalink    string        `json:"permalink_url"`
	Assignee     *AsanaUser    `json:"assignee"`
	Memberships  []AsanaMember `json:"memberships"`
}

// AsanaUser represents an Asana user.
type AsanaUser struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

// AsanaMember represents a task's project membership with section.
type AsanaMember struct {
	Project AsanaRef `json:"project"`
	Section AsanaRef `json:"section"`
}

// AsanaRef is a minimal Asana object reference.
type AsanaRef struct {
	GID  string `json:"gid"`
	Name string `json:"name"`
}

// NewAsanaAdapter creates a new Asana adapter.
func NewAsanaAdapter(cfg *config.AsanaAdapterConfig, opts ...AdapterOption) (*AsanaAdapter, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("Asana adapter is not enabled")
	}

	options := applyOptions(opts)
	tokenEnv := cfg.TokenEnv
	if tokenEnv == "" {
		tokenEnv = "ASANA_TOKEN"
	}
	token := options.getEnv(tokenEnv)
	if token == "" {
		return nil, fmt.Errorf("Asana token not found. Set %s environment variable", tokenEnv)
	}

	return &AsanaAdapter{
		config:  cfg,
		client:  options.httpClient,
		getEnv:  options.getEnv,
		token:   token,
		baseURL: "https://app.asana.com/api/1.0",
	}, nil
}

// SetBaseURL overrides the API base URL (for testing).
func (a *AsanaAdapter) SetBaseURL(url string) { a.baseURL = url }

func (a *AsanaAdapter) Name() string { return "asana" }

func (a *AsanaAdapter) IsConfigured() bool {
	return a.config.Enabled && a.config.ProjectGID != "" && a.token != ""
}

func (a *AsanaAdapter) TestConnection() (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/users/me", a.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, fmt.Sprintf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := a.client.Do(req)
	if err != nil {
		return false, fmt.Sprintf("Connection failed: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return false, fmt.Sprintf("Connection failed: HTTP %d", resp.StatusCode)
	}

	var body struct {
		Data AsanaUser `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return false, fmt.Sprintf("Failed to parse response: %v", err)
	}
	return true, fmt.Sprintf("Connected as %s", body.Data.Name)
}

func (a *AsanaAdapter) FetchItems(query map[string]interface{}) ([]ExternalItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/projects/%s/tasks?opt_fields=gid,name,notes,completed,created_at,modified_at,permalink_url,assignee.name,memberships.section.name,memberships.project.name",
		a.baseURL, a.config.ProjectGID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: HTTP %d", resp.StatusCode)
	}

	var body struct {
		Data []AsanaTask `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	items := make([]ExternalItem, 0, len(body.Data))
	for _, task := range body.Data {
		items = append(items, a.taskToItem(task))
	}
	return items, nil
}

func (a *AsanaAdapter) GetItem(externalID string) (*ExternalItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/tasks/%s?opt_fields=gid,name,notes,completed,created_at,modified_at,permalink_url,assignee.name,memberships.section.name,memberships.project.name",
		a.baseURL, externalID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API error: HTTP %d", resp.StatusCode)
	}

	var body struct {
		Data AsanaTask `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	item := a.taskToItem(body.Data)
	return &item, nil
}

func (a *AsanaAdapter) CreateItem(req *database.Requirement) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/tasks", a.baseURL)
	notes := req.RequirementText
	if req.Notes != "" {
		notes += "\n\nNotes: " + req.Notes
	}
	notes += fmt.Sprintf("\n\nRTMX: %s", req.ReqID)

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"name":      fmt.Sprintf("[%s] %s", req.ReqID, truncateStr(req.RequirementText, 80)),
			"notes":     notes,
			"completed": req.IsComplete(),
			"projects":  []string{a.config.ProjectGID},
		},
	}

	payloadBytes, _ := json.Marshal(payload)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+a.token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 201 {
		return "", fmt.Errorf("API error: HTTP %d", resp.StatusCode)
	}

	var body struct {
		Data struct {
			GID string `json:"gid"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	return body.Data.GID, nil
}

func (a *AsanaAdapter) UpdateItem(externalID string, req *database.Requirement) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/tasks/%s", a.baseURL, externalID)
	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"name":      fmt.Sprintf("[%s] %s", req.ReqID, truncateStr(req.RequirementText, 80)),
			"completed": req.IsComplete(),
		},
	}
	payloadBytes, _ := json.Marshal(payload)

	httpReq, err := http.NewRequestWithContext(ctx, "PUT", url, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return false
	}
	httpReq.Header.Set("Authorization", "Bearer "+a.token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode == 200
}

func (a *AsanaAdapter) MapStatusToRTMX(status string) database.Status {
	if a.config.StatusMapping != nil {
		if mapped, ok := a.config.StatusMapping[status]; ok {
			return database.Status(mapped)
		}
	}
	switch strings.ToLower(status) {
	case "completed", "complete", "true":
		return database.StatusComplete
	case "in progress", "partial":
		return database.StatusPartial
	default:
		return database.StatusMissing
	}
}

func (a *AsanaAdapter) MapStatusFromRTMX(status database.Status) string {
	switch status {
	case database.StatusComplete:
		return "completed"
	case database.StatusPartial:
		return "in progress"
	default:
		return "not started"
	}
}

// SectionToCategory maps an Asana section name to an RTMX category.
func (a *AsanaAdapter) SectionToCategory(sectionName string) string {
	// Convention: section names map directly to categories (uppercased first word)
	parts := strings.Fields(sectionName)
	if len(parts) == 0 {
		return "UNCATEGORIZED"
	}
	return strings.ToUpper(parts[0])
}

// SectionToPhase maps an Asana section's position to an RTMX phase.
func (a *AsanaAdapter) SectionToPhase(sectionIndex int) int {
	return sectionIndex + 1
}

func (a *AsanaAdapter) taskToItem(task AsanaTask) ExternalItem {
	reqID := ExtractReqID(task.Notes)
	if reqID == "" {
		reqID = ExtractReqID(task.Name)
	}
	assignee := ""
	if task.Assignee != nil {
		assignee = task.Assignee.Name
	}
	status := "not started"
	if task.Completed {
		status = "completed"
	}
	category := ""
	if len(task.Memberships) > 0 {
		category = task.Memberships[0].Section.Name
	}

	return ExternalItem{
		ExternalID:    task.GID,
		Title:         task.Name,
		Description:   task.Notes,
		Status:        status,
		URL:           task.Permalink,
		CreatedAt:     task.CreatedAt,
		UpdatedAt:     task.ModifiedAt,
		Assignee:      assignee,
		Labels:        []string{category},
		RequirementID: reqID,
	}
}

