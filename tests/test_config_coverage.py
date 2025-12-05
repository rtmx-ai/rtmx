"""Comprehensive tests for rtmx.config module."""

from pathlib import Path
from typing import Any

import pytest
import yaml

from rtmx.config import (
    AdaptersConfig,
    AgentConfig,
    AgentsConfig,
    GitHubAdapterConfig,
    JiraAdapterConfig,
    MCPConfig,
    PytestConfig,
    RTMXConfig,
    SyncConfig,
    find_config_file,
    load_config,
    save_config,
)


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_agent_config_defaults():
    """Test AgentConfig has sensible defaults."""
    config = AgentConfig()
    assert config.enabled is True
    assert config.config_path == ""


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_agent_config_custom():
    """Test AgentConfig with custom values."""
    config = AgentConfig(enabled=False, config_path="custom/path.md")
    assert config.enabled is False
    assert config.config_path == "custom/path.md"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_agents_config_defaults():
    """Test AgentsConfig has proper defaults for all agents."""
    config = AgentsConfig()

    # Check Claude defaults
    assert config.claude.enabled is True
    assert config.claude.config_path == "CLAUDE.md"

    # Check Cursor defaults
    assert config.cursor.enabled is True
    assert config.cursor.config_path == ".cursorrules"

    # Check Copilot defaults
    assert config.copilot.enabled is True
    assert config.copilot.config_path == ".github/copilot-instructions.md"

    # Check template directory
    assert config.template_dir == ".rtmx/templates/"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_agents_config_from_dict_empty():
    """Test AgentsConfig.from_dict with empty dict uses defaults."""
    config = AgentsConfig.from_dict({})

    assert config.claude.enabled is True
    assert config.cursor.enabled is True
    assert config.copilot.enabled is True
    assert config.template_dir == ".rtmx/templates/"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_agents_config_from_dict_partial():
    """Test AgentsConfig.from_dict with partial configuration."""
    data = {
        "claude": {"enabled": False, "config_path": "custom.md"},
        "template_dir": "my_templates/",
    }
    config = AgentsConfig.from_dict(data)

    assert config.claude.enabled is False
    assert config.claude.config_path == "custom.md"
    assert config.cursor.enabled is True  # Should use default
    assert config.template_dir == "my_templates/"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_agents_config_from_dict_full():
    """Test AgentsConfig.from_dict with complete configuration."""
    data = {
        "claude": {"enabled": True, "config_path": "CLAUDE.md"},
        "cursor": {"enabled": False, "config_path": ".cursor"},
        "copilot": {"enabled": False, "config_path": ".copilot.md"},
        "template_dir": "templates/",
    }
    config = AgentsConfig.from_dict(data)

    assert config.claude.enabled is True
    assert config.cursor.enabled is False
    assert config.copilot.enabled is False
    assert config.template_dir == "templates/"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_github_adapter_config_defaults():
    """Test GitHubAdapterConfig defaults."""
    config = GitHubAdapterConfig()

    assert config.enabled is False
    assert config.repo == ""
    assert config.token_env == "GITHUB_TOKEN"
    assert config.labels == ["requirement", "feature"]
    assert config.status_mapping == {"open": "MISSING", "closed": "COMPLETE"}


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_github_adapter_config_from_dict():
    """Test GitHubAdapterConfig.from_dict with custom values."""
    data = {
        "enabled": True,
        "repo": "owner/repo",
        "token_env": "GH_TOKEN",
        "labels": ["bug", "enhancement"],
        "status_mapping": {"open": "PARTIAL", "closed": "DONE"},
    }
    config = GitHubAdapterConfig.from_dict(data)

    assert config.enabled is True
    assert config.repo == "owner/repo"
    assert config.token_env == "GH_TOKEN"
    assert config.labels == ["bug", "enhancement"]
    assert config.status_mapping == {"open": "PARTIAL", "closed": "DONE"}


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_jira_adapter_config_defaults():
    """Test JiraAdapterConfig defaults."""
    config = JiraAdapterConfig()

    assert config.enabled is False
    assert config.server == ""
    assert config.project == ""
    assert config.token_env == "JIRA_API_TOKEN"
    assert config.email_env == "JIRA_EMAIL"
    assert config.issue_type == "Requirement"
    assert config.jql_filter == ""
    assert config.labels == []
    assert config.status_mapping == {
        "To Do": "MISSING",
        "In Progress": "PARTIAL",
        "Done": "COMPLETE",
    }


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_jira_adapter_config_from_dict():
    """Test JiraAdapterConfig.from_dict with custom values."""
    data = {
        "enabled": True,
        "server": "https://jira.example.com",
        "project": "PROJ",
        "token_env": "MY_JIRA_TOKEN",
        "email_env": "MY_EMAIL",
        "issue_type": "Story",
        "jql_filter": "project = PROJ",
        "labels": ["requirement", "test"],
        "status_mapping": {"Todo": "MISSING", "Done": "COMPLETE"},
    }
    config = JiraAdapterConfig.from_dict(data)

    assert config.enabled is True
    assert config.server == "https://jira.example.com"
    assert config.project == "PROJ"
    assert config.token_env == "MY_JIRA_TOKEN"
    assert config.email_env == "MY_EMAIL"
    assert config.issue_type == "Story"
    assert config.jql_filter == "project = PROJ"
    assert config.labels == ["requirement", "test"]
    assert config.status_mapping == {"Todo": "MISSING", "Done": "COMPLETE"}


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_adapters_config_defaults():
    """Test AdaptersConfig has proper defaults."""
    config = AdaptersConfig()

    assert config.github.enabled is False
    assert config.jira.enabled is False


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_adapters_config_from_dict():
    """Test AdaptersConfig.from_dict with nested configs."""
    data = {
        "github": {"enabled": True, "repo": "owner/repo"},
        "jira": {"enabled": True, "server": "https://jira.example.com"},
    }
    config = AdaptersConfig.from_dict(data)

    assert config.github.enabled is True
    assert config.github.repo == "owner/repo"
    assert config.jira.enabled is True
    assert config.jira.server == "https://jira.example.com"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_mcp_config_defaults():
    """Test MCPConfig defaults."""
    config = MCPConfig()

    assert config.enabled is False
    assert config.port == 3000
    assert config.host == "localhost"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_mcp_config_from_dict():
    """Test MCPConfig.from_dict with custom values."""
    data = {"enabled": True, "port": 8080, "host": "0.0.0.0"}
    config = MCPConfig.from_dict(data)

    assert config.enabled is True
    assert config.port == 8080
    assert config.host == "0.0.0.0"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_sync_config_defaults():
    """Test SyncConfig defaults."""
    config = SyncConfig()

    assert config.conflict_resolution == "manual"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_sync_config_from_dict():
    """Test SyncConfig.from_dict with custom values."""
    data = {"conflict_resolution": "prefer-local"}
    config = SyncConfig.from_dict(data)

    assert config.conflict_resolution == "prefer-local"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_pytest_config_defaults():
    """Test PytestConfig defaults."""
    config = PytestConfig()

    assert config.marker_prefix == "req"
    assert config.register_markers is True


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_pytest_config_from_dict():
    """Test PytestConfig.from_dict with custom values."""
    data = {"marker_prefix": "requirement", "register_markers": False}
    config = PytestConfig.from_dict(data)

    assert config.marker_prefix == "requirement"
    assert config.register_markers is False


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_rtmx_config_defaults():
    """Test RTMXConfig has proper defaults."""
    config = RTMXConfig()

    assert config.database == Path("docs/rtm_database.csv")
    assert config.requirements_dir == Path("docs/requirements")
    assert config.schema == "core"
    assert isinstance(config.pytest, PytestConfig)
    assert isinstance(config.agents, AgentsConfig)
    assert isinstance(config.adapters, AdaptersConfig)
    assert isinstance(config.mcp, MCPConfig)
    assert isinstance(config.sync, SyncConfig)
    assert config._config_path is None


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_rtmx_config_from_dict_minimal():
    """Test RTMXConfig.from_dict with minimal data."""
    data: dict[str, Any] = {}
    config = RTMXConfig.from_dict(data)

    assert config.database == Path("docs/rtm_database.csv")
    assert config.schema == "core"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_rtmx_config_from_dict_with_rtmx_key():
    """Test RTMXConfig.from_dict with 'rtmx' wrapper key."""
    data = {
        "rtmx": {
            "database": "custom.csv",
            "schema": "phoenix",
        }
    }
    config = RTMXConfig.from_dict(data)

    assert config.database == Path("custom.csv")
    assert config.schema == "phoenix"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_rtmx_config_from_dict_full():
    """Test RTMXConfig.from_dict with complete configuration."""
    data = {
        "rtmx": {
            "database": "my_rtm.csv",
            "requirements_dir": "specs/",
            "schema": "phoenix",
            "pytest": {"marker_prefix": "test", "register_markers": False},
            "agents": {"claude": {"enabled": False}},
            "adapters": {"github": {"enabled": True, "repo": "owner/repo"}},
            "mcp": {"enabled": True, "port": 9000},
            "sync": {"conflict_resolution": "prefer-remote"},
        }
    }
    config = RTMXConfig.from_dict(data)

    assert config.database == Path("my_rtm.csv")
    assert config.requirements_dir == Path("specs/")
    assert config.schema == "phoenix"
    assert config.pytest.marker_prefix == "test"
    assert config.agents.claude.enabled is False
    assert config.adapters.github.enabled is True
    assert config.mcp.enabled is True
    assert config.sync.conflict_resolution == "prefer-remote"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_rtmx_config_from_dict_stores_config_path():
    """Test RTMXConfig.from_dict stores config path."""
    data: dict[str, Any] = {}
    config_path = Path("/path/to/rtmx.yaml")
    config = RTMXConfig.from_dict(data, config_path)

    assert config._config_path == config_path


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_rtmx_config_to_dict():
    """Test RTMXConfig.to_dict serializes configuration."""
    config = RTMXConfig(
        database=Path("test.csv"),
        requirements_dir=Path("reqs/"),
        schema="custom",
    )

    result = config.to_dict()

    assert "rtmx" in result
    assert result["rtmx"]["database"] == "test.csv"
    # Path strips trailing slashes when serialized
    assert result["rtmx"]["requirements_dir"] == "reqs"
    assert result["rtmx"]["schema"] == "custom"
    assert "pytest" in result["rtmx"]
    assert "agents" in result["rtmx"]
    assert "adapters" in result["rtmx"]
    assert "mcp" in result["rtmx"]
    assert "sync" in result["rtmx"]


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_find_config_file_in_current_dir(tmp_path):
    """Test find_config_file finds rtmx.yaml in current directory."""
    config_file = tmp_path / "rtmx.yaml"
    config_file.write_text("rtmx:\n  schema: core\n")

    result = find_config_file(tmp_path)

    assert result == config_file


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_find_config_file_yml_extension(tmp_path):
    """Test find_config_file finds rtmx.yml as alternative."""
    config_file = tmp_path / "rtmx.yml"
    config_file.write_text("rtmx:\n  schema: core\n")

    result = find_config_file(tmp_path)

    assert result == config_file


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_find_config_file_prefers_yaml_over_yml(tmp_path):
    """Test find_config_file prefers .yaml over .yml."""
    yaml_file = tmp_path / "rtmx.yaml"
    yml_file = tmp_path / "rtmx.yml"
    yaml_file.write_text("rtmx:\n  schema: yaml\n")
    yml_file.write_text("rtmx:\n  schema: yml\n")

    result = find_config_file(tmp_path)

    assert result == yaml_file


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_find_config_file_searches_upward(tmp_path):
    """Test find_config_file searches parent directories."""
    config_file = tmp_path / "rtmx.yaml"
    config_file.write_text("rtmx:\n  schema: core\n")

    subdir = tmp_path / "subdir" / "nested"
    subdir.mkdir(parents=True)

    result = find_config_file(subdir)

    assert result == config_file


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_find_config_file_returns_none_when_not_found(tmp_path):
    """Test find_config_file returns None when file doesn't exist."""
    result = find_config_file(tmp_path)

    assert result is None


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_load_config_with_explicit_path(tmp_path):
    """Test load_config loads from specified path."""
    config_file = tmp_path / "custom.yaml"
    config_data = {"rtmx": {"database": "custom.csv", "schema": "phoenix"}}
    config_file.write_text(yaml.dump(config_data))

    config = load_config(config_file)

    assert config.database == Path("custom.csv")
    assert config.schema == "phoenix"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_load_config_returns_defaults_when_path_not_exists(tmp_path):
    """Test load_config returns defaults when specified path doesn't exist."""
    nonexistent = tmp_path / "nonexistent.yaml"

    config = load_config(nonexistent)

    assert config.database == Path("docs/rtm_database.csv")
    assert config.schema == "core"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_load_config_returns_defaults_when_no_file_found():
    """Test load_config returns defaults when no config file found."""
    config = load_config()

    assert isinstance(config, RTMXConfig)
    assert config.database == Path("docs/rtm_database.csv")


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_load_config_handles_empty_yaml(tmp_path):
    """Test load_config handles empty YAML file."""
    config_file = tmp_path / "rtmx.yaml"
    config_file.write_text("")

    config = load_config(config_file)

    assert config.database == Path("docs/rtm_database.csv")


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_save_config_creates_yaml_file(tmp_path):
    """Test save_config creates YAML file with correct content."""
    config = RTMXConfig(
        database=Path("test.csv"),
        schema="phoenix",
    )
    save_path = tmp_path / "output.yaml"

    result_path = save_config(config, save_path)

    assert result_path == save_path
    assert save_path.exists()

    # Verify content
    with save_path.open() as f:
        data = yaml.safe_load(f)

    assert data["rtmx"]["database"] == "test.csv"
    assert data["rtmx"]["schema"] == "phoenix"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_save_config_uses_default_path_when_none(tmp_path, monkeypatch):
    """Test save_config uses rtmx.yaml in cwd when no path specified."""
    monkeypatch.chdir(tmp_path)

    config = RTMXConfig()
    result_path = save_config(config)

    expected_path = tmp_path / "rtmx.yaml"
    assert result_path == expected_path
    assert expected_path.exists()


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_save_config_overwrites_existing_file(tmp_path):
    """Test save_config overwrites existing configuration file."""
    save_path = tmp_path / "rtmx.yaml"
    save_path.write_text("old content")

    config = RTMXConfig(schema="new_schema")
    save_config(config, save_path)

    with save_path.open() as f:
        data = yaml.safe_load(f)

    assert data["rtmx"]["schema"] == "new_schema"


@pytest.mark.req("REQ-CORE-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_config_roundtrip(tmp_path):
    """Test configuration can be saved and loaded without loss."""
    original = RTMXConfig(
        database=Path("my_db.csv"),
        requirements_dir=Path("my_reqs/"),
        schema="phoenix",
    )
    original.pytest.marker_prefix = "test"
    original.agents.claude.enabled = False
    original.adapters.github.enabled = True
    original.adapters.github.repo = "owner/repo"
    original.mcp.enabled = True
    original.mcp.port = 8080
    original.sync.conflict_resolution = "prefer-local"

    save_path = tmp_path / "config.yaml"
    save_config(original, save_path)
    loaded = load_config(save_path)

    assert loaded.database == original.database
    assert loaded.requirements_dir == original.requirements_dir
    assert loaded.schema == original.schema
    assert loaded.pytest.marker_prefix == original.pytest.marker_prefix
    assert loaded.agents.claude.enabled == original.agents.claude.enabled
    assert loaded.adapters.github.enabled == original.adapters.github.enabled
    assert loaded.adapters.github.repo == original.adapters.github.repo
    assert loaded.mcp.enabled == original.mcp.enabled
    assert loaded.mcp.port == original.mcp.port
    assert loaded.sync.conflict_resolution == original.sync.conflict_resolution
