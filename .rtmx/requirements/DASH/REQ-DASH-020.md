# REQ-DASH-020: Partial Handler Accepts Sort and Filter Query Parameters

## Summary

The requirements partial handler shall accept sort, order, status, priority, category,
assignee, and search query parameters, applying the same filtering and sorting logic
as the API handler to enable htmx-driven sort and filter interactions.

## Acceptance Criteria

1. `handlePartialRequirements` reads `sort`, `order`, `status`, `priority`, `category`, `assignee`, `search` from URL query parameters
2. Parameters are passed to `db.Filter()` and `sortRequirements()` using the same logic as the API handler
3. Partial response reflects the filtered and sorted result set
4. Missing parameters default to no filter / default sort order

## Dependencies

- REQ-DASH-001 (SPA framework)

## Blocks

- REQ-DASH-021, REQ-DASH-022, REQ-DASH-023, REQ-DASH-024

## Files to Modify

- `internal/cmd/serve_dashboard.go`

## Test Strategy

- Unit test: partial handler with sort param returns rows in correct order
- Unit test: partial handler with status filter returns only matching rows
- Unit test: partial handler with search param returns matching rows
- Unit test: partial handler with no params returns all rows in default order

## Effort

- 0.5 weeks

## Priority

- HIGH

## Phase

- 31
