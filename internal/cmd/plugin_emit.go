package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rtmx-ai/rtmx/internal/agentpack"
	"github.com/rtmx-ai/rtmx/internal/output"
	"github.com/spf13/cobra"
)

var (
	emitAgents   []string
	emitList     bool
	emitDryRun   bool
	emitForce    bool
	emitAll      bool
	emitAgentsMD bool
)

var pluginEmitCmd = &cobra.Command{
	Use:   "emit",
	Short: "Emit agent MCP/plugin configs from the shared pack",
	Long: `Write client-specific MCP configs from plugins/rtmx (marketplace pack).

Supported targets: claude, cursor, codex, gemini, windsurf, continue, gastown.

  rtmx plugin emit --list
  rtmx plugin emit --agents cursor,codex
  rtmx plugin emit --all --dry-run
  rtmx plugin emit --agents gemini,windsurf,continue --force
  rtmx plugin emit --agents codex --agents-md   # optional AGENTS.md nudge

Pack resolution: $RTMX_AGENT_PACK, walk-up plugins/rtmx/mcp.json, then
the embed shipped with the CLI (must match plugins/rtmx/mcp.json).

Overwrite policy: existing files are left untouched unless --force.
Dry-run prints paths without writing.`,
	RunE: runPluginEmit,
}

func init() {
	pluginEmitCmd.Flags().StringSliceVar(&emitAgents, "agents", nil, "targets to emit (see --list)")
	pluginEmitCmd.Flags().BoolVar(&emitList, "list", false, "list emit targets")
	pluginEmitCmd.Flags().BoolVar(&emitDryRun, "dry-run", false, "preview without writing")
	pluginEmitCmd.Flags().BoolVar(&emitForce, "force", false, "overwrite existing emit outputs")
	pluginEmitCmd.Flags().BoolVar(&emitAll, "all", false, "emit all targets")
	pluginEmitCmd.Flags().BoolVar(&emitAgentsMD, "agents-md", false, "also write optional AGENTS.md RTMX nudge (never forced)")
	pluginCmd.AddCommand(pluginEmitCmd)
	pluginCmd.Short = "Manage schema plugins and emit agent MCP packs"
	pluginCmd.Long = `Install, list, and remove schema plugins, or emit agent MCP configs.

Schema plugins add domain-specific columns to the RTMX database.

  rtmx plugin list
  rtmx plugin install ./myschema.yaml
  rtmx plugin remove myschema

Agent pack emit writes MCP configs from the shared marketplace pack
(plugins/rtmx) for Claude, Cursor, Codex, and long-tail agents:

  rtmx plugin emit --list
  rtmx plugin emit --agents cursor,codex,gemini
  rtmx plugin emit --all`
}

var emitTargets = []struct {
	Name string
	Desc string
}{
	{Name: "claude", Desc: "Claude marketplace-oriented MCP + install notes under .rtmx/emit/claude/"},
	{Name: "cursor", Desc: "Project .cursor/mcp.json from shared pack"},
	{Name: "codex", Desc: "Codex mcp_servers.rtmx TOML snippet under .rtmx/emit/codex/"},
	{Name: "gemini", Desc: "Gemini CLI extension JSON under .rtmx/emit/gemini/"},
	{Name: "windsurf", Desc: "Windsurf Cascade MCP JSON under .rtmx/emit/windsurf/"},
	{Name: "continue", Desc: "Continue.dev MCP YAML fragment under .rtmx/emit/continue/"},
	{Name: "gastown", Desc: "Gastown plugin MCP JSON under .rtmx/emit/gastown/"},
}

func runPluginEmit(cmd *cobra.Command, args []string) error {
	if noColor {
		output.DisableColor()
	}

	if emitList {
		cmd.Println(output.Header("Agent pack emit targets", output.TerminalWidth()))
		cmd.Println()
		for _, t := range emitTargets {
			cmd.Printf("  %-10s  %s\n", output.Color(t.Name, output.Cyan), t.Desc)
		}
		cmd.Println()
		return nil
	}

	targets := emitAgents
	if emitAll {
		targets = make([]string, 0, len(emitTargets))
		for _, t := range emitTargets {
			targets = append(targets, t.Name)
		}
	}
	if len(targets) == 0 {
		return fmt.Errorf("specify --agents, --all, or --list")
	}

	srv, src, err := agentpack.LoadMCP()
	if err != nil {
		return err
	}
	cmd.Printf("Pack MCP source: %s\n", src)
	cmd.Printf("  command=%s args=%v\n\n", srv.Command, srv.Args)

	if emitDryRun {
		cmd.Printf("%s\n\n", output.Color("DRY RUN - no files will be written", output.Yellow))
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	for _, name := range targets {
		name = strings.ToLower(strings.TrimSpace(name))
		switch name {
		case "claude":
			if err := emitClaude(cmd, cwd, srv); err != nil {
				return err
			}
		case "cursor":
			if err := emitCursor(cmd, cwd, srv); err != nil {
				return err
			}
		case "codex":
			if err := emitCodex(cmd, cwd, srv); err != nil {
				return err
			}
		case "gemini":
			if err := emitGemini(cmd, cwd, srv); err != nil {
				return err
			}
		case "windsurf":
			if err := emitWindsurf(cmd, cwd, srv); err != nil {
				return err
			}
		case "continue":
			if err := emitContinue(cmd, cwd, srv); err != nil {
				return err
			}
		case "gastown":
			if err := emitGastown(cmd, cwd, srv); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown emit target %q (use --list)", name)
		}
	}
	if emitAgentsMD {
		if err := emitAgentsMarkdown(cmd, cwd); err != nil {
			return err
		}
	}
	return nil
}

const agentsMDNudge = `# AGENTS.md — RTMX (optional)

When this repository has an RTM database (.rtmx/ or configured path) and the
rtmx MCP server is configured, prefer RTMX MCP tools (next, claim, verify,
status, backlog) over ad-hoc shell for requirements work.

Install Codex MCP (one command):

    codex mcp add rtmx -- rtmx mcp-server --stdio

Or run: rtmx plugin emit --agents codex
then merge the TOML snippet into ~/.codex/config.toml. Requires rtmx on PATH.
`

func emitWrite(cmd *cobra.Command, path string, data []byte) error {
	if _, err := os.Stat(path); err == nil && !emitForce && !emitDryRun {
		cmd.Printf("  %s %s (exists; use --force)\n", output.Color("Skip:", output.Yellow), path)
		return nil
	}
	if emitDryRun {
		cmd.Printf("  Would write: %s\n", path)
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}
	cmd.Printf("  %s %s\n", output.Color("Wrote:", output.Green), path)
	return nil
}

func emitClaude(cmd *cobra.Command, cwd string, srv agentpack.Server) error {
	dir := filepath.Join(cwd, ".rtmx", "emit", "claude")
	mcpPath := filepath.Join(dir, "mcp.json")
	readmePath := filepath.Join(dir, "INSTALL.md")
	body, err := json.MarshalIndent(map[string]interface{}{
		"mcpServers": map[string]agentpack.Server{"rtmx": srv},
	}, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	readme := `# Claude Code — RTMX emit

Prefer the first-party marketplace (REQ-MONO-062e):

` + "```bash\n" + `claude plugin marketplace add rtmx-ai/rtmx
claude plugin install rtmx@rtmx
` + "```\n\n" + `Local MCP payload (from shared pack) is in mcp.json next to this file.
Requires rtmx on PATH (brew install rtmx-ai/tap/rtmx).
`
	cmd.Println("Claude:")
	if err := emitWrite(cmd, mcpPath, body); err != nil {
		return err
	}
	return emitWrite(cmd, readmePath, []byte(readme))
}

func emitCursor(cmd *cobra.Command, cwd string, srv agentpack.Server) error {
	path := filepath.Join(cwd, ".cursor", "mcp.json")
	body, err := json.MarshalIndent(map[string]interface{}{
		"mcpServers": map[string]agentpack.Server{"rtmx": srv},
	}, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	cmd.Println("Cursor:")
	return emitWrite(cmd, path, body)
}

func emitCodex(cmd *cobra.Command, cwd string, srv agentpack.Server) error {
	dir := filepath.Join(cwd, ".rtmx", "emit", "codex")
	path := filepath.Join(dir, "mcp_servers.rtmx.toml")
	var b strings.Builder
	b.WriteString("# Append to ~/.codex/config.toml (or merge into project Codex config).\n")
	b.WriteString("# Preferred one-liner:\n#   codex mcp add rtmx -- ")
	b.WriteString(srv.Command)
	for _, a := range srv.Args {
		b.WriteByte(' ')
		b.WriteString(a)
	}
	b.WriteByte('\n')
	b.WriteString("\n[mcp_servers.rtmx]\n")
	b.WriteString(fmt.Sprintf("command = %q\n", srv.Command))
	b.WriteString("args = [")
	for i, a := range srv.Args {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(fmt.Sprintf("%q", a))
	}
	b.WriteString("]\n")
	cmd.Println("Codex:")
	readme := filepath.Join(dir, "INSTALL.md")
	installDoc := "# Codex MCP\n\n```bash\ncodex mcp add rtmx -- rtmx mcp-server --stdio\n```\n\n" +
		"Or merge `mcp_servers.rtmx.toml` into `~/.codex/config.toml`.\n" +
		"See docs/CODEX_MCP.md.\n"
	if err := emitWrite(cmd, path, []byte(b.String())); err != nil {
		return err
	}
	return emitWrite(cmd, readme, []byte(installDoc))
}

func emitAgentsMarkdown(cmd *cobra.Command, cwd string) error {
	path := filepath.Join(cwd, "AGENTS.md")
	cmd.Println("AGENTS.md (optional):")
	return emitWrite(cmd, path, []byte(agentsMDNudge))
}

func emitGemini(cmd *cobra.Command, cwd string, srv agentpack.Server) error {
	dir := filepath.Join(cwd, ".rtmx", "emit", "gemini")
	path := filepath.Join(dir, "rtmx-extension.json")
	body, err := json.MarshalIndent(map[string]interface{}{
		"name":        "rtmx",
		"description": "Requirements traceability management via RTMX",
		"transport": map[string]interface{}{
			"type":    "stdio",
			"command": srv.Command,
			"args":    srv.Args,
		},
	}, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	readme := "# Gemini CLI — RTMX emit\n\n" +
		"Copy `rtmx-extension.json` into your Gemini CLI extensions path.\n" +
		"MCP command/args match the shared marketplace pack (requires rtmx on PATH).\n" +
		"See docs/LONG_TAIL_EMIT.md.\n"
	cmd.Println("Gemini:")
	if err := emitWrite(cmd, path, body); err != nil {
		return err
	}
	return emitWrite(cmd, filepath.Join(dir, "INSTALL.md"), []byte(readme))
}

func emitWindsurf(cmd *cobra.Command, cwd string, srv agentpack.Server) error {
	dir := filepath.Join(cwd, ".rtmx", "emit", "windsurf")
	path := filepath.Join(dir, "mcp.json")
	body, err := json.MarshalIndent(map[string]interface{}{
		"mcpServers": map[string]agentpack.Server{"rtmx": srv},
	}, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	readme := "# Windsurf — RTMX emit\n\n" +
		"Merge `mcp.json` into Windsurf/Cascade MCP settings.\n" +
		"Same pack as Claude/Cursor: `rtmx mcp-server --stdio`.\n"
	cmd.Println("Windsurf:")
	if err := emitWrite(cmd, path, body); err != nil {
		return err
	}
	return emitWrite(cmd, filepath.Join(dir, "INSTALL.md"), []byte(readme))
}

func emitContinue(cmd *cobra.Command, cwd string, srv agentpack.Server) error {
	dir := filepath.Join(cwd, ".rtmx", "emit", "continue")
	path := filepath.Join(dir, "mcp-rtmx.yaml")
	var b strings.Builder
	b.WriteString("# Merge under Continue MCP / experimental MCP servers config.\n")
	b.WriteString("# Shared pack — do not diverge command/args.\n")
	b.WriteString("mcpServers:\n")
	b.WriteString("  - name: rtmx\n")
	b.WriteString(fmt.Sprintf("    command: %s\n", srv.Command))
	b.WriteString("    args:\n")
	for _, a := range srv.Args {
		b.WriteString(fmt.Sprintf("      - %s\n", a))
	}
	readme := "# Continue.dev — RTMX emit\n\nMerge `mcp-rtmx.yaml` into Continue MCP config. Requires rtmx on PATH.\n"
	cmd.Println("Continue:")
	if err := emitWrite(cmd, path, []byte(b.String())); err != nil {
		return err
	}
	return emitWrite(cmd, filepath.Join(dir, "INSTALL.md"), []byte(readme))
}

func emitGastown(cmd *cobra.Command, cwd string, srv agentpack.Server) error {
	dir := filepath.Join(cwd, ".rtmx", "emit", "gastown")
	path := filepath.Join(dir, "rtmx-plugin.json")
	body, err := json.MarshalIndent(map[string]interface{}{
		"name":        "rtmx",
		"version":     "1.0.0",
		"description": "RTMX requirements traceability plugin for Gastown",
		"transport": map[string]interface{}{
			"type":    "stdio",
			"command": srv.Command,
			"args":    srv.Args,
		},
		"capabilities": map[string]bool{"tools": true, "resources": true},
	}, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	readme := "# Gastown — RTMX emit\n\nUses shared-pack MCP stdio (`rtmx mcp-server --stdio`).\n"
	cmd.Println("Gastown:")
	if err := emitWrite(cmd, path, body); err != nil {
		return err
	}
	return emitWrite(cmd, filepath.Join(dir, "INSTALL.md"), []byte(readme))
}
