package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	loopOnce bool
	loopJSON bool
)

// loopEdgesOverride, when non-nil, replaces falling-edge detection.
// Tests inject edges; production leaves it nil until the watcher (REQ-ORCH-019a) is running.
var loopEdgesOverride *[]LoopEdge

var loopCmd = &cobra.Command{
	Use:   "loop",
	Short: "Run one delivery tick from a falling edge",
	Long: `Run the delivery loop tick: select the next unblocked requirement,
then decompose it, and print a JSON plan.

rtmx loop --once runs a single tick and exits. A falling edge must be
present; when next has nothing to claim the plan is {"idle":true}.

The tick does not create commits, branches, or pull requests.`,
	RunE: runLoop,
}

func init() {
	loopCmd.Flags().BoolVar(&loopOnce, "once", false, "run at most one tick and exit")
	loopCmd.Flags().BoolVar(&loopJSON, "json", false, "print the JSON plan (always on for --once)")
	rootCmd.AddCommand(loopCmd)
}

func runLoop(cmd *cobra.Command, args []string) error {
	if !loopOnce {
		return fmt.Errorf("rtmx loop --once runs one tick; the foreground watcher is not started without --once")
	}
	return runLoopOnce(cmd)
}
