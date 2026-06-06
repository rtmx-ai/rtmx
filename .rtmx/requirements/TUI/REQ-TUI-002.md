# REQ-TUI-002: Requirements Table View with Sort/Filter/Search

## Metadata
- **Category**: TUI
- **Subcategory**: View
- **Priority**: P0
- **Phase**: 26
- **Status**: MISSING
- **Dependencies**: REQ-TUI-001, REQ-API-001
- **Blocks**: REQ-TUI-003

## Requirement

The TUI shall provide a requirements table view as the default landing
screen, displaying all requirements in a sortable, filterable, searchable
table with columns for req_id, status, priority, category, effort, assignee,
and a truncated requirement text.

## Rationale

The requirements table is the most fundamental view in any project management
tool. Users need to quickly browse, filter, and locate specific requirements.
The current TUI shows only aggregate counts by category -- no individual
requirement visibility at all.

## Design

### Table Layout

```
 REQ ID         Status   Pri  Category  Effort  Assignee  Description
 REQ-MCP-007    MISSING  P0   MCP       0.5w              Response Size Logging
 REQ-MCP-008    MISSING  P0   MCP       1.5w              Tool Filtering
 REQ-CLI-001    COMPLETE P0   CLI       1.0w    rhino11   Static binary build
 ...
```

### Components

Uses `bubbles/table` for the table widget with custom column widths that
adapt to terminal width. Wide terminals show more of the description;
narrow terminals hide lower-priority columns (assignee, effort).

### Filter Bar

Pressing `/` activates a filter bar (bubbles/textinput) at the top of the
table. Filter syntax:

- Plain text: substring match across req_id and requirement_text
- `status:MISSING` -- filter by status
- `cat:MCP` -- filter by category
- `pri:P0` -- filter by priority
- `@rhino11` -- filter by assignee
- Combine with spaces: `cat:MCP status:MISSING`

### Sort

- Press `s` to cycle sort field (req_id -> priority -> status -> effort -> category)
- Press `S` (shift) to toggle sort direction

### Pagination

The table scrolls vertically. A "X of Y" indicator in the status bar shows
current position. Page Up/Page Down move by screenful.

## Acceptance Criteria

1. Table displays all requirements with correct column data.
2. Column widths adapt to terminal width.
3. `/` activates filter bar with structured filter syntax.
4. Each filter type correctly narrows displayed requirements.
5. Filters can be combined (e.g., `cat:MCP status:MISSING`).
6. `s` cycles sort field; `S` toggles direction.
7. Vertical scrolling works with j/k, arrow keys, and Page Up/Down.
8. Status bar shows "X of Y" count reflecting applied filters.
9. Pressing Enter on a row opens the detail view (REQ-TUI-003).
10. Pressing Escape clears active filter and returns to full list.

## Files to Create/Modify

- `internal/tui/views/requirements.go` -- Requirements table view model
- `internal/tui/views/requirements_test.go` -- Filter/sort/render tests
- `internal/tui/filter.go` -- Filter parser and application logic
- `internal/tui/filter_test.go` -- Filter syntax tests

## Effort Estimate

1 week

## Test Strategy

- Table-driven tests for filter parser (each filter type, combined filters)
- Sort correctness tests for each field and direction
- Column width adaptation with mocked terminal widths
- Render tests comparing output against golden files
- Integration: filter -> verify displayed rows match expected set
