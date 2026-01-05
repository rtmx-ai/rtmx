# REQ-SEC-003: OAuth2/SAML SSO Authentication

## Requirement
Enterprise tier shall support OAuth2 and SAML SSO.

## Phase
10 (Collaboration) - Enterprise Security

## Rationale
Enterprise customers require integration with their existing identity providers (Okta, Azure AD, Google Workspace). SSO enables centralized user management, MFA enforcement, and automatic deprovisioning when employees leave.

## Acceptance Criteria
- [ ] OAuth2 authorization code flow supported
- [ ] SAML 2.0 SP-initiated flow supported
- [ ] Integration with Okta verified
- [ ] Integration with Azure AD verified
- [ ] Integration with Google Workspace verified
- [ ] JIT (Just-In-Time) user provisioning works
- [ ] User deprovisioning via SCIM supported
- [ ] MFA status passed through from IdP

## Technical Notes
- Use `python-social-auth` or `authlib` for OAuth2
- Use `python3-saml` for SAML integration
- Store IdP metadata securely
- Implement PKCE for public clients

## Configuration
```yaml
auth:
  providers:
    - type: oauth2
      name: okta
      client_id: ${OKTA_CLIENT_ID}
      issuer: https://company.okta.com
    - type: saml
      name: azure_ad
      metadata_url: https://login.microsoftonline.com/.../metadata
```

## Test Cases
1. OAuth2 login flow completes successfully
2. SAML login flow completes successfully
3. Invalid tokens are rejected
4. Deprovisioned users lose access immediately

## Dependencies
- REQ-COLLAB-001 (sync server exists)

## Effort
3.0 weeks
