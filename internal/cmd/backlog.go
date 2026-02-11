package cmd

import (
	"github.com/spf13/cobra"
)

var (
	backlogView     string
	backlogPhase    int
	backlogCategory string
	backlogLimit    int
)

var backlogCmd = &cobra.Command{
	Use:   "backlog",
	Short: "Show prioritized backlog",
	Long: `Display the requirements backlog with various view modes.

View modes:
  all         All incomplete requirements (default)
  critical    High priority and blocking requirements
  quick-wins  Low effort, high value requirements
  blockers    Requirements blocking others
  list        Simple list format`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement backlog command
		cmd.Println("Backlog command not yet implemented")
		cmd.Printf("View: %s\n", backlogView)
		if backlogPhase > 0 {
			cmd.Printf("Phase filter: %d\n", backlogPhase)
		}
		if backlogCategory != "" {
			cmd.Printf("Category filter: %s\n", backlogCategory)
		}
		if backlogLimit > 0 {
			cmd.Printf("Limit: %d\n", backlogLimit)
		}
		return nil
	},
}

func init() {
	backlogCmd.Flags().StringVar(&backlogView, "view", "all", "view mode: all, critical, quick-wins, blockers, list")
	backlogCmd.Flags().IntVar(&backlogPhase, "phase", 0, "filter by phase number")
	backlogCmd.Flags().StringVar(&backlogCategory, "category", "", "filter by category")
	backlogCmd.Flags().IntVarP(&backlogLimit, "limit", "n", 0, "limit number of results")
}
