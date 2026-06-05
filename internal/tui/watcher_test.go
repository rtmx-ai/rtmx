package tui

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestTUIFileWatch validates the file watcher for auto-refresh.
// REQ-TUI-006: Auto-refresh on database file changes.
func TestTUIFileWatch(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-006")

	db, dbPath := testDB(t)

	t.Run("start_watching_returns_file_changed_msg", func(t *testing.T) {
		cmd := StartWatching(dbPath, 500*time.Millisecond)
		if cmd == nil {
			t.Fatal("StartWatching should return a cmd")
		}
		msg := cmd()
		fcm, ok := msg.(FileChangedMsg)
		if !ok {
			t.Fatalf("expected FileChangedMsg, got %T", msg)
		}
		if fcm.ModTime.IsZero() {
			t.Error("ModTime should not be zero")
		}
	})

	t.Run("watch_detects_no_change", func(t *testing.T) {
		// Use a very recent lastMod to ensure no change detected
		info, _ := os.Stat(dbPath)
		lastMod := info.ModTime().Add(time.Second) // future of actual modtime
		cmd := WatchDatabaseCmd(dbPath, 1*time.Millisecond, lastMod)
		msg := cmd()
		// Should return FileChangedMsg (no change) not DatabaseReloadedMsg
		if _, ok := msg.(DatabaseReloadedMsg); ok {
			t.Error("should not detect change when lastMod is after file mod time")
		}
		if fcm, ok := msg.(FileChangedMsg); ok {
			if fcm.ModTime != lastMod {
				t.Error("should preserve lastMod time")
			}
		}
	})

	t.Run("watch_detects_change", func(t *testing.T) {
		// Get current mod time
		info, _ := os.Stat(dbPath)
		oldMod := info.ModTime().Add(-2 * time.Second)

		// Modify the file
		_ = db.Update("REQ-MCP-001", map[string]interface{}{"status": "COMPLETE"})
		_ = db.Save(dbPath)

		cmd := WatchDatabaseCmd(dbPath, 1*time.Millisecond, oldMod)
		msg := cmd()
		reload, ok := msg.(DatabaseReloadedMsg)
		if !ok {
			t.Fatalf("expected DatabaseReloadedMsg after file change, got %T", msg)
		}
		if reload.DB == nil {
			t.Error("reloaded DB should not be nil")
		}
		if reload.Graph == nil {
			t.Error("reloaded Graph should not be nil")
		}
	})

	t.Run("watch_handles_missing_file", func(t *testing.T) {
		missingPath := filepath.Join(t.TempDir(), "nonexistent.csv")
		cmd := WatchDatabaseCmd(missingPath, 1*time.Millisecond, time.Time{})
		msg := cmd()
		if msg != nil {
			t.Error("should return nil for missing file")
		}
	})

	t.Run("start_watching_handles_missing_file", func(t *testing.T) {
		missingPath := filepath.Join(t.TempDir(), "nonexistent.csv")
		cmd := StartWatching(missingPath, 500*time.Millisecond)
		msg := cmd()
		if msg != nil {
			t.Error("should return nil for missing file")
		}
	})

	t.Run("app_model_enables_watching", func(t *testing.T) {
		m := NewAppModel(db, dbPath)
		// Init should enable watching
		m.Init()
		if !m.WatchEnabled() {
			t.Error("watching should be enabled after Init with valid dbPath")
		}
	})

	t.Run("file_changed_msg_continues_watching", func(t *testing.T) {
		m := NewAppModel(db, dbPath)
		m.Init()
		now := time.Now()
		updated, cmd := m.Update(FileChangedMsg{ModTime: now})
		app := updated.(*AppModel)
		if app.lastModTime != now {
			t.Error("lastModTime should be updated")
		}
		if cmd == nil {
			t.Error("should return a cmd to continue watching")
		}
	})

	t.Run("database_reloaded_continues_watching", func(t *testing.T) {
		m := NewAppModel(db, dbPath)
		m.Init()
		newDB, _ := database.Load(dbPath)
		_, cmd := m.Update(DatabaseReloadedMsg{
			DB:    newDB,
			Graph: nil,
		})
		if cmd == nil {
			t.Error("should return a cmd to continue watching after reload")
		}
	})
}
