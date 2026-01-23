# REQ-ZT-003: JWT Validation in rtmx-sync

## Status: NOT STARTED
## Priority: HIGH
## Phase: 10
## Effort: 2.0 weeks

## Description

rtmx-sync shall validate JWTs on every request to enforce authorization. Tokens issued by Zitadel include identity claims, repository grants, and role permissions. The sync server validates tokens without storing any secrets, using Zitadel's JWKS endpoint for signature verification.

## Acceptance Criteria

- [ ] All API endpoints require valid JWT in Authorization header
- [ ] JWT signature verified against Zitadel JWKS
- [ ] Token expiration checked on every request
- [ ] Issuer and audience claims validated
- [ ] Grant claims extracted for authorization decisions
- [ ] Invalid tokens return 401 Unauthorized
- [ ] Insufficient grants return 403 Forbidden
- [ ] JWKS cached with automatic refresh
- [ ] Token validation adds < 5ms latency

## Test Cases

- `tests/test_jwt.py::TestJWTValidation` - Signature verification tests
- `tests/test_jwt.py::TestClaimsExtraction` - Claim parsing tests
- `tests/test_jwt.py::TestExpiration` - Expiration handling tests
- `tests/test_jwt.py::TestAuthorization` - Grant-based authorization
- `tests/test_jwt.py::TestJWKSCache` - Key caching behavior
- `tests/test_jwt.py::TestPerformance` - Latency benchmarks

## Technical Notes

### JWT Structure

```json
{
  "iss": "https://auth.rtmx.ai",
  "sub": "user-123",
  "aud": "rtmx-sync",
  "exp": 1704067200,
  "iat": 1704063600,
  "email": "user@example.com",
  "grants": {
    "rtmx-ai/rtmx": ["dependency_viewer"],
    "sync-server": ["requirement_editor"]
  },
  "roles": ["rtmx-user"]
}
```

### Middleware Implementation

```python
# rtmx-sync/middleware/auth.py
from jose import jwt, JWTError
import httpx

class JWTAuthMiddleware:
    def __init__(self, issuer: str, audience: str):
        self.issuer = issuer
        self.audience = audience
        self._jwks_cache = None
        self._jwks_expires = 0

    async def __call__(self, request, call_next):
        auth_header = request.headers.get("Authorization")
        if not auth_header or not auth_header.startswith("Bearer "):
            return JSONResponse({"error": "Missing token"}, 401)

        token = auth_header[7:]
        try:
            payload = await self.validate_token(token)
            request.state.user = payload
            request.state.grants = payload.get("grants", {})
        except JWTError as e:
            return JSONResponse({"error": str(e)}, 401)

        return await call_next(request)

    async def validate_token(self, token: str) -> dict:
        jwks = await self.get_jwks()
        return jwt.decode(
            token,
            jwks,
            algorithms=["RS256"],
            issuer=self.issuer,
            audience=self.audience
        )
```

### Authorization Logic

```python
def check_repo_access(grants: dict, repo: str, required_role: str) -> bool:
    """Check if user has required role for repository."""
    repo_grants = grants.get(repo, [])

    # Role hierarchy: admin > editor > observer > viewer
    role_hierarchy = {
        "admin": 4,
        "requirement_editor": 3,
        "status_observer": 2,
        "dependency_viewer": 1
    }

    user_level = max(role_hierarchy.get(r, 0) for r in repo_grants)
    required_level = role_hierarchy.get(required_role, 0)

    return user_level >= required_level
```

### Security Properties

1. Stateless validation - no session storage
2. Short-lived tokens - 1 hour expiration
3. Cryptographic verification - JWKS from Zitadel
4. Fine-grained authorization - per-repo grants
5. Audit trail - all requests logged with user context

## Files to Create/Modify

- `rtmx-sync/src/middleware/auth.py` - JWT validation middleware
- `rtmx-sync/src/auth/jwks.py` - JWKS fetching and caching
- `rtmx-sync/src/auth/grants.py` - Grant extraction and checking
- `tests/test_jwt.py` - Comprehensive tests

## Dependencies

- REQ-ZT-001: Zitadel OIDC integration (issues JWTs)
- REQ-ZT-002: OpenZiti dark service (transport layer)

## Blocks

- REQ-COLLAB-002: Shadow requirements use JWT grants
- REQ-COLLAB-003: Grant delegation stored in JWT claims
