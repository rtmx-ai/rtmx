# REQ-PLAN-003: Filter by Target Version

## Metadata
- **Category**: PLAN
- **Subcategory**: Query
- **Priority**: P0
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: (none)
- **Blocks**: REQ-PLAN-004, REQ-PLAN-005, REQ-PLAN-006, REQ-PLAN-007

## Requirement

RTMX shall support filtering requirements by target version. The existing
`sprint` CSV column serves as the version targeting field. A `--version`
flag shall be added to `rtmx status`, `rtmx backlog`, and `rtmx verify`.
The `database.FilterOptions` struct shall be extended with a `TargetVersion`
field.

## Design

### Column Strategy

The `sprint` column already contains version strings (e.g., `v0.1.0`,
`v1.0.0`) in the live database. Rather than adding a 22nd column, we reuse
`sprint` semantically as the version targeting field:

```go
// internal/database/requirement.go

// TargetVersion returns the target release version for this requirement.
func (r *Requirement) TargetVersion() string {
    return r.Sprint
}

// SetTargetVersion sets the target release version.
func (r *Requirement) SetTargetVersion(v string) {
    r.Sprint = v
}
```

### Filter Extension

```go
// internal/database/database.go

type FilterOptions struct {
    // ... existing fields ...
    TargetVersion string  // filter by sprint/version field
}
```

### Grouping Method

```go
// ByVersion groups requirements by target version.
func (db *Database) ByVersion() map[string][]*Requirement
```

## Acceptance Criteria

1. `rtmx status --version v0.3.0` shows only requirements targeting v0.3.0
2. `rtmx backlog --version v0.3.0` filters backlog to version scope
3. `rtmx verify --version v0.3.0` verifies only version-scoped requirements
4. `database.Filter()` respects TargetVersion field
5. `database.ByVersion()` returns requirements grouped by version
6. Requirements with empty sprint field are grouped as "unversioned"

## Files to Modify

- `internal/database/database.go` -- FilterOptions, ByVersion()
- `internal/database/requirement.go` -- TargetVersion() accessors
- `internal/cmd/status.go` -- --version flag
- `internal/cmd/backlog.go` -- --version flag
- `internal/cmd/verify.go` -- --version flag
