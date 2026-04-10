package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/benchmark"
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestBenchmarkFramework validates that the benchmark infrastructure exists
// and is correctly configured.
// REQ-BENCH-001: Benchmark framework and orchestration
func TestBenchmarkFramework(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-001")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	benchDir := filepath.Join(projectRoot, "benchmarks")

	t.Run("benchmarks_directory_exists", func(t *testing.T) {
		makefile := filepath.Join(benchDir, "Makefile")
		if _, err := os.Stat(makefile); err != nil {
			t.Fatalf("benchmarks/Makefile not found: %v", err)
		}
	})

	t.Run("makefile_has_run_target", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(benchDir, "Makefile"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "run:") {
			t.Error("benchmarks/Makefile must have a 'run:' target")
		}
	})

	t.Run("makefile_has_compare_target", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(benchDir, "Makefile"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "compare:") {
			t.Error("benchmarks/Makefile must have a 'compare:' target")
		}
	})

	t.Run("run_benchmark_script_exists", func(t *testing.T) {
		script := filepath.Join(benchDir, "scripts", "run-benchmark.sh")
		if _, err := os.Stat(script); err != nil {
			t.Fatalf("benchmarks/scripts/run-benchmark.sh not found: %v", err)
		}
	})

	t.Run("report_script_exists", func(t *testing.T) {
		script := filepath.Join(benchDir, "scripts", "report.sh")
		if _, err := os.Stat(script); err != nil {
			t.Fatalf("benchmarks/scripts/report.sh not found: %v", err)
		}
	})

	t.Run("go_config_exists", func(t *testing.T) {
		cfg := filepath.Join(benchDir, "configs", "go.yaml")
		if _, err := os.Stat(cfg); err != nil {
			t.Fatalf("benchmarks/configs/go.yaml not found: %v", err)
		}
	})

	t.Run("go_config_is_valid", func(t *testing.T) {
		cfg, err := benchmark.LoadConfig(filepath.Join(benchDir, "configs", "go.yaml"))
		if err != nil {
			t.Fatalf("LoadConfig(go.yaml) error: %v", err)
		}
		if cfg.Language != "go" {
			t.Errorf("Language = %q, want %q", cfg.Language, "go")
		}
		if cfg.Exemplar.Repo == "" {
			t.Error("Exemplar.Repo must not be empty")
		}
		if cfg.ExpectedMarkers <= 0 {
			t.Error("ExpectedMarkers must be positive")
		}
	})

	t.Run("benchmark_workflow_exists", func(t *testing.T) {
		wf := filepath.Join(projectRoot, ".github", "workflows", "benchmark.yml")
		if _, err := os.Stat(wf); err != nil {
			t.Fatalf(".github/workflows/benchmark.yml not found: %v", err)
		}
	})

	t.Run("benchmark_workflow_has_schedule", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".github", "workflows", "benchmark.yml"))
		if err != nil {
			t.Fatal(err)
		}
		doc := string(content)
		if !strings.Contains(doc, "schedule:") {
			t.Error("benchmark workflow must have schedule trigger")
		}
		if !strings.Contains(doc, "cron:") {
			t.Error("benchmark workflow must have cron expression")
		}
	})

	t.Run("benchmark_workflow_has_matrix", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".github", "workflows", "benchmark.yml"))
		if err != nil {
			t.Fatal(err)
		}
		doc := string(content)
		if !strings.Contains(doc, "matrix:") {
			t.Error("benchmark workflow must use matrix strategy")
		}
		if !strings.Contains(doc, "language:") {
			t.Error("benchmark workflow matrix must include language")
		}
	})

	t.Run("run_script_clones_at_ref", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(benchDir, "scripts", "run-benchmark.sh"))
		if err != nil {
			t.Fatal(err)
		}
		doc := string(content)
		if !strings.Contains(doc, "git clone") {
			t.Error("run-benchmark.sh must clone the exemplar repo")
		}
		if !strings.Contains(doc, "--depth") {
			t.Error("run-benchmark.sh must use shallow clone (--depth)")
		}
	})

	t.Run("run_script_applies_patch", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(benchDir, "scripts", "run-benchmark.sh"))
		if err != nil {
			t.Fatal(err)
		}
		doc := string(content)
		if !strings.Contains(doc, "apply") || !strings.Contains(doc, "MARKER_PATCH") {
			t.Error("run-benchmark.sh must apply marker patches")
		}
	})

	t.Run("run_script_runs_scan", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(benchDir, "scripts", "run-benchmark.sh"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "rtmx from-tests") {
			t.Error("run-benchmark.sh must run rtmx from-tests")
		}
	})
}
