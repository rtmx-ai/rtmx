package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx-go/pkg/rtmx"
)

// TestAptInstall validates that release infrastructure is configured for
// Debian/Ubuntu APT and RPM-based Linux package distribution via GoReleaser nfpms.
// REQ-DIST-002: APT/DEB Repository (Linux)
func TestAptInstall(t *testing.T) {
	rtmx.Req(t, "REQ-DIST-002")

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

	// AC1: GoReleaser has nfpms section for Linux packages
	t.Run("nfpms_configured", func(t *testing.T) {
		if !strings.Contains(gr, "nfpms:") {
			t.Fatal("GoReleaser must have nfpms section for Linux package generation")
		}
	})

	// AC2: deb format is configured for Debian/Ubuntu
	t.Run("deb_format", func(t *testing.T) {
		if !strings.Contains(gr, "- deb") {
			t.Fatal("nfpms must include deb format for Debian/Ubuntu")
		}
	})

	// AC3: rpm format is configured for Red Hat/Fedora
	t.Run("rpm_format", func(t *testing.T) {
		if !strings.Contains(gr, "- rpm") {
			t.Fatal("nfpms must include rpm format for RPM-based distros")
		}
	})

	// AC4: Package name is rtmx
	t.Run("package_name", func(t *testing.T) {
		if !strings.Contains(gr, "package_name: rtmx") {
			t.Error("nfpms package_name must be rtmx")
		}
	})

	// AC5: Binary installed to /usr/bin
	t.Run("bindir", func(t *testing.T) {
		if !strings.Contains(gr, "bindir: /usr/bin") {
			t.Error("nfpms bindir must be /usr/bin for standard PATH availability")
		}
	})

	// AC6: Package metadata is correct
	t.Run("package_metadata", func(t *testing.T) {
		if !strings.Contains(gr, "vendor: ioTACTICAL LLC") {
			t.Error("nfpms vendor must be ioTACTICAL LLC")
		}
		if !strings.Contains(gr, "homepage: https://rtmx.ai") {
			t.Error("nfpms homepage must be https://rtmx.ai")
		}
		if !strings.Contains(gr, "maintainer: RTMX Engineering <dev@rtmx.ai>") {
			t.Error("nfpms maintainer must be RTMX Engineering <dev@rtmx.ai>")
		}
		if !strings.Contains(gr, "license: Apache-2.0") {
			t.Error("nfpms license must be Apache-2.0")
		}
	})

	// AC7: Package description is set
	t.Run("package_description", func(t *testing.T) {
		if !strings.Contains(gr, "description: Requirements Traceability Matrix toolkit") {
			t.Error("nfpms description must be set")
		}
	})

	// AC8: Linux is in the build matrix
	t.Run("linux_in_build_matrix", func(t *testing.T) {
		if !strings.Contains(gr, "- linux") {
			t.Error("Linux must be in the goos build matrix")
		}
	})

	// AC9: Both amd64 and arm64 architectures supported
	t.Run("architectures", func(t *testing.T) {
		if !strings.Contains(gr, "- amd64") {
			t.Error("amd64 must be in goarch build matrix")
		}
		if !strings.Contains(gr, "- arm64") {
			t.Error("arm64 must be in goarch build matrix")
		}
	})
}
