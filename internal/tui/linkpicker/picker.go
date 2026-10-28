// Package linkpicker provides an interactive issue-store provider picker
// for `rtmx link` / flagless bootstrap (REQ-GO-089 / REQ-MONO-044b).
package linkpicker

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Choice is one selectable issue-store / bootstrap source.
type Choice struct {
	ID    string // github|jira|asana|monday|gitlab|tests|skip
	Label string
}

// DefaultChoices is the ordered provider list for the link picker.
var DefaultChoices = []Choice{
	{ID: "github", Label: "GitHub"},
	{ID: "jira", Label: "Jira"},
	{ID: "asana", Label: "Asana"},
	{ID: "monday", Label: "Monday"},
	{ID: "gitlab", Label: "GitLab"},
	{ID: "tests", Label: "Tests"},
	{ID: "skip", Label: "Skip"},
}

// Model is a Bubble Tea list that returns a chosen provider ID.
type Model struct {
	choices []Choice
	cursor  int
	chosen  string
	done    bool
	title   string
}

// New builds a picker. hint is a provider ID ("github"/"gitlab") to preselect;
// empty hint leaves the cursor on the first item.
func New(hint string) Model {
	m := Model{
		choices: append([]Choice(nil), DefaultChoices...),
		title:   "Link an issue store",
	}
	hint = strings.ToLower(strings.TrimSpace(hint))
	if hint != "" {
		for i, c := range m.choices {
			if c.ID == hint {
				m.cursor = i
				break
			}
		}
	}
	return m
}

// Chosen returns the selected provider ID after quit, or "" if cancelled.
func (m Model) Chosen() string { return m.chosen }

// Done reports whether the user confirmed or cancelled.
func (m Model) Done() bool { return m.done }

// Cursor returns the highlighted index (for tests).
func (m Model) Cursor() int { return m.cursor }

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.done = true
			m.chosen = ""
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.chosen = m.choices[m.cursor].ID
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.done {
		return ""
	}
	var b strings.Builder
	b.WriteString(m.title)
	b.WriteString("\n\n")
	for i, c := range m.choices {
		cursor := "  "
		if i == m.cursor {
			cursor = "> "
		}
		b.WriteString(cursor)
		b.WriteString(c.Label)
		b.WriteByte('\n')
	}
	b.WriteString("\n↑/↓ move · enter confirm · q cancel\n")
	return b.String()
}
