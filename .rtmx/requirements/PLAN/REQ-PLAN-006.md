# REQ-PLAN-006: Release Scope Command

## Metadata
- **Category**: PLAN
- **Subcategory**: Release
- **Priority**: HIGH
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: REQ-PLAN-003

## Requirement

`rtmx release scope <version>` shall display a planning summary for a
release version: requirement count, completion percentage, total effort
estimate, remaining effort, and blocking requirements.

## Acceptance Criteria

1. Shows total requirements assigned to version
2. Shows completion breakdown (complete/partial/missing)
3. Shows total effort_weeks and remaining effort_weeks
4. Lists blocking requirements (incomplete dependencies from outside the version)
5. `--json` outputs machine-readable summary

## Files to Modify

- `internal/cmd/release.go`
