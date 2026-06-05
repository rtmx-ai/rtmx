# REQ-GO-078: Auth Command Tests

## Metadata
- **Category**: GO
- **Subcategory**: CLI
- **Priority**: HIGH
- **Phase**: 29
- **Status**: MISSING
- **Dependencies**: REQ-GO-061
- **Blocks**: (none)

## Requirement

The `auth` command (login, status, logout) shall have command-level tests
covering flag parsing, error messages when issuer/client_id are not
configured, token file paths, and the oidcClientFactory injection seam.
Currently this command has zero test coverage at the command wiring level.

## Acceptance Criteria

1. Test `auth login` with missing config (no issuer/client_id).
2. Test `auth status` with no stored token.
3. Test `auth status` with a valid stored token.
4. Test `auth logout` clears token.
5. Test oidcClientFactory injection seam is exercisable.
6. Test error output formatting for each failure mode.

## Files to Create/Modify

- `internal/cmd/auth_test.go` -- Command-level auth tests

## Effort Estimate

0.25 weeks
