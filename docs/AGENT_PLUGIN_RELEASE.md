# Agent plugin / marketplace release coupling (REQ-MONO-062o)

Marketplace plugins and the shared pack **track the CLI minor version**.

| CLI change | Pack / plugin version | Action |
|------------|----------------------|--------|
| Breaking MCP (rename/remove tool, incompatible schema) | **MAJOR** with CLI | Bump `plugins/rtmx/VERSION`, `pack.json`, Claude/Cursor `plugin.json`, marketplace manifests; validate; refresh listings |
| New MCP tool or additive optional fields; new CLI minor | **MINOR** with CLI | Same file bumps to match CLI minor |
| Pure CLI patch, no MCP name/schema change | Pack bump **optional** | No marketplace republish required |

Canonical contract: [MCP_PUBLIC_CONTRACT.md](./MCP_PUBLIC_CONTRACT.md).

## Release checklist (marketplace)

When cutting a CLI minor (or MCP-breaking major):

1. [ ] Bump `plugins/rtmx/VERSION` and `pack.json` `version` to the CLI minor.
2. [ ] Sync `internal/agentpack/mcp.json` with `plugins/rtmx/mcp.json` (parity test enforces).
3. [ ] Bump `plugins/claude-rtmx/.claude-plugin/plugin.json` and `.claude-plugin/marketplace.json` plugin version.
4. [ ] Bump `plugins/cursor-rtmx/.cursor-plugin/plugin.json` version.
5. [ ] Regenerate Cursor deeplink in `docs/CURSOR_MARKETPLACE.md` if launch args changed.
6. [ ] `claude plugin validate --strict ./plugins/claude-rtmx` and `claude plugin validate --strict .` (marketplace).
7. [ ] Run pack/plugin CI / `go test ./internal/agentpack/...`.
8. [ ] Note Marketplace re-index / Anthropic submit refresh if listed (062f/062h).
9. [ ] Run post-release marketplace smoke: `bash scripts/marketplace_smoke.sh`
   (also on `workflow_dispatch` of `.github/workflows/plugin-pack.yml`).
   Optional remote check: `RTMX_MARKETPLACE_URL=<raw marketplace.json URL> bash scripts/marketplace_smoke.sh`.
   Failures point back to this checklist (REQ-MONO-062r / 062o).

### When NOT to bump pack minor/major

- Documentation-only changes
- CLI bugfixes that do not change MCP tool names or input schemas
- Internal refactors with identical `tools/list` surface
