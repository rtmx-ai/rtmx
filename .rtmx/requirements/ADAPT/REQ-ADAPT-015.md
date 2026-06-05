# REQ-ADAPT-015: Adapter Factory Integration Tests

## Metadata
- **Category**: ADAPT
- **Subcategory**: Testing
- **Priority**: HIGH
- **Phase**: 29
- **Status**: MISSING
- **Dependencies**: REQ-ADAPT-014
- **Blocks**: (none)

## Requirement

The `getAdapter()` factory function in sync.go shall have integration tests
that verify the full config-to-adapter constructor path for all supported
services. Currently sync command tests use a mockAdapter, and adapter tests
use mock HTTP -- the factory that connects them is untested.

## Acceptance Criteria

1. Test `getAdapter()` for each service: github, jira, asana, monday, gitlab.
2. Test unknown service returns error.
3. Test disabled adapter returns error.
4. Test missing token returns error (via config with token env unset).
5. Tests exercise real adapter constructors (not mocks).

## Files to Create/Modify

- `internal/cmd/sync_factory_test.go` -- Factory integration tests

## Effort Estimate

0.25 weeks
