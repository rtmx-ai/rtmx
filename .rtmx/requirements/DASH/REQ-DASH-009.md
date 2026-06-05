# REQ-DASH-009: Authentication Middleware

## Metadata
- **Category**: DASH
- **Subcategory**: Security
- **Priority**: HIGH
- **Phase**: 28
- **Status**: MISSING
- **Dependencies**: REQ-DASH-001
- **Blocks**: (none)

## Requirement

The web dashboard shall implement authentication middleware that enforces
access control when the `--auth` flag is set, supporting API key
authentication for programmatic access and OAuth2/OIDC for browser-based
access, consistent with the deployment topology's auth requirements.

## Rationale

The `--auth` flag already exists on `rtmx serve` but is not wired to any
middleware. For self-managed and air-gapped deployments, the dashboard must
enforce access control. API key auth is needed for programmatic integration
(webhooks, external tools). OAuth2/OIDC is needed for browser-based access
in managed and self-managed deployments.

## Design

### API Key Auth (`--auth api-key`)

```
X-API-Key: rtmx_sk_...
```

- API keys are stored in `.rtmx/api-keys.json` (hashed with bcrypt)
- `rtmx serve --auth api-key --generate-key` creates a new key
- All `/api/*` and `/ws` endpoints require valid key
- Static assets (CSS, JS) are served without auth

### OAuth2/OIDC Auth (`--auth oauth`)

- Configurable OIDC provider via `rtmx.yaml` auth section
- Browser flow: redirect to provider -> callback -> session cookie
- Session stored in signed, encrypted cookie (no server-side session store)
- JWT validation for token-based access
- Configurable allowed email domains or OIDC groups

### Middleware Chain

```go
func authMiddleware(mode string, cfg *config.AuthConfig) func(http.Handler) http.Handler {
    switch mode {
    case "api-key":
        return apiKeyMiddleware(cfg)
    case "oauth":
        return oauthMiddleware(cfg)
    default:
        return noopMiddleware // no auth when flag not set
    }
}
```

### Security Requirements

- API keys: minimum 32 bytes, cryptographically random
- Cookies: Secure, HttpOnly, SameSite=Strict
- HTTPS required when auth is enabled (reject HTTP with warning)
- Rate limiting on auth endpoints (10 attempts/minute per IP)
- No plaintext key storage

## Acceptance Criteria

1. Without `--auth`: all endpoints accessible without credentials.
2. With `--auth api-key`: API endpoints require valid X-API-Key header.
3. With `--auth oauth`: browser users redirected to OIDC provider.
4. Invalid API key returns 401 with no information leakage.
5. OAuth callback sets secure session cookie.
6. Static assets served without authentication.
7. `--generate-key` creates and displays a new API key.
8. Rate limiting blocks brute-force attempts.
9. HTTP rejected with warning when auth is enabled.

## Files to Create/Modify

- `internal/cmd/serve_auth.go` -- Auth middleware implementations
- `internal/cmd/serve_auth_test.go` -- Auth middleware tests
- `internal/cmd/serve.go` -- Wire auth middleware into handler chain

## Effort Estimate

1 week

## Test Strategy

- API key: valid key passes, invalid key returns 401
- Missing key: returns 401 (not 500)
- OAuth flow: mock OIDC provider, verify redirect and callback
- Session cookie: verify Secure, HttpOnly, SameSite flags
- Rate limiting: exceed threshold, verify 429 response
- Static assets: verify no auth required for CSS/JS
- No auth mode: verify all endpoints accessible
