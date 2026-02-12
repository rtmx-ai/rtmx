package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/rtmx-ai/rtmx-go/internal/config"
	"github.com/rtmx-ai/rtmx-go/internal/database"
	"github.com/rtmx-ai/rtmx-go/internal/output"
	"github.com/spf13/cobra"
)

var statusVerbosity int

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show RTM completion status",
	Long: `Display the current status of the Requirements Traceability Matrix.

The verbosity level controls how much detail is shown:
  (default)  Summary statistics only
  -v         Show status by category
  -vv        Show status by category and phase
  -vvv       Show individual requirement details`,
	RunE: runStatus,
}

func init() {
	statusCmd.Flags().CountVarP(&statusVerbosity, "verbose", "v", "increase verbosity (-v, -vv, -vvv)")
}

func runStatus(cmd *cobra.Command, args []string) error {
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

	// Display status based on verbosity
	switch {
	case statusVerbosity >= 3:
		return displayDetailedStatus(cmd, db, cfg)
	case statusVerbosity >= 2:
		return displayPhaseStatus(cmd, db, cfg)
	case statusVerbosity >= 1:
		return displayCategoryStatus(cmd, db, cfg)
	default:
		return displaySummaryStatus(cmd, db, cfg)
	}
}

func displaySummaryStatus(cmd *cobra.Command, db *database.Database, cfg *config.Config) error {
	width := 80

	// Header
	cmd.Println(output.Header("RTM Status Check", width))
	cmd.Println()

	// Progress bar
	pct := db.CompletionPercentage()
	cmd.Printf("Requirements: %s  %s\n", output.ProgressBar(pct, 50), output.FormatPercent(pct))
	cmd.Println()

	// Status counts
	counts := db.StatusCounts()
	complete := counts[database.StatusComplete]
	partial := counts[database.StatusPartial]
	missing := counts[database.StatusMissing] + counts[database.StatusNotStarted]

	cmd.Printf("%s %d complete  %s %d partial  %s %d missing\n",
		output.StatusIcon("COMPLETE"), complete,
		output.StatusIcon("PARTIAL"), partial,
		output.StatusIcon("MISSING"), missing)
	cmd.Printf("(%d total)\n", db.Len())
	cmd.Println()

	// Phase status summary
	phases := db.Phases()
	if len(phases) > 0 {
		cmd.Println(output.Header("Phase Status", width))
		cmd.Println()

		byPhase := db.ByPhase()
		for _, phase := range phases {
			reqs := byPhase[phase]
			phasePct := phaseCompletion(reqs)

			var statusText string
			switch {
			case phasePct >= 100:
				statusText = output.Color("✓ Complete", output.Green)
			case phasePct > 0:
				statusText = output.Color("⚠ In Progress", output.Yellow)
			default:
				statusText = output.Color("✗ Not Started", output.Red)
			}

			// Count by status
			complete := 0
			partial := 0
			missing := 0
			for _, r := range reqs {
				switch r.Status {
				case database.StatusComplete:
					complete++
				case database.StatusPartial:
					partial++
				default:
					missing++
				}
			}

			phaseDesc := cfg.PhaseDescription(phase)
			if phaseDesc != "" && phaseDesc != fmt.Sprintf("Phase %d", phase) {
				cmd.Printf("Phase %d (%s): %6.1f%%  %s  (%d%s %d%s %d%s)\n",
					phase, phaseDesc, phasePct, statusText,
					complete, output.Color("✓", output.Green),
					partial, output.Color("⚠", output.Yellow),
					missing, output.Color("✗", output.Red))
			} else {
				cmd.Printf("Phase %d: %6.1f%%  %s  (%d%s %d%s %d%s)\n",
					phase, phasePct, statusText,
					complete, output.Color("✓", output.Green),
					partial, output.Color("⚠", output.Yellow),
					missing, output.Color("✗", output.Red))
			}
		}
		cmd.Println()
	}

	// Footer
	cmd.Println(output.Header(fmt.Sprintf("%d complete, %d partial, %d missing (%.1f%%)",
		complete, partial, missing, pct), width))

	return nil
}

func displayCategoryStatus(cmd *cobra.Command, db *database.Database, cfg *config.Config) error {
	width := 80

	cmd.Println(output.Header("RTM Status Check", width))
	cmd.Println()

	// Progress bar
	pct := db.CompletionPercentage()
	cmd.Printf("Requirements: %s  %s\n", output.ProgressBar(pct, 50), output.FormatPercent(pct))
	cmd.Println()

	// Status counts
	counts := db.StatusCounts()
	totalComplete := counts[database.StatusComplete]
	totalPartial := counts[database.StatusPartial]
	totalMissing := counts[database.StatusMissing] + counts[database.StatusNotStarted]

	cmd.Printf("%s %d complete  %s %d partial  %s %d missing\n",
		output.StatusIcon("COMPLETE"), totalComplete,
		output.StatusIcon("PARTIAL"), totalPartial,
		output.StatusIcon("MISSING"), totalMissing)
	cmd.Printf("(%d total)\n", db.Len())
	cmd.Println()

	// Category breakdown - Python style list
	cmd.Println("Requirements by Category:")
	cmd.Println()

	categories := db.Categories()
	byCategory := db.ByCategory()

	for _, cat := range categories {
		reqs := byCategory[cat]
		catPct := phaseCompletion(reqs)

		// Count by status
		complete := 0
		partial := 0
		missing := 0
		for _, r := range reqs {
			switch r.Status {
			case database.StatusComplete:
				complete++
			case database.StatusPartial:
				partial++
			default:
				missing++
			}
		}

		// Status icon based on completion
		var icon string
		switch {
		case catPct >= 100:
			icon = output.Color("✓", output.Green)
		case catPct >= 50:
			icon = output.Color("⚠", output.Yellow)
		default:
			icon = output.Color("✗", output.Red)
		}

		// Python-style format: "  ✓ CATEGORY        100.0%   N complete   N partial   N missing"
		cmd.Printf("  %s %s %6.1f%%   %d complete   %d partial   %d missing\n",
			icon,
			output.PadRight(cat, 16),
			catPct,
			complete, partial, missing)
	}

	return nil
}

func displayPhaseStatus(cmd *cobra.Command, db *database.Database, cfg *config.Config) error {
	width := 80

	cmd.Println(output.Header("RTM Status by Phase and Category", width))
	cmd.Println()

	phases := db.Phases()
	byPhase := db.ByPhase()

	for _, phase := range phases {
		phaseReqs := byPhase[phase]
		phasePct := phaseCompletion(phaseReqs)
		phaseDesc := cfg.PhaseDescription(phase)

		cmd.Printf("\n%s Phase %d: %s %s\n",
			output.Color("▶", output.Cyan),
			phase,
			phaseDesc,
			output.FormatPercent(phasePct))

		// Group by category within phase
		catMap := make(map[string][]*database.Requirement)
		for _, r := range phaseReqs {
			catMap[r.Category] = append(catMap[r.Category], r)
		}

		// Sort categories
		var cats []string
		for cat := range catMap {
			cats = append(cats, cat)
		}
		sort.Strings(cats)

		for _, cat := range cats {
			reqs := catMap[cat]
			catPct := phaseCompletion(reqs)

			cmd.Printf("  %s: %s %s (%d reqs)\n",
				output.PadRight(cat, 12),
				output.ProgressBar(catPct, 20),
				output.FormatPercent(catPct),
				len(reqs))
		}
	}

	return nil
}

func displayDetailedStatus(cmd *cobra.Command, db *database.Database, cfg *config.Config) error {
	width := 80

	cmd.Println(output.Header("RTM Detailed Status", width))
	cmd.Println()

	// Overall summary first
	pct := db.CompletionPercentage()
	cmd.Printf("Overall: %s  %s (%d requirements)\n",
		output.ProgressBar(pct, 40), output.FormatPercent(pct), db.Len())
	cmd.Println()

	// Group by phase, then category
	phases := db.Phases()
	byPhase := db.ByPhase()

	for _, phase := range phases {
		phaseReqs := byPhase[phase]
		phaseDesc := cfg.PhaseDescription(phase)
		phasePct := phaseCompletion(phaseReqs)

		cmd.Println(output.SubHeader(fmt.Sprintf("Phase %d: %s (%.1f%%)", phase, phaseDesc, phasePct), width))

		// Sort by category, then by ID
		sort.Slice(phaseReqs, func(i, j int) bool {
			if phaseReqs[i].Category != phaseReqs[j].Category {
				return phaseReqs[i].Category < phaseReqs[j].Category
			}
			return phaseReqs[i].ReqID < phaseReqs[j].ReqID
		})

		for _, req := range phaseReqs {
			icon := output.StatusIcon(req.Status.String())
			priorityColor := output.PriorityColor(req.Priority.String())

			// Truncate requirement text
			text := output.Truncate(req.RequirementText, 40)

			cmd.Printf("  %s %s [%s] %s\n",
				icon,
				output.Color(output.PadRight(req.ReqID, 15), output.Cyan),
				output.Color(string(req.Priority), priorityColor),
				text)
		}
		cmd.Println()
	}

	return nil
}

// phaseCompletion calculates completion percentage for a set of requirements.
func phaseCompletion(reqs []*database.Requirement) float64 {
	if len(reqs) == 0 {
		return 0
	}

	var total float64
	for _, r := range reqs {
		total += r.Status.CompletionPercent()
	}

	return total / float64(len(reqs))
}
