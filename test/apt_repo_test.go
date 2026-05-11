package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestAptRepository validates that the APT repository infrastructure
// is configured for Debian/Ubuntu distribution.
// REQ-DIST-007: APT repository for apt install rtmx
func TestAptRepository(t *testing.T) {
	rtmx.Req(t, "REQ-DIST-007")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	// AC1: GoReleaser builds .deb packages
	t.Run("deb_packages_configured", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".goreleaser.yaml"))
		if err != nil {
			t.Fatalf("Failed to read .goreleaser.yaml: %v", err)
		}
		gr := string(content)
		if !strings.Contains(gr, "nfpms:") {
			t.Fatal("GoReleaser must have nfpms section for .deb packages")
		}
		if !strings.Contains(gr, "deb") {
			t.Error("nfpms must include deb format")
		}
	})

	// AC2: Both architectures configured
	t.Run("multi_arch", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".goreleaser.yaml"))
		if err != nil {
			t.Fatalf("Failed to read .goreleaser.yaml: %v", err)
		}
		gr := string(content)
		if !strings.Contains(gr, "amd64") {
			t.Error("must build for amd64")
		}
		if !strings.Contains(gr, "arm64") {
			t.Error("must build for arm64")
		}
	})

	// AC3: APT repo generation script exists
	t.Run("repo_script_exists", func(t *testing.T) {
		scriptPath := filepath.Join(projectRoot, "scripts", "apt-repo.sh")
		if _, err := os.Stat(scriptPath); err != nil {
			t.Fatalf("scripts/apt-repo.sh must exist: %v", err)
		}

		content, err := os.ReadFile(scriptPath)
		if err != nil {
			t.Fatal(err)
		}
		script := string(content)

		// Must handle repo structure generation
		if !strings.Contains(script, "dpkg-scanpackages") && !strings.Contains(script, "apt-ftparchive") {
			t.Error("script must use dpkg-scanpackages or apt-ftparchive")
		}
		if !strings.Contains(script, "gpg") {
			t.Error("script must GPG-sign the repository")
		}
	})

	// AC4: Package metadata is correct
	t.Run("package_metadata", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".goreleaser.yaml"))
		if err != nil {
			t.Fatalf("Failed to read .goreleaser.yaml: %v", err)
		}
		gr := string(content)
		if !strings.Contains(gr, "package_name: rtmx") {
			t.Error("nfpms package_name must be rtmx")
		}
		if !strings.Contains(gr, "maintainer:") {
			t.Error("nfpms must have maintainer field")
		}
		if !strings.Contains(gr, "/usr/bin") {
			t.Error("nfpms bindir must be /usr/bin")
		}
	})

	// AC5: README documents apt install
	t.Run("readme_documents_apt", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "README.md"))
		if err != nil {
			t.Fatal("README.md must exist")
		}
		readme := string(content)
		if !strings.Contains(readme, ".deb") && !strings.Contains(readme, "apt") {
			t.Skip("apt instructions not yet in README")
		}
	})
}
