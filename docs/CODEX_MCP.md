# Codex MCP install (REQ-MONO-062k)

RTMX does not ship a fake Codex marketplace. Wire the same stdio MCP pack:

## One command

```bash
codex mcp add rtmx -- rtmx mcp-server --stdio
```

Verify: `codex mcp list`

Prerequisite: `rtmx` on `PATH` (`brew install rtmx-ai/tap/rtmx`).

## Emit / merge TOML

```bash
rtmx plugin emit --agents codex
# merge .rtmx/emit/codex/mcp_servers.rtmx.toml into ~/.codex/config.toml
```

```toml
[mcp_servers.rtmx]
command = "rtmx"
args = ["mcp-server", "--stdio"]
```

## Optional AGENTS.md nudge

Not installed by default:

```bash
rtmx plugin emit --agents codex --agents-md
```

## Legacy templates/codex tool JSON

The old shell-tool JSON under `templates/codex/` is obsolete relative to MCP.
`rtmx install --codex` now writes MCP install notes instead of that JSON.
