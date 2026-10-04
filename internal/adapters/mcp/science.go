package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
	"github.com/rtmx-ai/rtmx/internal/orchestration"
	"github.com/rtmx-ai/rtmx/internal/orchestration/deliverycheck"
	"github.com/rtmx-ai/rtmx/internal/orchestration/trade"
)

func (s *Server) projectRoot() string {
	// dbPath is typically <cwd>/.rtmx/database.csv
	return filepath.Dir(filepath.Dir(s.dbPath))
}

func scientificToolDefs() []toolDef {
	agentSchema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"agent_id": map[string]interface{}{"type": "string", "description": "Agent identity"},
		},
		"required": []string{"agent_id"},
	}
	return []toolDef{
		{
			Name:        "loop_tick",
			Description: "Next+claim+decompose plan for one agent cycle (~80 tokens)",
			InputSchema: agentSchema,
		},
		{
			Name:        "decompose",
			Description: "Split a coarse requirement into children (~40 tokens, mutation)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"req_id":   map[string]interface{}{"type": "string"},
					"agent_id": map[string]interface{}{"type": "string"},
					"dry_run":  map[string]interface{}{"type": "boolean"},
				},
				"required": []string{"req_id", "agent_id"},
			},
		},
		{
			Name:        "hygiene",
			Description: "RTM hygiene findings (~20 tokens/finding)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{"type": "integer"},
				},
			},
		},
		{
			Name:        "cycles",
			Description: "Circular dependency report (~15 tokens/cycle)",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
		{
			Name:        "webs",
			Description: "Independent work webs (~25 tokens/web)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"limit": map[string]interface{}{"type": "integer"},
				},
			},
		},
		{
			Name:        "context",
			Description: "Token-efficient RTM summary for agents (~200 tokens)",
			InputSchema: map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
		},
		{
			Name:        "delivery_check",
			Description: "Warn-first one-PR/one-commit delivery check (~50 tokens)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"req_id": map[string]interface{}{"type": "string"},
					"base":   map[string]interface{}{"type": "string"},
					"title":  map[string]interface{}{"type": "string"},
					"body":   map[string]interface{}{"type": "string"},
					"strict": map[string]interface{}{"type": "boolean"},
				},
			},
		},
		{
			Name:        "trade_open",
			Description: "Open a trade-analysis checkpoint (~35 tokens, mutation)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"req_id":   map[string]interface{}{"type": "string"},
					"title":    map[string]interface{}{"type": "string"},
					"agent_id": map[string]interface{}{"type": "string"},
				},
				"required": []string{"req_id", "agent_id"},
			},
		},
		{
			Name:        "trade_list",
			Description: "List trade-analysis checkpoints (~25 tokens/trade)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"req_id": map[string]interface{}{"type": "string"},
				},
			},
		},
		{
			Name:        "trade_resolve",
			Description: "Resolve a trade with a choice (~30 tokens, mutation)",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"trade_id": map[string]interface{}{"type": "string"},
					"choice":   map[string]interface{}{"type": "string"},
					"agent_id": map[string]interface{}{"type": "string"},
				},
				"required": []string{"trade_id", "choice", "agent_id"},
			},
		},
	}
}

type scienceLoopPlan struct {
	Idle             bool     `json:"idle,omitempty"`
	ReqID            string   `json:"req_id,omitempty"`
	Claimed          bool     `json:"claimed,omitempty"`
	Decomposed       bool     `json:"decomposed"`
	Children         []string `json:"children"`
	Redecomposed     []string `json:"redecomposed,omitempty"`
	PRPolicy         string   `json:"pr_policy"`
	CommitPolicy     string   `json:"commit_policy"`
	TradeRequired    bool     `json:"trade_required"`
	ReadyToImplement bool     `json:"ready_to_implement"`
	Note             string   `json:"note,omitempty"`
}

func (p scienceLoopPlan) MarshalJSON() ([]byte, error) {
	if p.Idle {
		return []byte(`{"idle":true}`), nil
	}
	children := p.Children
	if children == nil {
		children = []string{}
	}
	return json.Marshal(struct {
		ReqID            string   `json:"req_id"`
		Claimed          bool     `json:"claimed"`
		Decomposed       bool     `json:"decomposed"`
		Children         []string `json:"children"`
		Redecomposed     []string `json:"redecomposed,omitempty"`
		PRPolicy         string   `json:"pr_policy"`
		CommitPolicy     string   `json:"commit_policy"`
		TradeRequired    bool     `json:"trade_required"`
		ReadyToImplement bool     `json:"ready_to_implement"`
		Note             string   `json:"note,omitempty"`
	}{
		ReqID:            p.ReqID,
		Claimed:          p.Claimed,
		Decomposed:       p.Decomposed,
		Children:         children,
		Redecomposed:     p.Redecomposed,
		PRPolicy:         p.PRPolicy,
		CommitPolicy:     p.CommitPolicy,
		TradeRequired:    p.TradeRequired,
		ReadyToImplement: p.ReadyToImplement,
		Note:             p.Note,
	})
}

func (s *Server) toolLoopTick(db *database.Database, args map[string]interface{}) (interface{}, *rpcError) {
	agentID, _ := args["agent_id"].(string)
	if agentID == "" {
		return errorResult("agent_id is required"), nil
	}
	if s.claims == nil {
		return errorResult("claims store not initialized"), nil
	}
	root := s.projectRoot()
	req := pickDeliveryReq(db, graph.NewGraph(db).DetectWebs(), "")
	if req == nil {
		return scienceLoopPlan{Idle: true}, nil
	}
	// Skip if claimed by someone else
	if existing, _ := s.claims.Get(req.ReqID); existing != nil && existing.AgentID != agentID {
		return scienceLoopPlan{Idle: true}, nil
	}
	if _, err := s.claims.Claim(req.ReqID, agentID); err != nil {
		if ac, ok := err.(*orchestration.AlreadyClaimedError); !ok || ac.HeldBy != agentID {
			return errorResult(err.Error()), nil
		}
	}

	body := s.readRequirementMarkdown(req)
	children, decomposed := deliveryChildren(db, req, body)
	if decomposed {
		if err := s.applyScienceDecompose(db, req, children, body); err != nil {
			return errorResult(err.Error()), nil
		}
	}
	redecomposed := s.scienceRedecompose(db, req)
	tradeReq := false
	if open, _ := trade.HasOpen(root, req.ReqID); open {
		tradeReq = true
	}
	if trade.NeedsTrade(req.RequirementText, body) {
		tradeReq = true
	}
	ready := !tradeReq && !decomposed
	plan := scienceLoopPlan{
		ReqID:            req.ReqID,
		Claimed:          true,
		Decomposed:       decomposed,
		Children:         children,
		Redecomposed:     redecomposed,
		PRPolicy:         deliveryPRPolicy,
		CommitPolicy:     deliveryCommitPolicy,
		TradeRequired:    tradeReq,
		ReadyToImplement: ready,
	}
	if tradeReq {
		plan.Note = "resolve open trade before implementing"
	} else if decomposed {
		plan.Note = "parent is not the implementation target"
	}
	return plan, nil
}

func (s *Server) applyScienceDecompose(db *database.Database, parent *database.Requirement, childIDs []string, body string) error {
	if len(childIDs) == 0 {
		return nil
	}
	// Idempotent if children already exist
	existing := 0
	for _, id := range childIDs {
		if db.Get(id) != nil {
			existing++
		}
	}
	if existing == len(childIDs) {
		return nil
	}
	root := s.projectRoot()
	cat := parent.Category
	if cat == "" {
		cat = "GENERAL"
	}
	reqDir := filepath.Join(root, ".rtmx", "requirements", cat)
	if err := os.MkdirAll(reqDir, 0o755); err != nil {
		return err
	}
	titles := deliveryDecompositionTitles(body)
	for i, id := range childIDs {
		if db.Get(id) != nil {
			continue
		}
		title := id
		if i < len(titles) {
			title = titles[i]
		}
		md := fmt.Sprintf("# %s: %s\n\n## Acceptance Criteria\n\n1. %s\n", id, title, title)
		path := filepath.Join(reqDir, id+".md")
		if err := os.WriteFile(path, []byte(md), 0o644); err != nil {
			return err
		}
		rel := filepath.ToSlash(filepath.Join(".rtmx", "requirements", cat, id+".md"))
		child := &database.Requirement{
			ReqID:            id,
			Category:         cat,
			Subcategory:      parent.Subcategory,
			RequirementText:  title,
			Status:           database.StatusMissing,
			Priority:         parent.Priority,
			Phase:            parent.Phase,
			Dependencies:     database.NewStringSet(),
			Blocks:           database.NewStringSet(),
			RequirementFile:  rel,
			ValidationMethod: "System Test",
		}
		if parent.Blocks == nil {
			parent.Blocks = database.NewStringSet()
		}
		if parent.Dependencies == nil {
			parent.Dependencies = database.NewStringSet()
		}
		parent.Blocks.Add(id)
		parent.Dependencies.Add(id)
		if err := db.Add(child); err != nil {
			return err
		}
	}
	return db.Save(s.dbPath)
}

var scienceRedecomposeRE = regexp.MustCompile(`\bREQ-[A-Z0-9]+(?:-[A-Z0-9]+)*\b`)

func (s *Server) scienceRedecompose(db *database.Database, current *database.Requirement) []string {
	root := s.projectRoot()
	dir := filepath.Join(root, ".rtmx", "delivery")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var done []string
	seen := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		in := false
		for _, line := range strings.Split(string(raw), "\n") {
			trim := strings.TrimSpace(line)
			if strings.HasPrefix(trim, "## ") {
				if in {
					break
				}
				in = strings.EqualFold(trim, "## Follow-on decomposition")
				continue
			}
			if !in {
				continue
			}
			for _, id := range scienceRedecomposeRE.FindAllString(trim, -1) {
				if id == current.ReqID || seen[id] {
					continue
				}
				req := db.Get(id)
				if req == nil {
					continue
				}
				body := s.readRequirementMarkdown(req)
				children, decomposed := deliveryChildren(db, req, body)
				if !decomposed {
					continue
				}
				if err := s.applyScienceDecompose(db, req, children, body); err != nil {
					continue
				}
				seen[id] = true
				done = append(done, id)
			}
		}
	}
	return done
}

func (s *Server) toolDecompose(db *database.Database, args map[string]interface{}) (interface{}, *rpcError) {
	reqID, _ := args["req_id"].(string)
	agentID, _ := args["agent_id"].(string)
	dryRun, _ := args["dry_run"].(bool)
	if reqID == "" || agentID == "" {
		return errorResult("req_id and agent_id are required"), nil
	}
	req := db.Get(reqID)
	if req == nil {
		return errorResult("requirement not found"), nil
	}
	body := s.readRequirementMarkdown(req)
	children, decomposed := deliveryChildren(db, req, body)
	result := map[string]interface{}{
		"req_id":     reqID,
		"agent_id":   agentID,
		"atomic":     !decomposed,
		"decomposed": decomposed,
		"children":   children,
		"dry_run":    dryRun,
	}
	if !decomposed {
		result["note"] = "already atomic"
		return result, nil
	}
	if dryRun {
		return result, nil
	}
	if err := s.applyScienceDecompose(db, req, children, body); err != nil {
		return errorResult(err.Error()), nil
	}
	return result, nil
}

func (s *Server) toolHygiene(db *database.Database, args map[string]interface{}) interface{} {
	g := graph.NewGraph(db)
	cycles := g.FindCycles()
	type finding struct {
		Check   string `json:"check"`
		ReqID   string `json:"req_id,omitempty"`
		Message string `json:"message"`
	}
	var findings []finding
	for _, req := range db.All() {
		if req.EffortWeeks < 0.25 || req.EffortWeeks > 0.5 {
			findings = append(findings, finding{
				Check: "effort_bounds", ReqID: req.ReqID,
				Message: fmt.Sprintf("effort_weeks %.2f outside 0.25-0.50", req.EffortWeeks),
			})
		}
	}
	for _, cyc := range cycles {
		findings = append(findings, finding{
			Check: "cycle", Message: strings.Join(cyc, " -> "),
		})
	}
	limit := 0
	if v, ok := args["limit"].(float64); ok {
		limit = int(v)
	}
	if limit > 0 && len(findings) > limit {
		findings = findings[:limit]
	}
	return map[string]interface{}{
		"total":    db.Len(),
		"findings": findings,
		"cycles":   cycles,
	}
}

func (s *Server) toolCycles(db *database.Database) interface{} {
	cycles := graph.NewGraph(db).FindCycles()
	return map[string]interface{}{
		"found":  len(cycles) > 0,
		"count":  len(cycles),
		"cycles": cycles,
	}
}

func (s *Server) toolWebs(db *database.Database, args map[string]interface{}) interface{} {
	webs := graph.NewGraph(db).DetectWebs()
	limit := 0
	if v, ok := args["limit"].(float64); ok {
		limit = int(v)
	}
	type webJSON struct {
		Index     int      `json:"index"`
		IDs       []string `json:"ids"`
		Unblocked []string `json:"unblocked"`
	}
	out := make([]webJSON, 0, len(webs))
	for i, w := range webs {
		if limit > 0 && i >= limit {
			break
		}
		out = append(out, webJSON{Index: i + 1, IDs: w.IDs, Unblocked: w.Unblocked})
	}
	return map[string]interface{}{"webs": out, "count": len(webs)}
}

func (s *Server) toolContext(db *database.Database) interface{} {
	total := db.Len()
	complete := 0
	for _, req := range db.All() {
		if req.Status == database.StatusComplete {
			complete++
		}
	}
	pct := 0.0
	if total > 0 {
		pct = float64(complete) / float64(total) * 100
	}
	var blockers []string
	for _, req := range db.All() {
		if req.IsIncomplete() && req.Blocks != nil && req.Blocks.Len() > 0 {
			blockers = append(blockers, req.ReqID)
			if len(blockers) >= 3 {
				break
			}
		}
	}
	return map[string]interface{}{
		"completion_pct": pct,
		"total":          total,
		"complete":       complete,
		"blockers":       blockers,
	}
}

func (s *Server) toolDeliveryCheck(args map[string]interface{}) (interface{}, *rpcError) {
	reqID, _ := args["req_id"].(string)
	base, _ := args["base"].(string)
	title, _ := args["title"].(string)
	body, _ := args["body"].(string)
	strict, _ := args["strict"].(bool)
	res, err := deliverycheck.Check(deliverycheck.Input{
		Dir:   s.projectRoot(),
		Base:  base,
		ReqID: reqID,
		Title: title,
		Body:  body,
	})
	if err != nil {
		return errorResult(err.Error()), nil
	}
	out := map[string]interface{}{
		"ok":       res.OK,
		"req_id":   res.ReqID,
		"warnings": res.Warnings,
		"errors":   res.Errors,
		"commits":  res.Commits,
		"strict":   strict,
	}
	if strict && !res.OK {
		out["failed"] = true
	}
	return out, nil
}

func (s *Server) toolTradeOpen(args map[string]interface{}) (interface{}, *rpcError) {
	reqID, _ := args["req_id"].(string)
	title, _ := args["title"].(string)
	agentID, _ := args["agent_id"].(string)
	if reqID == "" || agentID == "" {
		return errorResult("req_id and agent_id are required"), nil
	}
	t, err := trade.Open(s.projectRoot(), reqID, title)
	if err != nil {
		return errorResult(err.Error()), nil
	}
	return map[string]interface{}{
		"id": t.ID, "req_id": t.ReqID, "title": t.Title, "status": t.Status,
		"path": t.Path, "agent_id": agentID,
	}, nil
}

func (s *Server) toolTradeList(args map[string]interface{}) (interface{}, *rpcError) {
	reqID, _ := args["req_id"].(string)
	trades, err := trade.List(s.projectRoot(), reqID)
	if err != nil {
		return errorResult(err.Error()), nil
	}
	return map[string]interface{}{"trades": trades, "count": len(trades)}, nil
}

func (s *Server) toolTradeResolve(args map[string]interface{}) (interface{}, *rpcError) {
	tradeID, _ := args["trade_id"].(string)
	choice, _ := args["choice"].(string)
	agentID, _ := args["agent_id"].(string)
	if tradeID == "" || choice == "" || agentID == "" {
		return errorResult("trade_id, choice, and agent_id are required"), nil
	}
	t, err := trade.Resolve(s.projectRoot(), tradeID, choice)
	if err != nil {
		return errorResult(err.Error()), nil
	}
	return map[string]interface{}{
		"id": t.ID, "status": t.Status, "choice": choice, "agent_id": agentID,
	}, nil
}
