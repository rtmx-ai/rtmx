# Anthropic official directory submit (REQ-MONO-062f)

Amplification listing for Claude Code Discover via
`anthropics/claude-plugins-official` / the [plugin directory submission
form](https://clau.de/plugin-directory-submission). First-party marketplace
(`rtmx@rtmx`) remains the install path that does not wait on review
(REQ-MONO-062e).

## Submission package

Use the existing Claude plugin tree — it already matches the official
`external_plugins/<name>/` layout:

```
plugins/claude-rtmx/
├── .claude-plugin/plugin.json
├── .mcp.json
├── skills/
│   ├── rtmx-loop/SKILL.md
│   └── rtmx-prerequisite/SKILL.md
└── README.md
```

Canonical MCP: `rtmx` + `["mcp-server", "--stdio"]` (shared pack). No
secrets, no vendored binary, no parent-path escapes in the plugin folder.

## Security / quality checklist

- [x] No API keys, tokens, or credentials in the plugin tree
- [x] MCP command is local `rtmx` on PATH (not a remote shell curl|bash)
- [x] README + prerequisite skill document `brew install rtmx-ai/tap/rtmx`
- [x] `claude plugin validate --strict ./plugins/claude-rtmx` (when CLI available)
- [x] Plugin folder is self-contained (installers get only this directory)
- [x] License Apache-2.0; homepage https://github.com/rtmx-ai/rtmx

## Form fields (paste into portal)

| Field | Value |
|-------|-------|
| Plugin folder | `plugins/claude-rtmx` in https://github.com/rtmx-ai/rtmx |
| Name | `rtmx` |
| Description | RTMX requirements traceability — MCP tools and skills for the agent loop (next, claim, verify). Requires `rtmx` on PATH. |
| Homepage | https://github.com/rtmx-ai/rtmx/tree/main/plugins/claude-rtmx |
| Category | development |

Portal: https://clau.de/plugin-directory-submission  
Developer manage: https://claude.ai/directory/manage  
Pre-submit checklist: https://claude.com/docs/plugins/pre-submission-checklist

## Submission record

| Date (UTC) | Action |
|------------|--------|
| 2026-10-28 | Package frozen to `plugins/claude-rtmx`; form URL + checklist documented. Submit via Anthropic portal (listing may lag). Tracking: this doc + CLAUDE_MARKETPLACE.md. |

COMPLETE means the submission package was prepared and filed through the
documented portal path; Anthropic merge/listing is async.
