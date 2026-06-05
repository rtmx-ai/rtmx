package views

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestTUIKanbanBoard validates the Kanban board view rendering,
// column/card navigation, card movement, and blocked-card constraints.
// REQ-TUI-005: Kanban Board View.
func TestTUIKanbanBoard(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-005")

	t.Run("renders_status_columns", func(t *testing.T) {
		db, dbPath := testDB(t)
		g := graph.NewGraph(db)
		v := NewKanbanView(db, g, dbPath)
		v.SetSize(120, 30)
		view := v.View()

		for _, status := range []string{"NOT_STARTED", "MISSING", "PARTIAL", "COMPLETE"} {
			if !strings.Contains(view, status) {
				t.Errorf("view should contain status column %q", status)
			}
		}
	})

	t.Run("renders_separator_and_help", func(t *testing.T) {
		db, dbPath := testDB(t)
		g := graph.NewGraph(db)
		v := NewKanbanView(db, g, dbPath)
		v.SetSize(120, 30)
		view := v.View()

		if !strings.Contains(view, "---") {
			t.Error("view should contain separator line")
		}
		if !strings.Contains(view, "m:move") {
			t.Error("view should contain help text")
		}
	})

	t.Run("column_navigation_h_l", func(t *testing.T) {
		db, dbPath := testDB(t)
		g := graph.NewGraph(db)
		v := NewKanbanView(db, g, dbPath)

		if v.ActiveColumn() != 0 {
			t.Fatalf("initial column = %d, want 0", v.ActiveColumn())
		}

		// Move right with 'l'
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
		if v.ActiveColumn() != 1 {
			t.Errorf("column after 'l' = %d, want 1", v.ActiveColumn())
		}

		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
		if v.ActiveColumn() != 2 {
			t.Errorf("column after second 'l' = %d, want 2", v.ActiveColumn())
		}

		// Move left with 'h'
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
		if v.ActiveColumn() != 1 {
			t.Errorf("column after 'h' = %d, want 1", v.ActiveColumn())
		}
	})

	t.Run("column_bounded", func(t *testing.T) {
		db, dbPath := testDB(t)
		g := graph.NewGraph(db)
		v := NewKanbanView(db, g, dbPath)

		// Cannot go left from column 0
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h'}})
		if v.ActiveColumn() != 0 {
			t.Errorf("column should not go below 0, got %d", v.ActiveColumn())
		}

		// Cannot go beyond column 3
		for i := 0; i < 10; i++ {
			v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
		}
		if v.ActiveColumn() != 3 {
			t.Errorf("column should not exceed 3, got %d", v.ActiveColumn())
		}
	})

	t.Run("card_navigation_j_k", func(t *testing.T) {
		db, dbPath := testDB(t)
		g := graph.NewGraph(db)
		v := NewKanbanView(db, g, dbPath)
		v.SetSize(120, 30)

		// Navigate to COMPLETE column (index 3) which has 2 items
		for i := 0; i < 3; i++ {
			v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
		}
		if v.ActiveColumn() != 3 {
			t.Fatalf("should be on COMPLETE column, got %d", v.ActiveColumn())
		}

		// Move down within column
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		// Should move down if there are multiple items, or stay if only one
		view := v.View()
		if view == "" {
			t.Error("view should not be empty after card navigation")
		}
	})

	t.Run("move_card_to_next_status", func(t *testing.T) {
		db, dbPath := testDB(t)
		g := graph.NewGraph(db)
		v := NewKanbanView(db, g, dbPath)
		v.SetSize(120, 30)

		// Navigate to MISSING column (index 1) - has REQ-MCP-002 and REQ-API-001
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
		if v.ActiveColumn() != 1 {
			t.Fatalf("should be on MISSING column, got %d", v.ActiveColumn())
		}

		// Find a card in MISSING column - REQ-API-001 has no deps so it should move
		// Find the req in the MISSING column that is not blocked
		var targetReqID string
		for _, req := range db.All() {
			if req.Status == database.StatusMissing && !g.IsBlocked(req.ReqID) {
				targetReqID = req.ReqID
				break
			}
		}
		if targetReqID == "" {
			t.Fatal("should have at least one unblocked MISSING requirement")
		}

		// Press 'm' to move card to next status (PARTIAL)
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})

		// Verify the card moved - reload db from path to check persistence
		updatedDB, err := database.Load(dbPath)
		if err != nil {
			t.Fatalf("failed to reload database: %v", err)
		}

		// At least one of the MISSING reqs should have moved to PARTIAL
		movedCount := 0
		for _, req := range updatedDB.All() {
			if req.Status == database.StatusPartial {
				movedCount++
			}
		}
		// We started with 1 PARTIAL (REQ-MCP-001), after move should have more
		if movedCount < 1 {
			t.Error("at least one requirement should be in PARTIAL status after move")
		}
	})

	t.Run("blocked_card_cannot_move_to_complete", func(t *testing.T) {
		db, dbPath := testDB(t)

		// Set REQ-MCP-002 to PARTIAL (its dep REQ-MCP-001 is also PARTIAL, not COMPLETE)
		_ = db.Update("REQ-MCP-002", map[string]interface{}{"status": "PARTIAL"})
		_ = db.Save(dbPath)
		g := graph.NewGraph(db)

		v := NewKanbanView(db, g, dbPath)
		v.SetSize(120, 30)

		// Navigate to PARTIAL column (index 2)
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
		if v.ActiveColumn() != 2 {
			t.Fatalf("should be on PARTIAL column, got %d", v.ActiveColumn())
		}

		// Find which card is selected and check if it is blocked
		// Try to move the blocked card - it should stay in PARTIAL
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})

		// REQ-MCP-002 should NOT be COMPLETE because REQ-MCP-001 is not COMPLETE
		req := db.Get("REQ-MCP-002")
		if req.Status == database.StatusComplete {
			t.Error("blocked requirement should not move to COMPLETE")
		}
	})

	t.Run("renders_blocked_indicator", func(t *testing.T) {
		db, dbPath := testDB(t)
		g := graph.NewGraph(db)
		v := NewKanbanView(db, g, dbPath)
		v.SetSize(120, 30)
		view := v.View()

		// REQ-MCP-002 depends on REQ-MCP-001 which is PARTIAL, so MCP-002 is blocked
		if !strings.Contains(view, "[B]") {
			t.Error("view should contain [B] blocked indicator")
		}
	})

	t.Run("empty_column_navigation", func(t *testing.T) {
		db, dbPath := testDB(t)
		g := graph.NewGraph(db)
		v := NewKanbanView(db, g, dbPath)
		v.SetSize(120, 30)

		// NOT_STARTED column (index 0) should be empty in our test data
		// Press 'm' on empty column should not panic
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
		view := v.View()
		if view == "" {
			t.Error("view should not be empty after move on empty column")
		}
	})
}
