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

func newTestMondayServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}
		if r.Header.Get("Authorization") != "test-token" {
			w.WriteHeader(401)
			return
		}

		var body struct {
			Query string `json:"query"`
		}
		json.NewDecoder(r.Body).Decode(&body)
		q := body.Query

		switch {
		case strings.Contains(q, "me {"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"me": map[string]interface{}{"name": "Test Monday User"},
				},
			})
		case strings.Contains(q, "boards(ids:"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"boards": []map[string]interface{}{
						{
							"items_page": map[string]interface{}{
								"items": []map[string]interface{}{
									{
										"id": "1001", "name": "[REQ-CLI-001] Build CLI", "state": "active",
										"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z",
										"group":         map[string]interface{}{"id": "g1", "title": "CLI Development"},
										"column_values": []map[string]interface{}{{"id": "status", "title": "Status", "text": "Done"}},
									},
									{
										"id": "1002", "name": "[REQ-MCP-001] MCP server", "state": "active",
										"created_at": "2026-01-03T00:00:00Z", "updated_at": "2026-01-04T00:00:00Z",
										"group":         map[string]interface{}{"id": "g2", "title": "MCP Backend"},
										"column_values": []map[string]interface{}{{"id": "status", "title": "Status", "text": "Working on it"}},
									},
								},
							},
						},
					},
				},
			})
		case strings.Contains(q, "items(ids:"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"items": []map[string]interface{}{
						{
							"id": "1001", "name": "[REQ-CLI-001] Build CLI", "state": "active",
							"created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-02T00:00:00Z",
							"group":         map[string]interface{}{"id": "g1", "title": "CLI Development"},
							"column_values": []map[string]interface{}{{"id": "status", "title": "Status", "text": "Done"}},
						},
					},
				},
			})
		case strings.Contains(q, "create_item"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"create_item": map[string]interface{}{"id": "2001"},
				},
			})
		case strings.Contains(q, "change_simple_column_value"):
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"change_simple_column_value": map[string]interface{}{"id": "1001"},
				},
			})
		default:
			w.WriteHeader(400)
		}
	}))
}

// TestMondayAdapter validates the Monday.com GraphQL API adapter.
// REQ-ADAPT-004: Monday.com GraphQL API adapter with token auth and ServiceAdapter interface.
func TestMondayAdapter(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-004")

	server := newTestMondayServer(t)
	defer server.Close()

	cfg := &config.MondayAdapterConfig{
		Enabled:  true,
		BoardID:  "board-1",
		TokenEnv: "MONDAY_TOKEN",
	}

	t.Run("creation_requires_enabled", func(t *testing.T) {
		_, err := NewMondayAdapter(&config.MondayAdapterConfig{Enabled: false}, WithEnvGetter(func(string) string { return "x" }))
		if err == nil {
			t.Error("should fail when not enabled")
		}
	})

	t.Run("creation_requires_token", func(t *testing.T) {
		_, err := NewMondayAdapter(cfg, WithEnvGetter(func(string) string { return "" }))
		if err == nil {
			t.Error("should fail without token")
		}
	})

	t.Run("name_returns_monday", func(t *testing.T) {
		m := mustMondayAdapter(t, server, cfg)
		if m.Name() != "monday" {
			t.Errorf("Name() = %q, want monday", m.Name())
		}
	})

	t.Run("is_configured", func(t *testing.T) {
		m := mustMondayAdapter(t, server, cfg)
		if !m.IsConfigured() {
			t.Error("should be configured")
		}
	})

	t.Run("test_connection", func(t *testing.T) {
		m := mustMondayAdapter(t, server, cfg)
		ok, msg := m.TestConnection()
		if !ok {
			t.Errorf("TestConnection failed: %s", msg)
		}
		if !strings.Contains(msg, "Test Monday User") {
			t.Errorf("should contain user name, got %q", msg)
		}
	})

	t.Run("fetch_items", func(t *testing.T) {
		m := mustMondayAdapter(t, server, cfg)
		items, err := m.FetchItems(nil)
		if err != nil {
			t.Fatalf("FetchItems error: %v", err)
		}
		if len(items) != 2 {
			t.Fatalf("got %d items, want 2", len(items))
		}
		if items[0].ExternalID != "1001" {
			t.Errorf("first item ID = %q, want 1001", items[0].ExternalID)
		}
	})

	t.Run("get_item", func(t *testing.T) {
		m := mustMondayAdapter(t, server, cfg)
		item, err := m.GetItem("1001")
		if err != nil {
			t.Fatalf("GetItem error: %v", err)
		}
		if item.ExternalID != "1001" {
			t.Errorf("item ID = %q, want 1001", item.ExternalID)
		}
	})

	t.Run("create_item", func(t *testing.T) {
		m := mustMondayAdapter(t, server, cfg)
		req := &database.Requirement{ReqID: "REQ-NEW-001", RequirementText: "New requirement"}
		id, err := m.CreateItem(req)
		if err != nil {
			t.Fatalf("CreateItem error: %v", err)
		}
		if id != "2001" {
			t.Errorf("created ID = %q, want 2001", id)
		}
	})

	t.Run("update_item", func(t *testing.T) {
		m := mustMondayAdapter(t, server, cfg)
		req := &database.Requirement{ReqID: "REQ-CLI-001", RequirementText: "Updated"}
		if !m.UpdateItem("1001", req) {
			t.Error("UpdateItem should succeed")
		}
	})

	t.Run("status_mapping", func(t *testing.T) {
		m := mustMondayAdapter(t, server, cfg)
		if m.MapStatusToRTMX("Done") != database.StatusComplete {
			t.Error("Done should map to COMPLETE")
		}
		if m.MapStatusToRTMX("Working on it") != database.StatusPartial {
			t.Error("Working on it should map to PARTIAL")
		}
		if m.MapStatusToRTMX("Not Started") != database.StatusMissing {
			t.Error("Not Started should map to MISSING")
		}
		if m.MapStatusFromRTMX(database.StatusComplete) != "Done" {
			t.Error("COMPLETE should map to Done")
		}
	})
}

// TestMondayBidirectionalSync validates bidirectional sync between Monday.com items and RTMX requirements.
// REQ-ADAPT-005: Monday.com bidirectional sync.
func TestMondayBidirectionalSync(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-005")

	server := newTestMondayServer(t)
	defer server.Close()

	cfg := &config.MondayAdapterConfig{
		Enabled:  true,
		BoardID:  "board-1",
		TokenEnv: "MONDAY_TOKEN",
	}

	t.Run("fetch_and_map_status", func(t *testing.T) {
		m := mustMondayAdapter(t, server, cfg)
		items, _ := m.FetchItems(nil)
		// First item has status "Done"
		if m.MapStatusToRTMX(items[0].Status) != database.StatusComplete {
			t.Error("first item (Done) should map to COMPLETE")
		}
		// Second item has status "Working on it"
		if m.MapStatusToRTMX(items[1].Status) != database.StatusPartial {
			t.Error("second item (Working on it) should map to PARTIAL")
		}
	})

	t.Run("create_from_requirement", func(t *testing.T) {
		m := mustMondayAdapter(t, server, cfg)
		req := &database.Requirement{ReqID: "REQ-SYNC-001", RequirementText: "Sync test"}
		id, err := m.CreateItem(req)
		if err != nil {
			t.Fatalf("error: %v", err)
		}
		if id == "" {
			t.Error("should return non-empty ID")
		}
	})

	t.Run("update_propagates", func(t *testing.T) {
		m := mustMondayAdapter(t, server, cfg)
		req := &database.Requirement{ReqID: "REQ-CLI-001", RequirementText: "Updated CLI", Status: database.StatusComplete}
		if !m.UpdateItem("1001", req) {
			t.Error("should succeed")
		}
	})
}

// TestMondayGroupMapping validates mapping of Monday board groups to RTMX categories.
// REQ-ADAPT-006: Monday group-to-category mapping.
func TestMondayGroupMapping(t *testing.T) {
	rtmx.Req(t, "REQ-ADAPT-006")

	server := newTestMondayServer(t)
	defer server.Close()

	cfg := &config.MondayAdapterConfig{
		Enabled:  true,
		BoardID:  "board-1",
		TokenEnv: "MONDAY_TOKEN",
	}
	m := mustMondayAdapter(t, server, cfg)

	t.Run("group_title_to_category", func(t *testing.T) {
		if m.GroupToCategory("CLI Development") != "CLI" {
			t.Error("'CLI Development' should map to CLI")
		}
		if m.GroupToCategory("MCP Backend") != "MCP" {
			t.Error("'MCP Backend' should map to MCP")
		}
	})

	t.Run("empty_group_defaults", func(t *testing.T) {
		if m.GroupToCategory("") != "UNCATEGORIZED" {
			t.Error("empty group should map to UNCATEGORIZED")
		}
	})

	t.Run("fetched_items_include_group", func(t *testing.T) {
		items, _ := m.FetchItems(nil)
		if len(items[0].Labels) == 0 || items[0].Labels[0] != "CLI Development" {
			t.Error("first item should have CLI Development group label")
		}
		if len(items[1].Labels) == 0 || items[1].Labels[0] != "MCP Backend" {
			t.Error("second item should have MCP Backend group label")
		}
	})

	t.Run("custom_status_mapping", func(t *testing.T) {
		cfgCustom := &config.MondayAdapterConfig{
			Enabled:  true,
			BoardID:  "board-1",
			TokenEnv: "MONDAY_TOKEN",
			StatusMapping: map[string]string{
				"Review": "PARTIAL",
			},
		}
		mc := mustMondayAdapter(t, server, cfgCustom)
		if mc.MapStatusToRTMX("Review") != database.StatusPartial {
			t.Error("custom mapping should work")
		}
	})
}

func mustMondayAdapter(t *testing.T, server *httptest.Server, cfg *config.MondayAdapterConfig) *MondayAdapter {
	t.Helper()
	m, err := NewMondayAdapter(cfg,
		WithHTTPClient(server.Client()),
		WithEnvGetter(func(string) string { return "test-token" }),
	)
	if err != nil {
		t.Fatalf("NewMondayAdapter error: %v", err)
	}
	m.SetAPIURL(server.URL)
	return m
}
