# REQ-E2E-003: End-to-End Test Suite

## Status: MISSING
## Priority: HIGH
## Phase: 10

## Description

RTMX shall have a comprehensive E2E test suite that exercises the full platform loop: CLI authentication via Zitadel, requirement synchronization via rtmx-sync over OpenZiti, and concurrent multi-user collaboration scenarios. Tests run against the local Docker Compose stack.

## Rationale

Unit and integration tests validate individual components, but only E2E tests catch issues at the seams — auth token propagation, Ziti service resolution, CRDT merge conflicts under real concurrency, and the full sync round-trip. These tests also validate the deployment architecture that will ship in the Zarf package.

## Acceptance Criteria

- [ ] E2E test suite runs via `make e2e-test` (requires stack running)
- [ ] Tests use pytest with `@pytest.mark.e2e` marker for selective execution
- [ ] Tests auto-obtain auth tokens from local Zitadel (no manual login)
- [ ] Test scenarios cover:
  - Single-user auth flow (login → sync → verify → logout)
  - Multi-user concurrent edits (Alice and Bob edit same RTM)
  - Conflict resolution (concurrent status updates merge correctly)
  - Offline recovery (disconnect → edit → reconnect → sync)
  - Permission enforcement (readonly user cannot push changes)
- [ ] Tests report per-scenario timing for performance regression detection
- [ ] CI can run E2E tests using `docker-compose.ci.yml` overlay
- [ ] Test failures include diagnostic info (container logs, network state)
- [ ] Tests clean up after themselves (no cross-test contamination)
- [ ] E2E tests excluded from regular `make test` (require `make e2e-test`)

## Test Scenarios

### Authentication Flow
1. `tests/e2e/test_auth_flow.py::test_login_pkce_flow` - Full PKCE login against local Zitadel
2. `tests/e2e/test_auth_flow.py::test_token_refresh` - Token auto-refresh before expiry
3. `tests/e2e/test_auth_flow.py::test_logout_clears_tokens` - Logout removes stored tokens
4. `tests/e2e/test_auth_flow.py::test_expired_token_reauth` - Expired token triggers re-login

### Sync Round-Trip
5. `tests/e2e/test_sync_flow.py::test_push_pull_round_trip` - Edit → push → pull on second client
6. `tests/e2e/test_sync_flow.py::test_sync_via_ziti_overlay` - Verify traffic goes through Ziti
7. `tests/e2e/test_sync_flow.py::test_large_database_sync` - Sync with 500+ requirements

### Multi-User Collaboration
8. `tests/e2e/test_collaboration.py::test_concurrent_edits_different_reqs` - No conflict
9. `tests/e2e/test_collaboration.py::test_concurrent_edits_same_req` - CRDT merge
10. `tests/e2e/test_collaboration.py::test_concurrent_status_updates` - Status progression rules

### Access Control
11. `tests/e2e/test_access_control.py::test_readonly_user_cannot_push` - Permission enforcement
12. `tests/e2e/test_access_control.py::test_admin_can_manage_roles` - Role management

### Resilience
13. `tests/e2e/test_resilience.py::test_offline_edit_and_resync` - Offline → reconnect
14. `tests/e2e/test_resilience.py::test_sync_server_restart` - rtmx-sync restart during sync

## Dependencies

- REQ-E2E-001: Local development stack (infrastructure)
- REQ-E2E-002: Zitadel auto-configuration (auth fixtures)
- REQ-ZT-001: Zitadel OIDC integration (CLI auth code)

## Blocks

None

## Effort

3.0 weeks
