# REQ-DASH-031: Command Palette Modal Container with Cmd+K Trigger

## Summary

A command palette modal shall open on Cmd+K (Mac) / Ctrl+K with a centered overlay,
auto-focused search input, and dark theme styling for quick navigation and search.

## Acceptance Criteria

1. Cmd+K (Mac) / Ctrl+K opens a centered modal overlay with auto-focused search input
2. Modal has a backdrop, centered position (max-width 600px), and dark theme styling
3. Escape closes the palette
4. Typing elsewhere does not trigger the palette when it is closed
5. Cmd+K toggles the palette if it is already open

## Dependencies

- REQ-DASH-025 (detail panel container)

## Blocks

- REQ-DASH-032

## Files to Modify

- `internal/dashboard/templates/layout.html`
- `internal/dashboard/static/app.js`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: Cmd+K opens palette modal
- Unit test: Escape closes palette modal
- Unit test: search input receives focus on open
- Visual test: modal renders centered with correct width and theme

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 31
