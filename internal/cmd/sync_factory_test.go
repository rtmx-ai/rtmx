package cmd

import (
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/adapters"
	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestGetAdapterFactory validates the getAdapter factory function creates
// the correct adapter type for each supported service, rejects unknown
// services, and returns meaningful errors for disabled or misconfigured
// adapters.
func TestGetAdapterFactory(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-015")

	t.Run("github_enabled", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.RTMX.Adapters.GitHub.Enabled = true
		cfg.RTMX.Adapters.GitHub.Repo = "owner/repo"
		t.Setenv("GITHUB_TOKEN", "test-token")

		adapter, err := getAdapter("github", cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if _, ok := adapter.(*adapters.GitHubAdapter); !ok {
			t.Errorf("expected *adapters.GitHubAdapter, got %T", adapter)
		}
	})

	t.Run("jira_enabled", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.RTMX.Adapters.Jira.Enabled = true
		cfg.RTMX.Adapters.Jira.Server = "https://test.atlassian.net"
		cfg.RTMX.Adapters.Jira.Project = "TEST"
		t.Setenv("JIRA_API_TOKEN", "test-token")
		t.Setenv("JIRA_EMAIL", "test@example.com")

		adapter, err := getAdapter("jira", cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if _, ok := adapter.(*adapters.JiraAdapter); !ok {
			t.Errorf("expected *adapters.JiraAdapter, got %T", adapter)
		}
	})

	t.Run("asana_enabled", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.RTMX.Adapters.Asana.Enabled = true
		cfg.RTMX.Adapters.Asana.WorkspaceGID = "12345"
		cfg.RTMX.Adapters.Asana.ProjectGID = "67890"
		t.Setenv("ASANA_TOKEN", "test-token")

		adapter, err := getAdapter("asana", cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if _, ok := adapter.(*adapters.AsanaAdapter); !ok {
			t.Errorf("expected *adapters.AsanaAdapter, got %T", adapter)
		}
	})

	t.Run("monday_enabled", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.RTMX.Adapters.Monday.Enabled = true
		cfg.RTMX.Adapters.Monday.BoardID = "12345"
		t.Setenv("MONDAY_TOKEN", "test-token")

		adapter, err := getAdapter("monday", cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if _, ok := adapter.(*adapters.MondayAdapter); !ok {
			t.Errorf("expected *adapters.MondayAdapter, got %T", adapter)
		}
	})

	t.Run("gitlab_enabled", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.RTMX.Adapters.GitLab.Enabled = true
		cfg.RTMX.Adapters.GitLab.Server = "https://gitlab.com"
		cfg.RTMX.Adapters.GitLab.Project = "group/project"
		t.Setenv("GITLAB_TOKEN", "test-token")

		adapter, err := getAdapter("gitlab", cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if _, ok := adapter.(*adapters.GitLabAdapter); !ok {
			t.Errorf("expected *adapters.GitLabAdapter, got %T", adapter)
		}
	})

	t.Run("unknown_service", func(t *testing.T) {
		cfg := config.DefaultConfig()

		_, err := getAdapter("notion", cfg)
		if err == nil {
			t.Fatal("expected error for unknown service, got nil")
		}
		if !strings.Contains(err.Error(), "unknown service") {
			t.Errorf("expected error containing 'unknown service', got %q", err.Error())
		}
	})

	t.Run("github_disabled", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.RTMX.Adapters.GitHub.Enabled = false

		_, err := getAdapter("github", cfg)
		if err == nil {
			t.Fatal("expected error for disabled adapter, got nil")
		}
		if !strings.Contains(err.Error(), "not enabled") {
			t.Errorf("expected error containing 'not enabled', got %q", err.Error())
		}
	})

	t.Run("jira_disabled", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.RTMX.Adapters.Jira.Enabled = false

		_, err := getAdapter("jira", cfg)
		if err == nil {
			t.Fatal("expected error for disabled adapter, got nil")
		}
		if !strings.Contains(err.Error(), "not enabled") {
			t.Errorf("expected error containing 'not enabled', got %q", err.Error())
		}
	})

	t.Run("asana_disabled", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.RTMX.Adapters.Asana.Enabled = false

		_, err := getAdapter("asana", cfg)
		if err == nil {
			t.Fatal("expected error for disabled adapter, got nil")
		}
		if !strings.Contains(err.Error(), "not enabled") {
			t.Errorf("expected error containing 'not enabled', got %q", err.Error())
		}
	})

	t.Run("monday_disabled", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.RTMX.Adapters.Monday.Enabled = false

		_, err := getAdapter("monday", cfg)
		if err == nil {
			t.Fatal("expected error for disabled adapter, got nil")
		}
		if !strings.Contains(err.Error(), "not enabled") {
			t.Errorf("expected error containing 'not enabled', got %q", err.Error())
		}
	})

	t.Run("gitlab_disabled", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.RTMX.Adapters.GitLab.Enabled = false

		_, err := getAdapter("gitlab", cfg)
		if err == nil {
			t.Fatal("expected error for disabled adapter, got nil")
		}
		if !strings.Contains(err.Error(), "not enabled") {
			t.Errorf("expected error containing 'not enabled', got %q", err.Error())
		}
	})

	t.Run("github_missing_token", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.RTMX.Adapters.GitHub.Enabled = true
		cfg.RTMX.Adapters.GitHub.Repo = "owner/repo"
		t.Setenv("GITHUB_TOKEN", "")

		_, err := getAdapter("github", cfg)
		if err == nil {
			t.Fatal("expected error for missing token, got nil")
		}
		if !strings.Contains(strings.ToLower(err.Error()), "token") {
			t.Errorf("expected error containing 'token', got %q", err.Error())
		}
	})

	t.Run("asana_missing_token", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.RTMX.Adapters.Asana.Enabled = true
		cfg.RTMX.Adapters.Asana.WorkspaceGID = "12345"
		cfg.RTMX.Adapters.Asana.ProjectGID = "67890"
		t.Setenv("ASANA_TOKEN", "")

		_, err := getAdapter("asana", cfg)
		if err == nil {
			t.Fatal("expected error for missing token, got nil")
		}
		if !strings.Contains(strings.ToLower(err.Error()), "token") {
			t.Errorf("expected error containing 'token', got %q", err.Error())
		}
	})

	t.Run("monday_missing_token", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.RTMX.Adapters.Monday.Enabled = true
		cfg.RTMX.Adapters.Monday.BoardID = "12345"
		t.Setenv("MONDAY_TOKEN", "")

		_, err := getAdapter("monday", cfg)
		if err == nil {
			t.Fatal("expected error for missing token, got nil")
		}
		if !strings.Contains(strings.ToLower(err.Error()), "token") {
			t.Errorf("expected error containing 'token', got %q", err.Error())
		}
	})

	t.Run("gitlab_missing_token", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.RTMX.Adapters.GitLab.Enabled = true
		cfg.RTMX.Adapters.GitLab.Server = "https://gitlab.com"
		cfg.RTMX.Adapters.GitLab.Project = "group/project"
		t.Setenv("GITLAB_TOKEN", "")

		_, err := getAdapter("gitlab", cfg)
		if err == nil {
			t.Fatal("expected error for missing token, got nil")
		}
		if !strings.Contains(strings.ToLower(err.Error()), "token") {
			t.Errorf("expected error containing 'token', got %q", err.Error())
		}
	})
}
