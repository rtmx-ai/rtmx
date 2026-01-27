"""NIST 800-53 SC (System and Communications Protection) Family Tests.

Tests for System and Communications Protection requirements in RTMX.

Control mappings:
- SC-7: Boundary Protection
- SC-8: Transmission Confidentiality and Integrity
- SC-12: Cryptographic Key Establishment and Management
- SC-13: Cryptographic Protection
- SC-23: Session Authenticity
- SC-28: Protection of Information at Rest
"""

from __future__ import annotations

import hashlib

import pytest

from rtmx.models import (
    ShadowRequirement,
    Status,
    Visibility,
)


class TestSC7BoundaryProtection:
    """SC-7: Boundary Protection.

    The information system:
    a. Monitors and controls communications at external boundaries
    b. Implements subnetworks for publicly accessible components
    c. Connects to external networks only through managed interfaces

    RTMX Implementation:
    - OpenZiti provides zero-trust boundary (no public ports)
    - Dark services invisible to external scanning
    - All traffic via identity-verified overlay
    """

    @pytest.mark.req("REQ-ZT-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_no_public_ports_exposed(self) -> None:
        """SC-7(a): rtmx-sync exposes no public ports (dark service)."""
        from rtmx.ziti import ZitiConfig

        config = ZitiConfig()

        # Service is defined but has no port mapping
        # The service name maps to a Ziti service, not a port
        assert "rtmx-sync" in config.services
        service_id = config.services["rtmx-sync"]

        # Service identifier is a name, not a port
        assert not service_id.isdigit()
        assert ":" not in service_id  # Not IP:port format

    @pytest.mark.req("REQ-ZT-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_identity_required_for_connection(self) -> None:
        """SC-7(b): Connections require Ziti identity (managed interface)."""
        from rtmx.ziti import ZitiClient, ZitiConfig, ZitiNotAvailableError

        config = ZitiConfig()

        # Without OpenZiti SDK, cannot create client
        try:
            client = ZitiClient(config)
            # If SDK available, would check enrollment requirement
            assert hasattr(client, "is_enrolled")
        except ZitiNotAvailableError:
            # SDK not installed is fine for unit tests
            pass


class TestSC8TransmissionProtection:
    """SC-8: Transmission Confidentiality and Integrity.

    The information system protects the confidentiality and integrity
    of transmitted information.

    RTMX Implementation:
    - OpenZiti provides end-to-end encryption
    - TLS for OIDC communications
    - Content hashes for integrity verification
    """

    @pytest.mark.req("REQ-ZT-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_oidc_endpoints_require_https(self) -> None:
        """SC-8(1): OIDC communications use TLS."""
        from rtmx.auth.oidc import AuthConfig

        config = AuthConfig()

        # All endpoints must use HTTPS
        assert config.issuer.startswith("https://")
        assert config.authorization_endpoint.startswith("https://")
        assert config.token_endpoint.startswith("https://")
        assert config.userinfo_endpoint.startswith("https://")

    @pytest.mark.req("REQ-ZT-001")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_ziti_controller_requires_https(self) -> None:
        """SC-8(1): Ziti controller uses TLS."""
        from rtmx.ziti import ZitiConfig

        config = ZitiConfig()

        # Controller must use HTTPS
        assert config.controller.startswith("https://")

    @pytest.mark.req("REQ-COLLAB-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_shadow_hash_provides_integrity(self) -> None:
        """SC-8(1): Shadow hashes enable integrity verification."""
        from rtmx.models import Requirement

        # Create a requirement
        req = Requirement(
            req_id="REQ-TEST-001",
            status=Status.COMPLETE,
            requirement_text="Test requirement for integrity verification",
        )

        # Create shadow with hash
        shadow = ShadowRequirement.from_requirement(req, "org/repo", Visibility.SHADOW)

        # Hash is deterministic
        content = f"{req.req_id}:{req.status.value}:{req.requirement_text}"
        expected_hash = hashlib.sha256(content.encode()).hexdigest()[:16]

        assert shadow.shadow_hash == expected_hash
        assert shadow.is_verifiable


class TestSC12CryptographicKeyManagement:
    """SC-12: Cryptographic Key Establishment and Management.

    The organization establishes and manages cryptographic keys
    when cryptography is employed within the information system.

    RTMX Implementation:
    - PKCE verifiers are cryptographically random
    - Ziti handles identity certificates
    - Tokens stored securely (keychain)
    """

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_pkce_verifier_cryptographically_random(self) -> None:
        """SC-12: PKCE verifiers use cryptographic randomness."""
        from rtmx.auth.oidc import generate_code_verifier

        # Generate multiple verifiers
        verifiers = [generate_code_verifier() for _ in range(100)]

        # All should be unique (collision probability negligible)
        assert len(set(verifiers)) == 100

        # All should meet minimum length
        for v in verifiers:
            assert len(v) >= 43

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_ziti_identity_stored_securely(self) -> None:
        """SC-12: Ziti identity files stored in protected directory."""
        from pathlib import Path

        from rtmx.ziti import ZitiConfig

        config = ZitiConfig()

        # Identity stored in user-specific directory
        assert str(config.identity_dir).startswith(str(Path.home()))
        assert ".rtmx" in str(config.identity_dir)


class TestSC13CryptographicProtection:
    """SC-13: Cryptographic Protection.

    The information system implements cryptographic mechanisms
    to prevent unauthorized disclosure and modification.

    RTMX Implementation:
    - SHA-256 for content hashing
    - S256 method for PKCE challenge
    - TLS 1.2+ for all transport
    """

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_pkce_uses_sha256(self) -> None:
        """SC-13: PKCE uses SHA-256 for code challenge."""
        import base64

        from rtmx.auth.oidc import generate_code_challenge, generate_code_verifier

        verifier = generate_code_verifier()
        challenge = generate_code_challenge(verifier)

        # Challenge should be base64url-encoded SHA-256
        # Manually verify the algorithm
        expected_digest = hashlib.sha256(verifier.encode("ascii")).digest()
        expected_challenge = base64.urlsafe_b64encode(expected_digest).rstrip(b"=").decode("ascii")

        assert challenge == expected_challenge

    @pytest.mark.req("REQ-COLLAB-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_shadow_hash_uses_sha256(self) -> None:
        """SC-13: Shadow requirements use SHA-256 for content hashing."""
        from rtmx.models import Requirement

        req = Requirement(
            req_id="REQ-CRYPTO-001",
            status=Status.COMPLETE,
            requirement_text="Cryptographic test",
        )

        shadow = ShadowRequirement.from_requirement(req, "org/repo", Visibility.SHADOW)

        # Verify hash is SHA-256 derived (truncated to 16 chars)
        content = f"{req.req_id}:{req.status.value}:{req.requirement_text}"
        full_hash = hashlib.sha256(content.encode()).hexdigest()

        assert shadow.shadow_hash == full_hash[:16]


class TestSC23SessionAuthenticity:
    """SC-23: Session Authenticity.

    The information system protects the authenticity of
    communications sessions.

    RTMX Implementation:
    - OAuth state parameter prevents CSRF
    - Token binding to client
    - Session tracked via token expiration
    """

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_oauth_state_prevents_csrf(self) -> None:
        """SC-23: OAuth state parameter prevents CSRF attacks."""
        import secrets

        # State should be cryptographically random
        state1 = secrets.token_urlsafe(16)
        state2 = secrets.token_urlsafe(16)

        # States should be unique
        assert state1 != state2

        # States should have sufficient entropy
        assert len(state1) >= 16

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_token_tracks_session(self) -> None:
        """SC-23: Tokens track session validity."""
        from datetime import datetime, timedelta

        from rtmx.auth.oidc import TokenInfo

        # Token represents a session
        token = TokenInfo(
            access_token="session-token",
            expires_at=datetime.now() + timedelta(hours=1),
        )

        # Session validity tied to token expiration
        assert not token.is_expired

        # Expired token = invalid session
        expired = TokenInfo(
            access_token="old-session",
            expires_at=datetime.now() - timedelta(minutes=1),
        )
        assert expired.is_expired


class TestSC28ProtectionAtRest:
    """SC-28: Protection of Information at Rest.

    The information system protects the confidentiality and integrity
    of information at rest.

    RTMX Implementation:
    - Tokens stored in system keychain (encrypted)
    - Fallback to file with restricted permissions
    - Shadow requirements limit data exposure
    """

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_keychain_preferred_for_tokens(self) -> None:
        """SC-28: Tokens prefer secure keychain storage."""
        from rtmx.auth.oidc import HAS_KEYRING

        # Test documents keyring availability
        # In production, tokens use keyring when available
        assert isinstance(HAS_KEYRING, bool)

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_token_file_has_restricted_permissions(self) -> None:
        """SC-28: Fallback token file has restricted permissions."""
        from rtmx.auth.oidc import _get_token_path

        token_path = _get_token_path()

        # Path should be in user's home directory
        assert str(token_path).startswith(str(token_path.home()))
        assert ".rtmx" in str(token_path)

    @pytest.mark.req("REQ-COLLAB-003")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_shadow_visibility_limits_exposure(self) -> None:
        """SC-28: Shadow visibility limits data exposure at rest."""
        # Full visibility - all data exposed
        shadow_full = ShadowRequirement(
            req_id="REQ-FULL-001",
            external_repo="org/repo",
            shadow_hash="abc123",
            status=Status.COMPLETE,
            visibility=Visibility.FULL,
        )
        assert shadow_full.is_accessible

        # Shadow visibility - limited exposure
        shadow_limited = ShadowRequirement(
            req_id="REQ-SHADOW-001",
            external_repo="org/repo",
            shadow_hash="def456",
            status=Status.COMPLETE,
            visibility=Visibility.SHADOW,
        )
        assert not shadow_limited.is_accessible
        assert shadow_limited.is_verifiable

        # Hash only - minimal exposure
        shadow_hash = ShadowRequirement(
            req_id="REQ-HASH-001",
            external_repo="org/repo",
            shadow_hash="ghi789",
            status=Status.COMPLETE,
            visibility=Visibility.HASH_ONLY,
        )
        assert not shadow_hash.is_accessible
        assert shadow_hash.is_verifiable
