package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/rtmx-ai/rtmx/internal/tui"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch terminal dashboard",
	Long: `Launches a terminal-based dashboard showing requirements status,
backlog, and health metrics in a compact format suitable for terminal
multiplexers and IDE panels.

Examples:
    rtmx tui            # launch terminal dashboard
    rtmx tui --once     # print dashboard once and exit`,
	RunE: runTui,
}

var tuiOnce bool

func init() {
	tuiCmd.Flags().BoolVar(&tuiOnce, "once", false, "print dashboard once and exit (no interactive mode)")
	rootCmd.AddCommand(tuiCmd)
}

func runTui(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
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

	// Interactive mode: launch Bubble Tea app
	// Fall back to --once mode for non-TTY or explicit flag
	if !tuiOnce && isTerminal() {
		app := tui.NewAppModel(db, dbPath)
		p := tea.NewProgram(app, tea.WithAltScreen())
		_, err := p.Run()
		return err
	}

	return renderTuiDashboard(cmd, db)
}

func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func renderTuiDashboard(cmd *cobra.Command, db *database.Database) error {
	reqs := db.All()
	complete, partial, missing := 0, 0, 0
	for _, req := range reqs {
		switch {
		case req.IsComplete():
			complete++
		case req.Status == database.StatusPartial:
			partial++
		default:
			missing++
		}
	}

	total := len(reqs)
	pct := 0.0
	if total > 0 {
		pct = float64(complete) / float64(total) * 100
	}

	width := output.TerminalWidth()
	cmd.Println(output.Header("RTMX Dashboard", width))
	cmd.Println()

	// Status summary
	cmd.Printf("  %s  %.1f%% complete (%d/%d)\n",
		output.Color("Status:", output.Dim),
		pct, complete, total)
	cmd.Printf("  %s  %d complete, %d partial, %d missing\n",
		output.Color("       ", output.Dim),
		complete, partial, missing)
	cmd.Println()

	// Progress bar
	barWidth := width - 10
	if barWidth > 60 {
		barWidth = 60
	}
	if barWidth < 20 {
		barWidth = 20
	}
	filled := int(float64(barWidth) * float64(complete) / float64(total))
	bar := ""
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "#"
		} else {
			bar += "-"
		}
	}
	cmd.Printf("  [%s] %.0f%%\n", output.Color(bar, output.Green), pct)
	cmd.Println()

	// Category breakdown
	categories := make(map[string][3]int) // complete, partial, missing
	catOrder := make([]string, 0)
	for _, req := range reqs {
		cat := req.Category
		counts := categories[cat]
		switch {
		case req.IsComplete():
			counts[0]++
		case req.Status == database.StatusPartial:
			counts[1]++
		default:
			counts[2]++
		}
		if _, exists := categories[cat]; !exists {
			catOrder = append(catOrder, cat)
		}
		categories[cat] = counts
	}

	if len(categories) > 0 {
		cmd.Printf("  %s\n", output.Color("Categories:", output.Dim))
		for _, cat := range catOrder {
			counts := categories[cat]
			catTotal := counts[0] + counts[1] + counts[2]
			catPct := 0.0
			if catTotal > 0 {
				catPct = float64(counts[0]) / float64(catTotal) * 100
			}
			cmd.Printf("    %-12s %3d/%d (%.0f%%)\n", cat, counts[0], catTotal, catPct)
		}
	}

	return nil
}
