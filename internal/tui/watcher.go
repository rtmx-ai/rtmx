package tui

import (
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
)

// FileChangedMsg signals that the database file was modified on disk.
type FileChangedMsg struct {
	ModTime time.Time
}

// WatchDatabaseCmd returns a tea.Cmd that polls for database file changes.
// It checks the file modification time every interval and sends a
// DatabaseReloadedMsg when a change is detected.
func WatchDatabaseCmd(dbPath string, interval time.Duration, lastMod time.Time) tea.Cmd {
	return func() tea.Msg {
		time.Sleep(interval)
		info, err := os.Stat(dbPath)
		if err != nil {
			return nil
		}
		if info.ModTime().After(lastMod) {
			db, err := database.Load(dbPath)
			if err != nil {
				return nil
			}
			return DatabaseReloadedMsg{
				DB:    db,
				Graph: graph.NewGraph(db),
			}
		}
		return FileChangedMsg{ModTime: lastMod}
	}
}

// StartWatching returns a tea.Cmd that begins file watching with the given interval.
func StartWatching(dbPath string, interval time.Duration) tea.Cmd {
	return func() tea.Msg {
		info, err := os.Stat(dbPath)
		if err != nil {
			return nil
		}
		return FileChangedMsg{ModTime: info.ModTime()}
	}
}
