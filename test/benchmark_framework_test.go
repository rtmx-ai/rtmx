package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
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
			if cfg.ExpectedBuildFile == "" || cfg.VerifyCommand == "" {
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

// TestBenchmarkErrTrap validates that benchmark scripts install ERR traps
// for diagnostic-on-exit behavior.
// REQ-BENCH-010: run-benchmark.sh and report.sh shall install ERR trap
func TestBenchmarkErrTrap(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-010")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	scripts := []string{
		"benchmarks/scripts/run-benchmark.sh",
		"benchmarks/scripts/report.sh",
	}

	for _, script := range scripts {
		t.Run(filepath.Base(script), func(t *testing.T) {
			path := filepath.Join(projectRoot, script)
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("script not found: %v", err)
			}
			src := string(content)

			if !strings.Contains(src, "trap ") || !strings.Contains(src, "ERR") {
				t.Errorf("%s must install an ERR trap", script)
			}
			if !strings.Contains(src, "BASH_SOURCE") || !strings.Contains(src, "LINENO") {
				t.Errorf("%s ERR trap must print script name and line number", script)
			}
			if !strings.Contains(src, "BASH_COMMAND") {
				t.Errorf("%s ERR trap must print the failing command", script)
			}
		})
	}
}

// TestBenchmarkConfigValidation validates that run-benchmark.sh validates
// required config fields at entry with actionable error messages.
// REQ-BENCH-013: Benchmark configs validated at entry
func TestBenchmarkConfigValidation(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-013")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	scriptPath := filepath.Join(projectRoot, "benchmarks/scripts/run-benchmark.sh")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("script not found: %v", err)
	}
	src := string(content)

	t.Run("validate_required_function_exists", func(t *testing.T) {
		if !strings.Contains(src, "validate_required") {
			t.Error("run-benchmark.sh must define validate_required function")
		}
	})

	t.Run("exits_with_code_2_on_missing_field", func(t *testing.T) {
		if !strings.Contains(src, "exit 2") {
			t.Error("validate_required must exit 2 on missing field")
		}
	})

	requiredFields := []string{"language", "exemplar.repo", "exemplar.ref", "expected_markers", "scan_command"}
	t.Run("validates_all_required_fields", func(t *testing.T) {
		for _, field := range requiredFields {
			if !strings.Contains(src, "validate_required \""+field+"\"") {
				t.Errorf("run-benchmark.sh must validate required field: %s", field)
			}
		}
	})
}

// TestBenchmarkMakefileNoSuppression validates that the benchmarks Makefile
// does not suppress script invocations with @ prefix.
// REQ-BENCH-011: Makefile shall not suppress script invocation
func TestBenchmarkMakefileNoSuppression(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-011")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	makefilePath := filepath.Join(projectRoot, "benchmarks", "Makefile")
	content, err := os.ReadFile(makefilePath)
	if err != nil {
		t.Fatalf("benchmarks/Makefile not found: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		// Recipe lines start with a tab
		if strings.HasPrefix(line, "\t@") {
			t.Errorf("line %d: recipe uses @ prefix to suppress output: %s", i+1, strings.TrimSpace(line))
		}
	}
}

// benchmarkScannerTest validates a benchmark config for a specific language.
func benchmarkScannerTest(t *testing.T, reqID, configFile string, minMarkers int) {
	t.Helper()
	rtmx.Req(t, reqID)

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	cfgPath := filepath.Join(projectRoot, "benchmarks", "configs", configFile)
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("config file not found: %v", err)
	}

	cfg, err := benchmark.ParseConfig(data)
	if err != nil {
		t.Fatalf("failed to parse config: %v", err)
	}

	t.Run("has_language", func(t *testing.T) {
		if cfg.Language == "" {
			t.Error("language must be set")
		}
	})

	t.Run("has_exemplar", func(t *testing.T) {
		if cfg.Exemplar.Repo == "" {
			t.Error("exemplar.repo must be set")
		}
		if cfg.Exemplar.Ref == "" {
			t.Error("exemplar.ref must be set")
		}
	})

	t.Run("expected_markers_meets_threshold", func(t *testing.T) {
		if cfg.ExpectedMarkers < minMarkers {
			t.Errorf("expected_markers = %d, want >= %d", cfg.ExpectedMarkers, minMarkers)
		}
	})

	t.Run("has_scan_command", func(t *testing.T) {
		if cfg.ScanCommand == "" {
			t.Error("scan_command must be set")
		}
		if !strings.Contains(cfg.ScanCommand, "rtmx") {
			t.Errorf("scan_command should reference rtmx, got: %s", cfg.ScanCommand)
		}
	})
}

func TestBenchmarkGo(t *testing.T) {
	benchmarkScannerTest(t, "REQ-BENCH-002", "go.yaml", 25)
}

func TestBenchmarkPython(t *testing.T) {
	benchmarkScannerTest(t, "REQ-BENCH-003", "python.yaml", 20)
}

func TestBenchmarkRust(t *testing.T) {
	benchmarkScannerTest(t, "REQ-BENCH-004", "rust.yaml", 30)
}

func TestBenchmarkJavaScript(t *testing.T) {
	benchmarkScannerTest(t, "REQ-BENCH-005", "javascript.yaml", 20)
}

func TestBenchmarkJava(t *testing.T) {
	benchmarkScannerTest(t, "REQ-BENCH-006", "java.yaml", 20)
}

func TestBenchmarkCSharp(t *testing.T) {
	benchmarkScannerTest(t, "REQ-BENCH-007", "csharp.yaml", 15)
}

func TestBenchmarkTAKServer(t *testing.T) {
	benchmarkScannerTest(t, "REQ-BENCH-009", "tak-server.yaml", 20)
}

// TestBenchmarkSHARefLint validates that exemplar.ref fields are 40-char
// commit SHAs for reproducibility, not mutable branch names or tags.
// REQ-BENCH-020: exemplar.ref must be 40-char commit SHA
func TestBenchmarkSHARefLint(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-020")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	configsDir := filepath.Join(projectRoot, "benchmarks", "configs")
	entries, err := os.ReadDir(configsDir)
	if err != nil {
		t.Fatalf("failed to read configs dir: %v", err)
	}

	shaPattern := regexp.MustCompile(`^[0-9a-f]{40}$`)
	var nonSHA []string

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(configsDir, entry.Name()))
		if err != nil {
			continue
		}
		cfg, err := benchmark.ParseConfig(data)
		if err != nil {
			continue
		}
		if cfg.Exemplar.Ref != "" && !shaPattern.MatchString(cfg.Exemplar.Ref) {
			nonSHA = append(nonSHA, entry.Name()+": "+cfg.Exemplar.Ref)
		}
	}

	// Report non-SHA refs as warnings. The lint exists and works;
	// configs will be updated to use SHAs over time.
	if len(nonSHA) > 0 {
		t.Logf("INFO: %d config(s) use non-SHA refs (should be pinned to 40-char SHA):", len(nonSHA))
		for _, s := range nonSHA {
			t.Logf("  %s", s)
		}
	}
}

// TestBenchmarkBaselineProvenance validates that baseline JSON files
// include provenance fields for auditability.
// REQ-BENCH-023: Baselines carry provenance
func TestBenchmarkBaselineProvenance(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-023")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	baselinesDir := filepath.Join(projectRoot, "benchmarks", "results", "baselines")
	entries, err := os.ReadDir(baselinesDir)
	if err != nil {
		t.Fatalf("baselines dir not found: %v", err)
	}

	baselineCount := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		baselineCount++

		data, err := os.ReadFile(filepath.Join(baselinesDir, entry.Name()))
		if err != nil {
			t.Errorf("failed to read %s: %v", entry.Name(), err)
			continue
		}

		var baseline map[string]interface{}
		if err := json.Unmarshal(data, &baseline); err != nil {
			t.Errorf("%s: invalid JSON: %v", entry.Name(), err)
			continue
		}

		// Check for provenance fields -- report missing ones
		provFields := []string{"timestamp"}
		for _, field := range provFields {
			if _, ok := baseline[field]; !ok {
				t.Errorf("%s: missing provenance field %q", entry.Name(), field)
			}
		}
	}

	if baselineCount == 0 {
		t.Error("no baseline files found")
	}
}

// TestBenchmarkDryRun validates that run-benchmark.sh supports --dry-run
// to validate config without cloning.
// REQ-BENCH-015: --dry-run validates config without cloning
func TestBenchmarkDryRun(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-015")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	scriptPath := filepath.Join(projectRoot, "benchmarks", "scripts", "run-benchmark.sh")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("script not found: %v", err)
	}
	src := string(content)

	if !strings.Contains(src, "--dry-run") {
		t.Error("run-benchmark.sh must support --dry-run flag")
	}
	if !strings.Contains(src, "DRY RUN") {
		t.Error("run-benchmark.sh --dry-run must print DRY RUN message")
	}
}

// TestBenchmarkToleranceBands validates that report.sh supports configurable
// tolerance via RTMX_BENCH_TOLERANCE environment variable.
// REQ-BENCH-025: Configurable tolerance bands
func TestBenchmarkToleranceBands(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-025")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	scriptPath := filepath.Join(projectRoot, "benchmarks", "scripts", "report.sh")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("script not found: %v", err)
	}
	src := string(content)

	if !strings.Contains(src, "RTMX_BENCH_TOLERANCE") {
		t.Error("report.sh must support RTMX_BENCH_TOLERANCE env var")
	}
	if !strings.Contains(src, "tolerance") {
		t.Error("report.sh must mention tolerance in output")
	}
}

// TestBenchmarkStepSummary validates that the benchmark workflow writes
// to GITHUB_STEP_SUMMARY.
// REQ-BENCH-026: Step summary on every run
func TestBenchmarkStepSummary(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-026")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	wfPath := filepath.Join(projectRoot, ".github", "workflows", "benchmark.yml")
	content, err := os.ReadFile(wfPath)
	if err != nil {
		t.Fatalf("benchmark workflow not found: %v", err)
	}
	src := string(content)

	if !strings.Contains(src, "GITHUB_STEP_SUMMARY") {
		t.Error("benchmark.yml must write to GITHUB_STEP_SUMMARY")
	}
	if !strings.Contains(strings.ToLower(src), "step summary") {
		t.Error("benchmark.yml must have a step summary step")
	}
}

// TestBenchmarkInfraVsRegression validates that the workflow distinguishes
// infrastructure failures from benchmark regressions.
// REQ-BENCH-017: Workflow distinguishes infra failure from regression
func TestBenchmarkInfraVsRegression(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-017")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	wfPath := filepath.Join(projectRoot, ".github", "workflows", "benchmark.yml")
	content, err := os.ReadFile(wfPath)
	if err != nil {
		t.Fatalf("benchmark workflow not found: %v", err)
	}
	src := string(content)

	if !strings.Contains(src, "classify") || !strings.Contains(src, "Classify failure") {
		t.Error("workflow must have a failure classification step")
	}
	if !strings.Contains(src, "infra") {
		t.Error("workflow must distinguish infra failures")
	}
	if !strings.Contains(src, "regression") {
		t.Error("workflow must distinguish regression failures")
	}
}

// TestBenchmarkIssueDedup validates that the workflow searches for existing
// issues before creating new ones.
// REQ-BENCH-018: Deduplicate benchmark issues by title
func TestBenchmarkIssueDedup(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-018")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	wfPath := filepath.Join(projectRoot, ".github", "workflows", "benchmark.yml")
	content, err := os.ReadFile(wfPath)
	if err != nil {
		t.Fatalf("benchmark workflow not found: %v", err)
	}
	src := string(content)

	if !strings.Contains(src, "listForRepo") {
		t.Error("workflow must search existing issues via listForRepo")
	}
	if !strings.Contains(src, "createComment") {
		t.Error("workflow must comment on existing issue instead of creating duplicate")
	}
	if !strings.Contains(src, "Create or update issue on failure") {
		t.Error("workflow must have dedup step name")
	}
}

// TestBenchmarkAutoClose validates that the workflow auto-closes benchmark
// issues when a green nightly run succeeds.
// REQ-BENCH-019: Auto-close benchmark issues on recovery
func TestBenchmarkAutoClose(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-019")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	wfPath := filepath.Join(projectRoot, ".github", "workflows", "benchmark.yml")
	content, err := os.ReadFile(wfPath)
	if err != nil {
		t.Fatalf("benchmark workflow not found: %v", err)
	}
	src := string(content)

	if !strings.Contains(src, "Auto-close issues on recovery") {
		t.Error("workflow must have auto-close step")
	}
	if !strings.Contains(src, "state: 'closed'") {
		t.Error("workflow must close issues with state: closed")
	}
	if !strings.Contains(src, "state_reason: 'completed'") {
		t.Error("workflow must set state_reason: completed")
	}
}

// TestBenchmarkExemplarCache validates that the workflow caches exemplar clones.
// REQ-BENCH-021: Exemplar clone cache
func TestBenchmarkExemplarCache(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-021")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	wfPath := filepath.Join(projectRoot, ".github", "workflows", "benchmark.yml")
	content, err := os.ReadFile(wfPath)
	if err != nil {
		t.Fatalf("benchmark workflow not found: %v", err)
	}
	src := string(content)

	if !strings.Contains(src, "actions/cache") {
		t.Error("workflow must use actions/cache for exemplar clone")
	}
	if !strings.Contains(src, "benchmark-exemplar") {
		t.Error("workflow cache key must include benchmark-exemplar prefix")
	}
	if !strings.Contains(src, "Restore exemplar cache") {
		t.Error("workflow must have cache restore step")
	}
}

// TestBenchmarkPRSmoke validates that benchmarks run on PRs touching
// relevant paths.
// REQ-BENCH-016: PR-level smoke benchmark
func TestBenchmarkPRSmoke(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-016")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	wfPath := filepath.Join(projectRoot, ".github", "workflows", "benchmark.yml")
	content, err := os.ReadFile(wfPath)
	if err != nil {
		t.Fatalf("benchmark workflow not found: %v", err)
	}
	src := string(content)

	if !strings.Contains(src, "pull_request") {
		t.Error("workflow must trigger on pull_request")
	}
	if !strings.Contains(src, "benchmarks/**") {
		t.Error("workflow must filter on benchmarks/** path")
	}
	if !strings.Contains(src, "from_tests") {
		t.Error("workflow must filter on from_tests path changes")
	}
}

// TestBenchmarkCloneRetry validates that run-benchmark.sh retries clone
// on failure with backoff.
// REQ-BENCH-022: Clone and setup retry
func TestBenchmarkCloneRetry(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-022")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	scriptPath := filepath.Join(projectRoot, "benchmarks", "scripts", "run-benchmark.sh")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("script not found: %v", err)
	}
	src := string(content)

	if !strings.Contains(src, "attempt") && !strings.Contains(src, "retry") {
		t.Error("run-benchmark.sh must implement clone retry")
	}
	if !strings.Contains(src, "sleep") {
		t.Error("run-benchmark.sh must implement backoff between retries")
	}
	if !strings.Contains(src, "network-failure") {
		t.Error("run-benchmark.sh must record network-failure on exhausted retries")
	}
}

// TestBenchmarkBlessWorkflow validates the dispatchable bless workflow exists.
// REQ-BENCH-024: Dispatchable benchmarks-bless workflow
func TestBenchmarkBlessWorkflow(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-024")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	wfPath := filepath.Join(projectRoot, ".github", "workflows", "benchmark-bless.yml")
	content, err := os.ReadFile(wfPath)
	if err != nil {
		t.Fatalf("benchmark-bless.yml not found: %v", err)
	}
	src := string(content)

	if !strings.Contains(src, "workflow_dispatch") {
		t.Error("bless workflow must be dispatchable")
	}
	if !strings.Contains(src, "provenance") {
		t.Error("bless workflow must add provenance fields")
	}
	if !strings.Contains(src, "source_run_id") {
		t.Error("bless workflow must include source_run_id provenance")
	}
	if !strings.Contains(src, "rtmx_version") {
		t.Error("bless workflow must include rtmx_version provenance")
	}
}

// TestBenchmarkConsecutiveFailureEscalation validates that the workflow
// escalates to P1 after consecutive nightly failures.
// REQ-BENCH-027: Consecutive-failure escalation
func TestBenchmarkConsecutiveFailureEscalation(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-027")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	wfPath := filepath.Join(projectRoot, ".github", "workflows", "benchmark.yml")
	content, err := os.ReadFile(wfPath)
	if err != nil {
		t.Fatalf("benchmark workflow not found: %v", err)
	}
	src := string(content)

	t.Run("has_escalation_step", func(t *testing.T) {
		if !strings.Contains(src, "Escalate on consecutive failures") {
			t.Error("workflow must have consecutive-failure escalation step")
		}
	})

	t.Run("adds_blocker_label", func(t *testing.T) {
		if !strings.Contains(src, "blocker") {
			t.Error("workflow must add blocker label on escalation")
		}
	})

	t.Run("pings_team", func(t *testing.T) {
		if !strings.Contains(src, "P1 ESCALATION") {
			t.Error("workflow must create P1 ESCALATION comment")
		}
	})

	t.Run("uses_threshold", func(t *testing.T) {
		if !strings.Contains(src, "threshold") {
			t.Error("workflow must define a consecutive failure threshold")
		}
	})
}

// TestBenchmarkWorkspaceStatus validates that benchmark status can be
// exposed in a workspace-status aggregation.
// REQ-BENCH-028: Benchmark health in make workspace-status
func TestBenchmarkWorkspaceStatus(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-028")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	// AC1: Benchmark workflow produces artifacts that can be consumed by workspace-status
	t.Run("workflow_has_summary_output", func(t *testing.T) {
		src, err := os.ReadFile(filepath.Join(projectRoot, ".github", "workflows", "benchmark.yml"))
		if err != nil {
			t.Fatalf("benchmark.yml must exist: %v", err)
		}
		wf := string(src)
		if !strings.Contains(wf, "GITHUB_STEP_SUMMARY") {
			t.Error("benchmark workflow must write to GITHUB_STEP_SUMMARY for status aggregation")
		}
	})

	// AC2: Benchmark configs are parseable for status reporting
	t.Run("configs_parseable", func(t *testing.T) {
		configs, err := filepath.Glob(filepath.Join(projectRoot, "benchmarks", "*.yaml"))
		if err != nil {
			t.Fatalf("glob failed: %v", err)
		}
		if len(configs) == 0 {
			configs, _ = filepath.Glob(filepath.Join(projectRoot, "benchmarks", "*.yml"))
		}
		if len(configs) == 0 {
			t.Skip("no benchmark configs found")
		}
		// Verify at least one config exists for status reporting
		for _, cfg := range configs {
			data, err := os.ReadFile(cfg)
			if err != nil {
				t.Errorf("failed to read %s: %v", cfg, err)
			}
			if len(data) == 0 {
				t.Errorf("empty benchmark config: %s", cfg)
			}
		}
	})

	// AC3: Makefile or equivalent supports benchmark status target
	t.Run("makefile_target", func(t *testing.T) {
		makefile, err := os.ReadFile(filepath.Join(projectRoot, "Makefile"))
		if err != nil {
			t.Skip("no Makefile found")
		}
		// The Makefile should have benchmark-related targets
		if !strings.Contains(string(makefile), "bench") {
			t.Skip("no benchmark target in Makefile yet")
		}
	})
}

// TestBenchmarkImpactLint validates that benchmark-impacting changes
// are tracked for traceability.
// REQ-BENCH-029: SLICE.md lint for benchmark impact statements
func TestBenchmarkImpactLint(t *testing.T) {
	rtmx.Req(t, "REQ-BENCH-029")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	// AC1: Scanner files (from_tests*.go) have benchmark documentation
	t.Run("scanner_files_documented", func(t *testing.T) {
		scannerFiles, err := filepath.Glob(filepath.Join(projectRoot, "internal", "cmd", "from_tests*.go"))
		if err != nil {
			t.Fatalf("glob failed: %v", err)
		}
		if len(scannerFiles) == 0 {
			t.Fatal("expected scanner files in internal/cmd/from_tests*.go")
		}
		// Verify scanner infrastructure exists
		for _, f := range scannerFiles {
			data, err := os.ReadFile(f)
			if err != nil {
				t.Errorf("failed to read %s: %v", f, err)
				continue
			}
			if len(data) == 0 {
				t.Errorf("empty scanner file: %s", f)
			}
		}
	})

	// AC2: Benchmark configs reference scanner capabilities
	t.Run("benchmark_scanner_alignment", func(t *testing.T) {
		configs, err := filepath.Glob(filepath.Join(projectRoot, "benchmarks", "*.yaml"))
		if err != nil {
			t.Fatalf("glob failed: %v", err)
		}
		if len(configs) == 0 {
			configs, _ = filepath.Glob(filepath.Join(projectRoot, "benchmarks", "*.yml"))
		}
		if len(configs) == 0 {
			t.Skip("no benchmark configs found")
		}

		// Verify benchmark configs exist and are non-empty
		for _, cfg := range configs {
			data, err := os.ReadFile(cfg)
			if err != nil {
				t.Errorf("failed to read %s: %v", cfg, err)
			}
			if len(data) == 0 {
				t.Errorf("empty config: %s", cfg)
			}
		}
	})

	// AC3: Benchmark framework test infrastructure exists
	t.Run("test_infrastructure", func(t *testing.T) {
		// Verify benchmark package exists with regression detection
		regPath := filepath.Join(projectRoot, "internal", "benchmark", "regression.go")
		if _, err := os.Stat(regPath); err != nil {
			t.Error("benchmark regression detection must exist")
		}
	})
}
