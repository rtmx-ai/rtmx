package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/output"
	"github.com/spf13/cobra"
)

var (
	setupDryRun      bool
	setupMinimal     bool
	setupForce       bool
	setupSkipAgents  bool
	setupSkipMakefile bool
	setupBranch      bool
	setupPR          bool
	setupScaffold    bool
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Complete rtmx setup in one command",
	Long: `Perform full integration: config, RTM, agent prompts, Makefile.
Safe to run multiple times (idempotent). Creates backups before modifying files.

Examples:
    rtmx setup              # Full setup with smart defaults
    rtmx setup --dry-run    # Preview what would happen
    rtmx setup --minimal    # Just config and RTM database
    rtmx setup --branch     # Create git branch for review workflow
    rtmx setup --pr         # Create branch and pull request
    rtmx setup --scaffold   # Generate spec files for all requirements`,
	RunE: runSetup,
}

func init() {
	setupCmd.Flags().BoolVar(&setupDryRun, "dry-run", false, "preview changes without making them")
	setupCmd.Flags().BoolVar(&setupMinimal, "minimal", false, "only create config and RTM, skip agents and Makefile")
	setupCmd.Flags().BoolVar(&setupForce, "force", false, "overwrite existing files")
	setupCmd.Flags().BoolVar(&setupSkipAgents, "skip-agents", false, "skip agent config injection")
	setupCmd.Flags().BoolVar(&setupSkipMakefile, "skip-makefile", false, "skip Makefile targets")
	setupCmd.Flags().BoolVar(&setupBranch, "branch", false, "create git branch for isolation")
	setupCmd.Flags().BoolVar(&setupPR, "pr", false, "create pull request after setup (implies --branch)")
	setupCmd.Flags().BoolVar(&setupScaffold, "scaffold", false, "auto-generate requirement spec files from database entries")

	rootCmd.AddCommand(setupCmd)
}

// SetupResult holds the result of setup operation.
type SetupResult struct {
	Success        bool     `json:"success"`
	StepsCompleted []string `json:"steps_completed"`
	StepsSkipped   []string `json:"steps_skipped"`
	FilesCreated   []string `json:"files_created"`
	FilesModified  []string `json:"files_modified"`
	FilesBackedUp  []string `json:"files_backed_up"`
	Warnings       []string `json:"warnings"`
	Errors         []string `json:"errors"`
	BranchName     string   `json:"branch_name,omitempty"`
	RollbackPoint  string   `json:"rollback_point,omitempty"`
	PRUrl          string   `json:"pr_url,omitempty"`
}

func runSetup(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	result := &SetupResult{
		Success:        false,
		StepsCompleted: []string{},
		StepsSkipped:   []string{},
		FilesCreated:   []string{},
		FilesModified:  []string{},
		FilesBackedUp:  []string{},
		Warnings:       []string{},
		Errors:         []string{},
	}

	// If --pr specified, enable branch mode
	if setupPR && !setupBranch {
		setupBranch = true
	}

	cmd.Println(output.Header("RTMX Setup", 60))
	cmd.Println()
	cmd.Printf("Project: %s\n", cwd)
	if setupDryRun {
		cmd.Printf("%s\n", output.Color("DRY RUN - no changes will be made", output.Yellow))
	}
	if setupBranch {
		cmd.Printf("%s\n", output.Color("Branch mode: changes will be on a new branch", output.Cyan))
	}
	if setupPR {
		cmd.Printf("%s\n", output.Color("PR mode: will create pull request", output.Cyan))
	}
	cmd.Println()

	// Phase 1: Detection
	cmd.Println(output.SubHeader("Phase 1: Project Detection", 60))
	detection := detectProject(cwd)

	cmd.Printf("  Git repository: %s\n", boolToYesNo(detection["is_git_repo"].(bool)))
	if detection["is_git_repo"].(bool) {
		if branch := detection["git_branch"]; branch != nil {
			cmd.Printf("  Git branch: %s\n", branch)
		}
		cmd.Printf("  Working tree: %s\n", cleanToStatus(detection["git_clean"].(bool)))
	}
	cmd.Printf("  RTMX config: %s\n", foundToStatus(detection["has_rtmx_config"].(bool)))
	cmd.Printf("  RTM database: %s\n", foundToStatus(detection["has_rtm_database"].(bool)))
	cmd.Printf("  Tests directory: %s\n", foundToStatus(detection["has_tests"].(bool)))
	cmd.Printf("  Makefile: %s\n", foundToStatus(detection["has_makefile"].(bool)))

	// Show agent configs
	agentConfigs := detection["agent_configs"].(map[string]map[string]interface{})
	var agentSummary []string
	for name, info := range agentConfigs {
		if info["exists"].(bool) {
			status := "exists"
			if info["has_rtmx"].(bool) {
				status = "configured"
			}
			agentSummary = append(agentSummary, fmt.Sprintf("%s(%s)", name, status))
		}
	}
	if len(agentSummary) > 0 {
		cmd.Printf("  Agent configs: %s\n", strings.Join(agentSummary, ", "))
	} else {
		cmd.Println("  Agent configs: None found")
	}
	cmd.Println()

	// Phase 1.5: Git branch setup (if requested)
	if setupBranch && detection["is_git_repo"].(bool) {
		cmd.Println(output.SubHeader("Phase 1.5: Git Branch Setup", 60))

		// Check for uncommitted changes
		if !detection["git_clean"].(bool) && !setupForce {
			cmd.Printf("  %s Uncommitted changes detected\n", output.Color("[FAIL]", output.Red))
			cmd.Printf("  %s\n", output.Color("Commit or stash changes first, or use --force", output.Dim))
			result.Errors = append(result.Errors, "Uncommitted changes block branch creation")
			return printSetupSummary(cmd, result)
		}

		// Generate branch name
		branchName := fmt.Sprintf("rtmx/setup-%s", time.Now().Format("20060102-150405"))
		result.BranchName = branchName

		// Create rollback point
		out, err := exec.Command("git", "rev-parse", "HEAD").Output()
		if err == nil {
			result.RollbackPoint = strings.TrimSpace(string(out))
			cmd.Printf("  Rollback point: %s\n", result.RollbackPoint[:8])
		}

		// Create and checkout branch
		if !setupDryRun {
			if err := exec.Command("git", "checkout", "-b", branchName).Run(); err != nil {
				cmd.Printf("  %s Could not create branch: %v\n", output.Color("[FAIL]", output.Red), err)
				result.Errors = append(result.Errors, fmt.Sprintf("Branch creation failed: %v", err))
				return printSetupSummary(cmd, result)
			}
			cmd.Printf("  %s Branch: %s\n", output.Color("[CREATE]", output.Green), branchName)
			result.StepsCompleted = append(result.StepsCompleted, "create_branch")
		} else {
			cmd.Printf("  %s Would create branch: %s\n", output.Color("[SKIP]", output.Dim), branchName)
		}
		cmd.Println()
	} else if setupBranch && !detection["is_git_repo"].(bool) {
		cmd.Printf("%s --branch requested but not a git repo, ignoring\n", output.Color("Warning:", output.Yellow))
		result.Warnings = append(result.Warnings, "--branch ignored: not a git repository")
		setupBranch = false
		cmd.Println()
	}

	// Phase 2: Create config
	cmd.Println(output.SubHeader("Phase 2: Configuration", 60))
	configPath := filepath.Join(cwd, "rtmx.yaml")

	if detection["has_rtmx_config"].(bool) && !setupForce {
		cmd.Printf("  %s rtmx.yaml already exists\n", output.Color("[SKIP]", output.Dim))
		result.StepsSkipped = append(result.StepsSkipped, "create_config")
	} else {
		configContent := `# RTMX Configuration
# See https://rtmx.ai for documentation

rtmx:
  database: docs/rtm_database.csv
  requirements_dir: docs/requirements
  schema: core
  pytest:
    marker_prefix: "req"
    register_markers: true
`
		if !setupDryRun {
			if _, err := os.Stat(configPath); err == nil {
				backupPath := backupFile(configPath)
				if backupPath != "" {
					result.FilesBackedUp = append(result.FilesBackedUp, backupPath)
				}
			}
			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to create config: %v", err))
			} else {
				result.FilesCreated = append(result.FilesCreated, configPath)
			}
		}
		cmd.Printf("  %s rtmx.yaml\n", output.Color("[CREATE]", output.Green))
		result.StepsCompleted = append(result.StepsCompleted, "create_config")
	}
	cmd.Println()

	// Phase 3: Create RTM database
	cmd.Println(output.SubHeader("Phase 3: RTM Database", 60))
	docsDir := filepath.Join(cwd, "docs")
	rtmPath := filepath.Join(docsDir, "rtm_database.csv")
	reqDir := filepath.Join(docsDir, "requirements")

	if detection["has_rtm_database"].(bool) && !setupForce {
		cmd.Printf("  %s RTM database already exists\n", output.Color("[SKIP]", output.Dim))
		result.StepsSkipped = append(result.StepsSkipped, "create_rtm")
	} else {
		rtmContent := `req_id,category,subcategory,requirement_text,target_value,test_module,test_function,validation_method,status,priority,phase,notes,effort_weeks,dependencies,blocks,assignee,sprint,started_date,completed_date,requirement_file
REQ-INIT-001,SETUP,RTMX,RTMX integration complete,Fully configured,tests/test_rtmx.py,test_rtmx_configured,Unit Test,MISSING,HIGH,1,Auto-generated by rtmx setup,0.5,,,developer,v0.1,,,docs/requirements/SETUP/REQ-INIT-001.md
`
		if !setupDryRun {
			os.MkdirAll(docsDir, 0755)
			os.MkdirAll(reqDir, 0755)

			if _, err := os.Stat(rtmPath); err == nil {
				backupPath := backupFile(rtmPath)
				if backupPath != "" {
					result.FilesBackedUp = append(result.FilesBackedUp, backupPath)
				}
			}
			if err := os.WriteFile(rtmPath, []byte(rtmContent), 0644); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to create RTM: %v", err))
			} else {
				result.FilesCreated = append(result.FilesCreated, rtmPath)
			}

			// Create sample requirement spec
			specDir := filepath.Join(reqDir, "SETUP")
			os.MkdirAll(specDir, 0755)
			specPath := filepath.Join(specDir, "REQ-INIT-001.md")
			specContent := `# REQ-INIT-001: RTMX Integration Complete

## Description
RTMX has been integrated into this project for requirements traceability.

## Acceptance Criteria
- [ ] rtmx.yaml configuration exists
- [ ] RTM database initialized
- [ ] Agent configs include RTMX guidance
- [ ] ` + "`rtmx status`" + ` runs without errors

## Validation
- **Test**: tests/test_rtmx.py::test_rtmx_configured
- **Method**: Unit Test
`
			if err := os.WriteFile(specPath, []byte(specContent), 0644); err == nil {
				result.FilesCreated = append(result.FilesCreated, specPath)
			}
		}
		cmd.Printf("  %s docs/rtm_database.csv\n", output.Color("[CREATE]", output.Green))
		cmd.Printf("  %s docs/requirements/SETUP/REQ-INIT-001.md\n", output.Color("[CREATE]", output.Green))
		result.StepsCompleted = append(result.StepsCompleted, "create_rtm")
	}
	cmd.Println()

	// Phase 4: Scan tests for markers (if tests exist)
	if detection["has_tests"].(bool) && !setupMinimal {
		cmd.Println(output.SubHeader("Phase 4: Test Marker Scan", 60))

		testDir := filepath.Join(cwd, "tests")
		if _, err := os.Stat(testDir); os.IsNotExist(err) {
			testDir = filepath.Join(cwd, "test")
		}

		var testFiles []string
		filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasPrefix(filepath.Base(path), "test_") && strings.HasSuffix(path, ".py") {
				testFiles = append(testFiles, path)
			}
			return nil
		})

		markersFound := 0
		for _, tf := range testFiles {
			markersFound += countMarkersInFile(tf)
		}

		cmd.Printf("  Scanned %d test files\n", len(testFiles))
		cmd.Printf("  Found %d requirement markers\n", markersFound)
		result.StepsCompleted = append(result.StepsCompleted, "scan_tests")
		cmd.Println()
	}

	// Phase 5: Agent configs
	if !setupMinimal && !setupSkipAgents {
		cmd.Println(output.SubHeader("Phase 5: Agent Configurations", 60))

		rtmxSection := `
## RTMX

This project uses RTMX for requirements traceability.

### Quick Commands
- ` + "`rtmx status`" + ` - Show RTM progress
- ` + "`rtmx backlog`" + ` - View prioritized backlog
- ` + "`rtmx health`" + ` - Run health checks

### When Implementing Requirements
1. Check the RTM: ` + "`rtmx status`" + `
2. Mark tests with ` + "`@pytest.mark.req(\"REQ-XXX-NNN\")`" + `
3. Update status when complete

### RTM Location
- Database: ` + "`docs/rtm_database.csv`" + `
- Specs: ` + "`docs/requirements/`" + `
`

		for name, info := range agentConfigs {
			agentPath := info["path"].(string)

			if info["has_rtmx"].(bool) {
				cmd.Printf("  %s %s already has RTMX section\n", output.Color("[SKIP]", output.Dim), filepath.Base(agentPath))
				result.StepsSkipped = append(result.StepsSkipped, fmt.Sprintf("agent_%s", name))
			} else if info["exists"].(bool) {
				// Append to existing file
				if !setupDryRun {
					backupPath := backupFile(agentPath)
					if backupPath != "" {
						result.FilesBackedUp = append(result.FilesBackedUp, backupPath)
					}
					f, err := os.OpenFile(agentPath, os.O_APPEND|os.O_WRONLY, 0644)
					if err == nil {
						f.WriteString("\n" + rtmxSection)
						f.Close()
						result.FilesModified = append(result.FilesModified, agentPath)
					}
				}
				cmd.Printf("  %s %s\n", output.Color("[UPDATE]", output.Green), filepath.Base(agentPath))
				result.StepsCompleted = append(result.StepsCompleted, fmt.Sprintf("agent_%s", name))
			} else if name == "claude" || name == "cursor" {
				// Create new file for key agents
				var newPath string
				if name == "claude" {
					newPath = filepath.Join(cwd, "CLAUDE.md")
				} else if name == "cursor" {
					newPath = filepath.Join(cwd, ".cursorrules")
				}
				if newPath != "" && !setupDryRun {
					os.MkdirAll(filepath.Dir(newPath), 0755)
					if err := os.WriteFile(newPath, []byte(fmt.Sprintf("# %s\n%s", filepath.Base(newPath), rtmxSection)), 0644); err == nil {
						result.FilesCreated = append(result.FilesCreated, newPath)
					}
				}
				if newPath != "" {
					cmd.Printf("  %s %s\n", output.Color("[CREATE]", output.Green), filepath.Base(newPath))
					result.StepsCompleted = append(result.StepsCompleted, fmt.Sprintf("agent_%s", name))
				}
			}
		}
		cmd.Println()
	}

	// Phase 6: Makefile
	if !setupMinimal && !setupSkipMakefile && detection["has_makefile"].(bool) {
		cmd.Println(output.SubHeader("Phase 6: Makefile Targets", 60))
		makefilePath := filepath.Join(cwd, "Makefile")

		content, err := os.ReadFile(makefilePath)
		if err == nil && strings.Contains(strings.ToLower(string(content)), "rtmx") && strings.Contains(string(content), "rtm:") {
			cmd.Printf("  %s Makefile already has rtmx targets\n", output.Color("[SKIP]", output.Dim))
			result.StepsSkipped = append(result.StepsSkipped, "makefile")
		} else {
			makefileTargets := `
# RTMX targets
.PHONY: rtm backlog health

rtm:
	@rtmx status

backlog:
	@rtmx backlog

health:
	@rtmx health
`
			if !setupDryRun {
				backupPath := backupFile(makefilePath)
				if backupPath != "" {
					result.FilesBackedUp = append(result.FilesBackedUp, backupPath)
				}
				f, err := os.OpenFile(makefilePath, os.O_APPEND|os.O_WRONLY, 0644)
				if err == nil {
					f.WriteString(makefileTargets)
					f.Close()
					result.FilesModified = append(result.FilesModified, makefilePath)
				}
			}
			cmd.Printf("  %s Makefile (added rtm, backlog, health targets)\n", output.Color("[UPDATE]", output.Green))
			result.StepsCompleted = append(result.StepsCompleted, "makefile")
		}
		cmd.Println()
	}

	// Phase 7: Health check
	cmd.Println(output.SubHeader("Phase 7: Validation", 60))

	if !setupDryRun {
		cfg, err := config.LoadFromDir(cwd)
		if err == nil {
			// Run basic health checks
			dbPath := cfg.DatabasePath(cwd)
			if _, err := os.Stat(dbPath); err == nil {
				cmd.Printf("  %s Health: READY\n", output.Color("[PASS]", output.Green))
				result.StepsCompleted = append(result.StepsCompleted, "health_check")
			} else {
				cmd.Printf("  %s Health: Database not found\n", output.Color("[WARN]", output.Yellow))
				result.Warnings = append(result.Warnings, "Health check shows warnings")
			}
		}
	} else {
		cmd.Printf("  %s Health check (dry run)\n", output.Color("[SKIP]", output.Dim))
	}
	cmd.Println()

	// Phase 8: Git commit and PR (if branch mode)
	if setupBranch && !setupDryRun && detection["is_git_repo"].(bool) {
		cmd.Println(output.SubHeader("Phase 8: Git Commit", 60))

		// Stage all changes
		exec.Command("git", "add", "-A").Run()

		// Commit
		commitMsg := "chore: Add rtmx configuration\n\nAdded by rtmx setup command."
		if err := exec.Command("git", "commit", "-m", commitMsg).Run(); err == nil {
			cmd.Printf("  %s Changes committed to %s\n", output.Color("[COMMIT]", output.Green), result.BranchName)
			result.StepsCompleted = append(result.StepsCompleted, "git_commit")

			// Create PR if requested
			if setupPR {
				cmd.Println()
				cmd.Println(output.SubHeader("Phase 9: Create Pull Request", 60))

				// Check if gh is installed
				if _, err := exec.LookPath("gh"); err != nil {
					cmd.Printf("  %s GitHub CLI (gh) not installed\n", output.Color("[SKIP]", output.Yellow))
					result.Warnings = append(result.Warnings, "PR creation skipped: gh CLI not installed")
				} else {
					// Push branch
					exec.Command("git", "push", "-u", "origin", result.BranchName).Run()

					// Create PR
					prTitle := "chore: Add rtmx for requirements traceability"
					prBody := `## Summary
- Added rtmx configuration (` + "`rtmx.yaml`" + `)
- Initialized RTM database (` + "`docs/rtm_database.csv`" + `)
- Updated agent configurations with RTMX guidance

## Validation
Run ` + "`rtmx health`" + ` to verify setup.

---
Generated by ` + "`rtmx setup --pr`"

					out, err := exec.Command("gh", "pr", "create", "--title", prTitle, "--body", prBody).Output()
					if err == nil {
						prURL := strings.TrimSpace(string(out))
						result.PRUrl = prURL
						cmd.Printf("  %s PR: %s\n", output.Color("[CREATE]", output.Green), prURL)
						result.StepsCompleted = append(result.StepsCompleted, "create_pr")
					} else {
						cmd.Printf("  %s Could not create PR\n", output.Color("[WARN]", output.Yellow))
						result.Warnings = append(result.Warnings, "PR creation failed")
					}
				}
			}
		} else {
			cmd.Printf("  %s Git commit failed\n", output.Color("[WARN]", output.Yellow))
			result.Warnings = append(result.Warnings, "Git commit failed")
		}
		cmd.Println()
	}

	return printSetupSummary(cmd, result)
}

func printSetupSummary(cmd *cobra.Command, result *SetupResult) error {
	cmd.Println(output.Header("Setup Complete", 60))

	if len(result.Errors) > 0 {
		result.Success = false
		cmd.Printf("%s\n", output.Color("Setup completed with errors", output.Red))
	} else {
		result.Success = true
		cmd.Printf("%s\n", output.Color("Setup completed successfully!", output.Green))
	}

	cmd.Println()
	cmd.Printf("  Steps completed: %d\n", len(result.StepsCompleted))
	cmd.Printf("  Steps skipped: %d\n", len(result.StepsSkipped))
	cmd.Printf("  Files created: %d\n", len(result.FilesCreated))
	cmd.Printf("  Files modified: %d\n", len(result.FilesModified))
	cmd.Printf("  Backups created: %d\n", len(result.FilesBackedUp))

	if result.BranchName != "" {
		cmd.Printf("  Branch: %s\n", result.BranchName)
	}
	if result.RollbackPoint != "" {
		cmd.Printf("  Rollback: git reset --hard %s\n", result.RollbackPoint[:8])
	}
	if result.PRUrl != "" {
		cmd.Printf("  PR: %s\n", result.PRUrl)
	}

	if len(result.Warnings) > 0 {
		cmd.Println()
		cmd.Printf("%s\n", output.Color("Warnings:", output.Yellow))
		for _, w := range result.Warnings {
			cmd.Printf("  - %s\n", w)
		}
	}

	if len(result.Errors) > 0 {
		cmd.Println()
		cmd.Printf("%s\n", output.Color("Errors:", output.Red))
		for _, e := range result.Errors {
			cmd.Printf("  - %s\n", e)
		}
	}

	cmd.Println()
	cmd.Println("Next steps:")
	if result.PRUrl != "" {
		cmd.Printf("  1. Review PR: %s\n", result.PRUrl)
		cmd.Println("  2. Merge when ready")
	} else if result.BranchName != "" {
		cmd.Printf("  1. Review changes: git diff main..%s\n", result.BranchName)
		cmd.Println("  2. Create PR: gh pr create")
		cmd.Println("  3. Merge when ready")
	} else {
		cmd.Println("  1. Run 'rtmx status' to see your RTM")
		cmd.Println("  2. Add requirements to docs/rtm_database.csv")
		cmd.Println("  3. Mark tests with @pytest.mark.req('REQ-XXX-NNN')")
	}

	if len(result.Errors) > 0 {
		return NewExitError(1, "setup completed with errors")
	}
	return nil
}

func detectProject(path string) map[string]interface{} {
	detection := map[string]interface{}{
		"is_git_repo":      false,
		"git_clean":        true,
		"git_branch":       nil,
		"has_rtmx_config":  false,
		"has_rtm_database": false,
		"has_tests":        false,
		"has_makefile":     false,
		"agent_configs":    map[string]map[string]interface{}{},
	}

	// Check git
	if _, err := os.Stat(filepath.Join(path, ".git")); err == nil {
		detection["is_git_repo"] = true

		// Get branch
		out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
		if err == nil {
			detection["git_branch"] = strings.TrimSpace(string(out))
		}

		// Check if clean
		out, err = exec.Command("git", "status", "--porcelain").Output()
		if err == nil {
			detection["git_clean"] = len(strings.TrimSpace(string(out))) == 0
		}
	}

	// Check config
	if _, err := os.Stat(filepath.Join(path, "rtmx.yaml")); err == nil {
		detection["has_rtmx_config"] = true
	} else if _, err := os.Stat(filepath.Join(path, ".rtmx.yaml")); err == nil {
		detection["has_rtmx_config"] = true
	}

	// Check RTM database
	if _, err := os.Stat(filepath.Join(path, "docs", "rtm_database.csv")); err == nil {
		detection["has_rtm_database"] = true
	}

	// Check tests
	if _, err := os.Stat(filepath.Join(path, "tests")); err == nil {
		detection["has_tests"] = true
	} else if _, err := os.Stat(filepath.Join(path, "test")); err == nil {
		detection["has_tests"] = true
	}

	// Check Makefile
	if _, err := os.Stat(filepath.Join(path, "Makefile")); err == nil {
		detection["has_makefile"] = true
	}

	// Detect agent configs
	agentFiles := map[string]string{
		"claude":   "CLAUDE.md",
		"cursor":   ".cursorrules",
		"copilot":  ".github/copilot-instructions.md",
		"windsurf": ".windsurfrules",
		"aider":    ".aider.conf.yml",
	}

	agentConfigs := make(map[string]map[string]interface{})
	for name, filename := range agentFiles {
		agentPath := filepath.Join(path, filename)
		info := map[string]interface{}{
			"exists":   false,
			"has_rtmx": false,
			"path":     agentPath,
		}

		if _, err := os.Stat(agentPath); err == nil {
			info["exists"] = true
			content, err := os.ReadFile(agentPath)
			if err == nil {
				info["has_rtmx"] = strings.Contains(string(content), "RTMX") || strings.Contains(string(content), "rtmx")
			}
		}

		agentConfigs[name] = info
	}
	detection["agent_configs"] = agentConfigs

	return detection
}

func backupFile(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ""
	}

	timestamp := time.Now().Format("20060102-150405")
	ext := filepath.Ext(path)
	backupPath := strings.TrimSuffix(path, ext) + fmt.Sprintf(".rtmx-backup-%s%s", timestamp, ext)

	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}

	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return ""
	}

	return backupPath
}

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

func cleanToStatus(clean bool) string {
	if clean {
		return "Clean"
	}
	return "Has changes"
}

func foundToStatus(found bool) string {
	if found {
		return "Found"
	}
	return "Not found"
}

// JSON output helper
func (r *SetupResult) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}
