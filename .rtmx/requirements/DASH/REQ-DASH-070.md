# REQ-DASH-070: Timeline Assignee Swimlanes

## Summary

The timeline view supports an alternative grouping mode where rows represent assignees instead of categories. Each assignee row displays their assigned requirements as horizontal bars within the sprint columns. When the total effort_weeks assigned to an assignee in a single sprint exceeds a configurable threshold, the row background turns a warning color to flag potential overload. A toggle control switches between Category and Assignee grouping modes.

## Acceptance Criteria

1. A toggle control (Category | Assignee) renders next to the zoom level control above the timeline SVG.
2. The default grouping mode is Category (existing behavior from REQ-DASH-067).
3. In Assignee mode: each row header displays an assignee name. Requirements within that row are the assignee's assigned items, laid out across sprint columns.
4. Requirements with empty Assignee field are grouped under an "Unassigned" row.
5. Within each assignee row, if multiple requirements fall in the same sprint column, they are stacked vertically.
6. Overload detection: if the sum of effort_weeks for an assignee in a single sprint column exceeds a threshold (default: 4.0 weeks), the cell background renders with --rtmx-status-missing at 15% opacity as a warning.
7. The overload threshold is configurable via the rtmx.yaml config under dashboard.timeline.overload_threshold_weeks.
8. Switching grouping mode issues an htmx GET to `/partials/graph?view=timeline&group={mode}` and swaps the timeline container.
9. The group parameter is persisted in the URL query string (e.g., ?view=timeline&group=assignee).
10. Assignee grouping works correctly at all three zoom levels (Sprint, Phase, Quarter).
11. The Go handler reads the group query parameter and builds row structure by either category or assignee.

## Dependencies

- REQ-DASH-067 (timeline view provides the base row/column layout)

## Blocks

(none)

## Files to Modify

- `internal/dashboard/templates/partials/timeline.html` (add grouping toggle, assignee row rendering)
- `internal/dashboard/static/app.js` (grouping state management, overload cell highlighting)
- `internal/cmd/serve_dashboard.go` (assignee grouping logic, overload calculation, config reading)

## Test Strategy

- Unit test: Go handler with group=assignee produces rows keyed by assignee name
- Unit test: requirements with empty assignee grouped under "Unassigned"
- Unit test: overload detection flags cells where sum of effort_weeks exceeds threshold
- Unit test: overload threshold reads from config, defaults to 4.0 when absent
- Unit test: assignee grouping combined with phase zoom produces correct layout
- Integration test: switch to Assignee grouping, verify row headers show assignee names
- Integration test: verify overloaded cells have warning background color
- Integration test: verify URL updates to include group parameter

## Effort

- 1.00 weeks

## Priority

- MEDIUM

## Phase

- 32
