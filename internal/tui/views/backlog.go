package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
)

// BacklogView shows prioritized incomplete requirements.
type BacklogView struct {
	db     *database.Database
	graph  *graph.Graph
	items  []*database.Requirement
	cursor int
	offset int
	width  int
	height int
}

// NewBacklogView creates a backlog view.
func NewBacklogView(db *database.Database, g *graph.Graph) *BacklogView {
	v := &BacklogView{db: db, graph: g}
	v.refreshData()
	return v
}

func (v *BacklogView) Init() tea.Cmd { return nil }

func (v *BacklogView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if v.cursor < len(v.items)-1 {
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
		}
	}
	return v, nil
}

func (v *BacklogView) View() string {
	if len(v.items) == 0 {
		return "  Backlog is empty -- all requirements complete."
	}

	blocking := v.graph.BlockingAnalysis()
	var b strings.Builder

	fmt.Fprintf(&b, "  Backlog: %d incomplete requirements\n\n", len(v.items))

	header := fmt.Sprintf("  %-16s %-10s %-8s %6s %6s  %s",
		"REQ ID", "STATUS", "PRI", "EFFORT", "BLOCKS", "DESCRIPTION")
	b.WriteString(header)
	b.WriteByte('\n')
	b.WriteString(strings.Repeat("-", min(v.width, len(header)+10)))
	b.WriteByte('\n')

	rows := v.visibleRows()
	end := v.offset + rows
	if end > len(v.items) {
		end = len(v.items)
	}

	for i := v.offset; i < end; i++ {
		r := v.items[i]
		prefix := "  "
		if i == v.cursor {
			prefix = "> "
		}
		blocksCount := blocking[r.ReqID]
		blocked := ""
		if v.graph.IsBlocked(r.ReqID) {
			blocked = " [B]"
		}
		desc := r.RequirementText
		descWidth := v.width - 65
		if descWidth < 10 {
			descWidth = 10
		}
		if len(desc) > descWidth {
			desc = desc[:descWidth-3] + "..."
		}
		line := fmt.Sprintf("%s%-16s %-10s %-8s %5.1fw %5d   %s%s",
			prefix, r.ReqID, r.Status, r.Priority, r.EffortWeeks, blocksCount, desc, blocked)
		b.WriteString(line)
		b.WriteByte('\n')
	}

	return b.String()
}

func (v *BacklogView) ShortHelp() []key.Binding { return nil }
func (v *BacklogView) SetSize(w, h int)         { v.width = w; v.height = h }

func (v *BacklogView) Reload(db *database.Database, g *graph.Graph) {
	v.db = db
	v.graph = g
	v.refreshData()
}

func (v *BacklogView) visibleRows() int {
	rows := v.height - 6
	if rows < 1 {
		rows = 10
	}
	return rows
}

func (v *BacklogView) refreshData() {
	v.items = v.db.Backlog()
	if v.cursor >= len(v.items) {
		v.cursor = max(0, len(v.items)-1)
	}
}
