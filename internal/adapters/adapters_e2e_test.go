package adapters

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// ---------------------------------------------------------------------------
// E2E lifecycle tests -- full create-to-verify sequences
// ---------------------------------------------------------------------------

// TestE2EAsanaLifecycle exercises the full adapter lifecycle:
// create adapter -> TestConnection -> FetchItems -> GetItem -> CreateItem -> UpdateItem -> status roundtrip.
func TestE2EAsanaLifecycle(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-001")

	server := newTestAsanaServer(t)
	defer server.Close()

	cfg := &config.AsanaAdapterConfig{
		Enabled:    true,
		ProjectGID: "proj-1",
		TokenEnv:   "ASANA_TOKEN",
	}
	adapter := mustAsanaAdapter(t, server, cfg)

	// Step 1: verify connection
	ok, msg := adapter.TestConnection()
	if !ok {
		t.Fatalf("TestConnection failed: %s", msg)
	}
	if !strings.Contains(msg, "Test User") {
		t.Fatalf("TestConnection message missing user, got %q", msg)
	}

	// Step 2: fetch all items
	items, err := adapter.FetchItems(nil)
	if err != nil {
		t.Fatalf("FetchItems error: %v", err)
	}
	if len(items) < 1 {
		t.Fatal("FetchItems returned no items")
	}

	// Step 3: get specific item from the fetched list
	firstID := items[0].ExternalID
	item, err := adapter.GetItem(firstID)
	if err != nil {
		t.Fatalf("GetItem(%q) error: %v", firstID, err)
	}
	if item.ExternalID != firstID {
		t.Errorf("GetItem ID = %q, want %q", item.ExternalID, firstID)
	}

	// Step 4: create a new item from a requirement
	req := &database.Requirement{
		ReqID:           "REQ-E2E-001",
		RequirementText: "End-to-end lifecycle test requirement",
		Status:          database.StatusPartial,
		Notes:           "Created during E2E test",
	}
	createdID, err := adapter.CreateItem(req)
	if err != nil {
		t.Fatalf("CreateItem error: %v", err)
	}
	if createdID == "" {
		t.Fatal("CreateItem returned empty ID")
	}

	// Step 5: update the item
	req.Status = database.StatusComplete
	if !adapter.UpdateItem(firstID, req) {
		t.Error("UpdateItem returned false, expected true")
	}

	// Step 6: status mapping roundtrip
	externalStatus := adapter.MapStatusFromRTMX(database.StatusComplete)
	rtmxStatus := adapter.MapStatusToRTMX(externalStatus)
	if rtmxStatus != database.StatusComplete {
		t.Errorf("status roundtrip: COMPLETE -> %q -> %v, want COMPLETE", externalStatus, rtmxStatus)
	}

	externalStatus = adapter.MapStatusFromRTMX(database.StatusPartial)
	rtmxStatus = adapter.MapStatusToRTMX(externalStatus)
	if rtmxStatus != database.StatusPartial {
		t.Errorf("status roundtrip: PARTIAL -> %q -> %v, want PARTIAL", externalStatus, rtmxStatus)
	}

	externalStatus = adapter.MapStatusFromRTMX(database.StatusMissing)
	rtmxStatus = adapter.MapStatusToRTMX(externalStatus)
	if rtmxStatus != database.StatusMissing {
		t.Errorf("status roundtrip: MISSING -> %q -> %v, want MISSING", externalStatus, rtmxStatus)
	}
}

// TestE2EMondayLifecycle exercises the full Monday.com adapter lifecycle.
func TestE2EMondayLifecycle(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-004")

	server := newTestMondayServer(t)
	defer server.Close()

	cfg := &config.MondayAdapterConfig{
		Enabled:  true,
		BoardID:  "board-1",
		TokenEnv: "MONDAY_TOKEN",
	}
	adapter := mustMondayAdapter(t, server, cfg)

	// Step 1: verify connection
	ok, msg := adapter.TestConnection()
	if !ok {
		t.Fatalf("TestConnection failed: %s", msg)
	}

	// Step 2: fetch all items
	items, err := adapter.FetchItems(nil)
	if err != nil {
		t.Fatalf("FetchItems error: %v", err)
	}
	if len(items) < 1 {
		t.Fatal("FetchItems returned no items")
	}

	// Step 3: get specific item
	firstID := items[0].ExternalID
	item, err := adapter.GetItem(firstID)
	if err != nil {
		t.Fatalf("GetItem(%q) error: %v", firstID, err)
	}
	if item.ExternalID != firstID {
		t.Errorf("GetItem ID = %q, want %q", item.ExternalID, firstID)
	}

	// Step 4: create item from requirement
	req := &database.Requirement{
		ReqID:           "REQ-E2E-002",
		RequirementText: "Monday E2E lifecycle test",
		Status:          database.StatusMissing,
	}
	createdID, err := adapter.CreateItem(req)
	if err != nil {
		t.Fatalf("CreateItem error: %v", err)
	}
	if createdID == "" {
		t.Fatal("CreateItem returned empty ID")
	}

	// Step 5: update item
	req.Status = database.StatusComplete
	if !adapter.UpdateItem(firstID, req) {
		t.Error("UpdateItem returned false")
	}

	// Step 6: status mapping roundtrip
	for _, status := range []database.Status{database.StatusComplete, database.StatusPartial, database.StatusMissing} {
		ext := adapter.MapStatusFromRTMX(status)
		back := adapter.MapStatusToRTMX(ext)
		if back != status {
			t.Errorf("Monday status roundtrip: %v -> %q -> %v", status, ext, back)
		}
	}
}

// TestE2EGitLabLifecycle exercises the full GitLab adapter lifecycle.
func TestE2EGitLabLifecycle(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-007")

	now := time.Now()
	callLog := []string{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method
		callLog = append(callLog, method+" "+path)

		switch {
		// TestConnection: GET /api/v4/projects/<project>
		case method == "GET" && !strings.Contains(path, "issues"):
			_ = json.NewEncoder(w).Encode(map[string]string{
				"path_with_namespace": "e2e-group/e2e-project",
			})

		// FetchItems: GET .../issues?...
		case method == "GET" && strings.Contains(path, "issues") && !strings.Contains(path, "issues/"):
			_ = json.NewEncoder(w).Encode([]GitLabIssue{
				{
					IID:         10,
					Title:       "[REQ-LIFE-010] E2E requirement",
					Description: "Lifecycle test\n\n---\nRTMX: REQ-LIFE-010",
					State:       "opened",
					WebURL:      "https://gitlab.com/e2e-group/e2e-project/-/issues/10",
					CreatedAt:   now,
					UpdatedAt:   now,
					Labels:      []string{"p1"},
				},
			})

		// GetItem: GET .../issues/<id>
		case method == "GET" && strings.Contains(path, "issues/"):
			_ = json.NewEncoder(w).Encode(GitLabIssue{
				IID:         10,
				Title:       "[REQ-LIFE-010] E2E requirement",
				Description: "Lifecycle test\n\n---\nRTMX: REQ-LIFE-010",
				State:       "opened",
				WebURL:      "https://gitlab.com/e2e-group/e2e-project/-/issues/10",
				CreatedAt:   now,
				UpdatedAt:   now,
				Assignee: &struct {
					Username string `json:"username"`
				}{Username: "e2e-dev"},
			})

		// CreateItem: POST .../issues
		case method == "POST" && strings.Contains(path, "issues"):
			w.WriteHeader(201)
			_ = json.NewEncoder(w).Encode(GitLabIssue{IID: 42})

		// UpdateItem: PUT .../issues/<id>
		case method == "PUT" && strings.Contains(path, "issues/"):
			_ = json.NewEncoder(w).Encode(GitLabIssue{IID: 42})

		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	cfg := &config.GitLabAdapterConfig{
		Enabled: true,
		Server:  server.URL,
		Project: "e2e-group/e2e-project",
	}
	adapter, err := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "glpat-e2e" }))
	if err != nil {
		t.Fatalf("NewGitLabAdapter error: %v", err)
	}

	// Step 1: connection
	ok, msg := adapter.TestConnection()
	if !ok {
		t.Fatalf("TestConnection failed: %s", msg)
	}
	if !strings.Contains(msg, "e2e-group/e2e-project") {
		t.Errorf("TestConnection message = %q, missing project path", msg)
	}

	// Step 2: fetch
	items, err := adapter.FetchItems(nil)
	if err != nil {
		t.Fatalf("FetchItems error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("FetchItems returned %d items, want 1", len(items))
	}
	if items[0].RequirementID != "REQ-LIFE-010" {
		t.Errorf("FetchItems[0].RequirementID = %q, want REQ-LIFE-010", items[0].RequirementID)
	}

	// Step 3: get item
	item, err := adapter.GetItem("10")
	if err != nil {
		t.Fatalf("GetItem error: %v", err)
	}
	if item.Assignee != "e2e-dev" {
		t.Errorf("GetItem Assignee = %q, want e2e-dev", item.Assignee)
	}

	// Step 4: create item
	req := &database.Requirement{
		ReqID:           "REQ-LIFE-011",
		RequirementText: "GitLab E2E lifecycle test",
		Notes:           "Created during E2E",
		Status:          database.StatusMissing,
	}
	createdID, err := adapter.CreateItem(req)
	if err != nil {
		t.Fatalf("CreateItem error: %v", err)
	}
	if createdID != "42" {
		t.Errorf("CreateItem returned %q, want 42", createdID)
	}

	// Step 5: update item
	req.Status = database.StatusComplete
	if !adapter.UpdateItem(createdID, req) {
		t.Error("UpdateItem returned false")
	}

	// Step 6: status mapping roundtrip
	if adapter.MapStatusToRTMX("closed") != database.StatusComplete {
		t.Error("closed should map to COMPLETE")
	}
	if adapter.MapStatusFromRTMX(database.StatusComplete) != "close" {
		t.Error("COMPLETE should map to close")
	}

	// Step 7: milestone roundtrip
	version := adapter.MilestoneToVersion("1.2.3")
	milestone := adapter.VersionToMilestone(version)
	if milestone != "1.2.3" {
		t.Errorf("milestone roundtrip: 1.2.3 -> %q -> %q", version, milestone)
	}

	// Verify the server received all expected API calls
	if len(callLog) < 5 {
		t.Errorf("expected at least 5 API calls, got %d: %v", len(callLog), callLog)
	}
}

// ---------------------------------------------------------------------------
// Error scenario tests
// ---------------------------------------------------------------------------

// TestErrorAsanaAuthFailure tests Asana adapter behavior with missing/invalid tokens.
func TestErrorAsanaAuthFailure(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-001")

	t.Run("missing_token_prevents_creation", func(t *testing.T) {
		cfg := &config.AsanaAdapterConfig{
			Enabled:    true,
			ProjectGID: "proj-1",
			TokenEnv:   "ASANA_TOKEN",
		}
		_, err := NewAsanaAdapter(cfg, WithEnvGetter(func(string) string { return "" }))
		if err == nil {
			t.Error("expected error when token is empty")
		}
		if !strings.Contains(err.Error(), "token not found") {
			t.Errorf("error = %q, want 'token not found'", err.Error())
		}
	})

	t.Run("invalid_token_fails_connection", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer valid-token" {
				w.WriteHeader(401)
				return
			}
		}))
		defer server.Close()

		cfg := &config.AsanaAdapterConfig{
			Enabled:    true,
			ProjectGID: "proj-1",
			TokenEnv:   "ASANA_TOKEN",
		}
		adapter, _ := NewAsanaAdapter(cfg,
			WithHTTPClient(server.Client()),
			WithEnvGetter(func(string) string { return "wrong-token" }),
		)
		adapter.SetBaseURL(server.URL)

		ok, msg := adapter.TestConnection()
		if ok {
			t.Error("expected connection to fail with wrong token")
		}
		if !strings.Contains(msg, "401") {
			t.Errorf("message = %q, should contain 401", msg)
		}
	})

	t.Run("invalid_token_fails_fetch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(401)
		}))
		defer server.Close()

		cfg := &config.AsanaAdapterConfig{
			Enabled:    true,
			ProjectGID: "proj-1",
			TokenEnv:   "ASANA_TOKEN",
		}
		adapter, _ := NewAsanaAdapter(cfg,
			WithHTTPClient(server.Client()),
			WithEnvGetter(func(string) string { return "bad" }),
		)
		adapter.SetBaseURL(server.URL)

		_, err := adapter.FetchItems(nil)
		if err == nil {
			t.Error("expected error on 401")
		}
	})
}

// TestErrorMondayAuthFailure tests Monday adapter behavior with auth issues.
func TestErrorMondayAuthFailure(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-004")

	t.Run("missing_token_prevents_creation", func(t *testing.T) {
		cfg := &config.MondayAdapterConfig{
			Enabled:  true,
			BoardID:  "board-1",
			TokenEnv: "MONDAY_TOKEN",
		}
		_, err := NewMondayAdapter(cfg, WithEnvGetter(func(string) string { return "" }))
		if err == nil {
			t.Error("expected error when token is empty")
		}
	})

	t.Run("invalid_token_fails_connection", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(401)
		}))
		defer server.Close()

		cfg := &config.MondayAdapterConfig{
			Enabled:  true,
			BoardID:  "board-1",
			TokenEnv: "MONDAY_TOKEN",
		}
		adapter, _ := NewMondayAdapter(cfg,
			WithHTTPClient(server.Client()),
			WithEnvGetter(func(string) string { return "bad-token" }),
		)
		adapter.SetAPIURL(server.URL)

		ok, msg := adapter.TestConnection()
		if ok {
			t.Error("expected connection to fail")
		}
		if !strings.Contains(msg, "401") {
			t.Errorf("message = %q, should contain 401", msg)
		}
	})
}

// TestErrorGitLabAuthFailure tests GitLab adapter behavior with auth issues.
func TestErrorGitLabAuthFailure(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-007")

	t.Run("missing_token_prevents_creation", func(t *testing.T) {
		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Project: "g/p",
		}
		_, err := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "" }))
		if err == nil {
			t.Error("expected error when token is empty")
		}
	})

	t.Run("invalid_token_fails_connection", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(401)
		}))
		defer server.Close()

		cfg := &config.GitLabAdapterConfig{
			Enabled: true,
			Server:  server.URL,
			Project: "g/p",
		}
		adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "bad" }))

		ok, msg := adapter.TestConnection()
		if ok {
			t.Error("expected connection to fail")
		}
		if !strings.Contains(msg, "401") {
			t.Errorf("message = %q, should contain 401", msg)
		}
	})
}

// TestErrorMalformedJSON tests adapter behavior when the server returns invalid JSON.
func TestErrorMalformedJSON(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-001")

	malformedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("this is not valid json{{{"))
	}))
	defer malformedServer.Close()

	t.Run("asana_malformed_connection", func(t *testing.T) {
		cfg := &config.AsanaAdapterConfig{Enabled: true, ProjectGID: "p", TokenEnv: "T"}
		a, _ := NewAsanaAdapter(cfg, WithHTTPClient(malformedServer.Client()), WithEnvGetter(func(string) string { return "tok" }))
		a.SetBaseURL(malformedServer.URL)

		ok, msg := a.TestConnection()
		if ok {
			t.Error("expected failure on malformed JSON")
		}
		if !strings.Contains(msg, "parse") {
			t.Errorf("message = %q, should mention parse failure", msg)
		}
	})

	t.Run("asana_malformed_fetch", func(t *testing.T) {
		cfg := &config.AsanaAdapterConfig{Enabled: true, ProjectGID: "p", TokenEnv: "T"}
		a, _ := NewAsanaAdapter(cfg, WithHTTPClient(malformedServer.Client()), WithEnvGetter(func(string) string { return "tok" }))
		a.SetBaseURL(malformedServer.URL)

		_, err := a.FetchItems(nil)
		if err == nil {
			t.Error("expected error on malformed JSON")
		}
	})

	t.Run("asana_malformed_getitem", func(t *testing.T) {
		cfg := &config.AsanaAdapterConfig{Enabled: true, ProjectGID: "p", TokenEnv: "T"}
		a, _ := NewAsanaAdapter(cfg, WithHTTPClient(malformedServer.Client()), WithEnvGetter(func(string) string { return "tok" }))
		a.SetBaseURL(malformedServer.URL)

		_, err := a.GetItem("123")
		if err == nil {
			t.Error("expected error on malformed JSON")
		}
	})

	t.Run("asana_malformed_create", func(t *testing.T) {
		// CreateItem expects 201, but malformed server returns 200 with bad JSON
		// We need a server that returns 201 with bad JSON
		badCreateServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(201)
			_, _ = w.Write([]byte("not json"))
		}))
		defer badCreateServer.Close()

		cfg := &config.AsanaAdapterConfig{Enabled: true, ProjectGID: "p", TokenEnv: "T"}
		a, _ := NewAsanaAdapter(cfg, WithHTTPClient(badCreateServer.Client()), WithEnvGetter(func(string) string { return "tok" }))
		a.SetBaseURL(badCreateServer.URL)

		_, err := a.CreateItem(&database.Requirement{ReqID: "REQ-X", RequirementText: "test"})
		if err == nil {
			t.Error("expected error on malformed JSON")
		}
	})

	t.Run("monday_malformed_fetch", func(t *testing.T) {
		cfg := &config.MondayAdapterConfig{Enabled: true, BoardID: "b", TokenEnv: "T"}
		m, _ := NewMondayAdapter(cfg, WithHTTPClient(malformedServer.Client()), WithEnvGetter(func(string) string { return "tok" }))
		m.SetAPIURL(malformedServer.URL)

		// Monday graphQL method returns error if decode fails
		_, err := m.FetchItems(nil)
		if err == nil {
			t.Error("expected error on malformed JSON")
		}
	})

	t.Run("monday_malformed_getitem", func(t *testing.T) {
		cfg := &config.MondayAdapterConfig{Enabled: true, BoardID: "b", TokenEnv: "T"}
		m, _ := NewMondayAdapter(cfg, WithHTTPClient(malformedServer.Client()), WithEnvGetter(func(string) string { return "tok" }))
		m.SetAPIURL(malformedServer.URL)

		_, err := m.GetItem("123")
		if err == nil {
			t.Error("expected error on malformed JSON")
		}
	})

	t.Run("gitlab_malformed_fetch", func(t *testing.T) {
		cfg := &config.GitLabAdapterConfig{Enabled: true, Server: malformedServer.URL, Project: "g/p"}
		g, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		_, err := g.FetchItems(nil)
		if err == nil {
			t.Error("expected error on malformed JSON")
		}
	})

	t.Run("gitlab_malformed_getitem", func(t *testing.T) {
		cfg := &config.GitLabAdapterConfig{Enabled: true, Server: malformedServer.URL, Project: "g/p"}
		g, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		_, err := g.GetItem("1")
		if err == nil {
			t.Error("expected error on malformed JSON")
		}
	})

	t.Run("gitlab_malformed_create", func(t *testing.T) {
		badCreateServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(201)
			_, _ = w.Write([]byte("not json"))
		}))
		defer badCreateServer.Close()

		cfg := &config.GitLabAdapterConfig{Enabled: true, Server: badCreateServer.URL, Project: "g/p"}
		g, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		_, err := g.CreateItem(&database.Requirement{ReqID: "REQ-X", RequirementText: "test"})
		if err == nil {
			t.Error("expected error on malformed JSON for create")
		}
	})
}

// TestErrorHTTPStatusCodes tests adapter behavior with various HTTP error codes.
func TestErrorHTTPStatusCodes(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-007")

	statusCodes := []struct {
		code int
		name string
	}{
		{500, "internal_server_error"},
		{404, "not_found"},
		{429, "rate_limit"},
		{403, "forbidden"},
	}

	for _, sc := range statusCodes {
		sc := sc
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(sc.code)
		}))

		t.Run(fmt.Sprintf("asana_fetch_%s", sc.name), func(t *testing.T) {
			cfg := &config.AsanaAdapterConfig{Enabled: true, ProjectGID: "p", TokenEnv: "T"}
			a, _ := NewAsanaAdapter(cfg, WithHTTPClient(server.Client()), WithEnvGetter(func(string) string { return "tok" }))
			a.SetBaseURL(server.URL)

			_, err := a.FetchItems(nil)
			if err == nil {
				t.Errorf("expected error on HTTP %d", sc.code)
			}
			if !strings.Contains(err.Error(), fmt.Sprintf("%d", sc.code)) {
				t.Errorf("error = %q, should contain %d", err.Error(), sc.code)
			}
		})

		t.Run(fmt.Sprintf("asana_create_%s", sc.name), func(t *testing.T) {
			cfg := &config.AsanaAdapterConfig{Enabled: true, ProjectGID: "p", TokenEnv: "T"}
			a, _ := NewAsanaAdapter(cfg, WithHTTPClient(server.Client()), WithEnvGetter(func(string) string { return "tok" }))
			a.SetBaseURL(server.URL)

			_, err := a.CreateItem(&database.Requirement{ReqID: "REQ-X", RequirementText: "test"})
			if err == nil {
				t.Errorf("expected error on HTTP %d", sc.code)
			}
		})

		t.Run(fmt.Sprintf("asana_update_%s", sc.name), func(t *testing.T) {
			cfg := &config.AsanaAdapterConfig{Enabled: true, ProjectGID: "p", TokenEnv: "T"}
			a, _ := NewAsanaAdapter(cfg, WithHTTPClient(server.Client()), WithEnvGetter(func(string) string { return "tok" }))
			a.SetBaseURL(server.URL)

			ok := a.UpdateItem("123", &database.Requirement{ReqID: "REQ-X", RequirementText: "test"})
			if ok {
				t.Errorf("UpdateItem should return false on HTTP %d", sc.code)
			}
		})

		t.Run(fmt.Sprintf("monday_fetch_%s", sc.name), func(t *testing.T) {
			cfg := &config.MondayAdapterConfig{Enabled: true, BoardID: "b", TokenEnv: "T"}
			m, _ := NewMondayAdapter(cfg, WithHTTPClient(server.Client()), WithEnvGetter(func(string) string { return "tok" }))
			m.SetAPIURL(server.URL)

			_, err := m.FetchItems(nil)
			if err == nil {
				t.Errorf("expected error on HTTP %d", sc.code)
			}
		})

		t.Run(fmt.Sprintf("monday_create_%s", sc.name), func(t *testing.T) {
			cfg := &config.MondayAdapterConfig{Enabled: true, BoardID: "b", TokenEnv: "T"}
			m, _ := NewMondayAdapter(cfg, WithHTTPClient(server.Client()), WithEnvGetter(func(string) string { return "tok" }))
			m.SetAPIURL(server.URL)

			_, err := m.CreateItem(&database.Requirement{ReqID: "REQ-X", RequirementText: "test"})
			if err == nil {
				t.Errorf("expected error on HTTP %d", sc.code)
			}
		})

		t.Run(fmt.Sprintf("gitlab_fetch_%s", sc.name), func(t *testing.T) {
			cfg := &config.GitLabAdapterConfig{Enabled: true, Server: server.URL, Project: "g/p"}
			g, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

			_, err := g.FetchItems(nil)
			if err == nil {
				t.Errorf("expected error on HTTP %d", sc.code)
			}
		})

		t.Run(fmt.Sprintf("gitlab_create_%s", sc.name), func(t *testing.T) {
			cfg := &config.GitLabAdapterConfig{Enabled: true, Server: server.URL, Project: "g/p"}
			g, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

			_, err := g.CreateItem(&database.Requirement{ReqID: "REQ-X", RequirementText: "test"})
			if err == nil {
				t.Errorf("expected error on HTTP %d", sc.code)
			}
		})

		t.Run(fmt.Sprintf("gitlab_update_%s", sc.name), func(t *testing.T) {
			cfg := &config.GitLabAdapterConfig{Enabled: true, Server: server.URL, Project: "g/p"}
			g, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

			ok := g.UpdateItem("1", &database.Requirement{ReqID: "REQ-X", RequirementText: "test", Status: database.StatusComplete})
			if ok {
				t.Errorf("UpdateItem should return false on HTTP %d", sc.code)
			}
		})

		server.Close()
	}
}

// TestErrorNetworkTimeout tests adapter behavior when requests time out.
func TestErrorNetworkTimeout(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-001")

	// Server that never responds (blocks until context deadline)
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
			return
		case <-time.After(30 * time.Second):
			return
		}
	}))
	defer slowServer.Close()

	// Use a client with a very short timeout for this test
	shortTimeoutClient := &http.Client{Timeout: 50 * time.Millisecond}

	t.Run("asana_timeout", func(t *testing.T) {
		cfg := &config.AsanaAdapterConfig{Enabled: true, ProjectGID: "p", TokenEnv: "T"}
		a, _ := NewAsanaAdapter(cfg, WithHTTPClient(shortTimeoutClient), WithEnvGetter(func(string) string { return "tok" }))
		a.SetBaseURL(slowServer.URL)

		ok, _ := a.TestConnection()
		if ok {
			t.Error("expected connection to fail on timeout")
		}
	})

	t.Run("monday_timeout", func(t *testing.T) {
		cfg := &config.MondayAdapterConfig{Enabled: true, BoardID: "b", TokenEnv: "T"}
		m, _ := NewMondayAdapter(cfg, WithHTTPClient(shortTimeoutClient), WithEnvGetter(func(string) string { return "tok" }))
		m.SetAPIURL(slowServer.URL)

		ok, _ := m.TestConnection()
		if ok {
			t.Error("expected connection to fail on timeout")
		}
	})

	t.Run("gitlab_timeout", func(t *testing.T) {
		cfg := &config.GitLabAdapterConfig{Enabled: true, Server: slowServer.URL, Project: "g/p"}
		g, _ := NewGitLabAdapter(cfg,
			WithHTTPClient(shortTimeoutClient),
			WithEnvGetter(func(string) string { return "tok" }),
		)

		ok, _ := g.TestConnection()
		if ok {
			t.Error("expected connection to fail on timeout")
		}
	})
}

// TestErrorEmptyResponseBody tests adapter behavior when server returns empty bodies.
func TestErrorEmptyResponseBody(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-004")

	emptyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		// Return empty body (no JSON)
	}))
	defer emptyServer.Close()

	t.Run("asana_empty_connection", func(t *testing.T) {
		cfg := &config.AsanaAdapterConfig{Enabled: true, ProjectGID: "p", TokenEnv: "T"}
		a, _ := NewAsanaAdapter(cfg, WithHTTPClient(emptyServer.Client()), WithEnvGetter(func(string) string { return "tok" }))
		a.SetBaseURL(emptyServer.URL)

		ok, _ := a.TestConnection()
		if ok {
			t.Error("expected failure on empty response")
		}
	})

	t.Run("asana_empty_fetch", func(t *testing.T) {
		cfg := &config.AsanaAdapterConfig{Enabled: true, ProjectGID: "p", TokenEnv: "T"}
		a, _ := NewAsanaAdapter(cfg, WithHTTPClient(emptyServer.Client()), WithEnvGetter(func(string) string { return "tok" }))
		a.SetBaseURL(emptyServer.URL)

		_, err := a.FetchItems(nil)
		if err == nil {
			t.Error("expected error on empty response")
		}
	})

	t.Run("monday_empty_fetch", func(t *testing.T) {
		cfg := &config.MondayAdapterConfig{Enabled: true, BoardID: "b", TokenEnv: "T"}
		m, _ := NewMondayAdapter(cfg, WithHTTPClient(emptyServer.Client()), WithEnvGetter(func(string) string { return "tok" }))
		m.SetAPIURL(emptyServer.URL)

		_, err := m.FetchItems(nil)
		if err == nil {
			t.Error("expected error on empty response")
		}
	})

	t.Run("gitlab_empty_fetch", func(t *testing.T) {
		cfg := &config.GitLabAdapterConfig{Enabled: true, Server: emptyServer.URL, Project: "g/p"}
		g, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		_, err := g.FetchItems(nil)
		if err == nil {
			t.Error("expected error on empty response")
		}
	})
}

// TestErrorMissingRequiredFields tests adapter behavior when responses lack required fields.
func TestErrorMissingRequiredFields(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-007")

	t.Run("asana_create_missing_gid", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(201)
			// Return valid JSON but without gid field
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{},
			})
		}))
		defer server.Close()

		cfg := &config.AsanaAdapterConfig{Enabled: true, ProjectGID: "p", TokenEnv: "T"}
		a, _ := NewAsanaAdapter(cfg, WithHTTPClient(server.Client()), WithEnvGetter(func(string) string { return "tok" }))
		a.SetBaseURL(server.URL)

		id, err := a.CreateItem(&database.Requirement{ReqID: "REQ-X", RequirementText: "test"})
		// The adapter does not validate for empty GID, so it returns "" without error
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if id != "" {
			t.Errorf("expected empty ID from missing GID, got %q", id)
		}
	})

	t.Run("monday_getitem_empty_items", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"items": []interface{}{},
				},
			})
		}))
		defer server.Close()

		cfg := &config.MondayAdapterConfig{Enabled: true, BoardID: "b", TokenEnv: "T"}
		m, _ := NewMondayAdapter(cfg, WithHTTPClient(server.Client()), WithEnvGetter(func(string) string { return "tok" }))
		m.SetAPIURL(server.URL)

		_, err := m.GetItem("999")
		if err == nil {
			t.Error("expected error when item not found")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("error = %q, should contain 'not found'", err.Error())
		}
	})

	t.Run("gitlab_create_missing_iid", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(201)
			// Return valid JSON but IID defaults to 0
			_ = json.NewEncoder(w).Encode(GitLabIssue{})
		}))
		defer server.Close()

		cfg := &config.GitLabAdapterConfig{Enabled: true, Server: server.URL, Project: "g/p"}
		g, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

		id, err := g.CreateItem(&database.Requirement{ReqID: "REQ-X", RequirementText: "test"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		// IID defaults to 0 which formats as "0"
		if id != "0" {
			t.Errorf("expected ID '0' from zero IID, got %q", id)
		}
	})
}

// ---------------------------------------------------------------------------
// Bidirectional roundtrip fidelity tests
// ---------------------------------------------------------------------------

// TestRoundtripAsanaFidelity verifies that a requirement created via CreateItem
// is faithfully represented when later fetched via FetchItems.
func TestRoundtripAsanaFidelity(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-002")

	// Stateful server that records the created task and returns it in fetches.
	var createdTask map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/tasks"):
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			data, _ := body["data"].(map[string]interface{})
			createdTask = map[string]interface{}{
				"gid":           "300",
				"name":          data["name"],
				"notes":         data["notes"],
				"completed":     data["completed"],
				"created_at":    "2026-06-01T00:00:00Z",
				"modified_at":   "2026-06-01T00:00:00Z",
				"permalink_url": "https://app.asana.com/0/proj-1/300",
				"memberships": []map[string]interface{}{
					{"project": map[string]interface{}{"gid": "proj-1"}, "section": map[string]interface{}{"gid": "s1", "name": "Testing"}},
				},
			}
			w.WriteHeader(201)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"gid": "300"}})

		case r.Method == "GET" && strings.Contains(r.URL.Path, "/projects/") && strings.Contains(r.URL.Path, "/tasks"):
			items := []interface{}{}
			if createdTask != nil {
				items = append(items, createdTask)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"data": items})

		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	cfg := &config.AsanaAdapterConfig{Enabled: true, ProjectGID: "proj-1", TokenEnv: "T"}
	adapter, _ := NewAsanaAdapter(cfg, WithHTTPClient(server.Client()), WithEnvGetter(func(string) string { return "tok" }))
	adapter.SetBaseURL(server.URL)

	// Create a requirement and push it to the adapter
	req := &database.Requirement{
		ReqID:           "REQ-ROUND-001",
		RequirementText: "Roundtrip fidelity test requirement",
		Status:          database.StatusComplete,
	}
	gid, err := adapter.CreateItem(req)
	if err != nil {
		t.Fatalf("CreateItem error: %v", err)
	}
	if gid != "300" {
		t.Fatalf("created GID = %q, want 300", gid)
	}

	// Now fetch and verify the item preserves the requirement ID
	items, err := adapter.FetchItems(nil)
	if err != nil {
		t.Fatalf("FetchItems error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].RequirementID != "REQ-ROUND-001" {
		t.Errorf("RequirementID = %q, want REQ-ROUND-001", items[0].RequirementID)
	}
	if items[0].ExternalID != "300" {
		t.Errorf("ExternalID = %q, want 300", items[0].ExternalID)
	}
	// Status mapping: completed task -> COMPLETE
	mapped := adapter.MapStatusToRTMX(items[0].Status)
	if mapped != database.StatusComplete {
		t.Errorf("status roundtrip: created as COMPLETE, fetched status=%q maps to %v", items[0].Status, mapped)
	}
}

// TestRoundtripMondayFidelity verifies Monday.com create-then-fetch roundtrip.
func TestRoundtripMondayFidelity(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-005")

	var createdItemName string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}
		var body struct {
			Query string `json:"query"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		q := body.Query

		switch {
		case strings.Contains(q, "create_item"):
			// Extract name from mutation for later retrieval
			if idx := strings.Index(q, "item_name:"); idx >= 0 {
				rest := q[idx+10:]
				if start := strings.Index(rest, "\""); start >= 0 {
					end := strings.Index(rest[start+1:], "\"")
					if end >= 0 {
						createdItemName = rest[start+1 : start+1+end]
					}
				}
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"create_item": map[string]interface{}{"id": "5001"},
				},
			})

		case strings.Contains(q, "items(ids:"):
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"items": []map[string]interface{}{
						{
							"id":           "5001",
							"name":         createdItemName,
							"state":        "active",
							"created_at":   "2026-06-01T00:00:00Z",
							"updated_at":   "2026-06-01T00:00:00Z",
							"group":        map[string]interface{}{"id": "g1", "title": "Testing"},
							"column_values": []map[string]interface{}{{"id": "status", "title": "Status", "text": "Done"}},
						},
					},
				},
			})

		case strings.Contains(q, "boards(ids:"):
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"boards": []map[string]interface{}{
						{
							"items_page": map[string]interface{}{
								"items": []map[string]interface{}{
									{
										"id":           "5001",
										"name":         createdItemName,
										"state":        "active",
										"created_at":   "2026-06-01T00:00:00Z",
										"updated_at":   "2026-06-01T00:00:00Z",
										"group":        map[string]interface{}{"id": "g1", "title": "Testing"},
										"column_values": []map[string]interface{}{{"id": "status", "title": "Status", "text": "Done"}},
									},
								},
							},
						},
					},
				},
			})

		default:
			w.WriteHeader(400)
		}
	}))
	defer server.Close()

	cfg := &config.MondayAdapterConfig{Enabled: true, BoardID: "board-1", TokenEnv: "T"}
	adapter, _ := NewMondayAdapter(cfg, WithHTTPClient(server.Client()), WithEnvGetter(func(string) string { return "tok" }))
	adapter.SetAPIURL(server.URL)

	req := &database.Requirement{
		ReqID:           "REQ-ROUND-002",
		RequirementText: "Monday roundtrip fidelity",
		Status:          database.StatusComplete,
	}
	id, err := adapter.CreateItem(req)
	if err != nil {
		t.Fatalf("CreateItem error: %v", err)
	}
	if id != "5001" {
		t.Fatalf("created ID = %q, want 5001", id)
	}

	// Fetch via board items
	items, err := adapter.FetchItems(nil)
	if err != nil {
		t.Fatalf("FetchItems error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].ExternalID != "5001" {
		t.Errorf("ExternalID = %q, want 5001", items[0].ExternalID)
	}

	// Verify the title contains the requirement ID
	if !strings.Contains(items[0].Title, "REQ-ROUND-002") {
		t.Errorf("Title = %q, should contain REQ-ROUND-002", items[0].Title)
	}

	// Get item by ID and verify
	item, err := adapter.GetItem("5001")
	if err != nil {
		t.Fatalf("GetItem error: %v", err)
	}
	if item.ExternalID != "5001" {
		t.Errorf("GetItem ExternalID = %q, want 5001", item.ExternalID)
	}

	// Status mapping: "Done" -> COMPLETE
	mapped := adapter.MapStatusToRTMX(items[0].Status)
	if mapped != database.StatusComplete {
		t.Errorf("status roundtrip: created as COMPLETE, fetched status=%q maps to %v", items[0].Status, mapped)
	}
}

// TestRoundtripGitLabFidelity verifies GitLab create-then-fetch roundtrip
// including MilestoneToVersion/VersionToMilestone bidirectional mapping.
func TestRoundtripGitLabFidelity(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-008")

	now := time.Now()
	var createdPayload map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		method := r.Method

		switch {
		// CreateItem
		case method == "POST" && strings.Contains(path, "issues"):
			_ = json.NewDecoder(r.Body).Decode(&createdPayload)
			w.WriteHeader(201)
			_ = json.NewEncoder(w).Encode(GitLabIssue{
				IID:         77,
				Title:       createdPayload["title"].(string),
				Description: createdPayload["description"].(string),
				State:       "opened",
				WebURL:      "https://gitlab.com/g/p/-/issues/77",
				CreatedAt:   now,
				UpdatedAt:   now,
				Milestone:   &GitLabMilestone{ID: 1, Title: "1.0.0"},
			})

		// FetchItems
		case method == "GET" && strings.Contains(path, "issues") && !strings.Contains(path, "issues/"):
			title := "[REQ-ROUND-003] GitLab roundtrip fidelity"
			desc := "GitLab roundtrip fidelity\n\n---\nRTMX: REQ-ROUND-003"
			if createdPayload != nil {
				if t, ok := createdPayload["title"].(string); ok {
					title = t
				}
				if d, ok := createdPayload["description"].(string); ok {
					desc = d
				}
			}
			_ = json.NewEncoder(w).Encode([]GitLabIssue{
				{
					IID:         77,
					Title:       title,
					Description: desc,
					State:       "opened",
					WebURL:      "https://gitlab.com/g/p/-/issues/77",
					CreatedAt:   now,
					UpdatedAt:   now,
					Labels:      []string{"p2"},
					Milestone:   &GitLabMilestone{ID: 1, Title: "1.0.0"},
				},
			})

		// GetItem
		case method == "GET" && strings.Contains(path, "issues/"):
			_ = json.NewEncoder(w).Encode(GitLabIssue{
				IID:         77,
				Title:       "[REQ-ROUND-003] GitLab roundtrip fidelity",
				Description: "RTMX: REQ-ROUND-003",
				State:       "opened",
				WebURL:      "https://gitlab.com/g/p/-/issues/77",
				CreatedAt:   now,
				UpdatedAt:   now,
				Milestone:   &GitLabMilestone{ID: 1, Title: "1.0.0"},
			})

		// UpdateItem
		case method == "PUT" && strings.Contains(path, "issues/"):
			_ = json.NewEncoder(w).Encode(GitLabIssue{IID: 77, State: "closed"})

		default:
			w.WriteHeader(404)
		}
	}))
	defer server.Close()

	cfg := &config.GitLabAdapterConfig{Enabled: true, Server: server.URL, Project: "g/p"}
	adapter, _ := NewGitLabAdapter(cfg, WithEnvGetter(func(string) string { return "tok" }))

	// Create
	req := &database.Requirement{
		ReqID:           "REQ-ROUND-003",
		RequirementText: "GitLab roundtrip fidelity",
		Status:          database.StatusMissing,
		Notes:           "Roundtrip test",
	}
	id, err := adapter.CreateItem(req)
	if err != nil {
		t.Fatalf("CreateItem error: %v", err)
	}
	if id != "77" {
		t.Fatalf("created ID = %q, want 77", id)
	}

	// Fetch and verify requirement ID extraction
	items, err := adapter.FetchItems(nil)
	if err != nil {
		t.Fatalf("FetchItems error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].RequirementID != "REQ-ROUND-003" {
		t.Errorf("RequirementID = %q, want REQ-ROUND-003", items[0].RequirementID)
	}
	if items[0].ExternalID != "77" {
		t.Errorf("ExternalID = %q, want 77", items[0].ExternalID)
	}

	// Priority extracted from label
	if items[0].Priority != "MEDIUM" {
		t.Errorf("Priority = %q, want MEDIUM (from p2 label)", items[0].Priority)
	}

	// Update to complete and verify
	req.Status = database.StatusComplete
	ok := adapter.UpdateItem(id, req)
	if !ok {
		t.Error("UpdateItem returned false")
	}

	// MilestoneToVersion / VersionToMilestone roundtrip
	versions := []string{"1.0.0", "v2.3.4", "0.1"}
	for _, v := range versions {
		milestone := adapter.VersionToMilestone(v)
		versionBack := adapter.MilestoneToVersion(milestone)
		milestoneBack := adapter.VersionToMilestone(versionBack)
		if milestoneBack != milestone {
			t.Errorf("milestone roundtrip failed: %q -> %q -> %q -> %q", v, milestone, versionBack, milestoneBack)
		}
	}

	// Specific roundtrip: semver without v prefix
	version := adapter.MilestoneToVersion("2.0.0")
	if version != "v2.0.0" {
		t.Errorf("MilestoneToVersion(2.0.0) = %q, want v2.0.0", version)
	}
	back := adapter.VersionToMilestone(version)
	if back != "2.0.0" {
		t.Errorf("VersionToMilestone(v2.0.0) = %q, want 2.0.0", back)
	}
}
