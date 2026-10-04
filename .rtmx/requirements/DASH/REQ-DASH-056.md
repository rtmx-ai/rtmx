# REQ-DASH-056: Always-Visible Node Labels with Zoom-Responsive Detail

## Summary

Each graph node shall display an always-visible label showing the requirement ID positioned below the node circle, replacing the current hover-only label behavior. When the user zooms in beyond 1.5x magnification, labels expand to include a truncated requirement description (first 40 characters). Labels use the design token monospace font for requirement IDs and sans-serif for descriptions, ensuring readability at all zoom levels.

## Acceptance Criteria

1. Every node in the graph has a persistent `<text>` SVG element positioned below the node circle, displaying the requirement ID (e.g., "REQ-DASH-050").
2. Label y-offset is computed as: node center y + node radius + 12px, ensuring the label clears the node circle regardless of node size.
3. Labels use `var(--rtmx-text-muted)` (#9ca3af) fill color and font-size 10px at 1x zoom.
4. The font-family for requirement IDs is the design token monospace font (e.g., `'JetBrains Mono', 'Fira Code', monospace`).
5. When the D3 zoom transform scale exceeds 1.5x, a second line of text appears below the ID showing the requirement description truncated to 40 characters with an ellipsis.
6. The description line uses `var(--rtmx-text-dim)` (#6b7280) fill color and font-size 9px.
7. The zoom threshold check runs inside the D3 zoom event handler and toggles a CSS class (`.labels-expanded`) on the SVG container, not on individual labels.
8. Labels do not overlap with other nodes. Dagre node spacing (nodesep) accounts for label height in its layout calculation.
9. At zoom levels below 0.5x, labels are hidden entirely (opacity: 0) to prevent visual clutter on zoomed-out views.
10. The hover-only label behavior from the previous implementation is removed.

## Dependencies

- REQ-DASH-048 (dagre layout engine provides stable node positions for label placement)
- REQ-DASH-044 (enriched graph JSON provides requirement description text)
- REQ-DASH-010 (design tokens for font and color references)

## Blocks

- None currently identified

## Files to Modify

- `internal/dashboard/static/app.js` (add persistent labels to renderGraph, add zoom-level responsive detail toggling in zoom handler, remove hover-only label logic)
- `internal/dashboard/static/styles.css` (define .node-label, .node-label-desc classes, .labels-expanded modifier, zoom-level visibility rules)

## Test Strategy

- Unit test (JS): verify every node has a <text> element with the correct requirement ID
- Unit test (JS): verify label y-position is below the node circle by radius + 12px
- Unit test (JS): verify zoom scale > 1.5 adds .labels-expanded class to SVG container
- Unit test (JS): verify zoom scale < 0.5 hides labels (opacity 0)
- Unit test (JS): verify description text is truncated to 40 characters with ellipsis
- Visual test: screenshot at 1x zoom showing IDs only, screenshot at 2x zoom showing IDs + descriptions
- Layout test: verify dagre nodesep accounts for label height, no label-node overlaps in 100+ node graph
- Regression test: hover zoom and pan behaviors remain functional

## Effort

- 1.00 weeks

## Priority

- HIGH

## Phase

- 32
