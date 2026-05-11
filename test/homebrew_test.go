package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestHomebrewTap validates that the Homebrew tap is configured and
// the GoReleaser formula generates correctly.
// REQ-GO-044: Go CLI shall be installable via Homebrew tap
func TestHomebrewTap(t *testing.T) {
	rtmx.Req(t, "REQ-GO-044")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	goreleaserPath := filepath.Join(projectRoot, ".goreleaser.yaml")
	content, err := os.ReadFile(goreleaserPath)
	if err != nil {
		t.Fatalf("Failed to read .goreleaser.yaml: %v", err)
	}
	gr := string(content)

	// AC1: GoReleaser has brews section targeting rtmx-ai/homebrew-tap
	t.Run("homebrew_tap_configured", func(t *testing.T) {
		if !strings.Contains(gr, "brews:") {
			t.Fatal("GoReleaser must have brews section for Homebrew tap")
		}
		if !strings.Contains(gr, "homebrew-tap") {
			t.Error("Brews section must reference homebrew-tap repository")
		}
	})

	// AC2: Formula produces correct binary name
	t.Run("formula_binary_name", func(t *testing.T) {
		if !strings.Contains(gr, "bin.install") {
			t.Error("Formula must include bin.install for binary")
		}
	})

	// AC3: Formula has test stanza
	t.Run("formula_test_stanza", func(t *testing.T) {
		if !strings.Contains(gr, `system "#{bin}/rtmx"`) {
			t.Error("Formula must include test stanza that runs rtmx")
		}
	})

	// AC4: README documents brew install
	t.Run("readme_install_documented", func(t *testing.T) {
		readme, err := os.ReadFile(filepath.Join(projectRoot, "README.md"))
		if err != nil {
			t.Fatal("README.md must exist")
		}
		if !strings.Contains(string(readme), "brew install rtmx-ai/tap/rtmx") {
			t.Error("README must document brew install command")
		}
	})
}
