package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestNpmInstall validates that npm distribution infrastructure is configured.
// REQ-DIST-003: RTMX shall be installable via npm with Go binary bundled
func TestNpmInstall(t *testing.T) {
	rtmx.Req(t, "REQ-DIST-003")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	// AC1: npm package.json exists in npm/ directory
	t.Run("package_json_exists", func(t *testing.T) {
		path := filepath.Join(projectRoot, "npm", "package.json")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("npm/package.json must exist: %v", err)
		}
		pkg := string(content)
		if !strings.Contains(pkg, `"name"`) {
			t.Error("package.json must have a name field")
		}
		if !strings.Contains(pkg, "rtmx") {
			t.Error("package name must reference rtmx")
		}
	})

	// AC2: Install script handles platform detection
	t.Run("install_script_exists", func(t *testing.T) {
		path := filepath.Join(projectRoot, "npm", "install.js")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("npm/install.js must exist: %v", err)
		}
		script := string(content)
		if !strings.Contains(script, "platform") {
			t.Error("install script must handle platform detection")
		}
		if !strings.Contains(script, "arch") {
			t.Error("install script must handle architecture detection")
		}
	})

	// AC3: Binary wrapper exists
	t.Run("bin_wrapper_exists", func(t *testing.T) {
		// Check for either bin/rtmx or index.js wrapper
		binPath := filepath.Join(projectRoot, "npm", "bin", "rtmx")
		indexPath := filepath.Join(projectRoot, "npm", "index.js")
		_, binErr := os.Stat(binPath)
		_, indexErr := os.Stat(indexPath)
		if binErr != nil && indexErr != nil {
			t.Fatal("npm package must have bin/rtmx or index.js wrapper")
		}
	})

	// AC4: GoReleaser produces archives for npm consumption
	t.Run("goreleaser_archives", func(t *testing.T) {
		path := filepath.Join(projectRoot, ".goreleaser.yaml")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("Failed to read .goreleaser.yaml: %v", err)
		}
		gr := string(content)
		// GoReleaser must produce tar.gz archives that npm install.js can download
		if !strings.Contains(gr, "archives:") {
			t.Error("GoReleaser must have archives section for binary distribution")
		}
	})
}
