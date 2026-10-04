# REQ-DASH-059: Rich Hover Tooltip on Graph Nodes

## Summary

Hovering over a graph node shall display a styled tooltip info card showing the
requirement ID, description (truncated to 80 characters), status badge, priority,
effort estimate, and assignee. The tooltip uses design token colors and typography,
positions itself near the cursor, and repositions to avoid clipping at viewport edges.

## Acceptance Criteria

1. Hovering a node for 200ms displays a tooltip card adjacent to the cursor position
2. Tooltip contains: requirement ID (bold), description (truncated at 80 chars with ellipsis), status badge (colored per design tokens), priority label, effort value, and assignee name
3. Tooltip uses `--color-surface-secondary` background, `--color-text-primary` text, `--radius-md` border radius, and `--shadow-lg` elevation from design tokens
4. Tooltip repositions when it would overflow the viewport: flips horizontally if too close to right edge, flips vertically if too close to bottom edge
5. Tooltip hides immediately when the cursor leaves the node
6. Tooltip hides when the user starts dragging, panning, or zooming the graph
7. Tooltip does not interfere with node click events (REQ-DASH-058)
8. Fields with no data (e.g., no assignee) display "Unassigned" or equivalent placeholder text
9. Tooltip is a single reusable DOM element repositioned on each hover, not one per node

## Dependencies

- REQ-DASH-044 (enriched JSON endpoint provides description, status, priority, effort, assignee)
- REQ-DASH-048 (dagre layout engine renders hoverable node elements)

## Blocks

- None

## Files to Modify

- `internal/dashboard/static/graph.js`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: hovering a node for 200ms renders tooltip with correct fields populated
- Unit test: tooltip repositions when placed near viewport edges
- Unit test: tooltip hides on mouseout, drag start, and zoom start
- Unit test: missing fields display placeholder text
- Integration test: tooltip renders with live data from enriched JSON endpoint

## Effort

- 0.50 weeks

## Priority

- HIGH

## Phase

- 32
