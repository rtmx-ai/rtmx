# Positioning one-liner (REQ-MONO-063h)

Canonical discovery copy — reuse verbatim on README, plugin descriptions,
site MCP docs, and registry/listing surfaces.

## One-liner

> Git-native requirements RTM for agents (next, claim, verify) — not another Jira/Linear MCP.

## What it means

| Phrase | Intent |
|--------|--------|
| Git-native requirements RTM | Local CSV-in-git traceability; durable intent in the repo |
| next, claim, verify | The agent loop MCP tools — not a ticket-board clone |
| not another Jira/Linear MCP | We are not a SaaS issue-tracker connector |

## What we do **not** claim

- Replacing Jira/Linear as a general project-management suite
- Vendoring the Go binary inside marketplace plugins (PATH / brew remains SoT)
- Default CLI phone-home / install telemetry

## Surfaces that must stay aligned

| Surface | File |
|---------|------|
| Canonical doc | this file |
| README agent section | `README.md` |
| Cursor plugin | `plugins/cursor-rtmx/.cursor-plugin/plugin.json` + README |
| Site MCP docs | `rtmx.ai` adapters/mcp |
| MCP Registry | `server.json` description (may append launch note) |
