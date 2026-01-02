# RTMX Configuration Guide

This guide covers all configuration options for RTMX.

## Configuration File

RTMX uses `rtmx.yaml` for project-specific configuration. The file is automatically discovered by searching upward from the current directory.

### File Discovery Order

1. `rtmx.yaml` in current directory
2. `rtmx.yml` in current directory (`.yaml` preferred)
3. Search parent directories until filesystem root

### CLI Override

```bash
rtmx --config /path/to/rtmx.yaml status
```

## Complete Configuration Reference

```yaml
rtmx:
  # ==========================================================================
  # Core Settings
  # ==========================================================================

  # Path to the RTM database CSV file
  # Default: docs/rtm_database.csv
  database: docs/rtm_database.csv

  # Directory containing requirement specification files
  # Default: docs/requirements
  requirements_dir: docs/requirements

  # Schema to use: "core" (20 columns) or "phoenix" (45+ columns)
  # Default: core
  schema: core

  # ==========================================================================
  # Pytest Integration
  # ==========================================================================

  pytest:
    # Prefix for requirement markers (@pytest.mark.req)
    # Default: req
    marker_prefix: "req"

    # Automatically register markers with pytest
    # Default: true
    register_markers: true

  # ==========================================================================
  # AI Agent Integration
  # ==========================================================================

  agents:
    # Claude Code configuration
    claude:
      enabled: true                    # Enable Claude integration
      config_path: "CLAUDE.md"         # Path to Claude config file

    # Cursor IDE configuration
    cursor:
      enabled: true
      config_path: ".cursorrules"

    # GitHub Copilot configuration
    copilot:
      enabled: true
      config_path: ".github/copilot-instructions.md"

    # Directory for custom prompt templates
    # Default: .rtmx/templates/
    template_dir: ".rtmx/templates/"

  # ==========================================================================
  # External Adapters
  # ==========================================================================

  adapters:
    # GitHub Issues integration
    github:
      enabled: false                   # Enable GitHub sync
      repo: "owner/repo"               # Repository (owner/name)
      token_env: "GITHUB_TOKEN"        # Environment variable for token
      labels:                          # Labels to filter issues
        - "requirement"
        - "feature"
      status_mapping:                  # GitHub state to RTMX status
        open: "MISSING"
        closed: "COMPLETE"

    # Jira integration
    jira:
      enabled: false
      server: "https://company.atlassian.net"
      project: "PROJ"                  # Jira project key
      token_env: "JIRA_API_TOKEN"      # Environment variable for API token
      email_env: "JIRA_EMAIL"          # Environment variable for email
      issue_type: "Requirement"        # Jira issue type to create
      jql_filter: ""                   # Optional JQL to filter tickets
      labels: []                       # Labels to apply to created issues
      status_mapping:
        "To Do": "MISSING"
        "In Progress": "PARTIAL"
        "Done": "COMPLETE"

  # ==========================================================================
  # MCP Server (Model Context Protocol)
  # ==========================================================================

  mcp:
    enabled: false                     # Enable MCP server
    host: "localhost"                  # Bind address
    port: 3000                         # Server port

  # ==========================================================================
  # Synchronization Settings
  # ==========================================================================

  sync:
    # How to resolve conflicts during bidirectional sync
    # Options: "manual", "prefer-local", "prefer-remote"
    # Default: manual
    conflict_resolution: "manual"
```

## Configuration Sections

### Core Settings

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `database` | string | `docs/rtm_database.csv` | Path to RTM CSV file |
| `requirements_dir` | string | `docs/requirements` | Directory for spec files |
| `schema` | string | `core` | Schema type (`core` or `phoenix`) |

### Pytest Settings

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `marker_prefix` | string | `req` | Marker prefix for tests |
| `register_markers` | bool | `true` | Auto-register with pytest |

### Agent Settings

Each agent (claude, cursor, copilot) supports:

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `enabled` | bool | `true` | Enable this agent |
| `config_path` | string | varies | Path to agent config file |

Default config paths:
- Claude: `CLAUDE.md`
- Cursor: `.cursorrules`
- Copilot: `.github/copilot-instructions.md`

### GitHub Adapter Settings

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `enabled` | bool | `false` | Enable GitHub sync |
| `repo` | string | `""` | Repository (owner/name) |
| `token_env` | string | `GITHUB_TOKEN` | Environment variable for token |
| `labels` | list | `[requirement, feature]` | Issue labels to filter |
| `status_mapping` | object | see below | State to status mapping |

Default status mapping:
```yaml
status_mapping:
  open: "MISSING"
  closed: "COMPLETE"
```

### Jira Adapter Settings

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `enabled` | bool | `false` | Enable Jira sync |
| `server` | string | `""` | Jira server URL |
| `project` | string | `""` | Jira project key |
| `token_env` | string | `JIRA_API_TOKEN` | Env var for API token |
| `email_env` | string | `JIRA_EMAIL` | Env var for email |
| `issue_type` | string | `Requirement` | Jira issue type |
| `jql_filter` | string | `""` | Optional JQL filter |
| `labels` | list | `[]` | Labels for new issues |
| `status_mapping` | object | see below | Status mapping |

Default status mapping:
```yaml
status_mapping:
  "To Do": "MISSING"
  "In Progress": "PARTIAL"
  "Done": "COMPLETE"
```

### MCP Settings

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `enabled` | bool | `false` | Enable MCP server |
| `host` | string | `localhost` | Bind address |
| `port` | int | `3000` | Server port |

### Sync Settings

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| `conflict_resolution` | string | `manual` | Conflict resolution strategy |

Conflict resolution options:
- `manual` - Prompt for each conflict
- `prefer-local` - RTM database wins
- `prefer-remote` - External service wins

## Environment Variables

RTMX uses environment variables for sensitive credentials:

| Variable | Used By | Description |
|----------|---------|-------------|
| `GITHUB_TOKEN` | GitHub adapter | Personal access token |
| `JIRA_API_TOKEN` | Jira adapter | Jira API token |
| `JIRA_EMAIL` | Jira adapter | Email for Jira authentication |

## Validation

Validate your configuration:

```bash
rtmx config --validate
```

This checks:
- Database file exists and is readable
- Requirements directory exists
- Schema is valid (`core` or `phoenix`)
- Required environment variables are set (for enabled adapters)

## Show Current Configuration

```bash
rtmx config                  # Terminal output
rtmx config --format yaml    # YAML output
rtmx config --format json    # JSON output
```

## Minimal Configuration

For simple projects, a minimal config is sufficient:

```yaml
rtmx:
  database: docs/rtm_database.csv
```

All other settings use sensible defaults.

## Programmatic Access

```python
from rtmx import RTMXConfig, load_config

# Load configuration
config = load_config()  # Auto-discover rtmx.yaml

# Or specify path
config = load_config("/path/to/rtmx.yaml")

# Access settings
print(config.database)
print(config.adapters.github.enabled)
print(config.to_dict())
```

## See Also

- [Schema Documentation](schema.md) - RTM database schema
- [LIFECYCLE.md](LIFECYCLE.md) - Full lifecycle documentation
- [README.md](../README.md) - Quick start guide
