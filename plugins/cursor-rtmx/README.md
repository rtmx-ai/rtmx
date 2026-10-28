# RTMX Cursor plugin

Cursor Plugin format (`.cursor-plugin/plugin.json` + root `mcp.json`) wrapping
the shared pack at `../rtmx/`. MCP launch must stay identical to
`../rtmx/mcp.json`: `rtmx mcp-server --stdio` with `rtmx` on `PATH`.

## Prerequisite

```bash
brew install rtmx-ai/tap/rtmx   # macOS / Linux
scoop install rtmx             # Windows
```

## Local install (dogfood)

```bash
mkdir -p ~/.cursor/plugins/local
ln -sfn "$(pwd)/plugins/cursor-rtmx" ~/.cursor/plugins/local/rtmx
# Developer: Reload Window — then Customize → confirm plugin MCP `rtmx`
```

On Enterprise, admins must allow local plugin imports.

## One-click MCP deeplink (REQ-MONO-062h)

```text
cursor://anysphere.cursor-deeplink/mcp/install?name=rtmx&config=eyJjb21tYW5kIjoicnRteCIsImFyZ3MiOlsibWNwLXNlcnZlciIsIi0tc3RkaW8iXX0%3D
```

HTTPS hand-off: https://cursor.com/link/mcp?name=rtmx&config=eyJjb21tYW5kIjoicnRteCIsImFyZ3MiOlsibWNwLXNlcnZlciIsIi0tc3RkaW8iXX0%3D

Full publish checklist + regenerate script: [docs/CURSOR_MARKETPLACE.md](../../docs/CURSOR_MARKETPLACE.md).

cursor.directory (live): https://cursor.directory/plugins/rtmx  

Official Marketplace: https://cursor.com/marketplace/publish (review may lag; deeplink works now). See [docs/COMMUNITY_LISTINGS.md](../../docs/COMMUNITY_LISTINGS.md).
