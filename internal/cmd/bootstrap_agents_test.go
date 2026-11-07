package cmd

import (
	"github.com/rtmx-ai/rtmx/pkg/rtmx"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func resetBootstrapAgentsFlags() func() {
	origDry := bootstrapAgentsDryRun
	origForce := bootstrapAgentsForce
	origEmit := bootstrapAgentsEmit
	origMD := bootstrapAgentsMD
	origTargets := bootstrapAgentsTargets
	origEmitDry := emitDryRun
	origEmitForce := emitForce
	origEmitMD := emitAgentsMD
	origEmitAll := emitAll
	origEmitList := emitList
	origEmitAgents := emitAgents

	bootstrapAgentsDryRun = false
	bootstrapAgentsForce = false
	bootstrapAgentsEmit = true
	bootstrapAgentsMD = false
	bootstrapAgentsTargets = []string{"claude", "cursor", "codex"}
	emitDryRun, emitForce, emitAgentsMD, emitAll, emitList = false, false, false, false, false
	emitAgents = nil

	return func() {
		bootstrapAgentsDryRun = origDry
		bootstrapAgentsForce = origForce
		bootstrapAgentsEmit = origEmit
		bootstrapAgentsMD = origMD
		bootstrapAgentsTargets = origTargets
		emitDryRun = origEmitDry
		emitForce = origEmitForce
		emitAgentsMD = origEmitMD
		emitAll = origEmitAll
		emitList = origEmitList
		emitAgents = origEmitAgents
	}
}

func TestBootstrapAgentsDryRun(t *testing.T) {
	rtmx.Req(t, "REQ-GO-001")

	defer resetBootstrapAgentsFlags()()
	bootstrapAgentsDryRun = true
	bootstrapAgentsEmit = true
	bootstrapAgentsTargets = []string{"cursor"}

	tmp := t.TempDir()
	t.Setenv("RTMX_AGENT_PACK", tmp)
	_ = os.WriteFile(filepath.Join(tmp, "mcp.json"),
		[]byte(`{"mcpServers":{"rtmx":{"command":"rtmx","args":["mcp-server","--stdio"]}}}`), 0644)
	cwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(cwd) }()
	work := filepath.Join(tmp, "proj")
	_ = os.MkdirAll(work, 0755)
	_ = os.Chdir(work)

	cmd := &cobra.Command{Use: "agents", RunE: runBootstrapAgents}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := runBootstrapAgents(cmd, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "DRY RUN") {
		t.Fatalf("expected dry-run banner:\n%s", out)
	}
	if !strings.Contains(out, "marketplace") && !strings.Contains(out, "Marketplace") {
		t.Fatalf("expected marketplace guidance:\n%s", out)
	}
	if !strings.Contains(out, "brew install") {
		t.Fatalf("expected PATH/brew reminder:\n%s", out)
	}
	if _, err := os.Stat(filepath.Join(work, ".cursor", "mcp.json")); !os.IsNotExist(err) {
		t.Fatal("dry-run must not write emit files")
	}
}

func TestBootstrapAgentsSkipEmit(t *testing.T) {
	defer resetBootstrapAgentsFlags()()
	bootstrapAgentsEmit = false

	cmd := &cobra.Command{Use: "agents", RunE: runBootstrapAgents}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	if err := runBootstrapAgents(cmd, nil); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "Emit skipped") {
		t.Fatalf("expected skip emit:\n%s", out)
	}
	if !strings.Contains(out, "plugin emit") {
		t.Fatalf("expected offline emit hint:\n%s", out)
	}
}

func TestBootstrapAgentsNoOverwriteAgentsMDWithoutForce(t *testing.T) {
	defer resetBootstrapAgentsFlags()()
	bootstrapAgentsEmit = true
	bootstrapAgentsMD = true
	bootstrapAgentsForce = false
	bootstrapAgentsTargets = []string{"codex"}

	tmp := t.TempDir()
	t.Setenv("RTMX_AGENT_PACK", tmp)
	_ = os.WriteFile(filepath.Join(tmp, "mcp.json"),
		[]byte(`{"mcpServers":{"rtmx":{"command":"rtmx","args":["mcp-server","--stdio"]}}}`), 0644)
	cwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(cwd) }()
	work := filepath.Join(tmp, "proj")
	_ = os.MkdirAll(work, 0755)
	_ = os.Chdir(work)

	existing := "# keep me\n"
	if err := os.WriteFile(filepath.Join(work, "AGENTS.md"), []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{Use: "agents", RunE: runBootstrapAgents}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	if err := runBootstrapAgents(cmd, nil); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(work, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != existing {
		t.Fatalf("AGENTS.md overwritten without --force:\n%s", data)
	}
	if !strings.Contains(buf.String(), "Skip:") && !strings.Contains(buf.String(), "exists") {
		t.Fatalf("expected skip notice:\n%s", buf.String())
	}
}

func TestBootstrapAgentsForceAgentsMD(t *testing.T) {
	defer resetBootstrapAgentsFlags()()
	bootstrapAgentsEmit = true
	bootstrapAgentsMD = true
	bootstrapAgentsForce = true
	bootstrapAgentsTargets = []string{"codex"}

	tmp := t.TempDir()
	t.Setenv("RTMX_AGENT_PACK", tmp)
	_ = os.WriteFile(filepath.Join(tmp, "mcp.json"),
		[]byte(`{"mcpServers":{"rtmx":{"command":"rtmx","args":["mcp-server","--stdio"]}}}`), 0644)
	cwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(cwd) }()
	work := filepath.Join(tmp, "proj")
	_ = os.MkdirAll(work, 0755)
	_ = os.Chdir(work)
	_ = os.WriteFile(filepath.Join(work, "AGENTS.md"), []byte("# old\n"), 0644)

	cmd := &cobra.Command{Use: "agents", RunE: runBootstrapAgents}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	if err := runBootstrapAgents(cmd, nil); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(work, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "codex mcp add rtmx") {
		t.Fatalf("expected AGENTS.md nudge after --force:\n%s", data)
	}
}

func TestBootstrapAgentsCmdRegistered(t *testing.T) {
	found := false
	for _, c := range bootstrapCmd.Commands() {
		if c.Name() == "agents" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("bootstrap agents subcommand not registered")
	}
	if bootstrapAgentsCmd.Flags().Lookup("dry-run") == nil {
		t.Fatal("missing --dry-run")
	}
	if bootstrapAgentsCmd.Flags().Lookup("force") == nil {
		t.Fatal("missing --force")
	}
	if bootstrapAgentsCmd.Flags().Lookup("emit") == nil {
		t.Fatal("missing --emit")
	}
}
