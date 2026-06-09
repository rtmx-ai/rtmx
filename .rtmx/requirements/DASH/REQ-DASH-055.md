# REQ-DASH-055: Cycle Member Node Warning Indicator

## Summary

Nodes that participate in a dependency cycle shall display a small warning triangle badge and an orange pulsing glow animation, making cycles immediately visible in the graph without requiring manual inspection. Clicking a cycle member node highlights the full cycle path (all nodes and edges in the cycle), enabling rapid diagnosis of circular dependencies that block topological ordering.

## Acceptance Criteria

1. `renderGraph()` reads the `is_cycle_member` field (boolean) from each node in the enriched JSON payload.
2. Cycle member nodes display a small warning triangle badge (SVG polygon) positioned at the top-right of the node circle. The badge uses `var(--rtmx-status-partial)` (#f59e0b) fill with a `!` character or exclamation mark inside.
3. Cycle member nodes have an orange pulsing glow effect via CSS animation:
   - `box-shadow` (or SVG `filter: drop-shadow`) using #f59e0b at 50% opacity
   - Pulsing animation: opacity alternates between 0.3 and 0.8 over a 2s cycle
4. The pulsing animation respects `prefers-reduced-motion` (animation paused, glow at static 0.5 opacity).
5. Clicking a cycle member node triggers a highlight mode:
   - All nodes in the same cycle receive a bright orange outline
   - All edges in the cycle are highlighted with orange stroke and increased width
   - Non-cycle nodes and edges are dimmed to 0.3 opacity
   - A second click (or click on background) exits highlight mode and restores normal rendering
6. The cycle path data (which nodes form the cycle) is provided in the enriched JSON or computed client-side from the dependency graph.
7. The warning badge scales proportionally with node size but has a minimum visible size of 8px.
8. Non-cycle-member nodes have no badge and no glow.
9. The legend panel includes a "Badge = Cycle" entry showing the warning triangle indicator.

## Dependencies

- REQ-DASH-044 (enriched graph JSON provides is_cycle_member per node)
- REQ-DASH-048 (dagre layout engine)
- REQ-DASH-010 (design tokens for orange/amber color)

## Blocks

- REQ-DASH-057 (legend must reference the cycle indicator encoding)

## Files to Modify

- `internal/dashboard/static/app.js` (add cycle badge rendering, click handler for cycle highlight mode, cycle path computation or lookup)
- `internal/dashboard/static/styles.css` (define .node-cycle-member class with glow animation, .cycle-highlight and .cycle-dimmed classes)

## Test Strategy

- Unit test (JS): verify nodes with is_cycle_member=true receive warning badge SVG element
- Unit test (JS): verify nodes with is_cycle_member=false have no badge
- Unit test (JS): verify click on cycle member triggers highlight mode with correct node/edge set
- Unit test (JS): verify second click or background click exits highlight mode
- Unit test (CSS): verify .node-cycle-member defines pulsing animation keyframes
- Accessibility test: verify animation pauses under prefers-reduced-motion
- Visual test: screenshot comparison showing cycle member nodes with badges and glow
- Edge case test: node in multiple cycles (belongs to two overlapping cycles) -- click highlights the cycle containing the clicked node
- Edge case test: graph with no cycles (no badges rendered)

## Effort

- 1.00 weeks

## Priority

- HIGH

## Phase

- 32
