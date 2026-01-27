"""RTMX OpenZiti Integration Module.

Provides zero-trust network overlay for RTMX using OpenZiti.

This module enables:
- Dark services (no public ports)
- End-to-end encryption
- Identity-based access control
- Ziti identity enrollment from OIDC tokens

The rtmx-sync server uses this module to become a dark service,
invisible to port scanners and accessible only via Ziti overlay.

Example:
    >>> from rtmx.ziti import ZitiConfig, enroll_identity, connect_service
    >>> config = ZitiConfig.load()
    >>> identity = await enroll_identity(jwt_token)
    >>> client = await connect_service("rtmx-sync")
"""

from rtmx.ziti.client import (
    ZitiClient,
    ZitiConfig,
    ZitiConnectionError,
    ZitiEnrollmentError,
    ZitiError,
    ZitiIdentity,
    ZitiNotAvailableError,
    connect_service,
    enroll_identity,
    is_ziti_available,
)

__all__ = [
    "ZitiClient",
    "ZitiConfig",
    "ZitiConnectionError",
    "ZitiEnrollmentError",
    "ZitiError",
    "ZitiIdentity",
    "ZitiNotAvailableError",
    "connect_service",
    "enroll_identity",
    "is_ziti_available",
]
