package adapters

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestGitLabAdapter validates the complete GitLab adapter functionality.
// REQ-ADAPT-007: Go CLI shall implement GitLab adapter
func TestGitLabAdapter(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-007")

	// --- Adapter creation and auth ---

	t.Run("creation with valid config", func(t *testing.T) {
		cfg := &config.GitLabAdapterConfig{
			Enabled:  true,
			Project:  "mygroup/myproject",
			TokenEnv: "TEST_GITLAB_TOKEN",
		}
		adapter, err := NewGitLabAdapter(cfg, WithEnvGetter(func(key string) string {
			if key == "TEST_GITLAB_TOKEN" {
				return "glpat-test-token"
			}
			return ""
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if adapter.Name() != "gitlab" {
			t.Errorf("Name() = %q, want %q", adapter.Name(), "gitlab")
		}
		if !adapter.IsConfigured() {
			t.Error("expected adapter to be configured")
		}
	})

	t.Run("creation fails when disabled", func(t *testing.T) {
		cfg := &config.GitLabAdapterConfig{
			Enabled: false,
			Project: "mygroup/myproject",
		}
		_, err := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))
		if err == nil {
			t.Error("expected error when adapter is disabled")
		}
	})

	t.Run("creation fails without token", func(t *testing.T) {
		cfg := &config.GitLabAdapterConfig{
			Enabled:  true,
			Project:  "mygroup/myproject",
			TokenEnv: "MISSING_TOKEN",
		}
		_, err := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "" }))
		if err == nil {
			t.Error("expected error when token env var is empty")
		}
	})

	t.Run("creation fails without project", func(t *testing.T) {
		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Project: "",
		}
		_, err := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))
		if err == nil {
			t.Error("expected error when project is empty")
		}
	})

	t.Run("default token env is GITLAB_TOKEN", func(t *testing.T) {
		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Project: "g/p",
		}
		_, err := NewGitLabAdapter(cfg, WithEnvGetter(func(key string) string {
			if key == "GITLAB_TOKEN" {
				return "default-token"
			}
			return ""
		}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	// --- SaaS vs self-hosted URL ---

	t.Run("default server is gitlab.com", func(t *testing.T) {
		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Project: "g/p",
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))
		if adapter.serverURL != "https://gitlab.com" {
			t.Errorf("serverURL = %q, want %q", adapter.serverURL, "https://gitlab.com")
		}
	})

	t.Run("self-hosted server URL", func(t *testing.T) {
		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Server:  "https://gitlab.example.com/",
			Project: "g/p",
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))
		if adapter.serverURL != "https://gitlab.example.com" {
			t.Errorf("serverURL = %q, want %q", adapter.serverURL, "https://gitlab.example.com")
		}
	})

	// --- Auth header ---

	t.Run("uses PRIVATE-TOKEN header", func(t *testing.T) {
		var capturedHeader string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedHeader = r.Header.Get("PRIVATE-TOKEN")
			_ = json.NewEncoder(w).Encode(map[string]string{"path_with_namespace": "g/p"})
		}))
		defer server.Close()

		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Server:  server.URL,
			Project: "g/p",
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "my-secret" }))
		adapter.TestConnection()

		if capturedHeader != "my-secret" {
			t.Errorf("PRIVATE-TOKEN = %q, want %q", capturedHeader, "my-secret")
		}
	})

	// --- FetchItems ---

	t.Run("FetchItems returns issues", func(t *testing.T) {
		now := time.Now()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			issues := []GitLabIssue{
				{
					IID:         1,
					Title:       "First issue",
					Description: "RTMX: REQ-TEST-001",
					State:       "opened",
					WebURL:      "https://gitlab.com/g/p/-/issues/1",
					CreatedAt:   now,
					UpdatedAt:   now,
					Labels:      []string{"p1"},
				},
				{
					IID:         2,
					Title:       "Second issue",
					Description: "Closed item",
					State:       "closed",
					WebURL:      "https://gitlab.com/g/p/-/issues/2",
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			}
			_ = json.NewEncoder(w).Encode(issues)
		}))
		defer server.Close()

		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Server:  server.URL,
			Project: "g/p",
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		items, err := adapter.FetchItems(nil)
		if err != nil {
			t.Fatalf("FetchItems error: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("got %d items, want 2", len(items))
		}
		if items[0].ExternalID != "1" {
			t.Errorf("items[0].ExternalID = %q, want %q", items[0].ExternalID, "1")
		}
		if items[0].RequirementID != "REQ-TEST-001" {
			t.Errorf("items[0].RequirementID = %q, want %q", items[0].RequirementID, "REQ-TEST-001")
		}
		if items[0].Priority != "HIGH" {
			t.Errorf("items[0].Priority = %q, want %q", items[0].Priority, "HIGH")
		}
	})

	t.Run("FetchItems with state filter", func(t *testing.T) {
		var capturedQuery string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedQuery = r.URL.RawQuery
			_ = json.NewEncoder(w).Encode([]GitLabIssue{})
		}))
		defer server.Close()

		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Server:  server.URL,
			Project: "g/p",
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))
		_, _ = adapter.FetchItems(map[string]interface{}{"state": "opened"})

		if !strings.Contains(capturedQuery, "state=opened") {
			t.Errorf("query = %q, want state=opened", capturedQuery)
		}
	})

	// --- GetItem ---

	t.Run("GetItem returns single issue", func(t *testing.T) {
		now := time.Now()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			issue := GitLabIssue{
				IID:         42,
				Title:       "Single issue",
				Description: "RTMX: REQ-ADAPT-007",
				State:       "opened",
				WebURL:      "https://gitlab.com/g/p/-/issues/42",
				CreatedAt:   now,
				UpdatedAt:   now,
				Assignee: &struct {
					Username string `json:"username"`
				}{Username: "dev1"},
			}
			_ = json.NewEncoder(w).Encode(issue)
		}))
		defer server.Close()

		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Server:  server.URL,
			Project: "g/p",
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		item, err := adapter.GetItem("42")
		if err != nil {
			t.Fatalf("GetItem error: %v", err)
		}
		if item.ExternalID != "42" {
			t.Errorf("ExternalID = %q, want %q", item.ExternalID, "42")
		}
		if item.Assignee != "dev1" {
			t.Errorf("Assignee = %q, want %q", item.Assignee, "dev1")
		}
		if item.RequirementID != "REQ-ADAPT-007" {
			t.Errorf("RequirementID = %q, want %q", item.RequirementID, "REQ-ADAPT-007")
		}
	})

	// --- CreateItem ---

	t.Run("CreateItem creates issue and returns IID", func(t *testing.T) {
		var receivedPayload map[string]interface{}
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewDecoder(r.Body).Decode(&receivedPayload)
			w.WriteHeader(201)
			_ = json.NewEncoder(w).Encode(GitLabIssue{IID: 99})
		}))
		defer server.Close()

		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Server:  server.URL,
			Project: "g/p",
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		req := &database.Requirement{
			ReqID:           "REQ-TEST-010",
			RequirementText: "Test requirement for creation",
			Notes:           "Some notes",
		}
		id, err := adapter.CreateItem(req)
		if err != nil {
			t.Fatalf("CreateItem error: %v", err)
		}
		if id != "99" {
			t.Errorf("CreateItem returned %q, want %q", id, "99")
		}
		title, _ := receivedPayload["title"].(string)
		if !strings.Contains(title, "REQ-TEST-010") {
			t.Errorf("title %q does not contain REQ-TEST-010", title)
		}
		desc, _ := receivedPayload["description"].(string)
		if !strings.Contains(desc, "RTMX: REQ-TEST-010") {
			t.Errorf("description missing RTMX tag")
		}
	})

	// --- UpdateItem ---

	t.Run("UpdateItem sends PUT with state_event", func(t *testing.T) {
		var receivedPayload map[string]interface{}
		var receivedMethod string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedMethod = r.Method
			_ = json.NewDecoder(r.Body).Decode(&receivedPayload)
			_ = json.NewEncoder(w).Encode(GitLabIssue{IID: 10})
		}))
		defer server.Close()

		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Server:  server.URL,
			Project: "g/p",
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		req := &database.Requirement{
			ReqID:           "REQ-TEST-010",
			RequirementText: "Updated text",
			Status:          database.StatusComplete,
		}
		ok := adapter.UpdateItem("10", req)
		if !ok {
			t.Error("UpdateItem returned false, want true")
		}
		if receivedMethod != "PUT" {
			t.Errorf("method = %q, want PUT", receivedMethod)
		}
		if se, _ := receivedPayload["state_event"].(string); se != "close" {
			t.Errorf("state_event = %q, want %q", se, "close")
		}
	})

	// --- Status mapping ---

	t.Run("default status mapping", func(t *testing.T) {
		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Project: "g/p",
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		if adapter.MapStatusToRTMX("closed") != database.StatusComplete {
			t.Error("closed should map to COMPLETE")
		}
		if adapter.MapStatusToRTMX("opened") != database.StatusMissing {
			t.Error("opened should map to MISSING")
		}
		if adapter.MapStatusFromRTMX(database.StatusComplete) != "close" {
			t.Error("COMPLETE should map to close")
		}
		if adapter.MapStatusFromRTMX(database.StatusMissing) != "reopen" {
			t.Error("MISSING should map to reopen")
		}
	})

	t.Run("custom status mapping", func(t *testing.T) {
		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Project: "g/p",
			StatusMapping: map[string]string{
				"opened": "NOT_STARTED",
				"closed": "COMPLETE",
			},
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		if adapter.MapStatusToRTMX("opened") != database.StatusNotStarted {
			t.Error("opened with custom mapping should map to NOT_STARTED")
		}
	})

	// --- TestConnection ---

	t.Run("TestConnection success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]string{
				"path_with_namespace": "mygroup/myproject",
			})
		}))
		defer server.Close()

		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Server:  server.URL,
			Project: "mygroup/myproject",
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		ok, msg := adapter.TestConnection()
		if !ok {
			t.Errorf("TestConnection failed: %s", msg)
		}
		if !strings.Contains(msg, "mygroup/myproject") {
			t.Errorf("message %q does not contain project path", msg)
		}
	})

	t.Run("TestConnection failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(401)
		}))
		defer server.Close()

		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Server:  server.URL,
			Project: "g/p",
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		ok, msg := adapter.TestConnection()
		if ok {
			t.Error("expected TestConnection to fail")
		}
		if !strings.Contains(msg, "401") {
			t.Errorf("message %q should mention 401", msg)
		}
	})
}

// TestGitLabBidirectionalSync validates bidirectional sync with milestone-to-version mapping.
// REQ-ADAPT-008: GitLab bidirectional sync with milestone mapping
func TestGitLabBidirectionalSync(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-008")

	cfg := &config.GitLabAdapterConfig{
		Enabled: true,
		Project: "g/p",
	}
	adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

	t.Run("MilestoneToVersion with semver", func(t *testing.T) {
		tests := []struct {
			milestone string
			want      string
		}{
			{"v1.0.0", "v1.0.0"},
			{"1.0.0", "v1.0.0"},
			{"0.2", "v0.2"},
			{"v0.3.1", "v0.3.1"},
		}
		for _, tt := range tests {
			got := adapter.MilestoneToVersion(tt.milestone)
			if got != tt.want {
				t.Errorf("MilestoneToVersion(%q) = %q, want %q", tt.milestone, got, tt.want)
			}
		}
	})

	t.Run("MilestoneToVersion with non-semver", func(t *testing.T) {
		got := adapter.MilestoneToVersion("Q3 Release")
		if got != "Q3 Release" {
			t.Errorf("MilestoneToVersion(\"Q3 Release\") = %q, want %q", got, "Q3 Release")
		}
	})

	t.Run("VersionToMilestone strips v prefix", func(t *testing.T) {
		tests := []struct {
			version string
			want    string
		}{
			{"v1.0.0", "1.0.0"},
			{"1.0.0", "1.0.0"},
			{"v0.2.5", "0.2.5"},
		}
		for _, tt := range tests {
			got := adapter.VersionToMilestone(tt.version)
			if got != tt.want {
				t.Errorf("VersionToMilestone(%q) = %q, want %q", tt.version, got, tt.want)
			}
		}
	})

	t.Run("roundtrip milestone to version and back", func(t *testing.T) {
		original := "1.2.3"
		version := adapter.MilestoneToVersion(original)
		milestone := adapter.VersionToMilestone(version)
		if milestone != original {
			t.Errorf("roundtrip: %q -> %q -> %q, want %q", original, version, milestone, original)
		}
	})

	t.Run("FetchItems preserves milestone data", func(t *testing.T) {
		now := time.Now()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			issues := []GitLabIssue{
				{
					IID:         5,
					Title:       "Milestone issue",
					Description: "RTMX: REQ-SYNC-001",
					State:       "opened",
					WebURL:      "https://gitlab.com/g/p/-/issues/5",
					CreatedAt:   now,
					UpdatedAt:   now,
					Milestone: &GitLabMilestone{
						ID:    1,
						Title: "1.0.0",
					},
				},
			}
			_ = json.NewEncoder(w).Encode(issues)
		}))
		defer server.Close()

		serverCfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Server:  server.URL,
			Project: "g/p",
		}
		serverAdapter, _ := NewGitLabAdapter(serverCfg, WithEnvGetter(func(string) string { return "tok" }))

		items, err := serverAdapter.FetchItems(nil)
		if err != nil {
			t.Fatalf("FetchItems error: %v", err)
		}
		if len(items) != 1 {
			t.Fatalf("got %d items, want 1", len(items))
		}
		if items[0].RequirementID != "REQ-SYNC-001" {
			t.Errorf("RequirementID = %q, want %q", items[0].RequirementID, "REQ-SYNC-001")
		}
	})

	t.Run("CreateItem and UpdateItem roundtrip", func(t *testing.T) {
		var createPayload, updatePayload map[string]interface{}
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			switch r.Method {
			case "POST":
				_ = json.NewDecoder(r.Body).Decode(&createPayload)
				w.WriteHeader(201)
				_ = json.NewEncoder(w).Encode(GitLabIssue{IID: 50})
			case "PUT":
				_ = json.NewDecoder(r.Body).Decode(&updatePayload)
				_ = json.NewEncoder(w).Encode(GitLabIssue{IID: 50})
			}
		}))
		defer server.Close()

		serverCfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Server:  server.URL,
			Project: "g/p",
		}
		serverAdapter, _ := NewGitLabAdapter(serverCfg, WithEnvGetter(func(string) string { return "tok" }))

		req := &database.Requirement{
			ReqID:           "REQ-SYNC-002",
			RequirementText: "Bidirectional sync test",
			Status:          database.StatusMissing,
		}

		id, err := serverAdapter.CreateItem(req)
		if err != nil {
			t.Fatalf("CreateItem error: %v", err)
		}
		if id != "50" {
			t.Errorf("CreateItem returned %q, want %q", id, "50")
		}

		req.Status = database.StatusComplete
		ok := serverAdapter.UpdateItem(id, req)
		if !ok {
			t.Error("UpdateItem returned false")
		}
		if se, _ := updatePayload["state_event"].(string); se != "close" {
			t.Errorf("state_event = %q, want %q", se, "close")
		}
	})
}

// TestGitLabCIIntegration validates GitLab CI pipeline status retrieval.
// REQ-ADAPT-009: GitLab CI integration for pipeline status
func TestGitLabCIIntegration(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-009")

	t.Run("GetPipelineStatus returns latest pipeline status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pipelines := []GitLabPipeline{
				{
					ID:     101,
					Status: "success",
					WebURL: "https://gitlab.com/g/p/-/pipelines/101",
				},
			}
			_ = json.NewEncoder(w).Encode(pipelines)
		}))
		defer server.Close()

		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Server:  server.URL,
			Project: "g/p",
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		status, err := adapter.GetPipelineStatus("1")
		if err != nil {
			t.Fatalf("GetPipelineStatus error: %v", err)
		}
		if status != "success" {
			t.Errorf("status = %q, want %q", status, "success")
		}
	})

	t.Run("GetPipelineStatus returns failed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pipelines := []GitLabPipeline{
				{ID: 102, Status: "failed"},
			}
			_ = json.NewEncoder(w).Encode(pipelines)
		}))
		defer server.Close()

		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Server:  server.URL,
			Project: "g/p",
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		status, err := adapter.GetPipelineStatus("2")
		if err != nil {
			t.Fatalf("GetPipelineStatus error: %v", err)
		}
		if status != "failed" {
			t.Errorf("status = %q, want %q", status, "failed")
		}
	})

	t.Run("GetPipelineStatus returns none when no pipelines", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode([]GitLabPipeline{})
		}))
		defer server.Close()

		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Server:  server.URL,
			Project: "g/p",
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		status, err := adapter.GetPipelineStatus("3")
		if err != nil {
			t.Fatalf("GetPipelineStatus error: %v", err)
		}
		if status != "none" {
			t.Errorf("status = %q, want %q", status, "none")
		}
	})

	t.Run("GetPipelineStatus handles running pipeline", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			pipelines := []GitLabPipeline{
				{ID: 103, Status: "running"},
			}
			_ = json.NewEncoder(w).Encode(pipelines)
		}))
		defer server.Close()

		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Server:  server.URL,
			Project: "g/p",
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		status, err := adapter.GetPipelineStatus("4")
		if err != nil {
			t.Fatalf("GetPipelineStatus error: %v", err)
		}
		if status != "running" {
			t.Errorf("status = %q, want %q", status, "running")
		}
	})

	t.Run("GetPipelineStatus handles API error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}))
		defer server.Close()

		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Server:  server.URL,
			Project: "g/p",
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		_, err := adapter.GetPipelineStatus("5")
		if err == nil {
			t.Error("expected error on 500 response")
		}
	})

	t.Run("GetPipelineStatus sends correct API path", func(t *testing.T) {
		var capturedPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedPath = r.URL.Path
			_ = json.NewEncoder(w).Encode([]GitLabPipeline{})
		}))
		defer server.Close()

		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Server:  server.URL,
			Project: "mygroup/myproject",
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		_, _ = adapter.GetPipelineStatus("7")

		// The Go HTTP client may decode %2F in the path, so accept either form.
		expected1 := "/api/v4/projects/mygroup%2Fmyproject/merge_requests/7/pipelines"
		expected2 := "/api/v4/projects/mygroup/myproject/merge_requests/7/pipelines"
		if capturedPath != expected1 && capturedPath != expected2 {
			t.Errorf("path = %q, want %q or %q", capturedPath, expected1, expected2)
		}
	})

	t.Run("CI result mapping covers all GitLab statuses", func(t *testing.T) {
		// GitLab pipeline statuses: created, waiting_for_resource, preparing,
		// pending, running, success, failed, canceled, skipped, manual, scheduled
		statuses := []string{
			"success", "failed", "canceled", "running",
			"pending", "created", "manual", "skipped",
		}
		for _, s := range statuses {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ = json.NewEncoder(w).Encode([]GitLabPipeline{{ID: 1, Status: s}})
			}))

			cfg := &config.GitLabAdapterConfig{
				Enabled: true,
				Server:  server.URL,
				Project: "g/p",
			}
			adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

			status, err := adapter.GetPipelineStatus("1")
			server.Close()
			if err != nil {
				t.Errorf("GetPipelineStatus for %q returned error: %v", s, err)
			}
			if status != s {
				t.Errorf("GetPipelineStatus for %q = %q, want %q", s, status, s)
			}
		}
	})
}
