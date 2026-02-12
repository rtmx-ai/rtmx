package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/internal/output"
	"github.com/spf13/cobra"
)

var (
	diffFormat string
	diffOutput string
)

var diffCmd = &cobra.Command{
	Use:   "diff BASELINE [CURRENT]",
	Short: "Compare RTM databases before and after changes",
	Long: `Compare baseline with current database.

If CURRENT is not specified, uses the default database path.

Exit codes:
  0  Stable or improved
  1  Regressed or degraded
  2  Breaking changes

Examples:
    rtmx diff backup.csv                    # Compare with backup
    rtmx diff v1.csv v2.csv                 # Compare two versions
    rtmx diff baseline.csv --format json    # JSON output`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runDiff,
}

func init() {
	diffCmd.Flags().StringVar(&diffFormat, "format", "terminal", "output format: terminal, markdown, json")
	diffCmd.Flags().StringVarP(&diffOutput, "output", "o", "", "output file")

	rootCmd.AddCommand(diffCmd)
}

// DiffResult holds the comparison results.
type DiffResult struct {
	Baseline    DiffStats       `json:"baseline"`
	Current     DiffStats       `json:"current"`
	Added       []string        `json:"added"`
	Removed     []string        `json:"removed"`
	Changed     []ChangedReq    `json:"changed"`
	Improved    int             `json:"improved"`
	Regressed   int             `json:"regressed"`
	ExitCode    int             `json:"exit_code"`
	Summary     string          `json:"summary"`
}

type DiffStats struct {
	Total       int     `json:"total"`
	Complete    int     `json:"complete"`
	Partial     int     `json:"partial"`
	Missing     int     `json:"missing"`
	Completion  float64 `json:"completion_percent"`
}

type ChangedReq struct {
	ReqID     string `json:"req_id"`
	Field     string `json:"field"`
	OldValue  string `json:"old_value"`
	NewValue  string `json:"new_value"`
}

func runDiff(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	baselinePath := args[0]

	// Determine current path
	var currentPath string
	if len(args) > 1 {
		currentPath = args[1]
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get working directory: %w", err)
		}
		cfg, err := config.LoadFromDir(cwd)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		currentPath = cfg.DatabasePath(cwd)
	}

	// Load databases
	baselineDB, err := database.Load(baselinePath)
	if err != nil {
		return fmt.Errorf("failed to load baseline: %w", err)
	}

	currentDB, err := database.Load(currentPath)
	if err != nil {
		return fmt.Errorf("failed to load current: %w", err)
	}

	// Compare
	result := compareDatabases(baselineDB, currentDB)

	// Output
	var outputContent string
	switch diffFormat {
	case "json":
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal result: %w", err)
		}
		outputContent = string(data)
	case "markdown":
		outputContent = formatDiffMarkdown(result)
	default:
		return formatDiffTerminal(cmd, result)
	}

	if diffOutput != "" {
		if err := os.WriteFile(diffOutput, []byte(outputContent), 0644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
		cmd.Printf("Written to %s\n", diffOutput)
	} else {
		cmd.Print(outputContent)
	}

	if result.ExitCode != 0 {
		return NewExitError(result.ExitCode, "")
	}
	return nil
}

func compareDatabases(baseline, current *database.Database) *DiffResult {
	result := &DiffResult{
		Added:   make([]string, 0),
		Removed: make([]string, 0),
		Changed: make([]ChangedReq, 0),
	}

	// Baseline stats
	baselineCounts := baseline.StatusCounts()
	result.Baseline = DiffStats{
		Total:      baseline.Len(),
		Complete:   baselineCounts[database.StatusComplete],
		Partial:    baselineCounts[database.StatusPartial],
		Missing:    baselineCounts[database.StatusMissing] + baselineCounts[database.StatusNotStarted],
		Completion: baseline.CompletionPercentage(),
	}

	// Current stats
	currentCounts := current.StatusCounts()
	result.Current = DiffStats{
		Total:      current.Len(),
		Complete:   currentCounts[database.StatusComplete],
		Partial:    currentCounts[database.StatusPartial],
		Missing:    currentCounts[database.StatusMissing] + currentCounts[database.StatusNotStarted],
		Completion: current.CompletionPercentage(),
	}

	// Find added requirements
	for _, req := range current.All() {
		if baseline.Get(req.ReqID) == nil {
			result.Added = append(result.Added, req.ReqID)
		}
	}

	// Find removed requirements
	for _, req := range baseline.All() {
		if current.Get(req.ReqID) == nil {
			result.Removed = append(result.Removed, req.ReqID)
		}
	}

	// Find changed requirements
	for _, currentReq := range current.All() {
		baselineReq := baseline.Get(currentReq.ReqID)
		if baselineReq == nil {
			continue
		}

		// Check status changes
		if currentReq.Status != baselineReq.Status {
			result.Changed = append(result.Changed, ChangedReq{
				ReqID:    currentReq.ReqID,
				Field:    "status",
				OldValue: baselineReq.Status.String(),
				NewValue: currentReq.Status.String(),
			})

			// Track improvements vs regressions
			if currentReq.Status.CompletionPercent() > baselineReq.Status.CompletionPercent() {
				result.Improved++
			} else if currentReq.Status.CompletionPercent() < baselineReq.Status.CompletionPercent() {
				result.Regressed++
			}
		}

		// Check priority changes
		if currentReq.Priority != baselineReq.Priority {
			result.Changed = append(result.Changed, ChangedReq{
				ReqID:    currentReq.ReqID,
				Field:    "priority",
				OldValue: string(baselineReq.Priority),
				NewValue: string(currentReq.Priority),
			})
		}
	}

	// Determine exit code and summary
	if result.Regressed > 0 {
		result.ExitCode = 1
		result.Summary = "REGRESSED"
	} else if len(result.Removed) > 0 {
		result.ExitCode = 2
		result.Summary = "BREAKING"
	} else if result.Improved > 0 || len(result.Added) > 0 {
		result.ExitCode = 0
		result.Summary = "IMPROVED"
	} else {
		result.ExitCode = 0
		result.Summary = "STABLE"
	}

	return result
}

func formatDiffTerminal(cmd *cobra.Command, result *DiffResult) error {
	width := 80
	cmd.Println(output.Header("RTM Database Comparison", width))
	cmd.Println()

	// Stats comparison
	cmd.Println("Statistics:")
	cmd.Printf("  %-20s %10s  →  %-10s\n", "", "Baseline", "Current")
	cmd.Printf("  %-20s %10d  →  %-10d\n", "Total requirements:", result.Baseline.Total, result.Current.Total)
	cmd.Printf("  %-20s %10d  →  %-10d\n", "Complete:", result.Baseline.Complete, result.Current.Complete)
	cmd.Printf("  %-20s %9.1f%%  →  %-.1f%%\n", "Completion:", result.Baseline.Completion, result.Current.Completion)
	cmd.Println()

	// Changes
	if len(result.Added) > 0 {
		cmd.Printf("%s Added (%d):\n", output.Color("+", output.Green), len(result.Added))
		for _, id := range result.Added {
			cmd.Printf("    %s %s\n", output.Color("+", output.Green), id)
		}
		cmd.Println()
	}

	if len(result.Removed) > 0 {
		cmd.Printf("%s Removed (%d):\n", output.Color("-", output.Red), len(result.Removed))
		for _, id := range result.Removed {
			cmd.Printf("    %s %s\n", output.Color("-", output.Red), id)
		}
		cmd.Println()
	}

	if len(result.Changed) > 0 {
		cmd.Printf("%s Changed (%d):\n", output.Color("~", output.Yellow), len(result.Changed))
		for _, c := range result.Changed {
			var arrow string
			if c.Field == "status" {
				arrow = output.Color("→", output.Yellow)
			} else {
				arrow = "→"
			}
			cmd.Printf("    %s %s.%s: %s %s %s\n",
				output.Color("~", output.Yellow), c.ReqID, c.Field, c.OldValue, arrow, c.NewValue)
		}
		cmd.Println()
	}

	// Summary
	cmd.Println(strings.Repeat("-", width))

	var summaryColor string
	switch result.Summary {
	case "IMPROVED":
		summaryColor = output.Green
	case "STABLE":
		summaryColor = output.Cyan
	case "REGRESSED":
		summaryColor = output.Yellow
	case "BREAKING":
		summaryColor = output.Red
	}

	cmd.Printf("Result: %s\n", output.Color(result.Summary, summaryColor))
	cmd.Printf("  %d improved, %d regressed, %d added, %d removed\n",
		result.Improved, result.Regressed, len(result.Added), len(result.Removed))

	if result.ExitCode != 0 {
		return NewExitError(result.ExitCode, "")
	}
	return nil
}

func formatDiffMarkdown(result *DiffResult) string {
	var sb strings.Builder

	sb.WriteString("# RTM Database Comparison\n\n")

	sb.WriteString("## Statistics\n\n")
	sb.WriteString("| Metric | Baseline | Current |\n")
	sb.WriteString("|--------|----------|--------|\n")
	sb.WriteString(fmt.Sprintf("| Total | %d | %d |\n", result.Baseline.Total, result.Current.Total))
	sb.WriteString(fmt.Sprintf("| Complete | %d | %d |\n", result.Baseline.Complete, result.Current.Complete))
	sb.WriteString(fmt.Sprintf("| Completion | %.1f%% | %.1f%% |\n", result.Baseline.Completion, result.Current.Completion))
	sb.WriteString("\n")

	if len(result.Added) > 0 {
		sb.WriteString("## Added\n\n")
		for _, id := range result.Added {
			sb.WriteString(fmt.Sprintf("- %s\n", id))
		}
		sb.WriteString("\n")
	}

	if len(result.Removed) > 0 {
		sb.WriteString("## Removed\n\n")
		for _, id := range result.Removed {
			sb.WriteString(fmt.Sprintf("- %s\n", id))
		}
		sb.WriteString("\n")
	}

	if len(result.Changed) > 0 {
		sb.WriteString("## Changed\n\n")
		sb.WriteString("| Requirement | Field | Old | New |\n")
		sb.WriteString("|-------------|-------|-----|-----|\n")
		for _, c := range result.Changed {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", c.ReqID, c.Field, c.OldValue, c.NewValue))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("## Summary: %s\n\n", result.Summary))
	sb.WriteString(fmt.Sprintf("- %d improved\n", result.Improved))
	sb.WriteString(fmt.Sprintf("- %d regressed\n", result.Regressed))
	sb.WriteString(fmt.Sprintf("- %d added\n", len(result.Added)))
	sb.WriteString(fmt.Sprintf("- %d removed\n", len(result.Removed)))

	return sb.String()
}
