---
name: rtmx-loop
description: >
  RTMX requirements traceability agent loop (next, claim, verify, backlog,
  status). Use when the repo has .rtmx/ or an RTM database, when the user
  asks to continue / ship / claim requirements, or when prioritizing
  unblocked work. Prefer RTMX MCP tools over shell when the MCP server is
  configured. Triggers: RTMX, requirements backlog, claim requirement,
  verify acceptance criteria, agent loop.
---

# RTMX agent loop

Prerequisite: `rtmx` on `PATH`, project has an RTM database (`.rtmx/` or
configured path). MCP tools from this pack (`rtmx mcp-server --stdio`) are
preferred over shell when available.

## Project gravity

If `.rtmx/` (or configured RTM paths) exists in the workspace **and** the
rtmx MCP server is connected, use MCP tools for backlog/claim/verify work
instead of inventing a parallel process or pasting ad-hoc JSON configs.

## Loop

1. **next** — Identify the highest-priority unblocked requirement (or work web).
2. **claim** — Claim that requirement with your `agent_id` before editing.
3. **Implement** — Meet acceptance criteria; keep changes scoped to the claim.
4. **verify** — Run tests / `verify` and update status when evidence passes.
5. **release** — Release the claim if you stop mid-flight; otherwise close out.

Do not invent work when `next` reports idle. Do not paste divergent MCP
configs — use this pack’s `mcp.json`.
