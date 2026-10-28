package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func newTestEmitRoot() *cobra.Command {
	root := &cobra.Command{Use: "rtmx", SilenceUsage: true, SilenceErrors: true}
	plugin := &cobra.Command{Use: "plugin"}
	emit := &cobra.Command{
		Use:  "emit",
		RunE: runPluginEmit,
	}
	emit.Flags().StringSliceVar(&emitAgents, "agents", nil, "")
	emit.Flags().BoolVar(&emitList, "list", false, "")
	emit.Flags().BoolVar(&emitDryRun, "dry-run", false, "")
	emit.Flags().BoolVar(&emitForce, "force", false, "")
	emit.Flags().BoolVar(&emitAll, "all", false, "")
	emit.Flags().BoolVar(&emitAgentsMD, "agents-md", false, "")
	plugin.AddCommand(emit)
	root.AddCommand(plugin)
	return root
}

func TestPluginEmitList(t *testing.T) {
	emitList, emitAgents, emitAll, emitDryRun, emitForce = true, nil, false, false, false
	root := newTestEmitRoot()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"plugin", "emit", "--list"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, name := range []string{"claude", "cursor", "codex", "gemini", "windsurf", "continue", "gastown"} {
		if !strings.Contains(out, name) {
			t.Fatalf("list missing %s:\n%s", name, out)
		}
	}
}

func TestPluginEmitWritesFromPack(t *testing.T) {
	emitList, emitAll, emitDryRun, emitForce = false, true, false, true
	emitAgents = nil

	tmp := t.TempDir()
	packDir := filepath.Join(tmp, "pack")
	_ = os.MkdirAll(packDir, 0755)
	packBody := []byte(`{"mcpServers":{"rtmx":{"command":"rtmx","args":["mcp-server","--stdio"]}}}`)
	if err := os.WriteFile(filepath.Join(packDir, "mcp.json"), packBody, 0644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RTMX_AGENT_PACK", packDir)

	cwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(cwd) }()
	work := filepath.Join(tmp, "proj")
	_ = os.MkdirAll(work, 0755)
	if err := os.Chdir(work); err != nil {
		t.Fatal(err)
	}

	root := newTestEmitRoot()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"plugin", "emit", "--all", "--force"})
	if err := root.Execute(); err != nil {
		t.Fatalf("emit: %v\n%s", err, buf.String())
	}

	cursorPath := filepath.Join(work, ".cursor", "mcp.json")
	data, err := os.ReadFile(cursorPath)
	if err != nil {
		t.Fatal(err)
	}
	var parsed map[string]map[string]struct {
		Command string   `json:"command"`
		Args    []string `json:"args"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatal(err)
	}
	srv := parsed["mcpServers"]["rtmx"]
	if srv.Command != "rtmx" || strings.Join(srv.Args, " ") != "mcp-server --stdio" {
		t.Fatalf("cursor mcp diverged: %#v", srv)
	}

	codex := filepath.Join(work, ".rtmx", "emit", "codex", "mcp_servers.rtmx.toml")
	toml, err := os.ReadFile(codex)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(toml), `[mcp_servers.rtmx]`) || !strings.Contains(string(toml), `--stdio`) {
		t.Fatalf("codex snippet bad:\n%s", toml)
	}

	claudeMCP := filepath.Join(work, ".rtmx", "emit", "claude", "mcp.json")
	if _, err := os.Stat(claudeMCP); err != nil {
		t.Fatal(err)
	}

	for _, rel := range []string{
		".rtmx/emit/gemini/rtmx-extension.json",
		".rtmx/emit/windsurf/mcp.json",
		".rtmx/emit/continue/mcp-rtmx.yaml",
		".rtmx/emit/gastown/rtmx-plugin.json",
	} {
		p := filepath.Join(work, filepath.FromSlash(rel))
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("long-tail missing %s: %v", rel, err)
		}
		if !strings.Contains(string(data), "mcp-server") || !strings.Contains(string(data), "stdio") {
			t.Fatalf("%s must reuse pack args:\n%s", rel, data)
		}
	}
}

func TestPluginEmitDryRun(t *testing.T) {
	emitList, emitAll, emitDryRun, emitForce, emitAgentsMD = false, false, true, false, false
	emitAgents = []string{"cursor"}
	tmp := t.TempDir()
	t.Setenv("RTMX_AGENT_PACK", tmp)
	_ = os.WriteFile(filepath.Join(tmp, "mcp.json"),
		[]byte(`{"mcpServers":{"rtmx":{"command":"rtmx","args":["mcp-server","--stdio"]}}}`), 0644)
	cwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(cwd) }()
	work := filepath.Join(tmp, "proj")
	_ = os.MkdirAll(work, 0755)
	_ = os.Chdir(work)

	root := newTestEmitRoot()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"plugin", "emit", "--agents", "cursor", "--dry-run"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(work, ".cursor", "mcp.json")); !os.IsNotExist(err) {
		t.Fatal("dry-run must not write")
	}
	if !strings.Contains(buf.String(), "Would write") {
		t.Fatalf("expected dry-run preview:\n%s", buf.String())
	}
}

func TestPluginEmitAgentsMDOptional(t *testing.T) {
	emitList, emitAll, emitDryRun, emitForce, emitAgentsMD = false, false, false, true, true
	emitAgents = []string{"codex"}
	tmp := t.TempDir()
	t.Setenv("RTMX_AGENT_PACK", tmp)
	_ = os.WriteFile(filepath.Join(tmp, "mcp.json"),
		[]byte(`{"mcpServers":{"rtmx":{"command":"rtmx","args":["mcp-server","--stdio"]}}}`), 0644)
	cwd, _ := os.Getwd()
	defer func() { _ = os.Chdir(cwd) }()
	work := filepath.Join(tmp, "proj")
	_ = os.MkdirAll(work, 0755)
	_ = os.Chdir(work)

	root := newTestEmitRoot()
	root.SetArgs([]string{"plugin", "emit", "--agents", "codex", "--agents-md", "--force"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(work, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "codex mcp add rtmx") {
		t.Fatalf("AGENTS.md missing Codex one-liner:\n%s", data)
	}
}
