package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/internal/output"
	"github.com/spf13/cobra"
)

var healthJSON bool

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Run health check for CI/CD pipelines",
	Long: `Run a comprehensive health check on the RTM database.

Exit codes:
  0  All checks passed
  1  Warnings present (non-blocking issues)
  2  Errors present (blocking issues)

Use --json for machine-readable output in CI/CD pipelines.`,
	RunE: runHealth,
}

func init() {
	healthCmd.Flags().BoolVar(&healthJSON, "json", false, "output as JSON")
}

// HealthResult represents the result of a health check.
type HealthResult struct {
	Status   string   `json:"status"`
	ExitCode int      `json:"exit_code"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
	Stats    struct {
		Total          int     `json:"total"`
		Complete       int     `json:"complete"`
		Partial        int     `json:"partial"`
		Missing        int     `json:"missing"`
		Completion     float64 `json:"completion_percent"`
		WithTests      int     `json:"with_tests"`
		WithoutTests   int     `json:"without_tests"`
		Blocked        int     `json:"blocked"`
		CycleCount     int     `json:"cycle_count"`
		OrphanedDeps   int     `json:"orphaned_deps"`
		MissingRecip   int     `json:"missing_reciprocity"`
	} `json:"stats"`
}

func runHealth(cmd *cobra.Command, args []string) error {
	// Apply color settings
	if noColor {
		output.DisableColor()
	}

	// Find and load config
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Load database
	dbPath := cfg.DatabasePath(cwd)
	db, err := database.Load(dbPath)
	if err != nil {
		return fmt.Errorf("failed to load database: %w", err)
	}

	// Run health checks
	result := runHealthChecks(db)

	// Output
	if healthJSON {
		return outputHealthJSON(cmd, result)
	}
	return outputHealthText(cmd, result)
}

func runHealthChecks(db *database.Database) *HealthResult {
	result := &HealthResult{
		Errors:   make([]string, 0),
		Warnings: make([]string, 0),
	}

	// Basic stats
	counts := db.StatusCounts()
	result.Stats.Total = db.Len()
	result.Stats.Complete = counts[database.StatusComplete]
	result.Stats.Partial = counts[database.StatusPartial]
	result.Stats.Missing = counts[database.StatusMissing] + counts[database.StatusNotStarted]
	result.Stats.Completion = db.CompletionPercentage()

	// Test coverage
	for _, req := range db.All() {
		if req.HasTest() {
			result.Stats.WithTests++
		} else {
			result.Stats.WithoutTests++
		}
	}

	// Check for orphaned dependencies
	for _, req := range db.All() {
		for dep := range req.Dependencies {
			// Skip cross-repo deps
			if len(dep) > 0 && dep[0] != '@' && !db.Exists(dep) {
				result.Stats.OrphanedDeps++
				result.Errors = append(result.Errors,
					fmt.Sprintf("%s depends on non-existent %s", req.ReqID, dep))
			}
		}
	}

	// Check for missing reciprocity
	for _, req := range db.All() {
		for dep := range req.Dependencies {
			if depReq := db.Get(dep); depReq != nil {
				if !depReq.Blocks.Contains(req.ReqID) {
					result.Stats.MissingRecip++
					result.Warnings = append(result.Warnings,
						fmt.Sprintf("%s depends on %s but %s doesn't block %s",
							req.ReqID, dep, dep, req.ReqID))
				}
			}
		}
	}

	// Check for blocked requirements
	for _, req := range db.All() {
		if req.IsBlocked(db) && req.IsIncomplete() {
			result.Stats.Blocked++
		}
	}

	// Check for requirements without tests
	incompleteWithoutTest := 0
	for _, req := range db.Incomplete() {
		if !req.HasTest() {
			incompleteWithoutTest++
		}
	}
	if incompleteWithoutTest > 0 {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("%d incomplete requirements have no tests assigned", incompleteWithoutTest))
	}

	// Low test coverage warning
	if result.Stats.Total > 0 {
		testCoverage := float64(result.Stats.WithTests) / float64(result.Stats.Total) * 100
		if testCoverage < 80 {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Test coverage is %.1f%% (recommend >80%%)", testCoverage))
		}
	}

	// TODO: Add cycle detection when graph package is implemented
	// For now, just set to 0
	result.Stats.CycleCount = 0

	// Determine overall status
	if len(result.Errors) > 0 {
		result.Status = "error"
		result.ExitCode = 2
	} else if len(result.Warnings) > 0 {
		result.Status = "warning"
		result.ExitCode = 1
	} else {
		result.Status = "healthy"
		result.ExitCode = 0
	}

	return result
}

func outputHealthJSON(cmd *cobra.Command, result *HealthResult) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize health result: %w", err)
	}
	cmd.Println(string(data))

	// Set exit code
	if result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}
	return nil
}

func outputHealthText(cmd *cobra.Command, result *HealthResult) error {
	width := 80

	// Header
	cmd.Println(output.Header("RTM Health Check", width))
	cmd.Println()

	// Overall status
	var statusIcon, statusColor string
	switch result.Status {
	case "healthy":
		statusIcon = "✓"
		statusColor = output.Green
	case "warning":
		statusIcon = "⚠"
		statusColor = output.Yellow
	case "error":
		statusIcon = "✗"
		statusColor = output.Red
	}
	cmd.Printf("Status: %s %s\n\n",
		output.Color(statusIcon, statusColor),
		output.Color(result.Status, statusColor))

	// Stats
	cmd.Println(output.SubHeader("Statistics", width))
	cmd.Printf("  Total requirements: %d\n", result.Stats.Total)
	cmd.Printf("  Completion: %s\n", output.FormatPercent(result.Stats.Completion))
	cmd.Printf("    %s Complete: %d\n", output.StatusIcon("COMPLETE"), result.Stats.Complete)
	cmd.Printf("    %s Partial: %d\n", output.StatusIcon("PARTIAL"), result.Stats.Partial)
	cmd.Printf("    %s Missing: %d\n", output.StatusIcon("MISSING"), result.Stats.Missing)
	cmd.Println()
	cmd.Printf("  Test coverage: %d/%d (%.1f%%)\n",
		result.Stats.WithTests, result.Stats.Total,
		float64(result.Stats.WithTests)/float64(max(1, result.Stats.Total))*100)
	cmd.Printf("  Blocked requirements: %d\n", result.Stats.Blocked)
	cmd.Println()

	// Errors
	if len(result.Errors) > 0 {
		cmd.Println(output.Color(output.SubHeader("Errors", width), output.Red))
		for _, err := range result.Errors {
			cmd.Printf("  %s %s\n", output.Color("✗", output.Red), err)
		}
		cmd.Println()
	}

	// Warnings
	if len(result.Warnings) > 0 {
		cmd.Println(output.Color(output.SubHeader("Warnings", width), output.Yellow))
		for _, warn := range result.Warnings {
			cmd.Printf("  %s %s\n", output.Color("⚠", output.Yellow), warn)
		}
		cmd.Println()
	}

	// Footer
	if result.Status == "healthy" {
		cmd.Println(output.Color("All health checks passed!", output.Green))
	}

	// Set exit code
	if result.ExitCode != 0 {
		os.Exit(result.ExitCode)
	}
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
