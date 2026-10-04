package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/spf13/cobra"
)

var planParallelJSON bool

var planParallelCmd = &cobra.Command{
	Use:   "plan-parallel",
	Short: "Produce concrete execution plan with agent assignments and merge gates",
	Long: `Analyze work webs and produce a phased execution plan showing which
webs can run in parallel and in what order they should be merged.

Each phase contains webs that can execute concurrently. Merge gates
between phases ensure upstream dependencies are satisfied before
downstream webs begin.

Examples:
    rtmx plan-parallel              # table output
    rtmx plan-parallel --json       # machine-readable JSON`,
	RunE: runPlanParallel,
}

func init() {
	planParallelCmd.Flags().BoolVar(&planParallelJSON, "json", false, "output as JSON")
	rootCmd.AddCommand(planParallelCmd)
}

// planOutput is the JSON-serializable output of plan-parallel.
type planOutput struct {
	Phases       []planPhase `json:"phases"`
	TotalWebs    int         `json:"total_webs"`
	TotalEffort  float64     `json:"total_effort"`
	CriticalPath float64     `json:"critical_path_effort"`
	Parallelism  float64     `json:"parallelism_factor"`
}

type planPhase struct {
	Phase       int          `json:"phase"`
	Webs        []planWeb    `json:"webs"`
	MergeOrder  []int        `json:"merge_order"`
	PhaseEffort float64      `json:"phase_effort"`
}

type planWeb struct {
	WebIndex    int      `json:"web_index"`
	IDs         []string `json:"ids"`
	TotalEffort float64  `json:"total_effort"`
	Unblocked   int      `json:"unblocked_count"`
	Blocked     int      `json:"blocked_count"`
}

func runPlanParallel(cmd *cobra.Command, args []string) error {
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

	db, err := database.Load(cfg.DatabasePath(cwd))
	if err != nil {
		return fmt.Errorf("failed to load database: %w", err)
	}

	g := graph.NewGraph(db)
	webs := g.DetectWebs()
	groups := g.ParallelGroups(webs)
	mergeOrder := g.MergeOrder(webs)

	// Build plan output
	plan := planOutput{
		TotalWebs: len(webs),
	}

	var totalEffort float64
	for _, w := range webs {
		totalEffort += w.TotalEffort
	}
	plan.TotalEffort = totalEffort

	// Build phases from parallel groups
	critPath := 0.0
	for i, grp := range groups {
		phase := planPhase{Phase: i}
		maxEffort := 0.0
		for _, wi := range grp.WebIndices {
			w := webs[wi]
			phase.Webs = append(phase.Webs, planWeb{
				WebIndex:    wi,
				IDs:         w.IDs,
				TotalEffort: w.TotalEffort,
				Unblocked:   len(w.Unblocked),
				Blocked:     len(w.Blocked),
			})
			if w.TotalEffort > maxEffort {
				maxEffort = w.TotalEffort
			}
			phase.PhaseEffort += w.TotalEffort
		}
		critPath += maxEffort

		// Compute merge order for this phase's webs
		for _, mi := range mergeOrder {
			for _, wi := range grp.WebIndices {
				if mi == wi {
					phase.MergeOrder = append(phase.MergeOrder, mi)
				}
			}
		}

		plan.Phases = append(plan.Phases, phase)
	}

	plan.CriticalPath = critPath
	if critPath > 0 {
		plan.Parallelism = totalEffort / critPath
	}

	if planParallelJSON {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(plan)
	}

	// Table output
	w := cmd.OutOrStdout()
	_, _ = fmt.Fprintf(w, "Parallel Execution Plan: %d webs, %d phases\n", plan.TotalWebs, len(plan.Phases))
	_, _ = fmt.Fprintf(w, "Total effort: %.1fw, Critical path: %.1fw, Parallelism: %.1fx\n\n",
		plan.TotalEffort, plan.CriticalPath, plan.Parallelism)

	for _, phase := range plan.Phases {
		_, _ = fmt.Fprintf(w, "Phase %d (%.1fw effort):\n", phase.Phase, phase.PhaseEffort)
		for _, pw := range phase.Webs {
			_, _ = fmt.Fprintf(w, "  Web %d: %d reqs (%.1fw), %d unblocked, %d blocked\n",
				pw.WebIndex, len(pw.IDs), pw.TotalEffort, pw.Unblocked, pw.Blocked)
		}
		_, _ = fmt.Fprintf(w, "  Merge order: %v\n\n", phase.MergeOrder)
	}

	return nil
}
