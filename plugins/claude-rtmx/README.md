# RTMX Claude Code plugin

Wraps the shared agent-plugin pack (`../rtmx/`) for Claude Code:
`rtmx mcp-server --stdio` plus `rtmx-loop` and `rtmx-prerequisite` skills.

## Prerequisite

`rtmx` must be on `PATH`:

```bash
brew install rtmx-ai/tap/rtmx   # macOS / Linux
scoop install rtmx             # Windows
```

## Install

```bash
claude plugin marketplace add rtmx-ai/rtmx
claude plugin install rtmx@rtmx
```

Session: `/plugin marketplace add rtmx-ai/rtmx` then `/plugin install rtmx@rtmx`.

Local dogfood from a checkout: see [docs/CLAUDE_MARKETPLACE.md](../../docs/CLAUDE_MARKETPLACE.md).

```bash
claude plugin validate --strict ./plugins/claude-rtmx
```

MCP command/args must stay identical to `../rtmx/mcp.json` (source of truth).
Requires `rtmx` on `PATH`.
