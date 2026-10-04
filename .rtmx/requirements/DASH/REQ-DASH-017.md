# REQ-DASH-017: Click-to-Edit Status and Priority Cells with Dropdown Select

## Summary

Status and priority cells in the requirements table shall be editable inline via
Alpine-driven dropdown selects, committing changes through hx-patch with row partial swap.

## Acceptance Criteria

1. Clicking a status cell shows an Alpine-driven `<select>` with valid status options
2. Clicking a priority cell shows an Alpine-driven `<select>` with valid priority options
3. On change, the select issues `hx-patch` with the updated field value
4. The PATCH response targets the closest `<tr>` via row partial swap
5. The select pre-selects the current value

## Dependencies

- REQ-DASH-016 (row partial endpoint)
- REQ-DASH-010 (design tokens)

## Blocks

- REQ-DASH-019

## Files to Modify

- `internal/dashboard/templates/partials/requirements.html`
- `internal/dashboard/static/app.js`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: clicking status cell renders select with correct options
- Unit test: selecting a new value triggers PATCH request with correct payload
- Integration test: row updates in place after status change

## Effort

- 0.5 weeks

## Priority

- HIGH

## Phase

- 31
