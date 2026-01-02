"""RTMX configuration management.

Loads and validates rtmx.yaml configuration with sensible defaults.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

import yaml

# Directory and file name constants
RTMX_DIR_NAME = ".rtmx"
CONFIG_FILE_NAME = "config.yaml"
LEGACY_CONFIG_NAMES = ("rtmx.yaml", "rtmx.yml")


@dataclass
class AgentConfig:
    """Configuration for a single AI agent integration."""

    enabled: bool = True
    config_path: str = ""


@dataclass
class AgentsConfig:
    """Configuration for all AI agent integrations."""

    claude: AgentConfig = field(default_factory=lambda: AgentConfig(config_path="CLAUDE.md"))
    cursor: AgentConfig = field(default_factory=lambda: AgentConfig(config_path=".cursorrules"))
    copilot: AgentConfig = field(
        default_factory=lambda: AgentConfig(config_path=".github/copilot-instructions.md")
    )
    template_dir: str = ".rtmx/templates/"

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> AgentsConfig:
        """Create AgentsConfig from dictionary."""
        config = cls()
        if "claude" in data:
            config.claude = AgentConfig(**data["claude"])
        if "cursor" in data:
            config.cursor = AgentConfig(**data["cursor"])
        if "copilot" in data:
            config.copilot = AgentConfig(**data["copilot"])
        if "template_dir" in data:
            config.template_dir = data["template_dir"]
        return config


@dataclass
class GitHubAdapterConfig:
    """Configuration for GitHub Issues adapter."""

    enabled: bool = False
    repo: str = ""
    token_env: str = "GITHUB_TOKEN"
    labels: list[str] = field(default_factory=lambda: ["requirement", "feature"])
    status_mapping: dict[str, str] = field(
        default_factory=lambda: {"open": "MISSING", "closed": "COMPLETE"}
    )

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> GitHubAdapterConfig:
        """Create GitHubAdapterConfig from dictionary."""
        return cls(
            enabled=data.get("enabled", False),
            repo=data.get("repo", ""),
            token_env=data.get("token_env", "GITHUB_TOKEN"),
            labels=data.get("labels", ["requirement", "feature"]),
            status_mapping=data.get("status_mapping", {"open": "MISSING", "closed": "COMPLETE"}),
        )


@dataclass
class JiraAdapterConfig:
    """Configuration for Jira adapter."""

    enabled: bool = False
    server: str = ""
    project: str = ""
    token_env: str = "JIRA_API_TOKEN"
    email_env: str = "JIRA_EMAIL"
    issue_type: str = "Requirement"
    jql_filter: str = ""
    labels: list[str] = field(default_factory=list)
    status_mapping: dict[str, str] = field(
        default_factory=lambda: {"To Do": "MISSING", "In Progress": "PARTIAL", "Done": "COMPLETE"}
    )

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> JiraAdapterConfig:
        """Create JiraAdapterConfig from dictionary."""
        return cls(
            enabled=data.get("enabled", False),
            server=data.get("server", ""),
            project=data.get("project", ""),
            token_env=data.get("token_env", "JIRA_API_TOKEN"),
            email_env=data.get("email_env", "JIRA_EMAIL"),
            issue_type=data.get("issue_type", "Requirement"),
            jql_filter=data.get("jql_filter", ""),
            labels=data.get("labels", []),
            status_mapping=data.get(
                "status_mapping",
                {"To Do": "MISSING", "In Progress": "PARTIAL", "Done": "COMPLETE"},
            ),
        )


@dataclass
class AdaptersConfig:
    """Configuration for external service adapters."""

    github: GitHubAdapterConfig = field(default_factory=GitHubAdapterConfig)
    jira: JiraAdapterConfig = field(default_factory=JiraAdapterConfig)

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> AdaptersConfig:
        """Create AdaptersConfig from dictionary."""
        return cls(
            github=GitHubAdapterConfig.from_dict(data.get("github", {})),
            jira=JiraAdapterConfig.from_dict(data.get("jira", {})),
        )


@dataclass
class MCPConfig:
    """Configuration for MCP server."""

    enabled: bool = False
    port: int = 3000
    host: str = "localhost"

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> MCPConfig:
        """Create MCPConfig from dictionary."""
        return cls(
            enabled=data.get("enabled", False),
            port=data.get("port", 3000),
            host=data.get("host", "localhost"),
        )


@dataclass
class SyncConfig:
    """Configuration for synchronization behavior."""

    conflict_resolution: str = "manual"  # manual, prefer-local, prefer-remote

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> SyncConfig:
        """Create SyncConfig from dictionary."""
        return cls(
            conflict_resolution=data.get("conflict_resolution", "manual"),
        )


@dataclass
class PytestConfig:
    """Configuration for pytest integration."""

    marker_prefix: str = "req"
    register_markers: bool = True

    @classmethod
    def from_dict(cls, data: dict[str, Any]) -> PytestConfig:
        """Create PytestConfig from dictionary."""
        return cls(
            marker_prefix=data.get("marker_prefix", "req"),
            register_markers=data.get("register_markers", True),
        )


@dataclass
class RTMXConfig:
    """Complete RTMX configuration."""

    database: Path = field(default_factory=lambda: Path("docs/rtm_database.csv"))
    requirements_dir: Path = field(default_factory=lambda: Path("docs/requirements"))
    schema: str = "core"
    pytest: PytestConfig = field(default_factory=PytestConfig)
    agents: AgentsConfig = field(default_factory=AgentsConfig)
    adapters: AdaptersConfig = field(default_factory=AdaptersConfig)
    mcp: MCPConfig = field(default_factory=MCPConfig)
    sync: SyncConfig = field(default_factory=SyncConfig)

    # Path where config was loaded from (if any)
    _config_path: Path | None = None

    @classmethod
    def from_dict(cls, data: dict[str, Any], config_path: Path | None = None) -> RTMXConfig:
        """Create RTMXConfig from dictionary.

        Args:
            data: Configuration dictionary (usually from YAML)
            config_path: Path where config was loaded from

        Returns:
            RTMXConfig instance
        """
        rtmx_data = data.get("rtmx", data)

        config = cls(
            database=Path(rtmx_data.get("database", "docs/rtm_database.csv")),
            requirements_dir=Path(rtmx_data.get("requirements_dir", "docs/requirements")),
            schema=rtmx_data.get("schema", "core"),
        )

        if "pytest" in rtmx_data:
            config.pytest = PytestConfig.from_dict(rtmx_data["pytest"])
        if "agents" in rtmx_data:
            config.agents = AgentsConfig.from_dict(rtmx_data["agents"])
        if "adapters" in rtmx_data:
            config.adapters = AdaptersConfig.from_dict(rtmx_data["adapters"])
        if "mcp" in rtmx_data:
            config.mcp = MCPConfig.from_dict(rtmx_data["mcp"])
        if "sync" in rtmx_data:
            config.sync = SyncConfig.from_dict(rtmx_data["sync"])

        config._config_path = config_path
        return config

    def to_dict(self) -> dict[str, Any]:
        """Convert config to dictionary for serialization."""
        return {
            "rtmx": {
                "database": str(self.database),
                "requirements_dir": str(self.requirements_dir),
                "schema": self.schema,
                "pytest": {
                    "marker_prefix": self.pytest.marker_prefix,
                    "register_markers": self.pytest.register_markers,
                },
                "agents": {
                    "claude": {
                        "enabled": self.agents.claude.enabled,
                        "config_path": self.agents.claude.config_path,
                    },
                    "cursor": {
                        "enabled": self.agents.cursor.enabled,
                        "config_path": self.agents.cursor.config_path,
                    },
                    "copilot": {
                        "enabled": self.agents.copilot.enabled,
                        "config_path": self.agents.copilot.config_path,
                    },
                    "template_dir": self.agents.template_dir,
                },
                "adapters": {
                    "github": {
                        "enabled": self.adapters.github.enabled,
                        "repo": self.adapters.github.repo,
                        "token_env": self.adapters.github.token_env,
                        "labels": self.adapters.github.labels,
                        "status_mapping": self.adapters.github.status_mapping,
                    },
                    "jira": {
                        "enabled": self.adapters.jira.enabled,
                        "server": self.adapters.jira.server,
                        "project": self.adapters.jira.project,
                        "token_env": self.adapters.jira.token_env,
                        "issue_type": self.adapters.jira.issue_type,
                        "status_mapping": self.adapters.jira.status_mapping,
                    },
                },
                "mcp": {
                    "enabled": self.mcp.enabled,
                    "port": self.mcp.port,
                    "host": self.mcp.host,
                },
                "sync": {
                    "conflict_resolution": self.sync.conflict_resolution,
                },
            }
        }


def find_config_file(start_path: Path | None = None) -> Path | None:
    """Find RTMX config by searching upward from start_path.

    Checks for config files in this order:
    1. .rtmx/config.yaml (new standard)
    2. rtmx.yaml (legacy)
    3. rtmx.yml (legacy)

    Args:
        start_path: Starting directory (defaults to cwd)

    Returns:
        Path to config file if found, None otherwise
    """
    current = Path.cwd() if start_path is None else Path(start_path).resolve()

    while current != current.parent:
        # Check .rtmx/config.yaml first (new standard)
        rtmx_dir_config = current / RTMX_DIR_NAME / CONFIG_FILE_NAME
        if rtmx_dir_config.exists():
            return rtmx_dir_config

        # Fall back to legacy config names
        for config_name in LEGACY_CONFIG_NAMES:
            config_path = current / config_name
            if config_path.exists():
                return config_path

        current = current.parent

    return None


def load_config(path: Path | str | None = None) -> RTMXConfig:
    """Load RTMX configuration from file.

    If path is not specified, searches upward from cwd for rtmx.yaml.
    If no config file is found, returns default configuration.

    Args:
        path: Path to config file (optional)

    Returns:
        RTMXConfig instance
    """
    config_path: Path | None = None

    if path is not None:
        config_path = Path(path)
        if not config_path.exists():
            # Return defaults if specified path doesn't exist
            return RTMXConfig()
    else:
        config_path = find_config_file()

    if config_path is None:
        # No config file found, return defaults
        return RTMXConfig()

    with config_path.open() as f:
        data = yaml.safe_load(f) or {}

    return RTMXConfig.from_dict(data, config_path)


def save_config(config: RTMXConfig, path: Path | str | None = None) -> Path:
    """Save RTMX configuration to file.

    Args:
        config: Configuration to save
        path: Path to save to (defaults to rtmx.yaml in cwd)

    Returns:
        Path where config was saved
    """
    save_path = Path.cwd() / "rtmx.yaml" if path is None else Path(path)

    with save_path.open("w") as f:
        yaml.dump(config.to_dict(), f, default_flow_style=False, sort_keys=False)

    return save_path
