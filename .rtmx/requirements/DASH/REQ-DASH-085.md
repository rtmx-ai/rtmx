# REQ-DASH-085: Legend Interactive Filter and Highlight Mode (057b)

## Summary

Clicking legend items in the legend panel (REQ-DASH-057) activates a filter/highlight mode that dims non-matching graph elements and emphasizes matching ones. Multiple legend items can be active simultaneously using AND logic. The active state is stored in Alpine.js and survives htmx partial reloads.

## Acceptance Criteria

1. Clicking a legend item activates highlight mode: all graph elements matching that encoding value are rendered at full opacity with an emphasis outline.
2. All non-matching elements are dimmed to 0.3 opacity.
3. The clicked legend item receives an active/selected visual state (e.g., highlighted border).
4. Clicking the same item again (or clicking a "Clear" button) exits highlight mode and restores all elements to full opacity.
5. Multiple legend items can be active simultaneously (AND filter): only elements matching all active filters are highlighted.
6. Active filter state is stored in Alpine.js component state and survives htmx partial reloads within the same session.
7. A "Clear all" button appears when any filter is active, clearing all filters in one click.

## Dependencies

- REQ-DASH-057 (legend panel structure must exist with clickable items)
- REQ-DASH-048 (dagre layout engine renders the graph elements to filter)
- REQ-DASH-010 (design tokens for active state styling)

## Blocks

- None currently identified

## Files to Modify

- `internal/dashboard/static/app.js` (click handlers for legend items, filter/highlight logic, Alpine.js state management)
- `internal/dashboard/static/styles.css` (.legend-item-active, .graph-dimmed, .graph-highlighted classes)

## Test Strategy

- Unit test (JS): verify clicking a legend item adds filter state and dims non-matching elements
- Unit test (JS): verify clicking the same item again removes filter state and restores full opacity
- Unit test (JS): verify multiple active filters apply AND logic
- Integration test: activate legend filter, verify correct graph elements are highlighted via DOM inspection
- State persistence test: activate filter, trigger htmx partial reload, verify filter state persists

## Effort

- 0.75 weeks

## Priority

- MEDIUM

## Phase

- 32
