package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestTUIGraphView validates the dependency graph view rendering,
// layer display, critical path markers, and navigation.
// REQ-TUI-004: Dependency Graph View.
func TestTUIGraphView(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-004")

	db, _ := testDB(t)
	g := graph.NewGraph(db)

	t.Run("renders_header_with_counts", func(t *testing.T) {
		v := NewGraphView(db, g)
		v.SetSize(120, 30)
		view := v.View()

		if !strings.Contains(view, "Dependency Graph:") {
			t.Error("view should contain 'Dependency Graph:' header")
		}
		if !strings.Contains(view, "nodes") {
			t.Error("view should display node count")
		}
		if !strings.Contains(view, "edges") {
			t.Error("view should display edge count")
		}
	})

	t.Run("renders_layers", func(t *testing.T) {
		v := NewGraphView(db, g)
		v.SetSize(120, 30)
		view := v.View()

		if !strings.Contains(view, "Layer 0:") {
			t.Error("view should contain 'Layer 0:'")
		}
		if !strings.Contains(view, "Layer 1:") {
			t.Error("view should contain 'Layer 1:'")
		}
	})

	t.Run("critical_path_markers", func(t *testing.T) {
		v := NewGraphView(db, g)
		v.SetSize(120, 30)
		view := v.View()

		if !strings.Contains(view, "* = critical path") {
			t.Error("view should contain critical path legend")
		}
		// Critical path nodes are marked with *
		if !strings.Contains(view, "*") {
			t.Error("view should contain at least one critical path marker")
		}
	})

	t.Run("web_count_displayed", func(t *testing.T) {
		v := NewGraphView(db, g)
		v.SetSize(120, 30)
		view := v.View()

		if !strings.Contains(view, "independent webs") {
			t.Error("view should display web count")
		}
	})

	t.Run("navigation_j_k_between_layers", func(t *testing.T) {
		v := NewGraphView(db, g)
		v.SetSize(120, 30)

		// Initial cursor at layer 0
		view0 := v.View()
		// Layer 0 should have the cursor marker
		lines0 := strings.Split(view0, "\n")
		foundCursor0 := false
		for _, line := range lines0 {
			if strings.Contains(line, "> Layer 0:") {
				foundCursor0 = true
				break
			}
		}
		if !foundCursor0 {
			t.Error("cursor should be on Layer 0 initially")
		}

		// Move down
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		view1 := v.View()
		lines1 := strings.Split(view1, "\n")
		foundCursor1 := false
		for _, line := range lines1 {
			if strings.Contains(line, "> Layer 1:") {
				foundCursor1 = true
				break
			}
		}
		if !foundCursor1 {
			t.Error("cursor should be on Layer 1 after 'j'")
		}

		// Move back up
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		view2 := v.View()
		lines2 := strings.Split(view2, "\n")
		foundCursor2 := false
		for _, line := range lines2 {
			if strings.Contains(line, "> Layer 0:") {
				foundCursor2 = true
				break
			}
		}
		if !foundCursor2 {
			t.Error("cursor should be back on Layer 0 after 'k'")
		}
	})

	t.Run("cursor_bounded", func(t *testing.T) {
		v := NewGraphView(db, g)

		// Try moving up from initial position
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		view := v.View()
		if !strings.Contains(view, "> Layer 0:") {
			t.Error("cursor should not go above Layer 0")
		}

		// Move past the end
		for i := 0; i < 20; i++ {
			v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		}
		// Should not panic and should still render
		finalView := v.View()
		if finalView == "" {
			t.Error("view should not be empty after excessive navigation")
		}
	})

	t.Run("nil_graph_shows_no_graph", func(t *testing.T) {
		emptyDB := database.NewDatabase()
		v := NewGraphView(emptyDB, nil)
		v.SetSize(80, 24)
		view := v.View()

		if !strings.Contains(view, "No dependency graph") {
			t.Errorf("nil graph should show 'No dependency graph', got: %q", view)
		}
	})

	t.Run("empty_database_renders", func(t *testing.T) {
		emptyDB := database.NewDatabase()
		emptyGraph := graph.NewGraph(emptyDB)
		v := NewGraphView(emptyDB, emptyGraph)
		v.SetSize(80, 24)
		view := v.View()

		// Empty database with a graph object still renders (0 nodes)
		if view == "" {
			t.Error("empty database graph view should not be empty")
		}
		if !strings.Contains(view, "0 nodes") {
			t.Error("empty database should show 0 nodes")
		}
	})

	t.Run("all_req_ids_visible", func(t *testing.T) {
		v := NewGraphView(db, g)
		v.SetSize(120, 30)
		view := v.View()

		// All reqs with dependencies should appear in the graph
		for _, id := range []string{"REQ-CLI-001", "REQ-CLI-002", "REQ-MCP-001", "REQ-MCP-002"} {
			if !strings.Contains(view, id) {
				t.Errorf("view should contain %q", id)
			}
		}
	})
}
