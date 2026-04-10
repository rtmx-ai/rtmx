# REQ-PLUGIN-004: TUI Output Mode for Inline Agent Display

## Metadata
- **Category**: PLUGIN
- **Subcategory**: TUI
- **Priority**: HIGH
- **Phase**: 19
- **Status**: MISSING
- **Dependencies**: REQ-PLUGIN-001
- **Blocks**: (none)

## Requirement

RTMX shall provide a `--format=tui` output mode on status, backlog, health, next, and claims commands that emits compact, pre-formatted blocks designed for inline display within agent TUIs (Claude Code, gemini-cli, Cursor chat). This mode produces output optimized for embedding in a conversation, not for a standalone terminal.

## Rationale

Today, rtmx output is designed for a full terminal: wide tables, progress bars, color codes. When embedded inline in an agent conversation, this output is too verbose and can break formatting. A TUI mode produces compact, markdown-compatible output that looks good inside a chat bubble or conversation turn.

The TUI mode is the rendering layer that all plugin integrations (REQ-PLUGIN-001, 002, 003) use to format their responses.

## Design

### Format Specification

`--format=tui` produces compact, markdown-compatible output:

**rtmx status --format=tui:**
```
RTMX  82% complete  138/168 reqs
COMPLETE 138 | PARTIAL 12 | MISSING 18
Next unblocked: REQ-SYNC-003 (P0, 2wk)
```
(3 lines max for status)

**rtmx backlog --format=tui --limit 5:**
```
Open Requirements (18 remaining)
1. REQ-SYNC-003   P0    2wk  SYNC    MISSING
2. REQ-MCP-003    P0    1wk  MCP     MISSING
3. REQ-AGENT-001  P0    1wk  AGENT   MISSING
4. REQ-PLUGIN-001 P0    1wk  PLUGIN  MISSING
5. REQ-SYNC-005   HIGH  2wk  SYNC    MISSING
```
(Fixed-width aligned, no table borders)

**rtmx health --format=tui:**
```
Health: WARNING
  coverage: 82% (threshold 80%)  PASS
  consistency: 1 issue            WARN
  dependencies: no cycles         PASS
```

**rtmx next --show --format=tui:**
```
5 work webs (18 open reqs, 23 effort-weeks)
  SYNC      4 reqs  6wk  REQ-SYNC-003 -> 005 -> 007 -> 008
  MCP       2 reqs  3wk  REQ-MCP-003 -> MCP-005
  AGENT     7 reqs  8wk  REQ-AGENT-001 -> 002..007
  PLUGIN    4 reqs  4wk  REQ-PLUGIN-001 -> 002..004
  singles   3 reqs  2wk  REQ-DIST-004, REQ-E2E-003, REQ-GO-045
```

### Design Principles

1. **Compact**: 3-8 lines for any command. No full tables.
2. **Monospace-friendly**: aligned columns using spaces, no Unicode box drawing.
3. **Markdown-safe**: no characters that break markdown rendering.
4. **No color codes**: TUI mode never emits ANSI escapes (agent TUIs handle their own styling).
5. **Self-contained**: each output block is meaningful without surrounding context.

### Implementation

Add to `internal/output/`:
```go
type Format string
const (
    FormatDefault Format = ""
    FormatJSON    Format = "json"
    FormatTUI     Format = "tui"
)

func RenderTUI(data interface{}, template string) string
```

Each command checks the --format flag and delegates to the appropriate renderer.

## Acceptance Criteria

1. `--format=tui` is accepted on status, backlog, health, next, claims commands.
2. TUI output is <= 8 lines for any command.
3. No ANSI color codes in TUI output.
4. Output renders correctly in Claude Code, gemini-cli, and Cursor chat.
5. Output is valid markdown (can be copy-pasted).
6. Fixed-width alignment works in monospace fonts.
7. --format=tui --json is an error (mutually exclusive).

## Files to Create/Modify

- `internal/output/tui.go` -- TUI renderer
- `internal/output/tui_test.go` -- Output format tests
- `internal/cmd/status.go` -- Add --format flag
- `internal/cmd/backlog.go` -- Add --format flag
- `internal/cmd/health.go` -- Add --format flag
- `internal/cmd/next.go` -- Add --format flag (shared with AGENT reqs)

## Test Strategy

- Unit tests for TUI renderer with golden file comparisons
- Table-driven tests for each command output in TUI mode
- Verify no ANSI escape codes in TUI output
- Verify line count <= 8 for all commands
- Verify fixed-width alignment is consistent across varying data lengths
- Verify --format=tui and --json are mutually exclusive (returns error)
- Golden file tests for each command TUI output format
