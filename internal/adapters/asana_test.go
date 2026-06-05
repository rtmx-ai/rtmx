package adapters

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func newTestAsanaServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/users/me", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(401)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{"gid": "1", "name": "Test User"},
		})
	})

	mux.HandleFunc("/projects/proj-1/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(401)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"gid": "100", "name": "[REQ-CLI-001] Build CLI", "notes": "Build CLI framework\n\nRTMX: REQ-CLI-001",
					"completed": true, "created_at": "2026-01-01T00:00:00Z", "modified_at": "2026-01-02T00:00:00Z",
					"permalink_url": "https://app.asana.com/0/proj-1/100",
					"assignee":      map[string]interface{}{"gid": "u1", "name": "Alice"},
					"memberships": []map[string]interface{}{
						{"project": map[string]interface{}{"gid": "proj-1", "name": "RTMX"}, "section": map[string]interface{}{"gid": "s1", "name": "CLI Tasks"}},
					},
				},
				{
					"gid": "101", "name": "[REQ-MCP-001] MCP server", "notes": "MCP server\n\nRTMX: REQ-MCP-001",
					"completed": false, "created_at": "2026-01-03T00:00:00Z", "modified_at": "2026-01-04T00:00:00Z",
					"memberships": []map[string]interface{}{
						{"project": map[string]interface{}{"gid": "proj-1", "name": "RTMX"}, "section": map[string]interface{}{"gid": "s2", "name": "MCP Tasks"}},
					},
				},
			},
		})
	})

	mux.HandleFunc("/tasks/100", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"gid": "100", "name": "[REQ-CLI-001] Build CLI", "notes": "Build CLI framework\n\nRTMX: REQ-CLI-001",
					"completed": true, "created_at": "2026-01-01T00:00:00Z", "modified_at": "2026-01-02T00:00:00Z",
					"assignee":  map[string]interface{}{"gid": "u1", "name": "Alice"},
					"memberships": []map[string]interface{}{
						{"project": map[string]interface{}{"gid": "proj-1"}, "section": map[string]interface{}{"gid": "s1", "name": "CLI Tasks"}},
					},
				},
			})
		} else if r.Method == "PUT" {
			json.NewEncoder(w).Encode(map[string]interface{}{"data": map[string]interface{}{"gid": "100"}})
		}
	})

	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			w.WriteHeader(201)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{"gid": "200"},
			})
		}
	})

	return httptest.NewServer(mux)
}

// TestAsanaAdapter validates the Asana REST API adapter.
// REQ-ADAPT-001: Asana REST API adapter with PAT/OAuth2 auth and ServiceAdapter interface.
func TestAsanaAdapter(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-001")

	server := newTestAsanaServer(t)
	defer server.Close()

	cfg := &config.AsanaAdapterConfig{
		Enabled:      true,
		WorkspaceGID: "ws-1",
		ProjectGID:   "proj-1",
		TokenEnv:     "ASANA_TOKEN",
	}

	t.Run("creation_requires_enabled", func(t *testing.T) {
		_, err := NewAsanaAdapter(&config.AsanaAdapterConfig{Enabled: false}, WithEnvGetter(func(string) string { return "x" }))
		if err == nil {
			t.Error("should fail when not enabled")
		}
	})

	t.Run("creation_requires_token", func(t *testing.T) {
		_, err := NewAsanaAdapter(cfg, WithEnvGetter(func(string) string { return "" }))
		if err == nil {
			t.Error("should fail without token")
		}
	})

	t.Run("name_returns_asana", func(t *testing.T) {
		a := mustAsanaAdapter(t, server, cfg)
		if a.Name() != "asana" {
			t.Errorf("Name() = %q, want asana", a.Name())
		}
	})

	t.Run("is_configured", func(t *testing.T) {
		a := mustAsanaAdapter(t, server, cfg)
		if !a.IsConfigured() {
			t.Error("should be configured")
		}
	})

	t.Run("test_connection", func(t *testing.T) {
		a := mustAsanaAdapter(t, server, cfg)
		ok, msg := a.TestConnection()
		if !ok {
			t.Errorf("TestConnection failed: %s", msg)
		}
		if !strings.Contains(msg, "Test User") {
			t.Errorf("message should contain user name, got %q", msg)
		}
	})

	t.Run("fetch_items", func(t *testing.T) {
		a := mustAsanaAdapter(t, server, cfg)
		items, err := a.FetchItems(nil)
		if err != nil {
			t.Fatalf("FetchItems error: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("got %d items, want 2", len(items))
		}
		if items[0].ExternalID != "100" {
			t.Errorf("first item ID = %q, want 100", items[0].ExternalID)
		}
		if items[0].RequirementID != "REQ-CLI-001" {
			t.Errorf("first item reqID = %q, want REQ-CLI-001", items[0].RequirementID)
		}
	})

	t.Run("get_item", func(t *testing.T) {
		a := mustAsanaAdapter(t, server, cfg)
		item, err := a.GetItem("100")
		if err != nil {
			t.Fatalf("GetItem error: %v", err)
		}
		if item.Assignee != "Alice" {
			t.Errorf("assignee = %q, want Alice", item.Assignee)
		}
	})

	t.Run("create_item", func(t *testing.T) {
		a := mustAsanaAdapter(t, server, cfg)
		req := &database.Requirement{ReqID: "REQ-NEW-001", RequirementText: "New requirement"}
		gid, err := a.CreateItem(req)
		if err != nil {
			t.Fatalf("CreateItem error: %v", err)
		}
		if gid != "200" {
			t.Errorf("created GID = %q, want 200", gid)
		}
	})

	t.Run("update_item", func(t *testing.T) {
		a := mustAsanaAdapter(t, server, cfg)
		req := &database.Requirement{ReqID: "REQ-CLI-001", RequirementText: "Updated", Status: database.StatusComplete}
		if !a.UpdateItem("100", req) {
			t.Error("UpdateItem should succeed")
		}
	})

	t.Run("status_mapping", func(t *testing.T) {
		a := mustAsanaAdapter(t, server, cfg)
		if a.MapStatusToRTMX("completed") != database.StatusComplete {
			t.Error("completed should map to COMPLETE")
		}
		if a.MapStatusToRTMX("in progress") != database.StatusPartial {
			t.Error("in progress should map to PARTIAL")
		}
		if a.MapStatusToRTMX("not started") != database.StatusMissing {
			t.Error("not started should map to MISSING")
		}
		if a.MapStatusFromRTMX(database.StatusComplete) != "completed" {
			t.Error("COMPLETE should map to completed")
		}
	})
}

// TestAsanaBidirectionalSync validates bidirectional sync between Asana tasks and RTMX requirements.
// REQ-ADAPT-002: Asana bidirectional sync.
func TestAsanaBidirectionalSync(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-002")

	server := newTestAsanaServer(t)
	defer server.Close()

	cfg := &config.AsanaAdapterConfig{
		Enabled:      true,
		ProjectGID:   "proj-1",
		TokenEnv:     "ASANA_TOKEN",
	}

	t.Run("fetch_maps_to_requirements", func(t *testing.T) {
		a := mustAsanaAdapter(t, server, cfg)
		items, err := a.FetchItems(nil)
		if err != nil {
			t.Fatalf("FetchItems error: %v", err)
		}
		// Verify we can map back to RTMX
		for _, item := range items {
			status := a.MapStatusToRTMX(item.Status)
			if status == "" {
				t.Errorf("item %s has empty mapped status", item.ExternalID)
			}
		}
		// First task is completed
		if a.MapStatusToRTMX(items[0].Status) != database.StatusComplete {
			t.Error("completed task should map to COMPLETE")
		}
		// Second task is not completed
		if a.MapStatusToRTMX(items[1].Status) != database.StatusMissing {
			t.Error("incomplete task should map to MISSING")
		}
	})

	t.Run("create_from_requirement", func(t *testing.T) {
		a := mustAsanaAdapter(t, server, cfg)
		req := &database.Requirement{
			ReqID:           "REQ-SYNC-001",
			RequirementText: "Sync test requirement",
			Status:          database.StatusPartial,
			Priority:        database.PriorityHigh,
		}
		gid, err := a.CreateItem(req)
		if err != nil {
			t.Fatalf("CreateItem error: %v", err)
		}
		if gid == "" {
			t.Error("should return non-empty GID")
		}
	})

	t.Run("update_syncs_status", func(t *testing.T) {
		a := mustAsanaAdapter(t, server, cfg)
		req := &database.Requirement{
			ReqID:           "REQ-CLI-001",
			RequirementText: "Build CLI framework",
			Status:          database.StatusComplete,
		}
		if !a.UpdateItem("100", req) {
			t.Error("UpdateItem should succeed")
		}
	})

	t.Run("requirement_id_extracted", func(t *testing.T) {
		a := mustAsanaAdapter(t, server, cfg)
		item, _ := a.GetItem("100")
		if item.RequirementID != "REQ-CLI-001" {
			t.Errorf("RequirementID = %q, want REQ-CLI-001", item.RequirementID)
		}
	})
}

// TestAsanaSectionMapping validates mapping of Asana sections to categories and phases.
// REQ-ADAPT-003: Asana section-to-category mapping.
func TestAsanaSectionMapping(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-003")

	server := newTestAsanaServer(t)
	defer server.Close()

	cfg := &config.AsanaAdapterConfig{
		Enabled:    true,
		ProjectGID: "proj-1",
		TokenEnv:   "ASANA_TOKEN",
	}
	a := mustAsanaAdapter(t, server, cfg)

	t.Run("section_name_to_category", func(t *testing.T) {
		if a.SectionToCategory("CLI Tasks") != "CLI" {
			t.Error("'CLI Tasks' should map to CLI")
		}
		if a.SectionToCategory("MCP Tasks") != "MCP" {
			t.Error("'MCP Tasks' should map to MCP")
		}
		if a.SectionToCategory("API Integration") != "API" {
			t.Error("'API Integration' should map to API")
		}
	})

	t.Run("empty_section_defaults", func(t *testing.T) {
		if a.SectionToCategory("") != "UNCATEGORIZED" {
			t.Error("empty section should map to UNCATEGORIZED")
		}
	})

	t.Run("section_index_to_phase", func(t *testing.T) {
		if a.SectionToPhase(0) != 1 {
			t.Error("first section should be phase 1")
		}
		if a.SectionToPhase(2) != 3 {
			t.Error("third section should be phase 3")
		}
	})

	t.Run("fetched_items_include_section", func(t *testing.T) {
		items, _ := a.FetchItems(nil)
		// First item is in "CLI Tasks" section
		if len(items[0].Labels) == 0 || items[0].Labels[0] != "CLI Tasks" {
			t.Error("first item should have CLI Tasks section label")
		}
		// Second item is in "MCP Tasks" section
		if len(items[1].Labels) == 0 || items[1].Labels[0] != "MCP Tasks" {
			t.Error("second item should have MCP Tasks section label")
		}
	})

	t.Run("custom_status_mapping", func(t *testing.T) {
		cfgCustom := &config.AsanaAdapterConfig{
			Enabled:    true,
			ProjectGID: "proj-1",
			TokenEnv:   "ASANA_TOKEN",
			StatusMapping: map[string]string{
				"In Review": "PARTIAL",
			},
		}
		ac := mustAsanaAdapter(t, server, cfgCustom)
		if ac.MapStatusToRTMX("In Review") != database.StatusPartial {
			t.Error("custom mapping should work")
		}
	})
}

func mustAsanaAdapter(t *testing.T, server *httptest.Server, cfg *config.AsanaAdapterConfig) *AsanaAdapter {
	t.Helper()
	a, err := NewAsanaAdapter(cfg,
		WithHTTPClient(server.Client()),
		WithEnvGetter(func(string) string { return "test-token" }),
	)
	if err != nil {
		t.Fatalf("NewAsanaAdapter error: %v", err)
	}
	a.SetBaseURL(server.URL)
	return a
}
