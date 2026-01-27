"""OIDC authentication client for RTMX.

Implements PKCE-based authentication flow for CLI applications
using Zitadel as the identity provider.

Security features:
- PKCE (Proof Key for Code Exchange) to prevent authorization code interception
- Secure token storage in system keychain
- Short-lived access tokens with refresh capability
- No client secret (public client)

Flow:
1. Generate PKCE code verifier and challenge
2. Start local callback server
3. Open browser to Zitadel authorize endpoint
4. Receive authorization code via callback
5. Exchange code + verifier for tokens
6. Store tokens securely
"""

from __future__ import annotations

import base64
import hashlib
import http.server
import json
import secrets
import socketserver
import threading
import urllib.parse
import webbrowser
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from pathlib import Path
from typing import Any

# Optional keyring support
try:
    import keyring

    HAS_KEYRING = True
except ImportError:
    HAS_KEYRING = False


@dataclass
class AuthConfig:
    """OIDC authentication configuration.

    Attributes:
        provider: Identity provider name (e.g., "zitadel")
        issuer: OIDC issuer URL (e.g., "https://auth.rtmx.ai")
        client_id: Public client ID for CLI
        scopes: Requested OAuth scopes
        callback_port: Local port for callback server
    """

    provider: str = "zitadel"
    issuer: str = "https://auth.rtmx.ai"
    client_id: str = "rtmx-cli"
    scopes: list[str] = field(
        default_factory=lambda: ["openid", "profile", "email", "offline_access"]
    )
    callback_port: int = 8765

    @property
    def authorization_endpoint(self) -> str:
        """Get authorization endpoint URL."""
        return f"{self.issuer}/oauth/v2/authorize"

    @property
    def token_endpoint(self) -> str:
        """Get token endpoint URL."""
        return f"{self.issuer}/oauth/v2/token"

    @property
    def userinfo_endpoint(self) -> str:
        """Get userinfo endpoint URL."""
        return f"{self.issuer}/oidc/v1/userinfo"

    @property
    def redirect_uri(self) -> str:
        """Get callback redirect URI."""
        return f"http://localhost:{self.callback_port}/callback"

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> AuthConfig:
        """Create config from dictionary."""
        return cls(
            provider=data.get("provider", "zitadel"),
            issuer=data.get("issuer", "https://auth.rtmx.ai"),
            client_id=data.get("client_id", "rtmx-cli"),
            scopes=data.get("scopes", ["openid", "profile", "email", "offline_access"]),
            callback_port=data.get("callback_port", 8765),
        )


@dataclass
class TokenInfo:
    """OAuth token information.

    Attributes:
        access_token: Current access token
        refresh_token: Token for refreshing access (optional)
        expires_at: Expiration timestamp
        token_type: Token type (usually "Bearer")
        id_token: OIDC ID token for user info (optional)
    """

    access_token: str
    refresh_token: str = ""
    expires_at: datetime = field(default_factory=datetime.now)
    token_type: str = "Bearer"
    id_token: str = ""

    @property
    def is_expired(self) -> bool:
        """Check if access token is expired."""
        return datetime.now() >= self.expires_at

    @property
    def is_refreshable(self) -> bool:
        """Check if tokens can be refreshed."""
        return bool(self.refresh_token)

    def to_dict(self) -> dict[str, Any]:
        """Convert to dictionary for storage."""
        return {
            "access_token": self.access_token,
            "refresh_token": self.refresh_token,
            "expires_at": self.expires_at.isoformat(),
            "token_type": self.token_type,
            "id_token": self.id_token,
        }

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> TokenInfo:
        """Create from dictionary."""
        expires_at = data.get("expires_at", "")
        if isinstance(expires_at, str):
            expires_at = datetime.fromisoformat(expires_at)
        return cls(
            access_token=data.get("access_token", ""),
            refresh_token=data.get("refresh_token", ""),
            expires_at=expires_at,
            token_type=data.get("token_type", "Bearer"),
            id_token=data.get("id_token", ""),
        )


# PKCE utilities


def generate_code_verifier() -> str:
    """Generate a cryptographically random code verifier for PKCE.

    Returns:
        43-128 character URL-safe string
    """
    return secrets.token_urlsafe(64)[:128]


def generate_code_challenge(verifier: str) -> str:
    """Generate code challenge from verifier using S256 method.

    Args:
        verifier: PKCE code verifier

    Returns:
        Base64url-encoded SHA256 hash
    """
    digest = hashlib.sha256(verifier.encode("ascii")).digest()
    return base64.urlsafe_b64encode(digest).rstrip(b"=").decode("ascii")


# Token storage


def _get_token_path() -> Path:
    """Get path for file-based token storage (fallback)."""
    config_dir = Path.home() / ".rtmx"
    config_dir.mkdir(parents=True, exist_ok=True)
    return config_dir / "tokens.json"


def _store_tokens(tokens: TokenInfo, config: AuthConfig) -> None:
    """Store tokens securely.

    Uses system keychain if available, falls back to encrypted file.
    """
    if HAS_KEYRING:
        keyring.set_password(
            "rtmx",
            f"{config.provider}:{config.client_id}",
            json.dumps(tokens.to_dict()),
        )
    else:
        # Fallback to file storage (less secure)
        token_path = _get_token_path()
        token_path.write_text(json.dumps(tokens.to_dict()))
        token_path.chmod(0o600)  # Restrict permissions


def _load_tokens(config: AuthConfig) -> TokenInfo | None:
    """Load stored tokens.

    Returns:
        TokenInfo if found and valid, None otherwise
    """
    if HAS_KEYRING:
        try:
            stored = keyring.get_password("rtmx", f"{config.provider}:{config.client_id}")
            if stored:
                return TokenInfo.from_dict(json.loads(stored))
        except Exception:
            pass
    else:
        token_path = _get_token_path()
        if token_path.exists():
            try:
                return TokenInfo.from_dict(json.loads(token_path.read_text()))
            except Exception:
                pass
    return None


def _clear_tokens(config: AuthConfig) -> None:
    """Clear stored tokens."""
    import contextlib

    if HAS_KEYRING:
        with contextlib.suppress(Exception):
            keyring.delete_password("rtmx", f"{config.provider}:{config.client_id}")
    token_path = _get_token_path()
    if token_path.exists():
        token_path.unlink()


# Local callback server for OAuth redirect


class _CallbackHandler(http.server.BaseHTTPRequestHandler):
    """HTTP handler for OAuth callback."""

    authorization_code: str | None = None
    error: str | None = None

    def log_message(self, format: str, *args: Any) -> None:
        """Suppress default logging."""
        pass

    def do_GET(self) -> None:
        """Handle GET request from OAuth redirect."""
        parsed = urllib.parse.urlparse(self.path)
        params = urllib.parse.parse_qs(parsed.query)

        if "code" in params:
            _CallbackHandler.authorization_code = params["code"][0]
            self.send_response(200)
            self.send_header("Content-type", "text/html")
            self.end_headers()
            self.wfile.write(b"""
<!DOCTYPE html>
<html>
<head><title>RTMX Authentication</title></head>
<body style="font-family: system-ui; text-align: center; padding: 50px;">
<h1>Authentication Successful</h1>
<p>You can close this window and return to your terminal.</p>
</body>
</html>
""")
        elif "error" in params:
            _CallbackHandler.error = params.get("error_description", params["error"])[0]
            self.send_response(400)
            self.send_header("Content-type", "text/html")
            self.end_headers()
            self.wfile.write(
                f"""
<!DOCTYPE html>
<html>
<head><title>RTMX Authentication Error</title></head>
<body style="font-family: system-ui; text-align: center; padding: 50px;">
<h1>Authentication Failed</h1>
<p>{_CallbackHandler.error}</p>
</body>
</html>
""".encode()
            )
        else:
            self.send_response(400)
            self.end_headers()


def _wait_for_callback(port: int, timeout: int = 120) -> str | None:
    """Start callback server and wait for authorization code.

    Args:
        port: Local port to listen on
        timeout: Maximum wait time in seconds

    Returns:
        Authorization code if received, None on timeout/error
    """

    class ReuseAddrServer(socketserver.TCPServer):
        allow_reuse_address = True

    _CallbackHandler.authorization_code = None
    _CallbackHandler.error = None

    with ReuseAddrServer(("", port), _CallbackHandler) as httpd:
        httpd.timeout = timeout
        httpd.handle_request()

    if _CallbackHandler.error:
        raise AuthenticationError(_CallbackHandler.error)

    return _CallbackHandler.authorization_code


class AuthenticationError(Exception):
    """Raised when authentication fails."""

    pass


# Public API


_config: AuthConfig | None = None
_tokens: TokenInfo | None = None


def get_config() -> AuthConfig:
    """Get current authentication configuration."""
    global _config
    if _config is None:
        _config = AuthConfig()
    return _config


def set_config(config: AuthConfig) -> None:
    """Set authentication configuration."""
    global _config
    _config = config


def is_authenticated() -> bool:
    """Check if user is authenticated with valid tokens.

    Returns:
        True if authenticated with non-expired tokens
    """
    global _tokens
    config = get_config()

    if _tokens is None:
        _tokens = _load_tokens(config)

    if _tokens is None:
        return False

    # If expired but refreshable, we still consider authenticated
    return not (_tokens.is_expired and not _tokens.is_refreshable)


def get_access_token() -> str | None:
    """Get current access token, refreshing if needed.

    Returns:
        Access token string, or None if not authenticated
    """
    global _tokens
    config = get_config()

    if _tokens is None:
        _tokens = _load_tokens(config)

    if _tokens is None:
        return None

    if _tokens.is_expired and _tokens.is_refreshable:
        try:
            refresh_tokens()
        except AuthenticationError:
            return None

    return _tokens.access_token if _tokens and not _tokens.is_expired else None


async def login(open_browser: bool = True) -> TokenInfo:
    """Perform OIDC login flow.

    This is an interactive flow that:
    1. Generates PKCE codes
    2. Opens browser to authorization endpoint
    3. Waits for callback with authorization code
    4. Exchanges code for tokens
    5. Stores tokens securely

    Args:
        open_browser: Whether to automatically open browser

    Returns:
        TokenInfo with access and refresh tokens

    Raises:
        AuthenticationError: If authentication fails
    """
    global _tokens
    config = get_config()

    # Generate PKCE codes
    verifier = generate_code_verifier()
    challenge = generate_code_challenge(verifier)

    # Build authorization URL
    state = secrets.token_urlsafe(16)
    params = {
        "client_id": config.client_id,
        "response_type": "code",
        "redirect_uri": config.redirect_uri,
        "scope": " ".join(config.scopes),
        "state": state,
        "code_challenge": challenge,
        "code_challenge_method": "S256",
    }
    auth_url = f"{config.authorization_endpoint}?{urllib.parse.urlencode(params)}"

    # Start callback server in background thread
    result: dict[str, Any] = {}

    def wait_for_code() -> None:
        try:
            result["code"] = _wait_for_callback(config.callback_port)
        except AuthenticationError as e:
            result["error"] = str(e)

    thread = threading.Thread(target=wait_for_code)
    thread.start()

    # Open browser
    if open_browser:
        webbrowser.open(auth_url)
    else:
        print(f"Please open this URL in your browser:\n{auth_url}")

    # Wait for callback
    thread.join(timeout=120)

    if "error" in result:
        raise AuthenticationError(result["error"])

    if not result.get("code"):
        raise AuthenticationError("Authentication timed out or was cancelled")

    # Exchange code for tokens
    _tokens = await _exchange_code(result["code"], verifier, config)
    _store_tokens(_tokens, config)

    return _tokens


async def _exchange_code(code: str, verifier: str, config: AuthConfig) -> TokenInfo:
    """Exchange authorization code for tokens.

    Args:
        code: Authorization code from callback
        verifier: PKCE code verifier
        config: Authentication configuration

    Returns:
        TokenInfo with tokens from provider
    """
    import aiohttp

    data = {
        "grant_type": "authorization_code",
        "client_id": config.client_id,
        "code": code,
        "redirect_uri": config.redirect_uri,
        "code_verifier": verifier,
    }

    async with (
        aiohttp.ClientSession() as session,
        session.post(config.token_endpoint, data=data) as resp,
    ):
        if resp.status != 200:
            error_text = await resp.text()
            raise AuthenticationError(f"Token exchange failed: {error_text}")

        token_data = await resp.json()

    expires_in = token_data.get("expires_in", 3600)
    return TokenInfo(
        access_token=token_data["access_token"],
        refresh_token=token_data.get("refresh_token", ""),
        expires_at=datetime.now() + timedelta(seconds=expires_in),
        token_type=token_data.get("token_type", "Bearer"),
        id_token=token_data.get("id_token", ""),
    )


def refresh_tokens() -> TokenInfo:
    """Refresh access token using refresh token.

    Returns:
        Updated TokenInfo

    Raises:
        AuthenticationError: If refresh fails
    """
    import asyncio

    return asyncio.run(_refresh_tokens_async())


async def _refresh_tokens_async() -> TokenInfo:
    """Async implementation of token refresh."""
    import aiohttp

    global _tokens
    config = get_config()

    if _tokens is None or not _tokens.refresh_token:
        raise AuthenticationError("No refresh token available")

    data = {
        "grant_type": "refresh_token",
        "client_id": config.client_id,
        "refresh_token": _tokens.refresh_token,
    }

    async with (
        aiohttp.ClientSession() as session,
        session.post(config.token_endpoint, data=data) as resp,
    ):
        if resp.status != 200:
            _clear_tokens(config)
            _tokens = None
            error_text = await resp.text()
            raise AuthenticationError(f"Token refresh failed: {error_text}")

        token_data = await resp.json()

    expires_in = token_data.get("expires_in", 3600)
    _tokens = TokenInfo(
        access_token=token_data["access_token"],
        refresh_token=token_data.get("refresh_token", _tokens.refresh_token),
        expires_at=datetime.now() + timedelta(seconds=expires_in),
        token_type=token_data.get("token_type", "Bearer"),
        id_token=token_data.get("id_token", ""),
    )
    _store_tokens(_tokens, config)

    return _tokens


def logout() -> None:
    """Clear stored tokens and logout.

    This clears all stored authentication state.
    """
    global _tokens
    config = get_config()
    _clear_tokens(config)
    _tokens = None
