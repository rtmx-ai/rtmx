package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rtmx-ai/rtmx/internal/graph"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestTUIDetailPane validates the requirement detail pane view.
// REQ-TUI-003: Requirement detail pane with deps, history, and notes.
func TestTUIDetailPane(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-003")

	db, _ := testDB(t)
	g := graph.NewGraph(db)

	t.Run("empty_shows_select_prompt", func(t *testing.T) {
		v := NewDetailView(db, g)
		v.SetSize(100, 30)
		view := v.View()
		if !strings.Contains(view, "Select a requirement") {
			t.Error("empty detail should prompt to select a requirement")
		}
	})

	t.Run("shows_requirement_metadata", func(t *testing.T) {
		v := NewDetailView(db, g)
		v.SetSize(100, 30)
		v.SetReqID("REQ-CLI-001")
		view := v.View()
		for _, want := range []string{"REQ-CLI-001", "CLI", "COMPLETE", "P0"} {
			if !strings.Contains(view, want) {
				t.Errorf("detail should contain %q", want)
			}
		}
	})

	t.Run("shows_category_and_effort", func(t *testing.T) {
		v := NewDetailView(db, g)
		v.SetSize(100, 30)
		v.SetReqID("REQ-MCP-001")
		view := v.View()
		if !strings.Contains(view, "MCP") {
			t.Error("should show category MCP")
		}
		if !strings.Contains(view, "2.0 weeks") {
			t.Error("should show effort 2.0 weeks")
		}
	})

	t.Run("shows_upstream_dependencies", func(t *testing.T) {
		v := NewDetailView(db, g)
		v.SetSize(100, 30)
		v.SetReqID("REQ-CLI-002")
		view := v.View()
		if !strings.Contains(view, "Upstream Dependencies") {
			t.Error("should show upstream section")
		}
		if !strings.Contains(view, "REQ-CLI-001") {
			t.Error("REQ-CLI-002 should show REQ-CLI-001 as upstream")
		}
	})

	t.Run("shows_downstream_dependents", func(t *testing.T) {
		v := NewDetailView(db, g)
		v.SetSize(100, 30)
		v.SetReqID("REQ-CLI-001")
		view := v.View()
		if !strings.Contains(view, "Downstream Dependents") {
			t.Error("should show downstream section")
		}
		if !strings.Contains(view, "REQ-CLI-002") {
			t.Error("REQ-CLI-001 should show REQ-CLI-002 as downstream")
		}
	})

	t.Run("shows_blocked_status", func(t *testing.T) {
		v := NewDetailView(db, g)
		v.SetSize(100, 30)
		v.SetReqID("REQ-MCP-002")
		view := v.View()
		if !strings.Contains(view, "BLOCKED") {
			t.Error("REQ-MCP-002 should show blocked status (dep REQ-MCP-001 is partial)")
		}
	})

	t.Run("not_found_requirement", func(t *testing.T) {
		v := NewDetailView(db, g)
		v.SetSize(100, 30)
		v.SetReqID("REQ-FAKE-999")
		view := v.View()
		if !strings.Contains(view, "not found") {
			t.Error("should show not found for missing req")
		}
	})

	t.Run("escape_clears_selection", func(t *testing.T) {
		v := NewDetailView(db, g)
		v.SetReqID("REQ-CLI-001")
		if v.ReqID() != "REQ-CLI-001" {
			t.Fatal("should have req set")
		}
		v.Update(tea.KeyMsg{Type: tea.KeyEsc})
		if v.ReqID() != "" {
			t.Error("escape should clear reqID")
		}
	})

	t.Run("req_id_getter", func(t *testing.T) {
		v := NewDetailView(db, g)
		if v.ReqID() != "" {
			t.Error("initial reqID should be empty")
		}
		v.SetReqID("REQ-API-001")
		if v.ReqID() != "REQ-API-001" {
			t.Error("reqID should be REQ-API-001")
		}
	})

	t.Run("reload_preserves_selection", func(t *testing.T) {
		v := NewDetailView(db, g)
		v.SetReqID("REQ-CLI-001")
		newG := graph.NewGraph(db)
		v.Reload(db, newG)
		if v.ReqID() != "REQ-CLI-001" {
			t.Error("reload should preserve selected reqID")
		}
	})
}
