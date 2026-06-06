# REQ-API-003: Requirement Update Endpoint

## Metadata
- **Category**: API
- **Subcategory**: REST
- **Priority**: HIGH
- **Phase**: 25
- **Status**: MISSING
- **Dependencies**: REQ-API-001
- **Blocks**: REQ-DASH-003, REQ-DASH-005, REQ-DASH-006

## Requirement

The RTMX serve command shall expose a `PATCH /api/requirements/:id` endpoint
that allows updating mutable fields on a requirement (status, assignee,
sprint, priority, notes) and persists changes to the CSV database.

## Rationale

A project management dashboard that cannot update requirements is read-only
and forces users back to the CLI or direct CSV editing for every change. The
GUI Kanban board and release planning views both need server-side mutation
to support drag-and-drop status transitions and version assignment.

## Design

### Request Schema

```json
{
  "status": "PARTIAL",
  "assignee": "rhino11",
  "sprint": "v1.3.0",
  "priority": "HIGH",
  "notes": "In progress -- blocking on API review"
}
```

All fields are optional. Only provided fields are updated. Unknown fields
return 400.

### Mutable vs Immutable Fields

| Mutable | Immutable |
|---------|-----------|
| status, assignee, sprint, priority, notes | req_id, category, subcategory, requirement_text, dependencies, blocks, phase |

Immutable fields require CLI-level operations (`rtmx edit`, direct CSV)
because they affect the dependency graph and requirement identity.

### Persistence

Updates call `db.Update(reqID, changes)` which writes back to the CSV file.
A file lock (flock) prevents concurrent write corruption when multiple
dashboard users are editing simultaneously.

### Validation

- `status` must be one of: `COMPLETE`, `PARTIAL`, `MISSING`, `NOT_STARTED`
- `priority` must be one of: `P0`, `HIGH`, `MEDIUM`, `LOW`
- `sprint` must match semver pattern or be empty string
- Setting status to `COMPLETE` auto-sets `completed_date` to today
- Setting status from `COMPLETE` to another value clears `completed_date`

### Response

Returns the full updated requirement (same schema as GET detail) with
HTTP 200 on success.

## Acceptance Criteria

1. `PATCH /api/requirements/REQ-MCP-007` with `{"status": "PARTIAL"}` updates status.
2. Only mutable fields are accepted; immutable fields return 400.
3. Unknown fields return 400 with descriptive error.
4. Invalid status/priority values return 400.
5. Changes are persisted to CSV and survive server restart.
6. Concurrent PATCH requests are serialized via file lock.
7. Setting status to COMPLETE auto-populates completed_date.
8. Response returns the full updated requirement.

## Files to Create/Modify

- `internal/cmd/serve_api.go` -- Add PATCH handler
- `internal/cmd/serve_api_test.go` -- Mutation tests
- `internal/database/database.go` -- Add Update method with flock

## Effort Estimate

0.5 weeks

## Test Strategy

- Table-driven tests for each mutable field
- Validation tests for invalid values
- Immutable field rejection tests
- Concurrent write test (two PATCH requests racing)
- Persistence test: PATCH then re-load database, verify change stuck
- Auto-date tests: completed_date set/cleared on status transitions
