package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
)

var kanbanColumns = []database.Status{
	database.StatusNotStarted,
	database.StatusMissing,
	database.StatusPartial,
	database.StatusComplete,
}

// KanbanView shows a Kanban board with status columns.
type KanbanView struct {
	db       *database.Database
	graph    *graph.Graph
	dbPath   string
	columns  [4][]*database.Requirement
	colIdx   int // active column
	rowIdx   [4]int // cursor per column
	width    int
	height   int
}

// NewKanbanView creates a Kanban board view.
func NewKanbanView(db *database.Database, g *graph.Graph, dbPath string) *KanbanView {
	v := &KanbanView{db: db, graph: g, dbPath: dbPath}
	v.refreshData()
	return v
}

func (v *KanbanView) Init() tea.Cmd { return nil }

func (v *KanbanView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "h", "left":
			if v.colIdx > 0 {
				v.colIdx--
			}
		case "l", "right":
			if v.colIdx < 3 {
				v.colIdx++
			}
		case "j", "down":
			col := v.columns[v.colIdx]
			if v.rowIdx[v.colIdx] < len(col)-1 {
				v.rowIdx[v.colIdx]++
			}
		case "k", "up":
			if v.rowIdx[v.colIdx] > 0 {
				v.rowIdx[v.colIdx]--
			}
		case "m":
			return v, v.moveCard()
		}
	}
	return v, nil
}

func (v *KanbanView) View() string {
	colWidth := v.width/4 - 1
	if colWidth < 15 {
		colWidth = 15
	}

	var b strings.Builder

	// Column headers
	for i, status := range kanbanColumns {
		label := fmt.Sprintf(" %s (%d) ", status, len(v.columns[i]))
		if i == v.colIdx {
			label = ">" + label
		} else {
			label = " " + label
		}
		if len(label) > colWidth {
			label = label[:colWidth]
		}
		b.WriteString(fmt.Sprintf("%-*s", colWidth, label))
		if i < 3 {
			b.WriteByte('|')
		}
	}
	b.WriteByte('\n')
	b.WriteString(strings.Repeat("-", v.width))
	b.WriteByte('\n')

	// Cards
	maxRows := 0
	for _, col := range v.columns {
		if len(col) > maxRows {
			maxRows = len(col)
		}
	}
	visRows := v.height - 4
	if visRows < 5 {
		visRows = 10
	}
	if maxRows > visRows {
		maxRows = visRows
	}

	for row := 0; row < maxRows; row++ {
		for i, col := range v.columns {
			cell := strings.Repeat(" ", colWidth)
			if row < len(col) {
				r := col[row]
				cursor := " "
				if i == v.colIdx && row == v.rowIdx[i] {
					cursor = ">"
				}
				blocked := ""
				if v.graph.IsBlocked(r.ReqID) {
					blocked = "[B]"
				}
				text := r.RequirementText
				textW := colWidth - len(r.ReqID) - len(blocked) - 4
				if textW < 5 {
					textW = 5
				}
				if len(text) > textW {
					text = text[:textW-2] + ".."
				}
				card := fmt.Sprintf("%s%s %s%s", cursor, r.ReqID, text, blocked)
				if len(card) > colWidth {
					card = card[:colWidth]
				}
				cell = fmt.Sprintf("%-*s", colWidth, card)
			}
			b.WriteString(cell)
			if i < 3 {
				b.WriteByte('|')
			}
		}
		b.WriteByte('\n')
	}

	b.WriteString(fmt.Sprintf("\n  m:move  h/l:columns  j/k:cards"))
	return b.String()
}

func (v *KanbanView) ShortHelp() []key.Binding { return nil }
func (v *KanbanView) SetSize(w, h int)         { v.width = w; v.height = h }

func (v *KanbanView) Reload(db *database.Database, g *graph.Graph) {
	v.db = db
	v.graph = g
	v.refreshData()
}

// ActiveColumn returns the active column index.
func (v *KanbanView) ActiveColumn() int { return v.colIdx }

func (v *KanbanView) moveCard() tea.Cmd {
	col := v.columns[v.colIdx]
	if len(col) == 0 || v.rowIdx[v.colIdx] >= len(col) {
		return nil
	}
	req := col[v.rowIdx[v.colIdx]]

	// Move to next status column
	nextCol := (v.colIdx + 1) % 4
	newStatus := kanbanColumns[nextCol]

	// Block moving to COMPLETE if upstream deps incomplete
	if newStatus == database.StatusComplete && v.graph.IsBlocked(req.ReqID) {
		return nil
	}

	_ = v.db.Update(req.ReqID, map[string]interface{}{"status": string(newStatus)})
	if v.dbPath != "" {
		_ = v.db.Save(v.dbPath)
	}
	v.graph = graph.NewGraph(v.db)
	v.refreshData()
	return nil
}

func (v *KanbanView) refreshData() {
	for i := range v.columns {
		v.columns[i] = nil
	}
	for _, req := range v.db.All() {
		for i, status := range kanbanColumns {
			if req.Status == status {
				v.columns[i] = append(v.columns[i], req)
				break
			}
		}
	}
	for i := range v.rowIdx {
		if v.rowIdx[i] >= len(v.columns[i]) {
			v.rowIdx[i] = max(0, len(v.columns[i])-1)
		}
	}
}
