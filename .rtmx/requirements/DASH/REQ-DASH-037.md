# REQ-DASH-037: Keyboard Shortcut Help Overlay on ? Key

## Summary

Pressing the ? key shall open a modal overlay listing all keyboard shortcuts grouped
by context (Global, Table, Panel, Palette), providing discoverability for power users.

## Acceptance Criteria

1. ? key opens an overlay listing all keyboard shortcuts
2. Shortcuts are grouped by context: Global, Table, Panel, Palette
3. Overlay uses dark theme styling consistent with design tokens
4. Escape closes the overlay
5. Overlay does not interfere with other modal states (palette, panel)

## Dependencies

- REQ-DASH-036 (g-prefix shortcuts)

## Blocks

- None

## Files to Modify

- `internal/dashboard/templates/layout.html`
- `internal/dashboard/static/app.js`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: ? key opens shortcut help overlay
- Unit test: overlay lists all registered shortcuts in correct groups
- Unit test: Escape closes the overlay
- Visual test: overlay renders with correct dark theme styling

## Effort

- 0.25 weeks

## Priority

- MEDIUM

## Phase

- 31
