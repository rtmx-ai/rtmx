package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
)

// DetailView shows full requirement metadata with dependency chains.
type DetailView struct {
	db     *database.Database
	graph  *graph.Graph
	reqID  string
	scroll int
	width  int
	height int
}

// NewDetailView creates a requirement detail pane.
func NewDetailView(db *database.Database, g *graph.Graph) *DetailView {
	return &DetailView{db: db, graph: g}
}

func (v *DetailView) Init() tea.Cmd { return nil }

func (v *DetailView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			v.scroll++
		case "k", "up":
			if v.scroll > 0 {
				v.scroll--
			}
		case "esc", "escape", "backspace":
			v.reqID = ""
			v.scroll = 0
		}
	}
	return v, nil
}

func (v *DetailView) View() string {
	if v.reqID == "" {
		return "  Select a requirement to view details. (Enter from Status tab)"
	}

	req := v.db.Get(v.reqID)
	if req == nil {
		return fmt.Sprintf("  Requirement %s not found.", v.reqID)
	}

	var b strings.Builder

	// Header
	b.WriteString(fmt.Sprintf("  %s\n", req.ReqID))
	b.WriteString(strings.Repeat("-", min(v.width, 60)))
	b.WriteByte('\n')

	// Metadata fields
	b.WriteString(fmt.Sprintf("  Category:    %s\n", req.Category))
	b.WriteString(fmt.Sprintf("  Status:      %s\n", req.Status))
	b.WriteString(fmt.Sprintf("  Priority:    %s\n", req.Priority))
	b.WriteString(fmt.Sprintf("  Phase:       %d\n", req.Phase))
	b.WriteString(fmt.Sprintf("  Effort:      %.1f weeks\n", req.EffortWeeks))
	if req.Assignee != "" {
		b.WriteString(fmt.Sprintf("  Assignee:    %s\n", req.Assignee))
	}
	if req.Sprint != "" {
		b.WriteString(fmt.Sprintf("  Sprint:      %s\n", req.Sprint))
	}
	b.WriteByte('\n')

	// Description
	b.WriteString(fmt.Sprintf("  Description: %s\n", req.RequirementText))
	if req.TargetValue != "" {
		b.WriteString(fmt.Sprintf("  Target:      %s\n", req.TargetValue))
	}
	if req.Notes != "" {
		b.WriteString(fmt.Sprintf("  Notes:       %s\n", req.Notes))
	}
	b.WriteByte('\n')

	// Dependencies
	upstream := v.graph.TransitiveDependencies(v.reqID)
	downstream := v.graph.TransitiveDependents(v.reqID)

	b.WriteString(fmt.Sprintf("  Upstream Dependencies (%d):\n", len(upstream)))
	if len(upstream) == 0 {
		b.WriteString("    (none)\n")
	}
	for _, id := range upstream {
		status := "?"
		if r := v.db.Get(id); r != nil {
			status = string(r.Status)
		}
		b.WriteString(fmt.Sprintf("    %s [%s]\n", id, status))
	}
	b.WriteByte('\n')

	b.WriteString(fmt.Sprintf("  Downstream Dependents (%d):\n", len(downstream)))
	if len(downstream) == 0 {
		b.WriteString("    (none)\n")
	}
	for _, id := range downstream {
		status := "?"
		if r := v.db.Get(id); r != nil {
			status = string(r.Status)
		}
		b.WriteString(fmt.Sprintf("    %s [%s]\n", id, status))
	}

	// Blocked status
	if v.graph.IsBlocked(v.reqID) {
		b.WriteString("\n  [BLOCKED] -- upstream dependencies incomplete\n")
	}

	return b.String()
}

func (v *DetailView) ShortHelp() []key.Binding { return nil }
func (v *DetailView) SetSize(w, h int)         { v.width = w; v.height = h }

func (v *DetailView) Reload(db *database.Database, g *graph.Graph) {
	v.db = db
	if g != nil {
		v.graph = g
	} else {
		v.graph = graph.NewGraph(db)
	}
}

// SetReqID sets the requirement to display.
func (v *DetailView) SetReqID(id string) {
	v.reqID = id
	v.scroll = 0
}

// ReqID returns the currently displayed requirement ID.
func (v *DetailView) ReqID() string {
	return v.reqID
}
