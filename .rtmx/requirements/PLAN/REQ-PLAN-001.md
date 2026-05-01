# REQ-PLAN-001: Display Assignee and Version in CLI Output

## Metadata
- **Category**: PLAN
- **Subcategory**: Display
- **Priority**: HIGH
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: (none)

## Requirement

The `rtmx backlog` table views shall display "Assignee" and "Version" columns
sourced from the existing `assignee` and `sprint` database fields. The
`rtmx status -vvv` detailed output and `rtmx context` LLM injection shall
also include these fields.

## Rationale

Both fields are stored in the database but completely hidden from all CLI
output. Users cannot see who is assigned to a requirement or which version
it targets without opening the CSV directly.

## Acceptance Criteria

1. `rtmx backlog` shows Assignee and Version columns in all table views
2. `rtmx status -vvv` includes assignee and version per requirement
3. `rtmx context` includes assignee and version in LLM context output
4. Empty fields render as blank, not "N/A" or similar
5. `--json` output includes both fields

## Files to Modify

- `internal/cmd/backlog.go`
- `internal/cmd/status.go`
- `internal/cmd/context.go`
