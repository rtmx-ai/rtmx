package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/spf13/cobra"
)

var websJSON bool

var websCmd = &cobra.Command{
	Use:   "webs",
	Short: "Display independent work webs with parallel group and dependency metadata",
	Long: `Analyze the dependency graph to find independent work webs,
compute parallel execution groups, and report cross-web dependencies
and file surface overlaps.

Examples:
    rtmx webs              # table output
    rtmx webs --json       # machine-readable JSON`,
	RunE: runWebs,
}

func init() {
	websCmd.Flags().BoolVar(&websJSON, "json", false, "output as JSON")
	rootCmd.AddCommand(websCmd)
}

// websOutput is the JSON-serializable output of the webs command.
type websOutput struct {
	Webs           []websWebJSON       `json:"webs"`
	ParallelGroups []websGroupJSON     `json:"parallel_groups"`
	Dependencies   []websDepJSON       `json:"dependencies"`
	Overlaps       []websOverlapJSON   `json:"overlaps"`
}

type websWebJSON struct {
	Index       int      `json:"index"`
	IDs         []string `json:"ids"`
	Unblocked   []string `json:"unblocked"`
	Blocked     []string `json:"blocked"`
	TotalEffort float64  `json:"total_effort"`
}

type websGroupJSON struct {
	Group      int   `json:"group"`
	WebIndices []int `json:"web_indices"`
}

type websDepJSON struct {
	From int `json:"from"`
	To   int `json:"to"`
}

type websOverlapJSON struct {
	WebA        int      `json:"web_a"`
	WebB        int      `json:"web_b"`
	SharedFiles []string `json:"shared_files"`
}

func runWebs(cmd *cobra.Command, args []string) error {
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
	deps := g.WebDependencies(webs)
	overlaps := g.DetectOverlaps(webs)
	groups := g.ParallelGroups(webs)

	if websJSON {
		out := websOutput{
			Webs:           make([]websWebJSON, len(webs)),
			ParallelGroups: make([]websGroupJSON, len(groups)),
			Dependencies:   make([]websDepJSON, len(deps)),
			Overlaps:       make([]websOverlapJSON, len(overlaps)),
		}
		for i, w := range webs {
			out.Webs[i] = websWebJSON{
				Index: i, IDs: w.IDs, Unblocked: w.Unblocked,
				Blocked: w.Blocked, TotalEffort: w.TotalEffort,
			}
		}
		for i, grp := range groups {
			out.ParallelGroups[i] = websGroupJSON{Group: i, WebIndices: grp.WebIndices}
		}
		for i, d := range deps {
			out.Dependencies[i] = websDepJSON{From: d.From, To: d.To}
		}
		for i, o := range overlaps {
			out.Overlaps[i] = websOverlapJSON{WebA: o.WebA, WebB: o.WebB, SharedFiles: o.SharedFiles}
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(out)
	}

	// Table output
	w := cmd.OutOrStdout()
	fmt.Fprintf(w, "Work Webs: %d webs, %d parallel groups\n\n", len(webs), len(groups))

	for i, web := range webs {
		groupIdx := -1
		for gi, grp := range groups {
			for _, wi := range grp.WebIndices {
				if wi == i {
					groupIdx = gi
				}
			}
		}
		fmt.Fprintf(w, "Web %d (Group %d) -- %d reqs, %.1fw effort\n",
			i, groupIdx, len(web.IDs), web.TotalEffort)
		fmt.Fprintf(w, "  IDs: %s\n", strings.Join(web.IDs, ", "))
		if len(web.Unblocked) > 0 {
			fmt.Fprintf(w, "  Unblocked: %s\n", strings.Join(web.Unblocked, ", "))
		}
		if len(web.Blocked) > 0 {
			fmt.Fprintf(w, "  Blocked: %s\n", strings.Join(web.Blocked, ", "))
		}
		fmt.Fprintln(w)
	}

	if len(deps) > 0 {
		fmt.Fprintln(w, "Cross-Web Dependencies:")
		for _, d := range deps {
			fmt.Fprintf(w, "  Web %d -> Web %d\n", d.From, d.To)
		}
		fmt.Fprintln(w)
	}

	if len(overlaps) > 0 {
		fmt.Fprintln(w, "File Surface Overlaps:")
		for _, o := range overlaps {
			fmt.Fprintf(w, "  Web %d <-> Web %d: %s\n", o.WebA, o.WebB, strings.Join(o.SharedFiles, ", "))
		}
		fmt.Fprintln(w)
	}

	return nil
}
