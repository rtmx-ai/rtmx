# REQ-COLLAB-007: Browser-Based CLI Authentication

## Description

CLI shall authenticate users via browser-based OAuth/OIDC flow, storing tokens locally while persisting sync endpoint configuration in versioned project config.

## Rationale

Developers expect modern authentication UX similar to `gcloud auth login`, `gh auth login`, and Claude Code. Browser-based OAuth provides:
- Familiar SSO experience (Google, GitHub, Okta, etc.)
- No passwords in terminal history
- Organization-controlled token policies
- Seamless enterprise IdP integration

## Acceptance Criteria

### Authentication Flow
- [ ] `rtmx sync login` opens system browser to auth endpoint
- [ ] CLI displays fallback URL if browser cannot open automatically
- [ ] Local HTTP server on ephemeral port receives OAuth callback
- [ ] Device code flow available for headless/SSH environments
- [ ] Authentication completes within 5 minutes or times out gracefully
- [ ] Success message displays authenticated user, org, and token expiry

### Token Storage
- [ ] Access token stored in platform-appropriate secure storage:
  - macOS: Keychain
  - Linux: Secret Service (libsecret) or encrypted file
  - Windows: Credential Manager
- [ ] Refresh token stored alongside access token
- [ ] Token expiry timestamp tracked for proactive refresh
- [ ] Tokens are NEVER written to version-controlled files
- [ ] `~/.config/rtmx/credentials.yaml` used as fallback with 600 permissions

### Configuration Management
- [ ] Sync endpoint saved to `rtmx.yaml` (versioned with project)
- [ ] Endpoint derived from git remote origin when not explicitly set
- [ ] Endpoint format: `wss://sync.rtmx.io/{provider}/{org}/{project}`
- [ ] Manual endpoint override supported via `sync.endpoint` config key

### Token Lifecycle
- [ ] Automatic token refresh when access token expires but refresh token valid
- [ ] Background refresh triggered before expiry (at 75% of TTL)
- [ ] Re-authentication prompt when refresh token expires
- [ ] Token expiry follows OIDC provider policy (org-configurable)
- [ ] Default expiry: 30 days for free tier, org-defined for enterprise

### Git Remote Integration
- [ ] Detect git remote origin change on `rtmx sync` commands
- [ ] Warn user if configured endpoint doesn't match current remote
- [ ] Prompt to run `rtmx sync login` to update endpoint
- [ ] Support GitHub, GitLab, Bitbucket remote URL formats

### CLI Commands
- [ ] `rtmx sync login` - Initiate browser OAuth flow
- [ ] `rtmx sync logout` - Clear local tokens
- [ ] `rtmx sync status` - Show connection state, user, token expiry
- [ ] `rtmx sync whoami` - Display authenticated identity

### Error Handling
- [ ] Clear error message if auth server unreachable
- [ ] Graceful handling of OAuth callback timeout
- [ ] Informative message if user denies OAuth consent
- [ ] Retry logic for transient network failures

## Technical Design

### OAuth Callback Flow

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│   CLI    │────>│ Browser  │────>│   IdP    │────>│ Callback │
│          │     │          │     │ (SSO)    │     │ Server   │
└──────────┘     └──────────┘     └──────────┘     └──────────┘
     │                                                   │
     │<──────────────── Token ───────────────────────────┘
     │
     ▼
┌──────────┐
│ Keychain │
└──────────┘
```

### Configuration Schema

```yaml
# rtmx.yaml (versioned)
sync:
  endpoint: wss://sync.rtmx.io/github/rtmx-ai/rtmx
  # Optional overrides
  provider: cloud  # cloud | self-hosted
  auto_connect: true
```

```yaml
# ~/.config/rtmx/credentials.yaml (NOT versioned, 600 perms)
tokens:
  sync.rtmx.io:
    access_token: eyJ...
    refresh_token: eyJ...
    expires_at: 2026-01-17T00:00:00Z
    issuer: https://auth.rtmx.ai
    user: ryan@iotactical.co
    org: iotactical
```

### Device Code Flow (Headless)

```
$ rtmx sync login

No browser detected. Use device code flow:

  1. Visit: https://sync.rtmx.io/device
  2. Enter code: ABCD-1234

Waiting for authentication... ✓
```

### Endpoint Derivation Logic

```python
def derive_endpoint(git_remote: str) -> str:
    """
    git@github.com:rtmx-ai/rtmx.git
    → wss://sync.rtmx.io/github/rtmx-ai/rtmx

    https://gitlab.com/org/project.git
    → wss://sync.rtmx.io/gitlab/org/project
    """
```

## Dependencies

- REQ-COLLAB-001: Sync server must exist to authenticate against
- REQ-SEC-003: OAuth2/SAML SSO infrastructure

## Test Cases

### Unit Tests
1. Parse git remote URL formats (GitHub, GitLab, Bitbucket, SSH, HTTPS)
2. Derive endpoint from remote URL
3. Detect remote URL change
4. Token expiry calculation
5. Credentials file permission validation

### Integration Tests
1. Complete OAuth flow with mock IdP
2. Token refresh before expiry
3. Keychain storage and retrieval
4. Device code flow completion
5. Callback server ephemeral port binding

### System Tests
1. End-to-end login with real OAuth provider (staging)
2. Cross-platform keychain integration
3. SSH environment device code fallback
4. Token expiry and re-authentication prompt

## Security Considerations

- Callback server binds only to localhost (127.0.0.1)
- State parameter prevents CSRF attacks
- PKCE (Proof Key for Code Exchange) required for public clients
- Tokens encrypted at rest in keychain
- No tokens in shell history, environment variables, or logs

## Effort Estimate

2.5 weeks:
- OAuth flow + callback server: 1 week
- Keychain integration (cross-platform): 0.5 weeks
- Token refresh + lifecycle: 0.5 weeks
- CLI commands + tests: 0.5 weeks

## Notes

This requirement establishes the developer-facing authentication experience for RTMX Sync. The pattern follows industry standards (gcloud, gh, claude) to minimize learning curve and maximize security.
