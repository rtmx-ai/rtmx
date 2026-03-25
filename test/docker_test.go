package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx-go/pkg/rtmx"
)

// TestDockerImage validates that the project is configured to publish
// multi-arch Docker images to ghcr.io via GoReleaser.
// REQ-REL-005: Release workflow shall publish multi-arch Docker images to ghcr.io.
func TestDockerImage(t *testing.T) {
	rtmx.Req(t, "REQ-REL-005")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	// AC1: Dockerfile exists
	t.Run("dockerfile_exists", func(t *testing.T) {
		dockerfilePath := filepath.Join(projectRoot, "Dockerfile")
		if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
			t.Fatal("Dockerfile must exist in the project root")
		}
	})

	// AC2: Dockerfile uses a multi-stage build
	t.Run("dockerfile_multi_stage", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "Dockerfile"))
		if err != nil {
			t.Fatalf("Failed to read Dockerfile: %v", err)
		}
		df := string(content)
		fromCount := strings.Count(strings.ToUpper(df), "FROM ")
		if fromCount < 2 {
			t.Fatalf("Dockerfile must use multi-stage build (found %d FROM directives, need at least 2)", fromCount)
		}
		if !strings.Contains(df, "AS ") && !strings.Contains(df, "as ") {
			t.Error("Dockerfile multi-stage build should use named stages (AS keyword)")
		}
	})

	// AC3: Dockerfile runtime stage uses alpine or scratch or distroless
	t.Run("dockerfile_minimal_base", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "Dockerfile"))
		if err != nil {
			t.Fatalf("Failed to read Dockerfile: %v", err)
		}
		df := strings.ToLower(string(content))
		// Find the last FROM line (runtime stage)
		lines := strings.Split(df, "\n")
		var lastFrom string
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "from ") {
				lastFrom = trimmed
			}
		}
		if lastFrom == "" {
			t.Fatal("No FROM directive found in Dockerfile")
		}
		if !strings.Contains(lastFrom, "scratch") &&
			!strings.Contains(lastFrom, "distroless") &&
			!strings.Contains(lastFrom, "alpine") {
			t.Fatalf("Runtime stage must use scratch, distroless, or alpine base; got: %s", lastFrom)
		}
	})

	// AC4: GoReleaser has docker configuration for multi-arch
	t.Run("goreleaser_docker_config", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".goreleaser.yaml"))
		if err != nil {
			t.Fatalf("Failed to read .goreleaser.yaml: %v", err)
		}
		gr := string(content)
		if !strings.Contains(gr, "dockers:") {
			t.Fatal("GoReleaser must have a dockers: section for Docker image builds")
		}
		if !strings.Contains(gr, "ghcr.io/rtmx-ai/rtmx") {
			t.Error("Docker image templates must target ghcr.io/rtmx-ai/rtmx")
		}
		if !strings.Contains(gr, "linux/amd64") {
			t.Error("Docker config must include linux/amd64 platform")
		}
		if !strings.Contains(gr, "linux/arm64") {
			t.Error("Docker config must include linux/arm64 platform")
		}
	})

	// AC5: GoReleaser has docker manifest configuration for multi-arch tags
	t.Run("goreleaser_docker_manifests", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".goreleaser.yaml"))
		if err != nil {
			t.Fatalf("Failed to read .goreleaser.yaml: %v", err)
		}
		gr := string(content)
		if !strings.Contains(gr, "docker_manifests:") {
			t.Fatal("GoReleaser must have a docker_manifests: section for multi-arch manifest lists")
		}
		if !strings.Contains(gr, "latest") {
			t.Error("Docker manifests should include a :latest tag")
		}
	})

	// AC6: Release workflow has packages:write permission
	t.Run("release_workflow_packages_permission", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".github", "workflows", "release.yml"))
		if err != nil {
			t.Fatalf("Failed to read release.yml: %v", err)
		}
		wf := string(content)
		if !strings.Contains(wf, "packages: write") {
			t.Fatal("Release workflow must have packages: write permission for ghcr.io push")
		}
	})

	// AC7: Dockerfile sets ENTRYPOINT to rtmx
	t.Run("dockerfile_entrypoint", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "Dockerfile"))
		if err != nil {
			t.Fatalf("Failed to read Dockerfile: %v", err)
		}
		df := string(content)
		if !strings.Contains(df, "ENTRYPOINT") {
			t.Fatal("Dockerfile must set an ENTRYPOINT")
		}
		if !strings.Contains(df, "rtmx") {
			t.Error("Dockerfile ENTRYPOINT should reference the rtmx binary")
		}
	})
}
