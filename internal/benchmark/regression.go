package benchmark

import "fmt"

// BenchmarkResult captures the outcome of running a single language benchmark.
type BenchmarkResult struct {
	Language     string `json:"language"`
	MarkerCount  int    `json:"marker_count"`
	MarkersFound int    `json:"markers_found"`
	TestsPassed  int    `json:"tests_passed"`
	TestsFailed  int    `json:"tests_failed"`
	VerifyStatus string `json:"verify_status"`
	Timestamp    string `json:"timestamp"`
}

// Regression describes a single metric that regressed between baseline and current.
type Regression struct {
	Field    string `json:"field"`
	Baseline any    `json:"baseline"`
	Current  any    `json:"current"`
	Message  string `json:"message"`
}

// RegressionReport is the result of comparing a current benchmark run to a baseline.
type RegressionReport struct {
	Language    string       `json:"language"`
	Regressions []Regression `json:"regressions,omitempty"`
}

// CompareResults compares current benchmark results against a baseline and
// returns a report of any regressions. Improvements (higher marker counts,
// fewer failures) are not flagged.
func CompareResults(baseline, current BenchmarkResult) RegressionReport {
	report := RegressionReport{Language: current.Language}

	if current.MarkerCount < baseline.MarkerCount {
		report.Regressions = append(report.Regressions, Regression{
			Field:    "marker_count",
			Baseline: baseline.MarkerCount,
			Current:  current.MarkerCount,
			Message:  fmt.Sprintf("marker count dropped from %d to %d", baseline.MarkerCount, current.MarkerCount),
		})
	}

	if current.TestsFailed > baseline.TestsFailed {
		report.Regressions = append(report.Regressions, Regression{
			Field:    "tests_failed",
			Baseline: baseline.TestsFailed,
			Current:  current.TestsFailed,
			Message:  fmt.Sprintf("test failures increased from %d to %d", baseline.TestsFailed, current.TestsFailed),
		})
	}

	if current.VerifyStatus != baseline.VerifyStatus && current.VerifyStatus == "fail" {
		report.Regressions = append(report.Regressions, Regression{
			Field:    "verify_status",
			Baseline: baseline.VerifyStatus,
			Current:  current.VerifyStatus,
			Message:  fmt.Sprintf("verify status changed from %q to %q", baseline.VerifyStatus, current.VerifyStatus),
		})
	}

	return report
}

// HasRegressions returns true if any regressions were detected.
func (r *RegressionReport) HasRegressions() bool {
	return len(r.Regressions) > 0
}

// ExitCode returns 1 if there are regressions, 0 otherwise.
func (r *RegressionReport) ExitCode() int {
	if r.HasRegressions() {
		return 1
	}
	return 0
}
