package cmd

import (
	"testing"

	"github.com/rtmx-ai/rtmx/internal/config"
	"github.com/rtmx-ai/rtmx/internal/database"
	"github.com/rtmx-ai/rtmx/internal/results"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

func compRes(scope, technique, env string, passed bool) results.Result {
	return results.Result{
		Passed: passed,
		Marker: results.Marker{
			ReqID:     "REQ-SW-001",
			TestName:  "t_" + scope + "_" + technique,
			TestFile:  "test_x.py",
			Scope:     scope,
			Technique: technique,
			Env:       env,
		},
	}
}

func boolPtr(b bool) *bool { return &b }

func TestDetermineStatusWithPolicy(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-009",
		rtmx.Scope("unit"),
		rtmx.Technique("nominal"),
		rtmx.Env("simulation"),
	)
	simple := config.CompletenessConfig{Policy: "simple"}
	combo3 := config.CompletenessConfig{
		Policy:          "combinations",
		Dimensions:      []string{"scope", "technique"},
		MinCombinations: 3,
	}

	tests := []struct {
		name     string
		results  []results.Result
		current  database.Status
		policy   config.CompletenessConfig
		expected database.Status
	}{
		{
			name:     "simple: one passing test completes",
			results:  []results.Result{compRes("unit", "nominal", "simulation", true)},
			current:  database.StatusMissing,
			policy:   simple,
			expected: database.StatusComplete,
		},
		{
			name:     "simple: a failing test downgrades complete to partial",
			results:  []results.Result{compRes("unit", "nominal", "simulation", false)},
			current:  database.StatusComplete,
			policy:   simple,
			expected: database.StatusPartial,
		},
		{
			name: "combinations: three distinct tuples complete the requirement",
			results: []results.Result{
				compRes("unit", "nominal", "simulation", true),
				compRes("integration", "monte_carlo", "simulation", true),
				compRes("system", "stress", "simulation", true),
			},
			current:  database.StatusMissing,
			policy:   combo3,
			expected: database.StatusComplete,
		},
		{
			name: "combinations: two distinct tuples is only partial",
			results: []results.Result{
				compRes("unit", "nominal", "simulation", true),
				compRes("integration", "monte_carlo", "simulation", true),
			},
			current:  database.StatusMissing,
			policy:   combo3,
			expected: database.StatusPartial,
		},
		{
			name: "combinations: duplicate tuples do not count twice",
			results: []results.Result{
				compRes("unit", "nominal", "simulation", true),
				compRes("unit", "nominal", "hil", true), // same (scope,technique)
				compRes("integration", "stress", "simulation", true),
			},
			current:  database.StatusMissing,
			policy:   combo3,
			expected: database.StatusPartial, // only 2 distinct (scope,technique)
		},
		{
			name: "combinations: a failure still downgrades complete by default",
			results: []results.Result{
				compRes("unit", "nominal", "simulation", true),
				compRes("integration", "monte_carlo", "simulation", true),
				compRes("system", "stress", "simulation", false),
			},
			current:  database.StatusComplete,
			policy:   combo3,
			expected: database.StatusPartial,
		},
		{
			name: "combinations: require_all_pass=false ignores failures when combos satisfied",
			results: []results.Result{
				compRes("unit", "nominal", "simulation", true),
				compRes("integration", "monte_carlo", "simulation", true),
				compRes("system", "stress", "simulation", true),
				compRes("system", "parametric", "simulation", false),
			},
			current: database.StatusMissing,
			policy: config.CompletenessConfig{
				Policy:          "combinations",
				Dimensions:      []string{"scope", "technique"},
				MinCombinations: 3,
				RequireAllPass:  boolPtr(false),
			},
			expected: database.StatusComplete,
		},
		{
			name:     "combinations: no dimensions falls back to simple",
			results:  []results.Result{compRes("unit", "nominal", "simulation", true)},
			current:  database.StatusMissing,
			policy:   config.CompletenessConfig{Policy: "combinations", MinCombinations: 3},
			expected: database.StatusComplete,
		},
		{
			name:     "no passing evidence keeps current status",
			results:  []results.Result{},
			current:  database.StatusPartial,
			policy:   combo3,
			expected: database.StatusPartial,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineStatusWithPolicy(tt.results, tt.current, tt.policy)
			if got != tt.expected {
				t.Errorf("determineStatusWithPolicy() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDimensionTupleKey(t *testing.T) {
	rtmx.Req(t, "REQ-VERIFY-009",
		rtmx.Scope("unit"),
		rtmx.Technique("boundary"),
		rtmx.Env("simulation"),
	)
	m := results.Marker{Scope: "unit", Technique: "nominal", Env: "simulation"}
	if got := dimensionTupleKey(m, []string{"scope", "technique"}); got != "unit\x1fnominal" {
		t.Errorf("dimensionTupleKey = %q", got)
	}
	// unknown dimension contributes an empty component (no panic)
	if got := dimensionTupleKey(m, []string{"scope", "bogus"}); got != "unit\x1f" {
		t.Errorf("dimensionTupleKey with unknown dim = %q", got)
	}
}
