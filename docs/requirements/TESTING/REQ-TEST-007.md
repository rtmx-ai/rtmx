# REQ-TEST-007: Sync Command E2E Tests

## Status: NOT_STARTED
## Priority: HIGH
## Phase: 4
## Effort: 1.0 weeks

## Description

Add comprehensive E2E tests for the sync command covering GitHub and Jira adapters with mocked HTTP responses.

## Acceptance Criteria

- [ ] E2E test for GitHub import (mocked HTTP)
- [ ] E2E test for GitHub export (mocked HTTP)
- [ ] E2E test for Jira import (mocked HTTP)
- [ ] E2E test for Jira export (mocked HTTP)
- [ ] E2E test for bidirectional sync with conflict resolution
- [ ] E2E test for --dry-run mode
- [ ] E2E test for network error handling
- [ ] E2E test for authentication failure
- [ ] At least 12 new scope_system tests

## Test Scenarios

### GitHub Adapter
1. Import issues from GitHub to RTM database
2. Export requirements to GitHub issues
3. Handle rate limiting gracefully
4. Handle authentication errors

### Jira Adapter
1. Import tickets from Jira project
2. Export requirements as Jira tickets
3. Handle Jira server vs cloud differences
4. Handle authentication errors

### Sync Logic
1. Bidirectional sync with --prefer-local
2. Bidirectional sync with --prefer-remote
3. Dry-run shows changes without applying
4. Network timeout handling

## Files to Create

- `tests/test_sync_e2e.py`

## Dependencies

- REQ-TEST-006 (bug fixes must be complete first)

## Notes

All HTTP interactions must be mocked - no real API calls in tests.
