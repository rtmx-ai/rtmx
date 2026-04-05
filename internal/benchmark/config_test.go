package benchmark

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseBenchmarkConfig(t *testing.T) {
	tests := []struct {
		name            string
		yaml            string
		wantLanguage    string
		wantRepo        string
		wantRef         string
		wantLicense     string
		wantDepth       int
		wantPatch       string
		wantMarkers     int
		wantScan        string
		wantVerify      string
		wantTimeout     int
		wantSetupCount  int
	}{
		{
			name: "full config",
			yaml: `
language: go
exemplar:
  repo: cli/cli
  ref: v2.60.0
  license: MIT
clone_depth: 1
setup_commands:
  - go mod download
marker_patch: patches/go/cli-cli.patch
expected_markers: 25
scan_command: rtmx from-tests --format json .
verify_command: go test -json ./...
timeout_minutes: 10
`,
			wantLanguage:   "go",
			wantRepo:       "cli/cli",
			wantRef:        "v2.60.0",
			wantLicense:    "MIT",
			wantDepth:      1,
			wantPatch:      "patches/go/cli-cli.patch",
			wantMarkers:    25,
			wantScan:       "rtmx from-tests --format json .",
			wantVerify:     "go test -json ./...",
			wantTimeout:    10,
			wantSetupCount: 1,
		},
		{
			name: "minimal config with defaults",
			yaml: `
language: python
exemplar:
  repo: psf/requests
  ref: v2.32.0
expected_markers: 20
scan_command: rtmx from-tests --format json .
`,
			wantLanguage: "python",
			wantRepo:     "psf/requests",
			wantRef:      "v2.32.0",
			wantMarkers:  20,
			wantScan:     "rtmx from-tests --format json .",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := ParseConfig([]byte(tt.yaml))
			if err != nil {
				t.Fatalf("ParseConfig() error = %v", err)
			}
			if cfg.Language != tt.wantLanguage {
				t.Errorf("Language = %q, want %q", cfg.Language, tt.wantLanguage)
			}
			if cfg.Exemplar.Repo != tt.wantRepo {
				t.Errorf("Exemplar.Repo = %q, want %q", cfg.Exemplar.Repo, tt.wantRepo)
			}
			if cfg.Exemplar.Ref != tt.wantRef {
				t.Errorf("Exemplar.Ref = %q, want %q", cfg.Exemplar.Ref, tt.wantRef)
			}
			if tt.wantLicense != "" && cfg.Exemplar.License != tt.wantLicense {
				t.Errorf("Exemplar.License = %q, want %q", cfg.Exemplar.License, tt.wantLicense)
			}
			if tt.wantDepth != 0 && cfg.CloneDepth != tt.wantDepth {
				t.Errorf("CloneDepth = %d, want %d", cfg.CloneDepth, tt.wantDepth)
			}
			if tt.wantPatch != "" && cfg.MarkerPatch != tt.wantPatch {
				t.Errorf("MarkerPatch = %q, want %q", cfg.MarkerPatch, tt.wantPatch)
			}
			if cfg.ExpectedMarkers != tt.wantMarkers {
				t.Errorf("ExpectedMarkers = %d, want %d", cfg.ExpectedMarkers, tt.wantMarkers)
			}
			if cfg.ScanCommand != tt.wantScan {
				t.Errorf("ScanCommand = %q, want %q", cfg.ScanCommand, tt.wantScan)
			}
			if tt.wantVerify != "" && cfg.VerifyCommand != tt.wantVerify {
				t.Errorf("VerifyCommand = %q, want %q", cfg.VerifyCommand, tt.wantVerify)
			}
			if tt.wantTimeout != 0 && cfg.TimeoutMinutes != tt.wantTimeout {
				t.Errorf("TimeoutMinutes = %d, want %d", cfg.TimeoutMinutes, tt.wantTimeout)
			}
			if tt.wantSetupCount != 0 && len(cfg.SetupCommands) != tt.wantSetupCount {
				t.Errorf("SetupCommands count = %d, want %d", len(cfg.SetupCommands), tt.wantSetupCount)
			}
		})
	}
}

func TestValidateBenchmarkConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  BenchmarkConfig
		wantErr string
	}{
		{
			name:    "missing language",
			config:  BenchmarkConfig{Exemplar: ExemplarConfig{Repo: "a/b", Ref: "v1"}, ExpectedMarkers: 1, ScanCommand: "scan"},
			wantErr: "language is required",
		},
		{
			name:    "missing repo",
			config:  BenchmarkConfig{Language: "go", Exemplar: ExemplarConfig{Ref: "v1"}, ExpectedMarkers: 1, ScanCommand: "scan"},
			wantErr: "exemplar.repo is required",
		},
		{
			name:    "missing ref",
			config:  BenchmarkConfig{Language: "go", Exemplar: ExemplarConfig{Repo: "a/b"}, ExpectedMarkers: 1, ScanCommand: "scan"},
			wantErr: "exemplar.ref is required",
		},
		{
			name:    "zero expected markers",
			config:  BenchmarkConfig{Language: "go", Exemplar: ExemplarConfig{Repo: "a/b", Ref: "v1"}, ExpectedMarkers: 0, ScanCommand: "scan"},
			wantErr: "expected_markers must be positive",
		},
		{
			name:    "missing scan command",
			config:  BenchmarkConfig{Language: "go", Exemplar: ExemplarConfig{Repo: "a/b", Ref: "v1"}, ExpectedMarkers: 1},
			wantErr: "scan_command is required",
		},
		{
			name:   "defaults applied for clone_depth",
			config: BenchmarkConfig{Language: "go", Exemplar: ExemplarConfig{Repo: "a/b", Ref: "v1"}, ExpectedMarkers: 1, ScanCommand: "scan"},
		},
		{
			name:   "defaults applied for timeout",
			config: BenchmarkConfig{Language: "go", Exemplar: ExemplarConfig{Repo: "a/b", Ref: "v1"}, ExpectedMarkers: 1, ScanCommand: "scan"},
		},
		{
			name:   "valid full config",
			config: BenchmarkConfig{Language: "go", Exemplar: ExemplarConfig{Repo: "a/b", Ref: "v1"}, ExpectedMarkers: 25, ScanCommand: "scan", CloneDepth: 1, TimeoutMinutes: 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Validate() expected error containing %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("Validate() error = %q, want containing %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Validate() unexpected error: %v", err)
			}
			if tt.config.CloneDepth <= 0 {
				t.Error("CloneDepth should have been defaulted")
			}
			if tt.config.TimeoutMinutes <= 0 {
				t.Error("TimeoutMinutes should have been defaulted")
			}
		})
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "go.yaml")
		data := []byte(`
language: go
exemplar:
  repo: cli/cli
  ref: v2.60.0
expected_markers: 25
scan_command: rtmx from-tests --format json .
`)
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
		cfg, err := LoadConfig(path)
		if err != nil {
			t.Fatalf("LoadConfig() error = %v", err)
		}
		if cfg.Language != "go" {
			t.Errorf("Language = %q, want %q", cfg.Language, "go")
		}
	})

	t.Run("missing file", func(t *testing.T) {
		_, err := LoadConfig("/nonexistent/path.yaml")
		if err == nil {
			t.Fatal("LoadConfig() expected error for missing file")
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "bad.yaml")
		if err := os.WriteFile(path, []byte("{{invalid"), 0o644); err != nil {
			t.Fatal(err)
		}
		_, err := LoadConfig(path)
		if err == nil {
			t.Fatal("LoadConfig() expected error for invalid YAML")
		}
	})
}
