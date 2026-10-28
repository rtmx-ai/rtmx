# Long-tail agent emitters (REQ-MONO-062l)

Beyond Claude / Cursor / Codex, `rtmx plugin emit` writes MCP configs for
agents already touched by `rtmx install`. All targets reuse the **shared
pack** command/args (`rtmx` + `mcp-server --stdio`).

## Agent → install / emit

| Agent | Prefer | Emit |
|-------|--------|------|
| Claude Code | `claude plugin marketplace add rtmx-ai/rtmx` + `install rtmx@rtmx` | `rtmx plugin emit --agents claude` |
| Cursor | Marketplace / deeplink ([CURSOR_MARKETPLACE.md](./CURSOR_MARKETPLACE.md)) | `rtmx plugin emit --agents cursor` |
| Codex | `codex mcp add rtmx -- rtmx mcp-server --stdio` | `rtmx plugin emit --agents codex` |
| Gemini CLI | emit extension JSON | `rtmx plugin emit --agents gemini` |
| Windsurf | emit MCP JSON | `rtmx plugin emit --agents windsurf` |
| Continue.dev | emit YAML fragment | `rtmx plugin emit --agents continue` |
| Gastown | emit plugin JSON | `rtmx plugin emit --agents gastown` |

```bash
rtmx plugin emit --list
rtmx plugin emit --agents gemini,windsurf,continue,gastown
rtmx plugin emit --all
```

Outputs land under `.rtmx/emit/<agent>/` (except Cursor → `.cursor/mcp.json`).
Prompt-only `rtmx install --agents …` remains additive; emit is the MCP path.
