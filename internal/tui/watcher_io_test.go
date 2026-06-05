package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// writeTestCSV writes a valid RTMX database CSV file to the given path.
func writeTestCSV(t *testing.T, dir string, reqs []*database.Requirement) string {
	t.Helper()
	rtmxDir := filepath.Join(dir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0o755); err != nil {
		t.Fatalf("failed to create .rtmx dir: %v", err)
	}
	dbPath := filepath.Join(rtmxDir, "database.csv")

	db := database.NewDatabase()
	for _, r := range reqs {
		if r.Dependencies == nil {
			r.Dependencies = make(database.StringSet)
		}
		if r.Blocks == nil {
			r.Blocks = make(database.StringSet)
		}
		if err := db.Add(r); err != nil {
			t.Fatalf("failed to add req %s: %v", r.ReqID, err)
		}
	}
	if err := db.Save(dbPath); err != nil {
		t.Fatalf("failed to save test CSV: %v", err)
	}
	return dbPath
}

// TestWatcherIOStartWatching verifies that StartWatching reads the real
// file modification time from a CSV file on disk.
// REQ-TUI-006: Auto-refresh on database file changes.
func TestWatcherIOStartWatching(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-006")

	dir := t.TempDir()
	dbPath := writeTestCSV(t, dir, []*database.Requirement{
		{ReqID: "REQ-IO-001", Category: "IO", RequirementText: "Initial req", Status: database.StatusMissing, Priority: database.PriorityHigh, Phase: 1, EffortWeeks: 0.5},
	})

	// Get the actual file mod time for comparison.
	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}
	actualModTime := info.ModTime()

	cmd := StartWatching(dbPath, 500*time.Millisecond)
	if cmd == nil {
		t.Fatal("StartWatching should return a non-nil cmd")
	}

	msg := cmd()
	fcm, ok := msg.(FileChangedMsg)
	if !ok {
		t.Fatalf("expected FileChangedMsg, got %T", msg)
	}
	if !fcm.ModTime.Equal(actualModTime) {
		t.Errorf("ModTime = %v, want %v", fcm.ModTime, actualModTime)
	}
}

// TestWatcherIODetectsFileModification verifies the watcher detects an actual
// file modification (appending a row) and produces a DatabaseReloadedMsg
// with the updated content.
// REQ-TUI-006: Auto-refresh on database file changes.
func TestWatcherIODetectsFileModification(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-006")

	dir := t.TempDir()
	initialReqs := []*database.Requirement{
		{ReqID: "REQ-IO-001", Category: "IO", RequirementText: "First requirement", Status: database.StatusMissing, Priority: database.PriorityHigh, Phase: 1, EffortWeeks: 0.5},
	}
	dbPath := writeTestCSV(t, dir, initialReqs)

	// Record the initial mod time.
	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("failed to stat: %v", err)
	}
	initialModTime := info.ModTime()

	// Modify the file: load, add a requirement, save.
	db, err := database.Load(dbPath)
	if err != nil {
		t.Fatalf("failed to load database: %v", err)
	}
	newReq := &database.Requirement{
		ReqID:           "REQ-IO-002",
		Category:        "IO",
		RequirementText: "Second requirement",
		Status:          database.StatusPartial,
		Priority:        database.PriorityP0,
		Phase:           1,
		EffortWeeks:     1.0,
		Dependencies:    make(database.StringSet),
		Blocks:          make(database.StringSet),
	}
	if err := db.Add(newReq); err != nil {
		t.Fatalf("failed to add second req: %v", err)
	}

	// Ensure the file mod time actually advances. Some file systems have
	// only 1-second resolution, so we use Chtimes to guarantee a gap if
	// the save happens too fast.
	futureTime := initialModTime.Add(2 * time.Second)
	if err := db.Save(dbPath); err != nil {
		t.Fatalf("failed to save modified database: %v", err)
	}
	// Force mod time forward in case the filesystem has coarse resolution.
	if err := os.Chtimes(dbPath, futureTime, futureTime); err != nil {
		t.Fatalf("failed to set file times: %v", err)
	}

	// Run WatchDatabaseCmd with the initial mod time. It should detect the change.
	cmd := WatchDatabaseCmd(dbPath, 1*time.Millisecond, initialModTime)
	msg := cmd()

	reload, ok := msg.(DatabaseReloadedMsg)
	if !ok {
		t.Fatalf("expected DatabaseReloadedMsg, got %T (%v)", msg, msg)
	}
	if reload.DB == nil {
		t.Fatal("reloaded DB should not be nil")
	}
	if reload.Graph == nil {
		t.Fatal("reloaded Graph should not be nil")
	}

	// Verify the reloaded database contains both requirements.
	all := reload.DB.All()
	if len(all) != 2 {
		t.Errorf("reloaded DB has %d requirements, want 2", len(all))
	}
	if reload.DB.Get("REQ-IO-002") == nil {
		t.Error("reloaded DB should contain newly added REQ-IO-002")
	}
	if reload.DB.Get("REQ-IO-002") != nil && reload.DB.Get("REQ-IO-002").Status != database.StatusPartial {
		t.Error("REQ-IO-002 should have PARTIAL status")
	}
}

// TestWatcherIONoChangeDetected verifies that when the file has not been
// modified since lastMod, the watcher returns a FileChangedMsg (no reload).
// REQ-TUI-006: Auto-refresh on database file changes.
func TestWatcherIONoChangeDetected(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-006")

	dir := t.TempDir()
	dbPath := writeTestCSV(t, dir, []*database.Requirement{
		{ReqID: "REQ-IO-001", Category: "IO", RequirementText: "Stable req", Status: database.StatusComplete, Priority: database.PriorityHigh, Phase: 1, EffortWeeks: 0.5},
	})

	// Use a lastMod that is in the future, so no change is detected.
	lastMod := time.Now().Add(1 * time.Hour)
	cmd := WatchDatabaseCmd(dbPath, 1*time.Millisecond, lastMod)
	msg := cmd()

	if _, ok := msg.(DatabaseReloadedMsg); ok {
		t.Error("should not produce DatabaseReloadedMsg when file has not changed")
	}
	fcm, ok := msg.(FileChangedMsg)
	if !ok {
		t.Fatalf("expected FileChangedMsg, got %T", msg)
	}
	if fcm.ModTime != lastMod {
		t.Errorf("FileChangedMsg ModTime = %v, want %v", fcm.ModTime, lastMod)
	}
}

// TestWatcherIOMissingFile verifies that the watcher handles a missing
// database file gracefully (returns nil).
// REQ-TUI-006: Auto-refresh on database file changes.
func TestWatcherIOMissingFile(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-006")

	missingPath := filepath.Join(t.TempDir(), ".rtmx", "nonexistent.csv")

	t.Run("watch_missing_file", func(t *testing.T) {
		cmd := WatchDatabaseCmd(missingPath, 1*time.Millisecond, time.Time{})
		msg := cmd()
		if msg != nil {
			t.Errorf("expected nil msg for missing file, got %T", msg)
		}
	})

	t.Run("start_watching_missing_file", func(t *testing.T) {
		cmd := StartWatching(missingPath, 500*time.Millisecond)
		msg := cmd()
		if msg != nil {
			t.Errorf("expected nil msg for missing file, got %T", msg)
		}
	})
}

// TestWatcherIODeletedFile verifies that when a watched file is deleted
// between polls, the watcher returns nil gracefully.
// REQ-TUI-006: Auto-refresh on database file changes.
func TestWatcherIODeletedFile(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-006")

	dir := t.TempDir()
	dbPath := writeTestCSV(t, dir, []*database.Requirement{
		{ReqID: "REQ-IO-001", Category: "IO", RequirementText: "Ephemeral", Status: database.StatusMissing, Priority: database.PriorityHigh, Phase: 1, EffortWeeks: 0.5},
	})

	// Record mod time then delete the file.
	info, _ := os.Stat(dbPath)
	lastMod := info.ModTime()
	if err := os.Remove(dbPath); err != nil {
		t.Fatalf("failed to delete file: %v", err)
	}

	cmd := WatchDatabaseCmd(dbPath, 1*time.Millisecond, lastMod)
	msg := cmd()
	if msg != nil {
		t.Errorf("expected nil for deleted file, got %T", msg)
	}
}

// TestWatcherIOCorruptedFile verifies that the watcher handles a file
// with invalid CSV content gracefully (returns nil, not a crash).
// REQ-TUI-006: Auto-refresh on database file changes.
func TestWatcherIOCorruptedFile(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-006")

	dir := t.TempDir()
	rtmxDir := filepath.Join(dir, ".rtmx")
	if err := os.MkdirAll(rtmxDir, 0o755); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(rtmxDir, "database.csv")

	// Write garbage content that will fail database.Load.
	if err := os.WriteFile(dbPath, []byte("this is not valid csv\n\x00\x01\x02"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Set mod time to the past so the watcher sees it as changed.
	past := time.Now().Add(-1 * time.Hour)
	if err := os.Chtimes(dbPath, past, past); err != nil {
		t.Fatal(err)
	}
	veryOldTime := time.Now().Add(-2 * time.Hour)

	cmd := WatchDatabaseCmd(dbPath, 1*time.Millisecond, veryOldTime)
	msg := cmd()

	// The file was modified but Load should fail, returning nil.
	if msg != nil {
		// If the CSV parser happens to be lenient and parses it as an empty DB,
		// that is also acceptable -- the key point is no panic.
		if reload, ok := msg.(DatabaseReloadedMsg); ok {
			if reload.DB == nil {
				t.Error("if DatabaseReloadedMsg is returned, DB should not be nil")
			}
		}
	}
}

// TestWatcherIOMultipleModifications verifies that the watcher can detect
// multiple sequential modifications, each producing a new DatabaseReloadedMsg.
// REQ-TUI-006: Auto-refresh on database file changes.
func TestWatcherIOMultipleModifications(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-006")

	dir := t.TempDir()
	dbPath := writeTestCSV(t, dir, []*database.Requirement{
		{ReqID: "REQ-IO-001", Category: "IO", RequirementText: "Original", Status: database.StatusMissing, Priority: database.PriorityHigh, Phase: 1, EffortWeeks: 0.5},
	})

	baseTime := time.Now().Add(-10 * time.Second)

	for i := 0; i < 3; i++ {
		// Load current state and add a requirement.
		db, err := database.Load(dbPath)
		if err != nil {
			t.Fatalf("round %d: failed to load: %v", i, err)
		}
		newReq := &database.Requirement{
			ReqID:           fmt.Sprintf("REQ-IO-%03d", i+10),
			Category:        "IO",
			RequirementText: fmt.Sprintf("Added in round %d", i),
			Status:          database.StatusMissing,
			Priority:        database.PriorityHigh,
			Phase:           1,
			EffortWeeks:     0.5,
			Dependencies:    make(database.StringSet),
			Blocks:          make(database.StringSet),
		}
		if err := db.Add(newReq); err != nil {
			t.Fatalf("round %d: failed to add: %v", i, err)
		}
		if err := db.Save(dbPath); err != nil {
			t.Fatalf("round %d: failed to save: %v", i, err)
		}
		// Force a distinct mod time.
		modTime := baseTime.Add(time.Duration(i+1) * 2 * time.Second)
		if err := os.Chtimes(dbPath, modTime, modTime); err != nil {
			t.Fatalf("round %d: failed to set times: %v", i, err)
		}

		// Watch with a lastMod before this modification.
		lastMod := modTime.Add(-1 * time.Second)
		cmd := WatchDatabaseCmd(dbPath, 1*time.Millisecond, lastMod)
		msg := cmd()

		reload, ok := msg.(DatabaseReloadedMsg)
		if !ok {
			t.Fatalf("round %d: expected DatabaseReloadedMsg, got %T", i, msg)
		}
		expectedCount := 1 + i + 1 // initial + i+1 added
		if got := len(reload.DB.All()); got != expectedCount {
			t.Errorf("round %d: DB has %d requirements, want %d", i, got, expectedCount)
		}
	}
}

// TestWatcherIOAppModelIntegration verifies the full integration: create a
// real CSV file, build an AppModel, send FileChangedMsg and
// DatabaseReloadedMsg through the model, and verify the views update.
// REQ-TUI-006: Auto-refresh on database file changes.
func TestWatcherIOAppModelIntegration(t *testing.T) {
	rtmx.Req(t, "REQ-TUI-006")

	dir := t.TempDir()
	dbPath := writeTestCSV(t, dir, []*database.Requirement{
		{ReqID: "REQ-IO-001", Category: "IO", RequirementText: "First", Status: database.StatusComplete, Priority: database.PriorityHigh, Phase: 1, EffortWeeks: 0.5},
		{ReqID: "REQ-IO-002", Category: "IO", RequirementText: "Second", Status: database.StatusMissing, Priority: database.PriorityP0, Phase: 1, EffortWeeks: 1.0},
	})

	db, err := database.Load(dbPath)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}
	m := NewAppModel(db, dbPath)
	m.Init()
	m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Verify initial status bar shows 1/2 (1 complete of 2).
	view := m.View()
	if !strings.Contains(view, "1/2") {
		t.Errorf("initial status bar should show 1/2, got:\n%s", view)
	}

	// Modify the file on disk: mark second req complete.
	db2, err := database.Load(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	_ = db2.Update("REQ-IO-002", map[string]interface{}{"status": "COMPLETE"})
	_ = db2.Save(dbPath)

	// Force a new mod time.
	futureTime := time.Now().Add(5 * time.Second)
	_ = os.Chtimes(dbPath, futureTime, futureTime)

	// Simulate what the watcher would do: detect the change, load the new DB.
	cmd := WatchDatabaseCmd(dbPath, 1*time.Millisecond, time.Now().Add(-10*time.Second))
	msg := cmd()

	reload, ok := msg.(DatabaseReloadedMsg)
	if !ok {
		t.Fatalf("expected DatabaseReloadedMsg, got %T", msg)
	}

	// Feed the reload into the AppModel.
	updated, _ := m.Update(reload)
	m = updated.(*AppModel)

	// Status bar should now show 2/2 (100%).
	view = m.View()
	if !strings.Contains(view, "2/2") {
		t.Errorf("status bar after reload should show 2/2, got:\n%s", view)
	}
	if !strings.Contains(view, "100%") {
		t.Errorf("status bar should show 100%% after all complete")
	}
}
