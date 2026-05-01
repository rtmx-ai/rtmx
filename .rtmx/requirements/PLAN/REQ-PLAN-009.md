# REQ-PLAN-009: Set Assignee via CLI

## Metadata
- **Category**: PLAN
- **Subcategory**: Assignment
- **Priority**: HIGH
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: REQ-PLAN-008

## Requirement

`rtmx assign <req-id> [--to <user>]` shall set the `assignee` field on the
specified requirement. If `--to` is omitted, the current authenticated user
(from REQ-PLAN-008) is used. Also sets `started_date` if not already set.
`rtmx unassign <req-id>` clears the assignee.

## Design

This is the manual/human assignment path. The ORCH claim protocol
(REQ-ORCH-005, future) provides the concurrent agent claiming path. Both
write to `assignee`. The claim protocol additionally uses `claims.json`
for concurrency control -- `rtmx assign` does not need `claims.json`.

## Acceptance Criteria

1. `rtmx assign REQ-PLAN-001 --to alice` sets assignee to "alice"
2. `rtmx assign REQ-PLAN-001` (no --to) uses current authenticated user
3. `rtmx assign REQ-PLAN-001` without auth gracefully prompts for --to
4. `rtmx unassign REQ-PLAN-001` clears assignee
5. Assignment sets started_date via existing SetStartedDate() if empty
6. Invalid requirement ID produces clear error

## Files to Create

- `internal/cmd/assign.go`
- `internal/cmd/assign_test.go`
