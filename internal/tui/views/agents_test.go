package views

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rtmx-ai/rtmx/internal/orchestration"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestTUIAgentMonitor validates the agent claims monitor view rendering,
// stale detection, claim display, and navigation.
// REQ-TUI-007: Agent Activity Monitor.
func TestTUIAgentMonitor(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-007")

	t.Run("empty_claims_shows_no_active", func(t *testing.T) {
		db, dbPath := testDB(t)
		v := NewAgentsView(db, dbPath)
		v.SetSize(120, 30)
		view := v.View()

		if !strings.Contains(view, "No active claims") {
			t.Errorf("empty claims should show 'No active claims', got: %q", view)
		}
	})

	t.Run("no_dbpath_shows_no_active", func(t *testing.T) {
		db, _ := testDB(t)
		v := NewAgentsView(db, "")
		v.SetSize(120, 30)
		view := v.View()

		if !strings.Contains(view, "No active claims") {
			t.Error("nil dbPath should show 'No active claims'")
		}
	})

	t.Run("displays_claims_after_refresh", func(t *testing.T) {
		db, dbPath := testDB(t)
		claimsDir := filepath.Join(filepath.Dir(dbPath), "claims")

		store, err := orchestration.NewClaimStore(claimsDir)
		if err != nil {
			t.Fatalf("failed to create claim store: %v", err)
		}

		_, err = store.Claim("REQ-CLI-001", "agent-alpha")
		if err != nil {
			t.Fatalf("failed to create claim: %v", err)
		}
		_, err = store.Claim("REQ-MCP-001", "agent-beta")
		if err != nil {
			t.Fatalf("failed to create claim: %v", err)
		}

		v := NewAgentsView(db, dbPath)
		v.SetSize(120, 30)
		view := v.View()

		if strings.Contains(view, "No active claims") {
			t.Error("view should not show 'No active claims' when claims exist")
		}
		if !strings.Contains(view, "REQ-CLI-001") {
			t.Error("view should contain REQ-CLI-001 claim")
		}
		if !strings.Contains(view, "REQ-MCP-001") {
			t.Error("view should contain REQ-MCP-001 claim")
		}
		if !strings.Contains(view, "agent-alpha") {
			t.Error("view should contain agent-alpha")
		}
		if !strings.Contains(view, "agent-beta") {
			t.Error("view should contain agent-beta")
		}
	})

	t.Run("displays_claim_header_counts", func(t *testing.T) {
		db, dbPath := testDB(t)
		claimsDir := filepath.Join(filepath.Dir(dbPath), "claims")

		store, err := orchestration.NewClaimStore(claimsDir)
		if err != nil {
			t.Fatalf("failed to create claim store: %v", err)
		}
		_, _ = store.Claim("REQ-CLI-001", "agent-alpha")

		v := NewAgentsView(db, dbPath)
		v.SetSize(120, 30)
		view := v.View()

		if !strings.Contains(view, "Agent Claims:") {
			t.Error("view should contain 'Agent Claims:' header")
		}
		if !strings.Contains(view, "1 active") {
			t.Error("view should show '1 active' count")
		}
	})

	t.Run("displays_column_headers_with_claims", func(t *testing.T) {
		db, dbPath := testDB(t)
		claimsDir := filepath.Join(filepath.Dir(dbPath), "claims")

		store, err := orchestration.NewClaimStore(claimsDir)
		if err != nil {
			t.Fatalf("failed to create claim store: %v", err)
		}
		_, _ = store.Claim("REQ-CLI-001", "agent-alpha")

		v := NewAgentsView(db, dbPath)
		v.SetSize(120, 30)
		view := v.View()

		for _, header := range []string{"REQ ID", "AGENT", "CLAIMED AT", "STATUS"} {
			if !strings.Contains(view, header) {
				t.Errorf("view should contain column header %q", header)
			}
		}
	})

	t.Run("stale_detection", func(t *testing.T) {
		db, dbPath := testDB(t)
		claimsDir := filepath.Join(filepath.Dir(dbPath), "claims")
		_ = os.MkdirAll(claimsDir, 0o755)

		// Write a claim file with a timestamp older than 15 minutes
		staleClaim := orchestration.Claim{
			ReqID:     "REQ-CLI-001",
			AgentID:   "agent-stale",
			ClaimedAt: time.Now().UTC().Add(-20 * time.Minute),
		}
		data, err := json.MarshalIndent(&staleClaim, "", "  ")
		if err != nil {
			t.Fatalf("failed to marshal stale claim: %v", err)
		}
		err = os.WriteFile(filepath.Join(claimsDir, "REQ-CLI-001.json"), data, 0o644)
		if err != nil {
			t.Fatalf("failed to write stale claim: %v", err)
		}

		v := NewAgentsView(db, dbPath)
		v.SetSize(120, 30)
		view := v.View()

		if !strings.Contains(view, "STALE") {
			t.Error("stale claim should display 'STALE' status")
		}
		if !strings.Contains(view, "1 stale") {
			t.Error("header should show '1 stale' count")
		}
	})

	t.Run("active_claim_shows_active_status", func(t *testing.T) {
		db, dbPath := testDB(t)
		claimsDir := filepath.Join(filepath.Dir(dbPath), "claims")

		store, err := orchestration.NewClaimStore(claimsDir)
		if err != nil {
			t.Fatalf("failed to create claim store: %v", err)
		}
		_, _ = store.Claim("REQ-CLI-001", "agent-fresh")

		v := NewAgentsView(db, dbPath)
		v.SetSize(120, 30)
		view := v.View()

		if !strings.Contains(view, "active") {
			t.Error("fresh claim should display 'active' status")
		}
	})

	t.Run("navigation_j_k", func(t *testing.T) {
		db, dbPath := testDB(t)
		claimsDir := filepath.Join(filepath.Dir(dbPath), "claims")

		store, err := orchestration.NewClaimStore(claimsDir)
		if err != nil {
			t.Fatalf("failed to create claim store: %v", err)
		}
		_, _ = store.Claim("REQ-CLI-001", "agent-alpha")
		_, _ = store.Claim("REQ-MCP-001", "agent-beta")

		v := NewAgentsView(db, dbPath)
		v.SetSize(120, 30)

		// Initial render should have cursor on first claim
		view0 := v.View()
		lines0 := strings.Split(view0, "\n")
		foundFirstCursor := false
		for _, line := range lines0 {
			if strings.HasPrefix(line, "> ") {
				foundFirstCursor = true
				break
			}
		}
		if !foundFirstCursor {
			t.Error("initial view should have cursor marker on first claim")
		}

		// Move down
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		view1 := v.View()
		// The cursor position should change
		if view0 == view1 {
			t.Error("view should change after cursor move")
		}

		// Move back up
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		view2 := v.View()
		if view2 != view0 {
			t.Error("view should return to initial state after k")
		}
	})

	t.Run("cursor_bounded", func(t *testing.T) {
		db, dbPath := testDB(t)
		claimsDir := filepath.Join(filepath.Dir(dbPath), "claims")

		store, err := orchestration.NewClaimStore(claimsDir)
		if err != nil {
			t.Fatalf("failed to create claim store: %v", err)
		}
		_, _ = store.Claim("REQ-CLI-001", "agent-alpha")

		v := NewAgentsView(db, dbPath)

		// Cannot go up from 0
		v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		// Cannot go down past last claim
		for i := 0; i < 10; i++ {
			v.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		}
		// Should not panic
		view := v.View()
		if view == "" {
			t.Error("view should not be empty after excessive navigation")
		}
	})

	t.Run("mixed_stale_and_active", func(t *testing.T) {
		db, dbPath := testDB(t)
		claimsDir := filepath.Join(filepath.Dir(dbPath), "claims")
		_ = os.MkdirAll(claimsDir, 0o755)

		// Write a stale claim
		staleClaim := orchestration.Claim{
			ReqID:     "REQ-CLI-001",
			AgentID:   "agent-old",
			ClaimedAt: time.Now().UTC().Add(-30 * time.Minute),
		}
		data, _ := json.MarshalIndent(&staleClaim, "", "  ")
		_ = os.WriteFile(filepath.Join(claimsDir, "REQ-CLI-001.json"), data, 0o644)

		// Write a fresh claim
		freshClaim := orchestration.Claim{
			ReqID:     "REQ-MCP-001",
			AgentID:   "agent-new",
			ClaimedAt: time.Now().UTC(),
		}
		data, _ = json.MarshalIndent(&freshClaim, "", "  ")
		_ = os.WriteFile(filepath.Join(claimsDir, "REQ-MCP-001.json"), data, 0o644)

		v := NewAgentsView(db, dbPath)
		v.SetSize(120, 30)
		view := v.View()

		if !strings.Contains(view, "2 active") {
			t.Error("header should show '2 active' total count")
		}
		if !strings.Contains(view, "1 stale") {
			t.Error("header should show '1 stale' count")
		}
		if !strings.Contains(view, "2 agents") {
			t.Error("header should show '2 agents' count")
		}
		if !strings.Contains(view, "STALE") {
			t.Error("view should contain STALE for old claim")
		}
	})
}
