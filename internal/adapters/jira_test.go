package adapters

import (
	"os"
	"testing"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
)

// TestJiraAdapter validates the complete Jira adapter functionality
// REQ-GO-032: Go CLI shall implement Jira adapter
func TestJiraAdapter(t *testing.T) {
	os.Setenv("TEST_JIRA_TOKEN", "test-token")
	os.Setenv("TEST_JIRA_EMAIL", "test@example.com")
	defer func() {
		os.Unsetenv("TEST_JIRA_TOKEN")
		os.Unsetenv("TEST_JIRA_EMAIL")
	}()

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_JIRA_TOKEN",
		EmailEnv: "TEST_JIRA_EMAIL",
	}

	adapter, err := NewJiraAdapter(&cfg)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	// Validate adapter interface
	if adapter.Name() != "jira" {
		t.Errorf("Expected name 'jira', got '%s'", adapter.Name())
	}
	if !adapter.IsConfigured() {
		t.Error("Expected adapter to be configured")
	}

	// Test status mapping (RTMX <-> Jira)
	if adapter.MapStatusToRTMX("Done") != database.StatusComplete {
		t.Error("Expected 'Done' to map to 'COMPLETE'")
	}
	if adapter.MapStatusFromRTMX(database.StatusComplete) != "Done" {
		t.Error("Expected 'COMPLETE' to map to 'Done'")
	}
}

func TestNewJiraAdapter(t *testing.T) {
	// Test without token
	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_JIRA_TOKEN",
		EmailEnv: "TEST_JIRA_EMAIL",
	}

	// Ensure env vars are not set
	os.Unsetenv("TEST_JIRA_TOKEN")
	os.Unsetenv("TEST_JIRA_EMAIL")

	_, err := NewJiraAdapter(&cfg)
	if err == nil {
		t.Error("Expected error when token is not set")
	}

	// Test with token but no email
	os.Setenv("TEST_JIRA_TOKEN", "test-token")
	defer os.Unsetenv("TEST_JIRA_TOKEN")

	_, err = NewJiraAdapter(&cfg)
	if err == nil {
		t.Error("Expected error when email is not set")
	}

	// Test with both token and email
	os.Setenv("TEST_JIRA_EMAIL", "test@example.com")
	defer os.Unsetenv("TEST_JIRA_EMAIL")

	adapter, err := NewJiraAdapter(&cfg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if adapter.Name() != "jira" {
		t.Errorf("Expected name 'jira', got '%s'", adapter.Name())
	}

	if !adapter.IsConfigured() {
		t.Error("Expected adapter to be configured")
	}
}

func TestJiraAdapterDisabled(t *testing.T) {
	cfg := config.JiraAdapterConfig{
		Enabled: false,
		Server:  "https://test.atlassian.net",
		Project: "TEST",
	}

	_, err := NewJiraAdapter(&cfg)
	if err == nil {
		t.Error("Expected error when adapter is disabled")
	}
}

func TestJiraMapStatusToRTMX(t *testing.T) {
	os.Setenv("TEST_JIRA_TOKEN", "test-token")
	os.Setenv("TEST_JIRA_EMAIL", "test@example.com")
	defer func() {
		os.Unsetenv("TEST_JIRA_TOKEN")
		os.Unsetenv("TEST_JIRA_EMAIL")
	}()

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_JIRA_TOKEN",
		EmailEnv: "TEST_JIRA_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg)

	tests := []struct {
		status   string
		expected database.Status
	}{
		{"Done", database.StatusComplete},
		{"Closed", database.StatusComplete},
		{"Complete", database.StatusComplete},
		{"In Progress", database.StatusPartial},
		{"In Review", database.StatusPartial},
		{"Open", database.StatusMissing},
		{"To Do", database.StatusMissing},
		{"Unknown", database.StatusMissing},
	}

	for _, tt := range tests {
		result := adapter.MapStatusToRTMX(tt.status)
		if result != tt.expected {
			t.Errorf("MapStatusToRTMX(%s) = %s, want %s", tt.status, result, tt.expected)
		}
	}
}

func TestJiraMapStatusFromRTMX(t *testing.T) {
	os.Setenv("TEST_JIRA_TOKEN", "test-token")
	os.Setenv("TEST_JIRA_EMAIL", "test@example.com")
	defer func() {
		os.Unsetenv("TEST_JIRA_TOKEN")
		os.Unsetenv("TEST_JIRA_EMAIL")
	}()

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_JIRA_TOKEN",
		EmailEnv: "TEST_JIRA_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg)

	tests := []struct {
		status   database.Status
		expected string
	}{
		{database.StatusComplete, "Done"},
		{database.StatusPartial, "In Progress"},
		{database.StatusMissing, "Open"},
	}

	for _, tt := range tests {
		result := adapter.MapStatusFromRTMX(tt.status)
		if result != tt.expected {
			t.Errorf("MapStatusFromRTMX(%s) = %s, want %s", tt.status, result, tt.expected)
		}
	}
}

func TestJiraIssueToItem(t *testing.T) {
	os.Setenv("TEST_JIRA_TOKEN", "test-token")
	os.Setenv("TEST_JIRA_EMAIL", "test@example.com")
	defer func() {
		os.Unsetenv("TEST_JIRA_TOKEN")
		os.Unsetenv("TEST_JIRA_EMAIL")
	}()

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_JIRA_TOKEN",
		EmailEnv: "TEST_JIRA_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg)

	issue := JiraIssue{
		Key:  "TEST-123",
		Self: "https://test.atlassian.net/rest/api/3/issue/TEST-123",
	}
	issue.Fields.Summary = "Test Issue"
	issue.Fields.Description = "RTMX: REQ-TEST-001\nDescription here"
	issue.Fields.Status.Name = "In Progress"
	issue.Fields.Labels = []string{"requirement", "p1"}
	issue.Fields.Created = "2024-01-01T00:00:00Z"
	issue.Fields.Updated = "2024-01-02T00:00:00Z"
	issue.Fields.Priority = &struct {
		Name string `json:"name"`
	}{Name: "High"}
	issue.Fields.Assignee = &struct {
		DisplayName string `json:"displayName"`
	}{DisplayName: "Test User"}

	item := adapter.issueToItem(issue)

	if item.ExternalID != "TEST-123" {
		t.Errorf("Expected ExternalID 'TEST-123', got '%s'", item.ExternalID)
	}
	if item.Title != "Test Issue" {
		t.Errorf("Expected Title 'Test Issue', got '%s'", item.Title)
	}
	if item.RequirementID != "REQ-TEST-001" {
		t.Errorf("Expected RequirementID 'REQ-TEST-001', got '%s'", item.RequirementID)
	}
	if item.Priority != "High" {
		t.Errorf("Expected Priority 'High', got '%s'", item.Priority)
	}
	if item.Assignee != "Test User" {
		t.Errorf("Expected Assignee 'Test User', got '%s'", item.Assignee)
	}
	if item.Status != "In Progress" {
		t.Errorf("Expected Status 'In Progress', got '%s'", item.Status)
	}
}

func TestJiraWithCustomStatusMapping(t *testing.T) {
	os.Setenv("TEST_JIRA_TOKEN", "test-token")
	os.Setenv("TEST_JIRA_EMAIL", "test@example.com")
	defer func() {
		os.Unsetenv("TEST_JIRA_TOKEN")
		os.Unsetenv("TEST_JIRA_EMAIL")
	}()

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_JIRA_TOKEN",
		EmailEnv: "TEST_JIRA_EMAIL",
		StatusMapping: map[string]string{
			"Finished": "COMPLETE",
			"WIP":      "PARTIAL",
			"Backlog":  "MISSING",
		},
	}

	adapter, _ := NewJiraAdapter(&cfg)

	// Test custom status mapping
	if adapter.MapStatusToRTMX("Finished") != database.StatusComplete {
		t.Error("Expected 'Finished' to map to 'COMPLETE' via custom mapping")
	}
	if adapter.MapStatusToRTMX("WIP") != database.StatusPartial {
		t.Error("Expected 'WIP' to map to 'PARTIAL' via custom mapping")
	}
	if adapter.MapStatusToRTMX("Backlog") != database.StatusMissing {
		t.Error("Expected 'Backlog' to map to 'MISSING' via custom mapping")
	}

	// Test reverse mapping
	if adapter.MapStatusFromRTMX(database.StatusComplete) != "Finished" {
		t.Error("Expected 'COMPLETE' to map to 'Finished' via custom mapping")
	}
}

func TestJiraIsConfigured(t *testing.T) {
	os.Setenv("TEST_JIRA_TOKEN", "test-token")
	os.Setenv("TEST_JIRA_EMAIL", "test@example.com")
	defer func() {
		os.Unsetenv("TEST_JIRA_TOKEN")
		os.Unsetenv("TEST_JIRA_EMAIL")
	}()

	tests := []struct {
		name     string
		cfg      config.JiraAdapterConfig
		expected bool
	}{
		{
			name: "fully configured",
			cfg: config.JiraAdapterConfig{
				Enabled:  true,
				Server:   "https://test.atlassian.net",
				Project:  "TEST",
				TokenEnv: "TEST_JIRA_TOKEN",
				EmailEnv: "TEST_JIRA_EMAIL",
			},
			expected: true,
		},
		{
			name: "missing server",
			cfg: config.JiraAdapterConfig{
				Enabled:  true,
				Server:   "",
				Project:  "TEST",
				TokenEnv: "TEST_JIRA_TOKEN",
				EmailEnv: "TEST_JIRA_EMAIL",
			},
			expected: false,
		},
		{
			name: "missing project",
			cfg: config.JiraAdapterConfig{
				Enabled:  true,
				Server:   "https://test.atlassian.net",
				Project:  "",
				TokenEnv: "TEST_JIRA_TOKEN",
				EmailEnv: "TEST_JIRA_EMAIL",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter, err := NewJiraAdapter(&tt.cfg)
			if err != nil {
				// Some configs will fail to create - that's expected
				if tt.expected {
					t.Errorf("Unexpected error creating adapter: %v", err)
				}
				return
			}

			if adapter.IsConfigured() != tt.expected {
				t.Errorf("IsConfigured() = %v, want %v", adapter.IsConfigured(), tt.expected)
			}
		})
	}
}

func TestJiraExtractReqIDFromDescription(t *testing.T) {
	os.Setenv("TEST_JIRA_TOKEN", "test-token")
	os.Setenv("TEST_JIRA_EMAIL", "test@example.com")
	defer func() {
		os.Unsetenv("TEST_JIRA_TOKEN")
		os.Unsetenv("TEST_JIRA_EMAIL")
	}()

	cfg := config.JiraAdapterConfig{
		Enabled:  true,
		Server:   "https://test.atlassian.net",
		Project:  "TEST",
		TokenEnv: "TEST_JIRA_TOKEN",
		EmailEnv: "TEST_JIRA_EMAIL",
	}

	adapter, _ := NewJiraAdapter(&cfg)

	tests := []struct {
		description string
		expected    string
	}{
		{"RTMX: REQ-TEST-001", "REQ-TEST-001"},
		{"Some text with RTMX: REQ-TEST-002 in the middle", "REQ-TEST-002"},
		{"No requirement here", ""},
		{"RTMX: REQ-FOO-123\nMore content", "REQ-FOO-123"},
	}

	for _, tt := range tests {
		issue := JiraIssue{}
		issue.Fields.Description = tt.description
		item := adapter.issueToItem(issue)
		if item.RequirementID != tt.expected {
			t.Errorf("For description %q, expected RequirementID %q, got %q",
				tt.description, tt.expected, item.RequirementID)
		}
	}
}
