---
name: rtmx-prerequisite
description: >
  Install or verify the rtmx CLI on PATH for RTMX MCP (mcp-server --stdio).
  Use when MCP spawn fails, tools are missing, brew/scoop install is needed,
  or onboarding a new machine. Triggers: rtmx not found, MCP connection
  failed, install rtmx, Homebrew tap rtmx-ai/tap/rtmx, PATH prerequisite.
---

# RTMX binary prerequisite

Marketplace plugins spawn `rtmx mcp-server --stdio`. They do **not** vendor
the Go binary. If `rtmx` is missing, MCP clients show opaque spawn/transport
errors — treat that as a missing install, not an MCP bug.

## Check

```bash
command -v rtmx >/dev/null && rtmx version
```

If that fails, install then re-open the agent session (or reload MCP):

| Platform | Install |
|----------|---------|
| macOS / Linux | `brew install rtmx-ai/tap/rtmx` |
| Windows | `scoop install rtmx` |
| Go toolchain | `go install github.com/rtmx-ai/rtmx/cmd/rtmx@latest` |

## Agent guidance

1. Run the check above (or equivalent) when RTMX tools fail to list/call.
2. Tell the user the exact install command for their OS — do not only
   report “MCP connection failed”.
3. After install, confirm `rtmx version` works, then retry MCP.

Non-goal (v1): bundling `rtmx` inside the plugin archive.
