package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"github.com/spf13/cobra"
)

// createSecurityTestCmd creates a root command with the security subcommand for testing.
func createSecurityTestCmd(jsonFlag, strictFlag bool) *cobra.Command {
	root := &cobra.Command{
		Use:           "rtmx",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	var jsonOut, strict bool

	secCmd := &cobra.Command{
		Use:   "security",
		Short: "Audit security posture",
		RunE: func(cmd *cobra.Command, args []string) error {
			securityJSON = jsonOut
			securityStrict = strict
			return runSecurity(cmd, args)
		},
	}
	secCmd.Flags().BoolVar(&jsonOut, "json", jsonFlag, "output as JSON")
	secCmd.Flags().BoolVar(&strict, "strict", strictFlag, "treat warnings as failures")
	root.AddCommand(secCmd)

	return root
}

// setupSecurityTestProject creates a temp directory with the given files.
// files is a map from relative path to content.
func setupSecurityTestProject(t *testing.T, files map[string]string) string {
	t.Helper()
	tmpDir := t.TempDir()
	for relPath, content := range files {
		absPath := filepath.Join(tmpDir, relPath)
		if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
			t.Fatalf("failed to create dir for %s: %v", relPath, err)
		}
		if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", relPath, err)
		}
	}
	return tmpDir
}

func TestSecurityCheckVerifyThresholds(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	tests := []struct {
		name       string
		config     string
		wantStatus CheckStatus
	}{
		{
			name: "thresholds configured under rtmx key",
			config: `rtmx:
  verify:
    thresholds:
      warn: 5
      fail: 15
`,
			wantStatus: CheckPass,
		},
		{
			name: "thresholds configured at top level",
			config: `verify:
  thresholds:
    warn: 10
    fail: 20
`,
			wantStatus: CheckPass,
		},
		{
			name:       "no thresholds configured",
			config:     "rtmx:\n  database: .rtmx/database.csv\n",
			wantStatus: CheckWarn,
		},
		{
			name: "partial thresholds (only warn)",
			config: `rtmx:
  verify:
    thresholds:
      warn: 5
`,
			wantStatus: CheckWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := setupSecurityTestProject(t, map[string]string{
				"rtmx.yaml": tt.config,
			})

			opts := &SecurityOptions{Dir: tmpDir}
			ghFalse := false
			opts.GhAvailable = &ghFalse
			result := runSecurityChecks(opts)

			found := false
			for _, check := range result.Checks {
				if check.Name == "verify_thresholds" {
					found = true
					if check.Status != tt.wantStatus {
						t.Errorf("verify_thresholds: got status %s, want %s", check.Status, tt.wantStatus)
					}
				}
			}
			if !found {
				t.Error("verify_thresholds check not found in results")
			}
		})
	}
}

func TestSecurityCheckCIVerifyJob(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	tests := []struct {
		name       string
		ciContent  string
		wantStatus CheckStatus
	}{
		{
			name: "verify-requirements job exists",
			ciContent: `name: CI
jobs:
  verify-requirements:
    runs-on: ubuntu-latest
    steps:
      - run: rtmx verify
`,
			wantStatus: CheckPass,
		},
		{
			name: "rtmx verify in steps",
			ciContent: `name: CI
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: rtmx verify --update
`,
			wantStatus: CheckPass,
		},
		{
			name: "rtmx health in CI",
			ciContent: `name: CI
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: rtmx health --json
`,
			wantStatus: CheckPass,
		},
		{
			name: "no verify job",
			ciContent: `name: CI
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: go test ./...
`,
			wantStatus: CheckWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := setupSecurityTestProject(t, map[string]string{
				"rtmx.yaml":                    "rtmx:\n  database: db.csv\n",
				".github/workflows/ci.yml":     tt.ciContent,
			})

			ghFalse := false
			opts := &SecurityOptions{Dir: tmpDir, GhAvailable: &ghFalse}
			result := runSecurityChecks(opts)

			found := false
			for _, check := range result.Checks {
				if check.Name == "ci_verify_job" {
					found = true
					if check.Status != tt.wantStatus {
						t.Errorf("ci_verify_job: got status %s, want %s", check.Status, tt.wantStatus)
					}
				}
			}
			if !found {
				t.Error("ci_verify_job check not found in results")
			}
		})
	}
}

func TestSecurityCheckCIVerifyJobMissing(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	tmpDir := setupSecurityTestProject(t, map[string]string{
		"rtmx.yaml": "rtmx:\n  database: db.csv\n",
	})

	ghFalse := false
	opts := &SecurityOptions{Dir: tmpDir, GhAvailable: &ghFalse}
	result := runSecurityChecks(opts)

	for _, check := range result.Checks {
		if check.Name == "ci_verify_job" {
			if check.Status != CheckWarn {
				t.Errorf("ci_verify_job: got status %s, want WARN when no CI file", check.Status)
			}
			return
		}
	}
	t.Error("ci_verify_job check not found in results")
}

func TestSecurityCheckActionsPinned(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	tests := []struct {
		name       string
		workflows  map[string]string
		wantStatus CheckStatus
		wantSubstr string
	}{
		{
			name: "all pinned",
			workflows: map[string]string{
				"ci.yml": `name: CI
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@a5ac7e51b41094c92402da3b24376905380afc29
      - uses: actions/setup-go@cdcb36043654635271a94b9a6d1392de5bb323a7
`,
			},
			wantStatus: CheckPass,
			wantSubstr: "2/2",
		},
		{
			name: "none pinned",
			workflows: map[string]string{
				"ci.yml": `name: CI
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
`,
			},
			wantStatus: CheckWarn,
			wantSubstr: "0/2",
		},
		{
			name: "mixed pinning",
			workflows: map[string]string{
				"ci.yml": `name: CI
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@a5ac7e51b41094c92402da3b24376905380afc29
      - uses: actions/setup-go@v5
`,
			},
			wantStatus: CheckWarn,
			wantSubstr: "1/2",
		},
		{
			name: "local actions only",
			workflows: map[string]string{
				"ci.yml": `name: CI
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: ./my-action
`,
			},
			wantStatus: CheckPass,
			wantSubstr: "No external actions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{
				"rtmx.yaml": "rtmx:\n  database: db.csv\n",
			}
			for name, content := range tt.workflows {
				files[filepath.Join(".github", "workflows", name)] = content
			}
			tmpDir := setupSecurityTestProject(t, files)

			ghFalse := false
			opts := &SecurityOptions{Dir: tmpDir, GhAvailable: &ghFalse}
			result := runSecurityChecks(opts)

			for _, check := range result.Checks {
				if check.Name == "actions_pinned" {
					if check.Status != tt.wantStatus {
						t.Errorf("actions_pinned: got status %s, want %s", check.Status, tt.wantStatus)
					}
					if tt.wantSubstr != "" && !strings.Contains(check.Message, tt.wantSubstr) {
						t.Errorf("actions_pinned message %q does not contain %q", check.Message, tt.wantSubstr)
					}
					return
				}
			}
			t.Error("actions_pinned check not found in results")
		})
	}
}

func TestSecurityCheckGPGSigning(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	tests := []struct {
		name           string
		releaseContent string
		wantStatus     CheckStatus
	}{
		{
			name: "GPG signing present",
			releaseContent: `name: Release
jobs:
  release:
    steps:
      - name: Import GPG key
        run: echo "$GPG_PRIVATE_KEY" | gpg --import
`,
			wantStatus: CheckPass,
		},
		{
			name: "cosign signing present",
			releaseContent: `name: Release
jobs:
  release:
    steps:
      - name: Sign with cosign
        run: cosign sign
`,
			wantStatus: CheckPass,
		},
		{
			name: "no signing",
			releaseContent: `name: Release
jobs:
  release:
    steps:
      - run: goreleaser release
`,
			wantStatus: CheckWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := setupSecurityTestProject(t, map[string]string{
				"rtmx.yaml":                        "rtmx:\n  database: db.csv\n",
				".github/workflows/release.yml":     tt.releaseContent,
			})

			ghFalse := false
			opts := &SecurityOptions{Dir: tmpDir, GhAvailable: &ghFalse}
			result := runSecurityChecks(opts)

			for _, check := range result.Checks {
				if check.Name == "gpg_signing" {
					if check.Status != tt.wantStatus {
						t.Errorf("gpg_signing: got status %s, want %s", check.Status, tt.wantStatus)
					}
					return
				}
			}
			t.Error("gpg_signing check not found in results")
		})
	}
}

func TestSecurityCheckGPGSigningMissingRelease(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	tmpDir := setupSecurityTestProject(t, map[string]string{
		"rtmx.yaml": "rtmx:\n  database: db.csv\n",
	})

	ghFalse := false
	opts := &SecurityOptions{Dir: tmpDir, GhAvailable: &ghFalse}
	result := runSecurityChecks(opts)

	for _, check := range result.Checks {
		if check.Name == "gpg_signing" {
			if check.Status != CheckWarn {
				t.Errorf("gpg_signing: got status %s, want WARN when release workflow missing", check.Status)
			}
			return
		}
	}
	t.Error("gpg_signing check not found in results")
}

func TestSecurityCheckInstallGPG(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	tests := []struct {
		name       string
		script     string
		wantStatus CheckStatus
	}{
		{
			name:       "gpg verification present",
			script:     "#!/bin/bash\ngpg --verify checksums.txt.sig checksums.txt\n",
			wantStatus: CheckPass,
		},
		{
			name:       "checksums verification",
			script:     "#!/bin/bash\nsha256sum --check checksums\n",
			wantStatus: CheckPass,
		},
		{
			name:       "no verification",
			script:     "#!/bin/bash\ncurl -L https://example.com/rtmx | tar xz\n",
			wantStatus: CheckWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := setupSecurityTestProject(t, map[string]string{
				"rtmx.yaml":         "rtmx:\n  database: db.csv\n",
				"scripts/install.sh": tt.script,
			})

			ghFalse := false
			opts := &SecurityOptions{Dir: tmpDir, GhAvailable: &ghFalse}
			result := runSecurityChecks(opts)

			for _, check := range result.Checks {
				if check.Name == "install_gpg" {
					if check.Status != tt.wantStatus {
						t.Errorf("install_gpg: got status %s, want %s", check.Status, tt.wantStatus)
					}
					return
				}
			}
			t.Error("install_gpg check not found in results")
		})
	}
}

func TestSecurityCheckTrustedPeers(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	tests := []struct {
		name       string
		config     string
		wantStatus CheckStatus
	}{
		{
			name: "trusted peers configured",
			config: `rtmx:
  sync:
    trusted_peers:
      - name: alice
        public_key: "ssh-ed25519 AAAA..."
`,
			wantStatus: CheckPass,
		},
		{
			name: "empty trusted peers list",
			config: `rtmx:
  sync:
    trusted_peers: []
`,
			wantStatus: CheckWarn,
		},
		{
			name:       "no sync config",
			config:     "rtmx:\n  database: db.csv\n",
			wantStatus: CheckWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := setupSecurityTestProject(t, map[string]string{
				"rtmx.yaml": tt.config,
			})

			ghFalse := false
			opts := &SecurityOptions{Dir: tmpDir, GhAvailable: &ghFalse}
			result := runSecurityChecks(opts)

			for _, check := range result.Checks {
				if check.Name == "trusted_peers" {
					if check.Status != tt.wantStatus {
						t.Errorf("trusted_peers: got status %s, want %s", check.Status, tt.wantStatus)
					}
					return
				}
			}
			t.Error("trusted_peers check not found in results")
		})
	}
}

func TestSecurityCheckCodeowners(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	tests := []struct {
		name       string
		files      map[string]string
		wantStatus CheckStatus
		wantSubstr string
	}{
		{
			name: "CODEOWNERS with .rtmx coverage",
			files: map[string]string{
				".github/CODEOWNERS": "/.rtmx/ @rtmx-ai/core\n",
			},
			wantStatus: CheckPass,
			wantSubstr: ".rtmx/ coverage",
		},
		{
			name: "CODEOWNERS without .rtmx",
			files: map[string]string{
				".github/CODEOWNERS": "* @rtmx-ai/team\n",
			},
			wantStatus: CheckWarn,
			wantSubstr: "missing .rtmx/",
		},
		{
			name:       "no CODEOWNERS",
			files:      map[string]string{},
			wantStatus: CheckWarn,
			wantSubstr: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := map[string]string{
				"rtmx.yaml": "rtmx:\n  database: db.csv\n",
			}
			for k, v := range tt.files {
				files[k] = v
			}
			tmpDir := setupSecurityTestProject(t, files)

			ghTrue := true
			opts := &SecurityOptions{
				Dir:         tmpDir,
				GhAvailable: &ghTrue,
				GhRunner: func(args ...string) (string, error) {
					if len(args) >= 2 && args[0] == "repo" && args[1] == "view" {
						return "test-org/test-repo\n", nil
					}
					if len(args) >= 2 && args[0] == "api" {
						return "2", nil // mock branch protection rules
					}
					return "", fmt.Errorf("unexpected gh call: %v", args)
				},
			}
			result := runSecurityChecks(opts)

			for _, check := range result.Checks {
				if check.Name == "codeowners" {
					if check.Status != tt.wantStatus {
						t.Errorf("codeowners: got status %s, want %s", check.Status, tt.wantStatus)
					}
					if tt.wantSubstr != "" && !strings.Contains(check.Message, tt.wantSubstr) {
						t.Errorf("codeowners message %q does not contain %q", check.Message, tt.wantSubstr)
					}
					return
				}
			}
			t.Error("codeowners check not found in results")
		})
	}
}

func TestSecurityCheckBranchProtection(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	tests := []struct {
		name       string
		ghOutput   string
		ghErr      error
		wantStatus CheckStatus
	}{
		{
			name:       "protection enabled",
			ghOutput:   "3",
			wantStatus: CheckPass,
		},
		{
			name:       "no rules",
			ghOutput:   "0",
			wantStatus: CheckWarn,
		},
		{
			name:       "API error",
			ghOutput:   "",
			ghErr:      fmt.Errorf("permission denied"),
			wantStatus: CheckWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := setupSecurityTestProject(t, map[string]string{
				"rtmx.yaml": "rtmx:\n  database: db.csv\n",
			})

			ghTrue := true
			opts := &SecurityOptions{
				Dir:         tmpDir,
				GhAvailable: &ghTrue,
				GhRunner: func(args ...string) (string, error) {
					if len(args) >= 2 && args[0] == "repo" && args[1] == "view" {
						return "test-org/test-repo\n", nil
					}
					if len(args) >= 2 && args[0] == "api" {
						return tt.ghOutput, tt.ghErr
					}
					return "", fmt.Errorf("unexpected gh call: %v", args)
				},
			}
			result := runSecurityChecks(opts)

			for _, check := range result.Checks {
				if check.Name == "branch_protection" {
					if check.Status != tt.wantStatus {
						t.Errorf("branch_protection: got status %s, want %s (msg: %s)", check.Status, tt.wantStatus, check.Message)
					}
					return
				}
			}
			t.Error("branch_protection check not found in results")
		})
	}
}

func TestSecurityRepoChecksSkipWhenNoGh(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	tmpDir := setupSecurityTestProject(t, map[string]string{
		"rtmx.yaml": "rtmx:\n  database: db.csv\n",
	})

	ghFalse := false
	opts := &SecurityOptions{Dir: tmpDir, GhAvailable: &ghFalse}
	result := runSecurityChecks(opts)

	skipCount := 0
	for _, check := range result.Checks {
		if check.Category == "repository" && check.Status == CheckSkip {
			skipCount++
		}
	}
	if skipCount != 2 {
		t.Errorf("expected 2 skipped repository checks when gh unavailable, got %d", skipCount)
	}
}

func TestSecurityJSONOutput(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	tmpDir := setupSecurityTestProject(t, map[string]string{
		"rtmx.yaml": `rtmx:
  verify:
    thresholds:
      warn: 5
      fail: 15
`,
		".github/workflows/ci.yml": "name: CI\njobs:\n  verify-requirements:\n    runs-on: ubuntu-latest\n",
	})

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	root := createSecurityTestCmd(true, false)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"security", "--json"})

	err := root.Execute()
	if err != nil {
		var exitErr *ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("security --json failed unexpectedly: %v", err)
		}
	}

	out := strings.TrimSpace(buf.String())
	var result SecurityResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput: %s", err, out)
	}

	// Verify JSON structure
	if len(result.Checks) == 0 {
		t.Error("expected non-empty checks array")
	}

	// Verify each check has required fields
	for _, check := range result.Checks {
		if check.Category == "" {
			t.Error("check missing category field")
		}
		if check.Name == "" {
			t.Error("check missing name field")
		}
		if check.Status == "" {
			t.Error("check missing status field")
		}
		if check.Message == "" {
			t.Error("check missing message field")
		}
	}

	// Verify summary (SKIP checks are not counted in passed/warnings/failed)
	skipCount := 0
	for _, check := range result.Checks {
		if check.Status == CheckSkip {
			skipCount++
		}
	}
	totalCounted := result.Summary.Passed + result.Summary.Warnings + result.Summary.Failed
	expectedCounted := len(result.Checks) - skipCount
	if totalCounted != expectedCounted {
		t.Errorf("summary counts (%d) do not match non-skip checks (%d)", totalCounted, expectedCounted)
	}
}

func TestSecurityJSONSchema(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	tmpDir := setupSecurityTestProject(t, map[string]string{
		"rtmx.yaml": "rtmx:\n  database: db.csv\n",
	})

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	root := createSecurityTestCmd(true, false)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"security", "--json"})

	err := root.Execute()
	if err != nil {
		var exitErr *ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("security --json failed: %v", err)
		}
	}

	// Validate top-level JSON keys
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &raw); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	requiredKeys := []string{"platform", "repository", "checks", "summary"}
	for _, key := range requiredKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("JSON output missing required key %q", key)
		}
	}

	// Validate summary keys
	summary, ok := raw["summary"].(map[string]interface{})
	if !ok {
		t.Fatal("summary is not an object")
	}
	summaryKeys := []string{"passed", "warnings", "failed", "score_percent"}
	for _, key := range summaryKeys {
		if _, ok := summary[key]; !ok {
			t.Errorf("summary missing required key %q", key)
		}
	}
}

func TestSecurityStrictFlag(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	// Create project with some warnings (no trusted peers, etc.)
	tmpDir := setupSecurityTestProject(t, map[string]string{
		"rtmx.yaml": "rtmx:\n  database: db.csv\n",
	})

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	// Without --strict: warnings are OK, should not exit with error
	// (unless there are FAILs, but we have none here -- just WARNs)
	t.Run("without strict warnings do not cause exit error", func(t *testing.T) {
		root := createSecurityTestCmd(true, false)
		buf := new(bytes.Buffer)
		root.SetOut(buf)
		root.SetArgs([]string{"security", "--json"})

		err := root.Execute()
		// Parse result to check there are warnings
		var result SecurityResult
		_ = json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &result)

		if result.Summary.Warnings == 0 {
			t.Skip("no warnings in test setup, cannot verify strict behavior")
		}

		if err != nil {
			var exitErr *ExitError
			if errors.As(err, &exitErr) {
				t.Errorf("without --strict, warnings should not cause exit error, got exit code %d", exitErr.Code)
			}
		}
	})

	t.Run("with strict warnings cause exit error", func(t *testing.T) {
		root := createSecurityTestCmd(true, true)
		buf := new(bytes.Buffer)
		root.SetOut(buf)
		root.SetArgs([]string{"security", "--json", "--strict"})

		err := root.Execute()
		// Parse to verify warnings exist
		var result SecurityResult
		_ = json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &result)

		if result.Summary.Warnings == 0 {
			t.Skip("no warnings in test setup, cannot verify strict behavior")
		}

		if err == nil {
			t.Error("with --strict, warnings should cause exit error")
		} else {
			var exitErr *ExitError
			if !errors.As(err, &exitErr) {
				t.Errorf("expected ExitError, got %T: %v", err, err)
			} else if exitErr.Code != 1 {
				t.Errorf("expected exit code 1, got %d", exitErr.Code)
			}
		}
	})
}

func TestSecurityTextOutput(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	tmpDir := setupSecurityTestProject(t, map[string]string{
		"rtmx.yaml": `rtmx:
  verify:
    thresholds:
      warn: 5
      fail: 15
`,
		".github/workflows/ci.yml":     "name: CI\njobs:\n  verify-requirements:\n    runs-on: ubuntu-latest\n",
		".github/workflows/release.yml": "name: Release\njobs:\n  release:\n    steps:\n      - run: gpg --verify\n",
		"scripts/install.sh":            "#!/bin/bash\ngpg --verify checksums.txt.sig\n",
	})

	origDir, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origDir) }()

	root := createSecurityTestCmd(false, false)
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"security"})

	err := root.Execute()
	if err != nil {
		var exitErr *ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("security failed unexpectedly: %v", err)
		}
	}

	out := buf.String()

	expectedPhrases := []string{
		"Security Posture Check",
		"RTMX Controls",
		"[PASS]",
		"Score:",
		"passed",
		"warnings",
		"failures",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(out, phrase) {
			t.Errorf("expected output to contain %q, got:\n%s", phrase, out)
		}
	}
}

func TestSecuritySummaryCalculation(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	// Set up a project where all RTMX checks pass
	tmpDir := setupSecurityTestProject(t, map[string]string{
		"rtmx.yaml": `rtmx:
  verify:
    thresholds:
      warn: 5
      fail: 15
  sync:
    trusted_peers:
      - name: alice
        key: "ssh-ed25519 AAAA"
`,
		".github/workflows/ci.yml":      "name: CI\njobs:\n  verify-requirements:\n    runs-on: ubuntu-latest\n",
		".github/workflows/release.yml": "name: Release\njobs:\n  release:\n    steps:\n      - run: gpg --verify\n",
		"scripts/install.sh":            "#!/bin/bash\ngpg --verify checksums.txt.sig\n",
	})

	ghFalse := false
	opts := &SecurityOptions{Dir: tmpDir, GhAvailable: &ghFalse}
	result := runSecurityChecks(opts)

	// All 6 RTMX checks should pass, 2 repo checks skipped
	rtmxPassed := 0
	for _, check := range result.Checks {
		if check.Category == "rtmx" && check.Status == CheckPass {
			rtmxPassed++
		}
	}
	if rtmxPassed != 6 {
		t.Errorf("expected 6 RTMX checks to pass, got %d", rtmxPassed)
		for _, check := range result.Checks {
			t.Logf("  %s/%s: %s - %s", check.Category, check.Name, check.Status, check.Message)
		}
	}

	// Score should reflect passed checks
	if result.Summary.Passed < 6 {
		t.Errorf("expected at least 6 passed checks, got %d", result.Summary.Passed)
	}
}

func TestSecurityExitCodeOnFail(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	// Verify that the exit code calculation works correctly
	result := &SecurityResult{
		Checks: []SecurityCheck{
			{Status: CheckPass},
			{Status: CheckFail},
		},
	}
	result.Summary.Passed = 1
	result.Summary.Failed = 1

	// Without strict
	securityStrict = false
	code := securityExitCode(result)
	if code != 1 {
		t.Errorf("expected exit code 1 on FAIL, got %d", code)
	}

	// All pass
	result2 := &SecurityResult{
		Checks: []SecurityCheck{
			{Status: CheckPass},
		},
	}
	result2.Summary.Passed = 1
	code = securityExitCode(result2)
	if code != 0 {
		t.Errorf("expected exit code 0 on all PASS, got %d", code)
	}

	// Warnings only, no strict
	result3 := &SecurityResult{
		Checks: []SecurityCheck{
			{Status: CheckWarn},
		},
	}
	result3.Summary.Warnings = 1
	securityStrict = false
	code = securityExitCode(result3)
	if code != 0 {
		t.Errorf("expected exit code 0 for warnings without --strict, got %d", code)
	}

	// Warnings with strict
	securityStrict = true
	code = securityExitCode(result3)
	if code != 1 {
		t.Errorf("expected exit code 1 for warnings with --strict, got %d", code)
	}

	// Reset
	securityStrict = false
}

func TestSecurityAllChecksPresent(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	tmpDir := setupSecurityTestProject(t, map[string]string{
		"rtmx.yaml":                     "rtmx:\n  database: db.csv\n",
		".github/workflows/ci.yml":      "name: CI\njobs:\n  test:\n    runs-on: ubuntu-latest\n",
		".github/workflows/release.yml": "name: Release\njobs:\n  release:\n    steps:\n      - run: echo\n",
		"scripts/install.sh":            "#!/bin/bash\necho hello\n",
	})

	ghFalse := false
	opts := &SecurityOptions{Dir: tmpDir, GhAvailable: &ghFalse}
	result := runSecurityChecks(opts)

	expectedRTMXChecks := []string{
		"verify_thresholds",
		"ci_verify_job",
		"actions_pinned",
		"gpg_signing",
		"install_gpg",
		"trusted_peers",
	}

	expectedRepoChecks := []string{
		"branch_protection",
		"codeowners",
	}

	checkNames := make(map[string]bool)
	for _, check := range result.Checks {
		checkNames[check.Name] = true
	}

	for _, name := range expectedRTMXChecks {
		if !checkNames[name] {
			t.Errorf("missing RTMX check: %s", name)
		}
	}

	for _, name := range expectedRepoChecks {
		if !checkNames[name] {
			t.Errorf("missing repository check: %s", name)
		}
	}
}

func TestSecurityRepoDetectionFailure(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	tmpDir := setupSecurityTestProject(t, map[string]string{
		"rtmx.yaml": "rtmx:\n  database: db.csv\n",
	})

	ghTrue := true
	opts := &SecurityOptions{
		Dir:         tmpDir,
		GhAvailable: &ghTrue,
		GhRunner: func(args ...string) (string, error) {
			return "", fmt.Errorf("not a git repository")
		},
	}
	result := runSecurityChecks(opts)

	// Should have a skip for branch_protection when repo detection fails
	for _, check := range result.Checks {
		if check.Name == "branch_protection" {
			if check.Status != CheckSkip {
				t.Errorf("expected branch_protection to be SKIP when repo detection fails, got %s", check.Status)
			}
			return
		}
	}
	t.Error("branch_protection check not found")
}

func TestSecurityNoWorkflowsDir(t *testing.T) {
	rtmx.Req(t, "REQ-SEC-012")

	// No .github/workflows at all
	tmpDir := setupSecurityTestProject(t, map[string]string{
		"rtmx.yaml": "rtmx:\n  database: db.csv\n",
	})

	ghFalse := false
	opts := &SecurityOptions{Dir: tmpDir, GhAvailable: &ghFalse}
	result := runSecurityChecks(opts)

	for _, check := range result.Checks {
		if check.Name == "actions_pinned" {
			if check.Status != CheckWarn {
				t.Errorf("actions_pinned: got status %s, want WARN when no workflows dir", check.Status)
			}
			return
		}
	}
	t.Error("actions_pinned check not found")
}
