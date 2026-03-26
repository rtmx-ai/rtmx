# REQ-GO-036: Grant Delegation for Access Control

## Metadata
- **Category**: SYNC
- **Subcategory**: Grants
- **Priority**: MEDIUM
- **Phase**: 11
- **Status**: MISSING
- **Effort**: 1.5 weeks
- **Dependencies**: REQ-GO-035 (Shadow requirements)
- **Blocks**: None

## Requirement

Go CLI shall implement grant delegation commands (`rtmx grant create`, `rtmx grant list`, `rtmx grant revoke`) for controlling which requirement data is visible to remote collaborators at each visibility level.

## Rationale

When organizations share requirement boundaries (e.g., rtmx-go depends on rtmx-sync), they may not want to expose full requirement details. Grant delegation allows a grantor to specify what data a grantee can see: full requirement text, status-only shadows, or hash-only verification. This builds on REQ-GO-035's shadow requirements by adding access control.

Grants are stored locally in the project's config, consistent with RTMX's local-first philosophy. Zitadel backend integration is deferred to Phase 13.

## Design

### Data Model

```go
// internal/sync/grant.go

// Role defines what operations a grantee can perform.
type Role string
const (
    RoleDependencyViewer  Role = "dependency_viewer"   // See status + deps only
    RoleStatusObserver    Role = "status_observer"      // See status of all reqs
    RoleRequirementEditor Role = "requirement_editor"   // Read + propose changes
    RoleAdmin             Role = "admin"                // Full access
)

// GrantConstraint limits which requirements are visible.
type GrantConstraint struct {
    Categories         []string  `yaml:"categories,omitempty"`          // Whitelist categories
    RequirementIDs     []string  `yaml:"requirement_ids,omitempty"`     // Whitelist specific IDs
    ExcludeCategories  []string  `yaml:"exclude_categories,omitempty"`  // Blacklist categories
    ExpiresAt          string    `yaml:"expires_at,omitempty"`          // ISO date, empty = no expiry
}

// Grant represents a delegation from grantor to grantee.
type Grant struct {
    ID          string          `yaml:"id"`
    Grantee     string          `yaml:"grantee"`       // Remote alias or org identifier
    Role        Role            `yaml:"role"`
    Constraints GrantConstraint `yaml:"constraints"`
    CreatedAt   string          `yaml:"created_at"`
    CreatedBy   string          `yaml:"created_by"`
}
```

### Storage

Grants are stored in `.rtmx/config.yaml` under the sync section:

```yaml
rtmx:
  sync:
    grants:
      - id: grant-001
        grantee: "upstream"
        role: "status_observer"
        constraints:
          categories: ["AUTH", "API"]
        created_at: "2026-03-25"
        created_by: "dev@rtmx.ai"
```

### CLI Commands

```
rtmx grant create ALIAS --role ROLE [--categories CAT1,CAT2] [--ids REQ-1,REQ-2] [--exclude CAT3]
rtmx grant list
rtmx grant revoke GRANT-ID
```

### Visibility Resolution

When a remote resolves a shadow requirement, the grant determines visibility:
- `admin` / `requirement_editor` -> full visibility
- `status_observer` -> status, phase, category only (no requirement text)
- `dependency_viewer` -> status + dependency graph only

Visibility enforcement applies when serving data to a grantee. The local project always has full access to its own data.

### Files to Create

- `internal/sync/grant.go` -- Grant model, constraint checking, role permissions
- `internal/sync/grant_test.go` -- Tests
- `internal/cmd/grant.go` -- CLI commands
- `internal/cmd/grant_test.go` -- CLI tests

### Files to Modify

- `internal/config/config.go` -- Add Grants field to SyncConfig

## Acceptance Criteria

1. `rtmx grant create ALIAS --role ROLE` creates a grant stored in config
2. `rtmx grant list` shows all grants with grantee, role, and constraints
3. `rtmx grant revoke GRANT-ID` removes a grant from config
4. Constraints correctly filter requirements by category and ID whitelist/blacklist
5. Expired grants (past `expires_at`) are treated as revoked
6. Grant roles map to correct visibility levels (full/shadow/hash_only)
7. Duplicate grants for same grantee+role are rejected

## Test Strategy

- **Test Module**: `internal/sync/grants_test.go`
- **Test Function**: `TestGrantDelegation`
- **Validation Method**: Integration Test

### Test Cases

1. Create grant with role and constraints, verify persisted to config
2. List grants shows correct table output
3. Revoke grant removes it from config
4. Constraint filtering: category whitelist allows matching reqs only
5. Constraint filtering: ID whitelist allows specific reqs only
6. Constraint filtering: category blacklist excludes matching reqs
7. Expired grant treated as inactive
8. Duplicate grant rejected with error
9. Unknown alias rejected with error
