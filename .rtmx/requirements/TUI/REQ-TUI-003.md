# REQ-TUI-003: Requirement Detail Pane

## Metadata
- **Category**: TUI
- **Subcategory**: View
- **Priority**: HIGH
- **Phase**: 26
- **Status**: MISSING
- **Dependencies**: REQ-TUI-001, REQ-TUI-002, REQ-API-002
- **Blocks**: (none)

## Requirement

The TUI shall provide a requirement detail pane that displays full metadata,
requirement text, acceptance criteria, upstream/downstream dependencies with
their statuses, and test information for a selected requirement.

## Rationale

Browsing a table of requirements is only useful if you can drill into the
details of any single requirement. The detail pane answers "what exactly
does this require?", "what blocks it?", and "what tests cover it?" --
the three questions asked most often during triage and planning.

## Design

### Layout

The detail pane opens as a full-screen overlay when Enter is pressed on a
table row. It uses a scrollable viewport (bubbles/viewport) for content
that exceeds terminal height.

```
+-- REQ-MCP-007: Response Size Logging --+
| Status: MISSING          Priority: P0  |
| Category: MCP            Phase: 27     |
| Effort: 0.5 weeks       Assignee: --   |
| Sprint: --               Started: --    |
|                                         |
| REQUIREMENT                             |
| RTMX MCP server shall log response     |
| byte count and estimated token count... |
|                                         |
| DEPENDENCIES (upstream)                 |
|   [COMPLETE] REQ-MCP-003  Read tools   |
|   [COMPLETE] REQ-MCP-006  Stdio        |
|                                         |
| BLOCKS (downstream)                     |
|   [MISSING]  REQ-MCP-008  Filtering    |
|   [MISSING]  REQ-MCP-009  Size hints   |
|                                         |
| TEST                                    |
|   Module: .../mcp/server_test.go        |
|   Function: TestMCPResponseSizeLogging  |
|   Method: Integration Test              |
+------ Esc: back | e: edit | g: graph --+
```

### Navigation

| Key | Action |
|-----|--------|
| Esc | Return to table view |
| j/k | Scroll content |
| Enter (on dependency) | Jump to that requirement's detail |
| e | Open status/assignee edit (future: REQ-TUI-008) |
| g | Jump to graph view centered on this requirement |

### Data Source

Reads directly from the in-memory database and graph packages. No HTTP
call needed. Dependency resolution uses the same graph traversal as the
`rtmx deps` command.

## Acceptance Criteria

1. Enter on a table row opens the detail pane for that requirement.
2. All metadata fields are displayed with correct values.
3. Upstream and downstream dependencies are listed with status indicators.
4. Dependencies are navigable -- Enter on a dependency opens its detail.
5. Scrolling works for content taller than terminal height.
6. Esc returns to the table view with selection preserved.
7. Requirement text wraps correctly at terminal width.
8. Long notes and acceptance criteria are displayed without truncation.

## Files to Create/Modify

- `internal/tui/views/detail.go` -- Detail pane view model
- `internal/tui/views/detail_test.go` -- Render and navigation tests

## Effort Estimate

1 week

## Test Strategy

- Render test: verify all fields present for a known requirement
- Dependency list: verify upstream/downstream correctly resolved
- Navigation: Enter on dependency opens correct detail
- Scroll: content taller than viewport scrolls without crash
- Esc: returns to table with original selection index preserved
