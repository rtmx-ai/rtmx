# REQ-DASH-005: Kanban Board View

## Metadata
- **Category**: DASH
- **Subcategory**: View
- **Priority**: HIGH
- **Phase**: 28
- **Status**: MISSING
- **Dependencies**: REQ-DASH-001, REQ-API-005, REQ-API-003
- **Blocks**: (none)

## Requirement

The web dashboard shall provide a Kanban board view with drag-and-drop
status transitions, displaying requirements as cards in status-based
columns with visual indicators for priority, effort, blocked state,
and assignee.

## Rationale

Kanban boards are the most widely adopted project management visualization.
Drag-and-drop status transitions provide the most intuitive way to update
requirement status. This is the view most likely to replace external tools
(Asana, Monday, Trello) for teams already using RTMX for traceability.

## Design

### Layout

Four columns: NOT_STARTED, MISSING, PARTIAL, COMPLETE. Column headers
show count and total effort.

```
+-- NOT STARTED (12) ----+  +-- MISSING (3) ---------+  +-- PARTIAL (0) --+  +-- COMPLETE (226) ------+
| 8.5 weeks remaining    |  | 2.5 weeks remaining    |  | 0 weeks        |  | Done                   |
|                        |  |                        |  |                 |  |                        |
| +--------------------+ |  | +--------------------+ |  |                 |  | +--------------------+ |
| | REQ-TUI-001        | |  | | REQ-MCP-007    P0  | |  |                 |  | | REQ-CLI-001    P0  | |
| | TUI Framework      | |  | | Size Logging       | |  |                 |  | | Static Binary      | |
| | 2.0w  [unassigned] | |  | | 0.5w  [unassigned] | |  |                 |  | | 1.0w  rhino11      | |
| +--------------------+ |  | +--------------------+ |  |                 |  | +--------------------+ |
| +--------------------+ |  | +--------------------+ |  |                 |  | ...                    |
| | REQ-TUI-002        | |  | | REQ-MCP-008    P0  | |  |                 |  |                        |
| | Req Table      [B] | |  | | Filtering      [B] | |  |                 |  |                        |
| | 1.0w  [unassigned] | |  | | 1.5w  [unassigned] | |  |                 |  |                        |
| +--------------------+ |  | +--------------------+ |  |                 |  |                        |
+------------------------+  +------------------------+  +-----------------+  +------------------------+
```

### Drag and Drop

HTML5 Drag and Drop API with Alpine.js state management. Dragging a card
to a new column triggers `PATCH /api/requirements/:id` with the new status.

### Validation

- Cards with incomplete upstream dependencies cannot be dropped in COMPLETE
  column. The drop target shows a red border and tooltip explaining the
  blocking dependencies.
- Cards animate back to their original column on rejected drops.

### Card Content

Each card shows:
- req_id and priority badge (color-coded)
- Truncated requirement text (2 lines max)
- Effort estimate and assignee
- `[B]` badge for blocked requirements
- Category tag (small, dimmed)

### Filters

- Category filter: dropdown above the board
- Version/Sprint filter: dropdown
- Assignee filter: dropdown
- The COMPLETE column is collapsed by default (expandable) to focus on
  active work

## Acceptance Criteria

1. Four columns render for each status value.
2. Cards display in the correct column based on current status.
3. Drag and drop moves a card to the target column and persists the status change.
4. Blocked cards cannot be dropped in the COMPLETE column.
5. Rejected drops animate the card back to its original position.
6. Column headers show requirement count and total effort.
7. COMPLETE column is collapsed by default with expand toggle.
8. Category, version, and assignee filters work.
9. Card click navigates to the requirement detail page.
10. Cards show priority badge, effort, assignee, and blocked indicator.

## Files to Create/Modify

- `dashboard/kanban.html` -- Kanban board template
- `dashboard/components/card.html` -- Card component
- `dashboard/js/kanban.js` -- Drag-and-drop logic with Alpine.js

## Effort Estimate

1 week

## Test Strategy

- Column assignment: verify requirements in correct column by status
- Drag and drop: simulate drag, verify PATCH request with new status
- Blocked enforcement: drag blocked card to COMPLETE, verify rejection
- Filter: apply category filter, verify card visibility
- Card rendering: verify all card fields present
- COMPLETE column collapse: verify collapsed by default, expandable
