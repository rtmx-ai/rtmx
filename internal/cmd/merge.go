package cmd

import (
	"fmt"
	"os"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/graph"
	"github.com/rtmx-ai/rtmx/internal/orchestration"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/spf13/cobra"
)

var mergeWebID int

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "Validate and merge a completed work web",
	Long: `Validates that all requirements in a work web are COMPLETE,
releases their claims, and optionally cleans up the worktree.

Examples:
    rtmx merge --web 1    # validate and merge web 1`,
	RunE: runMerge,
}

func init() {
	mergeCmd.Flags().IntVar(&mergeWebID, "web", 0, "work web ID to merge (required)")
	_ = mergeCmd.MarkFlagRequired("web")
	rootCmd.AddCommand(mergeCmd)
}

func runMerge(cmd *cobra.Command, args []string) error {
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

	if mergeWebID < 1 || mergeWebID > len(webs) {
		return fmt.Errorf("invalid web ID %d (have %d webs)", mergeWebID, len(webs))
	}

	web := webs[mergeWebID-1]

	// Check all requirements in the web are COMPLETE
	var incomplete []string
	for _, id := range web.IDs {
		req := db.Get(id)
		if req != nil && req.IsIncomplete() {
			incomplete = append(incomplete, id)
		}
	}

	if len(incomplete) > 0 {
		cmd.Printf("Cannot merge Web %d: %d incomplete requirement(s):\n", mergeWebID, len(incomplete))
		for _, id := range incomplete {
			req := db.Get(id)
			if req != nil {
				cmd.Printf("  %s  [%s]  %s\n", id, req.Status, output.Truncate(req.RequirementText, 50))
			}
		}
		return fmt.Errorf("web has incomplete requirements")
	}

	// Release all claims
	claimsDir := fmt.Sprintf("%s/.rtmx/claims", cwd)
	store, err := orchestration.NewClaimStore(claimsDir)
	if err != nil {
		return fmt.Errorf("failed to open claim store: %w", err)
	}

	released := 0
	for _, id := range web.IDs {
		claim, _ := store.Get(id)
		if claim != nil {
			_ = store.ForceRelease(id)
			released++
		}
	}

	// Clean up worktree if it exists
	_ = orchestration.RemoveWorktree(cwd, mergeWebID)

	cmd.Printf("Merged Web %d: %d requirement(s), %d claim(s) released\n",
		mergeWebID, len(web.IDs), released)
	return nil
}
