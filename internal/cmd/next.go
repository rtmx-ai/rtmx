package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
	"github.com/rtmx-ai/rtmx/internal/orchestration"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/spf13/cobra"
)

var (
	nextOne      bool
	nextJSON     bool
	nextBatch    bool
	nextWorktree bool
	nextAgentID  string
)

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Show available work webs and pick next requirement",
	Long: `Analyze the dependency graph to find independent work webs --
groups of incomplete requirements connected by dependency edges.
Requirements in different webs can be worked on in parallel.

By default, displays all work webs with summary stats. Use --one
to claim the highest-priority unblocked requirement.

Examples:
    rtmx next              # show all work webs
    rtmx next --one        # pick highest-priority unblocked item
    rtmx next --json       # machine-readable output`,
	RunE: runNext,
}

func init() {
	nextCmd.Flags().BoolVar(&nextOne, "one", false, "pick single highest-priority unblocked requirement")
	nextCmd.Flags().BoolVar(&nextJSON, "json", false, "output as JSON")
	nextCmd.Flags().BoolVar(&nextBatch, "batch", false, "claim entire highest-priority work web")
	nextCmd.Flags().BoolVar(&nextWorktree, "worktree", false, "create git worktree for claimed web (requires --batch)")
	nextCmd.Flags().StringVar(&nextAgentID, "agent-id", "", "agent identifier for claims (defaults to hostname)")
	rootCmd.AddCommand(nextCmd)
}

func runNext(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
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

	g := graph.NewGraph(db)
	webs := g.DetectWebs()

	if len(webs) == 0 {
		cmd.Println("No incomplete requirements found.")
		return nil
	}

	if nextBatch {
		return runNextBatch(cmd, db, webs, cwd)
	}

	if nextOne {
		return runNextOne(cmd, db, webs)
	}

	return runNextShow(cmd, db, webs)
}

func runNextShow(cmd *cobra.Command, db *database.Database, webs []graph.Web) error {
	width := output.TerminalWidth()
	cmd.Println(output.Header("Work Webs", width))
	cmd.Println()

	totalReqs := 0
	totalEffort := 0.0
	for _, web := range webs {
		totalReqs += len(web.IDs)
		totalEffort += web.TotalEffort
	}

	cmd.Printf("  %d web(s), %d incomplete requirement(s), %.1f effort-weeks total\n\n",
		len(webs), totalReqs, totalEffort)

	for i, web := range webs {
		// Find top-priority unblocked item
		topItem := ""
		topPriority := ""
		if len(web.Unblocked) > 0 {
			best := pickHighestPriority(db, web.Unblocked)
			if best != nil {
				topItem = best.ReqID
				topPriority = string(best.Priority)
			}
		}

		label := fmt.Sprintf("Web %d", i+1)
		cmd.Printf("  %s  %d reqs (%d unblocked, %d blocked)  %.1fw effort",
			output.Color(label, output.Cyan),
			len(web.IDs),
			len(web.Unblocked),
			len(web.Blocked),
			web.TotalEffort)

		if topItem != "" {
			cmd.Printf("  -> %s [%s]",
				output.Color(topItem, output.Green),
				topPriority)
		}
		cmd.Println()

		// List members
		for _, id := range web.IDs {
			req := db.Get(id)
			if req == nil {
				continue
			}
			icon := output.StatusIcon(req.Status.String())
			blocked := ""
			if contains(web.Blocked, id) {
				blocked = output.Color(" (blocked)", output.Dim)
			}
			cmd.Printf("    %s %s  [%s]  %s%s\n",
				icon,
				output.Color(output.PadRight(id, 15), output.Cyan),
				string(req.Priority),
				output.Truncate(req.RequirementText, 40),
				blocked)
		}
		cmd.Println()
	}

	return nil
}

func runNextOne(cmd *cobra.Command, db *database.Database, webs []graph.Web) error {
	// Collect all unblocked items across all webs
	var allUnblocked []string
	for _, web := range webs {
		allUnblocked = append(allUnblocked, web.Unblocked...)
	}

	if len(allUnblocked) == 0 {
		cmd.Println("No unblocked requirements available.")
		return nil
	}

	best := pickHighestPriority(db, allUnblocked)
	if best == nil {
		cmd.Println("No unblocked requirements available.")
		return nil
	}

	if nextJSON {
		cmd.Printf("{\"req_id\":%q,\"priority\":%q,\"category\":%q,\"effort_weeks\":%.1f,\"text\":%q}\n",
			best.ReqID, string(best.Priority), best.Category, best.EffortWeeks, best.RequirementText)
		return nil
	}

	width := output.TerminalWidth()
	cmd.Println(output.Header("Next Requirement", width))
	cmd.Println()
	cmd.Printf("  %s  %s\n", output.Color("Requirement:", output.Dim), output.Color(best.ReqID, output.Cyan))
	cmd.Printf("  %s  %s\n", output.Color("Priority:   ", output.Dim), string(best.Priority))
	cmd.Printf("  %s  %s\n", output.Color("Category:   ", output.Dim), best.Category)
	if best.EffortWeeks > 0 {
		cmd.Printf("  %s  %.1f weeks\n", output.Color("Effort:     ", output.Dim), best.EffortWeeks)
	}
	cmd.Printf("  %s  %s\n", output.Color("Description:", output.Dim), best.RequirementText)

	if best.RequirementFile != "" {
		cmd.Printf("  %s  %s\n", output.Color("Spec:       ", output.Dim), best.RequirementFile)
	}

	// Show dependencies
	if len(best.Dependencies) > 0 {
		cmd.Printf("  %s  ", output.Color("Depends on: ", output.Dim))
		first := true
		for dep := range best.Dependencies {
			if !first {
				cmd.Print(", ")
			}
			depReq := db.Get(dep)
			if depReq != nil {
				icon := output.StatusIcon(depReq.Status.String())
				cmd.Printf("%s %s", icon, dep)
			} else {
				cmd.Print(dep)
			}
			first = false
		}
		cmd.Println()
	}

	cmd.Println()
	return nil
}

// pickHighestPriority returns the highest-priority requirement from a list of IDs.
// Ties are broken by effort (smallest first, for quick wins).
func pickHighestPriority(db *database.Database, ids []string) *database.Requirement {
	if len(ids) == 0 {
		return nil
	}

	reqs := make([]*database.Requirement, 0, len(ids))
	for _, id := range ids {
		if r := db.Get(id); r != nil {
			reqs = append(reqs, r)
		}
	}

	if len(reqs) == 0 {
		return nil
	}

	sort.Slice(reqs, func(i, j int) bool {
		pi := reqs[i].Priority.Weight()
		pj := reqs[j].Priority.Weight()
		if pi != pj {
			return pi < pj // lower sort order = higher priority
		}
		return reqs[i].EffortWeeks < reqs[j].EffortWeeks
	})

	return reqs[0]
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// runNextBatch claims all requirements in the highest-priority work web.
func runNextBatch(cmd *cobra.Command, db *database.Database, webs []graph.Web, cwd string) error {
	if len(webs) == 0 {
		cmd.Println("No work webs available.")
		return nil
	}

	agentID := nextAgentID
	if agentID == "" {
		hostname, _ := os.Hostname()
		agentID = hostname
	}

	// Pick the web with the highest-priority unblocked requirement
	bestWebIdx := 0
	var bestPriority *database.Requirement
	for i, web := range webs {
		candidate := pickHighestPriority(db, web.Unblocked)
		if candidate == nil {
			continue
		}
		if bestPriority == nil || candidate.Priority.Weight() < bestPriority.Priority.Weight() {
			bestWebIdx = i
			bestPriority = candidate
		}
	}

	web := webs[bestWebIdx]

	// Claim all requirements in topological order
	claimsDir := fmt.Sprintf("%s/.rtmx/claims", cwd)
	store, err := orchestration.NewClaimStore(claimsDir)
	if err != nil {
		return fmt.Errorf("failed to create claim store: %w", err)
	}

	// Topological order: unblocked first, then blocked
	ordered := make([]string, 0, len(web.IDs))
	ordered = append(ordered, web.Unblocked...)
	ordered = append(ordered, web.Blocked...)

	var claimed []string
	for _, id := range ordered {
		_, claimErr := store.Claim(id, agentID)
		if claimErr != nil {
			// Already claimed -- skip silently
			continue
		}
		claimed = append(claimed, id)
	}

	if nextJSON {
		cmd.Printf("{\"web\":%d,\"claimed\":%d,\"agent_id\":%q,\"ids\":[", bestWebIdx+1, len(claimed), agentID)
		for i, id := range claimed {
			if i > 0 {
				cmd.Print(",")
			}
			cmd.Printf("%q", id)
		}
		cmd.Println("]}")
		return nil
	}

	cmd.Printf("Claimed %d requirement(s) in Web %d for agent %s:\n", len(claimed), bestWebIdx+1, agentID)
	for _, id := range claimed {
		req := db.Get(id)
		if req != nil {
			cmd.Printf("  %s  [%s]  %s\n", id, string(req.Priority), output.Truncate(req.RequirementText, 50))
		}
	}

	// Worktree binding (REQ-ORCH-008)
	if nextWorktree && len(claimed) > 0 {
		wtPath, branch, wtErr := orchestration.CreateWorktree(cwd, bestWebIdx+1)
		if wtErr != nil {
			return fmt.Errorf("failed to create worktree: %w", wtErr)
		}
		cmd.Printf("\nWorktree created:\n  Path:   %s\n  Branch: %s\n", wtPath, branch)
	}

	return nil
}
