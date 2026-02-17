package test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestV010Release validates the v0.1.0 release requirements.
// REQ-GO-073: Go CLI v0.1.0 release signals architectural transition from Python
func TestV010Release(t *testing.T) {
	// Build the Go CLI binary
	tmpDir, err := os.MkdirTemp("", "rtmx-release-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

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
