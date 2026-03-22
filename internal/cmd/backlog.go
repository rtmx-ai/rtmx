package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/internal/graph"
	"github.com/rtmx-ai/rtmx-go/internal/output"
	"github.com/spf13/cobra"
)

var (
	backlogView     string
	backlogPhase    int
	backlogCategory string
	backlogLimit    int
	backlogJSON     bool
)

var backlogCmd = &cobra.Command{
	Use:   "backlog",
	Short: "Show prioritized backlog",
	Long: `Display the requirements backlog with various view modes.

View modes:
  all         All incomplete requirements (default)
  critical    High priority and blocking requirements
  quick-wins  Low effort, high value requirements
  blockers    Requirements blocking others
  list        Simple list format`,
	RunE: runBacklog,
}

func init() {
	backlogCmd.Flags().StringVar(&backlogView, "view", "all", "view mode: all, critical, quick-wins, blockers, list")
	backlogCmd.Flags().IntVar(&backlogPhase, "phase", 0, "filter by phase number")
	backlogCmd.Flags().StringVar(&backlogCategory, "category", "", "filter by category")
	backlogCmd.Flags().IntVarP(&backlogLimit, "limit", "n", 0, "limit number of results")
	backlogCmd.Flags().BoolVar(&backlogJSON, "json", false, "output as JSON")
}

func runBacklog(cmd *cobra.Command, args []string) error {
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

	// Get incomplete requirements
	reqs := db.Incomplete()

	// Apply filters
	if backlogPhase > 0 {
		var filtered []*database.Requirement
		for _, r := range reqs {
			if r.Phase == backlogPhase {
				filtered = append(filtered, r)
			}
		}
		reqs = filtered
	}

	if backlogCategory != "" {
		var filtered []*database.Requirement
		for _, r := range reqs {
			if strings.EqualFold(r.Category, backlogCategory) {
				filtered = append(filtered, r)
			}
		}
		reqs = filtered
	}

	// Apply view-specific filtering and sorting
	switch backlogView {
	case "critical":
		reqs = filterCritical(reqs, db)
	case "quick-wins":
		reqs = filterQuickWins(reqs, db)
	case "blockers":
		reqs = filterBlockers(reqs, db)
	case "list":
		// Simple list, sorted by ID
		sort.Slice(reqs, func(i, j int) bool {
			return reqs[i].ReqID < reqs[j].ReqID
		})
	default:
		// "all" - prioritized order
		sortByPriority(reqs)
	}

	// Apply limit
	if backlogLimit > 0 && len(reqs) > backlogLimit {
		reqs = reqs[:backlogLimit]
	}

	// JSON output mode
	if backlogJSON {
		return displayBacklogJSON(cmd, reqs, db)
	}

	// Display
	return displayBacklog(cmd, reqs, db, cfg)
}

func filterCritical(reqs []*database.Requirement, db *database.Database) []*database.Requirement {
	g := graph.NewGraph(db)
	var critical []*database.Requirement
	for _, r := range reqs {
		// P0 or HIGH priority
		if r.Priority == database.PriorityP0 || r.Priority == database.PriorityHigh {
			critical = append(critical, r)
			continue
		}
		// Or transitively blocking many others
		blocked := countBlockedTransitive(r.ReqID, g)
		if blocked >= 2 {
			critical = append(critical, r)
		}
	}
	sortByPriority(critical)
	return critical
}

func filterQuickWins(reqs []*database.Requirement, db *database.Database) []*database.Requirement {
	var quickWins []*database.Requirement
	for _, r := range reqs {
		// Low effort AND high priority AND not blocked by incomplete dependencies
		if r.EffortWeeks > 0 && r.EffortWeeks <= 1.0 &&
			(r.Priority == database.PriorityP0 || r.Priority == database.PriorityHigh) &&
			!r.IsBlocked(db) {
			quickWins = append(quickWins, r)
		}
	}
	// Sort by effort (lowest first), then priority
	sort.Slice(quickWins, func(i, j int) bool {
		if quickWins[i].EffortWeeks != quickWins[j].EffortWeeks {
			return quickWins[i].EffortWeeks < quickWins[j].EffortWeeks
		}
		return quickWins[i].Priority.Weight() < quickWins[j].Priority.Weight()
	})
	return quickWins
}

func filterBlockers(reqs []*database.Requirement, db *database.Database) []*database.Requirement {
	g := graph.NewGraph(db)
	type blockerInfo struct {
		req     *database.Requirement
		blocked int
	}
	var blockers []blockerInfo
	for _, r := range reqs {
		blocked := countBlockedTransitive(r.ReqID, g)
		if blocked > 0 {
			blockers = append(blockers, blockerInfo{r, blocked})
		}
	}
	// Sort by number blocked (descending)
	sort.Slice(blockers, func(i, j int) bool {
		return blockers[i].blocked > blockers[j].blocked
	})
	result := make([]*database.Requirement, len(blockers))
	for i, b := range blockers {
		result[i] = b.req
	}
	return result
}

func countBlocked(req *database.Requirement, db *database.Database) int {
	count := 0
	for _, r := range db.All() {
		if r.Dependencies.Contains(req.ReqID) && r.IsIncomplete() {
			count++
		}
	}
	return count
}

// countBlockedTransitive returns the number of incomplete requirements transitively
// blocked by this requirement, using the graph's transitive dependents analysis.
func countBlockedTransitive(reqID string, g *graph.Graph) int {
	count := 0
	for _, dep := range g.TransitiveDependents(reqID) {
		if g.IsIncomplete(dep) {
			count++
		}
	}
	return count
}

// formatBlocksColumn returns the "X (Y)" format for the Blocks column
// where X is the transitive count and Y is the direct count.
func formatBlocksColumn(transitive, direct int) string {
	return fmt.Sprintf("%d (%d)", transitive, direct)
}

func sortByPriority(reqs []*database.Requirement) {
	sort.Slice(reqs, func(i, j int) bool {
		// Sort by priority weight (lower = higher priority)
		pi := reqs[i].Priority.Weight()
		pj := reqs[j].Priority.Weight()
		if pi != pj {
			return pi < pj
		}
		// Then by phase
		if reqs[i].Phase != reqs[j].Phase {
			return reqs[i].Phase < reqs[j].Phase
		}
		// Then by ID
		return reqs[i].ReqID < reqs[j].ReqID
	})
}

func displayBacklog(cmd *cobra.Command, reqs []*database.Requirement, db *database.Database, cfg *config.Config) error {
	width := 80

	// Header
	cmd.Println(output.Header("Prioritized Backlog", width))
	cmd.Println()

	if len(reqs) == 0 {
		cmd.Println("No items in backlog matching criteria.")
		return nil
	}

	// Summary statistics
	totalMissing := 0
	totalPartial := 0
	totalEffort := 0.0
	for _, r := range reqs {
		switch r.Status {
		case database.StatusMissing, database.StatusNotStarted:
			totalMissing++
		case database.StatusPartial:
			totalPartial++
		}
		totalEffort += r.EffortWeeks
	}

	cmd.Printf("Total Requirements: %d\n", db.Len())
	cmd.Printf("  %s MISSING: %d (%.1f%%)\n",
		output.StatusIcon("MISSING"), totalMissing, float64(totalMissing)/float64(db.Len())*100)
	cmd.Printf("  %s PARTIAL: %d (%.1f%%)\n",
		output.StatusIcon("PARTIAL"), totalPartial, float64(totalPartial)/float64(db.Len())*100)
	cmd.Printf("Estimated Effort: %.1f weeks\n", totalEffort)
	cmd.Println()

	// Display based on view
	switch backlogView {
	case "list":
		return displaySimpleList(cmd, reqs)
	case "critical":
		return displayCriticalTable(cmd, reqs, db, cfg)
	case "quick-wins":
		return displayQuickWinsTable(cmd, reqs, cfg)
	case "blockers":
		return displayBlockersTable(cmd, reqs, db, cfg)
	default:
		return displayAllBacklog(cmd, reqs, db, cfg)
	}
}

func displaySimpleList(cmd *cobra.Command, reqs []*database.Requirement) error {
	for _, r := range reqs {
		icon := output.StatusIcon(r.Status.String())
		cmd.Printf("%s %s %s\n", icon, r.ReqID, output.Truncate(r.RequirementText, 50))
	}
	return nil
}

func displayAllBacklog(cmd *cobra.Command, reqs []*database.Requirement, db *database.Database, cfg *config.Config) error {
	g := graph.NewGraph(db)
	// Split into critical, quick wins, and remaining
	var critical, quickWins, remaining []*database.Requirement
	for _, r := range reqs {
		blocked := countBlockedTransitive(r.ReqID, g)
		if r.Priority == database.PriorityP0 || r.Priority == database.PriorityHigh || blocked >= 2 {
			critical = append(critical, r)
		} else if r.EffortWeeks > 0 && r.EffortWeeks <= 1.0 {
			quickWins = append(quickWins, r)
		} else {
			remaining = append(remaining, r)
		}
	}

	// Display critical path items
	if len(critical) > 0 {
		limit := 5
		if backlogLimit > 0 && backlogLimit < limit {
			limit = backlogLimit
		}
		if len(critical) > limit {
			critical = critical[:limit]
		}
		cmd.Printf("CRITICAL PATH ITEMS (TOP %d)\n\n", len(critical))
		_ = displayCriticalTable(cmd, critical, db, cfg)
		cmd.Println()
	}

	// Display quick wins
	if len(quickWins) > 0 {
		cmd.Println("QUICK WINS (<1 week, HIGH priority)")
		cmd.Println()
		_ = displayQuickWinsTable(cmd, quickWins, cfg)
		cmd.Println()
	}

	// Display remaining
	if len(remaining) > 0 {
		cmd.Println("REMAINING REQUIREMENTS")
		cmd.Println()
		_ = displayRemainingTable(cmd, remaining, db, cfg)
	}

	return nil
}

func displayCriticalTable(cmd *cobra.Command, reqs []*database.Requirement, db *database.Database, cfg *config.Config) error {
	g := graph.NewGraph(db)
	table := output.NewTable("#", "Status", "Requirement", "Description", "Effort", "Blocks", "Phase")

	for i, r := range reqs {
		icon := output.StatusIcon(r.Status.String())
		transitive := countBlockedTransitive(r.ReqID, g)
		direct := countBlocked(r, db)
		blocksStr := formatBlocksColumn(transitive, direct)

		phaseDesc := cfg.PhaseDescription(r.Phase)
		phaseStr := fmt.Sprintf("Phase %d (%s)", r.Phase, phaseDesc)

		effortStr := ""
		if r.EffortWeeks > 0 {
			effortStr = fmt.Sprintf("%.1fw", r.EffortWeeks)
		}

		table.AddRow(
			fmt.Sprintf("%d", i+1),
			icon,
			r.ReqID,
			output.TruncateCell(r.RequirementText, 35),
			effortStr,
			blocksStr,
			output.TruncateCell(phaseStr, 30),
		)
	}

	cmd.Print(table.Render())
	return nil
}

func displayQuickWinsTable(cmd *cobra.Command, reqs []*database.Requirement, cfg *config.Config) error {
	table := output.NewTable("#", "Status", "Requirement", "Description", "Effort", "Phase")

	for i, r := range reqs {
		icon := output.StatusIcon(r.Status.String())

		phaseDesc := cfg.PhaseDescription(r.Phase)
		phaseStr := fmt.Sprintf("Phase %d (%s)", r.Phase, phaseDesc)

		effortStr := ""
		if r.EffortWeeks > 0 {
			effortStr = fmt.Sprintf("%.1fw", r.EffortWeeks)
		}

		table.AddRow(
			fmt.Sprintf("%d", i+1),
			icon,
			r.ReqID,
			output.TruncateCell(r.RequirementText, 35),
			effortStr,
			output.TruncateCell(phaseStr, 14),
		)
	}

	cmd.Print(table.Render())
	return nil
}

func displayBlockersTable(cmd *cobra.Command, reqs []*database.Requirement, db *database.Database, cfg *config.Config) error {
	g := graph.NewGraph(db)
	table := output.NewTable("#", "Status", "Requirement", "Description", "Blocks", "Phase")

	for i, r := range reqs {
		icon := output.StatusIcon(r.Status.String())
		transitive := countBlockedTransitive(r.ReqID, g)
		direct := countBlocked(r, db)

		phaseDesc := cfg.PhaseDescription(r.Phase)
		phaseStr := fmt.Sprintf("Phase %d (%s)", r.Phase, phaseDesc)

		table.AddRow(
			fmt.Sprintf("%d", i+1),
			icon,
			r.ReqID,
			output.TruncateCell(r.RequirementText, 35),
			formatBlocksColumn(transitive, direct),
			output.TruncateCell(phaseStr, 20),
		)
	}

	cmd.Print(table.Render())
	return nil
}

func displayRemainingTable(cmd *cobra.Command, reqs []*database.Requirement, db *database.Database, cfg *config.Config) error {
	g := graph.NewGraph(db)
	table := output.NewTable("#", "Status", "Requirement", "Description", "Priority", "Blocks", "⊘", "Phase")

	actionable := 0
	blocked := 0

	for i, r := range reqs {
		icon := output.StatusIcon(r.Status.String())
		transitive := countBlockedTransitive(r.ReqID, g)
		direct := countBlocked(r, db)

		phaseDesc := cfg.PhaseDescription(r.Phase)
		phaseStr := fmt.Sprintf("Phase %d (%s)", r.Phase, phaseDesc)

		// Check if blocked by incomplete dependencies
		blockingDeps := r.BlockingDeps(db)
		blockedMarker := ""
		if len(blockingDeps) > 0 {
			blockedMarker = "⊘"
			blocked++
		} else {
			actionable++
		}

		table.AddRow(
			fmt.Sprintf("%d", i+1),
			icon,
			r.ReqID,
			output.TruncateCell(r.RequirementText, 35),
			string(r.Priority),
			formatBlocksColumn(transitive, direct),
			blockedMarker,
			output.TruncateCell(phaseStr, 26),
		)
	}

	cmd.Print(table.Render())
	cmd.Println()
	cmd.Println("⊘ = blocked by incomplete dependencies")
	cmd.Printf("%d actionable, %d blocked\n", actionable, blocked)

	return nil
}

// backlogJSONCriticalItem represents a critical path item in JSON output.
type backlogJSONCriticalItem struct {
	ReqID       string  `json:"req_id"`
	Description string  `json:"description"`
	Effort      float64 `json:"effort"`
	Blocks      int     `json:"blocks"`
}

// backlogJSONRemainingItem represents a remaining item in JSON output.
type backlogJSONRemainingItem struct {
	ReqID       string `json:"req_id"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Blocked     bool   `json:"blocked"`
}

// backlogJSONOutput represents the full JSON output structure.
type backlogJSONOutput struct {
	TotalMissing        int                        `json:"total_missing"`
	EstimatedEffortWeeks float64                   `json:"estimated_effort_weeks"`
	CriticalPath        []backlogJSONCriticalItem  `json:"critical_path"`
	Remaining           []backlogJSONRemainingItem `json:"remaining"`
}

func displayBacklogJSON(cmd *cobra.Command, reqs []*database.Requirement, db *database.Database) error {
	totalMissing := 0
	totalEffort := 0.0
	for _, r := range reqs {
		if r.Status == database.StatusMissing || r.Status == database.StatusNotStarted {
			totalMissing++
		}
		totalEffort += r.EffortWeeks
	}

	result := backlogJSONOutput{
		TotalMissing:         totalMissing,
		EstimatedEffortWeeks: totalEffort,
		CriticalPath:         make([]backlogJSONCriticalItem, 0),
		Remaining:            make([]backlogJSONRemainingItem, 0),
	}

	// Build critical path and remaining items
	for _, r := range reqs {
		blocked := countBlocked(r, db)
		isBlocked := r.IsBlocked(db)

		if r.Priority == database.PriorityP0 || r.Priority == database.PriorityHigh || blocked >= 2 {
			result.CriticalPath = append(result.CriticalPath, backlogJSONCriticalItem{
				ReqID:       r.ReqID,
				Description: r.RequirementText,
				Effort:      r.EffortWeeks,
				Blocks:      blocked,
			})
		} else {
			result.Remaining = append(result.Remaining, backlogJSONRemainingItem{
				ReqID:       r.ReqID,
				Description: r.RequirementText,
				Priority:    string(r.Priority),
				Blocked:     isBlocked,
			})
		}
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	cmd.Println(string(data))
	return nil
}
