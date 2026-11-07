package cmd

import (
	"os/exec"

	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/spf13/cobra"
)

var (
	bootstrapAgentsDryRun  bool
	bootstrapAgentsForce   bool
	bootstrapAgentsEmit    bool
	bootstrapAgentsMD      bool
	bootstrapAgentsTargets []string
)

var bootstrapAgentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "One-shot agent MCP setup (marketplace + optional emit)",
	Long: `Print marketplace / deeplink install commands, remind PATH/brew, and
optionally run offline plugin emit so agents can reach next+claim quickly
(REQ-MONO-063f / first-session success).

  rtmx bootstrap agents
  rtmx bootstrap agents --dry-run
  rtmx bootstrap agents --emit=false          # guidance only
  rtmx bootstrap agents --agents cursor,codex --agents-md
  rtmx bootstrap agents --force               # overwrite emit outputs

Emit works offline (air-gapped). Marketplace commands need network when the
user runs them. Existing AGENTS.md is not overwritten unless --force.`,
	RunE: runBootstrapAgents,
}

func init() {
	bootstrapAgentsCmd.Flags().BoolVar(&bootstrapAgentsDryRun, "dry-run", false, "preview without writing files")
	bootstrapAgentsCmd.Flags().BoolVar(&bootstrapAgentsForce, "force", false, "overwrite existing emit outputs / AGENTS.md")
	bootstrapAgentsCmd.Flags().BoolVar(&bootstrapAgentsEmit, "emit", true, "run rtmx plugin emit for selected agents (offline)")
	bootstrapAgentsCmd.Flags().BoolVar(&bootstrapAgentsMD, "agents-md", false, "also write optional AGENTS.md RTMX nudge")
	bootstrapAgentsCmd.Flags().StringSliceVar(&bootstrapAgentsTargets, "agents", []string{"claude", "cursor", "codex"}, "emit targets when --emit")
	bootstrapCmd.AddCommand(bootstrapAgentsCmd)

	bootstrapCmd.Long += `

Agent MCP one-shot (marketplace + optional offline emit):
    rtmx bootstrap agents
    rtmx bootstrap agents --dry-run
    rtmx bootstrap agents --agents-md`
}

func runBootstrapAgents(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	cmd.Println("=== RTMX bootstrap agents ===")
	cmd.Println()
	cmd.Println("Goal: agent can call MCP next + claim within minutes")
	cmd.Println("(see docs/FIRST_SESSION_SUCCESS.md).")
	cmd.Println()

	printBootstrapAgentsPATH(cmd)
	cmd.Println()
	printBootstrapAgentsMarketplace(cmd)
	cmd.Println()

	if bootstrapAgentsDryRun {
		cmd.Printf("%s\n\n", output.Color("DRY RUN - no files will be written", output.Yellow))
	}

	if !bootstrapAgentsEmit {
		cmd.Println("Emit skipped (--emit=false). Offline path:")
		cmd.Println("  rtmx plugin emit --agents claude,cursor,codex")
		cmd.Println()
		cmd.Println("Done. Install marketplace or emit, then ask the agent to next + claim.")
		return nil
	}

	// Drive plugin emit via shared flags (offline; no network).
	emitDryRun = bootstrapAgentsDryRun
	emitForce = bootstrapAgentsForce
	emitAgentsMD = bootstrapAgentsMD
	emitAll = false
	emitList = false
	emitAgents = append([]string(nil), bootstrapAgentsTargets...)

	cmd.Println("Offline emit (works air-gapped):")
	if err := runPluginEmit(cmd, nil); err != nil {
		return err
	}
	cmd.Println()
	cmd.Println("Done. Prefer marketplace when online; emit covers air-gap.")
	cmd.Println("Ask the agent to run next, then claim.")
	return nil
}

func printBootstrapAgentsPATH(cmd *cobra.Command) {
	cmd.Println("Prerequisite — rtmx on PATH:")
	if path, err := exec.LookPath("rtmx"); err == nil {
		cmd.Printf("  %s %s\n", output.Color("Found:", output.Green), path)
	} else {
		cmd.Printf("  %s rtmx not found on PATH\n", output.Color("Missing:", output.Yellow))
	}
	cmd.Println("  brew install rtmx-ai/tap/rtmx   # macOS / Linux")
	cmd.Println("  # or: scoop / go install / npm i -g rtmx")
}

func printBootstrapAgentsMarketplace(cmd *cobra.Command) {
	cmd.Println("Marketplace / one-click (network):")
	cmd.Println("  claude plugin marketplace add rtmx-ai/rtmx && claude plugin install rtmx@rtmx")
	cmd.Println("  # Cursor: https://cursor.com/link/mcp?name=rtmx&config=eyJjb21tYW5kIjoicnRteCIsImFyZ3MiOlsibWNwLXNlcnZlciIsIi0tc3RkaW8iXX0%3D")
	cmd.Println("  # cursor.directory: https://cursor.directory/plugins/rtmx")
	cmd.Println("  codex mcp add rtmx -- rtmx mcp-server --stdio")
}
