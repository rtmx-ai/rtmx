# Dogfood marketplace install (REQ-MONO-062p)

Contributors and the RTMX monorepo use the **same** agent install path as
end users: marketplace / deeplink first, `rtmx plugin emit` when offline.

## Contributor onboarding

See [CONTRIBUTING.md](../CONTRIBUTING.md) § Agent / MCP setup.

| Path | When |
|------|------|
| `claude plugin marketplace add` + `install rtmx@rtmx` | Claude Code with network |
| Cursor Marketplace / deeplink | Cursor with network |
| `rtmx plugin emit --agents …` | Offline, air-gapped, or CI |
| `codex mcp add rtmx -- rtmx mcp-server --stdio` | Codex |

`rtmx install --agents …` remains for **prompt** injection and hooks. Its
help text points at marketplace/emit for MCP. Do not treat prompt install as
the primary MCP distribution mode.

## Init tip

`rtmx init` prints marketplace + emit next steps after creating `.rtmx/`.

## Project gravity

Root [AGENTS.md](../AGENTS.md) and plugin skills tell agents: when `.rtmx/`
exists and MCP is configured, prefer RTMX tools over shell.
