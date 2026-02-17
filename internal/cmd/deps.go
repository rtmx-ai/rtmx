package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/internal/graph"
	"github.com/rtmx-ai/rtmx-go/internal/output"
	"github.com/spf13/cobra"
)

var (
	depsReverse  bool
	depsAll      bool
	depsWorkable bool
)

var depsCmd = &cobra.Command{
	Use:   "deps [req_id]",
	Short: "Show requirement dependencies",
	Long: `Display dependency information for requirements.

Without arguments, shows an overview of the dependency graph.
With a requirement ID, shows dependencies for that requirement.

Flags:
  --reverse   Show dependents instead of dependencies
  --all       Show transitive dependencies (not just direct)
  --workable  Show only unblocked incomplete requirements`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDeps,
}

func init() {
	depsCmd.Flags().BoolVarP(&depsReverse, "reverse", "r", false, "show dependents instead of dependencies")
	depsCmd.Flags().BoolVarP(&depsAll, "all", "a", false, "show transitive dependencies")
	depsCmd.Flags().BoolVarP(&depsWorkable, "workable", "w", false, "show only unblocked incomplete requirements")

	rootCmd.AddCommand(depsCmd)
}

func runDeps(cmd *cobra.Command, args []string) error {
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

	if len(args) > 0 {
		return showReqDeps(cmd, args[0], db, g)
	}

	if depsWorkable {
		return showWorkable(cmd, db, g)
	}

	return showDepsOverview(cmd, db, g)
}

func showReqDeps(cmd *cobra.Command, reqID string, db *database.Database, g *graph.Graph) error {
	req := db.Get(reqID)
	if req == nil {
		return fmt.Errorf("requirement %s not found", reqID)
	}

	width := 80
	cmd.Println(output.Header(fmt.Sprintf("Dependencies: %s", reqID), width))
	cmd.Println()

	// Show requirement info
	icon := output.StatusIcon(req.Status.String())
	cmd.Printf("%s %s [%s]\n", icon, reqID, req.Priority)
	cmd.Printf("   %s\n\n", output.Truncate(req.RequirementText, 70))

	var deps []string
	var label string

	if depsReverse {
		if depsAll {
			deps = g.TransitiveDependents(reqID)
			label = "All Dependents (transitive)"
		} else {
			deps = g.Dependents(reqID)
			label = "Direct Dependents"
		}
	} else {
		if depsAll {
			deps = g.TransitiveDependencies(reqID)
			label = "All Dependencies (transitive)"
		} else {
			deps = g.Dependencies(reqID)
			label = "Direct Dependencies"
		}
	}

	cmd.Println(output.SubHeader(label, width))
	if len(deps) == 0 {
		cmd.Println("  (none)")
	} else {
		for _, depID := range deps {
			depReq := db.Get(depID)
			if depReq != nil {
				icon := output.StatusIcon(depReq.Status.String())
				cmd.Printf("  %s %s %s\n", icon, depID, output.Truncate(depReq.RequirementText, 50))
			}
		}
	}

	// Show blocking info
	if !depsReverse {
		cmd.Println()
		blocking := g.BlockingDependencies(reqID)
		if len(blocking) > 0 {
			cmd.Printf("%s Blocked by %d incomplete requirement(s): %s\n",
				output.Color("!", output.Yellow),
				len(blocking),
				strings.Join(blocking, ", "))
		} else if len(deps) > 0 {
			cmd.Printf("%s All dependencies are complete\n", output.Color("✓", output.Green))
		}
	}

	return nil
}

func showWorkable(cmd *cobra.Command, db *database.Database, g *graph.Graph) error {
	width := 80
	cmd.Println(output.Header("Workable Requirements", width))
	cmd.Println()

	workable := g.NextWorkable()

	if len(workable) == 0 {
		cmd.Println("No unblocked incomplete requirements found.")
		cmd.Println("All incomplete requirements are blocked by dependencies.")
		return nil
	}

	cmd.Printf("Found %d requirement(s) ready to work on:\n\n", len(workable))

	for _, reqID := range workable {
		req := db.Get(reqID)
		if req != nil {
			icon := output.StatusIcon(req.Status.String())
			priorityColor := output.PriorityColor(req.Priority.String())
			cmd.Printf("%s %s [%s] Phase %d\n",
				icon,
				output.Color(reqID, output.Cyan),
				output.Color(string(req.Priority), priorityColor),
				req.Phase)
			cmd.Printf("   %s\n\n", output.Truncate(req.RequirementText, 70))
		}
	}

	return nil
}

func showDepsOverview(cmd *cobra.Command, db *database.Database, g *graph.Graph) error {
	width := 80
	cmd.Println(output.Header("Dependencies", width))
	cmd.Println()

	// Get all requirements and build dependency info
	type reqInfo struct {
		id          string
		depsCount   int
		blocksCount int
		description string
	}

	var reqs []reqInfo
	for _, r := range db.All() {
		deps := len(g.Dependencies(r.ReqID))
		blocks := len(g.Dependents(r.ReqID))
		reqs = append(reqs, reqInfo{
			id:          r.ReqID,
			depsCount:   deps,
			blocksCount: blocks,
			description: r.RequirementText,
		})
	}

	// Sort by blocks descending (most blockers first)
	for i := 0; i < len(reqs)-1; i++ {
		for j := i + 1; j < len(reqs); j++ {
			if reqs[i].blocksCount < reqs[j].blocksCount {
				reqs[i], reqs[j] = reqs[j], reqs[i]
			}
		}
	}

	// Print header
	cmd.Printf("%-18s %5s  %6s   %s\n",
		"ID", "Deps", "Blocks", "Description")
	cmd.Println(strings.Repeat("-", width+5))

	// Print rows
	for _, r := range reqs {
		desc := output.TruncateCell(r.description, 45)
		cmd.Printf("%-18s %5d  %6d   %s\n",
			r.id, r.depsCount, r.blocksCount, desc)
	}

	return nil
}
