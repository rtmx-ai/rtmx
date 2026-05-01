# REQ-PLAN-002: Display Dates and Duration in Status Output

## Metadata
- **Category**: PLAN
- **Subcategory**: Display
- **Priority**: MEDIUM
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: (none)

## Requirement

The `rtmx status -vvv` detailed output shall display `started_date`,
`completed_date`, and computed duration for completed requirements. Duration
is `completed_date - started_date` displayed in human-readable form (e.g.,
"12 days", "3 weeks").

## Rationale

Started and completed dates are persisted in the database but never displayed
in any command output. Surfacing duration gives visibility into actual effort
versus estimated effort_weeks.

## Acceptance Criteria

1. `rtmx status -vvv` shows started and completed dates for requirements that have them
2. Duration is computed and displayed for requirements with both dates
3. Dates render in YYYY-MM-DD format consistent with CSV storage
4. Missing dates render as blank

## Files to Modify

- `internal/cmd/status.go`
