# Community listings (REQ-MONO-062i)

Interim discovery while Cursor Marketplace / Anthropic official review lag.
Primary install docs remain [CURSOR_MARKETPLACE.md](./CURSOR_MARKETPLACE.md)
and [CLAUDE_MARKETPLACE.md](./CLAUDE_MARKETPLACE.md).

## Live listing

**cursor.directory (live):** https://cursor.directory/plugins/rtmx  

MCP payload on that page matches the shared pack (`rtmx` + `mcp-server --stdio`).
Requires `rtmx` on PATH (`brew install rtmx-ai/tap/rtmx`).

| Field | Value |
|-------|-------|
| Name | RTMX |
| Description | Git-native requirements traceability MCP: next/claim/verify agent loop. Not a Jira clone — durable local RTM + verify gates. |
| MCP launch | `rtmx mcp-server --stdio` |
| Prerequisite | `brew install rtmx-ai/tap/rtmx` (or scoop / go install) — **rtmx on PATH** |
| Cursor deeplink | `cursor://anysphere.cursor-deeplink/mcp/install?name=rtmx&config=eyJjb21tYW5kIjoicnRteCIsImFyZ3MiOlsibWNwLXNlcnZlciIsIi0tc3RkaW8iXX0%3D` |
| HTTPS hand-off | https://cursor.com/link/mcp?name=rtmx&config=eyJjb21tYW5kIjoicnRteCIsImFyZ3MiOlsibWNwLXNlcnZlciIsIi0tc3RkaW8iXX0%3D |
| Claude marketplace | `claude plugin marketplace add rtmx-ai/rtmx` then `claude plugin install rtmx@rtmx` |
| Install docs | https://github.com/rtmx-ai/rtmx/blob/main/docs/CURSOR_MARKETPLACE.md |
| Plugin path | https://github.com/rtmx-ai/rtmx/tree/main/plugins/cursor-rtmx |

| Date (UTC) | Action |
|------------|--------|
| 2026-10-28 | Published on cursor.directory: https://cursor.directory/plugins/rtmx |

## Team / Default-On playbook (Sync customers)

For orgs on RTMX Sync that want agents to pull RTMX by default:

1. Document in the org runbook: install `rtmx` via brew/scoop on every
   developer machine (or ship via MDM).
2. Claude: add first-party marketplace + `rtmx@rtmx` in the team Claude
   Code setup guide (or org plugin allowlist when available).
3. Cursor: Enterprise admins allow the Marketplace plugin or pin the
   deeplink / local plugin path; enable team MCP Default-On when the
   Cursor team policy supports it.
4. Codex / long-tail: bake `rtmx plugin emit --all` (or `codex mcp add`)
   into the contributor bootstrap script.
5. Do **not** paste divergent MCP JSON — always the shared pack command/args.
