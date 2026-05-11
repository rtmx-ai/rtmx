package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestV1ReleaseGate validates that all prerequisites for v1.0.0 are met.
// REQ-REL-007: v1.0.0 release gate
func TestV1ReleaseGate(t *testing.T) {
	rtmx.Req(t, "REQ-REL-007")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	// AC1: All requirements are COMPLETE
	t.Run("all_requirements_complete", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".rtmx", "database.csv"))
		if err != nil {
			t.Fatalf("database.csv must exist: %v", err)
		}
		csv := string(content)
		lines := strings.Split(csv, "\n")
		partial := 0
		for _, line := range lines[1:] {
			if line == "" {
				continue
			}
			fields := strings.Split(line, ",")
			if len(fields) < 9 {
				continue
			}
			status := fields[8]
			if status == "PARTIAL" {
				partial++
			}
		}
		if partial > 0 {
			t.Errorf("%d requirements are PARTIAL, want 0", partial)
		}
	})

	// AC2: GoReleaser config produces all required artifact types
	t.Run("goreleaser_artifacts", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".goreleaser.yaml"))
		if err != nil {
			t.Fatalf("Failed to read .goreleaser.yaml: %v", err)
		}
		gr := string(content)
		required := []string{
			"archives:",    // binaries
			"nfpms:",       // .deb and .rpm
			"brews:",       // Homebrew formula
			"scoops:",      // Scoop manifest
			"signs:",       // GPG signatures
			"sboms:",       // SBOM
		}
		for _, section := range required {
			if !strings.Contains(gr, section) {
				t.Errorf("GoReleaser must have %s section", section)
			}
		}
	})

	// AC3: GPG signing configured
	t.Run("gpg_signing", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".goreleaser.yaml"))
		if err != nil {
			t.Fatalf("Failed to read .goreleaser.yaml: %v", err)
		}
		if !strings.Contains(string(content), "GPG_FINGERPRINT") {
			t.Error("GoReleaser must use GPG_FINGERPRINT for signing")
		}
	})

	// AC4: Release workflow exists
	t.Run("release_workflow", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".github", "workflows", "release.yml"))
		if err != nil {
			t.Fatalf("release.yml must exist: %v", err)
		}
		wf := string(content)
		if !strings.Contains(wf, "goreleaser") {
			t.Error("release workflow must use GoReleaser")
		}
	})

	// AC5: Version injection via ldflags
	t.Run("version_ldflags", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".goreleaser.yaml"))
		if err != nil {
			t.Fatalf("Failed to read .goreleaser.yaml: %v", err)
		}
		gr := string(content)
		if !strings.Contains(gr, "Version={{.Version}}") {
			t.Error("GoReleaser must inject version via ldflags")
		}
	})

	// AC6: Pre-tag hook or gate mechanism exists
	t.Run("release_gate_command", func(t *testing.T) {
		// Verify the release gate command infrastructure exists
		releasePath := filepath.Join(projectRoot, "internal", "cmd", "release.go")
		if _, err := os.Stat(releasePath); err != nil {
			t.Fatal("internal/cmd/release.go must exist for release gate")
		}
		content, err := os.ReadFile(releasePath)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(content), "gate") {
			t.Error("release command must support gate subcommand")
		}
	})
}
