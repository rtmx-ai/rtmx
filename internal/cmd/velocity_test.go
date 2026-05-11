package cmd

import (
	"testing"
	"time"

	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func TestComputeVelocity(t *testing.T) {
	rtmx.Req(t, "REQ-PLAN-011",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	today := time.Now().Format("2006-01-02")
	twoWeeksAgo := time.Now().AddDate(0, 0, -14).Format("2006-01-02")
	sixWeeksAgo := time.Now().AddDate(0, 0, -42).Format("2006-01-02")

	tests := []struct {
		name           string
		reqs           []*database.Requirement
		window         int
		wantInsuffData bool
		wantCount      int
		wantEffort     float64
	}{
		{
			name:           "no_data",
			reqs:           []*database.Requirement{},
			wantInsuffData: true,
		},
		{
			name: "missing_effort",
			reqs: []*database.Requirement{
				{Status: database.StatusComplete, CompletedDate: today},
			},
			wantInsuffData: true,
		},
		{
			name: "missing_date",
			reqs: []*database.Requirement{
				{Status: database.StatusComplete, EffortWeeks: 1.0},
			},
			wantInsuffData: true,
		},
		{
			name: "incomplete_status",
			reqs: []*database.Requirement{
				{Status: database.StatusMissing, EffortWeeks: 1.0, CompletedDate: today},
			},
			wantInsuffData: true,
		},
		{
			name: "single_requirement",
			reqs: []*database.Requirement{
				{Status: database.StatusComplete, EffortWeeks: 2.0, CompletedDate: today},
			},
			wantCount:  1,
			wantEffort: 2.0,
		},
		{
			name: "multiple_requirements",
			reqs: []*database.Requirement{
				{Status: database.StatusComplete, EffortWeeks: 1.0, CompletedDate: twoWeeksAgo},
				{Status: database.StatusComplete, EffortWeeks: 2.0, CompletedDate: today},
				{Status: database.StatusMissing, EffortWeeks: 3.0, CompletedDate: today}, // excluded
			},
			wantCount:  2,
			wantEffort: 3.0,
		},
		{
			name: "window_filters",
			reqs: []*database.Requirement{
				{Status: database.StatusComplete, EffortWeeks: 1.0, CompletedDate: sixWeeksAgo},
				{Status: database.StatusComplete, EffortWeeks: 2.0, CompletedDate: today},
			},
			window:     4,
			wantCount:  1,
			wantEffort: 2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeVelocity(tt.reqs, tt.window)

			if result.InsufficientData != tt.wantInsuffData {
				t.Errorf("InsufficientData = %v, want %v", result.InsufficientData, tt.wantInsuffData)
			}
			if result.CompletedCount != tt.wantCount {
				t.Errorf("CompletedCount = %d, want %d", result.CompletedCount, tt.wantCount)
			}
			if result.TotalEffortWeeks != tt.wantEffort {
				t.Errorf("TotalEffortWeeks = %f, want %f", result.TotalEffortWeeks, tt.wantEffort)
			}
			if !tt.wantInsuffData && result.Velocity <= 0 {
				t.Error("Velocity should be positive for valid data")
			}
		})
	}
}

func TestComputeVelocityPositive(t *testing.T) {
	rtmx.Req(t, "REQ-PLAN-011",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)

	twoWeeksAgo := time.Now().AddDate(0, 0, -14).Format("2006-01-02")
	today := time.Now().Format("2006-01-02")

	reqs := []*database.Requirement{
		{Status: database.StatusComplete, EffortWeeks: 2.0, CompletedDate: twoWeeksAgo},
		{Status: database.StatusComplete, EffortWeeks: 3.0, CompletedDate: today},
	}

	result := ComputeVelocity(reqs, 0)

	if result.InsufficientData {
		t.Fatal("should have sufficient data")
	}
	if result.Velocity <= 0 {
		t.Errorf("velocity should be positive, got %f", result.Velocity)
	}
	if result.CalendarWeeks < 2.0 {
		t.Errorf("calendar weeks should be >= 2, got %f", result.CalendarWeeks)
	}
	if result.TotalEffortWeeks != 5.0 {
		t.Errorf("total effort should be 5.0, got %f", result.TotalEffortWeeks)
	}
}
