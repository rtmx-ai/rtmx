# REQ-DASH-028: Panel Close on Escape Key and Backdrop Click

## Summary

The detail panel shall close when the user presses Escape or clicks the backdrop overlay,
providing standard modal dismissal patterns.

## Acceptance Criteria

1. Pressing Escape sets `panelOpen=false` when the panel is open
2. Clicking the backdrop overlay also closes the panel
3. Escape only triggers panel close when the panel is actually open
4. Focus management is correct after panel close

## Dependencies

- REQ-DASH-027 (req-id links open panel)

## Blocks

- REQ-DASH-029, REQ-DASH-030, REQ-DASH-033

## Files to Modify

- `internal/dashboard/templates/layout.html`
- `internal/dashboard/static/app.js`

## Test Strategy

- Unit test: Escape keydown sets panelOpen to false
- Unit test: backdrop click sets panelOpen to false
- Unit test: Escape is no-op when panel is already closed

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 31
