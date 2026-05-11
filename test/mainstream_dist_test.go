package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestMainstreamDistribution validates that the infrastructure for
// mainstream package repository submissions is configured.
// REQ-DIST-005: RTMX shall be submitted to mainstream package repositories
func TestMainstreamDistribution(t *testing.T) {
	rtmx.Req(t, "REQ-DIST-005")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	// AC1: GoReleaser configured for cross-platform builds
	t.Run("cross_platform_builds", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".goreleaser.yaml"))
		if err != nil {
			t.Fatalf("Failed to read .goreleaser.yaml: %v", err)
		}
		gr := string(content)

		// Must support linux and darwin at minimum
		if !strings.Contains(gr, "linux") {
			t.Error("GoReleaser must target linux")
		}
		if !strings.Contains(gr, "darwin") {
			t.Error("GoReleaser must target darwin")
		}
		if !strings.Contains(gr, "amd64") {
			t.Error("GoReleaser must target amd64")
		}
		if !strings.Contains(gr, "arm64") {
			t.Error("GoReleaser must target arm64")
		}
	})

	// AC2: Homebrew tap configured (prerequisite for homebrew-core)
	t.Run("homebrew_tap_exists", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".goreleaser.yaml"))
		if err != nil {
			t.Fatalf("Failed to read .goreleaser.yaml: %v", err)
		}
		if !strings.Contains(string(content), "brews:") {
			t.Error("GoReleaser must configure Homebrew formula")
		}
	})

	// AC3: RPM/DEB packages configured
	t.Run("linux_packages", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".goreleaser.yaml"))
		if err != nil {
			t.Fatalf("Failed to read .goreleaser.yaml: %v", err)
		}
		gr := string(content)
		if !strings.Contains(gr, "nfpms:") {
			t.Error("GoReleaser must have nfpms section for .deb and .rpm packages")
		}
	})

	// AC4: Scoop bucket configured (Windows)
	t.Run("scoop_configured", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".goreleaser.yaml"))
		if err != nil {
			t.Fatalf("Failed to read .goreleaser.yaml: %v", err)
		}
		if !strings.Contains(string(content), "scoops:") {
			t.Error("GoReleaser must configure Scoop bucket for Windows distribution")
		}
	})

	// AC5: README documents multiple install methods
	t.Run("readme_install_methods", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "README.md"))
		if err != nil {
			t.Fatal("README.md must exist")
		}
		readme := string(content)
		methods := []string{"brew install", "scoop install", "go install"}
		documented := 0
		for _, m := range methods {
			if strings.Contains(readme, m) {
				documented++
			}
		}
		if documented < 2 {
			t.Errorf("README should document at least 2 install methods, found %d", documented)
		}
	})
}
