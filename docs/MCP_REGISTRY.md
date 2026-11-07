# Official MCP Registry (REQ-MONO-063c)

Publish RTMX metadata to
[registry.modelcontextprotocol.io](https://registry.modelcontextprotocol.io)
so aggregators and clients can discover the same stdio launch as the
marketplace pack: `rtmx mcp-server --stdio`.

## Artifacts

| File | Role |
|------|------|
| [`server.json`](../server.json) | Registry metadata (`io.github.rtmx-ai/rtmx`) |
| [`npm/package.json`](../npm/package.json) | `mcpName` ownership proof for npm |
| [`plugins/rtmx/mcp.json`](../plugins/rtmx/mcp.json) | Canonical command/args (must stay identical) |

Namespace uses GitHub org auth: `io.github.rtmx-ai/rtmx`.

Prerequisite for users: `rtmx` on `PATH` (`brew install rtmx-ai/tap/rtmx`
or the npm postinstall binary). The registry hosts **metadata**, not the Go
binary.

## Publish procedure

1. Ensure `server.json` `packageArguments` / launch match `plugins/rtmx/mcp.json`.
2. Bump `server.json` `version` with the agent-plugin pack minor when MCP surface
   changes (see [AGENT_PLUGIN_RELEASE.md](./AGENT_PLUGIN_RELEASE.md)).
3. If the npm `version` in `server.json` is new, publish npm first with
   `mcpName` set:

   ```bash
   cd npm && npm publish --access public
   ```

4. Install publisher: `brew install mcp-publisher` (or release binary).
5. Authenticate as a member of `rtmx-ai`:

   ```bash
   mcp-publisher login github
   ```

6. From the `rtmx/` repo root (where `server.json` lives):

   ```bash
   mcp-publisher publish
   ```

7. Verify:

   ```bash
   curl -fsSL 'https://registry.modelcontextprotocol.io/v0.1/servers?search=io.github.rtmx-ai/rtmx'
   ```

## Submission record

| Date (UTC) | Action |
|------------|--------|
| 2026-10-29 | `server.json` + `mcpName` + this runbook frozen. Human: `mcp-publisher login github` then `publish` (listing may lag / registry preview). |

## Non-goals

- Vendoring the Go binary inside an MCPB for v1 (PATH + brew remains SoT).
- Divergent command/args vs marketplace plugins.
