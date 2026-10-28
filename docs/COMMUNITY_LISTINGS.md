# Community listings (REQ-MONO-062i)

Interim discovery while Cursor Marketplace / Anthropic official review lag.
Primary install docs remain [CURSOR_MARKETPLACE.md](./CURSOR_MARKETPLACE.md)
and [CLAUDE_MARKETPLACE.md](./CLAUDE_MARKETPLACE.md).

## Live listing (canonical public pointer)

**This document** is the canonical community listing text, published in the
rtmx repo and linked from the Cursor plugin README and site MCP docs.

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

## cursor.directory form

Submit at https://cursor.directory/mcp/new (or `/plugins/new` for Open
Plugins auto-detect of `plugins/cursor-rtmx`).

Paste:

- **Name:** RTMX  
- **Description:** Requirements traceability MCP (next, claim, verify). Requires `rtmx` on PATH via Homebrew `rtmx-ai/tap/rtmx`.  
- **Cursor Deep Link:** (HTTPS hand-off above)  
- **Install instructions URL:** https://github.com/rtmx-ai/rtmx/blob/main/docs/CURSOR_MARKETPLACE.md  

| Date (UTC) | Action |
|------------|--------|
| 2026-10-28 | Listing text + brew/PATH prerequisite published; cursor.directory form fields ready. Directory UI submit is account-gated (listing may lag). |

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
