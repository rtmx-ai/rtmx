package benchmark

import (
	"encoding/json"
	"testing"
)

func TestCompareResults(t *testing.T) {
	tests := []struct {
		name       string
		baseline   BenchmarkResult
		current    BenchmarkResult
		wantCount  int
		wantFields []string
	}{
		{
			name:      "identical results no regressions",
			baseline:  BenchmarkResult{Language: "go", MarkerCount: 25, TestsPassed: 25, TestsFailed: 0, VerifyStatus: "pass"},
			current:   BenchmarkResult{Language: "go", MarkerCount: 25, TestsPassed: 25, TestsFailed: 0, VerifyStatus: "pass"},
			wantCount: 0,
		},
		{
			name:       "marker count dropped",
			baseline:   BenchmarkResult{Language: "go", MarkerCount: 25, TestsPassed: 25, TestsFailed: 0, VerifyStatus: "pass"},
			current:    BenchmarkResult{Language: "go", MarkerCount: 20, TestsPassed: 20, TestsFailed: 0, VerifyStatus: "pass"},
			wantCount:  1,
			wantFields: []string{"marker_count"},
		},
		{
			name:       "tests now failing",
			baseline:   BenchmarkResult{Language: "go", MarkerCount: 25, TestsPassed: 25, TestsFailed: 0, VerifyStatus: "pass"},
			current:    BenchmarkResult{Language: "go", MarkerCount: 25, TestsPassed: 22, TestsFailed: 3, VerifyStatus: "pass"},
			wantCount:  1,
			wantFields: []string{"tests_failed"},
		},
		{
			name:       "verify status changed to fail",
			baseline:   BenchmarkResult{Language: "go", MarkerCount: 25, TestsPassed: 25, TestsFailed: 0, VerifyStatus: "pass"},
			current:    BenchmarkResult{Language: "go", MarkerCount: 25, TestsPassed: 25, TestsFailed: 0, VerifyStatus: "fail"},
			wantCount:  1,
			wantFields: []string{"verify_status"},
		},
		{
			name:      "marker count increased is not regression",
			baseline:  BenchmarkResult{Language: "go", MarkerCount: 25, TestsPassed: 25, TestsFailed: 0, VerifyStatus: "pass"},
			current:   BenchmarkResult{Language: "go", MarkerCount: 30, TestsPassed: 30, TestsFailed: 0, VerifyStatus: "pass"},
			wantCount: 0,
		},
		{
			name:      "more tests pass is not regression",
			baseline:  BenchmarkResult{Language: "go", MarkerCount: 25, TestsPassed: 20, TestsFailed: 5, VerifyStatus: "pass"},
			current:   BenchmarkResult{Language: "go", MarkerCount: 25, TestsPassed: 25, TestsFailed: 0, VerifyStatus: "pass"},
			wantCount: 0,
		},
		{
			name:       "multiple regressions",
			baseline:   BenchmarkResult{Language: "go", MarkerCount: 25, TestsPassed: 25, TestsFailed: 0, VerifyStatus: "pass"},
			current:    BenchmarkResult{Language: "go", MarkerCount: 15, TestsPassed: 10, TestsFailed: 5, VerifyStatus: "fail"},
			wantCount:  3,
			wantFields: []string{"marker_count", "tests_failed", "verify_status"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := CompareResults(tt.baseline, tt.current)
			if len(report.Regressions) != tt.wantCount {
				t.Errorf("CompareResults() got %d regressions, want %d", len(report.Regressions), tt.wantCount)
				for _, r := range report.Regressions {
					t.Logf("  regression: %s (%s)", r.Field, r.Message)
				}
			}
			for _, wantField := range tt.wantFields {
				found := false
				for _, r := range report.Regressions {
					if r.Field == wantField {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected regression on field %q not found", wantField)
				}
			}
		})
	}
}

func TestRegressionReportHasRegressions(t *testing.T) {
	t.Run("no regressions", func(t *testing.T) {
		report := RegressionReport{Language: "go"}
		if report.HasRegressions() {
			t.Error("HasRegressions() = true, want false")
		}
		if report.ExitCode() != 0 {
			t.Errorf("ExitCode() = %d, want 0", report.ExitCode())
		}
	})

	t.Run("has regressions", func(t *testing.T) {
		report := RegressionReport{
			Language:    "go",
			Regressions: []Regression{{Field: "marker_count", Message: "dropped"}},
		}
		if !report.HasRegressions() {
			t.Error("HasRegressions() = false, want true")
		}
		if report.ExitCode() != 1 {
			t.Errorf("ExitCode() = %d, want 1", report.ExitCode())
		}
	})
}

func TestBenchmarkResultJSON(t *testing.T) {
	original := BenchmarkResult{
		Language:     "go",
		MarkerCount:  25,
		MarkersFound: 25,
		TestsPassed:  25,
		TestsFailed:  0,
		VerifyStatus: "pass",
		Timestamp:    "2026-04-05T04:00:00Z",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var roundtrip BenchmarkResult
	if err := json.Unmarshal(data, &roundtrip); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if roundtrip != original {
		t.Errorf("round-trip mismatch:\n  got:  %+v\n  want: %+v", roundtrip, original)
	}
}
