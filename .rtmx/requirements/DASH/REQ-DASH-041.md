# REQ-DASH-041: POST /api/requirements/bulk Endpoint for Batch Updates

## Summary

A new bulk update API endpoint shall accept a list of requirement IDs and field updates,
applying them atomically with validation to enable efficient multi-requirement edits.

## Acceptance Criteria

1. POST `/api/requirements/bulk` accepts JSON body: `{"req_ids": ["REQ-XXX", ...], "updates": {"status": "COMPLETE"}}`
2. Validates all `req_ids` exist before applying any updates
3. Validates all field values against allowed values (status, priority, etc.)
4. Applies updates atomically (all or none) and saves the database
5. Returns `{"updated": N, "req_ids": [...]}` on success
6. Returns appropriate error response if any validation fails

## Dependencies

- REQ-DASH-001 (SPA framework)

## Blocks

- REQ-DASH-042

## Files to Modify

- `internal/cmd/serve_api.go`
- `internal/cmd/serve.go`

## Test Strategy

- Unit test: valid bulk update returns correct updated count and IDs
- Unit test: nonexistent req_id returns validation error with no changes applied
- Unit test: invalid field value returns validation error with no changes applied
- Unit test: empty req_ids list returns error
- Integration test: bulk update persists changes to database file

## Effort

- 0.5 weeks

## Priority

- MEDIUM

## Phase

- 31
