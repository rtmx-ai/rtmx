# REQ-COLLAB-003: Grant Delegation for Cross-Repo Access Control

## Status: NOT STARTED
## Priority: HIGH
## Phase: 10
## Effort: 2.5 weeks

## Description

RTMX shall implement grant delegation following the Zitadel project grants pattern, allowing repository owners to delegate specific access roles to other repositories or organizations. This enables fine-grained cross-organizational collaboration without exposing full requirement details.

## Acceptance Criteria

- [ ] `Grant` dataclass stores delegation configuration (grantor, grantee, roles, constraints)
- [ ] Supported roles: `dependency_viewer`, `status_observer`, `requirement_editor`, `admin`
- [ ] Category constraints limit grants to specific requirement categories
- [ ] `rtmx grant create` command creates new grants
- [ ] `rtmx grant list` shows incoming and outgoing grants
- [ ] `rtmx grant revoke` removes existing grants
- [ ] Grants are stored in Zitadel (not local files)
- [ ] Grant changes propagate immediately via Ziti overlay
- [ ] Audit log tracks all grant operations

## Test Cases

- `tests/test_grants.py::TestGrantModel` - Grant dataclass tests
- `tests/test_grants.py::TestGrantRoles` - Role permission tests
- `tests/test_grants.py::TestGrantConstraints` - Category constraint tests
- `tests/test_grants.py::TestGrantCLI` - CLI command tests
- `tests/test_grants.py::TestGrantPropagation` - Grant sync tests
- `tests/test_grants.py::TestGrantAudit` - Audit logging tests

## Technical Notes

### Grant Model

```python
@dataclass
class Grant:
    grantor: str              # "rtmx-ai/rtmx"
    grantee: str              # "sync-server"
    roles: set[str]           # {"dependency_viewer", "status_observer"}
    constraints: dict         # {"categories": ["COLLABORATION", "CRDT"]}
    created_at: datetime
    created_by: str           # User who created grant
    expires_at: datetime | None = None
```

### Role Permissions

| Role | Can View | Can Update | Can Create | Can Delete |
|------|----------|------------|------------|------------|
| dependency_viewer | deps, status | - | - | - |
| status_observer | all fields read-only | - | - | - |
| requirement_editor | all | status, notes | - | - |
| admin | all | all | yes | yes |

### Zitadel Integration

Grants are stored as Zitadel project grants:
1. Each RTMX repository = Zitadel project
2. Grant creates project grant in Zitadel
3. Roles map to Zitadel roles
4. JWT includes granted roles as claims

## Files to Create/Modify

- `src/rtmx/grants.py` - Grant model and logic
- `src/rtmx/cli/grant.py` - Grant CLI commands
- `src/rtmx/cli/main.py` - Register grant commands
- `src/rtmx/auth/zitadel.py` - Zitadel grant API integration
- `tests/test_grants.py` - Comprehensive tests

## Dependencies

- REQ-COLLAB-001: Cross-repo dependency references
- REQ-ZT-001: Zitadel OIDC integration

## Blocks

- REQ-ZT-003: JWT validation uses grant claims
