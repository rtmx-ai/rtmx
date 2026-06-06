package views

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func testDB(t *testing.T) (*database.Database, string) {
	t.Helper()
	db := database.NewDatabase()
	reqs := []*database.Requirement{
		{ReqID: "REQ-CLI-001", Category: "CLI", RequirementText: "Build CLI framework", Status: database.StatusComplete, Priority: database.PriorityP0, Phase: 1, EffortWeeks: 1.0},
		{ReqID: "REQ-CLI-002", Category: "CLI", RequirementText: "Status command", Status: database.StatusComplete, Priority: database.PriorityHigh, Phase: 1, EffortWeeks: 0.5},
		{ReqID: "REQ-MCP-001", Category: "MCP", RequirementText: "MCP server", Status: database.StatusPartial, Priority: database.PriorityP0, Phase: 2, EffortWeeks: 2.0},
		{ReqID: "REQ-MCP-002", Category: "MCP", RequirementText: "MCP tools", Status: database.StatusMissing, Priority: database.PriorityHigh, Phase: 2, EffortWeeks: 1.0},
		{ReqID: "REQ-API-001", Category: "API", RequirementText: "Requirements endpoint", Status: database.StatusMissing, Priority: database.PriorityP0, Phase: 3, EffortWeeks: 0.5},
	}
	for _, r := range reqs {
		r.Dependencies = make(database.StringSet)
		r.Blocks = make(database.StringSet)
		_ = db.Add(r)
	}
	db.Get("REQ-CLI-002").Dependencies.Add("REQ-CLI-001")
	db.Get("REQ-CLI-001").Blocks.Add("REQ-CLI-002")
	db.Get("REQ-MCP-002").Dependencies.Add("REQ-MCP-001")
	db.Get("REQ-MCP-001").Blocks.Add("REQ-MCP-002")

	tmpDir := t.TempDir()
	rtmxDir := filepath.Join(tmpDir, ".rtmx")
	_ = os.MkdirAll(rtmxDir, 0o755)
	dbPath := filepath.Join(rtmxDir, "database.csv")
	_ = db.Save(dbPath)
	return db, dbPath
}

// TestTUIRequirementsTable validates the requirements table view rendering,
// sorting, navigation, and display of all requirement data.
// REQ-TUI-002: Requirements Table View.
func TestTUIRequirementsTable(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-002")

	db, _ := testDB(t)
	g := graph.NewGraph(db)

	t.Run("renders_column_headers", func(t *testing.T) {
		v := NewStatusView(db, g)
		v.SetSize(120, 30)
		view := v.View()

		for _, header := range []string{"REQ ID", "STATUS", "PRI"} {
			if !strings.Contains(view, header) {
				t.Errorf("view should contain column header %q", header)
			}
		}
	})

	t.Run("renders_all_requirements", func(t *testing.T) {
		v := NewStatusView(db, g)
		v.SetSize(120, 30)
		view := v.View()

		reqIDs := []string{"REQ-CLI-001", "REQ-CLI-002", "REQ-MCP-001", "REQ-MCP-002", "REQ-API-001"}
		for _, id := range reqIDs {
			if !strings.Contains(view, id) {
				t.Errorf("view should contain requirement %q", id)
			}
		}
	})

	t.Run("sort_field_cycles", func(t *testing.T) {
		v := NewStatusView(db, g)

		if v.SortField() != "req_id" {
			t.Fatalf("initial sort field = %q, want %q", v.SortField(), "req_id")
		}

		// Press 's' to cycle through sort fields
		expected := []string{"status", "priority", "category", "effort_weeks", "req_id"}
		for _, want := range expected {
			v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
			if v.SortField() != want {
				t.Errorf("sort field after press = %q, want %q", v.SortField(), want)
			}
		}
	})

	t.Run("sort_direction_toggles", func(t *testing.T) {
		v := NewStatusView(db, g)

		if v.SortDesc() {
			t.Fatal("initial sort should be ascending")
		}

		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
		if !v.SortDesc() {
			t.Error("sort should be descending after 'S'")
		}

		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
		if v.SortDesc() {
			t.Error("sort should be ascending after second 'S'")
		}
	})

	t.Run("navigation_j_k", func(t *testing.T) {
		v := NewStatusView(db, g)
		v.SetSize(120, 30)

		if v.Cursor() != 0 {
			t.Fatalf("initial cursor = %d, want 0", v.Cursor())
		}

		// Move down with 'j'
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if v.Cursor() != 1 {
			t.Errorf("cursor after j = %d, want 1", v.Cursor())
		}

		// Move down again
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		if v.Cursor() != 2 {
			t.Errorf("cursor after second j = %d, want 2", v.Cursor())
		}

		// Move up with 'k'
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if v.Cursor() != 1 {
			t.Errorf("cursor after k = %d, want 1", v.Cursor())
		}
	})

	t.Run("selected_req_id_tracks_cursor", func(t *testing.T) {
		v := NewStatusView(db, g)
		v.SetSize(120, 30)

		first := v.SelectedReqID()
		if first == "" {
			t.Fatal("SelectedReqID should not be empty")
		}

		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		second := v.SelectedReqID()
		if second == "" {
			t.Fatal("SelectedReqID should not be empty after move")
		}
		if first == second {
			t.Error("SelectedReqID should change after cursor move")
		}
	})

	t.Run("cursor_does_not_go_negative", func(t *testing.T) {
		v := NewStatusView(db, g)

		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		if v.Cursor() != 0 {
			t.Errorf("cursor should not go below 0, got %d", v.Cursor())
		}
	})

	t.Run("cursor_does_not_exceed_count", func(t *testing.T) {
		v := NewStatusView(db, g)
		v.SetSize(120, 30)

		// Press 'j' more times than there are requirements
		for i := 0; i < 10; i++ {
			v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		}
		if v.Cursor() != 4 {
			t.Errorf("cursor should cap at 4 (len-1), got %d", v.Cursor())
		}
	})

	t.Run("footer_shows_sort_info", func(t *testing.T) {
		v := NewStatusView(db, g)
		v.SetSize(120, 30)
		view := v.View()

		if !strings.Contains(view, "sort:") {
			t.Error("footer should display sort information")
		}
		if !strings.Contains(view, "req_id") {
			t.Error("footer should display current sort field")
		}
		if !strings.Contains(view, "asc") {
			t.Error("footer should display sort direction")
		}
	})

	t.Run("empty_database", func(t *testing.T) {
		emptyDB := database.NewDatabase()
		emptyGraph := graph.NewGraph(emptyDB)
		v := NewStatusView(emptyDB, emptyGraph)
		v.SetSize(80, 24)
		view := v.View()

		if !strings.Contains(view, "No requirements") {
			t.Error("empty view should show 'No requirements' message")
		}
	})
}
