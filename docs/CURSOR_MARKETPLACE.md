# Cursor Marketplace and MCP deeplink

**REQ-MONO-062h** — one-click install for the open-source RTMX client in Cursor.

Plugin package: [`plugins/cursor-rtmx/`](../plugins/cursor-rtmx/)  
Shared MCP SoT: [`plugins/rtmx/mcp.json`](../plugins/rtmx/mcp.json)

## One-click MCP deeplink (ships now)

Config: `{"command":"rtmx","args":["mcp-server","--stdio"]}`

```text
cursor://anysphere.cursor-deeplink/mcp/install?name=rtmx&config=eyJjb21tYW5kIjoicnRteCIsImFyZ3MiOlsibWNwLXNlcnZlciIsIi0tc3RkaW8iXX0%3D
```

Web redirect (opens Cursor):

```text
https://cursor.com/link/mcp?name=rtmx&config=eyJjb21tYW5kIjoicnRteCIsImFyZ3MiOlsibWNwLXNlcnZlciIsIi0tc3RkaW8iXX0%3D
```

Prerequisite: `rtmx` on `PATH` (`brew install rtmx-ai/tap/rtmx`).

Regenerate after MCP launch changes:

```bash
python3 - <<'PY'
import json, base64, urllib.parse
cfg = json.load(open("plugins/rtmx/mcp.json"))["mcpServers"]["rtmx"]
payload = {k: cfg[k] for k in ("command", "args") if k in cfg}
b64 = urllib.parse.quote(base64.b64encode(json.dumps(payload, separators=(",", ":")).encode()).decode(), safe="")
print(f"cursor://anysphere.cursor-deeplink/mcp/install?name=rtmx&config={b64}")
PY
```

## Local plugin dogfood

```bash
ln -sfn "$(pwd)/plugins/cursor-rtmx" ~/.cursor/plugins/local/rtmx
# Developer: Reload Window → Customize
```

## Marketplace publish checklist

Submit at https://cursor.com/marketplace/publish (Cursor reviews manually).

| Step | Status |
|------|--------|
| Plugin at `plugins/cursor-rtmx` with `.cursor-plugin/plugin.json` | Ready |
| `mcp.json` matches shared pack stdio | Ready |
| README documents install + PATH | Ready |
| Public Git host (`rtmx-ai/rtmx` after subtree push) | Ready for submit |
| Form submitted / listing live | Pending review — deeplink is the interim auto-pull |

Notes: 2026-10-28 — deeplink + docs COMPLETE for install; Marketplace form submit/listing may lag (not a blocker per 062h OR clause). Site CTAs consume these URLs in REQ-MONO-062m.

## Site / docs deep-link (062m)

Prefer linking the **https://cursor.com/link/mcp?…** URL on rtmx.ai and README so browsers hand off to Cursor. Keep the `cursor://` form for in-product docs.
