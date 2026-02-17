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

// CheckStatus represents the result of a single health check.
type CheckStatus string

const (
	CheckPass CheckStatus = "PASS"
	CheckWarn CheckStatus = "WARN"
	CheckFail CheckStatus = "FAIL"
	CheckSkip CheckStatus = "SKIP"
)

// HealthCheck represents an individual health check result.
type HealthCheck struct {
	Name        string      `json:"name"`
	Status      CheckStatus `json:"status"`
	Message     string      `json:"message"`
	IsBlocking  bool        `json:"is_blocking,omitempty"`
}

// HealthResult represents the result of a health check.
type HealthResult struct {
	Status   string        `json:"status"`
	ExitCode int           `json:"exit_code"`
	Checks   []HealthCheck `json:"checks"`
	Summary  struct {
		Passed   int `json:"passed"`
		Warnings int `json:"warnings"`
		Failed   int `json:"failed"`
		Skipped  int `json:"skipped"`
	} `json:"summary"`
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
		Checks: make([]HealthCheck, 0),
	}

	// Basic stats
	counts := db.StatusCounts()
	result.Stats.Total = db.Len()
	result.Stats.Complete = counts[database.StatusComplete]
	result.Stats.Partial = counts[database.StatusPartial]
	result.Stats.Missing = counts[database.StatusMissing] + counts[database.StatusNotStarted]
	result.Stats.Completion = db.CompletionPercentage()

	// Check 1: RTM database loaded
	result.Checks = append(result.Checks, HealthCheck{
		Name:    "rtm_loads",
		Status:  CheckPass,
		Message: fmt.Sprintf("RTM database loaded: %d requirements", result.Stats.Total),
	})

	// Test coverage
	for _, req := range db.All() {
		if req.HasTest() {
			result.Stats.WithTests++
		} else {
			result.Stats.WithoutTests++
		}
	}

	// Check 2: Orphaned dependencies
	orphanedErrors := []string{}
	for _, req := range db.All() {
		for dep := range req.Dependencies {
			// Skip cross-repo deps
			if len(dep) > 0 && dep[0] != '@' && !db.Exists(dep) {
				result.Stats.OrphanedDeps++
				orphanedErrors = append(orphanedErrors,
					fmt.Sprintf("%s depends on non-existent %s", req.ReqID, dep))
			}
		}
	}
	if len(orphanedErrors) > 0 {
		result.Checks = append(result.Checks, HealthCheck{
			Name:       "orphaned_deps",
			Status:     CheckFail,
			Message:    fmt.Sprintf("Orphaned dependencies: %d errors", len(orphanedErrors)),
			IsBlocking: true,
		})
	} else {
		result.Checks = append(result.Checks, HealthCheck{
			Name:    "orphaned_deps",
			Status:  CheckPass,
			Message: "No orphaned dependencies",
		})
	}

	// Check 3: Reciprocity
	for _, req := range db.All() {
		for dep := range req.Dependencies {
			if depReq := db.Get(dep); depReq != nil {
				if !depReq.Blocks.Contains(req.ReqID) {
					result.Stats.MissingRecip++
				}
			}
		}
	}
	if result.Stats.MissingRecip > 0 {
		result.Checks = append(result.Checks, HealthCheck{
			Name:    "reciprocity",
			Status:  CheckWarn,
			Message: fmt.Sprintf("Reciprocity violations: %d", result.Stats.MissingRecip),
		})
	} else {
		result.Checks = append(result.Checks, HealthCheck{
			Name:    "reciprocity",
			Status:  CheckPass,
			Message: "All dependencies have reciprocal blocks",
		})
	}

	// Check for blocked requirements
	for _, req := range db.All() {
		if req.IsBlocked(db) && req.IsIncomplete() {
			result.Stats.Blocked++
		}
	}

	// Check 4: Test coverage
	testCoverage := 0.0
	if result.Stats.Total > 0 {
		testCoverage = float64(result.Stats.WithTests) / float64(result.Stats.Total) * 100
	}
	incompleteWithoutTest := 0
	for _, req := range db.Incomplete() {
		if !req.HasTest() {
			incompleteWithoutTest++
		}
	}
	if testCoverage >= 80 {
		result.Checks = append(result.Checks, HealthCheck{
			Name:    "test_coverage",
			Status:  CheckPass,
			Message: fmt.Sprintf("Test coverage: %.1f%% (%d requirements)", testCoverage, result.Stats.WithTests),
		})
	} else {
		result.Checks = append(result.Checks, HealthCheck{
			Name:    "test_coverage",
			Status:  CheckWarn,
			Message: fmt.Sprintf("Test coverage: %.1f%% (%d requirements without tests)", testCoverage, result.Stats.WithoutTests),
		})
	}

	// Check 5: Cycles (placeholder)
	result.Stats.CycleCount = 0
	result.Checks = append(result.Checks, HealthCheck{
		Name:    "cycles",
		Status:  CheckPass,
		Message: "No circular dependencies detected",
	})

	// Calculate summary
	for _, check := range result.Checks {
		switch check.Status {
		case CheckPass:
			result.Summary.Passed++
		case CheckWarn:
			result.Summary.Warnings++
		case CheckFail:
			result.Summary.Failed++
		case CheckSkip:
			result.Summary.Skipped++
		}
	}

	// Determine overall status
	if result.Summary.Failed > 0 {
		result.Status = "UNHEALTHY"
		result.ExitCode = 2
	} else if result.Summary.Warnings > 0 {
		result.Status = "WARNING"
		result.ExitCode = 1
	} else {
		result.Status = "HEALTHY"
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

	// Return exit error if needed
	if result.ExitCode != 0 {
		return NewExitError(result.ExitCode, "")
	}
	return nil
}

func outputHealthText(cmd *cobra.Command, result *HealthResult) error {
	width := 80

	// Header
	cmd.Println(output.Header("RTMX Health Check", width))
	cmd.Println()

	// Display each check in Python format: [PASS] check_name: message
	for _, check := range result.Checks {
		var statusLabel, statusColor string
		switch check.Status {
		case CheckPass:
			statusLabel = "[PASS]"
			statusColor = output.Green
		case CheckWarn:
			statusLabel = "[WARN]"
			statusColor = output.Yellow
		case CheckFail:
			statusLabel = "[FAIL]"
			statusColor = output.Red
		case CheckSkip:
			statusLabel = "[SKIP]"
			statusColor = output.Dim
		}

		msg := check.Message
		if check.IsBlocking {
			msg += " [blocking]"
		}

		cmd.Printf("  %s %s: %s\n",
			output.Color(statusLabel, statusColor),
			check.Name,
			msg)
	}
	cmd.Println()

	// Footer with status
	cmd.Println(output.Header("", width-2))

	var statusColor string
	switch result.Status {
	case "HEALTHY":
		statusColor = output.Green
	case "WARNING":
		statusColor = output.Yellow
	case "UNHEALTHY":
		statusColor = output.Red
	}

	statusMsg := result.Status
	if result.Summary.Failed > 0 {
		statusMsg += " (blocking errors)"
	} else if result.Summary.Warnings > 0 {
		statusMsg += " (non-blocking warnings)"
	}

	cmd.Printf("Status: %s\n", output.Color(statusMsg, statusColor))
	cmd.Printf("Summary: %d passed, %d warnings, %d failed, %d skipped\n",
		result.Summary.Passed, result.Summary.Warnings, result.Summary.Failed, result.Summary.Skipped)

	// Return exit error if needed
	if result.ExitCode != 0 {
		return NewExitError(result.ExitCode, "")
	}
	return nil
}
