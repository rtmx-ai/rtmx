# REQ-AGENT-002: Claude Code Hooks Integration

## Metadata
- **Category**: ADAPT
- **Subcategory**: Claude
- **Priority**: P0
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-029
- **Blocks**: REQ-GO-047

## Requirement

`rtmx install --claude` shall install Claude Code hooks that automatically inject RTM context into conversations, matching the Python CLI's REQ-CLAUDE-001 implementation.

## Design

### Hook Installation

```bash
rtmx install --claude          # Install Claude Code hooks
rtmx install --claude --remove # Remove hooks
```

Creates `.claude/hooks.json`:
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": ".*",
        "command": "rtmx context --format claude"
      }
    ]
  }
}
```

### Context Command

```bash
rtmx context                    # Token-efficient RTM summary
rtmx context --format claude    # Claude Code hook format
rtmx context --format plain     # Plain text
```

Output: concise RTM status suitable for LLM context injection (< 500 tokens).

## Acceptance Criteria

1. `rtmx install --claude` creates `.claude/hooks.json`
2. `rtmx context` outputs token-efficient RTM summary
3. Hook fires automatically in Claude Code conversations
4. `--remove` cleanly uninstalls hooks
5. Context output includes: completion %, top blockers, quick wins

## Files to Create/Modify

- `internal/cmd/install.go` - Add `--claude` flag
- `internal/cmd/context.go` - New context command
- `internal/cmd/context_test.go` - Context output tests
