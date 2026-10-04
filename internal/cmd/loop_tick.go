package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
	"github.com/rtmx-ai/rtmx/internal/orchestration"
	"github.com/spf13/cobra"
)

const (
	loopPRPolicy     = "one_pr_per_requirement"
	loopCommitPolicy = "one_commit_per_ac"
	loopAgentID      = "rtmx-loop"
)

// loopPlan is the machine-readable tick result (REQ-ORCH-019c / REQ-ORCH-023b).
type loopPlan struct {
	Idle                 bool     `json:"idle,omitempty"`
	ReqID                string   `json:"req_id,omitempty"`
	Claimed              bool     `json:"claimed,omitempty"`
	Decomposed           bool     `json:"decomposed"`
	Children             []string `json:"children"`
	Redecomposed         []string `json:"redecomposed,omitempty"`
	PRPolicy             string   `json:"pr_policy"`
	CommitPolicy         string   `json:"commit_policy"`
	TradeRequired        bool     `json:"trade_required"`
	ReadyToImplement     bool     `json:"ready_to_implement"`
	Atomic               bool     `json:"atomic,omitempty"`
	PRTarget             string   `json:"pr_target,omitempty"`
	ImplementationTarget string   `json:"implementation_target,omitempty"`
	Note                 string   `json:"note,omitempty"`
}

var (
	loopTickAgentID string
	loopTickStrict  bool
)

var loopTickCmd = &cobra.Command{
	Use:   "tick",
	Short: "Select, claim, and plan the next requirement for an agent",
	Long: `Run one agent-callable delivery tick: next + claim + decompose,
optional learning re-decompose from .rtmx/delivery notes, and trade_required.

Does not host a coding agent. Emits a JSON plan.`,
	RunE: runLoopTickCLI,
}

func init() {
	loopTickCmd.Flags().StringVar(&loopTickAgentID, "agent-id", "", "agent identity for the claim")
	loopTickCmd.Flags().BoolVar(&loopTickStrict, "strict", false, "fail when trade_required and trade still open")
	_ = loopTickCmd.MarkFlagRequired("agent-id")
	loopCmd.AddCommand(loopTickCmd)
}

func runLoopTickCLI(cmd *cobra.Command, args []string) error {
	return runAgentDeliveryTick(cmd, loopTickAgentID, loopTickStrict)
}

func runLoopOnce(cmd *cobra.Command) error {
	edges, err := loopEdges()
	if err != nil {
		cmd.PrintErrf("warning: falling-edge lookup unavailable (%v); not ticking\n", err)
		return writeIdlePlan(cmd)
	}
	if len(edges) == 0 {
		return writeIdlePlan(cmd)
	}
	return runDeliveryTick(cmd)
}

func runDeliveryTick(cmd *cobra.Command) error {
	return runAgentDeliveryTick(cmd, loopAgentID, false)
}

func runAgentDeliveryTick(cmd *cobra.Command, agentID string, strict bool) error {
	if agentID == "" {
		return fmt.Errorf("--agent-id is required")
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	dbPath := cfg.DatabasePath(cwd)
	db, err := database.Load(dbPath)
	if err != nil {
		return fmt.Errorf("failed to load database: %w", err)
	}

	req := selectNextUnclaimed(cmd, db, cwd)
	if req == nil {
		return writeIdlePlan(cmd)
	}

	store, err := orchestration.NewClaimStore(filepath.Join(cwd, ".rtmx", "claims"))
	if err != nil {
		return err
	}
	if _, err := store.Claim(req.ReqID, agentID); err != nil {
		// Idempotent: already claimed by self is OK.
		if ac, ok := err.(*orchestration.AlreadyClaimedError); ok && ac.HeldBy == agentID {
			// no-op
		} else {
			return fmt.Errorf("claim %s: %w", req.ReqID, err)
		}
	}

	plan, err := decomposeForTick(cwd, db, dbPath, req)
	if err != nil {
		return err
	}
	plan.Claimed = true
	plan.Redecomposed = learningRedecompose(cwd, db, dbPath, req)

	mdPath := resolveRequirementFile(cwd, req)
	mdBody := ""
	if mdPath != "" {
		if b, err := os.ReadFile(mdPath); err == nil {
			mdBody = string(b)
		}
	}
	plan.TradeRequired = tradeRequiredFor(cwd, req.ReqID, req.RequirementText, mdBody)
	plan.ReadyToImplement = !plan.TradeRequired && (plan.Atomic || (!plan.Decomposed && len(plan.Children) == 0) || plan.ImplementationTarget == req.ReqID)
	if plan.Decomposed && plan.ImplementationTarget == "children" {
		plan.ReadyToImplement = false
		if plan.Note == "" {
			plan.Note = "parent is not the implementation target"
		}
	}
	if plan.TradeRequired {
		plan.ReadyToImplement = false
		if plan.Note == "" {
			plan.Note = "resolve open trade before implementing"
		}
	}

	raw, err := json.Marshal(plan)
	if err != nil {
		return err
	}
	cmd.Printf("%s\n", raw)
	if strict && plan.TradeRequired {
		return NewExitError(1, "trade_required")
	}
	return nil
}

func loopEdges() ([]LoopEdge, error) {
	if loopEdgesOverride != nil {
		return *loopEdgesOverride, nil
	}
	return nil, nil
}

func writeIdlePlan(cmd *cobra.Command) error {
	cmd.Println(`{"idle":true}`)
	return nil
}

func selectNextUnclaimed(cmd *cobra.Command, db *database.Database, cwd string) *database.Requirement {
	g := graph.NewGraph(db)
	webs := g.DetectWebs()
	claimed := claimedReqIDs(filepath.Join(cwd, ".rtmx", "claims"))
	var unblocked []string
	for _, web := range webs {
		for _, id := range web.Unblocked {
			if claimed[id] {
				continue
			}
			unblocked = append(unblocked, id)
		}
	}
	named := lookupOpenPRNames(cmd, unblocked)
	keep, _ := openPRSkips(unblocked, named)
	return pickHighestPriority(db, keep)
}

func claimedReqIDs(dir string) map[string]bool {
	out := map[string]bool{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".json") {
			out[strings.TrimSuffix(name, ".json")] = true
		}
	}
	return out
}

func decomposeForTick(cwd string, db *database.Database, dbPath string, req *database.Requirement) (*loopPlan, error) {
	mdPath := resolveRequirementFile(cwd, req)
	mdBody := ""
	if mdPath != "" {
		if b, err := os.ReadFile(mdPath); err == nil {
			mdBody = string(b)
		}
	}
	planned, err := planDecompose(req, mdBody, db)
	if err != nil {
		return nil, err
	}
	plan := &loopPlan{
		ReqID:        req.ReqID,
		Children:     []string{},
		PRPolicy:     loopPRPolicy,
		CommitPolicy: loopCommitPolicy,
	}
	if planned.Atomic {
		plan.Atomic = true
		plan.PRTarget = req.ReqID
		plan.ImplementationTarget = req.ReqID
		plan.Note = "already atomic"
		plan.ReadyToImplement = true
		return plan, nil
	}
	if err := applyDecompose(cwd, db, req, planned, mdPath); err != nil {
		return nil, err
	}
	if err := db.Save(dbPath); err != nil {
		return nil, fmt.Errorf("save database: %w", err)
	}
	plan.Decomposed = true
	plan.Children = append([]string{}, planned.ChildIDs...)
	plan.ImplementationTarget = "children"
	plan.Note = "parent is not the implementation target"
	return plan, nil
}

var redecomposeReqPattern = regexp.MustCompile(`\bREQ-[A-Z0-9]+(?:-[A-Z0-9]+)*\b`)

// learningRedecompose reads delivery notes / COMPLETE siblings for follow-on IDs and decomposes them.
func learningRedecompose(cwd string, db *database.Database, dbPath string, current *database.Requirement) []string {
	ids := collectRedecomposeIDs(cwd, db, current)
	var done []string
	for _, id := range ids {
		if id == current.ReqID {
			continue
		}
		req := db.Get(id)
		if req == nil || req.Status == database.StatusComplete {
			continue
		}
		mdPath := resolveRequirementFile(cwd, req)
		mdBody := ""
		if mdPath != "" {
			if b, err := os.ReadFile(mdPath); err == nil {
				mdBody = string(b)
			}
		}
		planned, err := planDecompose(req, mdBody, db)
		if err != nil || planned.Atomic {
			continue
		}
		if err := applyDecompose(cwd, db, req, planned, mdPath); err != nil {
			continue
		}
		done = append(done, id)
	}
	if len(done) > 0 {
		_ = db.Save(dbPath)
	}
	return done
}

func collectRedecomposeIDs(cwd string, db *database.Database, current *database.Requirement) []string {
	seen := map[string]bool{}
	var out []string
	add := func(id string) {
		if id == "" || seen[id] {
			return
		}
		seen[id] = true
		out = append(out, id)
	}

	deliveryDir := filepath.Join(cwd, ".rtmx", "delivery")
	entries, _ := os.ReadDir(deliveryDir)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(deliveryDir, e.Name()))
		if err != nil {
			continue
		}
		for _, id := range parseRedecomposeTargets(string(raw)) {
			add(id)
		}
	}

	// COMPLETE siblings that share a dependency web with current.
	g := graph.NewGraph(db)
	for _, web := range g.DetectWebs() {
		inWeb := false
		for _, id := range web.IDs {
			if id == current.ReqID {
				inWeb = true
				break
			}
		}
		if !inWeb {
			continue
		}
		for _, id := range web.IDs {
			r := db.Get(id)
			if r == nil || r.Status != database.StatusComplete {
				continue
			}
			mdPath := resolveRequirementFile(cwd, r)
			if mdPath == "" {
				continue
			}
			raw, err := os.ReadFile(mdPath)
			if err != nil {
				continue
			}
			for _, rid := range parseRedecomposeTargets(string(raw)) {
				add(rid)
			}
		}
	}
	return out
}

func parseRedecomposeTargets(body string) []string {
	var out []string
	seen := map[string]bool{}
	add := func(id string) {
		if seen[id] {
			return
		}
		seen[id] = true
		out = append(out, id)
	}
	// YAML front matter: redecompose: [REQ-…]
	if fm := extractYAMLFrontMatter(body); fm != "" {
		for _, id := range redecomposeReqPattern.FindAllString(fm, -1) {
			if strings.Contains(strings.ToLower(fm), "redecompose") {
				add(id)
			}
		}
		// Also match redecompose: line specifically
		for _, line := range strings.Split(fm, "\n") {
			if strings.HasPrefix(strings.TrimSpace(strings.ToLower(line)), "redecompose:") {
				for _, id := range redecomposeReqPattern.FindAllString(line, -1) {
					add(id)
				}
			}
		}
	}
	// ## Follow-on decomposition section
	in := false
	for _, line := range strings.Split(body, "\n") {
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
		for _, id := range redecomposeReqPattern.FindAllString(trim, -1) {
			add(id)
		}
	}
	return out
}

func extractYAMLFrontMatter(body string) string {
	if !strings.HasPrefix(body, "---\n") {
		return ""
	}
	rest := strings.TrimPrefix(body, "---\n")
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return ""
	}
	return rest[:idx]
}
