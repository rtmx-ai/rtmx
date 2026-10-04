# REQ-DASH-087: Treemap D3 Squarify Rendering (071b)

## Summary

Client-side D3 treemap rendering that reads the hierarchical JSON produced by REQ-DASH-071 and renders a space-filling rectangle layout. Rectangle area is proportional to effort_weeks, color reflects status via design tokens, and the D3 treemap uses the squarify tiling algorithm for optimal aspect ratios. Category group boundaries are shown with borders and labels. Hovering shows a tooltip; clicking navigates to the detail view.

## Acceptance Criteria

1. The treemap renders as an SVG element within the htmx swap target when view=treemap is selected.
2. The treemap uses d3.treemapSquarify tiling algorithm.
3. Rectangle area is proportional to effort_weeks as provided by the JSON.
4. Rectangle fill color uses status design tokens: --rtmx-status-complete (green), --rtmx-status-partial (amber), --rtmx-status-missing (red), --rtmx-status-not-started (gray).
5. Category group boundaries are shown with a 2px border in --rtmx-border color. Category labels render at the top of each group.
6. Individual requirement rectangles display the requirement ID as text if the rectangle is large enough (width > 60px and height > 20px). Smaller rectangles show no text.
7. Hovering a requirement rectangle shows a tooltip with: ID, title, status, effort_weeks, category, and assignee.
8. Clicking a requirement rectangle navigates to the detail view or opens the detail panel.
9. The treemap fills the available container width and uses a 16:9 aspect ratio for height, with a minimum height of 400px.
10. The treemap renders correctly for 1 to 300 requirements without text overlap or invisible rectangles.

## Dependencies

- REQ-DASH-071 (Go handler provides the hierarchical JSON data structure)
- REQ-DASH-093 (treemap partial template provides the SVG container)
- REQ-DASH-091 (reusable tooltip component for hover)
- REQ-DASH-010 (design tokens for status colors)

## Blocks

- REQ-DASH-072 (treemap color modes)
- REQ-DASH-073 (treemap drill-down)

## Files to Modify

- `internal/dashboard/static/app.js` (treemap D3 rendering function)
- `internal/dashboard/templates/partials/treemap.html` (D3 script block for rendering)

## Test Strategy

- Integration test: load treemap view, verify SVG element renders with correct number of leaf rectangles
- Integration test: verify rectangle colors match status values from database
- Integration test: hover a rectangle, verify tooltip appears with correct data
- Integration test: click a rectangle, verify navigation to detail view
- Performance test: treemap renders 300 requirements in under 200ms

## Effort

- 0.75 weeks

## Priority

- HIGH

## Phase

- 32
