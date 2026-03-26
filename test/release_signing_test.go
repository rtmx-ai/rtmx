package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestGPGSigning validates that release infrastructure is configured for
// GPG signing of all binaries and checksums.
// REQ-REL-001: GPG binary signing
func TestGPGSigning(t *testing.T) {
	rtmx.Req(t, "REQ-REL-001")

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

	// AC1: All release archives have .sig GPG detached signatures
	t.Run("signs_all_artifacts", func(t *testing.T) {
		if !strings.Contains(gr, "signs:") {
			t.Fatal("GoReleaser must have signs section")
		}
		if !strings.Contains(gr, "artifacts: all") {
			t.Error("GoReleaser must sign all artifacts (artifacts: all)")
		}
		if !strings.Contains(gr, "--detach-sign") {
			t.Error("GoReleaser must produce detached signatures")
		}
	})

	// AC2: checksums.txt has .sig signature (covered by artifacts: all)
	t.Run("checksums_signed", func(t *testing.T) {
		if !strings.Contains(gr, "checksums") {
			t.Error("GoReleaser must generate checksums")
		}
	})

	// AC3: GPG public key published at known URL
	t.Run("gpg_public_key_exists", func(t *testing.T) {
		keyPath := filepath.Join(projectRoot, "gpg.key")
		info, err := os.Stat(keyPath)
		if err != nil {
			t.Fatal("gpg.key must exist in repository root")
		}
		if info.Size() == 0 {
			t.Error("gpg.key must not be empty")
		}
	})

	// AC4: Verification instructions in release notes
	t.Run("verification_instructions", func(t *testing.T) {
		if !strings.Contains(gr, "gpg --verify") {
			t.Error("Release notes must include GPG verification instructions")
		}
		if !strings.Contains(gr, "gpg --import") {
			t.Error("Release notes must include GPG key import instructions")
		}
	})

	// AC5: CI release workflow has GPG signing enabled
	t.Run("release_workflow_gpg", func(t *testing.T) {
		if !strings.Contains(rel, "gpg --batch --import") {
			t.Error("Release workflow must import GPG private key")
		}
		if !strings.Contains(rel, "GPG_PRIVATE_KEY") {
			t.Error("Release workflow must reference GPG_PRIVATE_KEY secret")
		}
		if !strings.Contains(rel, "GPG_FINGERPRINT") {
			t.Error("Release workflow must pass GPG_FINGERPRINT to GoReleaser")
		}
	})
}
