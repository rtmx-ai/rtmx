"""RTMX Authentication Module.

Provides zero-trust authentication for RTMX CLI using Zitadel OIDC.

This module implements:
- PKCE-based OIDC login flow for CLI
- Token storage and refresh
- JWT validation utilities

Example:
    >>> from rtmx.auth import login, get_access_token, logout
    >>> await login()  # Opens browser for Zitadel auth
    >>> token = get_access_token()  # Returns cached token
    >>> logout()  # Clears stored tokens
"""

from rtmx.auth.oidc import (
    AuthConfig,
    TokenInfo,
    get_access_token,
    get_config,
    is_authenticated,
    login,
    logout,
    refresh_tokens,
)

__all__ = [
    "AuthConfig",
    "TokenInfo",
    "get_access_token",
    "get_config",
    "is_authenticated",
    "login",
    "logout",
    "refresh_tokens",
]
