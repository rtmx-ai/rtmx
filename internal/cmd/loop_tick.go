package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// loopPlan is the machine-readable tick result (REQ-ORCH-019c).
type loopPlan struct {
	ReqID                string   `json:"req_id"`
	Decomposed           bool     `json:"decomposed"`
	Children             []string `json:"children"`
	PRPolicy             string   `json:"pr_policy"`
	CommitPolicy         string   `json:"commit_policy"`
	Atomic               bool     `json:"atomic,omitempty"`
	PRTarget             string   `json:"pr_target,omitempty"`
	ImplementationTarget string   `json:"implementation_target,omitempty"`
	Note                 string   `json:"note,omitempty"`
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
	if _, err := store.Claim(req.ReqID, loopAgentID); err != nil {
		return fmt.Errorf("claim %s: %w", req.ReqID, err)
	}

	plan, err := decomposeForTick(cwd, db, dbPath, req)
	if err != nil {
		return err
	}
	raw, err := json.Marshal(plan)
	if err != nil {
		return err
	}
	cmd.Printf("%s\n", raw)
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
