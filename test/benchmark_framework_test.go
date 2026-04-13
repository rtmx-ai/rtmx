package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/internal/benchmark"
	"github.com/rtmx-ai/rtmx/internal/cmd"
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

	t.Run("detect_test_command_matches_configs", func(t *testing.T) {
		// For each benchmark config with an expected_build_file, verify that
		// DetectTestCommand returns a command consistent with the config's
		// verify_command. This exercises the auto-detection fix for non-Go projects.
		configs, err := filepath.Glob(filepath.Join(benchDir, "configs", "*.yaml"))
		if err != nil {
			t.Fatal(err)
		}
		if len(configs) == 0 {
			t.Fatal("no benchmark configs found")
		}

		tested := 0
		for _, cfgPath := range configs {
			cfg, err := benchmark.LoadConfig(cfgPath)
			if err != nil {
				t.Errorf("LoadConfig(%s) error: %v", filepath.Base(cfgPath), err)
				continue
			}
			if cfg.ExpectedBuildFile == "" {
				continue
			}

			// Create a temp dir with the expected build file
			dir := t.TempDir()
			buildFile := filepath.Join(dir, cfg.ExpectedBuildFile)
			if err := os.WriteFile(buildFile, []byte(""), 0644); err != nil {
				t.Fatal(err)
			}

			detectedCmd, _ := cmd.DetectTestCommand(dir)

			// The detected command should match the first word of verify_command
			verifyFirst := strings.Fields(cfg.VerifyCommand)[0]
			if detectedCmd != verifyFirst {
				t.Errorf("config %s (build_file=%s): DetectTestCommand=%q, verify_command starts with %q",
					filepath.Base(cfgPath), cfg.ExpectedBuildFile, detectedCmd, verifyFirst)
			}
			tested++
		}

		if tested < 5 {
			t.Errorf("expected at least 5 configs with expected_build_file, got %d", tested)
		}
	})
}
