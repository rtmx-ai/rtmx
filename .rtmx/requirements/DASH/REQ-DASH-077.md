# REQ-DASH-077: Screen Reader Support for Graph View

## Summary

The graph view shall be accessible to screen reader users through ARIA roles, live regions,
and a text summary. The SVG element has role="img" with an accessible label. Cluster groups
use role="group". Edges are described via aria-label. A live region announces node details
when focus changes. A visible summary paragraph above the graph states key metrics:
total nodes, total edges, critical path length, and blocked count.

## Acceptance Criteria

1. The graph SVG element has `role="img"` and `aria-label="Dependency graph for {project name}"`
2. Each cluster group within the SVG has `role="group"` and `aria-label="{category}: {count} requirements"`
3. Each node element has `role="button"`, `aria-label="{req-id}: {title} ({status})"`, and `tabindex="-1"`
4. Each edge path has `aria-label="{source-id} depends on {target-id}"` and `role="presentation"` (edges are supplementary, not interactive)
5. A `<div aria-live="polite">` region above the graph announces the focused node's ID, title, status, and dependency count when focus changes
6. A visible summary paragraph renders above the graph: "Dependency graph: {X} nodes, {Y} edges, {Z} on critical path, {W} blocked"
7. The summary updates dynamically when filters change the visible node set
8. When clusters are collapsed (REQ-DASH-075), the cluster node aria-label includes the member count: "{category}: {count} requirements (collapsed)"
9. Color is not the sole indicator of status -- each status-colored node also displays a text label or icon with a text alternative
10. All graph controls (buttons, dropdowns) have visible labels or aria-label attributes

## Dependencies

- REQ-DASH-048 (dagre layout provides the rendered graph structure)
- REQ-DASH-056 (node labels must exist for aria-label content)

## Blocks

- None

## Files to Modify

- `internal/dashboard/static/app.js` (ARIA attributes on SVG elements, live region updates)
- `internal/dashboard/static/styles.css` (summary paragraph styling, live region positioning)
- `internal/dashboard/templates/partials/graph.html` (summary paragraph, live region container, SVG ARIA attributes)

## Test Strategy

- Unit test: SVG element has role="img" and correct aria-label
- Unit test: each node has role="button" and aria-label with req-id, title, status
- Unit test: cluster groups have role="group" with category and count in aria-label
- Unit test: live region content updates on focus change
- Unit test: summary paragraph shows correct node count, edge count, critical path count, blocked count
- Unit test: summary updates when filters reduce visible node set
- Integration test: screen reader (axe-core audit) reports no critical accessibility violations on graph partial
- Integration test: keyboard navigation (REQ-DASH-076) triggers live region announcements

## Effort

- 1.00 weeks

## Priority

- HIGH

## Phase

- 32
