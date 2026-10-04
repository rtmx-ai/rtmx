# REQ-DASH-024: Search Input Above Requirements Table with Debounced Query

## Summary

A search input with magnifying glass icon shall be placed above the requirements table,
issuing debounced htmx requests to filter the table as the user types.

## Acceptance Criteria

1. A text input with magnifying glass icon is rendered above the requirements table
2. After a 200ms pause in typing, the input sends an `hx-get` with `search=` param
3. The htmx trigger is configured as `hx-trigger="keyup changed delay:200ms"`
4. The table partial re-renders with matching results
5. Clearing the input restores the unfiltered table

## Dependencies

- REQ-DASH-020 (sort/filter query parameters)

## Blocks

- None

## Files to Modify

- `internal/dashboard/templates/partials/requirements.html`
- `internal/dashboard/static/styles.css`

## Test Strategy

- Unit test: search input renders with correct hx-get and hx-trigger attributes
- Unit test: typing triggers partial request with search param after debounce
- Integration test: table filters in real time as user types

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 31
