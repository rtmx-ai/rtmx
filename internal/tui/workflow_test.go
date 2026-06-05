package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
	"github.com/rtmx-ai/rtmx/internal/tui/views"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// helper: send a sequence of key messages through an AppModel, returning
// the final model. Intermediate models are chained through Update.
func sendKeys(t *testing.T, m *AppModel, keys ...tea.KeyMsg) *AppModel {
	t.Helper()
	var model tea.Model = m
	for _, k := range keys {
		model, _ = model.Update(k)
	}
	return model.(*AppModel)
}

func keyRune(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func keyType(k tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: k}
}

// newTestApp creates an AppModel with a test database and a window size set,
// ready for workflow testing.
func newTestApp(t *testing.T) *AppModel {
	t.Helper()
	db, dbPath := testDB(t)
	m := NewAppModel(db, dbPath)
	// Set a reasonable terminal size so views render properly.
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	return m
}

// TestWorkflowTabNavigation exercises the full tab navigation cycle.
// Start on Status -> Tab through all views -> ShiftTab back.
// REQ-TUI-001: Interactive Terminal UI Framework.
func TestWorkflowTabNavigation(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-001")

	m := newTestApp(t)

	// Verify initial state: Status tab active.
	if m.ActiveTab() != TabStatus {
		t.Fatalf("initial tab = %d, want %d (Status)", m.ActiveTab(), TabStatus)
	}
	view := m.View()
	if !strings.Contains(view, "Status") {
		t.Error("initial view should render Status tab")
	}

	// Tab forward through every tab and verify each is active.
	expected := []int{TabBacklog, TabGraph, TabKanban, TabAgents, TabStatus}
	for _, want := range expected {
		m = sendKeys(t, m, keyType(tea.KeyTab))
		if m.ActiveTab() != want {
			t.Errorf("after Tab: got tab %d (%s), want %d (%s)",
				m.ActiveTab(), TabNames[m.ActiveTab()], want, TabNames[want])
		}
		v := m.View()
		if v == "" {
			t.Errorf("tab %d (%s) rendered empty", want, TabNames[want])
		}
	}

	// We are back on Status. Now ShiftTab backward through all tabs.
	backward := []int{TabAgents, TabKanban, TabGraph, TabBacklog, TabStatus}
	for _, want := range backward {
		m = sendKeys(t, m, keyType(tea.KeyShiftTab))
		if m.ActiveTab() != want {
			t.Errorf("after ShiftTab: got tab %d (%s), want %d (%s)",
				m.ActiveTab(), TabNames[m.ActiveTab()], want, TabNames[want])
		}
	}
}

// TestWorkflowTabJumpKeys exercises numeric jump keys (1-5).
// REQ-TUI-001: Interactive Terminal UI Framework.
func TestWorkflowTabJumpKeys(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-001")

	m := newTestApp(t)

	jumps := []struct {
		key  rune
		want int
	}{
		{'5', TabAgents},
		{'3', TabGraph},
		{'1', TabStatus},
		{'4', TabKanban},
		{'2', TabBacklog},
	}
	for _, tt := range jumps {
		m = sendKeys(t, m, keyRune(tt.key))
		if m.ActiveTab() != tt.want {
			t.Errorf("jump key %c: got tab %d, want %d", tt.key, m.ActiveTab(), tt.want)
		}
		v := m.View()
		if !strings.Contains(v, TabNames[tt.want]) {
			t.Errorf("view after jump to %s should contain tab name in tab bar", TabNames[tt.want])
		}
	}
}

// TestWorkflowStatusSortAndNavigate exercises sorting and cursor navigation
// within the Status view as a multi-step workflow:
// switch to Status -> sort by different fields -> navigate with j/k ->
// verify cursor and sort state after each step.
// REQ-TUI-002: Requirements Table View.
func TestWorkflowStatusSortAndNavigate(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-002")

	m := newTestApp(t)

	// Ensure we are on Status tab.
	m = sendKeys(t, m, keyRune('1'))
	if m.ActiveTab() != TabStatus {
		t.Fatal("should be on Status tab")
	}

	// Get the StatusView for inspection.
	sv := m.views[TabStatus].(*views.StatusView)
	if sv.SortField() != "req_id" {
		t.Fatalf("initial sort = %q, want req_id", sv.SortField())
	}

	// Step 1: press 's' to cycle sort to "status".
	m = sendKeys(t, m, keyRune('s'))
	if sv.SortField() != "status" {
		t.Errorf("sort after first s = %q, want status", sv.SortField())
	}

	// Verify the view reflects the sort change in its footer.
	view := m.View()
	if !strings.Contains(view, "status") {
		t.Error("view footer should show current sort field 'status'")
	}

	// Step 2: press 's' again to cycle to "priority".
	m = sendKeys(t, m, keyRune('s'))
	if sv.SortField() != "priority" {
		t.Errorf("sort after second s = %q, want priority", sv.SortField())
	}

	// Step 3: toggle sort direction with 'S'.
	if sv.SortDesc() {
		t.Error("initial sort direction should be ascending")
	}
	m = sendKeys(t, m, keyRune('S'))
	if !sv.SortDesc() {
		t.Error("sort direction should be descending after S")
	}

	// Step 4: navigate down with 'j' three times.
	m = sendKeys(t, m,
		keyRune('j'),
		keyRune('j'),
		keyRune('j'),
	)
	if sv.Cursor() != 3 {
		t.Errorf("cursor after 3x j = %d, want 3", sv.Cursor())
	}

	// The selected req should be non-empty and consistent.
	selectedID := sv.SelectedReqID()
	if selectedID == "" {
		t.Error("SelectedReqID should not be empty after navigation")
	}

	// Step 5: navigate up with 'k'.
	_ = sendKeys(t, m, keyRune('k'))
	if sv.Cursor() != 2 {
		t.Errorf("cursor after k = %d, want 2", sv.Cursor())
	}

	newID := sv.SelectedReqID()
	if newID == selectedID {
		t.Error("SelectedReqID should change after moving cursor")
	}
}

// TestWorkflowKanbanNavAndMove exercises Kanban board column navigation
// and card movement as a multi-step workflow:
// jump to Kanban -> navigate columns -> navigate cards -> move a card.
// REQ-TUI-005: Kanban Board View.
func TestWorkflowKanbanNavAndMove(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-005")

	m := newTestApp(t)

	// Jump to Kanban tab.
	m = sendKeys(t, m, keyRune('4'))
	if m.ActiveTab() != TabKanban {
		t.Fatal("should be on Kanban tab")
	}

	kv := m.views[TabKanban].(*views.KanbanView)

	// Initial column should be 0 (NOT_STARTED).
	if kv.ActiveColumn() != 0 {
		t.Errorf("initial column = %d, want 0", kv.ActiveColumn())
	}

	view := m.View()
	if !strings.Contains(view, "MISSING") {
		t.Error("kanban should show MISSING column")
	}
	if !strings.Contains(view, "COMPLETE") {
		t.Error("kanban should show COMPLETE column")
	}

	// Navigate right to column 1.
	m = sendKeys(t, m, keyRune('l'))
	if kv.ActiveColumn() != 1 {
		t.Errorf("column after l = %d, want 1", kv.ActiveColumn())
	}

	// Navigate right again to column 2.
	m = sendKeys(t, m, keyRune('l'))
	if kv.ActiveColumn() != 2 {
		t.Errorf("column after second l = %d, want 2", kv.ActiveColumn())
	}

	// Navigate left back to column 1.
	m = sendKeys(t, m, keyRune('h'))
	if kv.ActiveColumn() != 1 {
		t.Errorf("column after h = %d, want 1", kv.ActiveColumn())
	}

	// Navigate down within the column.
	m = sendKeys(t, m, keyRune('j'))

	// Move a card with 'm'. The card should transition status.
	// First go to a column with cards. Column 1 is MISSING which has REQ-MCP-002 and REQ-API-001.
	// Moving a MISSING card should move it to PARTIAL (next column).
	m = sendKeys(t, m, keyRune('m'))

	// Verify the kanban re-rendered without panic.
	view = m.View()
	if view == "" {
		t.Error("kanban should render after card move")
	}
}

// TestWorkflowCrossViewStatePreservation verifies that cursor state within
// a view is preserved when switching tabs and returning.
// REQ-TUI-001: Interactive Terminal UI Framework.
func TestWorkflowCrossViewStatePreservation(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-001")

	m := newTestApp(t)

	// On Status tab, move cursor down twice.
	m = sendKeys(t, m, keyRune('1'))
	sv := m.views[TabStatus].(*views.StatusView)
	m = sendKeys(t, m, keyRune('j'), keyRune('j'))
	cursorBefore := sv.Cursor()
	selectedBefore := sv.SelectedReqID()
	sortBefore := sv.SortField()

	if cursorBefore != 2 {
		t.Fatalf("cursor should be 2 before tab switch, got %d", cursorBefore)
	}

	// Switch to Kanban and back.
	m = sendKeys(t, m, keyRune('4')) // Kanban
	if m.ActiveTab() != TabKanban {
		t.Fatal("should be on Kanban tab")
	}
	m = sendKeys(t, m, keyRune('1')) // Back to Status
	if m.ActiveTab() != TabStatus {
		t.Fatal("should be on Status tab")
	}

	// Verify cursor state is preserved in StatusView.
	svAfter := m.views[TabStatus].(*views.StatusView)
	if svAfter.Cursor() != cursorBefore {
		t.Errorf("cursor after tab round-trip = %d, want %d", svAfter.Cursor(), cursorBefore)
	}
	if svAfter.SelectedReqID() != selectedBefore {
		t.Errorf("selectedReqID after tab round-trip = %q, want %q", svAfter.SelectedReqID(), selectedBefore)
	}
	if svAfter.SortField() != sortBefore {
		t.Errorf("sortField after tab round-trip = %q, want %q", svAfter.SortField(), sortBefore)
	}
}

// TestWorkflowBacklogNavigation exercises multi-step navigation in the
// Backlog view: switch to backlog -> navigate items -> verify rendering.
// REQ-TUI-005: Backlog view coverage.
func TestWorkflowBacklogNavigation(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-005")

	m := newTestApp(t)

	// Jump to Backlog tab.
	m = sendKeys(t, m, keyRune('2'))
	if m.ActiveTab() != TabBacklog {
		t.Fatal("should be on Backlog tab")
	}

	view := m.View()
	if !strings.Contains(view, "Backlog") {
		t.Error("backlog view should contain 'Backlog' header")
	}
	// Backlog should show incomplete items (3 of 5 are incomplete).
	if !strings.Contains(view, "incomplete") {
		t.Error("backlog should mention incomplete requirements")
	}

	// Navigate down to second item.
	m = sendKeys(t, m, keyRune('j'))
	view = m.View()
	// Cursor indicator '>' should appear on the second row.
	lines := strings.Split(view, "\n")
	foundCursor := false
	for _, line := range lines {
		if strings.HasPrefix(line, "> ") || strings.HasPrefix(line, ">") {
			foundCursor = true
			break
		}
	}
	if !foundCursor {
		t.Error("cursor indicator '>' should appear in backlog view after navigation")
	}

	// Navigate down past all items -- cursor should clamp.
	for i := 0; i < 10; i++ {
		m = sendKeys(t, m, keyRune('j'))
	}
	// Should not panic and should still render.
	view = m.View()
	if view == "" {
		t.Error("backlog should render after excessive navigation")
	}

	// Navigate up past top -- should clamp at 0.
	for i := 0; i < 20; i++ {
		m = sendKeys(t, m, keyRune('k'))
	}
	view = m.View()
	if view == "" {
		t.Error("backlog should render after upward navigation clamp")
	}
}

// TestWorkflowGraphNavigation exercises navigation through graph layers.
// REQ-TUI-004: Dependency Graph Visualization.
func TestWorkflowGraphNavigation(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-004")

	m := newTestApp(t)

	// Jump to Graph tab.
	m = sendKeys(t, m, keyRune('3'))
	if m.ActiveTab() != TabGraph {
		t.Fatal("should be on Graph tab")
	}

	view := m.View()
	if !strings.Contains(view, "Dependency Graph") {
		t.Error("graph view should contain 'Dependency Graph' header")
	}
	if !strings.Contains(view, "Layer") {
		t.Error("graph view should show layers")
	}

	// Navigate down through layers.
	m = sendKeys(t, m, keyRune('j'))
	m = sendKeys(t, m, keyRune('j'))

	// Should not panic and should render the cursor.
	view = m.View()
	if !strings.Contains(view, ">") {
		t.Error("graph view should show cursor '>' on active layer")
	}

	// Navigate up.
	m = sendKeys(t, m, keyRune('k'))
	view = m.View()
	if view == "" {
		t.Error("graph view should render after up navigation")
	}
}

// TestWorkflowDatabaseReloadPreservesViews verifies that a database reload
// propagates to all views and they continue to render correctly.
// REQ-TUI-006: Auto-refresh on database file changes.
func TestWorkflowDatabaseReloadPreservesViews(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-006")

	db, dbPath := testDB(t)
	m := NewAppModel(db, dbPath)
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Navigate to a specific cursor position on Status tab.
	m = sendKeys(t, m, keyRune('j'), keyRune('j'))

	// Simulate a database change: mark a requirement COMPLETE.
	_ = db.Update("REQ-MCP-001", map[string]interface{}{"status": "COMPLETE"})
	_ = db.Save(dbPath)

	// Load updated database and send reload message.
	newDB, err := database.Load(dbPath)
	if err != nil {
		t.Fatalf("failed to reload database: %v", err)
	}
	newGraph := graph.NewGraph(newDB)
	updated, _ := m.Update(DatabaseReloadedMsg{DB: newDB, Graph: newGraph})
	m = updated.(*AppModel)

	// Status bar should now show 3/5 (was 2/5).
	view := m.View()
	if !strings.Contains(view, "3/5") {
		t.Error("status bar should show 3/5 after reload (3 complete of 5)")
	}

	// Switch through all tabs and verify each renders without panic.
	for i := 0; i < TabCount; i++ {
		m = sendKeys(t, m, keyType(tea.KeyTab))
		v := m.View()
		if v == "" {
			t.Errorf("tab %d (%s) rendered empty after reload", m.ActiveTab(), TabNames[m.ActiveTab()])
		}
	}
}

// TestWorkflowMultiTabRoundTrip exercises a realistic user session:
// check status -> look at backlog -> check graph -> move a kanban card ->
// return to status -> verify everything is consistent.
// REQ-TUI-001: Interactive Terminal UI Framework.
func TestWorkflowMultiTabRoundTrip(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-001")

	m := newTestApp(t)

	// 1. Status tab: browse requirements.
	view := m.View()
	if !strings.Contains(view, "REQ-CLI-001") {
		t.Error("status view should show REQ-CLI-001")
	}

	// 2. Jump to Backlog: check incomplete items.
	m = sendKeys(t, m, keyRune('2'))
	view = m.View()
	if !strings.Contains(view, "Backlog") {
		t.Error("should be on backlog view")
	}
	// MCP-001 is PARTIAL, should be in backlog.
	if !strings.Contains(view, "REQ-MCP-001") {
		t.Error("backlog should show incomplete REQ-MCP-001")
	}

	// 3. Jump to Graph: inspect dependency structure.
	m = sendKeys(t, m, keyRune('3'))
	view = m.View()
	if !strings.Contains(view, "Dependency Graph") {
		t.Error("should be on graph view")
	}

	// 4. Jump to Kanban: move a card.
	m = sendKeys(t, m, keyRune('4'))
	view = m.View()
	if !strings.Contains(view, "m:move") {
		t.Error("kanban should show move hint")
	}

	// 5. Return to Status.
	m = sendKeys(t, m, keyRune('1'))
	if m.ActiveTab() != TabStatus {
		t.Fatal("should be back on Status tab")
	}
	view = m.View()
	if !strings.Contains(view, "REQ-CLI-001") {
		t.Error("status view should still show requirements after round trip")
	}

	// 6. Quit.
	m = sendKeys(t, m, keyRune('q'))
	if !m.Quitting() {
		t.Error("app should be quitting after 'q'")
	}
	if m.View() != "" {
		t.Error("quitting view should be empty")
	}
}

// TestWorkflowHelpOverlayDuringNavigation verifies that help toggle works
// correctly while navigating between tabs.
// REQ-TUI-001: Interactive Terminal UI Framework.
func TestWorkflowHelpOverlayDuringNavigation(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-001")

	m := newTestApp(t)

	// Toggle help on.
	m = sendKeys(t, m, keyRune('?'))
	if !m.ShowHelp() {
		t.Fatal("help should be visible")
	}

	// Tab navigation should still work while help is showing.
	m = sendKeys(t, m, keyType(tea.KeyTab))
	if m.ActiveTab() != TabBacklog {
		t.Error("tab should advance while help is open")
	}
	if !m.ShowHelp() {
		t.Error("help should remain visible after tab switch")
	}

	// Toggle help off.
	m = sendKeys(t, m, keyRune('?'))
	if m.ShowHelp() {
		t.Error("help should be hidden after second toggle")
	}

	// View should render without panic in both states.
	view := m.View()
	if view == "" {
		t.Error("view should not be empty after help toggle")
	}
}
