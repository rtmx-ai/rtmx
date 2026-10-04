# REQ-DASH-060: Group-By Control with Swimlane Layout

## Summary

A group-by control shall allow users to organize graph nodes into labeled swimlane
rectangles by dimension: None, Category, Assignee, Phase, or Sprint. The control
renders as a segmented button group in the graph toolbar. When a grouping is active,
dagre lays out nodes within their group boundaries, and labeled rounded rectangles
visually enclose each group. The default grouping is Category.

## Acceptance Criteria

1. A segmented button group in the graph toolbar offers options: None, Category, Assignee, Phase, Sprint
2. Selecting a group-by option re-renders the graph with nodes organized into swimlane groups
3. Each group is enclosed in a labeled rounded rectangle with `--color-surface-tertiary` fill and `--color-border-primary` stroke
4. Group labels are rendered at the top-left of each swimlane using `--font-size-sm` and `--color-text-secondary`
5. Dagre layout respects group boundaries: nodes within a group are laid out together, edges cross group boundaries cleanly
6. Selecting "None" removes all group rectangles and lays out nodes in a flat graph
7. Default grouping on page load is Category
8. Group-by selection persists in URL query parameter (`?groupBy=category`) so it survives page reloads
9. Empty groups (no matching nodes after filtering) are not rendered
10. Changing group-by preserves the current filter and highlight state

## Dependencies

- REQ-DASH-048 (dagre layout engine must support compound/grouped node layout)

## Blocks

- None

## Files to Modify

- `internal/dashboard/static/graph.js`
- `internal/dashboard/templates/partials/graph.html`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: selecting each group-by option produces correct node groupings
- Unit test: dagre layout positions nodes within their group boundaries
- Unit test: group rectangles render with correct labels and styling
- Unit test: "None" option removes all group rectangles
- Unit test: empty groups are excluded from rendering
- Integration test: group-by selection round-trips through URL query parameter

## Effort

- 1.00 weeks

## Priority

- MEDIUM

## Phase

- 32
