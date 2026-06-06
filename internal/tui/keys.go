package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all key bindings for the TUI application.
type KeyMap struct {
	Quit      key.Binding
	Help      key.Binding
	Tab       key.Binding
	ShiftTab  key.Binding
	Up        key.Binding
	Down      key.Binding
	Enter     key.Binding
	Escape    key.Binding
	Refresh   key.Binding
	Filter    key.Binding
	Sort      key.Binding
	SortDir   key.Binding
	Left      key.Binding
	Right     key.Binding
	Move      key.Binding
	Jump1     key.Binding
	Jump2     key.Binding
	Jump3     key.Binding
	Jump4     key.Binding
	Jump5     key.Binding
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit:     key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		Help:     key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next view")),
		ShiftTab: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev view")),
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("k/up", "up")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("j/down", "down")),
		Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		Escape:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Refresh:  key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		Filter:   key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
		Sort:     key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "sort")),
		SortDir:  key.NewBinding(key.WithKeys("S"), key.WithHelp("S", "sort dir")),
		Left:     key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("h/left", "left")),
		Right:    key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("l/right", "right")),
		Move:     key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "move")),
		Jump1:    key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "status")),
		Jump2:    key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "backlog")),
		Jump3:    key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "graph")),
		Jump4:    key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "kanban")),
		Jump5:    key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "agents")),
	}
}
