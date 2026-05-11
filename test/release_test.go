package test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestV010Release validates the v0.1.0 release requirements.
// REQ-GO-073: Go CLI v0.1.0 release signals architectural transition from Python
func TestV010Release(t *testing.T) {
	rtmx.Req(t, "REQ-GO-073")
	// Build the Go CLI binary
	tmpDir, err := os.MkdirTemp("", "rtmx-release-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	binaryPath := filepath.Join(tmpDir, binaryName())

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/rtmx")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build Go CLI: %v\n%s", err, output)
	}

	// Test 1: Binary exists and is executable
	t.Run("binary_exists", func(t *testing.T) {
		info, err := os.Stat(binaryPath)
		if err != nil {
			t.Fatalf("Binary not found: %v", err)
		}
		if info.Size() == 0 {
			t.Error("Binary is empty")
		}
	})

	// Test 2: Version command works
	t.Run("version_command", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "version")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("version command failed: %v\n%s", err, output)
		}
		if !bytes.Contains(output, []byte("rtmx")) {
			t.Error("version output missing 'rtmx'")
		}
	})

	// Test 3: Help shows all required commands
	t.Run("help_shows_commands", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Help may exit non-zero in some cases
			_ = err
		}

		requiredCommands := []string{
			"status",
			"backlog",
			"health",
			"init",
			"verify",
			"from-tests",
			"from-pytest",
			"deps",
			"cycles",
		}

		for _, cmdName := range requiredCommands {
			if !bytes.Contains(output, []byte(cmdName)) {
				t.Errorf("Help missing required command: %s", cmdName)
			}
		}
	})

	// Test 4: Status command works with real project
	t.Run("status_with_project", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "status")
		cmd.Dir = projectRoot
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("status command failed: %v\n%s", err, output)
		}
		if !bytes.Contains(output, []byte("RTM Status Check")) {
			t.Error("status output missing expected header")
		}
	})

	// Test 5: Cross-platform compatibility (verified by this test running)
	t.Run("cross_platform", func(t *testing.T) {
		t.Logf("Running on %s/%s", runtime.GOOS, runtime.GOARCH)
		// If we got here, the binary works on this platform
	})

	// Test 6: from-go command exists (Go testing integration - REQ-LANG-003)
	t.Run("from_go_command", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "from-go", "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Help may exit non-zero
			_ = err
		}
		if !bytes.Contains(output, []byte("from-go")) {
			t.Error("from-go command not available")
		}
	})
}

// TestV010ReleaseNotes validates release documentation exists.
func TestV010ReleaseNotes(t *testing.T) {
	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	// Check for README
	readmePath := filepath.Join(projectRoot, "README.md")
	if _, err := os.Stat(readmePath); err != nil {
		t.Error("README.md not found")
	}

	// Check for LICENSE
	licensePath := filepath.Join(projectRoot, "LICENSE")
	if _, err := os.Stat(licensePath); err != nil {
		t.Error("LICENSE not found")
	}

	// Check GoReleaser config
	goreleaserPath := filepath.Join(projectRoot, ".goreleaser.yaml")
	content, err := os.ReadFile(goreleaserPath)
	if err != nil {
		t.Fatalf("Failed to read .goreleaser.yaml: %v", err)
	}

	// Verify it targets all required platforms
	requiredPlatforms := []string{"linux", "darwin", "windows"}
	for _, platform := range requiredPlatforms {
		if !strings.Contains(string(content), platform) {
			t.Errorf("GoReleaser missing platform: %s", platform)
		}
	}

	// Verify Homebrew tap is configured
	if !strings.Contains(string(content), "homebrew-tap") {
		t.Error("GoReleaser missing Homebrew tap configuration")
	}
}

// TestV1Release validates the v1.0.0 release gate.
// REQ-GO-047: Go CLI v1.0.0 stable release with multi-language support and full documentation.
// All 22 dependencies must be COMPLETE before this gate passes.
func TestV1Release(t *testing.T) {
	rtmx.Req(t, "REQ-GO-047")

	// Build the Go CLI binary
	tmpDir, err := os.MkdirTemp("", "rtmx-v1-release-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	binaryPath := filepath.Join(tmpDir, binaryName())

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	buildCmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/rtmx")
	buildCmd.Dir = projectRoot
	if output, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build Go CLI: %v\n%s", err, output)
	}

	// Test 1: Binary builds and is executable
	t.Run("binary_builds", func(t *testing.T) {
		info, err := os.Stat(binaryPath)
		if err != nil {
			t.Fatalf("Binary not found: %v", err)
		}
		if info.Size() == 0 {
			t.Error("Binary is empty")
		}
		t.Logf("Binary size: %d bytes on %s/%s", info.Size(), runtime.GOOS, runtime.GOARCH)
	})

	// Test 2: Version command works
	t.Run("version_command", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "version")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("version command failed: %v\n%s", err, output)
		}
		if !bytes.Contains(output, []byte("rtmx")) {
			t.Error("version output missing 'rtmx'")
		}
	})

	// Test 3: All required v1.0 commands exist
	t.Run("all_required_commands", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "--help")
		output, err := cmd.CombinedOutput()
		if err != nil {
			_ = err // help may exit non-zero
		}

		requiredCommands := []string{
			"status",
			"backlog",
			"health",
			"deps",
			"cycles",
			"verify",
			"from-go",
			"from-tests",
			"from-pytest",
			"reconcile",
			"context",
			"install",
			"diff",
			"docs",
			"config",
			"makefile",
			"analyze",
			"bootstrap",
			"setup",
		}

		for _, cmdName := range requiredCommands {
			if !bytes.Contains(output, []byte(cmdName)) {
				t.Errorf("Help missing required command: %s", cmdName)
			}
		}
	})

	// Test 4: JSON output works for status
	t.Run("status_json", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "status", "--json")
		cmd.Dir = projectRoot
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("status --json failed: %v\n%s", err, output)
		}
		var result map[string]interface{}
		if err := json.Unmarshal(output, &result); err != nil {
			t.Fatalf("status --json produced invalid JSON: %v\nOutput: %s", err, output)
		}
	})

	// Test 5: JSON output works for backlog
	t.Run("backlog_json", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "backlog", "--json")
		cmd.Dir = projectRoot
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("backlog --json failed: %v\n%s", err, output)
		}
		var result map[string]interface{}
		if err := json.Unmarshal(output, &result); err != nil {
			t.Fatalf("backlog --json produced invalid JSON: %v\nOutput: %s", err, output)
		}
	})

	// Test 6: JSON output works for health
	t.Run("health_json", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "health", "--json")
		cmd.Dir = projectRoot
		output, err := cmd.CombinedOutput()
		if err != nil {
			_ = err // health may exit non-zero on warnings
		}
		// Extract JSON object from output (may have extra text from error exit)
		raw := string(output)
		start := strings.Index(raw, "{")
		end := strings.LastIndex(raw, "}")
		if start < 0 || end < 0 || end <= start {
			t.Fatalf("health --json produced no JSON object\nOutput: %s", raw)
		}
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(raw[start:end+1]), &result); err != nil {
			t.Fatalf("health --json produced invalid JSON: %v\nOutput: %s", err, raw)
		}
	})

	// Test 7: --fail-under flag works (exit 0 when above threshold)
	t.Run("fail_under_passes", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "status", "--fail-under", "1")
		cmd.Dir = projectRoot
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("status --fail-under 1 should pass: %v\n%s", err, output)
		}
	})

	// Test 8: --fail-under flag works (exit 1 when below threshold)
	t.Run("fail_under_fails", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "status", "--fail-under", "100")
		cmd.Dir = projectRoot
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Error("status --fail-under 100 should fail but exited 0")
		}
		_ = output
	})

	// Test 9: GoReleaser config exists and is valid
	t.Run("goreleaser_config", func(t *testing.T) {
		goreleaserPath := filepath.Join(projectRoot, ".goreleaser.yaml")
		content, err := os.ReadFile(goreleaserPath)
		if err != nil {
			t.Fatalf("Failed to read .goreleaser.yaml: %v", err)
		}
		// Must target all platforms
		for _, platform := range []string{"linux", "darwin", "windows"} {
			if !strings.Contains(string(content), platform) {
				t.Errorf("GoReleaser missing platform: %s", platform)
			}
		}
		// Must have Homebrew tap
		if !strings.Contains(string(content), "homebrew-tap") {
			t.Error("GoReleaser missing Homebrew tap configuration")
		}
		// Must have Scoop bucket
		if !strings.Contains(string(content), "scoop") {
			t.Error("GoReleaser missing Scoop bucket configuration")
		}
	})

	// Test 10: CI workflow exists
	t.Run("ci_workflow", func(t *testing.T) {
		ciPath := filepath.Join(projectRoot, ".github/workflows/ci.yml")
		if _, err := os.Stat(ciPath); err != nil {
			t.Fatalf("CI workflow not found: %v", err)
		}
	})

	// Test 11: Coverage badge in README
	t.Run("coverage_badge_in_readme", func(t *testing.T) {
		readmePath := filepath.Join(projectRoot, "README.md")
		content, err := os.ReadFile(readmePath)
		if err != nil {
			t.Fatalf("Failed to read README.md: %v", err)
		}
		if !strings.Contains(string(content), "coveralls") && !strings.Contains(string(content), "coverage") {
			t.Error("README.md missing coverage badge")
		}
	})

	// Test 12: SECURITY.md exists
	t.Run("security_md", func(t *testing.T) {
		secPath := filepath.Join(projectRoot, "SECURITY.md")
		if _, err := os.Stat(secPath); err != nil {
			t.Fatalf("SECURITY.md not found: %v", err)
		}
	})

	// Test 13: Install script exists
	t.Run("install_script", func(t *testing.T) {
		installPath := filepath.Join(projectRoot, "scripts/install.sh")
		info, err := os.Stat(installPath)
		if err != nil {
			t.Fatalf("Install script not found: %v", err)
		}
		if info.Size() == 0 {
			t.Error("Install script is empty")
		}
	})

	// Test 14: GPG signing is configured in release workflow
	t.Run("gpg_signing_configured", func(t *testing.T) {
		releasePath := filepath.Join(projectRoot, ".github/workflows/release.yml")
		content, err := os.ReadFile(releasePath)
		if err != nil {
			// Fall back to goreleaser config
			goreleaserPath := filepath.Join(projectRoot, ".goreleaser.yaml")
			content, err = os.ReadFile(goreleaserPath)
			if err != nil {
				t.Fatalf("Neither release.yml nor .goreleaser.yaml found: %v", err)
			}
		}
		if !strings.Contains(string(content), "gpg") && !strings.Contains(string(content), "GPG") &&
			!strings.Contains(string(content), "sign") {
			t.Error("Release pipeline missing GPG signing configuration")
		}
	})

	// Test 15: All 22 dependency requirements are COMPLETE in database
	t.Run("all_dependencies_complete", func(t *testing.T) {
		dbPath := filepath.Join(projectRoot, ".rtmx/database.csv")
		content, err := os.ReadFile(dbPath)
		if err != nil {
			t.Fatalf("Failed to read database.csv: %v", err)
		}

		dependencies := []string{
			"REQ-AGENT-001",
			"REQ-AGENT-002",
			"REQ-CI-001",
			"REQ-CI-002",
			"REQ-E2E-001",
			"REQ-GO-048",
			"REQ-GO-049",
			"REQ-GO-050",
			"REQ-GO-051",
			"REQ-GO-052",
			"REQ-GO-053",
			"REQ-GO-054",
			"REQ-GO-068",
			"REQ-GO-073",
			"REQ-PAR-001",
			"REQ-PAR-002",
			"REQ-PAR-003",
			"REQ-PAR-004",
			"REQ-REL-001",
			"REQ-REL-002",
			"REQ-REL-003",
			"REQ-REL-004",
		}

		lines := strings.Split(string(content), "\n")
		for _, dep := range dependencies {
			found := false
			for _, line := range lines {
				if strings.HasPrefix(line, dep+",") {
					found = true
					if !strings.Contains(line, ",COMPLETE,") {
						t.Errorf("Dependency %s is not COMPLETE", dep)
					}
					break
				}
			}
			if !found {
				t.Errorf("Dependency %s not found in database.csv", dep)
			}
		}
	})
}
