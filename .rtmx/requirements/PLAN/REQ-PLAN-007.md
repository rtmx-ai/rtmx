# REQ-PLAN-007: Assign Requirements to Version

## Metadata
- **Category**: PLAN
- **Subcategory**: Release
- **Priority**: HIGH
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: REQ-PLAN-003

## Requirement

`rtmx release assign <version> <req-id> [req-id...]` shall set the target
version (sprint field) for the specified requirements. `rtmx release unassign
<req-id> [req-id...]` shall clear it.

## Acceptance Criteria

1. `rtmx release assign v0.3.0 REQ-PLAN-001 REQ-PLAN-002` sets sprint for both
2. `rtmx release unassign REQ-PLAN-001` clears sprint
3. Invalid requirement IDs produce clear error messages
4. Database file is updated in place with backup
5. `--dry-run` previews changes without writing

## Files to Modify

- `internal/cmd/release.go`
