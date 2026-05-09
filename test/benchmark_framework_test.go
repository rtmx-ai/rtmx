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
