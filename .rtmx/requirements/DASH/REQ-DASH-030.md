# REQ-DASH-030: Saving in Panel Triggers Out-of-Band Swap on Underlying Table Row

## Summary

When a requirement is saved from the detail panel, the PATCH response shall include
an out-of-band swap for the corresponding table row, keeping the underlying table
in sync without requiring the user to close the panel.

## Acceptance Criteria

1. PATCH from the detail panel form returns a response with `hx-swap-oob="true"` on the `<tr>` for the edited row
2. The requirements table row updates in place without the user closing the panel
3. The out-of-band swap uses the same row partial as REQ-DASH-016
4. Panel content remains stable during the swap

## Dependencies

- REQ-DASH-028 (panel close)
- REQ-DASH-016 (row partial endpoint)

## Blocks

- None

## Files to Modify

- `internal/cmd/serve_dashboard.go`
- `internal/cmd/serve_api.go`
- `internal/dashboard/templates/partials/detail.html`

## Test Strategy

- Unit test: PATCH response includes hx-swap-oob attribute on tr element
- Unit test: out-of-band row matches the row partial format
- Integration test: editing in panel updates table row in background

## Effort

- 0.5 weeks

## Priority

- HIGH

## Phase

- 31
