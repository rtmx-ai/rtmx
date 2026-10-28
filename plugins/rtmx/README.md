# RTMX agent-plugin pack

**Canonical source of truth** for marketplace plugins and `rtmx plugin emit`.

| File | Role |
|------|------|
| `mcp.json` | Only MCP launch config: `rtmx` + `["mcp-server", "--stdio"]` |
| `VERSION` / `pack.json` | Pack version (tracks CLI **minor**; see `docs/MCP_PUBLIC_CONTRACT.md`) |
| `skills/rtmx-loop/` | Agent loop skill (next → claim → verify) |

## Rule

Claude wrappers, Cursor wrappers, Codex emit, and long-tail emitters MUST
copy or reference **this** `mcp.json`. Do not invent a second command/args
pair for the same server.

## Binary prerequisite (PATH)

Plugins spawn `rtmx`; they do **not** bundle the binary (v1 non-goal).

| Platform | Install |
|----------|---------|
| macOS / Linux | `brew install rtmx-ai/tap/rtmx` |
| Windows | `scoop install rtmx` |
| Go toolchain | `go install github.com/rtmx-ai/rtmx/cmd/rtmx@latest` |

If MCP fails with a spawn/transport error, run `command -v rtmx` first.
See skill `skills/rtmx-prerequisite/` for agent-facing failure guidance.

## Version

`VERSION` is the pack semver. Bump the minor when the CLI minor releases
with MCP surface changes; coordinate MAJOR with breaking MCP (062a / 062o).
