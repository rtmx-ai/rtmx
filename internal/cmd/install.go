package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/rtmx-ai/rtmx/internal/output"
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
	installCoder      bool
	installCodex      bool
	installGastown    bool
	installGeminiCLI  bool
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
	installCmd.Flags().BoolVar(&installCoder, "coder", false, "generate Coder workspace template with MCP server")
	installCmd.Flags().BoolVar(&installCodex, "codex", false, "generate OpenAI Codex CLI tool definition")
	installCmd.Flags().BoolVar(&installGastown, "gastown", false, "generate Gastown plugin config")
	installCmd.Flags().BoolVar(&installGeminiCLI, "gemini-cli", false, "generate Gemini CLI extension config")
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

// rtmxPromptCore is the standardized agent prompt content shared across all
// agent integrations. Language-agnostic, includes closed-loop verification
// workflow with dependency ordering.
const rtmxPromptCore = `This project uses RTMX for requirements traceability management.

Full patterns guide: https://rtmx.ai/patterns

### Critical: Closed-Loop Verification

Never manually edit the status field in the RTM database.
Status must be derived from test results:

    rtmx verify --command "npm test" --update        # Node.js
    rtmx verify --command "python -m pytest" --update # Python
    rtmx verify --command "go test ./..." --update    # Go
    rtmx verify --command "cargo test" --update       # Rust

If no --command is given, rtmx auto-detects the test runner from
project files (package.json, pyproject.toml, Cargo.toml, Makefile, etc).

### Development Workflow

1. Run rtmx backlog to see prioritized incomplete requirements
2. Run rtmx next to get the next unblocked requirement
3. Read the requirement spec in .rtmx/requirements/
4. Write tests that exercise the acceptance criteria
5. Implement code to pass the tests
6. Run rtmx verify --command "<test_command>" --update
7. Commit with the requirement ID in the message
8. Repeat from step 2

Requirements are discrete batches of work ordered by dependency.
Always respect dependency ordering -- do not skip ahead.

### Commands

- rtmx status      -- completion status (-v/-vv/-vvv for detail)
- rtmx backlog     -- prioritized incomplete requirements
- rtmx next        -- next unblocked requirement to work on
- rtmx verify      -- run tests and update status from results
- rtmx deps REQ-ID -- dependency graph for a requirement
- rtmx health      -- test coverage and traceability health

### Patterns and Anti-Patterns

Do: rtmx verify --update         Not: manual status edits
Do: respect dependency ordering   Not: skip blocked requirements
Do: one requirement per cycle     Not: batch multiple at once
Do: commit with REQ-ID            Not: orphan commits`

// Agent prompt templates -- each wraps rtmxPromptCore with the appropriate
// format (markdown with heading levels, YAML comments, plain text).

const claudePrompt = `
## RTMX Requirements Traceability

` + rtmxPromptCore + `

### MCP Tools (if configured)

When an MCP server is configured, these tools are available:

- mcp__rtmx__backlog -- prioritized backlog with dependency order
- mcp__rtmx__next    -- next unblocked requirement
- mcp__rtmx__claim   -- mark a requirement as in-progress
- mcp__rtmx__status  -- completion status
- mcp__rtmx__verify  -- run tests and update requirements
- mcp__rtmx__deps    -- dependency graph for a requirement
- mcp__rtmx__health  -- test coverage and traceability health
`

const cursorPrompt = `# RTMX Requirements Traceability

` + rtmxPromptCore + `
`

const copilotPrompt = `# RTMX Requirements Traceability

` + rtmxPromptCore + `
`

// clinePrompt is the RTMX context for Cline (.clinerules).
const clinePrompt = `# RTMX Requirements Traceability

` + rtmxPromptCore + `
`

// geminiPrompt is the RTMX context for Gemini CLI (GEMINI.md).
const geminiPrompt = `# RTMX Requirements Traceability

` + rtmxPromptCore + `
`

// windsurfPrompt is the RTMX context for Windsurf/Cascade (.windsurfrules).
const windsurfPrompt = `# RTMX Requirements Traceability

` + rtmxPromptCore + `
`

// amazonqPrompt is the RTMX context for Amazon Q Developer (.amazonq/rules).
const amazonqPrompt = `# RTMX Requirements Traceability

` + rtmxPromptCore + `
`

// aiderPrompt is the RTMX context for Aider (.aider.conf.yml), YAML format.
const aiderPrompt = `# RTMX Requirements Traceability
# Full patterns guide: https://rtmx.ai/patterns
#
# Critical Rule:
#   Never manually edit status in the RTM database.
#   Use 'rtmx verify --command "<test_cmd>" --update' to derive status.
#
# Workflow:
#   1. rtmx backlog       - see prioritized work
#   2. rtmx next          - get next unblocked requirement
#   3. Read spec, write tests, implement
#   4. rtmx verify --update  - run tests, update status
#   5. Commit with REQ-ID
#
# Commands:
#   rtmx status          - completion status
#   rtmx backlog         - prioritized incomplete requirements
#   rtmx next            - next unblocked requirement
#   rtmx verify --update - run tests and update status
#   rtmx deps REQ-ID     - dependency graph
read: []
`

// continuePrompt is the RTMX context for Continue.dev (.continue/config.yaml), YAML format.
const continuePrompt = `# RTMX Requirements Traceability
# Full patterns guide: https://rtmx.ai/patterns
#
# Critical Rule:
#   Never manually edit status in the RTM database.
#   Use 'rtmx verify --command "<test_cmd>" --update' to derive status.
#
# Workflow:
#   1. rtmx backlog       - see prioritized work
#   2. rtmx next          - get next unblocked requirement
#   3. Read spec, write tests, implement
#   4. rtmx verify --update  - run tests, update status
#   5. Commit with REQ-ID
#
# Commands:
#   rtmx status          - completion status
#   rtmx backlog         - prioritized incomplete requirements
#   rtmx next            - next unblocked requirement
#   rtmx verify --update - run tests and update status
#   rtmx deps REQ-ID     - dependency graph
`

// zedInstructions is the RTMX context string injected into Zed's
// .zed/settings.json under the assistant.instructions key.
const zedInstructions = `RTMX Requirements Traceability. Full patterns guide: https://rtmx.ai/patterns. Critical Rule: Never manually edit status in the RTM database. Use 'rtmx verify --command "<test_cmd>" --update' to derive status from test results. Workflow: rtmx backlog (prioritized work), rtmx next (next unblocked requirement), implement with tests, rtmx verify --update (run tests, update status), commit with REQ-ID. Commands: rtmx status, rtmx backlog, rtmx next, rtmx verify --update, rtmx deps REQ-ID.`

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

	if installCoder {
		return runCoderInstall(cmd)
	}

	if installCodex {
		return runCodexInstall(cmd)
	}

	if installGastown {
		return runGastownInstall(cmd)
	}

	if installGeminiCLI {
		return runGeminiCLIInstall(cmd)
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
		cmd.Printf("%s\n", output.Color("Removing Claude Code hooks and skills...", output.Bold))
		if installDryRun {
			cmd.Printf("  Would remove: %s\n", hooksPath)
			for name := range claudeSkillDefinitions() {
				cmd.Printf("  Would remove: %s\n", filepath.Join(claudeDir, "skills", name))
			}
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
			for name := range claudeSkillDefinitions() {
				skillDir := filepath.Join(claudeDir, "skills", name)
				if err := os.RemoveAll(skillDir); err == nil {
					cmd.Printf("  %s %s\n", output.Color("Removed:", output.Green), skillDir)
				}
			}
		}
		cmd.Println()
		cmd.Printf("%s\n", output.Color("Claude Code hooks and skills removed", output.Green))
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

	// Install skill pack
	cmd.Println()
	cmd.Printf("%s\n", output.Color("Installing Claude Code skill pack...", output.Bold))

	skills := claudeSkillDefinitions()
	for name, content := range skills {
		skillDir := filepath.Join(claudeDir, "skills", name)
		skillPath := filepath.Join(skillDir, "SKILL.md")

		if installDryRun {
			cmd.Printf("  Would create: %s\n", skillPath)
		} else {
			if err := os.MkdirAll(skillDir, 0755); err != nil {
				return fmt.Errorf("failed to create skill directory %s: %w", name, err)
			}
			if err := os.WriteFile(skillPath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write skill %s: %w", name, err)
			}
			cmd.Printf("  %s %s\n", output.Color("Created:", output.Green), skillPath)
		}
	}

	cmd.Println()
	cmd.Printf("%s\n", output.Color("Claude Code hooks and skills installed", output.Green))
	cmd.Println()
	cmd.Printf("Available slash commands: %s\n", output.Color("/rtmx-status /rtmx-backlog /rtmx-next /rtmx-verify /rtmx-claim", output.Cyan))
	cmd.Println("Use 'rtmx install --claude --remove' to uninstall.")

	return nil
}

// claudeSkillDefinitions returns the skill pack for Claude Code.
// Each key is a directory name under .claude/skills/, each value is SKILL.md content.
func claudeSkillDefinitions() map[string]string {
	return map[string]string{
		"rtmx-status": `---
name: rtmx-status
description: Show RTM completion status with progress bars, category breakdown, and version grouping.
argument-hint: "[-v|-vv|-vvv] [--by-version] [--json]"
---

Show the current RTM status. Pass arguments to control verbosity.

` + "```!" + `
rtmx status $ARGUMENTS
` + "```" + `
`,
		"rtmx-backlog": `---
name: rtmx-backlog
description: Show prioritized backlog with critical path, quick wins, and blocking analysis.
argument-hint: "[--json] [--version VERSION]"
---

Show the prioritized backlog of incomplete requirements.

` + "```!" + `
rtmx backlog $ARGUMENTS
` + "```" + `
`,
		"rtmx-next": `---
name: rtmx-next
description: Show independent work webs and pick the next highest-priority unblocked requirement.
argument-hint: "[--one] [--json]"
---

Analyze the dependency graph for parallelizable work. Use --one to pick a single requirement.

` + "```!" + `
rtmx next $ARGUMENTS
` + "```" + `
`,
		"rtmx-verify": `---
name: rtmx-verify
description: Run tests and update RTM status based on results. Use --audit to check for stale test references.
argument-hint: "[--update] [--audit] [--dry-run]"
---

Verify requirements by running tests and updating the RTM database.

` + "```!" + `
rtmx verify $ARGUMENTS
` + "```" + `
`,
		"rtmx-claim": `---
name: rtmx-claim
description: Claim a requirement for work. Prevents other agents from working on the same requirement.
argument-hint: "<req-id>"
arguments: [req_id]
---

Claim requirement $req_id for this agent session.

` + "```!" + `
rtmx next --one --json
` + "```" + `

If a specific requirement ID was provided, claim it. Otherwise use the output above to identify the highest-priority unblocked requirement and begin working on it.
`,
	}
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

	// Cursor MCP config (REQ-PLUGIN-003a)
	for _, agent := range targetAgents {
		if agent == "cursor" {
			if err := installCursorMCP(cmd, cwd); err != nil {
				cmd.Printf("  %s Cursor MCP: %v\n", output.Color("Error:", output.Red), err)
			}
		}
	}

	cmd.Println()
	cmd.Printf("%s\n", output.Color("✓ Installation complete", output.Green))

	return nil
}

const cursorMCPJSON = `{
  "mcpServers": {
    "rtmx": {
      "command": "rtmx",
      "args": ["mcp-server", "--port", "0"],
      "env": {}
    }
  }
}
`

func installCursorMCP(cmd *cobra.Command, cwd string) error {
	cursorDir := filepath.Join(cwd, ".cursor")
	mcpPath := filepath.Join(cursorDir, "mcp.json")

	if installDryRun {
		cmd.Printf("  Would create: %s\n", mcpPath)
		return nil
	}

	if err := os.MkdirAll(cursorDir, 0755); err != nil {
		return fmt.Errorf("failed to create .cursor directory: %w", err)
	}

	if err := os.WriteFile(mcpPath, []byte(cursorMCPJSON), 0644); err != nil {
		return fmt.Errorf("failed to write mcp.json: %w", err)
	}

	cmd.Printf("  %s %s\n", output.Color("Created:", output.Green), mcpPath)
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

// coderSetupScript is the workspace template for Coder integration.
const coderSetupScript = `#!/bin/bash
# RTMX Coder Workspace Setup
# Generated by: rtmx install --coder
# Starts RTMX MCP server on HTTP/SSE for remote agent access.

set -euo pipefail

echo "Setting up RTMX in Coder workspace..."

# Install rtmx if not present
if ! command -v rtmx >/dev/null 2>&1; then
    echo "Installing rtmx..."
    go install github.com/rtmx-ai/rtmx/cmd/rtmx@latest
fi

# Start MCP server on HTTP/SSE transport for remote agent access
echo "Starting RTMX MCP server (HTTP/SSE)..."
rtmx mcp-server --transport http --port 8484 &
MCP_PID=$!
echo "RTMX MCP server started (PID: $MCP_PID, port: 8484)"

# Write connection info for agents
cat > .rtmx-mcp.json <<EOF
{
  "transport": "http",
  "url": "http://localhost:8484",
  "pid": $MCP_PID
}
EOF

echo "RTMX workspace setup complete."
echo "MCP server accessible at http://localhost:8484"
`

// codexToolDefinition is the tool definition for OpenAI Codex CLI.
const codexToolDefinition = `{
  "name": "rtmx",
  "description": "Requirements traceability management. Query project status, backlog, dependencies, and verify requirements against test evidence.",
  "tools": [
    {
      "name": "rtmx_status",
      "description": "Show RTM completion status with progress bars and category breakdown.",
      "command": "rtmx status --json"
    },
    {
      "name": "rtmx_backlog",
      "description": "Show prioritized backlog of incomplete requirements.",
      "command": "rtmx backlog --json"
    },
    {
      "name": "rtmx_verify",
      "description": "Run tests and update requirement status based on results.",
      "command": "rtmx verify --update --json"
    },
    {
      "name": "rtmx_deps",
      "description": "Show dependency graph for a requirement.",
      "command": "rtmx deps --json"
    },
    {
      "name": "rtmx_next",
      "description": "Show the next highest-priority unblocked requirement.",
      "command": "rtmx next --one --json"
    }
  ]
}
`

// gastownPluginConfig is the MCP-based plugin config for Gastown.
const gastownPluginConfig = `{
  "name": "rtmx",
  "version": "1.0.0",
  "description": "RTMX requirements traceability plugin for Gastown",
  "transport": {
    "type": "stdio",
    "command": "rtmx",
    "args": ["mcp-server"]
  },
  "capabilities": {
    "tools": true,
    "resources": true
  }
}
`

func runCoderInstall(cmd *cobra.Command) error {
	cmd.Println("=== RTMX Coder Workspace Template ===")
	cmd.Println()

	if installDryRun {
		cmd.Printf("%s\n", output.Color("DRY RUN - no files will be written", output.Yellow))
		cmd.Println()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	coderDir := filepath.Join(cwd, "templates", "coder")
	scriptPath := filepath.Join(coderDir, "setup.sh")

	if installDryRun {
		cmd.Printf("  Would create: %s\n", scriptPath)
	} else {
		if err := os.MkdirAll(coderDir, 0755); err != nil {
			return fmt.Errorf("failed to create templates/coder directory: %w", err)
		}
		if err := os.WriteFile(scriptPath, []byte(coderSetupScript), 0755); err != nil {
			return fmt.Errorf("failed to write setup.sh: %w", err)
		}
		cmd.Printf("  %s %s\n", output.Color("Created:", output.Green), scriptPath)
	}

	cmd.Println()
	cmd.Printf("%s\n", output.Color("Coder workspace template installed", output.Green))
	cmd.Println("Copy templates/coder/ into your Coder template repository.")
	return nil
}

func runCodexInstall(cmd *cobra.Command) error {
	cmd.Println("=== RTMX OpenAI Codex Integration ===")
	cmd.Println()

	if installDryRun {
		cmd.Printf("%s\n", output.Color("DRY RUN - no files will be written", output.Yellow))
		cmd.Println()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	codexDir := filepath.Join(cwd, "templates", "codex")
	toolPath := filepath.Join(codexDir, "rtmx-tools.json")

	if installDryRun {
		cmd.Printf("  Would create: %s\n", toolPath)
	} else {
		if err := os.MkdirAll(codexDir, 0755); err != nil {
			return fmt.Errorf("failed to create templates/codex directory: %w", err)
		}
		if err := os.WriteFile(toolPath, []byte(codexToolDefinition), 0644); err != nil {
			return fmt.Errorf("failed to write rtmx-tools.json: %w", err)
		}
		cmd.Printf("  %s %s\n", output.Color("Created:", output.Green), toolPath)
	}

	cmd.Println()
	cmd.Printf("%s\n", output.Color("Codex tool definition installed", output.Green))
	cmd.Println("All commands use --json for structured output.")
	return nil
}

func runGastownInstall(cmd *cobra.Command) error {
	cmd.Println("=== RTMX Gastown Plugin ===")
	cmd.Println()

	if installDryRun {
		cmd.Printf("%s\n", output.Color("DRY RUN - no files will be written", output.Yellow))
		cmd.Println()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	gastownDir := filepath.Join(cwd, "templates", "gastown")
	pluginPath := filepath.Join(gastownDir, "rtmx-plugin.json")

	if installDryRun {
		cmd.Printf("  Would create: %s\n", pluginPath)
	} else {
		if err := os.MkdirAll(gastownDir, 0755); err != nil {
			return fmt.Errorf("failed to create templates/gastown directory: %w", err)
		}
		if err := os.WriteFile(pluginPath, []byte(gastownPluginConfig), 0644); err != nil {
			return fmt.Errorf("failed to write rtmx-plugin.json: %w", err)
		}
		cmd.Printf("  %s %s\n", output.Color("Created:", output.Green), pluginPath)
	}

	cmd.Println()
	cmd.Printf("%s\n", output.Color("Gastown plugin config installed", output.Green))
	cmd.Println("Uses MCP stdio transport for tool discovery.")
	return nil
}

// geminiCLIExtensionConfig is the Gemini CLI extension configuration
// that registers RTMX as an MCP tool provider via stdio transport.
const geminiCLIExtensionConfig = `{
  "name": "rtmx",
  "description": "Requirements traceability management via RTMX",
  "transport": {
    "type": "stdio",
    "command": "rtmx",
    "args": ["mcp", "serve"]
  },
  "tools": [
    {"name": "rtmx_status", "description": "Show RTM completion status"},
    {"name": "rtmx_backlog", "description": "Show prioritized backlog"},
    {"name": "rtmx_next", "description": "Get next claimable requirement"},
    {"name": "rtmx_claim", "description": "Claim a requirement for work"},
    {"name": "rtmx_verify", "description": "Verify requirement completion"}
  ]
}
`

func runGeminiCLIInstall(cmd *cobra.Command) error {
	cmd.Println("=== RTMX Gemini CLI Extension ===")
	cmd.Println()

	if installDryRun {
		cmd.Printf("%s\n", output.Color("DRY RUN - no files will be written", output.Yellow))
		cmd.Println()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	geminiDir := filepath.Join(cwd, "templates", "gemini-cli")
	configPath := filepath.Join(geminiDir, "rtmx-extension.json")

	if installDryRun {
		cmd.Printf("  Would create: %s\n", configPath)
	} else {
		if err := os.MkdirAll(geminiDir, 0755); err != nil {
			return fmt.Errorf("failed to create templates/gemini-cli directory: %w", err)
		}
		if err := os.WriteFile(configPath, []byte(geminiCLIExtensionConfig), 0644); err != nil {
			return fmt.Errorf("failed to write rtmx-extension.json: %w", err)
		}
		cmd.Printf("  %s %s\n", output.Color("Created:", output.Green), configPath)
	}

	cmd.Println()
	cmd.Printf("%s\n", output.Color("Gemini CLI extension config installed", output.Green))
	cmd.Println("Uses MCP stdio transport for tool discovery.")
	return nil
}
