// Package tui provides the interactive terminal UI for RTMX using Bubble Tea.
package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
	"github.com/rtmx-ai/rtmx/internal/tui/views"
)

// Tab indices
const (
	TabStatus  = 0
	TabBacklog = 1
	TabGraph   = 2
	TabKanban  = 3
	TabAgents  = 4
	TabCount   = 5
)

// TabNames are the display labels for each tab.
var TabNames = [TabCount]string{"Status", "Backlog", "Graph", "Kanban", "Agents"}

// DatabaseReloadedMsg signals that the database was reloaded from disk.
type DatabaseReloadedMsg struct {
	DB    *database.Database
	Graph *graph.Graph
}

// AppModel is the top-level Bubble Tea model.
type AppModel struct {
	db        *database.Database
	dbPath    string
	graph     *graph.Graph
	keys      KeyMap
	activeTab int
	views     [TabCount]tea.Model
	width     int
	height    int
	showHelp     bool
	quitting     bool
	watchEnabled bool
	lastModTime  time.Time
}

// viewSizer is implemented by views that accept resize.
type viewSizer interface {
	SetSize(width, height int)
}

// viewReloader is implemented by views that accept data reload.
type viewReloader interface {
	Reload(db *database.Database, g *graph.Graph)
}

// NewAppModel creates a new TUI application model.
func NewAppModel(db *database.Database, dbPath string) *AppModel {
	g := graph.NewGraph(db)
	m := &AppModel{
		db:     db,
		dbPath: dbPath,
		graph:  g,
		keys:   DefaultKeyMap(),
		width:  80,
		height: 24,
	}
	m.views[TabStatus] = views.NewStatusView(db, g)
	m.views[TabBacklog] = views.NewBacklogView(db, g)
	m.views[TabGraph] = views.NewGraphView(db, g)
	m.views[TabKanban] = views.NewKanbanView(db, g, dbPath)
	m.views[TabAgents] = views.NewAgentsView(db, dbPath)
	return m
}

// ActiveTab returns the currently active tab index.
func (m *AppModel) ActiveTab() int {
	return m.activeTab
}

// Quitting returns whether the app is shutting down.
func (m *AppModel) Quitting() bool {
	return m.quitting
}

// ShowHelp returns whether the help overlay is visible.
func (m *AppModel) ShowHelp() bool {
	return m.showHelp
}

// WatchEnabled returns whether file watching is active.
func (m *AppModel) WatchEnabled() bool {
	return m.watchEnabled
}

// Init implements tea.Model.
func (m *AppModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, v := range m.views {
		if v != nil {
			cmds = append(cmds, v.Init())
		}
	}
	if m.dbPath != "" {
		m.watchEnabled = true
		cmds = append(cmds, StartWatching(m.dbPath, 500*time.Millisecond))
	}
	return tea.Batch(cmds...)
}

// Update implements tea.Model.
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		viewHeight := m.height - 3 // tab bar (1) + status bar (1) + padding (1)
		for _, v := range m.views {
			if s, ok := v.(viewSizer); ok {
				s.SetSize(m.width, viewHeight)
			}
		}
		return m, nil

	case tea.KeyMsg:
		// Global keys (always handled at app level)
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		case key.Matches(msg, m.keys.Tab):
			m.activeTab = (m.activeTab + 1) % TabCount
			return m, nil
		case key.Matches(msg, m.keys.ShiftTab):
			m.activeTab = (m.activeTab - 1 + TabCount) % TabCount
			return m, nil
		case key.Matches(msg, m.keys.Jump1):
			m.activeTab = TabStatus
			return m, nil
		case key.Matches(msg, m.keys.Jump2):
			m.activeTab = TabBacklog
			return m, nil
		case key.Matches(msg, m.keys.Jump3):
			m.activeTab = TabGraph
			return m, nil
		case key.Matches(msg, m.keys.Jump4):
			m.activeTab = TabKanban
			return m, nil
		case key.Matches(msg, m.keys.Jump5):
			m.activeTab = TabAgents
			return m, nil
		case key.Matches(msg, m.keys.Refresh):
			return m, m.reloadDatabase()
		}

	case DatabaseReloadedMsg:
		m.db = msg.DB
		m.graph = msg.Graph
		m.lastModTime = time.Now()
		for _, v := range m.views {
			if r, ok := v.(viewReloader); ok {
				r.Reload(msg.DB, msg.Graph)
			}
		}
		if m.watchEnabled {
			return m, WatchDatabaseCmd(m.dbPath, 500*time.Millisecond, m.lastModTime)
		}
		return m, nil

	case FileChangedMsg:
		m.lastModTime = msg.ModTime
		if m.watchEnabled {
			return m, WatchDatabaseCmd(m.dbPath, 500*time.Millisecond, m.lastModTime)
		}
		return m, nil
	}

	// Delegate to active view
	if v := m.views[m.activeTab]; v != nil {
		newView, cmd := v.Update(msg)
		m.views[m.activeTab] = newView
		return m, cmd
	}

	return m, nil
}

// View implements tea.Model.
func (m *AppModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Tab bar
	b.WriteString(m.renderTabBar())
	b.WriteByte('\n')

	// Active view
	if v := m.views[m.activeTab]; v != nil {
		b.WriteString(v.View())
	}
	b.WriteByte('\n')

	// Status bar
	b.WriteString(m.renderStatusBar())

	return b.String()
}

func (m *AppModel) renderTabBar() string {
	var tabs []string
	for i, name := range TabNames {
		label := fmt.Sprintf(" %d:%s ", i+1, name)
		if i == m.activeTab {
			tabs = append(tabs, StyleTabActive.Render(label))
		} else {
			tabs = append(tabs, StyleTabInactive.Render(label))
		}
	}
	bar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	return lipgloss.NewStyle().Width(m.width).Render(bar)
}

func (m *AppModel) renderStatusBar() string {
	total := len(m.db.All())
	complete := 0
	for _, r := range m.db.All() {
		if r.IsComplete() {
			complete++
		}
	}
	pct := 0.0
	if total > 0 {
		pct = float64(complete) / float64(total) * 100
	}

	left := fmt.Sprintf(" %d/%d (%.0f%%)", complete, total, pct)
	right := " q:quit  ?:help  tab:switch  r:refresh "
	if m.showHelp {
		right = " ?:close help "
	}

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 0 {
		gap = 0
	}
	bar := left + strings.Repeat(" ", gap) + right
	return StyleStatusBar.Width(m.width).Render(bar)
}

func (m *AppModel) reloadDatabase() tea.Cmd {
	return func() tea.Msg {
		db, err := database.Load(m.dbPath)
		if err != nil {
			return nil
		}
		return DatabaseReloadedMsg{
			DB:    db,
			Graph: graph.NewGraph(db),
		}
	}
}
