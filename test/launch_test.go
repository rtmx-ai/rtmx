package test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rtmx-ai/rtmx/pkg/rtmx"
)

// TestReadmeLaunchReady validates that the README is structured as a
// landing page suitable for Show HN.
// REQ-LAUNCH-001: README rewritten as landing page
func TestReadmeLaunchReady(t *testing.T) {
	rtmx.Req(t, "REQ-LAUNCH-001")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	content, err := os.ReadFile(filepath.Join(projectRoot, "README.md"))
	if err != nil {
		t.Fatalf("README.md must exist: %v", err)
	}
	readme := string(content)

	// AC1: Hero section with workflow GIF reference
	t.Run("hero_section", func(t *testing.T) {
		// Must have a prominent heading and description within first 20 lines
		lines := strings.Split(readme, "\n")
		if len(lines) < 5 {
			t.Fatal("README too short")
		}
		if !strings.HasPrefix(lines[0], "#") {
			t.Error("README must start with a heading")
		}
		// GIF reference (may be commented out until generated)
		if !strings.Contains(readme, "workflow") && !strings.Contains(readme, "gif") {
			t.Error("README should reference a workflow GIF")
		}
	})

	// AC2: Install command within first screenful (first 30 lines)
	t.Run("install_prominent", func(t *testing.T) {
		lines := strings.Split(readme, "\n")
		first30 := strings.Join(lines[:min(30, len(lines))], "\n")
		if !strings.Contains(first30, "brew install") {
			t.Error("brew install command must appear in first 30 lines")
		}
	})

	// AC3: Mermaid diagrams embedded
	t.Run("mermaid_diagrams", func(t *testing.T) {
		mermaidCount := strings.Count(readme, "```mermaid")
		if mermaidCount < 2 {
			t.Errorf("README should have at least 2 Mermaid diagrams, got %d", mermaidCount)
		}
	})

	// AC4: Key sections present
	t.Run("key_sections", func(t *testing.T) {
		requiredSections := []string{
			"Install",
			"What It Does",
			"MCP",
		}
		for _, section := range requiredSections {
			if !strings.Contains(readme, section) {
				t.Errorf("README must contain %q section", section)
			}
		}
	})

	// AC5: Blog post link
	t.Run("blog_link", func(t *testing.T) {
		if !strings.Contains(readme, "rtmx.ai") {
			t.Error("README must link to rtmx.ai")
		}
	})

	// AC6: Self-referential dogfooding section
	t.Run("dogfooding", func(t *testing.T) {
		if !strings.Contains(readme, "requirements") && !strings.Contains(readme, "dogfood") {
			t.Error("README should mention self-referential requirements management")
		}
	})

	// AC7: Migration info moved to docs or below fold
	t.Run("migration_below_fold", func(t *testing.T) {
		// Migration guide content should not be in the first half
		lines := strings.Split(readme, "\n")
		midpoint := len(lines) / 2
		firstHalf := strings.Join(lines[:midpoint], "\n")
		if strings.Contains(firstHalf, "Migrating from Python") && strings.Contains(firstHalf, "pip install") {
			t.Error("Migration instructions should be below the fold or in docs/")
		}
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestLaunchChecklist validates that all Show HN launch prerequisites are met.
// REQ-LAUNCH-002: Show HN launch readiness
func TestLaunchChecklist(t *testing.T) {
	rtmx.Req(t, "REQ-LAUNCH-002")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	// AC1: README has install commands
	t.Run("install_commands_documented", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "README.md"))
		if err != nil {
			t.Fatal("README.md must exist")
		}
		readme := string(content)
		installs := []string{"brew install", "scoop install", "go install"}
		for _, cmd := range installs {
			if !strings.Contains(readme, cmd) {
				t.Errorf("README must document %q", cmd)
			}
		}
	})

	// AC2: GitHub releases page infrastructure
	t.Run("release_workflow_exists", func(t *testing.T) {
		if _, err := os.Stat(filepath.Join(projectRoot, ".github", "workflows", "release.yml")); err != nil {
			t.Fatal("release.yml workflow must exist")
		}
	})

	// AC3: GoReleaser produces all artifact types
	t.Run("all_artifact_types", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, ".goreleaser.yaml"))
		if err != nil {
			t.Fatal("GoReleaser config must exist")
		}
		gr := string(content)
		for _, section := range []string{"archives:", "nfpms:", "brews:", "scoops:"} {
			if !strings.Contains(gr, section) {
				t.Errorf("GoReleaser must have %s", section)
			}
		}
	})

	// AC4: Blog post link in README
	t.Run("blog_post_linked", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "README.md"))
		if err != nil {
			t.Fatal("README.md must exist")
		}
		if !strings.Contains(string(content), "rtmx.ai/blog/show-hn-rtmx") {
			t.Error("README must link to Show HN blog post")
		}
	})

	// AC5: Init and setup commands exist
	t.Run("onboarding_commands", func(t *testing.T) {
		for _, cmd := range []string{"init.go", "setup.go"} {
			path := filepath.Join(projectRoot, "internal", "cmd", cmd)
			if _, err := os.Stat(path); err != nil {
				t.Errorf("internal/cmd/%s must exist for onboarding flow", cmd)
			}
		}
	})

	// AC6: SECURITY.md exists
	t.Run("security_policy", func(t *testing.T) {
		if _, err := os.Stat(filepath.Join(projectRoot, "SECURITY.md")); err != nil {
			t.Error("SECURITY.md must exist")
		}
	})

	// AC7: LICENSE exists
	t.Run("license", func(t *testing.T) {
		if _, err := os.Stat(filepath.Join(projectRoot, "LICENSE")); err != nil {
			t.Error("LICENSE must exist")
		}
	})
}

// TestVhsGifGeneration validates that VHS tape files exist and are configured
// to generate GIF assets for the README.
// REQ-LAUNCH-003: VHS GIF generation from terminal demos
func TestVhsGifGeneration(t *testing.T) {
	rtmx.Req(t, "REQ-LAUNCH-003")

	wd, _ := os.Getwd()
	projectRoot := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(projectRoot, "cmd/rtmx")); err != nil {
		projectRoot = wd
	}

	// AC1: Workflow tape file exists with correct output directive
	t.Run("workflow_tape_exists", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "docs", "tapes", "workflow.tape"))
		if err != nil {
			t.Fatalf("docs/tapes/workflow.tape must exist: %v", err)
		}
		tape := string(content)
		if !strings.Contains(tape, "Output") {
			t.Error("tape must have Output directive")
		}
		if !strings.Contains(tape, "rtmx-workflow.gif") {
			t.Error("tape output must target rtmx-workflow.gif")
		}
	})

	// AC2: Agent-loop tape file exists with correct output directive
	t.Run("agent_loop_tape_exists", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "docs", "tapes", "agent-loop.tape"))
		if err != nil {
			t.Fatalf("docs/tapes/agent-loop.tape must exist: %v", err)
		}
		tape := string(content)
		if !strings.Contains(tape, "Output") {
			t.Error("tape must have Output directive")
		}
		if !strings.Contains(tape, "rtmx-agent-loop.gif") {
			t.Error("tape output must target rtmx-agent-loop.gif")
		}
	})

	// AC3: Tapes exercise core rtmx commands
	t.Run("tapes_exercise_commands", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "docs", "tapes", "workflow.tape"))
		if err != nil {
			t.Fatal(err)
		}
		tape := string(content)
		for _, cmd := range []string{"rtmx status", "rtmx verify"} {
			if !strings.Contains(tape, cmd) {
				t.Errorf("workflow tape must exercise %q", cmd)
			}
		}
	})

	// AC4: README references the GIF (commented or uncommented)
	t.Run("readme_references_gif", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "README.md"))
		if err != nil {
			t.Fatal(err)
		}
		readme := string(content)
		if !strings.Contains(readme, "rtmx-workflow.gif") && !strings.Contains(readme, "workflow") {
			t.Error("README must reference the workflow GIF")
		}
	})

	// AC5: Tape configuration uses readable settings
	t.Run("tape_configuration", func(t *testing.T) {
		content, err := os.ReadFile(filepath.Join(projectRoot, "docs", "tapes", "workflow.tape"))
		if err != nil {
			t.Fatal(err)
		}
		tape := string(content)
		if !strings.Contains(tape, "Set Width") {
			t.Error("tape must set width for consistent rendering")
		}
		if !strings.Contains(tape, "Set Height") {
			t.Error("tape must set height for consistent rendering")
		}
		if !strings.Contains(tape, "Set Theme") {
			t.Error("tape must set theme for consistent rendering")
		}
	})
}
