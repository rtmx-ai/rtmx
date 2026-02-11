package cmd

import (
	"github.com/spf13/cobra"
)

var statusVerbosity int

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show RTM completion status",
	Long: `Display the current status of the Requirements Traceability Matrix.

The verbosity level controls how much detail is shown:
  (default)  Summary statistics only
  -v         Show status by category
  -vv        Show status by category and phase
  -vvv       Show individual requirement details`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Implement status command
		cmd.Println("Status command not yet implemented")
		cmd.Printf("Verbosity level: %d\n", statusVerbosity)
		return nil
	},
}

func init() {
	statusCmd.Flags().CountVarP(&statusVerbosity, "verbose", "v", "increase verbosity (-v, -vv, -vvv)")
}
