package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/internal/output"
	"github.com/spf13/cobra"
)

var (
	contextFormat string
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Output token-efficient RTM summary for LLM context injection",
	Long: `Output a concise RTM status summary suitable for LLM context injection.

The output is designed to be under 500 tokens and includes:
- Overall completion percentage
- Top 3 blockers (high-priority incomplete requirements blocking others)
- Top 3 quick wins (actionable, high priority, low effort, unblocked)

Formats:
  plain  - Plain text output (default)
  claude - Formatted for Claude Code hooks

Examples:
  rtmx context                    # Token-efficient RTM summary
  rtmx context --format claude    # Claude Code hook format
  rtmx context --format plain     # Plain text`,
	RunE: runContext,
}

func init() {
	contextCmd.Flags().StringVar(&contextFormat, "format", "plain", "output format: plain, claude")

	rootCmd.AddCommand(contextCmd)
}

func runContext(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	// Validate format
	switch contextFormat {
	case "plain", "claude":
		// valid
	default:
		return fmt.Errorf("invalid format %q: must be 'plain' or 'claude'", contextFormat)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	cfg, err := config.LoadFromDir(cwd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	dbPath := cfg.DatabasePath(cwd)
	db, err := database.Load(dbPath)
	if err != nil {
		return fmt.Errorf("failed to load database: %w", err)
	}

	switch contextFormat {
	case "claude":
		return renderClaudeContext(cmd, db)
	default:
		return renderPlainContext(cmd, db)
	}
}

// contextBlocker represents a high-impact blocker requirement.
type contextBlocker struct {
	ReqID    string
	Text     string
	Blocks   int // number of requirements this blocks
	Priority database.Priority
}

// contextQuickWin represents an actionable quick win.
type contextQuickWin struct {
	ReqID    string
	Text     string
	Priority database.Priority
	Effort   float64
}

// computeContextData gathers the summary data from the database.
func computeContextData(db *database.Database) (pct float64, blockers []contextBlocker, quickWins []contextQuickWin) {
	pct = db.CompletionPercentage()

	// Find blockers: incomplete requirements that block others.
	// Count how many incomplete reqs each incomplete req blocks.
	blockCount := make(map[string]int)
	for _, req := range db.All() {
		if req.IsComplete() {
			continue
		}
		for dep := range req.Dependencies {
			if strings.Contains(dep, ":") {
				continue
			}
			depReq := db.Get(dep)
			if depReq != nil && depReq.IsIncomplete() {
				blockCount[dep]++
			}
		}
	}

	// Build blocker list from incomplete reqs that block at least one other
	for reqID, count := range blockCount {
		req := db.Get(reqID)
		if req == nil {
			continue
		}
		blockers = append(blockers, contextBlocker{
			ReqID:    reqID,
			Text:     truncate(req.RequirementText, 60),
			Blocks:   count,
			Priority: req.Priority,
		})
	}
	// Sort by block count descending, then priority
	sort.Slice(blockers, func(i, j int) bool {
		if blockers[i].Blocks != blockers[j].Blocks {
			return blockers[i].Blocks > blockers[j].Blocks
		}
		return blockers[i].Priority.Weight() < blockers[j].Priority.Weight()
	})
	if len(blockers) > 3 {
		blockers = blockers[:3]
	}

	// Find quick wins: incomplete, unblocked, high priority, low effort
	for _, req := range db.All() {
		if req.IsComplete() {
			continue
		}
		if req.IsBlocked(db) {
			continue
		}
		quickWins = append(quickWins, contextQuickWin{
			ReqID:    req.ReqID,
			Text:     truncate(req.RequirementText, 60),
			Priority: req.Priority,
			Effort:   req.EffortWeeks,
		})
	}
	// Sort by priority (ascending weight = higher priority first), then effort ascending
	sort.Slice(quickWins, func(i, j int) bool {
		wi := quickWins[i].Priority.Weight()
		wj := quickWins[j].Priority.Weight()
		if wi != wj {
			return wi < wj
		}
		return quickWins[i].Effort < quickWins[j].Effort
	})
	if len(quickWins) > 3 {
		quickWins = quickWins[:3]
	}

	return pct, blockers, quickWins
}

func renderPlainContext(cmd *cobra.Command, db *database.Database) error {
	pct, blockers, quickWins := computeContextData(db)

	total := db.Len()
	complete := len(db.Complete())

	cmd.Printf("RTM Status: %.0f%% (%d/%d)\n", pct, complete, total)

	if len(blockers) > 0 {
		cmd.Println("\nTop Blockers:")
		for _, b := range blockers {
			cmd.Printf("  %s [%s] blocks %d - %s\n", b.ReqID, b.Priority, b.Blocks, b.Text)
		}
	}

	if len(quickWins) > 0 {
		cmd.Println("\nQuick Wins:")
		for _, qw := range quickWins {
			cmd.Printf("  %s [%s] - %s\n", qw.ReqID, qw.Priority, qw.Text)
		}
	}

	return nil
}

func renderClaudeContext(cmd *cobra.Command, db *database.Database) error {
	pct, blockers, quickWins := computeContextData(db)

	total := db.Len()
	complete := len(db.Complete())
	incomplete := total - complete

	var sb strings.Builder
	fmt.Fprintf(&sb, "RTM: %.0f%% (%d/%d complete, %d remaining)\n", pct, complete, total, incomplete)

	// Status breakdown by category
	byCat := db.ByCategory()
	if len(byCat) > 0 {
		sb.WriteString("Categories: ")
		cats := db.Categories()
		parts := make([]string, 0, len(cats))
		for _, cat := range cats {
			reqs := byCat[cat]
			catComplete := 0
			for _, r := range reqs {
				if r.IsComplete() {
					catComplete++
				}
			}
			parts = append(parts, fmt.Sprintf("%s %d/%d", cat, catComplete, len(reqs)))
		}
		sb.WriteString(strings.Join(parts, ", "))
		sb.WriteString("\n")
	}

	if len(blockers) > 0 {
		sb.WriteString("Blockers: ")
		parts := make([]string, 0, len(blockers))
		for _, b := range blockers {
			parts = append(parts, fmt.Sprintf("%s(blocks %d)", b.ReqID, b.Blocks))
		}
		sb.WriteString(strings.Join(parts, ", "))
		sb.WriteString("\n")
	}

	if len(quickWins) > 0 {
		sb.WriteString("Quick Wins: ")
		parts := make([]string, 0, len(quickWins))
		for _, qw := range quickWins {
			parts = append(parts, fmt.Sprintf("%s[%s]", qw.ReqID, qw.Priority))
		}
		sb.WriteString(strings.Join(parts, ", "))
		sb.WriteString("\n")
	}

	cmd.Print(sb.String())
	return nil
}

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
