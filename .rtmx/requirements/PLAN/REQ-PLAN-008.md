# REQ-PLAN-008: JWT Identity Extraction

## Metadata
- **Category**: PLAN
- **Subcategory**: Identity
- **Priority**: HIGH
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: (none)
- **Blocks**: REQ-PLAN-009

## Requirement

RTMX shall provide a `CurrentUser()` function that extracts user identity
from the stored OIDC ID token. The ID token (JWT) is already stored by
`rtmx auth login` at `~/.rtmx/auth/tokens.json` but is never decoded.

## Design

```go
// internal/auth/identity.go

type UserIdentity struct {
    Sub   string // OIDC subject identifier
    Email string // User email
    Name  string // Display name
}

// CurrentUser loads the stored ID token and decodes JWT claims.
// Returns empty identity (not error) when no token exists.
func CurrentUser(tokenStorePath string) (UserIdentity, error)
```

JWT decode is base64 payload extraction only -- no signature re-verification.
The OIDC provider validated the token at login time. The existing
`auth.TokenSet.IDToken` field (internal/auth/oidc.go:62) holds the JWT.

## Acceptance Criteria

1. `CurrentUser()` returns identity from stored ID token
2. Missing or expired token returns empty identity, not error
3. Malformed JWT returns error
4. No external dependencies (standard library base64 + json decode)

## Files to Create

- `internal/auth/identity.go`
- `internal/auth/identity_test.go`
