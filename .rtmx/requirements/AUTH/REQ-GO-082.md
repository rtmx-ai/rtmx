# REQ-GO-082: Managed CLI login with loopback OAuth

## Metadata
- **Category**: AUTH
- **Subcategory**: Managed
- **Priority**: P0
- **Phase**: 8
- **Status**: COMPLETE
- **Dependencies**: REQ-GO-040
- **External ID**: rtmx-ai/REQ-MONO-021a

## Requirement

Go CLI SHALL provide zero-config managed login (`rtmx login` or
`rtmx auth login --managed`) using browser OAuth against the managed
sync host with localhost loopback completion, reusing PKCE/loopback
machinery from REQ-GO-040.

## Acceptance Criteria

1. [x] Managed defaults require no custom `auth.issuer` in rtmx.yaml.
2. [x] Browser + loopback complete without manual code paste.
3. [x] Tokens/session stored in the existing auth token store.
4. [x] `rtmx auth status` / `logout` work for the managed credential.
5. [x] Table-driven / injectable tests (no real browser in unit tests).

## Files (expected)

- `internal/cmd/auth.go` (and/or `login.go`)
- `internal/auth/` managed client wrapping OIDC or sync-mediated flow
- `internal/cmd/auth_managed_test.go`

## Effort Estimate

1.25 weeks
