package tui

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

// TestTUIAppInit validates the TUI app model initialization and interaction.
// REQ-TUI-001: Interactive Terminal UI Framework with Bubble Tea.
func TestTUIAppInit(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-001")

	db, dbPath := testDB(t)

	t.Run("model_initializes", func(t *testing.T) {
		m := NewAppModel(db, dbPath)
		if m == nil {
			t.Fatal("NewAppModel returned nil")
		}
		if m.ActiveTab() != TabStatus {
			t.Errorf("initial tab = %d, want %d (Status)", m.ActiveTab(), TabStatus)
		}
		for i, v := range m.views {
			if v == nil {
				t.Errorf("view[%d] (%s) is nil", i, TabNames[i])
			}
		}
	})

	t.Run("window_size_handled", func(t *testing.T) {
		m := NewAppModel(db, dbPath)
		updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		app := updated.(*AppModel)
		if app.width != 120 || app.height != 40 {
			t.Errorf("size = %dx%d, want 120x40", app.width, app.height)
		}
	})

	t.Run("tab_switches_forward", func(t *testing.T) {
		m := NewAppModel(db, dbPath)
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
		app := updated.(*AppModel)
		if app.ActiveTab() != TabBacklog {
			t.Errorf("tab = %d, want %d (Backlog)", app.ActiveTab(), TabBacklog)
		}
	})

	t.Run("tab_switches_backward", func(t *testing.T) {
		m := NewAppModel(db, dbPath)
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
		app := updated.(*AppModel)
		if app.ActiveTab() != TabAgents {
			t.Errorf("tab = %d, want %d (Agents)", app.ActiveTab(), TabAgents)
		}
	})

	t.Run("tab_wraps_around", func(t *testing.T) {
		m := NewAppModel(db, dbPath)
		for i := 0; i < TabCount; i++ {
			m.Update(tea.KeyMsg{Type: tea.KeyTab})
		}
		// After TabCount tabs, should be back to Status
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
		app := updated.(*AppModel)
		if app.ActiveTab() != TabBacklog {
			t.Errorf("tab after wrap = %d, want %d", app.ActiveTab(), TabBacklog)
		}
	})

	t.Run("number_keys_jump_to_tab", func(t *testing.T) {
		tests := []struct {
			key  rune
			want int
		}{
			{'1', TabStatus},
			{'2', TabBacklog},
			{'3', TabGraph},
			{'4', TabKanban},
			{'5', TabAgents},
		}
		for _, tt := range tests {
			m := NewAppModel(db, dbPath)
			updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{tt.key}})
			app := updated.(*AppModel)
			if app.ActiveTab() != tt.want {
				t.Errorf("key %c: tab = %d, want %d", tt.key, app.ActiveTab(), tt.want)
			}
		}
	})

	t.Run("quit_key", func(t *testing.T) {
		m := NewAppModel(db, dbPath)
		updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		app := updated.(*AppModel)
		if !app.Quitting() {
			t.Error("should be quitting after 'q'")
		}
		if cmd == nil {
			t.Error("quit should return a tea.Cmd")
		}
	})

	t.Run("help_toggle", func(t *testing.T) {
		m := NewAppModel(db, dbPath)
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
		app := updated.(*AppModel)
		if !app.ShowHelp() {
			t.Error("help should be visible after '?'")
		}
		updated, _ = app.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
		app = updated.(*AppModel)
		if app.ShowHelp() {
			t.Error("help should be hidden after second '?'")
		}
	})

	t.Run("view_renders_without_panic", func(t *testing.T) {
		m := NewAppModel(db, dbPath)
		m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		view := m.View()
		if view == "" {
			t.Error("view should not be empty")
		}
		// Tab bar should show all tabs
		for _, name := range TabNames {
			if !strings.Contains(view, name) {
				t.Errorf("view should contain tab name %q", name)
			}
		}
	})

	t.Run("status_bar_shows_counts", func(t *testing.T) {
		m := NewAppModel(db, dbPath)
		m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
		view := m.View()
		if !strings.Contains(view, "2/5") {
			t.Error("status bar should show 2/5 (2 complete of 5)")
		}
	})

	t.Run("each_tab_renders", func(t *testing.T) {
		m := NewAppModel(db, dbPath)
		m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
		for i := 0; i < TabCount; i++ {
			m.activeTab = i
			view := m.View()
			if view == "" {
				t.Errorf("tab %d (%s) rendered empty", i, TabNames[i])
			}
		}
	})

	t.Run("database_reload", func(t *testing.T) {
		m := NewAppModel(db, dbPath)
		// Modify and save
		_ = db.Update("REQ-MCP-001", map[string]interface{}{"status": "COMPLETE"})
		_ = db.Save(dbPath)
		// Simulate reload
		newDB, _ := database.Load(dbPath)
		newGraph := graph.NewGraph(newDB)
		m.Update(DatabaseReloadedMsg{DB: newDB, Graph: newGraph})
		// Model should have updated db
		if m.db.Get("REQ-MCP-001").Status != database.StatusComplete {
			t.Error("database should be reloaded")
		}
	})

	t.Run("quitting_renders_empty", func(t *testing.T) {
		m := NewAppModel(db, dbPath)
		m.quitting = true
		if m.View() != "" {
			t.Error("quitting view should be empty")
		}
	})
}
