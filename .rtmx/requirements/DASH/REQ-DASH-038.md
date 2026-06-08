# REQ-DASH-038: Keyboard Shortcuts Suppressed in Input and Textarea Fields

## Summary

All single-key shortcuts shall be suppressed when focus is in input, select, or textarea
fields, preventing accidental navigation while typing. Modifier shortcuts (Cmd+K) shall
continue to work in input fields.

## Acceptance Criteria

1. Single-key shortcuts (j, k, g, ?, e, x, o) are no-op when `document.activeElement` is `input`, `select`, or `textarea`
2. Modifier shortcuts (Cmd+K) still work when focus is in input fields
3. Shortcuts resume normal behavior when focus leaves input fields

## Dependencies

- REQ-DASH-035 (j/k table navigation)
- REQ-DASH-036 (g-prefix shortcuts)

## Blocks

- None

## Files to Modify

- `internal/dashboard/static/app.js`

## Test Strategy

- Unit test: j key is no-op when focus is in text input
- Unit test: g key is no-op when focus is in select element
- Unit test: Cmd+K still triggers palette when focus is in input
- Unit test: shortcuts work after focus leaves input field

## Effort

- 0.25 weeks

## Priority

- MEDIUM

## Phase

- 31
