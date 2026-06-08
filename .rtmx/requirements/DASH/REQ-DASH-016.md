# REQ-DASH-016: Row Partial Endpoint Returns Single Table Row HTML

## Summary

A new partial endpoint shall return a single `<tr>` element for a given requirement ID,
enabling htmx row-level swaps for inline editing without re-rendering the full table.

## Acceptance Criteria

1. GET `/partials/requirements/row/:id` returns one `<tr>` element for the specified requirement
2. Handler is registered in `serve.go` and implemented in `serve_dashboard.go`
3. Returned row contains the same columns as the full requirements table
4. Returns 404 if requirement ID does not exist

## Dependencies

- REQ-DASH-001 (SPA framework)

## Blocks

- REQ-DASH-017, REQ-DASH-018, REQ-DASH-030

## Files to Modify

- `internal/cmd/serve_dashboard.go`
- `internal/cmd/serve.go`

## Test Strategy

- Unit test: GET row partial returns valid HTML `<tr>` with correct data attributes
- Unit test: GET row partial for nonexistent ID returns 404
- Regression: existing partial endpoint tests continue to pass

## Effort

- 0.25 weeks

## Priority

- HIGH

## Phase

- 31
