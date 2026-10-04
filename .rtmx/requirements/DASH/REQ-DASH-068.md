# REQ-DASH-068: Timeline Dependency Arrows

## Summary

An SVG overlay on the timeline view draws curved arrows between dependent requirements, making dependency chains visible in the temporal context of sprints. Arrows on the critical path are rendered in red; all other dependency arrows are gray. Dependencies that cross sprint boundaries are visually highlighted with a dashed stroke to draw attention to cross-sprint coupling. The arrows are drawn as a separate SVG layer on top of the timeline bars so they do not interfere with bar hover interactions.

## Acceptance Criteria

1. For each dependency edge (source blocks target), a curved SVG path is drawn from the right edge of the source bar to the left edge of the target bar.
2. Arrows use cubic bezier curves that route around intermediate bars rather than straight lines that obscure content.
3. Critical path edges are colored --rtmx-status-missing (red) with 2px stroke width.
4. Non-critical dependency edges are colored --rtmx-text-dim (gray) with 1px stroke width.
5. Dependencies where source and target are in different sprint columns use a dashed stroke pattern (stroke-dasharray="4,4") in addition to their color.
6. Arrow endpoints include an arrowhead marker (SVG marker element) pointing at the target bar.
7. Hovering an arrow highlights both the source and target bars with increased opacity and border.
8. A toggle control (Alpine checkbox or button) allows hiding/showing dependency arrows. Default state is visible.
9. Arrows are rendered in a separate SVG group (g element) layered above bars but below tooltips.
10. Performance: arrow rendering for 500 dependency edges completes in under 100ms.

## Dependencies

- REQ-DASH-067 (timeline view provides the bar positions and SVG container)
- REQ-DASH-045 (critical path edge data identifies which edges are critical)

## Blocks

(none)

## Files to Modify

- `internal/dashboard/templates/partials/timeline.html` (add arrow toggle control, SVG group for arrows)
- `internal/dashboard/static/app.js` (arrow rendering logic: bezier path calculation, critical path coloring, cross-sprint detection)
- `internal/cmd/serve_dashboard.go` (include dependency edges and critical path flag in timeline JSON)

## Test Strategy

- Unit test: verify dependency edge JSON includes source, target, is_critical, and crosses_sprint fields
- Unit test: verify crosses_sprint is true when source.sprint differs from target.sprint
- Unit test: verify bezier path calculation produces valid SVG path data for various bar position layouts
- Integration test: load timeline with dependencies, verify SVG path elements exist for each edge
- Integration test: verify critical path arrows have red stroke, non-critical have gray
- Integration test: toggle arrow visibility off, verify arrow group has display:none

## Effort

- 1.00 weeks

## Priority

- MEDIUM

## Phase

- 32
