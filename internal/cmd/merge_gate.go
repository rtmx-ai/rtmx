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

var (
	mergeGateWebIndex int
	mergeGateJSON     bool
)

var mergeGateCmd = &cobra.Command{
	Use:   "merge-gate",
	Short: "Validate web merge safety checking completion and upstream status",
	Long: `Check whether a work web is safe to merge by verifying that all
requirements in the web are complete and all upstream web dependencies
are satisfied.

Exits 0 if safe to merge, exits 1 with failure details otherwise.

Examples:
    rtmx merge-gate --web 0        # check if web 0 is safe to merge
    rtmx merge-gate --web 0 --json # machine-readable output`,
	RunE: runMergeGate,
}

func init() {
	mergeGateCmd.Flags().IntVar(&mergeGateWebIndex, "web", -1, "web index to check (required)")
	mergeGateCmd.Flags().BoolVar(&mergeGateJSON, "json", false, "output as JSON")
	rootCmd.AddCommand(mergeGateCmd)
}

// mergeGateOutput is the JSON-serializable result.
type mergeGateOutput struct {
	WebIndex         int      `json:"web_index"`
	Safe             bool     `json:"safe"`
	IncompleteIDs    []string `json:"incomplete_ids,omitempty"`
	BlockedUpstream  []int    `json:"blocked_upstream,omitempty"`
	Failures         []string `json:"failures,omitempty"`
}

func runMergeGate(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	if mergeGateWebIndex < 0 {
		return fmt.Errorf("--web flag is required (use 'rtmx webs' to list web indices)")
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

	if mergeGateWebIndex >= len(webs) {
		return fmt.Errorf("web index %d out of range (have %d webs)", mergeGateWebIndex, len(webs))
	}

	web := webs[mergeGateWebIndex]
	deps := g.WebDependencies(webs)

	result := mergeGateOutput{
		WebIndex: mergeGateWebIndex,
		Safe:     true,
	}

	// Check all requirements in the web are complete
	for _, id := range web.IDs {
		req := db.Get(id)
		if req != nil && req.IsIncomplete() {
			result.IncompleteIDs = append(result.IncompleteIDs, id)
		}
	}
	if len(result.IncompleteIDs) > 0 {
		result.Safe = false
		result.Failures = append(result.Failures,
			fmt.Sprintf("%d incomplete requirements in web", len(result.IncompleteIDs)))
	}

	// Check upstream webs are complete
	for _, d := range deps {
		if d.To == mergeGateWebIndex {
			// d.From is an upstream web -- check if it has incomplete reqs
			upstream := webs[d.From]
			hasIncomplete := false
			for _, uid := range upstream.IDs {
				req := db.Get(uid)
				if req != nil && req.IsIncomplete() {
					hasIncomplete = true
					break
				}
			}
			if hasIncomplete {
				result.BlockedUpstream = append(result.BlockedUpstream, d.From)
			}
		}
	}
	if len(result.BlockedUpstream) > 0 {
		result.Safe = false
		result.Failures = append(result.Failures,
			fmt.Sprintf("%d upstream webs still incomplete", len(result.BlockedUpstream)))
	}

	if mergeGateJSON {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			return err
		}
	} else {
		w := cmd.OutOrStdout()
		if result.Safe {
			_, _ = fmt.Fprintf(w, "Web %d: SAFE to merge\n", mergeGateWebIndex)
		} else {
			_, _ = fmt.Fprintf(w, "Web %d: NOT safe to merge\n", mergeGateWebIndex)
			for _, f := range result.Failures {
				_, _ = fmt.Fprintf(w, "  - %s\n", f)
			}
			if len(result.IncompleteIDs) > 0 {
				_, _ = fmt.Fprintf(w, "  Incomplete: %v\n", result.IncompleteIDs)
			}
			if len(result.BlockedUpstream) > 0 {
				_, _ = fmt.Fprintf(w, "  Blocked by upstream webs: %v\n", result.BlockedUpstream)
			}
		}
	}

	if !result.Safe {
		return fmt.Errorf("merge gate failed: web %d is not safe to merge", mergeGateWebIndex)
	}
	return nil
}
