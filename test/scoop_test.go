package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestScoopInstall validates that release infrastructure is configured for
// Scoop package manager distribution on Windows.
// REQ-DIST-001: Scoop Package (Windows)
func TestScoopInstall(t *testing.T) {
	rtmx.Req(t, "REQ-DIST-001")

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

	// AC1: GoReleaser has Scoop bucket configuration
	t.Run("scoop_bucket_configured", func(t *testing.T) {
		if !strings.Contains(gr, "scoops:") {
			t.Fatal("GoReleaser must have scoops section")
		}
	})

	// AC2: Scoop bucket targets rtmx-ai/scoop-bucket
	t.Run("scoop_bucket_target", func(t *testing.T) {
		if !strings.Contains(gr, "name: scoop-bucket") {
			t.Error("Scoop repository must target scoop-bucket")
		}
		if !strings.Contains(gr, "owner: rtmx-ai") {
			t.Error("Scoop repository owner must be rtmx-ai")
		}
	})

	// AC3: Scoop bucket token is configured for automated publishing
	t.Run("scoop_bucket_token", func(t *testing.T) {
		if !strings.Contains(gr, "SCOOP_BUCKET_TOKEN") {
			t.Error("GoReleaser must reference SCOOP_BUCKET_TOKEN for automated publishing")
		}
	})

	// AC4: Binary name is rtmx
	t.Run("binary_name", func(t *testing.T) {
		if !strings.Contains(gr, "binary: rtmx") {
			t.Error("Build binary name must be rtmx")
		}
	})

	// AC5: Windows is in the build matrix
	t.Run("windows_in_build_matrix", func(t *testing.T) {
		if !strings.Contains(gr, "- windows") {
			t.Error("Windows must be in the goos build matrix")
		}
	})

	// AC6: Both amd64 and arm64 architectures supported
	t.Run("architectures", func(t *testing.T) {
		if !strings.Contains(gr, "- amd64") {
			t.Error("amd64 must be in goarch build matrix")
		}
		if !strings.Contains(gr, "- arm64") {
			t.Error("arm64 must be in goarch build matrix")
		}
	})

	// AC7: Windows archives use zip format
	t.Run("windows_zip_format", func(t *testing.T) {
		if !strings.Contains(gr, "goos: windows") {
			t.Error("GoReleaser must have Windows-specific format override")
		}
		if !strings.Contains(gr, "- zip") {
			t.Error("Windows archives must use zip format")
		}
	})

	// AC8: Scoop metadata is correct
	t.Run("scoop_metadata", func(t *testing.T) {
		if !strings.Contains(gr, "homepage: \"https://rtmx.ai\"") {
			t.Error("Scoop homepage must be https://rtmx.ai")
		}
		if !strings.Contains(gr, "license: \"Apache-2.0\"") {
			t.Error("Scoop license must be Apache-2.0")
		}
	})
}
