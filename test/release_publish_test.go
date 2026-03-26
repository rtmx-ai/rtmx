package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestHomebrewScoop validates that release infrastructure is configured
// to publish to the Homebrew tap and Scoop bucket repositories.
// REQ-REL-004: Release workflow shall publish to Homebrew tap and Scoop bucket
func TestHomebrewScoop(t *testing.T) {
	rtmx.Req(t, "REQ-REL-004")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	// Read goreleaser config
	goreleaserPath := filepath.Join(projectRoot, ".goreleaser.yaml")
	grContent, err := os.ReadFile(goreleaserPath)
	if err != nil {
		t.Fatalf("Failed to read .goreleaser.yaml: %v", err)
	}
	gr := string(grContent)

	// Read release workflow
	releasePath := filepath.Join(projectRoot, ".github", "workflows", "release.yml")
	relContent, err := os.ReadFile(releasePath)
	if err != nil {
		t.Fatalf("Failed to read release.yml: %v", err)
	}
	rel := string(relContent)

	// AC1: Homebrew tap configuration exists in GoReleaser
	t.Run("homebrew_tap_configured", func(t *testing.T) {
		if !strings.Contains(gr, "brews:") {
			t.Fatal("GoReleaser must have brews section")
		}
		if !strings.Contains(gr, "name: homebrew-tap") {
			t.Error("Homebrew tap must target rtmx-ai/homebrew-tap repository")
		}
		if !strings.Contains(gr, "owner: rtmx-ai") {
			t.Error("Homebrew tap must be owned by rtmx-ai")
		}
		if !strings.Contains(gr, "HOMEBREW_TAP_TOKEN") {
			t.Error("Homebrew tap must use HOMEBREW_TAP_TOKEN for authentication")
		}
		if !strings.Contains(gr, "directory: Formula") {
			t.Error("Homebrew formula must be placed in Formula directory")
		}
	})

	// AC2: Scoop bucket configuration exists in GoReleaser
	t.Run("scoop_bucket_configured", func(t *testing.T) {
		if !strings.Contains(gr, "scoops:") {
			t.Fatal("GoReleaser must have scoops section")
		}
		if !strings.Contains(gr, "name: scoop-bucket") {
			t.Error("Scoop bucket must target rtmx-ai/scoop-bucket repository")
		}
		if !strings.Contains(gr, "SCOOP_BUCKET_TOKEN") {
			t.Error("Scoop bucket must use SCOOP_BUCKET_TOKEN for authentication")
		}
	})

	// AC3: Release workflow passes both tokens to GoReleaser
	t.Run("workflow_passes_homebrew_token", func(t *testing.T) {
		if !strings.Contains(rel, "HOMEBREW_TAP_TOKEN") {
			t.Error("Release workflow must pass HOMEBREW_TAP_TOKEN to GoReleaser")
		}
		if !strings.Contains(rel, "secrets.HOMEBREW_TAP_TOKEN") {
			t.Error("Release workflow must source HOMEBREW_TAP_TOKEN from secrets")
		}
	})

	t.Run("workflow_passes_scoop_token", func(t *testing.T) {
		if !strings.Contains(rel, "SCOOP_BUCKET_TOKEN") {
			t.Error("Release workflow must pass SCOOP_BUCKET_TOKEN to GoReleaser")
		}
		if !strings.Contains(rel, "secrets.SCOOP_BUCKET_TOKEN") {
			t.Error("Release workflow must source SCOOP_BUCKET_TOKEN from secrets")
		}
	})

	// AC4: Homebrew formula includes install and test stanzas
	t.Run("homebrew_formula_stanzas", func(t *testing.T) {
		if !strings.Contains(gr, "bin.install") {
			t.Error("Homebrew formula must include install stanza with bin.install")
		}
		if !strings.Contains(gr, `system "#{bin}/rtmx"`) {
			t.Error("Homebrew formula must include test stanza that runs rtmx")
		}
	})

	// AC5: Both configs target the correct package name
	t.Run("package_name_rtmx", func(t *testing.T) {
		// Check the brews section has name: rtmx
		brewsIdx := strings.Index(gr, "brews:")
		if brewsIdx == -1 {
			t.Fatal("brews section not found")
		}
		brewSection := gr[brewsIdx:]
		scoopsIdx := strings.Index(brewSection, "scoops:")
		if scoopsIdx > 0 {
			brewSection = brewSection[:scoopsIdx]
		}
		if !strings.Contains(brewSection, "name: rtmx") {
			t.Error("Homebrew formula name must be 'rtmx'")
		}
	})

	// AC6: Release notes include installation instructions for both
	t.Run("release_notes_install_instructions", func(t *testing.T) {
		if !strings.Contains(gr, "brew install rtmx-ai/tap/rtmx") {
			t.Error("Release notes must include brew install command")
		}
		if !strings.Contains(gr, "scoop install rtmx") {
			t.Error("Release notes must include scoop install command")
		}
	})
}
