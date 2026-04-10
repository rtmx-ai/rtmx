# REQ-PLUGIN-003: Universal Agent Integration (Cursor, Coder, Codex, Gastown)

## Metadata
- **Category**: PLUGIN
- **Subcategory**: Universal
- **Priority**: MEDIUM
- **Phase**: 21
- **Status**: MISSING
- **Dependencies**: REQ-MCP-003, REQ-MCP-005
- **Blocks**: (none)

## Requirement

RTMX shall provide integration paths for Cursor, Coder, OpenAI Codex, and Steve Yegge's Gastown project, using the MCP server as the primary interface and thin adapters where platform-specific integration is needed.

## Rationale

Different agent platforms have different integration models. Rather than building deep integrations for each, RTMX should provide:
1. MCP server (works for MCP-speaking platforms: Cursor, Claude Code)
2. CLI with --json (works for any platform that can shell out)
3. HTTP API (works for remote/networked agents: Coder, Gastown)
4. Platform-specific adapters only where the MCP/CLI/HTTP path is insufficient

## Design

### Integration Matrix

| Platform | Transport | Integration Level | Notes |
|----------|-----------|-------------------|-------|
| Claude Code | MCP (stdio) | Native skill pack | REQ-PLUGIN-001 |
| Cursor | MCP (stdio) | MCP tool provider | Cursor natively speaks MCP; register as tool |
| gemini-cli | MCP or CLI | Extension | REQ-PLUGIN-002 |
| Coder | HTTP/SSE | Remote MCP or REST | Coder workspaces are remote; need HTTP transport |
| Codex | CLI | Shell-out | Codex runs commands; rtmx CLI with --json suffices |
| Gastown | HTTP/SSE or MCP | Plugin/tool provider | Depends on Gastown's plugin architecture |

### Cursor Integration

Cursor supports MCP servers via `.cursor/mcp.json`. Installation:
```
rtmx install --cursor
```
Writes MCP server config to `.cursor/mcp.json`. Cursor discovers RTMX tools automatically.

### Coder Integration

Coder workspaces are remote. The RTMX MCP server runs inside the workspace and exposes an HTTP/SSE endpoint. A Coder template can include RTMX setup:
```
rtmx install --coder
```
Writes a Coder module or script that starts the MCP server on workspace creation.

### Codex Integration

OpenAI Codex can run shell commands. No special integration needed beyond ensuring `rtmx` is on PATH and all commands support `--json`. Document the Codex integration pattern:
```
rtmx install --codex
```
Writes a system prompt snippet or tool definition for Codex to use.

### Gastown Integration

Depends on Gastown's architecture (to be investigated). Likely either:
- MCP tool provider (if Gastown speaks MCP)
- HTTP plugin (if Gastown has a plugin registry)
- CLI wrapper (if Gastown shells out to tools)

```
rtmx install --gastown
```

### install Command Architecture

`rtmx install` becomes the universal integration entry point:
```
rtmx install --claude-code    # Claude Code skill pack
rtmx install --gemini-cli     # gemini-cli extension
rtmx install --cursor         # Cursor MCP config
rtmx install --coder          # Coder workspace template
rtmx install --codex          # Codex tool definition
rtmx install --gastown        # Gastown plugin (TBD)
rtmx install --list           # Show available integrations and status
```

Each flag writes platform-specific configuration files from templates in `templates/<platform>/`.

## Acceptance Criteria

1. `rtmx install --cursor` writes a valid .cursor/mcp.json.
2. `rtmx install --coder` generates workspace setup for HTTP MCP server.
3. `rtmx install --codex` generates a Codex-compatible tool definition.
4. `rtmx install --list` shows all available integrations and whether they're installed.
5. All integrations use the same underlying MCP server or CLI.
6. HTTP/SSE transport works for remote agents.
7. Documentation for each integration path.

## Files to Create/Modify

- `internal/cmd/install.go` -- Platform flags and template rendering
- `internal/cmd/install_test.go` -- Tests for each platform
- `templates/cursor/mcp.json` -- Cursor MCP config
- `templates/coder/` -- Coder workspace templates
- `templates/codex/` -- Codex tool definition
- `templates/gastown/` -- Gastown plugin (placeholder)
- `docs/integrations/` -- Per-platform setup guides

## Test Strategy

- Unit test per platform: verify correct config file generation
- Golden file tests: generated configs match expected templates
- Integration test: `rtmx install --list` detects installed integrations
- Idempotency test: running install twice does not corrupt config
- Error handling: missing directories, permission errors, existing conflicting configs
