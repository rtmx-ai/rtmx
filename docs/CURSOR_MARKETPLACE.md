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

## Official Marketplace submit package (REQ-MONO-063d)

Submit at https://cursor.com/marketplace/publish (Cursor reviews manually).
cursor.directory remains the interim live listing.

### Form field values

| Field | Value |
|-------|--------|
| Plugin path | `plugins/cursor-rtmx` (`.cursor-plugin/plugin.json` + root `mcp.json`) |
| Public repo | https://github.com/rtmx-ai/rtmx |
| Homepage | https://github.com/rtmx-ai/rtmx/tree/main/plugins/cursor-rtmx |
| Description | Git-native requirements RTM for agents (next, claim, verify) — not another Jira/Linear MCP. ([POSITIONING.md](./POSITIONING.md)) |
| Prerequisite | `rtmx` on `PATH` — `brew install rtmx-ai/tap/rtmx` (or scoop / go install) |
| MCP launch | `rtmx` + `["mcp-server", "--stdio"]` (shared pack; do not vendor binary) |
| Deeplink (works now) | https://cursor.com/link/mcp?name=rtmx&config=eyJjb21tYW5kIjoicnRteCIsImFyZ3MiOlsibWNwLXNlcnZlciIsIi0tc3RkaW8iXX0%3D |

### Publish checklist

| Step | Status |
|------|--------|
| Plugin at `plugins/cursor-rtmx` with `.cursor-plugin/plugin.json` | Ready |
| `mcp.json` matches shared pack stdio | Ready |
| README documents install + PATH | Ready |
| Public Git host (`rtmx-ai/rtmx`) | Ready (sync via REQ-MONO-063e) |
| Form field values above frozen | Ready |
| Form submitted / listing live | **Human:** file at publish URL; listing may lag |
| cursor.directory | **Live:** https://cursor.directory/plugins/rtmx |

### Submission record

| Date (UTC) | Action |
|------------|--------|
| 2026-10-28 | Deeplink + docs; cursor.directory published |
| 2026-10-29 | Official submit package frozen (063d). Human: open https://cursor.com/marketplace/publish and paste field values; append row here when filed. |

Do not remove deeplink or cursor.directory links while official review lags.

## Site / docs deep-link (062m)

Prefer linking the **https://cursor.com/link/mcp?…** URL on rtmx.ai and README so browsers hand off to Cursor. Keep the `cursor://` form for in-product docs.

Discovery attribution (UTMs + Plausible): [DISCOVERY_FUNNEL.md](./DISCOVERY_FUNNEL.md) (REQ-MONO-063a).

## Managed Sync (multi-agent)

After MCP tools load, join a live room with the **same** stdio pack — no
second JSON: [rtmx.ai/sync](https://rtmx.ai/sync) → `rtmx login` → `rtmx sync`.
Details: [SYNC_MARKETPLACE_COHESION.md](./SYNC_MARKETPLACE_COHESION.md) (REQ-MONO-063g).
