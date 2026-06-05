package views

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
	"github.com/rtmx-ai/rtmx/internal/orchestration"
)

// AgentsView shows active agent claims and staleness.
type AgentsView struct {
	db        *database.Database
	dbPath    string
	claims    []*orchestration.Claim
	cursor    int
	width     int
	height    int
	staleTime time.Duration
}

// NewAgentsView creates an agent activity monitor view.
func NewAgentsView(db *database.Database, dbPath string) *AgentsView {
	v := &AgentsView{
		db:        db,
		dbPath:    dbPath,
		staleTime: 15 * time.Minute,
	}
	v.refreshData()
	return v
}

func (v *AgentsView) Init() tea.Cmd { return nil }

func (v *AgentsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if v.cursor < len(v.claims)-1 {
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

func (v *AgentsView) View() string {
	var b strings.Builder

	now := time.Now().UTC()
	staleCount := 0
	agents := make(map[string]bool)
	for _, c := range v.claims {
		if now.Sub(c.ClaimedAt) > v.staleTime {
			staleCount++
		}
		agents[c.AgentID] = true
	}

	b.WriteString(fmt.Sprintf("  Agent Claims: %d active, %d stale, %d agents\n\n",
		len(v.claims), staleCount, len(agents)))

	if len(v.claims) == 0 {
		b.WriteString("  No active claims.")
		return b.String()
	}

	header := fmt.Sprintf("  %-16s %-20s %-20s %-8s  %s",
		"REQ ID", "AGENT", "CLAIMED AT", "STATUS", "DESCRIPTION")
	b.WriteString(header)
	b.WriteByte('\n')
	b.WriteString(strings.Repeat("-", min(v.width, len(header)+10)))
	b.WriteByte('\n')

	for i, c := range v.claims {
		prefix := "  "
		if i == v.cursor {
			prefix = "> "
		}
		status := "active"
		if now.Sub(c.ClaimedAt) > v.staleTime {
			status = "STALE"
		}
		desc := ""
		if r := v.db.Get(c.ReqID); r != nil {
			desc = r.RequirementText
			descW := v.width - 75
			if descW < 10 {
				descW = 10
			}
			if len(desc) > descW {
				desc = desc[:descW-3] + "..."
			}
		}
		line := fmt.Sprintf("%s%-16s %-20s %-20s %-8s  %s",
			prefix, c.ReqID, c.AgentID,
			c.ClaimedAt.Format("2006-01-02 15:04"),
			status, desc)
		b.WriteString(line)
		b.WriteByte('\n')
	}

	return b.String()
}

func (v *AgentsView) ShortHelp() []key.Binding { return nil }
func (v *AgentsView) SetSize(w, h int)         { v.width = w; v.height = h }

func (v *AgentsView) Reload(db *database.Database, _ *graph.Graph) {
	v.db = db
	v.refreshData()
}

func (v *AgentsView) refreshData() {
	if v.dbPath == "" {
		v.claims = nil
		return
	}
	claimsDir := filepath.Join(filepath.Dir(v.dbPath), "claims")
	store, err := orchestration.NewClaimStore(claimsDir)
	if err != nil {
		v.claims = nil
		return
	}
	claims, err := store.List()
	if err != nil {
		v.claims = nil
		return
	}
	v.claims = claims
}
