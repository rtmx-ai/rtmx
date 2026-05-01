# REQ-PLAN-004: Version Summary View

## Metadata
- **Category**: PLAN
- **Subcategory**: Display
- **Priority**: MEDIUM
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: REQ-PLAN-003

## Requirement

`rtmx status --by-version` shall group requirements by target version and
display per-version completion percentages, following the existing pattern
of `ByPhase()` and `ByCategory()` grouping.

## Acceptance Criteria

1. `rtmx status --by-version` shows each version with completion % and counts
2. Unversioned requirements grouped under "(unversioned)"
3. Versions sorted lexicographically (v0.1.0, v0.2.0, v1.0.0)
4. Output format matches existing phase/category breakdown style

## Files to Modify

- `internal/database/database.go` -- ByVersion() method
- `internal/cmd/status.go` -- --by-version flag and rendering
