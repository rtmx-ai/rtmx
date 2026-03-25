package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestSecurityPolicy validates that SECURITY.md and install script exist
// and meet the acceptance criteria.
// REQ-REL-006: Security policy and install script
func TestSecurityPolicy(t *testing.T) {
	rtmx.Req(t, "REQ-REL-006")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	// AC1: SECURITY.md exists in repository root
	t.Run("security_md_exists", func(t *testing.T) {
		secPath := filepath.Join(projectRoot, "SECURITY.md")
		content, err := os.ReadFile(secPath)
		if err != nil {
			t.Fatal("SECURITY.md must exist in repository root")
		}
		sec := string(content)
		if !strings.Contains(sec, "Reporting a Vulnerability") {
			t.Error("SECURITY.md must include vulnerability reporting instructions")
		}
		if !strings.Contains(sec, "security@rtmx.ai") {
			t.Error("SECURITY.md must include security contact email")
		}
		if !strings.Contains(sec, "gpg --verify") {
			t.Error("SECURITY.md must include GPG verification instructions")
		}
	})

	// AC2: Install script works on Linux and macOS (amd64, arm64)
	t.Run("install_script_platform_support", func(t *testing.T) {
		scriptPath := filepath.Join(projectRoot, "scripts", "install.sh")
		content, err := os.ReadFile(scriptPath)
		if err != nil {
			t.Fatal("scripts/install.sh must exist")
		}
		script := string(content)
		for _, platform := range []string{"linux", "darwin"} {
			if !strings.Contains(script, platform) {
				t.Errorf("Install script must support %s", platform)
			}
		}
		for _, arch := range []string{"amd64", "arm64"} {
			if !strings.Contains(script, arch) {
				t.Errorf("Install script must support %s", arch)
			}
		}
	})

	// AC3: Install script verifies checksums
	t.Run("install_script_checksum_verification", func(t *testing.T) {
		scriptPath := filepath.Join(projectRoot, "scripts", "install.sh")
		content, err := os.ReadFile(scriptPath)
		if err != nil {
			t.Fatal("scripts/install.sh must exist")
		}
		script := string(content)
		if !strings.Contains(script, "checksums.txt") {
			t.Error("Install script must download checksums")
		}
		if !strings.Contains(script, "sha256sum") {
			t.Error("Install script must verify SHA256 checksums")
		}
		if !strings.Contains(script, "Checksum mismatch") {
			t.Error("Install script must detect checksum mismatches")
		}
	})

	// AC4: Install script provides clear error on unsupported platforms
	t.Run("install_script_unsupported_platform_error", func(t *testing.T) {
		scriptPath := filepath.Join(projectRoot, "scripts", "install.sh")
		content, err := os.ReadFile(scriptPath)
		if err != nil {
			t.Fatal("scripts/install.sh must exist")
		}
		script := string(content)
		if !strings.Contains(script, "Unsupported OS") {
			t.Error("Install script must error on unsupported OS")
		}
		if !strings.Contains(script, "Unsupported architecture") {
			t.Error("Install script must error on unsupported architecture")
		}
	})
}
