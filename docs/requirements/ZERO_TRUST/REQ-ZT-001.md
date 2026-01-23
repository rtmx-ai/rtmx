# REQ-ZT-001: Zitadel OIDC Integration for CLI Authentication

## Status: NOT STARTED
## Priority: HIGH
## Phase: 10
## Effort: 3.0 weeks

## Description

RTMX CLI shall integrate with Zitadel as the identity provider for zero-trust authentication. The CLI shall implement OIDC PKCE flow for secure authentication without storing client secrets, supporting SSO via GitHub, SAML, and other enterprise identity providers configured in Zitadel.

## Acceptance Criteria

- [ ] `rtmx sync login` initiates OIDC PKCE flow
- [ ] Local callback server receives authorization code
- [ ] Browser opens Zitadel authorize endpoint automatically
- [ ] Tokens exchanged and stored securely in system keychain
- [ ] `rtmx sync logout` clears stored tokens
- [ ] `rtmx sync status` shows authenticated user info
- [ ] Token refresh happens automatically before expiration
- [ ] Support for GitHub, Google, and SAML federation via Zitadel
- [ ] Offline token caching for limited offline access

## Test Cases

- `tests/test_auth.py::TestOIDCFlow` - PKCE flow implementation tests
- `tests/test_auth.py::TestTokenStorage` - Keychain storage tests
- `tests/test_auth.py::TestTokenRefresh` - Automatic refresh tests
- `tests/test_auth.py::TestLoginCLI` - CLI login command tests
- `tests/test_auth.py::TestLogoutCLI` - CLI logout command tests
- `tests/test_auth.py::TestOfflineTokens` - Offline access tests

## Technical Notes

### OIDC PKCE Flow

```python
async def login():
    # 1. Generate PKCE code verifier and challenge
    verifier = secrets.token_urlsafe(32)
    challenge = base64url(sha256(verifier))

    # 2. Start local callback server on random port
    server = await start_callback_server()
    redirect_uri = f"http://localhost:{server.port}/callback"

    # 3. Open browser to Zitadel authorize endpoint
    auth_url = (
        f"{ISSUER}/oauth/v2/authorize?"
        f"client_id={CLIENT_ID}&"
        f"redirect_uri={redirect_uri}&"
        f"response_type=code&"
        f"scope=openid profile email&"
        f"code_challenge={challenge}&"
        f"code_challenge_method=S256"
    )
    webbrowser.open(auth_url)

    # 4. Wait for callback with authorization code
    code = await server.wait_for_code()

    # 5. Exchange code for tokens
    tokens = await exchange_code(code, verifier, redirect_uri)

    # 6. Store tokens in keychain
    keyring.set_password("rtmx", "access_token", tokens.access_token)
    keyring.set_password("rtmx", "refresh_token", tokens.refresh_token)
```

### Configuration

```yaml
# .rtmx/config.yaml
rtmx:
  auth:
    provider: zitadel
    issuer: https://auth.rtmx.ai
    client_id: rtmx-cli
    # No client_secret (public client with PKCE)
```

### Token Storage

- Access token: Short-lived (1 hour), stored in keychain
- Refresh token: Long-lived (30 days), stored in keychain
- ID token: Used for user info display
- Tokens encrypted at rest via system keychain

### Security Properties

1. No client secret stored (PKCE eliminates need)
2. Tokens never logged or printed
3. Keychain protects at-rest storage
4. Refresh tokens can be revoked server-side

## Files to Create/Modify

- `src/rtmx/auth/__init__.py` - Auth module
- `src/rtmx/auth/oidc.py` - OIDC PKCE implementation
- `src/rtmx/auth/keychain.py` - Secure token storage
- `src/rtmx/cli/sync.py` - Add login/logout commands
- `tests/test_auth.py` - Comprehensive tests

## Dependencies

- REQ-COLLAB-001: Cross-repo dependency references

## Blocks

- REQ-ZT-002: OpenZiti enrollment uses Zitadel JWT
- REQ-ZT-003: JWT validation in rtmx-sync
- REQ-COLLAB-003: Grant delegation uses Zitadel
