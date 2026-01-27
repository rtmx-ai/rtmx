"""OpenZiti client for RTMX.

Provides zero-trust network connectivity via OpenZiti overlay.

Security model:
- No listening ports on client or server
- All traffic is end-to-end encrypted
- Identity-based policies control access
- Server is a "dark service" invisible to port scanners

Architecture:
    ┌──────────────┐     Ziti Overlay     ┌──────────────┐
    │  rtmx CLI    │◄───────────────────►│  rtmx-sync   │
    │  (client)    │  (encrypted tunnel)  │  (dark svc)  │
    └──────────────┘                      └──────────────┘
          │                                      │
          │ Identity enrolled                    │ No public ports
          │ from OIDC token                      │ Ziti SDK only
          ▼                                      ▼
    ┌─────────────────────────────────────────────────────┐
    │           OpenZiti Controller + Edge Routers        │
    │  (manages identities, policies, service discovery)  │
    └─────────────────────────────────────────────────────┘
"""

from __future__ import annotations

import json
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

# Check if OpenZiti SDK is available
try:
    import openziti

    HAS_OPENZITI = True
except ImportError:
    HAS_OPENZITI = False


class ZitiError(Exception):
    """Base exception for Ziti-related errors."""

    pass


class ZitiNotAvailableError(ZitiError):
    """Raised when OpenZiti SDK is not installed."""

    pass


class ZitiEnrollmentError(ZitiError):
    """Raised when identity enrollment fails."""

    pass


class ZitiConnectionError(ZitiError):
    """Raised when service connection fails."""

    pass


@dataclass
class ZitiConfig:
    """OpenZiti configuration.

    Attributes:
        controller: Ziti controller URL
        identity_dir: Directory for storing identities
        services: Map of service names to their Ziti service identifiers
    """

    controller: str = "https://ziti.rtmx.ai"
    identity_dir: Path = field(default_factory=lambda: Path.home() / ".rtmx" / "ziti")
    services: dict[str, str] = field(
        default_factory=lambda: {
            "rtmx-sync": "rtmx-sync-service",
        }
    )

    def __post_init__(self) -> None:
        """Ensure identity directory exists."""
        self.identity_dir.mkdir(parents=True, exist_ok=True)

    @property
    def identity_path(self) -> Path:
        """Get path to stored identity file."""
        return self.identity_dir / "identity.json"

    @property
    def has_identity(self) -> bool:
        """Check if an identity is enrolled."""
        return self.identity_path.exists()

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> ZitiConfig:
        """Create config from dictionary."""
        identity_dir_str = data.get("identity_dir")
        identity_dir = (
            Path(identity_dir_str) if identity_dir_str else Path.home() / ".rtmx" / "ziti"
        )

        return cls(
            controller=data.get("controller", "https://ziti.rtmx.ai"),
            identity_dir=identity_dir,
            services=data.get(
                "services",
                {
                    "rtmx-sync": "rtmx-sync-service",
                },
            ),
        )

    def to_dict(self) -> dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "controller": self.controller,
            "identity_dir": str(self.identity_dir),
            "services": self.services,
        }


@dataclass
class ZitiIdentity:
    """OpenZiti identity information.

    Represents an enrolled Ziti identity that can access services.

    Attributes:
        name: Identity name (usually matches user ID)
        fingerprint: Certificate fingerprint
        enrolled_at: Enrollment timestamp
        roles: Ziti roles assigned to this identity
    """

    name: str
    fingerprint: str = ""
    enrolled_at: str = ""
    roles: list[str] = field(default_factory=list)

    @classmethod
    def from_identity_file(cls, path: Path) -> ZitiIdentity:
        """Load identity info from Ziti identity file.

        Args:
            path: Path to identity.json file

        Returns:
            ZitiIdentity with info from file
        """
        if not path.exists():
            raise ZitiError(f"Identity file not found: {path}")

        data = json.loads(path.read_text())
        return cls(
            name=data.get("id", {}).get("name", "unknown"),
            fingerprint=data.get("id", {}).get("fingerprint", ""),
            enrolled_at=data.get("id", {}).get("enrolledAt", ""),
            roles=data.get("identity", {}).get("roles", []),
        )


class ZitiClient:
    """OpenZiti client for connecting to dark services.

    This client wraps the OpenZiti SDK to provide secure
    connections to rtmx-sync and other dark services.
    """

    def __init__(self, config: ZitiConfig | None = None) -> None:
        """Initialize Ziti client.

        Args:
            config: Ziti configuration. Uses defaults if not provided.

        Raises:
            ZitiNotAvailableError: If OpenZiti SDK is not installed
        """
        if not HAS_OPENZITI:
            raise ZitiNotAvailableError(
                "OpenZiti SDK not installed. Install with: pip install openziti"
            )

        self.config = config or ZitiConfig()
        self._context: Any = None
        self._identity: ZitiIdentity | None = None

    @property
    def is_enrolled(self) -> bool:
        """Check if identity is enrolled."""
        return self.config.has_identity

    @property
    def identity(self) -> ZitiIdentity | None:
        """Get current identity info."""
        if self._identity is None and self.is_enrolled:
            self._identity = ZitiIdentity.from_identity_file(self.config.identity_path)
        return self._identity

    async def enroll(self, jwt_token: str) -> ZitiIdentity:
        """Enroll a new identity using JWT enrollment token.

        The JWT token is typically derived from OIDC authentication.

        Args:
            jwt_token: One-time enrollment JWT from Ziti controller

        Returns:
            Enrolled ZitiIdentity

        Raises:
            ZitiEnrollmentError: If enrollment fails
        """
        if not HAS_OPENZITI:
            raise ZitiNotAvailableError("OpenZiti SDK not installed")

        try:
            # Write JWT to temp file for enrollment
            jwt_path = self.config.identity_dir / "enrollment.jwt"
            jwt_path.write_text(jwt_token)

            # Enroll using SDK
            openziti.enroll(
                jwt=str(jwt_path),
                identity=str(self.config.identity_path),
            )

            # Clean up JWT
            jwt_path.unlink()

            # Load and return identity
            self._identity = ZitiIdentity.from_identity_file(self.config.identity_path)
            return self._identity

        except Exception as e:
            raise ZitiEnrollmentError(f"Enrollment failed: {e}") from e

    def load_context(self) -> None:
        """Load Ziti context from stored identity.

        Must be called before connecting to services.

        Raises:
            ZitiError: If no identity enrolled or load fails
        """
        if not HAS_OPENZITI:
            raise ZitiNotAvailableError("OpenZiti SDK not installed")

        if not self.is_enrolled:
            raise ZitiError("No identity enrolled. Run enrollment first.")

        try:
            self._context = openziti.load(str(self.config.identity_path))
        except Exception as e:
            raise ZitiError(f"Failed to load Ziti context: {e}") from e

    async def connect(self, service_name: str) -> Any:
        """Connect to a Ziti service.

        Args:
            service_name: Name of service to connect to (e.g., "rtmx-sync")

        Returns:
            Connected socket-like object for communication

        Raises:
            ZitiConnectionError: If connection fails
        """
        if self._context is None:
            self.load_context()

        ziti_service = self.config.services.get(service_name, service_name)

        try:
            # Use Ziti SDK to dial service
            sock = openziti.socket(type=openziti.SOCK_STREAM)
            sock.connect((ziti_service, 0))  # Port 0 for Ziti service dial
            return sock
        except Exception as e:
            raise ZitiConnectionError(f"Failed to connect to {service_name}: {e}") from e

    def close(self) -> None:
        """Close Ziti context and release resources."""
        import contextlib

        if self._context is not None:
            with contextlib.suppress(Exception):
                # SDK cleanup if needed
                pass
            self._context = None


# Module-level convenience functions

_client: ZitiClient | None = None
_config: ZitiConfig | None = None


def is_ziti_available() -> bool:
    """Check if OpenZiti SDK is available.

    Returns:
        True if SDK is installed and importable
    """
    return HAS_OPENZITI


def get_config() -> ZitiConfig:
    """Get current Ziti configuration."""
    global _config
    if _config is None:
        _config = ZitiConfig()
    return _config


def set_config(config: ZitiConfig) -> None:
    """Set Ziti configuration."""
    global _config, _client
    _config = config
    _client = None  # Reset client on config change


def get_client() -> ZitiClient:
    """Get or create Ziti client instance."""
    global _client
    if _client is None:
        _client = ZitiClient(get_config())
    return _client


async def enroll_identity(jwt_token: str) -> ZitiIdentity:
    """Enroll a new Ziti identity.

    Args:
        jwt_token: One-time enrollment JWT

    Returns:
        Enrolled identity information
    """
    client = get_client()
    return await client.enroll(jwt_token)


async def connect_service(service_name: str) -> Any:
    """Connect to a Ziti service.

    Args:
        service_name: Service to connect to

    Returns:
        Connected socket for communication
    """
    client = get_client()
    return await client.connect(service_name)
