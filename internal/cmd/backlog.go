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
	backlogView     string
	backlogPhase    int
	backlogCategory string
	backlogLimit    int
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
		reqs = filterQuickWins(reqs)
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

	// Display
	return displayBacklog(cmd, reqs, db, cfg)
}

func filterCritical(reqs []*database.Requirement, db *database.Database) []*database.Requirement {
	var critical []*database.Requirement
	for _, r := range reqs {
		// P0 or HIGH priority
		if r.Priority == database.PriorityP0 || r.Priority == database.PriorityHigh {
			critical = append(critical, r)
			continue
		}
		// Or blocking many others
		blocked := countBlocked(r, db)
		if blocked >= 2 {
			critical = append(critical, r)
		}
	}
	sortByPriority(critical)
	return critical
}

func filterQuickWins(reqs []*database.Requirement) []*database.Requirement {
	var quickWins []*database.Requirement
	for _, r := range reqs {
		// Low effort AND high priority
		if r.EffortWeeks > 0 && r.EffortWeeks <= 1.0 &&
			(r.Priority == database.PriorityP0 || r.Priority == database.PriorityHigh) {
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
	type blockerInfo struct {
		req     *database.Requirement
		blocked int
	}
	var blockers []blockerInfo
	for _, r := range reqs {
		blocked := countBlocked(r, db)
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
	title := fmt.Sprintf("Backlog (%s)", backlogView)
	if backlogPhase > 0 {
		title += fmt.Sprintf(" - Phase %d", backlogPhase)
	}
	if backlogCategory != "" {
		title += fmt.Sprintf(" - %s", backlogCategory)
	}
	cmd.Println(output.Header(title, width))
	cmd.Println()

	if len(reqs) == 0 {
		cmd.Println("No items in backlog matching criteria.")
		return nil
	}

	cmd.Printf("Showing %d items\n\n", len(reqs))

	// Display based on view
	if backlogView == "list" {
		return displaySimpleList(cmd, reqs)
	}

	return displayDetailedBacklog(cmd, reqs, db, cfg)
}

func displaySimpleList(cmd *cobra.Command, reqs []*database.Requirement) error {
	for _, r := range reqs {
		icon := output.StatusIcon(r.Status.String())
		cmd.Printf("%s %s %s\n", icon, r.ReqID, output.Truncate(r.RequirementText, 50))
	}
	return nil
}

func displayDetailedBacklog(cmd *cobra.Command, reqs []*database.Requirement, db *database.Database, cfg *config.Config) error {
	for i, r := range reqs {
		if i > 0 {
			cmd.Println()
		}

		// Header line with ID and priority
		icon := output.StatusIcon(r.Status.String())
		priorityColor := output.PriorityColor(r.Priority.String())

		cmd.Printf("%s %s [%s] Phase %d\n",
			icon,
			output.Color(r.ReqID, output.Cyan),
			output.Color(string(r.Priority), priorityColor),
			r.Phase)

		// Requirement text
		cmd.Printf("   %s\n", output.Truncate(r.RequirementText, 70))

		// Blocking info
		blocked := countBlocked(r, db)
		if blocked > 0 {
			cmd.Printf("   %s Blocks %d other requirement(s)\n",
				output.Color("→", output.Yellow), blocked)
		}

		// Blocked by info
		if len(r.Dependencies) > 0 {
			blockingDeps := r.BlockingDeps(db)
			if len(blockingDeps) > 0 {
				cmd.Printf("   %s Blocked by: %s\n",
					output.Color("←", output.Red),
					strings.Join(blockingDeps, ", "))
			}
		}

		// Effort
		if r.EffortWeeks > 0 {
			cmd.Printf("   Effort: %.1f weeks\n", r.EffortWeeks)
		}
	}

	return nil
}
