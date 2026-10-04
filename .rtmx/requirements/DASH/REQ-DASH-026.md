# REQ-DASH-026: Panel Slide Animation and Semi-Transparent Backdrop CSS

## Summary

The detail panel shall slide in from the right edge with a 200ms ease transition,
and a semi-transparent backdrop shall cover the content area behind the panel.

## Acceptance Criteria

1. Panel slides from off-screen right with a 200ms ease transition
2. A semi-transparent backdrop (#000 at 40% opacity) covers the content area behind the panel
3. Backdrop is visible only when panel is open
4. Transition is smooth with no layout shift

## Dependencies

- REQ-DASH-025 (detail panel container)

## Blocks

- REQ-DASH-027

## Files to Modify

- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: panel CSS includes transform transition with 200ms duration
- Unit test: backdrop element has correct opacity and background color
- Visual test: panel slide-in animation is smooth

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 31
