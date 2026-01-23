# REQ-COLLAB-001: Cross-Repository Dependency References

## Status: COMPLETE
## Priority: HIGH
## Phase: 10
## Effort: 1.5 weeks

## Description

RTMX shall support cross-repository requirement dependencies, allowing requirements in one repository to depend on requirements in another repository. This enables enterprise-scale requirements traceability across organizational boundaries while maintaining clear dependency chains.

## Acceptance Criteria

- [x] `RemoteConfig` dataclass stores remote repository configuration (alias, repo, path, database)
- [x] `SyncConfig.remotes` dictionary maps aliases to `RemoteConfig` instances
- [x] `parse_requirement_ref()` parses local, aliased, and full-repo reference formats
- [x] Supported reference formats:
  - Local: `REQ-SW-001`
  - Aliased: `sync:REQ-SYNC-001`
  - Full repo: `sync-server:REQ-SYNC-001`
- [x] `validate_cross_repo_deps()` validates cross-repo dependencies
- [x] Graceful degradation when remote is unavailable (warning, not error)
- [x] `Requirement.is_blocked()` checks cross-repo dependency status
- [x] CLI commands: `rtmx remote add`, `rtmx remote remove`, `rtmx remote list`

## Test Cases

- `tests/test_cross_repo.py::TestRemoteConfig` - RemoteConfig dataclass tests
- `tests/test_cross_repo.py::TestSyncConfigRemotes` - SyncConfig remotes tests
- `tests/test_cross_repo.py::TestRequirementRef` - RequirementRef dataclass tests
- `tests/test_cross_repo.py::TestParseRequirementRef` - Reference parsing tests
- `tests/test_cross_repo.py::TestValidateCrossRepoDeps` - Cross-repo validation tests
- `tests/test_cross_repo.py::TestIsBlockedCrossRepo` - Cross-repo blocking tests
- `tests/test_cross_repo.py::TestCLIRemoteCommands` - CLI remote command tests

## Technical Notes

### Configuration Format

```yaml
rtmx:
  sync:
    remotes:
      sync:
        repo: sync-server
        path: ../rtmx-sync
        database: .rtmx/database.csv
```

### Reference Resolution

1. Parse reference string to extract alias/repo and req_id
2. Look up remote configuration by alias (or match full repo path)
3. Load remote database from configured path
4. Resolve requirement and check status

### Graceful Degradation

When a remote is unavailable:
- Validation issues warnings (not errors)
- `is_blocked()` returns False (can't verify, assume not blocked)
- User can still work offline

## Files Created/Modified

- `src/rtmx/config.py` - Added `RemoteConfig`, updated `SyncConfig`
- `src/rtmx/parser.py` - Added `RequirementRef`, `parse_requirement_ref()`
- `src/rtmx/validation.py` - Added `validate_cross_repo_deps()`
- `src/rtmx/models.py` - Updated `is_blocked()` for cross-repo
- `src/rtmx/cli/remote.py` - New CLI commands
- `src/rtmx/cli/main.py` - Registered remote command group
- `tests/test_cross_repo.py` - Comprehensive test suite

## Dependencies

None (foundation requirement)

## Blocks

- REQ-COLLAB-002: Shadow requirements for partial visibility
- REQ-COLLAB-003: Grant delegation for access control
- REQ-ZT-001: Zitadel OIDC integration
