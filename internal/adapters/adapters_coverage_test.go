package adapters

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
)

// REQ-GO-065: Go CLI adapters shall achieve 100% test coverage

// =============================================================================
// GitHub Adapter Tests
// =============================================================================

func TestGitHubFetchItems(t *testing.T) {
	tests := []struct {
		name       string
		query      map[string]interface{}
		response   string
		statusCode int
		wantErr    bool
		wantCount  int
	}{
		{
			name:  "fetch all issues",
			query: nil,
			response: `[
				{"number":1,"title":"Issue 1","body":"RTMX: REQ-TEST-001","state":"open","html_url":"https://github.com/test"},
				{"number":2,"title":"Issue 2","body":"Test body","state":"closed","html_url":"https://github.com/test"}
			]`,
			statusCode: 200,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:       "fetch with state filter",
			query:      map[string]interface{}{"state": "open"},
			response:   `[{"number":1,"title":"Issue 1","body":"","state":"open","html_url":"https://github.com/test"}]`,
			statusCode: 200,
			wantErr:    false,
			wantCount:  1,
		},
		{
			name:       "API error",
			query:      nil,
			response:   `{"error":"forbidden"}`,
			statusCode: 403,
			wantErr:    true,
			wantCount:  0,
		},
		{
			name:       "invalid JSON response",
			query:      nil,
			response:   `invalid json`,
			statusCode: 200,
			wantErr:    true,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				Response: &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(bytes.NewBufferString(tt.response)),
				},
			}

			cfg := config.GitHubAdapterConfig{
				Enabled:  true,
				Repo:     "owner/repo",
				TokenEnv: "TEST_TOKEN",
			}

			adapter, err := NewGitHubAdapter(&cfg,
				WithHTTPClient(mockClient),
				WithEnvGetter(func(key string) string {
					if key == "TEST_TOKEN" {
						return "test-token"
					}
					return ""
				}),
			)
			if err != nil {
				t.Fatalf("Failed to create adapter: %v", err)
			}

			items, err := adapter.FetchItems(tt.query)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(items) != tt.wantCount {
				t.Errorf("Expected %d items, got %d", tt.wantCount, len(items))
			}
		})
	}
}

func TestGitHubFetchItemsRequestError(t *testing.T) {
	mockClient := &MockHTTPClient{
		Err: errors.New("connection refused"),
	}

	cfg := config.GitHubAdapterConfig{
		Enabled:  true,
		Repo:     "owner/repo",
		TokenEnv: "TEST_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string { return "test-token" }),
	)

	_, err := adapter.FetchItems(nil)
	if err == nil {
		t.Error("Expected error on request failure")
	}
}

func TestGitHubGetItem(t *testing.T) {
	tests := []struct {
		name       string
		externalID string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "get existing issue",
			externalID: "123",
			response:   `{"number":123,"title":"Test Issue","body":"RTMX: REQ-TEST-001","state":"open","html_url":"https://github.com/test"}`,
			statusCode: 200,
			wantErr:    false,
		},
		{
			name:       "issue not found",
			externalID: "999",
			response:   `{"message":"Not Found"}`,
			statusCode: 404,
			wantErr:    true,
		},
		{
			name:       "invalid JSON",
			externalID: "123",
			response:   `{invalid}`,
			statusCode: 200,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				Response: &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(bytes.NewBufferString(tt.response)),
				},
			}

			cfg := config.GitHubAdapterConfig{
				Enabled:  true,
				Repo:     "owner/repo",
				TokenEnv: "TEST_TOKEN",
			}

			adapter, _ := NewGitHubAdapter(&cfg,
				WithHTTPClient(mockClient),
				WithEnvGetter(func(key string) string { return "test-token" }),
			)

			item, err := adapter.GetItem(tt.externalID)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if item == nil {
				t.Error("Expected item, got nil")
			}
		})
	}
}

func TestGitHubGetItemRequestError(t *testing.T) {
	mockClient := &MockHTTPClient{
		Err: errors.New("connection refused"),
	}

	cfg := config.GitHubAdapterConfig{
		Enabled:  true,
		Repo:     "owner/repo",
		TokenEnv: "TEST_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string { return "test-token" }),
	)

	_, err := adapter.GetItem("123")
	if err == nil {
		t.Error("Expected error on request failure")
	}
}

func TestGitHubCreateItem(t *testing.T) {
	tests := []struct {
		name       string
		req        *database.Requirement
		labels     config.GitHubLabels
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name: "create issue successfully",
			req: &database.Requirement{
				ReqID:           "REQ-TEST-001",
				RequirementText: "Test requirement",
			},
			response:   `{"number":42}`,
			statusCode: 201,
			wantErr:    false,
		},
		{
			name: "create issue with notes",
			req: &database.Requirement{
				ReqID:           "REQ-TEST-002",
				RequirementText: "Test requirement with notes",
				Notes:           "Additional notes here",
			},
			response:   `{"number":43}`,
			statusCode: 201,
			wantErr:    false,
		},
		{
			name: "create issue with labels",
			req: &database.Requirement{
				ReqID:           "REQ-TEST-003",
				RequirementText: "Test requirement",
			},
			labels:     config.GitHubLabels{Requirement: "requirement"},
			response:   `{"number":44}`,
			statusCode: 201,
			wantErr:    false,
		},
		{
			name: "API error",
			req: &database.Requirement{
				ReqID:           "REQ-TEST-004",
				RequirementText: "Test requirement",
			},
			response:   `{"message":"Forbidden"}`,
			statusCode: 403,
			wantErr:    true,
		},
		{
			name: "invalid JSON response",
			req: &database.Requirement{
				ReqID:           "REQ-TEST-005",
				RequirementText: "Test requirement",
			},
			response:   `{invalid}`,
			statusCode: 201,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				Response: &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(bytes.NewBufferString(tt.response)),
				},
			}

			cfg := config.GitHubAdapterConfig{
				Enabled:  true,
				Repo:     "owner/repo",
				TokenEnv: "TEST_TOKEN",
				Labels:   tt.labels,
			}

			adapter, _ := NewGitHubAdapter(&cfg,
				WithHTTPClient(mockClient),
				WithEnvGetter(func(key string) string { return "test-token" }),
			)

			externalID, err := adapter.CreateItem(tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if externalID == "" {
				t.Error("Expected external ID, got empty string")
			}
		})
	}
}

func TestGitHubCreateItemRequestError(t *testing.T) {
	mockClient := &MockHTTPClient{
		Err: errors.New("connection refused"),
	}

	cfg := config.GitHubAdapterConfig{
		Enabled:  true,
		Repo:     "owner/repo",
		TokenEnv: "TEST_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string { return "test-token" }),
	)

	req := &database.Requirement{
		ReqID:           "REQ-TEST-001",
		RequirementText: "Test requirement",
	}

	_, err := adapter.CreateItem(req)
	if err == nil {
		t.Error("Expected error on request failure")
	}
}

func TestGitHubUpdateItem(t *testing.T) {
	tests := []struct {
		name       string
		externalID string
		req        *database.Requirement
		statusCode int
		wantOK     bool
	}{
		{
			name:       "update issue successfully",
			externalID: "123",
			req: &database.Requirement{
				ReqID:           "REQ-TEST-001",
				RequirementText: "Updated requirement",
				Status:          database.StatusComplete,
			},
			statusCode: 200,
			wantOK:     true,
		},
		{
			name:       "update issue with notes",
			externalID: "124",
			req: &database.Requirement{
				ReqID:           "REQ-TEST-002",
				RequirementText: "Updated requirement",
				Notes:           "Updated notes",
				Status:          database.StatusMissing,
			},
			statusCode: 200,
			wantOK:     true,
		},
		{
			name:       "update fails with API error",
			externalID: "125",
			req: &database.Requirement{
				ReqID:           "REQ-TEST-003",
				RequirementText: "Test requirement",
				Status:          database.StatusPartial,
			},
			statusCode: 403,
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				Response: &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
				},
			}

			cfg := config.GitHubAdapterConfig{
				Enabled:  true,
				Repo:     "owner/repo",
				TokenEnv: "TEST_TOKEN",
			}

			adapter, _ := NewGitHubAdapter(&cfg,
				WithHTTPClient(mockClient),
				WithEnvGetter(func(key string) string { return "test-token" }),
			)

			ok := adapter.UpdateItem(tt.externalID, tt.req)
			if ok != tt.wantOK {
				t.Errorf("UpdateItem() = %v, want %v", ok, tt.wantOK)
			}
		})
	}
}

func TestGitHubUpdateItemRequestError(t *testing.T) {
	mockClient := &MockHTTPClient{
		Err: errors.New("connection refused"),
	}

	cfg := config.GitHubAdapterConfig{
		Enabled:  true,
		Repo:     "owner/repo",
		TokenEnv: "TEST_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string { return "test-token" }),
	)

	req := &database.Requirement{
		ReqID:           "REQ-TEST-001",
		RequirementText: "Test requirement",
		Status:          database.StatusComplete,
	}

	ok := adapter.UpdateItem("123", req)
	if ok {
		t.Error("Expected false on request failure")
	}
}

func TestGitHubTestConnectionFailure(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		statusCode int
		clientErr  error
	}{
		{
			name:       "API returns 401",
			statusCode: 401,
			response:   `{"message":"Unauthorized"}`,
		},
		{
			name:       "API returns 404",
			statusCode: 404,
			response:   `{"message":"Not Found"}`,
		},
		{
			name:       "invalid JSON response",
			statusCode: 200,
			response:   `{invalid}`,
		},
		{
			name:      "connection error",
			clientErr: errors.New("connection refused"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockClient *MockHTTPClient
			if tt.clientErr != nil {
				mockClient = &MockHTTPClient{Err: tt.clientErr}
			} else {
				mockClient = &MockHTTPClient{
					Response: &http.Response{
						StatusCode: tt.statusCode,
						Body:       io.NopCloser(bytes.NewBufferString(tt.response)),
					},
				}
			}

			cfg := config.GitHubAdapterConfig{
				Enabled:  true,
				Repo:     "owner/repo",
				TokenEnv: "TEST_TOKEN",
			}

			adapter, _ := NewGitHubAdapter(&cfg,
				WithHTTPClient(mockClient),
				WithEnvGetter(func(key string) string { return "test-token" }),
			)

			success, _ := adapter.TestConnection()
			if success {
				t.Error("Expected TestConnection to fail")
			}
		})
	}
}

func TestGitHubIsConfiguredFalse(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.GitHubAdapterConfig{
		Enabled:  true,
		Repo:     "", // empty repo
		TokenEnv: "TEST_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string { return "test-token" }),
	)

	if adapter.IsConfigured() {
		t.Error("Expected IsConfigured to return false for empty repo")
	}
}

func TestGitHubMapStatusEdgeCases(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.GitHubAdapterConfig{
		Enabled:  true,
		Repo:     "owner/repo",
		TokenEnv: "TEST_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string { return "test-token" }),
	)

	// Test with unknown status
	if adapter.MapStatusToRTMX("unknown") != database.StatusMissing {
		t.Error("Expected 'unknown' to map to MISSING")
	}

	// Test case insensitivity
	if adapter.MapStatusToRTMX("CLOSED") != database.StatusComplete {
		t.Error("Expected 'CLOSED' to map to COMPLETE")
	}

	// Test MapStatusFromRTMX with PARTIAL
	if adapter.MapStatusFromRTMX(database.StatusPartial) != "open" {
		t.Error("Expected PARTIAL to map to 'open'")
	}
}

func TestGitHubIssueToItemEdgeCases(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.GitHubAdapterConfig{
		Enabled:  true,
		Repo:     "owner/repo",
		TokenEnv: "TEST_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string { return "test-token" }),
	)

	// Test with no assignee
	issue := GitHubIssue{
		Number:  1,
		Title:   "Test",
		Body:    "No RTMX marker here",
		State:   "open",
		HTMLURL: "https://github.com/test",
	}

	item := adapter.issueToItem(issue)
	if item.RequirementID != "" {
		t.Error("Expected empty RequirementID when no marker in body")
	}
	if item.Assignee != "" {
		t.Error("Expected empty Assignee when nil")
	}

	// Test with empty body
	issue2 := GitHubIssue{
		Number: 2,
		Body:   "",
	}
	item2 := adapter.issueToItem(issue2)
	if item2.RequirementID != "" {
		t.Error("Expected empty RequirementID for empty body")
	}
}

func TestGitHubExtractPriorityAllCases(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.GitHubAdapterConfig{
		Enabled:  true,
		Repo:     "owner/repo",
		TokenEnv: "TEST_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string { return "test-token" }),
	)

	tests := []struct {
		labels   []string
		expected string
	}{
		{[]string{"priority:critical"}, "P0"},
		{[]string{"priority:high"}, "HIGH"},
		{[]string{"priority:low"}, "LOW"},
		{[]string{"P0"}, "P0"},
		{[]string{"P2"}, "MEDIUM"},
		{[]string{"P3"}, "LOW"},
		{[]string{"unrelated", "another"}, ""},
	}

	for _, tt := range tests {
		result := adapter.extractPriority(tt.labels)
		if result != tt.expected {
			t.Errorf("extractPriority(%v) = %s, want %s", tt.labels, result, tt.expected)
		}
	}
}

func TestGitHubDefaultTokenEnv(t *testing.T) {
	mockClient := &MockHTTPClient{}

	// Test with empty TokenEnv (should default to GITHUB_TOKEN)
	cfg := config.GitHubAdapterConfig{
		Enabled:  true,
		Repo:     "owner/repo",
		TokenEnv: "", // empty, should default
	}

	_, err := NewGitHubAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "GITHUB_TOKEN" {
				return "default-token"
			}
			return ""
		}),
	)
	if err != nil {
		t.Fatalf("Expected no error with default GITHUB_TOKEN: %v", err)
	}
}

// =============================================================================
// Jira Adapter Tests
// =============================================================================

func TestJiraFetchItems(t *testing.T) {
	tests := []struct {
		name       string
		query      map[string]interface{}
		issueType  string
		response   string
		statusCode int
		wantErr    bool
		wantCount  int
	}{
		{
			name:  "fetch all issues",
			query: nil,
			response: `{
				"issues": [
					{"key":"TEST-1","fields":{"summary":"Issue 1","description":"RTMX: REQ-TEST-001","status":{"name":"Open"}}},
					{"key":"TEST-2","fields":{"summary":"Issue 2","description":"","status":{"name":"Done"}}}
				],
				"total": 2,
				"maxResults": 50,
				"startAt": 0
			}`,
			statusCode: 200,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:      "fetch with issue type",
			query:     nil,
			issueType: "Story",
			response: `{
				"issues": [{"key":"TEST-1","fields":{"summary":"Issue 1","status":{"name":"Open"}}}],
				"total": 1,
				"maxResults": 50,
				"startAt": 0
			}`,
			statusCode: 200,
			wantErr:    false,
			wantCount:  1,
		},
		{
			name:  "fetch with JQL",
			query: map[string]interface{}{"jql": "project = TEST AND status = Open"},
			response: `{
				"issues": [],
				"total": 0,
				"maxResults": 50,
				"startAt": 0
			}`,
			statusCode: 200,
			wantErr:    false,
			wantCount:  0,
		},
		{
			name:  "fetch with status filter",
			query: map[string]interface{}{"status": "Open"},
			response: `{
				"issues": [{"key":"TEST-1","fields":{"summary":"Issue 1","status":{"name":"Open"}}}],
				"total": 1,
				"maxResults": 50,
				"startAt": 0
			}`,
			statusCode: 200,
			wantErr:    false,
			wantCount:  1,
		},
		{
			name:  "fetch with labels filter",
			query: map[string]interface{}{"labels": []string{"requirement"}},
			response: `{
				"issues": [{"key":"TEST-1","fields":{"summary":"Issue 1","status":{"name":"Open"},"labels":["requirement"]}}],
				"total": 1,
				"maxResults": 50,
				"startAt": 0
			}`,
			statusCode: 200,
			wantErr:    false,
			wantCount:  1,
		},
		{
			name:       "API error",
			query:      nil,
			response:   `{"errorMessages":["Forbidden"]}`,
			statusCode: 403,
			wantErr:    true,
		},
		{
			name:       "invalid JSON",
			query:      nil,
			response:   `{invalid}`,
			statusCode: 200,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				Response: &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(bytes.NewBufferString(tt.response)),
				},
			}

			cfg := config.JiraAdapterConfig{
				Enabled:   true,
				Server:    "https://test.atlassian.net",
				Project:   "TEST",
				IssueType: tt.issueType,
				TokenEnv:  "TEST_TOKEN",
				EmailEnv:  "TEST_EMAIL",
			}

			adapter, _ := NewJiraAdapter(&cfg,
				WithHTTPClient(mockClient),
				WithEnvGetter(func(key string) string {
					switch key {
					case "TEST_TOKEN":
						return "test-token"
					case "TEST_EMAIL":
						return "test@example.com"
					}
					return ""
				}),
			)

			items, err := adapter.FetchItems(tt.query)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(items) != tt.wantCount {
				t.Errorf("Expected %d items, got %d", tt.wantCount, len(items))
			}
		})
	}
}

func TestJiraFetchItemsRequestError(t *testing.T) {
	mockClient := &MockHTTPClient{
		Err: errors.New("connection refused"),
	}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	_, err := adapter.FetchItems(nil)
	if err == nil {
		t.Error("Expected error on request failure")
	}
}

func TestJiraGetItem(t *testing.T) {
	tests := []struct {
		name       string
		externalID string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "get existing issue",
			externalID: "TEST-123",
			response:   `{"key":"TEST-123","fields":{"summary":"Test Issue","description":"RTMX: REQ-TEST-001","status":{"name":"Open"}}}`,
			statusCode: 200,
			wantErr:    false,
		},
		{
			name:       "issue not found",
			externalID: "TEST-999",
			response:   `{"errorMessages":["Issue Does Not Exist"]}`,
			statusCode: 404,
			wantErr:    true,
		},
		{
			name:       "invalid JSON",
			externalID: "TEST-123",
			response:   `{invalid}`,
			statusCode: 200,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				Response: &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(bytes.NewBufferString(tt.response)),
				},
			}

			cfg := config.JiraAdapterConfig{
				Enabled:  true,
				Server:   "https://test.atlassian.net",
				Project:  "TEST",
				TokenEnv: "TEST_TOKEN",
				EmailEnv: "TEST_EMAIL",
			}

			adapter, _ := NewJiraAdapter(&cfg,
				WithHTTPClient(mockClient),
				WithEnvGetter(func(key string) string {
					if key == "TEST_TOKEN" {
						return "test-token"
					}
					if key == "TEST_EMAIL" {
						return "test@example.com"
					}
					return ""
				}),
			)

			item, err := adapter.GetItem(tt.externalID)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if item == nil {
				t.Error("Expected item, got nil")
			}
		})
	}
}

func TestJiraGetItemRequestError(t *testing.T) {
	mockClient := &MockHTTPClient{
		Err: errors.New("connection refused"),
	}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	_, err := adapter.GetItem("TEST-123")
	if err == nil {
		t.Error("Expected error on request failure")
	}
}

func TestJiraCreateItem(t *testing.T) {
	tests := []struct {
		name       string
		req        *database.Requirement
		issueType  string
		labels     []string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name: "create issue successfully",
			req: &database.Requirement{
				ReqID:           "REQ-TEST-001",
				RequirementText: "Test requirement",
			},
			response:   `{"key":"TEST-42"}`,
			statusCode: 201,
			wantErr:    false,
		},
		{
			name: "create issue with notes",
			req: &database.Requirement{
				ReqID:           "REQ-TEST-002",
				RequirementText: "Test requirement with notes",
				Notes:           "Additional notes here",
			},
			response:   `{"key":"TEST-43"}`,
			statusCode: 201,
			wantErr:    false,
		},
		{
			name: "create issue with custom issue type",
			req: &database.Requirement{
				ReqID:           "REQ-TEST-003",
				RequirementText: "Test requirement",
			},
			issueType:  "Story",
			response:   `{"key":"TEST-44"}`,
			statusCode: 201,
			wantErr:    false,
		},
		{
			name: "create issue with labels",
			req: &database.Requirement{
				ReqID:           "REQ-TEST-004",
				RequirementText: "Test requirement",
			},
			labels:     []string{"requirement", "rtmx"},
			response:   `{"key":"TEST-45"}`,
			statusCode: 201,
			wantErr:    false,
		},
		{
			name: "API error",
			req: &database.Requirement{
				ReqID:           "REQ-TEST-005",
				RequirementText: "Test requirement",
			},
			response:   `{"errorMessages":["Forbidden"]}`,
			statusCode: 403,
			wantErr:    true,
		},
		{
			name: "invalid JSON response",
			req: &database.Requirement{
				ReqID:           "REQ-TEST-006",
				RequirementText: "Test requirement",
			},
			response:   `{invalid}`,
			statusCode: 201,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				Response: &http.Response{
					StatusCode: tt.statusCode,
					Body:       io.NopCloser(bytes.NewBufferString(tt.response)),
				},
			}

			cfg := config.JiraAdapterConfig{
				Enabled:   true,
				Server:    "https://test.atlassian.net",
				Project:   "TEST",
				IssueType: tt.issueType,
				Labels:    tt.labels,
				TokenEnv:  "TEST_TOKEN",
				EmailEnv:  "TEST_EMAIL",
			}

			adapter, _ := NewJiraAdapter(&cfg,
				WithHTTPClient(mockClient),
				WithEnvGetter(func(key string) string {
					if key == "TEST_TOKEN" {
						return "test-token"
					}
					if key == "TEST_EMAIL" {
						return "test@example.com"
					}
					return ""
				}),
			)

			externalID, err := adapter.CreateItem(tt.req)
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if externalID == "" {
				t.Error("Expected external ID, got empty string")
			}
		})
	}
}

func TestJiraCreateItemRequestError(t *testing.T) {
	mockClient := &MockHTTPClient{
		Err: errors.New("connection refused"),
	}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	req := &database.Requirement{
		ReqID:           "REQ-TEST-001",
		RequirementText: "Test requirement",
	}

	_, err := adapter.CreateItem(req)
	if err == nil {
		t.Error("Expected error on request failure")
	}
}

func TestJiraUpdateItem(t *testing.T) {
	// Helper to create a mock client that returns different responses for different requests
	makeMultiResponseClient := func(responses []struct {
		statusCode int
		body       string
	}) *SequentialMockClient {
		return &SequentialMockClient{responses: responses}
	}

	tests := []struct {
		name       string
		externalID string
		req        *database.Requirement
		responses  []struct {
			statusCode int
			body       string
		}
		wantOK bool
	}{
		{
			name:       "update issue successfully with transition",
			externalID: "TEST-123",
			req: &database.Requirement{
				ReqID:           "REQ-TEST-001",
				RequirementText: "Updated requirement",
				Status:          database.StatusComplete,
			},
			responses: []struct {
				statusCode int
				body       string
			}{
				{204, ""},                                                                            // Update
				{200, `{"transitions":[{"id":"1","to":{"name":"Done"}}]}`},                           // Get transitions
				{204, ""},                                                                            // Post transition
			},
			wantOK: true,
		},
		{
			name:       "update issue with notes",
			externalID: "TEST-124",
			req: &database.Requirement{
				ReqID:           "REQ-TEST-002",
				RequirementText: "Updated requirement",
				Notes:           "Updated notes",
				Status:          database.StatusMissing,
			},
			responses: []struct {
				statusCode int
				body       string
			}{
				{200, ""},                                 // Update (200 is also acceptable)
				{200, `{"transitions":[]}`},               // Get transitions (none available)
			},
			wantOK: true,
		},
		{
			name:       "update fails with API error",
			externalID: "TEST-125",
			req: &database.Requirement{
				ReqID:           "REQ-TEST-003",
				RequirementText: "Test requirement",
				Status:          database.StatusPartial,
			},
			responses: []struct {
				statusCode int
				body       string
			}{
				{403, `{"errorMessages":["Forbidden"]}`},
			},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := makeMultiResponseClient(tt.responses)

			cfg := config.JiraAdapterConfig{
				Enabled:  true,
				Server:   "https://test.atlassian.net",
				Project:  "TEST",
				TokenEnv: "TEST_TOKEN",
				EmailEnv: "TEST_EMAIL",
			}

			adapter, _ := NewJiraAdapter(&cfg,
				WithHTTPClient(mockClient),
				WithEnvGetter(func(key string) string {
					if key == "TEST_TOKEN" {
						return "test-token"
					}
					if key == "TEST_EMAIL" {
						return "test@example.com"
					}
					return ""
				}),
			)

			ok := adapter.UpdateItem(tt.externalID, tt.req)
			if ok != tt.wantOK {
				t.Errorf("UpdateItem() = %v, want %v", ok, tt.wantOK)
			}
		})
	}
}

// SequentialMockClient returns responses in sequence for multiple requests
type SequentialMockClient struct {
	responses []struct {
		statusCode int
		body       string
	}
	index int
}

func (m *SequentialMockClient) Do(req *http.Request) (*http.Response, error) {
	if m.index >= len(m.responses) {
		return &http.Response{
			StatusCode: 500,
			Body:       io.NopCloser(bytes.NewBufferString(`{"error":"no more responses"}`)),
		}, nil
	}
	resp := m.responses[m.index]
	m.index++
	return &http.Response{
		StatusCode: resp.statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(resp.body)),
	}, nil
}

func TestJiraUpdateItemRequestError(t *testing.T) {
	mockClient := &MockHTTPClient{
		Err: errors.New("connection refused"),
	}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	req := &database.Requirement{
		ReqID:           "REQ-TEST-001",
		RequirementText: "Test requirement",
		Status:          database.StatusComplete,
	}

	ok := adapter.UpdateItem("TEST-123", req)
	if ok {
		t.Error("Expected false on request failure")
	}
}

func TestJiraTransitionIssue(t *testing.T) {
	tests := []struct {
		name         string
		targetStatus string
		responses    []struct {
			statusCode int
			body       string
		}
		wantOK bool
	}{
		{
			name:         "transition to Done",
			targetStatus: "Done",
			responses: []struct {
				statusCode int
				body       string
			}{
				{200, `{"transitions":[{"id":"1","to":{"name":"Done"}},{"id":"2","to":{"name":"Open"}}]}`},
				{204, ""},
			},
			wantOK: true,
		},
		{
			name:         "no matching transition",
			targetStatus: "Done",
			responses: []struct {
				statusCode int
				body       string
			}{
				{200, `{"transitions":[{"id":"2","to":{"name":"Open"}}]}`}, // No "Done" transition
			},
			wantOK: true, // Returns true when no transition needed
		},
		{
			name:         "get transitions fails",
			targetStatus: "Done",
			responses: []struct {
				statusCode int
				body       string
			}{
				{403, `{"error":"forbidden"}`},
			},
			wantOK: false,
		},
		{
			name:         "invalid transitions JSON",
			targetStatus: "Done",
			responses: []struct {
				statusCode int
				body       string
			}{
				{200, `{invalid}`},
			},
			wantOK: false,
		},
		{
			name:         "transition post fails",
			targetStatus: "Done",
			responses: []struct {
				statusCode int
				body       string
			}{
				{200, `{"transitions":[{"id":"1","to":{"name":"Done"}}]}`},
				{403, `{"error":"forbidden"}`},
			},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &SequentialMockClient{responses: tt.responses}

			cfg := config.JiraAdapterConfig{
				Enabled:  true,
				Server:   "https://test.atlassian.net",
				Project:  "TEST",
				TokenEnv: "TEST_TOKEN",
				EmailEnv: "TEST_EMAIL",
			}

			adapter, _ := NewJiraAdapter(&cfg,
				WithHTTPClient(mockClient),
				WithEnvGetter(func(key string) string {
					if key == "TEST_TOKEN" {
						return "test-token"
					}
					if key == "TEST_EMAIL" {
						return "test@example.com"
					}
					return ""
				}),
			)

			ctx := context.Background()
			ok := adapter.transitionIssue(ctx, "TEST-123", tt.targetStatus)
			if ok != tt.wantOK {
				t.Errorf("transitionIssue() = %v, want %v", ok, tt.wantOK)
			}
		})
	}
}

func TestJiraTransitionIssueRequestError(t *testing.T) {
	mockClient := &MockHTTPClient{
		Err: errors.New("connection refused"),
	}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	ctx := context.Background()
	ok := adapter.transitionIssue(ctx, "TEST-123", "Done")
	if ok {
		t.Error("Expected false on request failure")
	}
}

func TestJiraTestConnectionFailure(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		statusCode int
		clientErr  error
	}{
		{
			name:       "API returns 401",
			statusCode: 401,
			response:   `{"errorMessages":["Unauthorized"]}`,
		},
		{
			name:       "API returns 404",
			statusCode: 404,
			response:   `{"errorMessages":["Not Found"]}`,
		},
		{
			name:       "invalid JSON response",
			statusCode: 200,
			response:   `{invalid}`,
		},
		{
			name:      "connection error",
			clientErr: errors.New("connection refused"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockClient *MockHTTPClient
			if tt.clientErr != nil {
				mockClient = &MockHTTPClient{Err: tt.clientErr}
			} else {
				mockClient = &MockHTTPClient{
					Response: &http.Response{
						StatusCode: tt.statusCode,
						Body:       io.NopCloser(bytes.NewBufferString(tt.response)),
					},
				}
			}

			cfg := config.JiraAdapterConfig{
				Enabled:  true,
				Server:   "https://test.atlassian.net",
				Project:  "TEST",
				TokenEnv: "TEST_TOKEN",
				EmailEnv: "TEST_EMAIL",
			}

			adapter, _ := NewJiraAdapter(&cfg,
				WithHTTPClient(mockClient),
				WithEnvGetter(func(key string) string {
					if key == "TEST_TOKEN" {
						return "test-token"
					}
					if key == "TEST_EMAIL" {
						return "test@example.com"
					}
					return ""
				}),
			)

			success, _ := adapter.TestConnection()
			if success {
				t.Error("Expected TestConnection to fail")
			}
		})
	}
}

func TestJiraIsConfiguredFalse(t *testing.T) {
	mockClient := &MockHTTPClient{}

	tests := []struct {
		name     string
		server   string
		project  string
		expected bool
	}{
		{"empty server", "", "TEST", false},
		{"empty project", "https://test.atlassian.net", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.JiraAdapterConfig{
				Enabled:  true,
				Server:   tt.server,
				Project:  tt.project,
				TokenEnv: "TEST_TOKEN",
				EmailEnv: "TEST_EMAIL",
			}

			adapter, err := NewJiraAdapter(&cfg,
				WithHTTPClient(mockClient),
				WithEnvGetter(func(key string) string {
					if key == "TEST_TOKEN" {
						return "test-token"
					}
					if key == "TEST_EMAIL" {
						return "test@example.com"
					}
					return ""
				}),
			)
			if err != nil {
				t.Skipf("Adapter creation failed: %v", err)
			}

			if adapter.IsConfigured() != tt.expected {
				t.Errorf("IsConfigured() = %v, want %v", adapter.IsConfigured(), tt.expected)
			}
		})
	}
}

func TestJiraIssueToItemEdgeCases(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	// Test with no assignee and no priority
	issue := JiraIssue{
		Key: "TEST-1",
	}
	issue.Fields.Summary = "Test"
	issue.Fields.Description = "No RTMX marker"
	issue.Fields.Status.Name = "Open"

	item := adapter.issueToItem(issue)
	if item.RequirementID != "" {
		t.Error("Expected empty RequirementID when no marker")
	}
	if item.Assignee != "" {
		t.Error("Expected empty Assignee when nil")
	}
	if item.Priority != "" {
		t.Error("Expected empty Priority when nil")
	}

	// Test with empty description
	issue2 := JiraIssue{
		Key: "TEST-2",
	}
	issue2.Fields.Description = ""

	item2 := adapter.issueToItem(issue2)
	if item2.RequirementID != "" {
		t.Error("Expected empty RequirementID for empty description")
	}
}

func TestJiraDefaultEnvVars(t *testing.T) {
	mockClient := &MockHTTPClient{}

	// Test with empty env var names (should default to JIRA_API_TOKEN and JIRA_EMAIL)
	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "", // empty, should default
		EmailEnv: "", // empty, should default
	}

	_, err := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			switch key {
			case "JIRA_API_TOKEN":
				return "default-token"
			case "JIRA_EMAIL":
				return "default@example.com"
			}
			return ""
		}),
	)
	if err != nil {
		t.Fatalf("Expected no error with default env vars: %v", err)
	}
}

func TestJiraServerTrailingSlash(t *testing.T) {
	mockClient := &MockHTTPClient{
		Response: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(`{"key":"TEST","name":"Test Project"}`)),
		},
	}

	// Test with trailing slash in server URL
	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net/", // trailing slash
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	success, _ := adapter.TestConnection()
	if !success {
		t.Error("Expected TestConnection to succeed with trailing slash")
	}
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestTruncateStr(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"exactly10!", 10, "exactly10!"},
		{"this is a long string", 10, "this is..."},
		{"", 10, ""},
		{"a", 1, "a"},
		{"abc", 4, "abc"},
		{"abcd", 4, "abcd"},
		{"abcde", 4, "a..."},
	}

	for _, tt := range tests {
		result := truncateStr(tt.input, tt.maxLen)
		if result != tt.expected {
			t.Errorf("truncateStr(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
		}
	}
}

func TestDefaultGetEnv(t *testing.T) {
	// Test that defaultGetEnv works correctly
	result := defaultGetEnv("PATH")
	if result == "" {
		t.Skip("PATH environment variable not set")
	}

	// Test non-existent variable
	result = defaultGetEnv("RTMX_TEST_NONEXISTENT_VAR_12345")
	if result != "" {
		t.Error("Expected empty string for non-existent variable")
	}
}

// =============================================================================
// Additional Edge Case Tests
// =============================================================================

func TestGitHubIssueWithLabelsAndAssignee(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.GitHubAdapterConfig{
		Enabled:  true,
		Repo:     "owner/repo",
		TokenEnv: "TEST_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string { return "test-token" }),
	)

	now := time.Now()
	issue := GitHubIssue{
		Number:    1,
		Title:     "Test",
		Body:      "REQ-DEMO-001 is referenced here",
		State:     "open",
		HTMLURL:   "https://github.com/test",
		CreatedAt: now,
		UpdatedAt: now,
		Labels: []struct {
			Name string `json:"name"`
		}{
			{Name: "bug"},
			{Name: "priority:critical"},
		},
		Assignee: &struct {
			Login string `json:"login"`
		}{Login: "testuser"},
	}

	item := adapter.issueToItem(issue)
	if item.Assignee != "testuser" {
		t.Errorf("Expected Assignee 'testuser', got %q", item.Assignee)
	}
	if item.Priority != "P0" {
		t.Errorf("Expected Priority 'P0', got %q", item.Priority)
	}
	if len(item.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(item.Labels))
	}
	// Note: The regex in issueToItem doesn't match "REQ-DEMO-001" without "RTMX:" prefix
	// This is expected behavior based on the regex pattern
}

func TestJiraIssueWithAllFields(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	issue := JiraIssue{
		Key:  "TEST-1",
		Self: "https://test.atlassian.net/rest/api/3/issue/TEST-1",
	}
	issue.Fields.Summary = "Full Issue"
	issue.Fields.Description = "RTMX: REQ-FULL-001\nDescription content"
	issue.Fields.Status.Name = "In Progress"
	issue.Fields.Labels = []string{"requirement", "rtmx"}
	issue.Fields.Created = "2024-01-01T00:00:00Z"
	issue.Fields.Updated = "2024-01-02T00:00:00Z"
	issue.Fields.Priority = &struct {
		Name string `json:"name"`
	}{Name: "High"}
	issue.Fields.Assignee = &struct {
		DisplayName string `json:"displayName"`
	}{DisplayName: "John Doe"}

	item := adapter.issueToItem(issue)
	if item.ExternalID != "TEST-1" {
		t.Errorf("Expected ExternalID 'TEST-1', got %q", item.ExternalID)
	}
	if item.RequirementID != "REQ-FULL-001" {
		t.Errorf("Expected RequirementID 'REQ-FULL-001', got %q", item.RequirementID)
	}
	if item.Assignee != "John Doe" {
		t.Errorf("Expected Assignee 'John Doe', got %q", item.Assignee)
	}
	if item.Priority != "High" {
		t.Errorf("Expected Priority 'High', got %q", item.Priority)
	}
	if len(item.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(item.Labels))
	}
}

func TestJiraFetchItemsPagination(t *testing.T) {
	// Create a client that returns a full page first, then partial page
	responses := []struct {
		statusCode int
		body       string
	}{
		{200, `{"issues":[{"key":"TEST-1"},{"key":"TEST-2"}],"total":3,"maxResults":2,"startAt":0}`},
		{200, `{"issues":[{"key":"TEST-3"}],"total":3,"maxResults":2,"startAt":2}`},
	}
	mockClient := &SequentialMockClient{responses: responses}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	// Note: Due to the way our mock works, pagination isn't fully simulated
	// but we're testing that the adapter handles multiple API calls
	items, err := adapter.FetchItems(nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// First response only returns 2 items
	if len(items) != 2 {
		t.Errorf("Expected 2 items from first page, got %d", len(items))
	}
}

// =============================================================================
// Test Full Pagination Flow
// =============================================================================

func TestJiraFetchItemsFullPagination(t *testing.T) {
	// Simulate multiple pages of results where maxResults matches the number of issues
	// This forces the pagination loop to continue
	responses := []struct {
		statusCode int
		body       string
	}{
		// First page: 50 results (full page)
		{200, `{"issues":[{"key":"TEST-1"},{"key":"TEST-2"},{"key":"TEST-3"},{"key":"TEST-4"},{"key":"TEST-5"},{"key":"TEST-6"},{"key":"TEST-7"},{"key":"TEST-8"},{"key":"TEST-9"},{"key":"TEST-10"},{"key":"TEST-11"},{"key":"TEST-12"},{"key":"TEST-13"},{"key":"TEST-14"},{"key":"TEST-15"},{"key":"TEST-16"},{"key":"TEST-17"},{"key":"TEST-18"},{"key":"TEST-19"},{"key":"TEST-20"},{"key":"TEST-21"},{"key":"TEST-22"},{"key":"TEST-23"},{"key":"TEST-24"},{"key":"TEST-25"},{"key":"TEST-26"},{"key":"TEST-27"},{"key":"TEST-28"},{"key":"TEST-29"},{"key":"TEST-30"},{"key":"TEST-31"},{"key":"TEST-32"},{"key":"TEST-33"},{"key":"TEST-34"},{"key":"TEST-35"},{"key":"TEST-36"},{"key":"TEST-37"},{"key":"TEST-38"},{"key":"TEST-39"},{"key":"TEST-40"},{"key":"TEST-41"},{"key":"TEST-42"},{"key":"TEST-43"},{"key":"TEST-44"},{"key":"TEST-45"},{"key":"TEST-46"},{"key":"TEST-47"},{"key":"TEST-48"},{"key":"TEST-49"},{"key":"TEST-50"}],"total":51,"maxResults":50,"startAt":0}`},
		// Second page: 1 result (partial page, triggers break)
		{200, `{"issues":[{"key":"TEST-51"}],"total":51,"maxResults":50,"startAt":50}`},
	}
	mockClient := &SequentialMockClient{responses: responses}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	items, err := adapter.FetchItems(nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// We expect 51 items total from both pages
	if len(items) != 51 {
		t.Errorf("Expected 51 items from pagination, got %d", len(items))
	}
}

// =============================================================================
// Tests for Transition POST Request Failures
// =============================================================================

func TestJiraTransitionIssuePostRequestError(t *testing.T) {
	// Create a sequential client that succeeds on GET but fails on POST
	mockClient := &TransitionPostFailClient{}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	ctx := context.Background()
	ok := adapter.transitionIssue(ctx, "TEST-123", "Done")
	if ok {
		t.Error("Expected transitionIssue to fail when POST request fails")
	}
}

// TransitionPostFailClient succeeds on GET transitions but fails on POST
type TransitionPostFailClient struct {
	requestCount int
}

func (m *TransitionPostFailClient) Do(req *http.Request) (*http.Response, error) {
	m.requestCount++
	if req.Method == "GET" {
		// Return transitions with a matching "Done" transition
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBufferString(`{"transitions":[{"id":"1","to":{"name":"Done"}}]}`)),
		}, nil
	}
	// POST request fails
	return nil, errors.New("connection refused on POST")
}

// =============================================================================
// Additional Jira MapStatus Tests
// =============================================================================

func TestJiraMapStatusWithMapping(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
		StatusMapping: map[string]string{
			"Custom Done":    "COMPLETE",
			"Custom Open":    "MISSING",
			"Custom Partial": "PARTIAL",
		},
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	// Test custom mapping
	status := adapter.MapStatusToRTMX("Custom Done")
	if status != database.StatusComplete {
		t.Errorf("Expected COMPLETE, got %v", status)
	}

	status = adapter.MapStatusToRTMX("Custom Partial")
	if status != database.StatusPartial {
		t.Errorf("Expected PARTIAL, got %v", status)
	}

	// Test reverse mapping
	jiraStatus := adapter.MapStatusFromRTMX(database.StatusComplete)
	if jiraStatus != "Custom Done" {
		t.Errorf("Expected 'Custom Done', got %q", jiraStatus)
	}

	jiraStatus = adapter.MapStatusFromRTMX(database.StatusPartial)
	if jiraStatus != "Custom Partial" {
		t.Errorf("Expected 'Custom Partial', got %q", jiraStatus)
	}
}

func TestJiraMapStatusDefaultCases(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	// Test default mappings
	tests := []struct {
		input    string
		expected database.Status
	}{
		{"Done", database.StatusComplete},
		{"Closed", database.StatusComplete},
		{"Complete", database.StatusComplete},
		{"In Progress", database.StatusPartial},
		{"In Review", database.StatusPartial},
		{"Open", database.StatusMissing},
		{"Todo", database.StatusMissing},
		{"Unknown", database.StatusMissing},
	}

	for _, tt := range tests {
		status := adapter.MapStatusToRTMX(tt.input)
		if status != tt.expected {
			t.Errorf("MapStatusToRTMX(%q) = %v, want %v", tt.input, status, tt.expected)
		}
	}
}

func TestJiraMapStatusWithInvalidMapping(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
		StatusMapping: map[string]string{
			"Custom Status": "INVALID_STATUS", // Invalid RTMX status
		},
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	// Should fall through to default mapping since INVALID_STATUS won't parse
	status := adapter.MapStatusToRTMX("Custom Status")
	if status != database.StatusMissing {
		t.Errorf("Expected MISSING for invalid mapping, got %v", status)
	}
}

// =============================================================================
// GitHub Adapter Not Enabled Test
// =============================================================================

func TestGitHubAdapterNotEnabled(t *testing.T) {
	cfg := config.GitHubAdapterConfig{
		Enabled:  false,
		Repo:     "owner/repo",
		TokenEnv: "TEST_TOKEN",
	}

	_, err := NewGitHubAdapter(&cfg)
	if err == nil {
		t.Error("Expected error when adapter is not enabled")
	}
}

func TestJiraAdapterNotEnabled(t *testing.T) {
	cfg := config.JiraAdapterConfig{
		Enabled:  false,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	_, err := NewJiraAdapter(&cfg)
	if err == nil {
		t.Error("Expected error when adapter is not enabled")
	}
}

func TestJiraMissingEmail(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	_, err := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			// Missing email
			return ""
		}),
	)
	if err == nil {
		t.Error("Expected error when email is missing")
	}
}

// =============================================================================
// Tests for GitHub Name function
// =============================================================================

func TestGitHubName(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.GitHubAdapterConfig{
		Enabled:  true,
		Repo:     "owner/repo",
		TokenEnv: "TEST_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string { return "test-token" }),
	)

	if adapter.Name() != "github" {
		t.Errorf("Expected name 'github', got %q", adapter.Name())
	}
}

func TestJiraName(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	if adapter.Name() != "jira" {
		t.Errorf("Expected name 'jira', got %q", adapter.Name())
	}
}

// =============================================================================
// Test GitHub IsConfigured with all valid values
// =============================================================================

func TestGitHubIsConfiguredTrue(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.GitHubAdapterConfig{
		Enabled:  true,
		Repo:     "owner/repo",
		TokenEnv: "TEST_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string { return "test-token" }),
	)

	if !adapter.IsConfigured() {
		t.Error("Expected IsConfigured to return true")
	}
}

func TestJiraIsConfiguredTrue(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	if !adapter.IsConfigured() {
		t.Error("Expected IsConfigured to return true")
	}
}

// =============================================================================
// Test GitHub UpdateItem with various statuses
// =============================================================================

func TestGitHubUpdateItemWithAllStatuses(t *testing.T) {
	tests := []struct {
		name        string
		status      database.Status
		expectState string
	}{
		{"complete status", database.StatusComplete, "closed"},
		{"missing status", database.StatusMissing, "open"},
		{"partial status", database.StatusPartial, "open"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockHTTPClient{
				Response: &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(`{}`)),
				},
			}

			cfg := config.GitHubAdapterConfig{
				Enabled:  true,
				Repo:     "owner/repo",
				TokenEnv: "TEST_TOKEN",
			}

			adapter, _ := NewGitHubAdapter(&cfg,
				WithHTTPClient(mockClient),
				WithEnvGetter(func(key string) string { return "test-token" }),
			)

			req := &database.Requirement{
				ReqID:           "REQ-TEST-001",
				RequirementText: "Test requirement",
				Status:          tt.status,
			}

			ok := adapter.UpdateItem("123", req)
			if !ok {
				t.Error("Expected UpdateItem to succeed")
			}
		})
	}
}

// =============================================================================
// Tests for Invalid URL Edge Cases (Request Creation Errors)
// =============================================================================

// Note: http.NewRequestWithContext fails when URL contains control characters.
// We test these paths by using repos/servers with invalid characters.

func TestGitHubTestConnectionInvalidURL(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.GitHubAdapterConfig{
		Enabled:  true,
		Repo:     "owner/repo\x00invalid", // null character makes URL invalid
		TokenEnv: "TEST_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string { return "test-token" }),
	)

	success, msg := adapter.TestConnection()
	if success {
		t.Error("Expected TestConnection to fail with invalid URL")
	}
	if msg == "" {
		t.Error("Expected error message")
	}
}

func TestGitHubFetchItemsInvalidURL(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.GitHubAdapterConfig{
		Enabled:  true,
		Repo:     "owner/repo\x00invalid", // null character makes URL invalid
		TokenEnv: "TEST_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string { return "test-token" }),
	)

	_, err := adapter.FetchItems(nil)
	if err == nil {
		t.Error("Expected error with invalid URL")
	}
}

func TestGitHubGetItemInvalidURL(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.GitHubAdapterConfig{
		Enabled:  true,
		Repo:     "owner/repo\x00invalid", // null character makes URL invalid
		TokenEnv: "TEST_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string { return "test-token" }),
	)

	_, err := adapter.GetItem("123")
	if err == nil {
		t.Error("Expected error with invalid URL")
	}
}

func TestGitHubCreateItemInvalidURL(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.GitHubAdapterConfig{
		Enabled:  true,
		Repo:     "owner/repo\x00invalid", // null character makes URL invalid
		TokenEnv: "TEST_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string { return "test-token" }),
	)

	req := &database.Requirement{
		ReqID:           "REQ-TEST-001",
		RequirementText: "Test",
	}

	_, err := adapter.CreateItem(req)
	if err == nil {
		t.Error("Expected error with invalid URL")
	}
}

func TestGitHubUpdateItemInvalidURL(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.GitHubAdapterConfig{
		Enabled:  true,
		Repo:     "owner/repo\x00invalid", // null character makes URL invalid
		TokenEnv: "TEST_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string { return "test-token" }),
	)

	req := &database.Requirement{
		ReqID:           "REQ-TEST-001",
		RequirementText: "Test",
		Status:          database.StatusComplete,
	}

	ok := adapter.UpdateItem("123", req)
	if ok {
		t.Error("Expected UpdateItem to fail with invalid URL")
	}
}

func TestJiraTestConnectionInvalidURL(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net\x00invalid", // null character
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	success, msg := adapter.TestConnection()
	if success {
		t.Error("Expected TestConnection to fail with invalid URL")
	}
	if msg == "" {
		t.Error("Expected error message")
	}
}

func TestJiraFetchItemsInvalidURL(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net\x00invalid",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	_, err := adapter.FetchItems(nil)
	if err == nil {
		t.Error("Expected error with invalid URL")
	}
}

func TestJiraGetItemInvalidURL(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net\x00invalid",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	_, err := adapter.GetItem("TEST-123")
	if err == nil {
		t.Error("Expected error with invalid URL")
	}
}

func TestJiraCreateItemInvalidURL(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net\x00invalid",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	req := &database.Requirement{
		ReqID:           "REQ-TEST-001",
		RequirementText: "Test",
	}

	_, err := adapter.CreateItem(req)
	if err == nil {
		t.Error("Expected error with invalid URL")
	}
}

func TestJiraUpdateItemInvalidURL(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net\x00invalid",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	req := &database.Requirement{
		ReqID:           "REQ-TEST-001",
		RequirementText: "Test",
		Status:          database.StatusComplete,
	}

	ok := adapter.UpdateItem("TEST-123", req)
	if ok {
		t.Error("Expected UpdateItem to fail with invalid URL")
	}
}

func TestJiraTransitionIssueInvalidURL(t *testing.T) {
	mockClient := &MockHTTPClient{}

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net\x00invalid",
		Project:  "TEST",
		TokenEnv: "TEST_TOKEN",
		EmailEnv: "TEST_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg,
		WithHTTPClient(mockClient),
		WithEnvGetter(func(key string) string {
			if key == "TEST_TOKEN" {
				return "test-token"
			}
			if key == "TEST_EMAIL" {
				return "test@example.com"
			}
			return ""
		}),
	)

	ctx := context.Background()
	ok := adapter.transitionIssue(ctx, "TEST-123", "Done")
	if ok {
		t.Error("Expected transitionIssue to fail with invalid URL")
	}
}
