package cmd

import (
	"fmt"
	"os"
	"path/filepath"
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
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install RTM-aware prompts into AI agent configs or git hooks",
	Long: `Inject RTMX context and commands into Claude, Cursor, or Copilot configs.
With --hooks, installs git hooks for automated validation.

Examples:
    rtmx install                    # Interactive selection
    rtmx install --all              # Install to all detected agents
    rtmx install --agents claude    # Install only to Claude
    rtmx install --dry-run          # Preview changes
    rtmx install --hooks            # Install pre-commit hook (health check)
    rtmx install --hooks --validate # Install validation pre-commit hook
    rtmx install --hooks --pre-push # Install both hooks
    rtmx install --hooks --remove   # Remove rtmx hooks`,
	RunE: runInstall,
}

func init() {
	installCmd.Flags().BoolVar(&installDryRun, "dry-run", false, "preview changes without writing")
	installCmd.Flags().BoolVarP(&installYes, "yes", "y", false, "skip confirmation prompts")
	installCmd.Flags().BoolVar(&installForce, "force", false, "overwrite existing RTMX sections")
	installCmd.Flags().StringSliceVar(&installAgents, "agents", nil, "specific agents to install (claude, cursor, copilot)")
	installCmd.Flags().BoolVar(&installAll, "all", false, "install to all detected agents")
	installCmd.Flags().BoolVar(&installSkipBackup, "skip-backup", false, "don't create backup files")
	installCmd.Flags().BoolVar(&installHooks, "hooks", false, "install git hooks instead of agent configs")
	installCmd.Flags().BoolVar(&installPrePush, "pre-push", false, "also install pre-push hook (requires --hooks)")
	installCmd.Flags().BoolVar(&installRemove, "remove", false, "remove installed hooks (requires --hooks)")
	installCmd.Flags().BoolVar(&installValidate, "validate", false, "install validation hook (requires --hooks)")

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

const rtmxHookMarker = "# RTMX"

func runInstall(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	if installHooks {
		return runHooksInstall(cmd)
	}

	return runAgentInstall(cmd)
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

	// Show detected configs
	cmd.Printf("%s\n", output.Color("Detected agent configurations:", output.Bold))
	for agent, path := range detected {
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

		if path != "" {
			// Check if RTMX section already exists
			content, err := os.ReadFile(path)
			if err == nil && strings.Contains(string(content), "RTMX Requirements Traceability") && !installForce {
				cmd.Printf("  %s\n", output.Color("RTMX section already exists (use --force to overwrite)", output.Yellow))
				continue
			}

			// Create backup if needed
			if !installSkipBackup && !installDryRun && path != "" {
				timestamp := time.Now().Format("20060102-150405")
				ext := filepath.Ext(path)
				backupPath := strings.TrimSuffix(path, ext) + fmt.Sprintf(".rtmx-backup-%s%s", timestamp, ext)
				if content, err := os.ReadFile(path); err == nil {
					_ = os.WriteFile(backupPath, content, 0644)
					cmd.Printf("  %s\n", output.Color(fmt.Sprintf("Backup: %s", backupPath), output.Dim))
				}
			}

			// Append RTMX section
			newContent := string(content)
			if installForce && strings.Contains(newContent, "RTMX Requirements Traceability") {
				// Remove existing section
				lines := strings.Split(newContent, "\n")
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
				newContent = strings.Join(newLines, "\n")
			}

			newContent = strings.TrimRight(newContent, "\n") + "\n" + strings.TrimSpace(prompt) + "\n"

			if installDryRun {
				cmd.Printf("  Would append %d characters\n", len(prompt))
			} else {
				if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
					cmd.Printf("  %s Failed to update: %v\n", output.Color("Error:", output.Red), err)
					continue
				}
				cmd.Printf("  %s Updated %s\n", output.Color("✓", output.Green), path)
			}
		} else {
			// Create new file
			var newPath string
			switch agent {
			case "claude":
				newPath = filepath.Join(cwd, "CLAUDE.md")
			case "cursor":
				newPath = filepath.Join(cwd, ".cursorrules")
			case "copilot":
				newPath = filepath.Join(cwd, ".github", "copilot-instructions.md")
				_ = os.MkdirAll(filepath.Dir(newPath), 0755)
			default:
				cmd.Printf("  %s\n", output.Color(fmt.Sprintf("Unknown agent: %s", agent), output.Red))
				continue
			}

			if installDryRun {
				cmd.Printf("  Would create %s\n", newPath)
			} else {
				if err := os.WriteFile(newPath, []byte(strings.TrimSpace(prompt)+"\n"), 0644); err != nil {
					cmd.Printf("  %s Failed to create: %v\n", output.Color("Error:", output.Red), err)
					continue
				}
				cmd.Printf("  %s Created %s\n", output.Color("✓", output.Green), newPath)
			}
		}
	}

	cmd.Println()
	cmd.Printf("%s\n", output.Color("✓ Installation complete", output.Green))

	return nil
}

func detectAgentConfigs(cwd string) map[string]string {
	configs := make(map[string]string)

	// Claude
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
	if configs["claude"] == "" {
		configs["claude"] = ""
	}

	// Cursor
	cursorPath := filepath.Join(cwd, ".cursorrules")
	if _, err := os.Stat(cursorPath); err == nil {
		configs["cursor"] = cursorPath
	} else {
		configs["cursor"] = ""
	}

	// Copilot
	copilotPath := filepath.Join(cwd, ".github", "copilot-instructions.md")
	if _, err := os.Stat(copilotPath); err == nil {
		configs["copilot"] = copilotPath
	} else {
		configs["copilot"] = ""
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
	default:
		return ""
	}
}
