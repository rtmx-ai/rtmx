# REQ-GO-079: Grant Command Tests

## Metadata
- **Category**: GO
- **Subcategory**: CLI
- **Priority**: HIGH
- **Phase**: 29
- **Status**: MISSING
- **Dependencies**: REQ-GO-061
- **Blocks**: (none)

## Requirement

The `grant` command (create, list, revoke) shall have command-level tests
covering role validation, category/ID constraint filtering, expiry date
parsing, config file writing, and all error paths. Currently this command
has zero test coverage.

## Acceptance Criteria

1. Test `grant create` with valid parameters writes to config.
2. Test `grant create` with invalid role returns error.
3. Test `grant create` with expiry date parses correctly.
4. Test `grant list` with no grants shows empty message.
5. Test `grant list` with grants displays table.
6. Test `grant revoke` removes grant from config.
7. Test `grant revoke` with nonexistent ID returns error.
8. Test category and requirement ID constraints are stored.

## Files to Create/Modify

- `internal/cmd/grant_test.go` -- Command-level grant tests

## Effort Estimate

0.25 weeks
