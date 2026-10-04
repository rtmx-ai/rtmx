# REQ-DASH-057: Legend Panel Structure with All Encoding Sections (057a)

## Summary

The graph view shall include a collapsible legend panel that documents all visual encodings used in the roadmap graph: node size (effort), fill color (status), border ring (priority), edge style (critical path), dashed border (blocked), and warning badge (cycle member). The panel renders as a static reference, positioned in the bottom-left corner of the graph container. It supports collapse/expand toggling and uses design token styling. Interactive filtering is handled separately by REQ-DASH-085.

## Acceptance Criteria

1. The legend panel is rendered as an HTML element overlaying the bottom-left corner of the graph SVG container, using `position: absolute` with `z-index` above the SVG.
2. The legend panel has a collapse/expand toggle. Default state is expanded. Collapsed state shows only a small icon button.
3. Legend sections and their entries:
   - **Size = Effort**: three reference circles (small/0.5w, medium/2w, large/4w) with labels
   - **Fill = Status**: four colored circles (COMPLETE, PARTIAL, MISSING, NOT_STARTED) with labels
   - **Border = Priority**: four ring examples (P0 red 3px, HIGH orange 2px, MEDIUM blue 1px, LOW none) with labels
   - **Edge = Critical Path**: bold emerald animated line vs. thin gray line with labels
   - **Dashed = Blocked**: dashed-border circle vs. solid circle with labels
   - **Badge = Cycle**: warning triangle indicator with label
4. The legend panel uses design token colors and fonts: `var(--rtmx-bg-card)` background, `var(--rtmx-border)` border, `var(--rtmx-text)` labels.
5. Legend collapse/expand state is stored in Alpine.js component state and survives htmx partial reloads within the same session.
6. The legend panel is responsive: on narrow viewports (< 768px), it renders as a horizontal strip below the graph instead of an overlay.

## Dependencies

- REQ-DASH-050 (size encoding must exist for legend to reference)
- REQ-DASH-051 (fill color encoding must exist for legend to reference)
- REQ-DASH-052 (border ring encoding must exist for legend to reference)
- REQ-DASH-053 (critical path edge encoding must exist for legend to reference)
- REQ-DASH-054 (blocked node encoding must exist for legend to reference)
- REQ-DASH-055 (cycle indicator encoding must exist for legend to reference)
- REQ-DASH-048 (dagre layout engine)
- REQ-DASH-010 (design tokens)

## Blocks

- REQ-DASH-085 (interactive filter mode depends on legend panel structure)

## Files to Modify

- `internal/dashboard/static/app.js` (add legend rendering function, Alpine.js state for collapse/expand)
- `internal/dashboard/static/styles.css` (define .legend-panel, .legend-section, .legend-item, .legend-collapsed classes, responsive breakpoint rules)
- `internal/dashboard/templates/partials/graph.html` (add legend panel HTML structure with Alpine.js bindings)

## Test Strategy

- Unit test (JS): verify legend panel renders with all six sections and correct entries
- Unit test (JS): verify collapse/expand toggle works and persists in Alpine.js state
- Unit test (CSS): verify .legend-panel uses design token custom properties
- Responsive test: verify legend renders as horizontal strip on viewport < 768px
- Accessibility test: legend items are keyboard-focusable and have appropriate aria-labels

## Effort

- 0.75 weeks

## Priority

- MEDIUM

## Phase

- 32
