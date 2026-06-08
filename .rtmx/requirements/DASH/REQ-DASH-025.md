# REQ-DASH-025: Detail Panel Container in Layout with Alpine Toggle State

## Summary

The layout shell shall contain a detail panel container on the right edge of the viewport,
managed by Alpine.js state for open/close toggling with smooth transitions.

## Acceptance Criteria

1. `layout.html` contains a `<div id="detail-panel">` with `x-show="panelOpen"` and `x-transition`
2. Panel is 480px wide, positioned fixed on the right edge
3. Panel uses dark theme background consistent with design tokens
4. Panel state is managed by the `rtmxApp()` Alpine component with `panelOpen` (boolean) and `panelReqId` (string)

## Dependencies

- REQ-DASH-001 (SPA framework)
- REQ-DASH-010 (design tokens)

## Blocks

- REQ-DASH-026, REQ-DASH-031

## Files to Modify

- `internal/dashboard/templates/layout.html`
- `internal/dashboard/static/app.js`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: layout contains detail-panel element with correct Alpine bindings
- Unit test: Alpine component initializes with panelOpen=false
- Visual test: panel renders at correct width and position

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 31
