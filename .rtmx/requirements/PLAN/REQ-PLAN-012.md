# REQ-PLAN-012: Release Forecast

## Metadata
- **Category**: PLAN
- **Subcategory**: Forecast
- **Priority**: MEDIUM
- **Phase**: 23
- **Status**: MISSING
- **Dependencies**: REQ-PLAN-011, REQ-PLAN-005

## Requirement

`rtmx release forecast <version>` shall project a completion date for a
release version based on historical velocity and remaining effort.
Remaining effort is the sum of `effort_weeks` for incomplete requirements
assigned to the version. Projected weeks = remaining_effort / velocity.

## Acceptance Criteria

1. Displays projected completion date with confidence qualifier
2. Shows remaining effort, velocity used, and projected weeks
3. Warns when velocity data is insufficient for reliable forecast
4. `--json` outputs machine-readable forecast
5. Accounts for requirements without effort_weeks (warns about unmeasured scope)

## Files to Modify

- `internal/cmd/release.go` -- add forecast subcommand
