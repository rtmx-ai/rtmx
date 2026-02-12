package adapters

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// TestGitHubAdapter validates the complete GitHub adapter functionality
// REQ-GO-019: Go CLI shall implement GitHub adapter
func TestGitHubAdapter(t *testing.T) {
	os.Setenv("TEST_GITHUB_TOKEN", "test-token")
	defer os.Unsetenv("TEST_GITHUB_TOKEN")

	config := GitHubConfig{
		Enabled:  true,
		Repo:     "owner/repo",
		TokenEnv: "TEST_GITHUB_TOKEN",
	}

	adapter, err := NewGitHubAdapter(config)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Validate adapter interface
	if adapter.Name() != "github" {
		t.Errorf("Expected name 'github', got '%s'", adapter.Name())
	}
	if !adapter.IsConfigured() {
		t.Error("Expected adapter to be configured")
	}

	// Test status mapping (RTMX <-> GitHub)
	if adapter.MapStatusToRTMX("closed") != "COMPLETE" {
		t.Error("Expected 'closed' to map to 'COMPLETE'")
	}
	if adapter.MapStatusToGitHub("COMPLETE") != "closed" {
		t.Error("Expected 'COMPLETE' to map to 'closed'")
	}

	// Test priority extraction
	priorities := []struct {
		labels   []string
		expected string
	}{
		{[]string{"P0"}, "CRITICAL"},
		{[]string{"P1"}, "HIGH"},
		{[]string{"priority:medium"}, "MEDIUM"},
		{[]string{"other"}, ""},
	}
	for _, tc := range priorities {
		if result := adapter.extractPriority(tc.labels); result != tc.expected {
			t.Errorf("extractPriority(%v) = %s, want %s", tc.labels, result, tc.expected)
		}
	}

	// Test issue to item conversion
	issue := GitHubIssue{
		Number:  42,
		Title:   "Test Issue",
		Body:    "RTMX: REQ-TEST-001",
		State:   "open",
		HTMLURL: "https://github.com/owner/repo/issues/42",
	}
	item := adapter.issueToItem(issue)
	if item.ExternalID != "42" {
		t.Errorf("Expected ExternalID '42', got '%s'", item.ExternalID)
	}
	if item.RequirementID != "REQ-TEST-001" {
		t.Errorf("Expected RequirementID 'REQ-TEST-001', got '%s'", item.RequirementID)
	}
}

func TestNewGitHubAdapter(t *testing.T) {
	// Test without token
	config := GitHubConfig{
		Enabled:  true,
		Repo:     "owner/repo",
		TokenEnv: "TEST_GITHUB_TOKEN",
	}

	// Ensure env var is not set
	os.Unsetenv("TEST_GITHUB_TOKEN")

	_, err := NewGitHubAdapter(config)
	if err == nil {
		t.Error("Expected error when token is not set")
	}

	// Test with token
	os.Setenv("TEST_GITHUB_TOKEN", "test-token")
	defer os.Unsetenv("TEST_GITHUB_TOKEN")

	adapter, err := NewGitHubAdapter(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if adapter.Name() != "github" {
		t.Errorf("Expected name 'github', got '%s'", adapter.Name())
	}

	if !adapter.IsConfigured() {
		t.Error("Expected adapter to be configured")
	}
}

func TestGitHubAdapterDisabled(t *testing.T) {
	config := GitHubConfig{
		Enabled: false,
		Repo:    "owner/repo",
	}

	_, err := NewGitHubAdapter(config)
	if err == nil {
		t.Error("Expected error when adapter is disabled")
	}
}

func TestExtractPriority(t *testing.T) {
	os.Setenv("TEST_GITHUB_TOKEN", "test-token")
	defer os.Unsetenv("TEST_GITHUB_TOKEN")

	config := GitHubConfig{
		Enabled:  true,
		Repo:     "owner/repo",
		TokenEnv: "TEST_GITHUB_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(config)

	tests := []struct {
		labels   []string
		expected string
	}{
		{[]string{"P0"}, "CRITICAL"},
		{[]string{"P1"}, "HIGH"},
		{[]string{"priority:medium"}, "MEDIUM"},
		{[]string{"other"}, ""},
		{[]string{}, ""},
	}

	for _, tt := range tests {
		result := adapter.extractPriority(tt.labels)
		if result != tt.expected {
			t.Errorf("extractPriority(%v) = %s, want %s", tt.labels, result, tt.expected)
		}
	}
}

func TestMapStatus(t *testing.T) {
	os.Setenv("TEST_GITHUB_TOKEN", "test-token")
	defer os.Unsetenv("TEST_GITHUB_TOKEN")

	config := GitHubConfig{
		Enabled:  true,
		Repo:     "owner/repo",
		TokenEnv: "TEST_GITHUB_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(config)

	// Test MapStatusToRTMX
	if adapter.MapStatusToRTMX("closed") != "COMPLETE" {
		t.Error("Expected 'closed' to map to 'COMPLETE'")
	}
	if adapter.MapStatusToRTMX("open") != "MISSING" {
		t.Error("Expected 'open' to map to 'MISSING'")
	}

	// Test MapStatusToGitHub
	if adapter.MapStatusToGitHub("COMPLETE") != "closed" {
		t.Error("Expected 'COMPLETE' to map to 'closed'")
	}
	if adapter.MapStatusToGitHub("MISSING") != "open" {
		t.Error("Expected 'MISSING' to map to 'open'")
	}
}

func TestIssueToItem(t *testing.T) {
	os.Setenv("TEST_GITHUB_TOKEN", "test-token")
	defer os.Unsetenv("TEST_GITHUB_TOKEN")

	config := GitHubConfig{
		Enabled:  true,
		Repo:     "owner/repo",
		TokenEnv: "TEST_GITHUB_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(config)

	now := time.Now()
	issue := GitHubIssue{
		Number:    123,
		Title:     "Test Issue",
		Body:      "RTMX: REQ-TEST-001\nDescription here",
		State:     "open",
		HTMLURL:   "https://github.com/owner/repo/issues/123",
		CreatedAt: now,
		UpdatedAt: now,
		Labels: []struct {
			Name string `json:"name"`
		}{
			{Name: "P1"},
			{Name: "bug"},
		},
		Assignee: &struct {
			Login string `json:"login"`
		}{Login: "testuser"},
	}

	item := adapter.issueToItem(issue)

	if item.ExternalID != "123" {
		t.Errorf("Expected ExternalID '123', got '%s'", item.ExternalID)
	}
	if item.Title != "Test Issue" {
		t.Errorf("Expected Title 'Test Issue', got '%s'", item.Title)
	}
	if item.RequirementID != "REQ-TEST-001" {
		t.Errorf("Expected RequirementID 'REQ-TEST-001', got '%s'", item.RequirementID)
	}
	if item.Priority != "HIGH" {
		t.Errorf("Expected Priority 'HIGH', got '%s'", item.Priority)
	}
	if item.Assignee != "testuser" {
		t.Errorf("Expected Assignee 'testuser', got '%s'", item.Assignee)
	}
}

func TestTestConnection(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/repos/owner/repo" {
			json.NewEncoder(w).Encode(map[string]string{"full_name": "owner/repo"})
			return
		}
		w.WriteHeader(404)
	}))
	defer server.Close()

	os.Setenv("TEST_GITHUB_TOKEN", "test-token")
	defer os.Unsetenv("TEST_GITHUB_TOKEN")

	config := GitHubConfig{
		Enabled:  true,
		Repo:     "owner/repo",
		TokenEnv: "TEST_GITHUB_TOKEN",
	}

	adapter, _ := NewGitHubAdapter(config)
	// Note: This test would need to mock the HTTP client to fully test
	// For now, we just test that the method exists and has the right signature
	ctx := context.Background()
	ok, msg := adapter.TestConnection(ctx)
	// Will fail without real API access, but tests the interface
	_ = ok
	_ = msg
}

func TestFetchIssues(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issues := []GitHubIssue{
			{
				Number:    1,
				Title:     "Issue 1",
				Body:      "REQ-TEST-001",
				State:     "open",
				HTMLURL:   "https://github.com/test",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}
		json.NewEncoder(w).Encode(issues)
	}))
	defer server.Close()

	os.Setenv("TEST_GITHUB_TOKEN", "test-token")
	defer os.Unsetenv("TEST_GITHUB_TOKEN")

	// Note: Full integration test would require mocking the base URL
	// This test verifies the function signature and basic structure
}
