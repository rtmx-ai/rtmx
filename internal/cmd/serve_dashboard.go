package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/dashboard"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
)

// statusData holds template data for the status partial.
type statusData struct {
	Complete   int
	Partial    int
	Missing    int
	Total      int
	Percent    int
	Categories []categoryData
}

type categoryData struct {
	Name     string
	Complete int
	Partial  int
	Missing  int
	Total    int
	Percent  int
}

// requirementsListData holds template data for the requirements partial.
type requirementsListData struct {
	Requirements []reqRowData
	Categories   []string
	HasMore      bool
	NextPage     int
}

type reqRowData struct {
	ReqID       string
	Status      string
	StatusClass string
	Priority    string
	Category    string
	Description string
	Effort      string
}

// detailData holds template data for the detail partial.
type detailData struct {
	ReqID       string
	Status      string
	StatusClass string
	Priority    string
	Category    string
	Phase       int
	Effort      string
	Description string
	Assignee    string
	Notes       string
	Upstream    []depData
	Downstream  []depData
}

type depData struct {
	ReqID       string
	Status      string
	StatusClass string
}

// graphPartialData holds template data for the graph partial.
type graphPartialData struct {
	NodeCount  int
	EdgeCount  int
	WebCount   int
	Categories []string
	GraphJSON  string
}

// kanbanData holds template data for the kanban partial.
type kanbanData struct {
	Columns []kanbanColumnData
}

type kanbanColumnData struct {
	Status        string
	Label         string
	Count         int
	Cards         []kanbanCardData
}

type kanbanCardData struct {
	ReqID         string
	Status        string
	StatusClass   string
	Priority      string
	PriorityClass string
	Description   string
	Blocked       bool
}

// releasePageData holds template data for the releases partial.
type releasePageData struct {
	Versions []releaseVersionData
}

type releaseVersionData struct {
	Version      string
	Complete     int
	Total        int
	Percent      int
	GatePass     bool
	Requirements []reqRowData
}

// healthData holds template data for the health partial.
type healthData struct {
	Percent      int
	Velocity     string
	BlockedCount int
	Checks       []healthCheck
}

type healthCheck struct {
	Name   string
	Pass   bool
	Detail string
}

// agentPageData holds template data for the agents partial.
type agentPageData struct {
	ActiveCount int
	StaleCount  int
	AgentCount  int
	Claims      []agentClaimData
}

type agentClaimData struct {
	ReqID       string
	AgentID     string
	ClaimedAt   string
	Stale       bool
	Description string
}

func statusClassFor(status database.Status) string {
	switch status {
	case database.StatusComplete:
		return "complete"
	case database.StatusPartial:
		return "partial"
	case database.StatusMissing:
		return "missing"
	default:
		return "not-started"
	}
}

func priorityClassFor(pri database.Priority) string {
	switch pri {
	case database.PriorityP0:
		return "p0"
	case database.PriorityHigh:
		return "high"
	case database.PriorityMedium:
		return "med"
	default:
		return "low"
	}
}

// registerDashboardRoutes adds the SPA and partial routes to the mux.
func registerDashboardRoutes(mux *http.ServeMux, db *database.Database, g *graph.Graph) {
	// Serve static assets
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(dashboard.StaticFS())))

	// SPA shell -- serves layout for all non-API, non-partial routes
	mux.HandleFunc("/app", handleDashboardShell(db))
	mux.HandleFunc("/requirements", handleDashboardShell(db))
	mux.HandleFunc("/graph", handleDashboardShell(db))
	mux.HandleFunc("/kanban", handleDashboardShell(db))
	mux.HandleFunc("/releases", handleDashboardShell(db))
	mux.HandleFunc("/health", handleDashboardShell(db))
	mux.HandleFunc("/agents", handleDashboardShell(db))

	// Partials (htmx targets)
	mux.HandleFunc("/partials/status", handlePartialStatus(db))
	mux.HandleFunc("/partials/requirements", handlePartialRequirements(db))
	mux.HandleFunc("/partials/detail/", handlePartialDetail(db, g))
	mux.HandleFunc("/partials/graph", handlePartialGraph(db, g))
	mux.HandleFunc("/partials/kanban", handlePartialKanban(db, g))
	mux.HandleFunc("/partials/releases", handlePartialReleases(db))
	mux.HandleFunc("/partials/health", handlePartialHealth(db, g))
	mux.HandleFunc("/partials/agents", handlePartialAgents(db))
}

func handleDashboardShell(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Determine which page content to pre-render
		page := strings.TrimPrefix(r.URL.Path, "/")
		if page == "" || page == "app" {
			page = "status"
		}

		// Render initial content
		var content bytes.Buffer
		switch page {
		case "status":
			_ = dashboard.RenderPartial(&content, "status", buildStatusData(db))
		default:
			content.WriteString("<p>Loading...</p>")
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = dashboard.RenderLayout(w, dashboard.LayoutData{
			Title:      strings.Title(page),
			ActivePage: page,
			Content:    htmlFromString(content.String()),
		})
	}
}

func handlePartialStatus(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = dashboard.RenderPartial(w, "status", buildStatusData(db))
	}
}

func handlePartialRequirements(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqs := db.All()
		// Collect categories
		catSet := make(map[string]bool)
		for _, req := range reqs {
			catSet[req.Category] = true
		}
		cats := make([]string, 0, len(catSet))
		for c := range catSet {
			cats = append(cats, c)
		}

		// Apply filters
		status := r.URL.Query().Get("status")
		category := r.URL.Query().Get("category")
		search := r.URL.Query().Get("search")

		var filtered []*database.Requirement
		for _, req := range reqs {
			if status != "" && string(req.Status) != status {
				continue
			}
			if category != "" && req.Category != category {
				continue
			}
			if search != "" && !strings.Contains(strings.ToLower(req.ReqID+req.RequirementText), strings.ToLower(search)) {
				continue
			}
			filtered = append(filtered, req)
		}

		rows := make([]reqRowData, 0, len(filtered))
		for _, req := range filtered {
			rows = append(rows, reqRowData{
				ReqID:       req.ReqID,
				Status:      string(req.Status),
				StatusClass: statusClassFor(req.Status),
				Priority:    string(req.Priority),
				Category:    req.Category,
				Description: req.RequirementText,
				Effort:      fmt.Sprintf("%.1fw", req.EffortWeeks),
			})
		}

		data := requirementsListData{
			Requirements: rows,
			Categories:   cats,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = dashboard.RenderPartial(w, "requirements", data)
	}
}

func handlePartialDetail(db *database.Database, g *graph.Graph) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqID := strings.TrimPrefix(r.URL.Path, "/partials/detail/")
		req := db.Get(reqID)
		if req == nil {
			http.NotFound(w, r)
			return
		}

		upstream := g.TransitiveDependencies(reqID)
		downstream := g.TransitiveDependents(reqID)

		upDeps := make([]depData, 0, len(upstream))
		for _, id := range upstream {
			s := database.StatusMissing
			if r := db.Get(id); r != nil {
				s = r.Status
			}
			upDeps = append(upDeps, depData{ReqID: id, Status: string(s), StatusClass: statusClassFor(s)})
		}
		downDeps := make([]depData, 0, len(downstream))
		for _, id := range downstream {
			s := database.StatusMissing
			if r := db.Get(id); r != nil {
				s = r.Status
			}
			downDeps = append(downDeps, depData{ReqID: id, Status: string(s), StatusClass: statusClassFor(s)})
		}

		data := detailData{
			ReqID:       req.ReqID,
			Status:      string(req.Status),
			StatusClass: statusClassFor(req.Status),
			Priority:    string(req.Priority),
			Category:    req.Category,
			Phase:       req.Phase,
			Effort:      fmt.Sprintf("%.1f", req.EffortWeeks),
			Description: req.RequirementText,
			Assignee:    req.Assignee,
			Notes:       req.Notes,
			Upstream:    upDeps,
			Downstream:  downDeps,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = dashboard.RenderPartial(w, "detail", data)
	}
}

func handlePartialGraph(db *database.Database, g *graph.Graph) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stats := g.Statistics()
		nodeCount, _ := stats["nodes"].(int)
		edgeCount, _ := stats["edges"].(int)

		catSet := make(map[string]bool)
		for _, req := range db.All() {
			catSet[req.Category] = true
		}
		cats := make([]string, 0, len(catSet))
		for c := range catSet {
			cats = append(cats, c)
		}

		// Build JSON for D3
		type gNode struct {
			ID     string `json:"id"`
			Status string `json:"status"`
			Group  string `json:"group"`
		}
		type gEdge struct {
			Source string `json:"source"`
			Target string `json:"target"`
		}
		type gJSON struct {
			Nodes []gNode `json:"nodes"`
			Edges []gEdge `json:"edges"`
		}
		gData := gJSON{}
		for _, req := range db.All() {
			gData.Nodes = append(gData.Nodes, gNode{
				ID:     req.ReqID,
				Status: string(req.Status),
				Group:  req.Category,
			})
			for dep := range req.Dependencies {
				gData.Edges = append(gData.Edges, gEdge{Source: dep, Target: req.ReqID})
			}
		}
		graphJSON, _ := json.Marshal(gData)

		data := graphPartialData{
			NodeCount:  nodeCount,
			EdgeCount:  edgeCount,
			WebCount:   len(g.DetectWebs()),
			Categories: cats,
			GraphJSON:  string(graphJSON),
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = dashboard.RenderPartial(w, "graph", data)
	}
}

func handlePartialKanban(db *database.Database, g *graph.Graph) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		columns := []struct {
			status database.Status
			label  string
		}{
			{database.StatusNotStarted, "Not Started"},
			{database.StatusMissing, "Missing"},
			{database.StatusPartial, "Partial"},
			{database.StatusComplete, "Complete"},
		}

		data := kanbanData{}
		for _, col := range columns {
			kc := kanbanColumnData{
				Status: string(col.status),
				Label:  col.label,
			}
			for _, req := range db.All() {
				if req.Status != col.status {
					continue
				}
				desc := req.RequirementText
				if len(desc) > 60 {
					desc = desc[:57] + "..."
				}
				kc.Cards = append(kc.Cards, kanbanCardData{
					ReqID:         req.ReqID,
					Status:        string(req.Status),
					StatusClass:   statusClassFor(req.Status),
					Priority:      string(req.Priority),
					PriorityClass: priorityClassFor(req.Priority),
					Description:   desc,
					Blocked:       g.IsBlocked(req.ReqID),
				})
			}
			kc.Count = len(kc.Cards)
			data.Columns = append(data.Columns, kc)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = dashboard.RenderPartial(w, "kanban", data)
	}
}

func handlePartialReleases(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Group by version/sprint
		versions := make(map[string][]*database.Requirement)
		var versionOrder []string
		for _, req := range db.All() {
			v := req.Sprint
			if v == "" {
				continue
			}
			if _, exists := versions[v]; !exists {
				versionOrder = append(versionOrder, v)
			}
			versions[v] = append(versions[v], req)
		}

		data := releasePageData{}
		for _, v := range versionOrder {
			reqs := versions[v]
			complete := 0
			rows := make([]reqRowData, 0, len(reqs))
			for _, req := range reqs {
				if req.IsComplete() {
					complete++
				}
				rows = append(rows, reqRowData{
					ReqID:       req.ReqID,
					Status:      string(req.Status),
					StatusClass: statusClassFor(req.Status),
					Priority:    string(req.Priority),
					Description: req.RequirementText,
				})
			}
			pct := 0
			if len(reqs) > 0 {
				pct = complete * 100 / len(reqs)
			}
			data.Versions = append(data.Versions, releaseVersionData{
				Version:      v,
				Complete:     complete,
				Total:        len(reqs),
				Percent:      pct,
				GatePass:     complete == len(reqs),
				Requirements: rows,
			})
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = dashboard.RenderPartial(w, "releases", data)
	}
}

func handlePartialHealth(db *database.Database, g *graph.Graph) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqs := db.All()
		total := len(reqs)
		complete := 0
		blocked := 0
		for _, req := range reqs {
			if req.IsComplete() {
				complete++
			}
			if g.IsBlocked(req.ReqID) {
				blocked++
			}
		}
		pct := 0
		if total > 0 {
			pct = complete * 100 / total
		}

		checks := []healthCheck{
			{Name: "No circular dependencies", Pass: len(g.FindCycles()) == 0, Detail: fmt.Sprintf("%d cycles found", len(g.FindCycles()))},
			{Name: "Completion above 50%", Pass: pct >= 50, Detail: fmt.Sprintf("%d%% complete", pct)},
			{Name: "No stale blocked items", Pass: blocked < total/4, Detail: fmt.Sprintf("%d blocked of %d total", blocked, total)},
		}

		data := healthData{
			Percent:      pct,
			Velocity:     "N/A",
			BlockedCount: blocked,
			Checks:       checks,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = dashboard.RenderPartial(w, "health", data)
	}
}

func handlePartialAgents(db *database.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Agents partial with no claims by default
		data := agentPageData{}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = dashboard.RenderPartial(w, "agents", data)
	}
}

func buildStatusData(db *database.Database) statusData {
	reqs := db.All()
	complete, partial, missing := 0, 0, 0
	catMap := make(map[string]*[3]int)
	var catOrder []string
	for _, req := range reqs {
		switch {
		case req.IsComplete():
			complete++
		case req.Status == database.StatusPartial:
			partial++
		default:
			missing++
		}
		counts, exists := catMap[req.Category]
		if !exists {
			counts = &[3]int{}
			catMap[req.Category] = counts
			catOrder = append(catOrder, req.Category)
		}
		switch {
		case req.IsComplete():
			counts[0]++
		case req.Status == database.StatusPartial:
			counts[1]++
		default:
			counts[2]++
		}
	}

	total := len(reqs)
	pct := 0
	if total > 0 {
		pct = complete * 100 / total
	}

	cats := make([]categoryData, 0, len(catOrder))
	for _, name := range catOrder {
		c := catMap[name]
		ct := c[0] + c[1] + c[2]
		cp := 0
		if ct > 0 {
			cp = c[0] * 100 / ct
		}
		cats = append(cats, categoryData{Name: name, Complete: c[0], Partial: c[1], Missing: c[2], Total: ct, Percent: cp})
	}

	return statusData{
		Complete:   complete,
		Partial:    partial,
		Missing:    missing,
		Total:      total,
		Percent:    pct,
		Categories: cats,
	}
}

func htmlFromString(s string) htmlSafe {
	return htmlSafe(s)
}

type htmlSafe = dashboard.HTMLContent
