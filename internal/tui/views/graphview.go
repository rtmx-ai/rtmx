package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
)

// GraphView shows an ASCII dependency graph.
type GraphView struct {
	db     *database.Database
	graph  *graph.Graph
	layers [][]string
	cursor int // current layer
	width  int
	height int
}

// NewGraphView creates a dependency graph view.
func NewGraphView(db *database.Database, g *graph.Graph) *GraphView {
	v := &GraphView{db: db, graph: g}
	v.refreshData()
	return v
}

func (v *GraphView) Init() tea.Cmd { return nil }

func (v *GraphView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if v.cursor < len(v.layers)-1 {
				v.cursor++
			}
		case "k", "up":
			if v.cursor > 0 {
				v.cursor--
			}
		}
	}
	return v, nil
}

func (v *GraphView) View() string {
	if len(v.layers) == 0 {
		return "  No dependency graph to display."
	}

	stats := v.graph.Statistics()
	cp := v.graph.CriticalPath()
	cpSet := make(map[string]bool)
	for _, id := range cp {
		cpSet[id] = true
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("  Dependency Graph: %d nodes, %d edges, %d layers\n\n",
		stats["nodes"], stats["edges"], len(v.layers)))

	for i, layer := range v.layers {
		prefix := "  "
		if i == v.cursor {
			prefix = "> "
		}
		b.WriteString(fmt.Sprintf("%sLayer %d: ", prefix, i))

		var nodeStrs []string
		for _, id := range layer {
			label := id
			if req := v.db.Get(id); req != nil {
				status := string(req.Status)
				mark := ""
				if cpSet[id] {
					mark = "*"
				}
				label = fmt.Sprintf("[%s%s %s]", id, mark, status[:3])
			}
			nodeStrs = append(nodeStrs, label)
		}
		b.WriteString(strings.Join(nodeStrs, "  "))
		b.WriteByte('\n')
	}

	b.WriteString(fmt.Sprintf("\n  * = critical path | %d independent webs", len(v.graph.DetectWebs())))

	return b.String()
}

func (v *GraphView) ShortHelp() []key.Binding { return nil }
func (v *GraphView) SetSize(w, h int)         { v.width = w; v.height = h }

func (v *GraphView) Reload(db *database.Database, g *graph.Graph) {
	v.db = db
	if g != nil {
		v.graph = g
	} else {
		v.graph = graph.NewGraph(db)
	}
	v.refreshData()
}

func (v *GraphView) refreshData() {
	if v.graph == nil {
		v.layers = nil
		return
	}
	v.layers = v.graph.Layers()
}
