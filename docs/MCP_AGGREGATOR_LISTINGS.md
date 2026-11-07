# MCP aggregator listings playbook (REQ-MONO-063m)

Submit RTMX to major aggregators **after** the official registry package
([MCP_REGISTRY.md](./MCP_REGISTRY.md)) and cursor.directory listing are ready.
Always point at brew PATH + shared pack stdio — never a forked launch.

## Canonical pointers

| Asset | URL |
|-------|-----|
| Official registry | `io.github.rtmx-ai/rtmx` via [MCP_REGISTRY.md](./MCP_REGISTRY.md) |
| cursor.directory | https://cursor.directory/plugins/rtmx |
| Public repo | https://github.com/rtmx-ai/rtmx |
| One-liner | [POSITIONING.md](./POSITIONING.md) |

## Aggregators

### PulseMCP

- Site: https://www.pulsemcp.com/ (or current submit / suggest form on that domain)
- Fields: name `RTMX`, description from POSITIONING, install =
  `brew install rtmx-ai/tap/rtmx` + `rtmx mcp-server --stdio`, repo URL,
  registry name if supported
- Notes: prefer linking registry + GitHub over vendored binaries

### Glama

- Site: https://glama.ai/mcp (servers directory / submit flow)
- Fields: same pack command/args; homepage rtmx.ai or GitHub; tags
  requirements, traceability, mcp

### awesome-mcp (or awesome-mcp-servers)

- Repo: search GitHub `awesome-mcp` / `awesome-mcp-servers` (community list)
- PR: one-line entry with name, short POSITIONING blurb, link to
  `github.com/rtmx-ai/rtmx` and cursor.directory

## Submission record

| Date (UTC) | Aggregator | Action |
|------------|------------|--------|
| 2026-10-29 | — | Playbook frozen (063m). Human: file PulseMCP + Glama; append rows. |
