# REQ-TUI-005: Kanban Board View

## Metadata
- **Category**: TUI
- **Subcategory**: View
- **Priority**: HIGH
- **Phase**: 26
- **Status**: MISSING
- **Dependencies**: REQ-TUI-001, REQ-API-005
- **Blocks**: (none)

## Requirement

The TUI shall provide a Kanban board view with status-based columns
(NOT_STARTED, MISSING, PARTIAL, COMPLETE) where requirements are displayed
as cards that can be moved between columns via keyboard shortcuts.

## Rationale

Kanban boards are the standard project management visualization for tracking
work in progress. A terminal-based Kanban board lets CLI-first users manage
their backlog without context-switching to a web tool. Moving cards between
columns (status transitions) is the most common mutation in daily planning.

## Design

### Layout

```
+-- NOT STARTED --+  +--- MISSING ----+  +--- PARTIAL ----+  +-- COMPLETE ----+
| REQ-TUI-001     |  | REQ-MCP-007    |  |                |  | REQ-CLI-001    |
| TUI Framework   |  | Size Logging   |  |                |  | Static Binary  |
| P0  2.0w        |  | P0  0.5w       |  |                |  | P0  1.0w       |
|                  |  |                |  |                |  |                |
| REQ-TUI-002     |  | REQ-MCP-008    |  |                |  | REQ-CLI-002    |
| Req Table       |  | Filtering      |  |                |  | Config Load    |
| P0  1.0w        |  | P0  1.5w [B]   |  |                |  | P0  0.5w       |
+------------------+  +----------------+  +----------------+  +----------------+
```

`[B]` indicator shows the card is blocked by upstream dependencies.

### Column Layout

Columns divide the terminal width equally. Each column scrolls independently.
A focus indicator (bold border) shows which column is active.

### Card Content

Each card shows:
- req_id (bold)
- Truncated requirement text (first line)
- Priority badge and effort estimate
- `[B]` blocked indicator when applicable
- Category tag (dimmed)

### Navigation and Actions

| Key | Action |
|-----|--------|
| h/l or Left/Right | Move focus between columns |
| j/k or Up/Down | Move selection within column |
| Enter | Open detail pane for selected card |
| m | Move card: prompts for target status, updates requirement |
| Space | Toggle card selection for bulk operations |
| v | Filter by version/sprint |
| c | Filter by category |

### Status Transitions via Move

Pressing `m` on a card shows a status picker overlay. Selecting a target
status updates the requirement's status field and moves the card to the
new column. This calls the same `db.Update()` path as REQ-API-003.

Blocked cards cannot be moved to COMPLETE (enforced with an error message
explaining which dependencies are incomplete).

## Acceptance Criteria

1. Four columns displayed for each status value.
2. Cards show req_id, text, priority, effort, and blocked indicator.
3. Navigation moves between columns and within columns.
4. `m` moves a card to a new status column.
5. Blocked requirements cannot be moved to COMPLETE.
6. Card count per column shown in column header.
7. Each column scrolls independently.
8. Filter by version limits cards to a specific sprint scope.
9. Enter on a card opens the detail pane.
10. Status changes persist to the CSV database.

## Files to Create/Modify

- `internal/tui/views/kanban.go` -- Kanban board view model
- `internal/tui/views/kanban_test.go` -- Column layout and move tests
- `internal/tui/views/card.go` -- Card rendering component
- `internal/tui/views/card_test.go` -- Card render tests

## Effort Estimate

1.5 weeks

## Test Strategy

- Column assignment: verify requirements land in correct column by status
- Card rendering: golden file tests for card content
- Move operation: verify status update and column transfer
- Blocked enforcement: COMPLETE move rejected for blocked requirement
- Independent scroll: verify column scroll state isolation
- Filter: verify version and category filters reduce visible cards
