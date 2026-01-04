---
title: Configuration
description: Configure RTMX for your project
---

RTMX is configured via `rtmx.yaml` in your project root.

## Basic Configuration

```yaml
rtmx:
  database: docs/rtm_database.csv
  requirements_dir: docs/requirements
  schema: core
```

## Full Configuration

```yaml
rtmx:
  # Database location
  database: docs/rtm_database.csv

  # Requirements specification directory
  requirements_dir: docs/requirements

  # Schema type: core, phoenix, or custom
  schema: core

  # Pytest integration
  pytest:
    marker_prefix: "req"
    register_markers: true
    strict_markers: false

  # GitHub integration
  github:
    enabled: false
    repo: owner/repo
    label_prefix: "req:"
    sync_on_push: false

  # Jira integration
  jira:
    enabled: false
    url: https://your-org.atlassian.net
    project_key: PROJ
    sync_on_push: false

  # MCP server settings
  mcp:
    enabled: true
    tools:
      - status
      - backlog
      - get_requirement
      - update_status
      - deps
      - search

  # Validation rules
  validation:
    require_test_function: false
    require_acceptance_criteria: false
    require_owner: false
    max_phase: 10
```

## Configuration Options

### Database Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `database` | string | `docs/rtm_database.csv` | Path to RTM database |
| `requirements_dir` | string | `docs/requirements` | Path to requirement specs |
| `schema` | string | `core` | Schema type |

### Pytest Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `marker_prefix` | string | `req` | Prefix for requirement markers |
| `register_markers` | bool | `true` | Auto-register markers |
| `strict_markers` | bool | `false` | Fail on unknown markers |

### GitHub Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `false` | Enable GitHub sync |
| `repo` | string | - | Repository in `owner/repo` format |
| `label_prefix` | string | `req:` | Prefix for issue labels |
| `sync_on_push` | bool | `false` | Sync on git push |

### Jira Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `false` | Enable Jira sync |
| `url` | string | - | Jira instance URL |
| `project_key` | string | - | Jira project key |
| `sync_on_push` | bool | `false` | Sync on git push |

### MCP Settings

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `enabled` | bool | `true` | Enable MCP server |
| `tools` | list | all | Enabled MCP tools |

## Environment Variables

Override configuration with environment variables:

```bash
RTMX_DATABASE=custom/path.csv rtmx status
RTMX_GITHUB_TOKEN=ghp_xxx rtmx sync github
JIRA_API_TOKEN=xxx rtmx sync jira
```

## Multiple Configurations

Use different configs for different environments:

```bash
rtmx --config rtmx.production.yaml status
rtmx --config rtmx.staging.yaml status
```

## Validation

Validate your configuration:

```bash
rtmx config              # Show current config
rtmx config --validate   # Validate config
rtmx config --init       # Create default config
```
