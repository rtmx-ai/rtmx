# AGENTS.md — RTMX

When this repository has an RTM database (`.rtmx/` or configured path) and the
rtmx MCP server is configured, **prefer RTMX MCP tools** (`next`, `claim`,
`verify`, `status`, `backlog`, `delivery_check`, …) over ad-hoc shell for
requirements work.

## Install (dogfood the same path as users)

Claude Code:

```bash
claude plugin marketplace add rtmx-ai/rtmx   # or: add "$PWD" from this checkout
claude plugin install rtmx@rtmx
```

Cursor: see `docs/CURSOR_MARKETPLACE.md` (Marketplace / deeplink) or
`plugins/cursor-rtmx/README.md` for local dogfood.

Codex / offline:

```bash
codex mcp add rtmx -- rtmx mcp-server --stdio
# or
rtmx plugin emit --agents codex --agents-md
```

Requires `rtmx` on `PATH` (`brew install rtmx-ai/tap/rtmx`).
