package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rtmx-ai/rtmx-go/internal/output"
	"github.com/spf13/cobra"
)

var (
	installDryRun     bool
	installYes        bool
	installForce      bool
	installAgents     []string
	installAll        bool
	installSkipBackup bool
	installHooks      bool
	installPrePush    bool
	installRemove     bool
	installValidate   bool
	installClaude     bool
	installList       bool
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install RTM-aware prompts into AI agent configs or git hooks",
	Long: `Inject RTMX context and commands into AI agent configs or git hooks.

Supported agents: claude, cursor, copilot, cline, gemini, windsurf, aider,
amazonq, zed, continue.

With --hooks, installs git hooks for automated validation.

Examples:
    rtmx install                    # Interactive selection
    rtmx install --all              # Install to all detected agents
    rtmx install --agents claude    # Install only to Claude
    rtmx install --agents cline,gemini  # Install to multiple agents
    rtmx install --list             # Show all supported agents
    rtmx install --dry-run          # Preview changes
    rtmx install --hooks            # Install pre-commit hook (health check)
    rtmx install --hooks --validate # Install validation pre-commit hook
    rtmx install --hooks --pre-push # Install both hooks
    rtmx install --hooks --remove   # Remove rtmx hooks
    rtmx install --claude           # Install Claude Code hooks
    rtmx install --claude --remove  # Remove Claude Code hooks`,
	RunE: runInstall,
}

func init() {
	installCmd.Flags().BoolVar(&installDryRun, "dry-run", false, "preview changes without writing")
	installCmd.Flags().BoolVarP(&installYes, "yes", "y", false, "skip confirmation prompts")
	installCmd.Flags().BoolVar(&installForce, "force", false, "overwrite existing RTMX sections")
	installCmd.Flags().StringSliceVar(&installAgents, "agents", nil, "specific agents to install (use --list to see all)")
	installCmd.Flags().BoolVar(&installAll, "all", false, "install to all detected agents")
	installCmd.Flags().BoolVar(&installSkipBackup, "skip-backup", false, "don't create backup files")
	installCmd.Flags().BoolVar(&installHooks, "hooks", false, "install git hooks instead of agent configs")
	installCmd.Flags().BoolVar(&installPrePush, "pre-push", false, "also install pre-push hook (requires --hooks)")
	installCmd.Flags().BoolVar(&installRemove, "remove", false, "remove installed hooks (requires --hooks)")
	installCmd.Flags().BoolVar(&installValidate, "validate", false, "install validation hook (requires --hooks)")
	installCmd.Flags().BoolVar(&installClaude, "claude", false, "install Claude Code hooks (.claude/hooks.json)")
	installCmd.Flags().BoolVar(&installList, "list", false, "list all supported agents and their detection status")

	rootCmd.AddCommand(installCmd)
}

// Git hook templates
const preCommitHookTemplate = `#!/bin/sh
# RTMX pre-commit hook
# Installed by: rtmx install --hooks

echo "Running RTMX health check..."
if command -v rtmx >/dev/null 2>&1; then
    rtmx health --strict
    if [ $? -ne 0 ]; then
        echo "RTMX health check failed. Commit aborted."
        echo "Run 'rtmx health' for details, or commit with --no-verify to skip."
        exit 1
    fi
else
    echo "Warning: rtmx not found in PATH, skipping health check"
fi
`

const prePushHookTemplate = `#!/bin/sh
# RTMX pre-push hook
# Installed by: rtmx install --hooks --pre-push

echo "Checking test marker compliance..."
if command -v pytest >/dev/null 2>&1; then
    # Count tests with @pytest.mark.req marker
    WITH_REQ=$(pytest tests/ --collect-only -q -m req 2>/dev/null | grep -c "::test_" || echo "0")
    TOTAL=$(pytest tests/ --collect-only -q 2>/dev/null | grep -c "::test_" || echo "0")

    if [ "$TOTAL" -gt 0 ]; then
        PCT=$((WITH_REQ * 100 / TOTAL))
        if [ "$PCT" -lt 80 ]; then
            echo "Test marker compliance is ${PCT}% (requires 80%)."
            echo "Push aborted. Add @pytest.mark.req() markers to tests."
            exit 1
        fi
        echo "Test marker compliance: ${PCT}%"
    fi
else
    echo "Warning: pytest not found in PATH, skipping marker check"
fi
`

const validationHookTemplate = `#!/bin/sh
# RTMX pre-commit validation hook
# Installed by: rtmx install --hooks --validate

# Get list of staged RTM CSV files
STAGED_RTM=$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.csv$')

if [ -n "$STAGED_RTM" ]; then
    echo "Validating staged RTM files..."
    if command -v rtmx >/dev/null 2>&1; then
        rtmx validate-staged $STAGED_RTM
        if [ $? -ne 0 ]; then
            echo "RTM validation failed. Commit aborted."
            echo "Fix validation errors above, or commit with --no-verify to skip."
            exit 1
        fi
    else
        echo "Warning: rtmx not found in PATH, skipping RTM validation"
    fi
fi
`

// Agent prompt templates
const claudePrompt = `
## RTMX Requirements Traceability

This project uses RTMX for requirements traceability management.

**Full patterns guide**: https://rtmx.ai/patterns

### Critical: Closed-Loop Verification

**Never manually edit the ` + "`status`" + ` field in rtm_database.csv.**

Status must be derived from test results using ` + "`rtmx verify --update`" + `.

` + "```bash" + `
# RIGHT: Let tests determine status
rtmx verify --update

# WRONG: Manual status edit in CSV or code
` + "```" + `

### Quick Commands
- ` + "`rtmx status`" + ` - Completion status (-v/-vv/-vvv for detail)
- ` + "`rtmx backlog`" + ` - Prioritized incomplete requirements
- ` + "`rtmx verify --update`" + ` - Run tests and update status from results
- ` + "`rtmx from-tests --update`" + ` - Sync test metadata to RTM
- ` + "`make rtm`" + ` / ` + "`make backlog`" + ` - Makefile shortcuts (if available)

### Development Workflow
1. Read requirement spec from ` + "`docs/requirements/`" + `
2. Write tests with ` + "`@pytest.mark.req(\"REQ-XX-NNN\")`" + `
3. Implement code to pass tests
4. Run ` + "`rtmx verify --update`" + ` (status updated automatically)
5. Commit changes

### Test Markers
| Marker | Purpose |
|--------|---------|
| ` + "`@pytest.mark.req(\"ID\")`" + ` | Link to requirement |
| ` + "`@pytest.mark.scope_unit`" + ` | Single component |
| ` + "`@pytest.mark.scope_integration`" + ` | Multi-component |
| ` + "`@pytest.mark.technique_nominal`" + ` | Happy path |
| ` + "`@pytest.mark.technique_stress`" + ` | Edge cases |

### Patterns and Anti-Patterns

| Do This | Not This |
|---------|----------|
| ` + "`rtmx verify --update`" + ` | Manual status edits |
| ` + "`@pytest.mark.req()`" + ` on tests | Orphan tests |
| Respect ` + "`blockedBy`" + ` deps | Ignore dependencies |
`

const cursorPrompt = `# RTMX Requirements Traceability

Full patterns guide: https://rtmx.ai/patterns

## Critical Rule
Never manually edit ` + "`status`" + ` in rtm_database.csv.
Use ` + "`rtmx verify --update`" + ` to derive status from test results.

## Context Commands
- rtmx status -v        # Category-level completion
- rtmx backlog          # What needs work
- rtmx verify --update  # Run tests, update status
- rtmx deps --req ID    # Requirement dependencies

## Test Generation Rules
When generating tests, add @pytest.mark.req("REQ-XX-NNN") markers.
Include scope markers (scope_unit, scope_integration, scope_system).
Reference: docs/requirements/ for requirement details.
`

const copilotPrompt = `# RTMX Requirements Traceability

This project uses RTMX for requirements traceability.
Full patterns guide: https://rtmx.ai/patterns

## Critical Rule
Never manually edit ` + "`status`" + ` in rtm_database.csv.
Use ` + "`rtmx verify --update`" + ` to derive status from test results.

## Test Markers
- @pytest.mark.req("REQ-XX-NNN") - Links test to requirement
- @pytest.mark.scope_unit/integration/system - Test scope

## Commands
- rtmx status - Check completion status
- rtmx backlog - See incomplete requirements
- rtmx verify --update - Update status from test results
`

// clinePrompt is the RTMX context for Cline (.clinerules).
const clinePrompt = `# RTMX Requirements Traceability

This project uses RTMX for requirements traceability.
Full patterns guide: https://rtmx.ai/patterns

## Critical Rule
Never manually edit ` + "`status`" + ` in rtm_database.csv.
Use ` + "`rtmx verify --update`" + ` to derive status from test results.

## Commands
- rtmx status - Check completion status
- rtmx backlog - See incomplete requirements
- rtmx verify --update - Update status from test results

## Test Markers
- @pytest.mark.req("REQ-XX-NNN") - Links test to requirement
- @pytest.mark.scope_unit/integration/system - Test scope
`

// geminiPrompt is the RTMX context for Gemini CLI (GEMINI.md).
const geminiPrompt = `# RTMX Requirements Traceability

This project uses RTMX for requirements traceability.
Full patterns guide: https://rtmx.ai/patterns

## Critical Rule
Never manually edit ` + "`status`" + ` in rtm_database.csv.
Use ` + "`rtmx verify --update`" + ` to derive status from test results.

## Commands
- rtmx status - Check completion status
- rtmx backlog - See incomplete requirements
- rtmx verify --update - Update status from test results

## Test Markers
- @pytest.mark.req("REQ-XX-NNN") - Links test to requirement
- @pytest.mark.scope_unit/integration/system - Test scope
`

// windsurfPrompt is the RTMX context for Windsurf/Cascade (.windsurfrules).
const windsurfPrompt = `# RTMX Requirements Traceability

This project uses RTMX for requirements traceability.
Full patterns guide: https://rtmx.ai/patterns

## Critical Rule
Never manually edit ` + "`status`" + ` in rtm_database.csv.
Use ` + "`rtmx verify --update`" + ` to derive status from test results.

## Commands
- rtmx status - Check completion status
- rtmx backlog - See incomplete requirements
- rtmx verify --update - Update status from test results

## Test Markers
- @pytest.mark.req("REQ-XX-NNN") - Links test to requirement
- @pytest.mark.scope_unit/integration/system - Test scope
`

// amazonqPrompt is the RTMX context for Amazon Q Developer (.amazonq/rules).
const amazonqPrompt = `# RTMX Requirements Traceability

This project uses RTMX for requirements traceability.
Full patterns guide: https://rtmx.ai/patterns

## Critical Rule
Never manually edit ` + "`status`" + ` in rtm_database.csv.
Use ` + "`rtmx verify --update`" + ` to derive status from test results.

## Commands
- rtmx status - Check completion status
- rtmx backlog - See incomplete requirements
- rtmx verify --update - Update status from test results

## Test Markers
- @pytest.mark.req("REQ-XX-NNN") - Links test to requirement
- @pytest.mark.scope_unit/integration/system - Test scope
`

// aiderPrompt is the RTMX context for Aider (.aider.conf.yml), YAML format.
const aiderPrompt = `# RTMX Requirements Traceability
# Full patterns guide: https://rtmx.ai/patterns
#
# Critical Rule:
#   Never manually edit 'status' in rtm_database.csv.
#   Use 'rtmx verify --update' to derive status from test results.
#
# Commands:
#   rtmx status          - Check completion status
#   rtmx backlog         - See incomplete requirements
#   rtmx verify --update - Update status from test results
#
# Test Markers:
#   @pytest.mark.req("REQ-XX-NNN") - Links test to requirement
#   @pytest.mark.scope_unit/integration/system - Test scope
read: []
`

// continuePrompt is the RTMX context for Continue.dev (.continue/config.yaml), YAML format.
const continuePrompt = `# RTMX Requirements Traceability
# Full patterns guide: https://rtmx.ai/patterns
#
# Critical Rule:
#   Never manually edit 'status' in rtm_database.csv.
#   Use 'rtmx verify --update' to derive status from test results.
#
# Commands:
#   rtmx status          - Check completion status
#   rtmx backlog         - See incomplete requirements
#   rtmx verify --update - Update status from test results
#
# Test Markers:
#   @pytest.mark.req("REQ-XX-NNN") - Links test to requirement
#   @pytest.mark.scope_unit/integration/system - Test scope
`

// zedInstructions is the RTMX context string injected into Zed's
// .zed/settings.json under the assistant.instructions key.
const zedInstructions = `RTMX Requirements Traceability. Full patterns guide: https://rtmx.ai/patterns. Critical Rule: Never manually edit 'status' in rtm_database.csv. Use 'rtmx verify --update' to derive status from test results. Commands: rtmx status (completion status), rtmx backlog (incomplete requirements), rtmx verify --update (update status from test results). Test Markers: @pytest.mark.req("REQ-XX-NNN") links test to requirement, @pytest.mark.scope_unit/integration/system for test scope.`

// agentInfo describes a supported AI agent.
type agentInfo struct {
	Name       string // internal key
	Label      string // display name
	ConfigFile string // relative config file path
	Format     string // markdown, yaml, json
}

// supportedAgents is the canonical registry of all supported AI agents.
var supportedAgents = []agentInfo{
	{Name: "claude", Label: "Claude Code", ConfigFile: "CLAUDE.md", Format: "markdown"},
	{Name: "cursor", Label: "Cursor", ConfigFile: ".cursorrules", Format: "markdown"},
	{Name: "copilot", Label: "GitHub Copilot", ConfigFile: ".github/copilot-instructions.md", Format: "markdown"},
	{Name: "cline", Label: "Cline", ConfigFile: ".clinerules", Format: "markdown"},
	{Name: "gemini", Label: "Gemini CLI", ConfigFile: "GEMINI.md", Format: "markdown"},
	{Name: "windsurf", Label: "Windsurf/Cascade", ConfigFile: ".windsurfrules", Format: "markdown"},
	{Name: "aider", Label: "Aider", ConfigFile: ".aider.conf.yml", Format: "yaml"},
	{Name: "amazonq", Label: "Amazon Q Developer", ConfigFile: ".amazonq/rules", Format: "markdown"},
	{Name: "zed", Label: "Zed Editor", ConfigFile: ".zed/settings.json", Format: "json"},
	{Name: "continue", Label: "Continue.dev", ConfigFile: ".continue/config.yaml", Format: "yaml"},
}

const rtmxHookMarker = "# RTMX"

func runInstall(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	if installList {
		return runAgentList(cmd)
	}

	if installClaude {
		return runClaudeInstall(cmd)
	}

	if installHooks {
		return runHooksInstall(cmd)
	}

	return runAgentInstall(cmd)
}

func runAgentList(cmd *cobra.Command) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	detected := detectAgentConfigs(cwd)

	cmd.Printf("%s\n", output.Color("Supported AI Agents (10):", output.Bold))
	cmd.Println()

	for _, agent := range supportedAgents {
		path := detected[agent.Name]
		status := output.Color("not detected", output.Dim)
		if path != "" {
			status = output.Color("detected", output.Green)
		}
		cmd.Printf("  %-20s %-25s %s  (%s)\n", agent.Name, agent.Label, status, agent.ConfigFile)
	}

	cmd.Println()
	cmd.Println("Use --agents <name> to install to specific agents, or --all to install to all.")
	return nil
}

// claudeHooksJSON is the template for .claude/hooks.json.
const claudeHooksJSON = `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": ".*",
        "command": "rtmx context --format claude"
      }
    ]
  }
}
`

func runClaudeInstall(cmd *cobra.Command) error {
	cmd.Println("=== RTMX Claude Code Hooks ===")
	cmd.Println()

	if installDryRun {
		cmd.Printf("%s\n", output.Color("DRY RUN - no files will be written", output.Yellow))
		cmd.Println()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	claudeDir := filepath.Join(cwd, ".claude")
	hooksPath := filepath.Join(claudeDir, "hooks.json")

	if installRemove {
		cmd.Printf("%s\n", output.Color("Removing Claude Code hooks...", output.Bold))
		if installDryRun {
			cmd.Printf("  Would remove: %s\n", hooksPath)
		} else {
			if _, err := os.Stat(hooksPath); os.IsNotExist(err) {
				cmd.Printf("  %s\n", output.Color("No hooks.json to remove", output.Dim))
			} else {
				if err := os.Remove(hooksPath); err != nil {
					cmd.Printf("  %s Failed to remove: %v\n", output.Color("Error:", output.Red), err)
				} else {
					cmd.Printf("  %s %s\n", output.Color("Removed:", output.Green), hooksPath)
				}
			}
		}
		cmd.Println()
		cmd.Printf("%s\n", output.Color("Claude Code hooks removed", output.Green))
		return nil
	}

	cmd.Printf("%s\n", output.Color("Installing Claude Code hooks...", output.Bold))

	if installDryRun {
		cmd.Printf("  Would create: %s\n", hooksPath)
	} else {
		// Create .claude directory if needed
		if err := os.MkdirAll(claudeDir, 0755); err != nil {
			return fmt.Errorf("failed to create .claude directory: %w", err)
		}

		if err := os.WriteFile(hooksPath, []byte(claudeHooksJSON), 0644); err != nil {
			return fmt.Errorf("failed to write hooks.json: %w", err)
		}

		cmd.Printf("  %s %s\n", output.Color("Created:", output.Green), hooksPath)
	}

	cmd.Println()
	cmd.Printf("%s\n", output.Color("Claude Code hooks installed", output.Green))
	cmd.Println()
	cmd.Println("The PreToolUse hook will inject RTM context into Claude Code conversations.")
	cmd.Println("Use 'rtmx install --claude --remove' to uninstall.")

	return nil
}

func runHooksInstall(cmd *cobra.Command) error {
	cmd.Println("=== RTMX Git Hooks ===")
	cmd.Println()

	if installDryRun {
		cmd.Printf("%s\n", output.Color("DRY RUN - no files will be written", output.Yellow))
		cmd.Println()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	gitDir := filepath.Join(cwd, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		cmd.Printf("%s Not in a git repository\n", output.Color("Error:", output.Red))
		cmd.Println("Initialize a git repository first with: git init")
		return nil
	}

	hooksDir := filepath.Join(gitDir, "hooks")

	// Create hooks directory if needed
	if !installDryRun {
		if err := os.MkdirAll(hooksDir, 0755); err != nil {
			return fmt.Errorf("failed to create hooks directory: %w", err)
		}
	}

	// Determine which hooks to process
	type hookInfo struct {
		name     string
		template string
	}

	var hooks []hookInfo

	if installValidate {
		hooks = append(hooks, hookInfo{"pre-commit", validationHookTemplate})
	} else {
		hooks = append(hooks, hookInfo{"pre-commit", preCommitHookTemplate})
	}

	if installPrePush {
		hooks = append(hooks, hookInfo{"pre-push", prePushHookTemplate})
	}

	if installRemove {
		cmd.Printf("%s\n", output.Color("Removing RTMX hooks...", output.Bold))
	} else {
		hookType := "health check"
		if installValidate {
			hookType = "validation"
		}
		cmd.Printf("%s %s hooks...\n", output.Color("Installing RTMX", output.Bold), hookType)
	}

	for _, hook := range hooks {
		hookPath := filepath.Join(hooksDir, hook.name)

		if installRemove {
			// Remove hook only if it's an RTMX hook
			if isRTMXHook(hookPath) {
				if installDryRun {
					cmd.Printf("  Would remove: %s\n", hookPath)
				} else {
					if err := os.Remove(hookPath); err != nil {
						cmd.Printf("  %s Failed to remove %s: %v\n", output.Color("Error:", output.Red), hook.name, err)
					} else {
						cmd.Printf("  %s %s\n", output.Color("Removed:", output.Green), hook.name)
					}
				}
			} else {
				cmd.Printf("  %s\n", output.Color(fmt.Sprintf("No RTMX hook to remove: %s", hook.name), output.Dim))
			}
		} else {
			// Install hook
			if _, err := os.Stat(hookPath); err == nil && !isRTMXHook(hookPath) && !installDryRun {
				// Backup existing non-RTMX hook
				timestamp := time.Now().Format("20060102-150405")
				backupPath := filepath.Join(hooksDir, fmt.Sprintf("%s.rtmx-backup-%s", hook.name, timestamp))
				if err := os.Rename(hookPath, backupPath); err != nil {
					cmd.Printf("  %s Failed to backup %s: %v\n", output.Color("Warning:", output.Yellow), hook.name, err)
				} else {
					cmd.Printf("  %s\n", output.Color(fmt.Sprintf("Backup: %s", backupPath), output.Dim))
				}
			}

			if installDryRun {
				cmd.Printf("  Would create: %s\n", hookPath)
			} else {
				if err := os.WriteFile(hookPath, []byte(hook.template), 0755); err != nil {
					cmd.Printf("  %s Failed to install %s: %v\n", output.Color("Error:", output.Red), hook.name, err)
					continue
				}
				cmd.Printf("  %s %s\n", output.Color("Installed:", output.Green), hookPath)
			}
		}
	}

	cmd.Println()
	if installRemove {
		cmd.Printf("%s\n", output.Color("Hooks removed", output.Green))
	} else {
		cmd.Printf("%s\n", output.Color("Hooks installed", output.Green))
		cmd.Println()
		cmd.Println("Hooks will run automatically on git commit/push.")
		cmd.Println("Use --no-verify to bypass hooks when needed.")
	}

	return nil
}

func isRTMXHook(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(content), rtmxHookMarker)
}

func runAgentInstall(cmd *cobra.Command) error {
	cmd.Println("=== RTMX Agent Installation ===")
	cmd.Println()

	if installDryRun {
		cmd.Printf("%s\n", output.Color("DRY RUN - no files will be written", output.Yellow))
		cmd.Println()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Detect agent configs
	detected := detectAgentConfigs(cwd)

	// Show detected configs (sorted for stable output)
	cmd.Printf("%s\n", output.Color("Detected agent configurations:", output.Bold))
	agentNames := make([]string, 0, len(detected))
	for agent := range detected {
		agentNames = append(agentNames, agent)
	}
	sort.Strings(agentNames)
	for _, agent := range agentNames {
		path := detected[agent]
		if path != "" {
			cmd.Printf("  %s %s: %s\n", output.Color("✓", output.Green), agent, path)
		} else {
			cmd.Printf("  %s %s: not found\n", output.Color("○", output.Dim), agent)
		}
	}
	cmd.Println()

	// Determine which agents to install
	var targetAgents []string
	if len(installAgents) > 0 {
		targetAgents = installAgents
	} else if installAll {
		for agent := range detected {
			targetAgents = append(targetAgents, agent)
		}
	} else if installYes {
		// Non-interactive: only install to existing agents
		for agent, path := range detected {
			if path != "" {
				targetAgents = append(targetAgents, agent)
			}
		}
	} else {
		// Interactive mode - for now, just use existing agents
		for agent, path := range detected {
			if path != "" {
				targetAgents = append(targetAgents, agent)
			}
		}
	}

	if len(targetAgents) == 0 {
		cmd.Printf("%s\n", output.Color("No agents selected", output.Yellow))
		return nil
	}

	// Install to each agent
	for _, agent := range targetAgents {
		cmd.Printf("%s %s...\n", output.Color("Installing to", output.Bold), agent)

		path := detected[agent]
		prompt := getAgentPrompt(agent)
		if prompt == "" {
			cmd.Printf("  %s\n", output.Color(fmt.Sprintf("Unknown agent: %s", agent), output.Red))
			continue
		}

		if path != "" {
			if err := installToExistingFile(cmd, agent, path, prompt); err != nil {
				cmd.Printf("  %s %v\n", output.Color("Error:", output.Red), err)
			}
		} else {
			if err := installToNewFile(cmd, cwd, agent, prompt); err != nil {
				cmd.Printf("  %s %v\n", output.Color("Error:", output.Red), err)
			}
		}
	}

	cmd.Println()
	cmd.Printf("%s\n", output.Color("✓ Installation complete", output.Green))

	return nil
}

func detectAgentConfigs(cwd string) map[string]string {
	configs := make(map[string]string)

	for _, agent := range supportedAgents {
		configs[agent.Name] = ""

		// Claude has an alternate path
		if agent.Name == "claude" {
			claudePaths := []string{
				filepath.Join(cwd, "CLAUDE.md"),
				filepath.Join(cwd, ".claude", "CLAUDE.md"),
			}
			for _, p := range claudePaths {
				if _, err := os.Stat(p); err == nil {
					configs["claude"] = p
					break
				}
			}
			continue
		}

		p := filepath.Join(cwd, agent.ConfigFile)
		if _, err := os.Stat(p); err == nil {
			configs[agent.Name] = p
		}
	}

	return configs
}

func getAgentPrompt(agent string) string {
	switch agent {
	case "claude":
		return claudePrompt
	case "cursor":
		return cursorPrompt
	case "copilot":
		return copilotPrompt
	case "cline":
		return clinePrompt
	case "gemini":
		return geminiPrompt
	case "windsurf":
		return windsurfPrompt
	case "aider":
		return aiderPrompt
	case "amazonq":
		return amazonqPrompt
	case "zed":
		return zedInstructions
	case "continue":
		return continuePrompt
	default:
		return ""
	}
}

// getAgentFormat returns the config format for the given agent name.
func getAgentFormat(name string) string {
	for _, a := range supportedAgents {
		if a.Name == name {
			return a.Format
		}
	}
	return ""
}

// getAgentConfigFile returns the relative config file path for the given agent name.
func getAgentConfigFile(name string) string {
	for _, a := range supportedAgents {
		if a.Name == name {
			return a.ConfigFile
		}
	}
	return ""
}

// installToExistingFile appends or injects RTMX content into an existing agent config file.
func installToExistingFile(cmd *cobra.Command, agent, path, prompt string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	format := getAgentFormat(agent)

	// Check if RTMX section already exists
	if containsRTMXSection(string(content), format) && !installForce {
		cmd.Printf("  %s\n", output.Color("RTMX section already exists (use --force to overwrite)", output.Yellow))
		return nil
	}

	// Create backup if needed
	if !installSkipBackup && !installDryRun {
		timestamp := time.Now().Format("20060102-150405")
		ext := filepath.Ext(path)
		backupPath := strings.TrimSuffix(path, ext) + fmt.Sprintf(".rtmx-backup-%s%s", timestamp, ext)
		if backupContent, readErr := os.ReadFile(path); readErr == nil {
			_ = os.WriteFile(backupPath, backupContent, 0644)
			cmd.Printf("  %s\n", output.Color(fmt.Sprintf("Backup: %s", backupPath), output.Dim))
		}
	}

	var newContent string

	switch format {
	case "json":
		newContent, err = injectZedJSON(string(content), prompt)
		if err != nil {
			return fmt.Errorf("failed to inject into JSON: %w", err)
		}
	default:
		// Markdown and YAML: append or replace
		newContent = string(content)
		if installForce && containsRTMXSection(newContent, format) {
			newContent = removeRTMXSection(newContent, format)
		}
		newContent = strings.TrimRight(newContent, "\n") + "\n" + strings.TrimSpace(prompt) + "\n"
	}

	if installDryRun {
		cmd.Printf("  Would update %s\n", path)
	} else {
		if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to update: %w", err)
		}
		cmd.Printf("  %s Updated %s\n", output.Color("✓", output.Green), path)
	}
	return nil
}

// installToNewFile creates a new agent config file with RTMX content.
func installToNewFile(cmd *cobra.Command, cwd, agent, prompt string) error {
	configFile := getAgentConfigFile(agent)
	if configFile == "" {
		return fmt.Errorf("unknown agent: %s", agent)
	}

	newPath := filepath.Join(cwd, configFile)

	// Create parent directories if needed
	if dir := filepath.Dir(newPath); dir != cwd {
		if !installDryRun {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		}
	}

	format := getAgentFormat(agent)

	var fileContent string
	switch format {
	case "json":
		// Create a new Zed settings.json with assistant.instructions
		fileContent = buildZedJSON(prompt)
	default:
		fileContent = strings.TrimSpace(prompt) + "\n"
	}

	if installDryRun {
		cmd.Printf("  Would create %s\n", newPath)
	} else {
		if err := os.WriteFile(newPath, []byte(fileContent), 0644); err != nil {
			return fmt.Errorf("failed to create: %w", err)
		}
		cmd.Printf("  %s Created %s\n", output.Color("✓", output.Green), newPath)
	}
	return nil
}

// containsRTMXSection checks whether the content already has RTMX context injected.
func containsRTMXSection(content, format string) bool {
	switch format {
	case "json":
		return strings.Contains(content, "RTMX Requirements Traceability")
	default:
		return strings.Contains(content, "RTMX Requirements Traceability")
	}
}

// removeRTMXSection removes an existing RTMX section from markdown/yaml content.
func removeRTMXSection(content, format string) string {
	_ = format
	lines := strings.Split(content, "\n")
	var newLines []string
	inRTMXSection := false
	for _, line := range lines {
		if strings.Contains(line, "## RTMX Requirements Traceability") || strings.Contains(line, "# RTMX Requirements Traceability") {
			inRTMXSection = true
			continue
		}
		if inRTMXSection && strings.HasPrefix(line, "## ") {
			inRTMXSection = false
		}
		if !inRTMXSection {
			newLines = append(newLines, line)
		}
	}
	return strings.Join(newLines, "\n")
}

// injectZedJSON injects the RTMX instructions into an existing Zed settings.json.
func injectZedJSON(content, instructions string) (string, error) {
	var settings map[string]interface{}
	if err := json.Unmarshal([]byte(content), &settings); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	assistant, ok := settings["assistant"].(map[string]interface{})
	if !ok {
		assistant = make(map[string]interface{})
		settings["assistant"] = assistant
	}
	assistant["instructions"] = instructions

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data) + "\n", nil
}

// buildZedJSON creates a new Zed settings.json with RTMX instructions.
func buildZedJSON(instructions string) string {
	settings := map[string]interface{}{
		"assistant": map[string]interface{}{
			"instructions": instructions,
		},
	}
	data, _ := json.MarshalIndent(settings, "", "  ")
	return string(data) + "\n"
}
