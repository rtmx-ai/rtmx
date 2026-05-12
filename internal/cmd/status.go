package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/spf13/cobra"
)

var (
	statusVerbosity int
	statusJSON      bool
	statusFailUnder float64
	statusVersion   string
	statusByVersion bool
)

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

var (
	statusVerify bool
	statusNoWarn bool
)

func init() {
	statusCmd.Flags().CountVarP(&statusVerbosity, "verbose", "v", "increase verbosity (-v, -vv, -vvv)")
	statusCmd.Flags().BoolVar(&statusJSON, "json", false, "output as JSON")
	statusCmd.Flags().Float64Var(&statusFailUnder, "fail-under", 0, "fail if completion percentage is below threshold")
	statusCmd.Flags().BoolVar(&statusVerify, "verify", false, "run verify --update before displaying status")
	statusCmd.Flags().BoolVar(&statusNoWarn, "no-warn", false, "suppress staleness warning")
	statusCmd.Flags().StringVar(&statusVersion, "version", "", "filter by target version (sprint field)")
	statusCmd.Flags().BoolVar(&statusByVersion, "by-version", false, "show status grouped by target version")
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

	// Run verify first if --verify flag or auto_verify config
	if statusVerify || cfg.RTMX.Verify.AutoVerify {
		if err := runVerify(cmd, nil); err != nil {
			// Verify may fail due to test failures -- continue showing status
			cmd.Printf("%s Verify completed with errors: %v\n\n", output.Color("!", output.Yellow), err)
		}
	}

	// Check staleness (unless --no-warn or --json or --verify just ran)
	if !statusNoWarn && !statusJSON && !statusVerify && !cfg.RTMX.Verify.AutoVerify && cfg.RTMX.Verify.ShouldWarnStale() {
		dbPath := cfg.DatabasePath(cwd)
		rtmxDir := filepath.Dir(dbPath)
		warning := CheckStaleness(rtmxDir)
		if warning != "" {
			cmd.Printf("%s %s\n\n", output.Color("WARNING:", output.Yellow), warning)
		}
	}

	// Load database
	dbPath := cfg.DatabasePath(cwd)
	db, err := database.Load(dbPath)
	if err != nil {
		return fmt.Errorf("failed to load database: %w", err)
	}

	// Apply version filter if specified
	if statusVersion != "" {
		db = db.FilteredCopy(database.FilterOptions{TargetVersion: statusVersion})
		if db.Len() == 0 {
			cmd.Printf("No requirements assigned to version %s\n", statusVersion)
			return nil
		}
		cmd.Printf("Filtered to version: %s (%d requirements)\n\n", statusVersion, db.Len())
	}

	// JSON output mode
	if statusJSON {
		return displayStatusJSON(cmd, db, cfg)
	}

	// Display status based on verbosity
	if statusByVersion {
		err = displayVersionStatus(cmd, db, cfg)
	} else {
		switch {
		case statusVerbosity >= 3:
			err = displayDetailedStatus(cmd, db, cfg)
		case statusVerbosity >= 2:
			err = displayPhaseStatus(cmd, db, cfg)
		case statusVerbosity >= 1:
			err = displayCategoryStatus(cmd, db, cfg)
		default:
			err = displaySummaryStatus(cmd, db, cfg)
		}
	}

	if err != nil {
		return err
	}

	// Check fail-under threshold
	if statusFailUnder > 0 {
		pct := db.CompletionPercentage()
		if pct < statusFailUnder {
			return NewExitError(1, fmt.Sprintf("completion %.1f%% is below threshold %.1f%%", pct, statusFailUnder))
		}
	}

	return nil
}

// statusJSONPhase represents a phase entry in JSON output.
type statusJSONPhase struct {
	Phase    int     `json:"phase"`
	Name     string  `json:"name"`
	Total    int     `json:"total"`
	Complete int     `json:"complete"`
	Pct      float64 `json:"pct"`
}

// statusJSONCategory represents a category entry in JSON output.
type statusJSONCategory struct {
	Name     string  `json:"name"`
	Total    int     `json:"total"`
	Complete int     `json:"complete"`
	Pct      float64 `json:"pct"`
}

// statusJSONOutput represents the full JSON output structure.
type statusJSONOutput struct {
	Total           int                  `json:"total"`
	Complete        int                  `json:"complete"`
	Partial         int                  `json:"partial"`
	Missing         int                  `json:"missing"`
	CompletionPct   float64              `json:"completion_pct"`
	Phases          []statusJSONPhase    `json:"phases"`
	Categories      []statusJSONCategory `json:"categories"`
	FailUnder       *float64             `json:"fail_under,omitempty"`
	ThresholdPassed *bool                `json:"threshold_passed,omitempty"`
}

func displayStatusJSON(cmd *cobra.Command, db *database.Database, cfg *config.Config) error {
	counts := db.StatusCounts()
	complete := counts[database.StatusComplete]
	partial := counts[database.StatusPartial]
	missing := counts[database.StatusMissing] + counts[database.StatusNotStarted]
	pct := db.CompletionPercentage()

	// Round to 1 decimal place
	pct = math.Round(pct*10) / 10

	result := statusJSONOutput{
		Total:         db.Len(),
		Complete:      complete,
		Partial:       partial,
		Missing:       missing,
		CompletionPct: pct,
		Phases:        make([]statusJSONPhase, 0),
		Categories:    make([]statusJSONCategory, 0),
	}

	// Build phases
	phases := db.Phases()
	byPhase := db.ByPhase()
	for _, phase := range phases {
		reqs := byPhase[phase]
		phaseComplete := 0
		for _, r := range reqs {
			if r.Status == database.StatusComplete {
				phaseComplete++
			}
		}
		phasePct := phaseCompletion(reqs)
		phasePct = math.Round(phasePct*10) / 10

		result.Phases = append(result.Phases, statusJSONPhase{
			Phase:    phase,
			Name:     cfg.PhaseDescription(phase),
			Total:    len(reqs),
			Complete: phaseComplete,
			Pct:      phasePct,
		})
	}

	// Build categories
	categories := db.Categories()
	byCategory := db.ByCategory()
	for _, cat := range categories {
		reqs := byCategory[cat]
		catComplete := 0
		for _, r := range reqs {
			if r.Status == database.StatusComplete {
				catComplete++
			}
		}
		catPct := phaseCompletion(reqs)
		catPct = math.Round(catPct*10) / 10

		result.Categories = append(result.Categories, statusJSONCategory{
			Name:     cat,
			Total:    len(reqs),
			Complete: catComplete,
			Pct:      catPct,
		})
	}

	// Include fail-under info if specified
	var failUnderErr error
	if statusFailUnder > 0 {
		result.FailUnder = &statusFailUnder
		passed := pct >= statusFailUnder
		result.ThresholdPassed = &passed
		if !passed {
			failUnderErr = NewExitError(1, fmt.Sprintf("completion %.1f%% is below threshold %.1f%%", pct, statusFailUnder))
		}
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	cmd.Println(string(data))

	return failUnderErr
}

func displaySummaryStatus(cmd *cobra.Command, db *database.Database, cfg *config.Config) error {
	width := output.TerminalWidth()

	// Header
	cmd.Println(output.Header("RTM Status Check", width))
	cmd.Println()

	// Progress bar capped at MaxBarWidth for visual consistency
	barWidth := output.ClampBarWidth(width - 20)
	pct := db.CompletionPercentage()
	cmd.Printf("Requirements: %s  %s\n", output.ProgressBar(pct, barWidth), output.FormatPercent(pct))
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

	// Phase summary: count phases by completion status
	phases := db.Phases()
	if len(phases) > 0 {
		byPhase := db.ByPhase()
		donePhases := 0
		inProgressPhases := 0
		for _, phase := range phases {
			phasePct := phaseCompletion(byPhase[phase])
			if phasePct >= 100 {
				donePhases++
			} else {
				inProgressPhases++
			}
		}
		cmd.Printf("%d phases (%d complete", len(phases), donePhases)
		if inProgressPhases > 0 {
			cmd.Printf(", %d in progress", inProgressPhases)
		}
		cmd.Println(")")
		cmd.Println()
	}

	// Footer
	cmd.Println(output.Header(fmt.Sprintf("%d complete, %d partial, %d missing (%.1f%%)",
		complete, partial, missing, pct), width))

	return nil
}

func displayCategoryStatus(cmd *cobra.Command, db *database.Database, cfg *config.Config) error {
	width := output.TerminalWidth()

	cmd.Println(output.Header("RTM Status Check", width))
	cmd.Println()

	// Progress bar capped at MaxBarWidth
	barWidth := output.ClampBarWidth(width - 20)
	pct := db.CompletionPercentage()
	cmd.Printf("Requirements: %s  %s\n", output.ProgressBar(pct, barWidth), output.FormatPercent(pct))
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
	width := output.TerminalWidth()

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

		// Scale category progress bar to terminal width
		// "  " + cat(12) + ": " + bar + " " + pct(6) + " (" + N + " reqs)"
		catBarWidth := output.ClampBarWidth(width - 40)

		for _, cat := range cats {
			reqs := catMap[cat]
			catPct := phaseCompletion(reqs)

			cmd.Printf("  %s: %s %s (%d reqs)\n",
				output.PadRight(cat, 12),
				output.ProgressBar(catPct, catBarWidth),
				output.FormatPercent(catPct),
				len(reqs))
		}
	}

	return nil
}

func displayDetailedStatus(cmd *cobra.Command, db *database.Database, cfg *config.Config) error {
	width := output.TerminalWidth()

	cmd.Println(output.Header("RTM Detailed Status", width))
	cmd.Println()

	// Overall summary first - scale bar to terminal width
	overallBarWidth := output.ClampBarWidth(width - 35)
	pct := db.CompletionPercentage()
	cmd.Printf("Overall: %s  %s  (%d requirements)\n",
		output.ProgressBar(pct, overallBarWidth), output.FormatPercent(pct), db.Len())
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

			// Build suffix with assignee, version, dates
			var suffix []string
			if v := req.TargetVersion(); v != "" {
				suffix = append(suffix, v)
			}
			if req.Assignee != "" {
				suffix = append(suffix, "@"+req.Assignee)
			}
			if req.StartedDate != "" && req.CompletedDate != "" {
				suffix = append(suffix, req.StartedDate+".."+req.CompletedDate)
			} else if req.StartedDate != "" {
				suffix = append(suffix, "started:"+req.StartedDate)
			}

			line := fmt.Sprintf("  %s %s [%s] %s",
				icon,
				output.Color(output.PadRight(req.ReqID, 15), output.Cyan),
				output.Color(string(req.Priority), priorityColor),
				text)
			if len(suffix) > 0 {
				line += "  " + output.Color(strings.Join(suffix, " "), output.Dim)
			}
			cmd.Println(line)
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

func displayVersionStatus(cmd *cobra.Command, db *database.Database, _ *config.Config) error {
	width := output.TerminalWidth()

	cmd.Println(output.Header("RTM Status by Version", width))
	cmd.Println()

	// Overall summary
	overallBarWidth := output.ClampBarWidth(width - 35)
	pct := db.CompletionPercentage()
	cmd.Printf("Overall: %s  %s  (%d requirements)\n",
		output.ProgressBar(pct, overallBarWidth), output.FormatPercent(pct), db.Len())
	cmd.Println()

	byVersion := db.ByVersion()
	versions := db.Versions()

	barWidth := output.ClampBarWidth(width - 40)

	for _, ver := range versions {
		reqs := byVersion[ver]
		verPct := phaseCompletion(reqs)

		var complete, partial, missing int
		var totalEffort, remainingEffort float64
		for _, r := range reqs {
			switch r.Status {
			case database.StatusComplete:
				complete++
			case database.StatusPartial:
				partial++
			default:
				missing++
			}
			totalEffort += r.EffortWeeks
			if r.Status != database.StatusComplete {
				remainingEffort += r.EffortWeeks
			}
		}

		cmd.Printf("  %s  %s  %s  (%d/%d complete",
			output.Color(output.PadRight(ver, 10), output.Cyan),
			output.ProgressBar(verPct, barWidth),
			output.FormatPercent(verPct),
			complete, len(reqs))
		if remainingEffort > 0 {
			cmd.Printf(", %.1fw remaining", remainingEffort)
		}
		cmd.Println(")")
	}

	// Unversioned
	if unversioned := byVersion[""]; len(unversioned) > 0 {
		uPct := phaseCompletion(unversioned)
		var uComplete int
		for _, r := range unversioned {
			if r.Status == database.StatusComplete {
				uComplete++
			}
		}
		cmd.Printf("  %s  %s  %s  (%d/%d complete)\n",
			output.Color(output.PadRight("(none)", 10), output.Dim),
			output.ProgressBar(uPct, barWidth),
			output.FormatPercent(uPct),
			uComplete, len(unversioned))
	}

	cmd.Println()
	return nil
}
