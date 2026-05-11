package cmd

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/spf13/cobra"
)

var (
	velocityWindow int
	velocityJSON   bool
)

var velocityCmd = &cobra.Command{
	Use:   "velocity",
	Short: "Show team velocity from completed requirements",
	Long: `Compute team velocity from completed requirements that have both
effort_weeks and completed_date populated. Velocity is defined as total
effort-weeks completed divided by calendar-weeks elapsed.

Examples:
    rtmx velocity             # All-time velocity
    rtmx velocity --window 4  # Last 4 calendar weeks
    rtmx velocity --json      # Machine-readable output`,
	RunE: runVelocity,
}

func init() {
	velocityCmd.Flags().IntVar(&velocityWindow, "window", 0, "limit to last N calendar weeks")
	velocityCmd.Flags().BoolVar(&velocityJSON, "json", false, "output as JSON")
	rootCmd.AddCommand(velocityCmd)
}

// VelocityResult holds the computed velocity data.
type VelocityResult struct {
	TotalEffortWeeks   float64 `json:"total_effort_weeks"`
	CalendarWeeks      float64 `json:"calendar_weeks"`
	Velocity           float64 `json:"velocity"`
	CompletedCount     int     `json:"completed_count"`
	WindowWeeks        int     `json:"window_weeks,omitempty"`
	InsufficientData   bool    `json:"insufficient_data,omitempty"`
}

// ComputeVelocity calculates velocity from a set of requirements.
// Exported for use by release forecast.
func ComputeVelocity(reqs []*database.Requirement, windowWeeks int) *VelocityResult {
	now := time.Now()
	var cutoff time.Time
	if windowWeeks > 0 {
		cutoff = now.AddDate(0, 0, -windowWeeks*7)
	}

	var totalEffort float64
	var earliest, latest time.Time
	count := 0

	for _, req := range reqs {
		if req.Status != database.StatusComplete {
			continue
		}
		if req.EffortWeeks <= 0 || req.CompletedDate == "" {
			continue
		}

		completed, err := time.Parse("2006-01-02", req.CompletedDate)
		if err != nil {
			continue
		}

		if windowWeeks > 0 && completed.Before(cutoff) {
			continue
		}

		totalEffort += req.EffortWeeks
		count++

		if earliest.IsZero() || completed.Before(earliest) {
			earliest = completed
		}
		if latest.IsZero() || completed.After(latest) {
			latest = completed
		}
	}

	result := &VelocityResult{
		TotalEffortWeeks: totalEffort,
		CompletedCount:   count,
	}
	if windowWeeks > 0 {
		result.WindowWeeks = windowWeeks
	}

	if count == 0 {
		result.InsufficientData = true
		return result
	}

	// Calendar weeks elapsed: from earliest completion to now
	calendarDays := now.Sub(earliest).Hours() / 24
	result.CalendarWeeks = math.Max(calendarDays/7.0, 1.0) // at least 1 week
	result.Velocity = totalEffort / result.CalendarWeeks

	return result
}

func runVelocity(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	db, _, err := loadDB(cwd)
	if err != nil {
		return err
	}

	result := ComputeVelocity(db.All(), velocityWindow)

	if velocityJSON {
		data, _ := json.MarshalIndent(result, "", "  ")
		cmd.Println(string(data))
		return nil
	}

	if result.InsufficientData {
		cmd.Println("Not enough data for velocity calculation.")
		cmd.Println("Requirements need both effort_weeks > 0 and completed_date set.")
		return nil
	}

	width := 50
	title := "Team Velocity"
	if result.WindowWeeks > 0 {
		title = fmt.Sprintf("Team Velocity (last %d weeks)", result.WindowWeeks)
	}
	cmd.Println(output.Header(title, width))
	cmd.Println()
	cmd.Printf("  Completed requirements:  %d\n", result.CompletedCount)
	cmd.Printf("  Total effort:            %.1f weeks\n", result.TotalEffortWeeks)
	cmd.Printf("  Calendar span:           %.1f weeks\n", result.CalendarWeeks)
	cmd.Println()
	cmd.Printf("  Velocity:                %s effort-weeks/calendar-week\n",
		output.Color(fmt.Sprintf("%.2f", result.Velocity), output.Green))
	cmd.Println()

	return nil
}
