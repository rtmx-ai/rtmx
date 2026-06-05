// Package views provides the individual view models for the RTMX TUI.
package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
)

// StatusView shows requirements table with sort/filter/search.
type StatusView struct {
	db       *database.Database
	graph    *graph.Graph
	reqs     []*database.Requirement
	cursor   int
	offset   int
	width    int
	height   int
	sortField string
	sortDesc  bool
}

// NewStatusView creates a requirements table view.
func NewStatusView(db *database.Database, g *graph.Graph) *StatusView {
	v := &StatusView{
		db:        db,
		graph:     g,
		sortField: "req_id",
	}
	v.refreshData()
	return v
}

func (v *StatusView) Init() tea.Cmd { return nil }

func (v *StatusView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if v.cursor < len(v.reqs)-1 {
				v.cursor++
				if v.cursor-v.offset >= v.visibleRows() {
					v.offset++
				}
			}
		case "k", "up":
			if v.cursor > 0 {
				v.cursor--
				if v.cursor < v.offset {
					v.offset = v.cursor
				}
			}
		case "s":
			fields := []string{"req_id", "status", "priority", "category", "effort_weeks"}
			for i, f := range fields {
				if f == v.sortField {
					v.sortField = fields[(i+1)%len(fields)]
					break
				}
			}
			v.refreshData()
		case "S":
			v.sortDesc = !v.sortDesc
			v.refreshData()
		}
	}
	return v, nil
}

func (v *StatusView) View() string {
	if len(v.reqs) == 0 {
		return "  No requirements found."
	}

	var b strings.Builder

	// Header
	header := fmt.Sprintf("  %-16s %-10s %-8s %-10s %6s  %s",
		"REQ ID", "STATUS", "PRI", "CATEGORY", "EFFORT", "DESCRIPTION")
	b.WriteString(header)
	b.WriteByte('\n')
	b.WriteString(strings.Repeat("-", min(v.width, len(header)+10)))
	b.WriteByte('\n')

	rows := v.visibleRows()
	end := v.offset + rows
	if end > len(v.reqs) {
		end = len(v.reqs)
	}

	for i := v.offset; i < end; i++ {
		r := v.reqs[i]
		prefix := "  "
		if i == v.cursor {
			prefix = "> "
		}
		desc := r.RequirementText
		descWidth := v.width - 60
		if descWidth < 10 {
			descWidth = 10
		}
		if len(desc) > descWidth {
			desc = desc[:descWidth-3] + "..."
		}
		line := fmt.Sprintf("%s%-16s %-10s %-8s %-10s %5.1fw  %s",
			prefix, r.ReqID, r.Status, r.Priority, r.Category, r.EffortWeeks, desc)
		b.WriteString(line)
		b.WriteByte('\n')
	}

	// Footer
	fmt.Fprintf(&b, "\n  %d of %d requirements | sort: %s %s",
		v.cursor+1, len(v.reqs), v.sortField, sortDir(v.sortDesc))

	return b.String()
}

func (v *StatusView) ShortHelp() []key.Binding { return nil }

func (v *StatusView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

func (v *StatusView) Reload(db *database.Database, g *graph.Graph) {
	v.db = db
	v.graph = g
	v.refreshData()
}

// SelectedReqID returns the currently selected requirement ID.
func (v *StatusView) SelectedReqID() string {
	if v.cursor >= 0 && v.cursor < len(v.reqs) {
		return v.reqs[v.cursor].ReqID
	}
	return ""
}

// Cursor returns the current cursor position.
func (v *StatusView) Cursor() int { return v.cursor }

// SortField returns the current sort field.
func (v *StatusView) SortField() string { return v.sortField }

// SortDesc returns whether sort is descending.
func (v *StatusView) SortDesc() bool { return v.sortDesc }

func (v *StatusView) visibleRows() int {
	rows := v.height - 5 // header, separator, footer, padding
	if rows < 1 {
		rows = 10
	}
	return rows
}

func (v *StatusView) refreshData() {
	v.reqs = v.db.All()
	sortReqs(v.reqs, v.sortField, v.sortDesc)
	if v.cursor >= len(v.reqs) {
		v.cursor = max(0, len(v.reqs)-1)
	}
}

func sortReqs(reqs []*database.Requirement, field string, desc bool) {
	sort.SliceStable(reqs, func(i, j int) bool {
		var less bool
		switch field {
		case "status":
			less = reqs[i].Status.Weight() < reqs[j].Status.Weight()
		case "priority":
			less = reqs[i].Priority.Weight() < reqs[j].Priority.Weight()
		case "category":
			less = reqs[i].Category < reqs[j].Category
		case "effort_weeks":
			less = reqs[i].EffortWeeks < reqs[j].EffortWeeks
		default:
			less = reqs[i].ReqID < reqs[j].ReqID
		}
		if desc {
			return !less
		}
		return less
	})
}

func sortDir(desc bool) string {
	if desc {
		return "desc"
	}
	return "asc"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
