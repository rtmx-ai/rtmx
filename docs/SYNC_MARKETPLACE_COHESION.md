# Managed Sync + marketplace cohesion (REQ-MONO-063g)

Marketplace / deeplink install loads the **same** stdio MCP pack
(`rtmx mcp-server --stdio`). Multi-agent collaboration then uses Managed
Sync — not a second MCP JSON.

## Path

1. Install CLI: `brew install rtmx-ai/tap/rtmx` (PATH prerequisite).
2. Install agent plugin / deeplink (Claude marketplace, Cursor deeplink, or
   `rtmx bootstrap agents`).
3. Local loop: MCP `next` → `claim` → `verify` against CSV-in-git.
4. Multiplayer room: [rtmx.ai/sync](https://rtmx.ai/sync) → `rtmx login` →
   `rtmx sync` (org room). Plans: [rtmx.ai/pricing](https://rtmx.ai/pricing).

No secrets belong in the plugin pack. Sync auth is CLI login / token env —
never baked into `mcp.json`.

## Non-goals

- Divergent MCP command/args for Sync vs OSS
- Embedding API keys in marketplace plugins
