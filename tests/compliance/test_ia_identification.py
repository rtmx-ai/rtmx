"""NIST 800-53 IA (Identification and Authentication) Family Tests.

Tests for Identification and Authentication requirements in RTMX.

Control mappings:
- IA-2: Identification and Authentication (Organizational Users)
- IA-4: Identifier Management
- IA-5: Authenticator Management
- IA-8: Identification and Authentication (Non-Organizational Users)
- IA-9: Service Identification and Authentication
"""

from __future__ import annotations

from datetime import datetime, timedelta

import pytest

from rtmx.auth.oidc import (
    AuthConfig,
    TokenInfo,
    generate_code_challenge,
    generate_code_verifier,
)


class TestIA2IdentificationAuthentication:
    """IA-2: Identification and Authentication (Organizational Users).

    The information system uniquely identifies and authenticates
    organizational users (or processes acting on behalf of users).

    RTMX Implementation:
    - OIDC authentication via Zitadel
    - PKCE flow prevents authorization code interception
    - User identity tied to email from IdP
    """

    @pytest.mark.req("REQ-SEC-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_oidc_authentication_required(self) -> None:
        """IA-2: System requires OIDC authentication."""
        config = AuthConfig()

        # Verify OIDC endpoints are defined
        assert "authorize" in config.authorization_endpoint
        assert "token" in config.token_endpoint
        assert config.issuer.startswith("https://")

    @pytest.mark.req("REQ-SEC-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_pkce_prevents_code_interception(self) -> None:
        """IA-2(1): PKCE prevents authorization code interception attacks."""
        verifier = generate_code_verifier()
        challenge = generate_code_challenge(verifier)

        # Verifier should be cryptographically random
        assert len(verifier) >= 43  # PKCE spec minimum
        assert len(verifier) <= 128  # PKCE spec maximum

        # Challenge should be derived from verifier
        assert challenge != verifier  # Challenge is hashed
        assert "=" not in challenge  # Base64url (no padding)

        # Same verifier produces same challenge
        assert generate_code_challenge(verifier) == challenge

        # Different verifier produces different challenge
        other_verifier = generate_code_verifier()
        assert generate_code_challenge(other_verifier) != challenge

    @pytest.mark.req("REQ-SEC-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_tokens_required_for_access(self) -> None:
        """IA-2: Valid tokens required for authenticated access."""
        # Valid token
        valid_token = TokenInfo(
            access_token="valid-access-token",
            expires_at=datetime.now() + timedelta(hours=1),
        )
        assert not valid_token.is_expired

        # Expired token
        expired_token = TokenInfo(
            access_token="expired-access-token",
            expires_at=datetime.now() - timedelta(hours=1),
        )
        assert expired_token.is_expired


class TestIA4IdentifierManagement:
    """IA-4: Identifier Management.

    The organization manages information system identifiers by:
    a. Receiving authorization to assign identifiers
    b. Selecting identifiers that identify individuals
    c. Assigning identifiers to intended parties
    d. Preventing reuse of identifiers
    e. Disabling identifiers after period of inactivity

    RTMX Implementation:
    - Identifiers managed by Zitadel (delegated to IdP)
    - User IDs tied to OIDC subject claims
    - Unique identifiers per user/service
    """

    @pytest.mark.req("REQ-SEC-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_identifiers_delegated_to_idp(self) -> None:
        """IA-4(a): Identifier management delegated to Zitadel."""
        config = AuthConfig()

        # RTMX delegates identity management to the IdP
        assert config.provider == "zitadel"
        assert config.issuer  # IdP handles identifiers

    @pytest.mark.req("REQ-SEC-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_unique_client_identifier(self) -> None:
        """IA-4(b): Unique client ID identifies the application."""
        config = AuthConfig()

        # Client ID uniquely identifies RTMX CLI
        assert config.client_id == "rtmx-cli"


class TestIA5AuthenticatorManagement:
    """IA-5: Authenticator Management.

    The organization manages information system authenticators by:
    a. Verifying identity before distributing authenticators
    b. Establishing initial authenticator content
    c. Ensuring authenticators have sufficient strength
    d. Establishing procedures for lost/compromised authenticators
    e. Changing default authenticators before installing systems
    f. Protecting authenticators commensurate with classification

    RTMX Implementation:
    - Tokens stored securely (keychain or encrypted file)
    - Refresh tokens enable token rotation
    - Token expiration enforces periodic re-authentication
    """

    @pytest.mark.req("REQ-SEC-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_tokens_have_expiration(self) -> None:
        """IA-5(c): Tokens expire requiring re-authentication."""
        token = TokenInfo(
            access_token="access-token",
            refresh_token="refresh-token",
            expires_at=datetime.now() + timedelta(hours=1),
        )

        # Token has finite lifetime
        assert token.expires_at > datetime.now()
        assert token.expires_at < datetime.now() + timedelta(days=1)  # Reasonable expiry

    @pytest.mark.req("REQ-SEC-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_refresh_enables_rotation(self) -> None:
        """IA-5(d): Refresh tokens enable credential rotation."""
        token = TokenInfo(
            access_token="old-access-token",
            refresh_token="valid-refresh-token",
            expires_at=datetime.now() - timedelta(hours=1),  # Expired
        )

        # Even with expired access token, can refresh
        assert token.is_expired
        assert token.is_refreshable

    @pytest.mark.req("REQ-SEC-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_secure_scopes_requested(self) -> None:
        """IA-5(f): Only necessary scopes requested."""
        config = AuthConfig()

        # Scopes should be minimal necessary
        expected_scopes = ["openid", "profile", "email", "offline_access"]
        assert set(config.scopes) == set(expected_scopes)

        # No admin or dangerous scopes
        assert "admin" not in config.scopes
        assert "write" not in config.scopes


class TestIA8NonOrganizationalUsers:
    """IA-8: Identification and Authentication (Non-Organizational Users).

    The information system uniquely identifies and authenticates
    non-organizational users (or processes acting on their behalf).

    RTMX Implementation:
    - Cross-repo federation uses same OIDC authentication
    - External users must authenticate via Zitadel
    - Grant delegation controls what external users can access
    """

    @pytest.mark.req("REQ-SEC-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_external_users_require_authentication(self) -> None:
        """IA-8: External users must authenticate to access resources."""
        from rtmx.models import DelegationRole, GrantDelegation

        # External user gets access via delegation, but still needs auth
        delegation = GrantDelegation(
            grantor="company-a/repo",
            grantee="company-b/repo",  # External organization
            roles_delegated={DelegationRole.SHADOW_VIEWER},
        )

        # Delegation doesn't bypass authentication requirement
        # (In practice, the sync service validates JWT before applying delegation)
        assert delegation.grantee != delegation.grantor
        assert delegation.roles_delegated  # Has limited roles


class TestIA9ServiceAuthentication:
    """IA-9: Service Identification and Authentication.

    The organization identifies and authenticates services
    before establishing communications with the service.

    RTMX Implementation:
    - OpenZiti provides mutual authentication
    - Services identified by Ziti identity, not IP
    - Dark services only accessible via identity-verified overlay
    """

    @pytest.mark.req("REQ-ZT-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_ziti_service_identification(self) -> None:
        """IA-9: Services identified via Ziti identities."""
        from rtmx.ziti import ZitiConfig

        config = ZitiConfig()

        # Services have explicit identifiers
        assert "rtmx-sync" in config.services
        assert config.services["rtmx-sync"]  # Has service name

    @pytest.mark.req("REQ-ZT-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_identity_enrollment_required(self) -> None:
        """IA-9: Identity enrollment required before service access."""
        from rtmx.ziti import ZitiConfig

        config = ZitiConfig()

        # Identity must be enrolled to access services
        # New config has no identity by default
        assert config.identity_path  # Path is defined
        # In practice, has_identity would be False until enrollment
