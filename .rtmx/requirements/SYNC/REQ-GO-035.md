# REQ-GO-035: Shadow Requirements for Cross-Repo Dependencies

## Metadata
- **Category**: SYNC
- **Subcategory**: CrossRepo
- **Priority**: HIGH
- **Phase**: 11
- **Status**: MISSING
- **Effort**: 2 weeks
- **Dependencies**: REQ-GO-033 (Remote management)
- **Blocks**: REQ-GO-036 (Grant delegation), REQ-GO-042 (CRDT sync), REQ-GO-075 (Cross-repo move/clone)

## Requirement

Go CLI shall resolve cross-repo dependency references in the format `sync:ALIAS:REQ-ID`, loading the referenced requirement from a locally-cloned remote repository and reflecting its status in dependency analysis, blocking calculations, and verification output.

## Rationale

RTMX-enabled projects frequently depend on requirements in sibling repositories. For example, rtmx-go may depend on `sync:rtmx:REQ-AUTH-001` in the Python CLI repo. Today these references are recognized but skipped with a TODO comment in `IsBlocked()` and `BlockingDeps()`. Without resolution, backlog and deps commands cannot give accurate blocking analysis for cross-repo work.

Shadow requirements preserve the local-first philosophy: all requirement context lives in locally-cloned repositories accessible to the agent. No SaaS dependency (GitHub, Jira) is required for resolution -- only a filesystem path to the remote project's database.

## Design

### Data Model

```go
// internal/sync/shadow.go

// ShadowRequirement is a local projection of a requirement from another repository.
type ShadowRequirement struct {
    ReqID        string            // Original ID in the remote repo (e.g., "REQ-AUTH-001")
    RemoteAlias  string            // Config alias (e.g., "rtmx")
    RemoteRepo   string            // Repository identifier (e.g., "rtmx-ai/rtmx")
    Status       database.Status   // COMPLETE, PARTIAL, MISSING
    Description  string            // Requirement text
    Phase        int               // Phase number
    Dependencies []string          // Dependencies within the remote repo
    Visibility   string            // "full" (default; "shadow" and "hash_only" deferred to REQ-GO-036)
    ResolvedAt   time.Time         // When this shadow was last resolved
}
```

### Shadow Resolver

```go
// ShadowResolver resolves cross-repo dependency references to ShadowRequirements.
type ShadowResolver struct {
    Remotes map[string]config.SyncRemote
    Cache   map[string]*ShadowRequirement  // keyed by "alias:req_id"
}

// Resolve parses a "sync:ALIAS:REQ-ID" reference and returns the shadow.
// Returns an error if the alias is unknown or the remote database is not accessible.
func (r *ShadowResolver) Resolve(ref string) (*ShadowRequirement, error)

// ResolveAll resolves all cross-repo dependencies in a database.
// Returns resolved shadows and a list of warnings for unresolvable refs.
func (r *ShadowResolver) ResolveAll(db *database.Database) ([]*ShadowRequirement, []string)

// IsResolvable checks if a dependency reference is a cross-repo reference.
func IsResolvable(dep string) bool  // returns true if matches "sync:*:*"

// ParseRef splits "sync:ALIAS:REQ-ID" into alias and requirement ID.
func ParseRef(ref string) (alias string, reqID string, err error)
```

### Resolution Strategy

1. Parse the dependency string: `sync:ALIAS:REQ-ID`
2. Look up ALIAS in `config.RTMX.Sync.Remotes`
3. Require `remote.Path` to be set (local clone path)
4. Load the remote database from `remote.Path/remote.Database`
5. Find the requirement by ID
6. Return a ShadowRequirement with full visibility

If the remote path doesn't exist or the requirement isn't found, return a warning (not a hard error). Unresolvable shadows do not block local requirements -- they produce warnings in output.

### Integration Points

**`internal/database/requirement.go`** -- Extend `IsBlocked()` and `BlockingDeps()`:
- Accept an optional `ShadowResolver`
- For deps matching `sync:*:*`, resolve via the shadow resolver
- A shadow with status COMPLETE does not block; MISSING or PARTIAL does
- Unresolvable shadows produce warnings but do not block

**`internal/graph/graph.go`** -- Extend dependency graph:
- `BuildGraph()` accepts shadows as additional nodes
- Cross-repo edges marked with edge type for display
- Cycle detection works within repos only (cross-repo cycles are warnings)

**`internal/cmd/deps.go`** -- Display shadow status:
- Show `[SHADOW]` indicator for cross-repo dependencies
- Show remote alias and status: `sync:rtmx:REQ-AUTH-001 [COMPLETE]`
- Warn if shadow could not be resolved: `sync:rtmx:REQ-AUTH-001 [UNRESOLVED]`

**`internal/cmd/status.go`** -- Count shadow deps in blocking analysis

**`internal/cmd/verify.go`** -- `rtmx verify --update` refreshes shadow status:
- Re-resolve all cross-repo dependencies
- Update local database if shadow status changed
- Report changes: `sync:rtmx:REQ-AUTH-001: MISSING -> COMPLETE`

### Files to Create

- `internal/sync/shadow.go` -- ShadowRequirement model and ShadowResolver
- `internal/sync/shadow_test.go` -- Tests with real temp directories and databases

### Files to Modify

- `internal/database/requirement.go` -- IsBlocked/BlockingDeps accept resolver
- `internal/graph/graph.go` -- Cross-repo edge support
- `internal/cmd/deps.go` -- Shadow display in output
- `internal/cmd/status.go` -- Shadow-aware blocking counts
- `internal/cmd/verify.go` -- Shadow refresh on verify --update

## Acceptance Criteria

1. `sync:ALIAS:REQ-ID` references in dependency fields resolve to the correct requirement from the remote's local database
2. A COMPLETE shadow does not block the local requirement
3. A MISSING/PARTIAL shadow blocks the local requirement (shown in backlog/deps)
4. An unresolvable shadow (missing remote path, missing requirement) produces a warning, not a hard error
5. `rtmx deps REQ-ID` shows shadow dependencies with `[SHADOW]` indicator and status
6. `rtmx verify --update` refreshes shadow status from remote databases
7. Shadow resolution works only from local paths (no network/API calls)
8. The `Visibility` field defaults to `"full"` (other levels deferred to REQ-GO-036)

## Test Strategy

- **Test Module**: `internal/sync/shadow_test.go`
- **Test Function**: `TestShadowRequirements`
- **Validation Method**: Integration Test

### Test Cases

1. **ParseRef** -- valid and invalid `sync:ALIAS:REQ-ID` strings
2. **Resolve** -- happy path with two temp projects, one as remote of the other
3. **Resolve with COMPLETE shadow** -- local req not blocked
4. **Resolve with MISSING shadow** -- local req IS blocked
5. **Unresolvable alias** -- warning, not error
6. **Unresolvable path** -- remote configured but path doesn't exist, warning
7. **Unresolvable requirement** -- path exists but req not in database, warning
8. **ResolveAll** -- database with mix of local and cross-repo deps
9. **Integration with deps command** -- output includes [SHADOW] indicators
10. **Integration with verify --update** -- shadow status refreshed from remote
