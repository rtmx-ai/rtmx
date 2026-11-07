# Team Default-On playbook (REQ-MONO-063l)

One-pager for Sync customers who want every developer’s agents to pull RTMX
by default. Derived from [COMMUNITY_LISTINGS.md](./COMMUNITY_LISTINGS.md).

## Shared rules

1. **`rtmx` on PATH** on every machine (`brew install rtmx-ai/tap/rtmx`, scoop,
   or MDM). Plugins do **not** vendor the binary.
2. **One MCP pack** — always `rtmx mcp-server --stdio`. Do not paste divergent
   JSON per agent.
3. Prefer marketplace / deeplink online; `rtmx plugin emit` or
   `rtmx bootstrap agents` offline.

## Steps

### 1. Claude Code

1. Org runbook: `claude plugin marketplace add rtmx-ai/rtmx`
2. `claude plugin install rtmx@rtmx` (or team allowlist when available)
3. Confirm MCP tools `next` / `claim` appear in a sample repo with `.rtmx/`

### 2. Cursor

1. Enterprise: allow Marketplace plugin or pin deeplink /
   `plugins/cursor-rtmx` local path
2. Enable team MCP Default-On when Cursor policy supports it
3. Deeplink: see [CURSOR_MARKETPLACE.md](./CURSOR_MARKETPLACE.md)

### 3. Codex (and long-tail)

1. Bootstrap: `codex mcp add rtmx -- rtmx mcp-server --stdio`
2. Or `rtmx plugin emit --agents codex` / `rtmx bootstrap agents`
3. Document in contributor `AGENTS.md` via `--agents-md` when appropriate

## Sync rooms

After Default-On, multi-agent rooms: [SYNC_MARKETPLACE_COHESION.md](./SYNC_MARKETPLACE_COHESION.md)
(`rtmx login` → `rtmx sync`).
