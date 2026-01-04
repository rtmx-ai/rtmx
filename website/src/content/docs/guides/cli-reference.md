---
title: CLI Reference
description: Complete command reference for RTMX CLI
---

## Core Commands

| Command | Description |
|---------|-------------|
| `rtmx setup` | Complete RTMX setup (config, RTM, agents, Makefile) |
| `rtmx init` | Minimal setup (config and RTM database only) |
| `rtmx status` | Show completion progress (`-v`, `-vv`, `-vvv` for detail) |
| `rtmx backlog` | Show prioritized incomplete requirements |
| `rtmx health` | Run integration health checks |
| `rtmx config` | Show or validate configuration |

## Analysis Commands

| Command | Description |
|---------|-------------|
| `rtmx deps` | Show dependency graph |
| `rtmx cycles` | Detect circular dependencies |
| `rtmx reconcile` | Check/fix dependency reciprocity |
| `rtmx analyze` | Discover requirements from project artifacts |
| `rtmx diff` | Compare RTM versions (for PRs) |

## Integration Commands

| Command | Description |
|---------|-------------|
| `rtmx from-tests` | Scan tests for requirement markers |
| `rtmx bootstrap` | Generate RTM from tests, GitHub, or Jira |
| `rtmx sync` | Synchronize with GitHub Issues or Jira |
| `rtmx install` | Install prompts into AI agent configs |
| `rtmx makefile` | Generate Makefile targets |
| `rtmx mcp-server` | Start MCP server for AI agent integration |

## Command Details

### rtmx setup

Complete project setup with all RTMX features:

```bash
rtmx setup              # Interactive setup
rtmx setup --branch     # Create git branch for review
rtmx setup --pr         # Create branch and open PR
```

Creates:
- `rtmx.yaml` configuration
- `docs/rtm_database.csv` requirements database
- `docs/requirements/` specification directory
- Makefile targets
- AI agent configs (CLAUDE.md, .cursorrules)

### rtmx status

Show project progress:

```bash
rtmx status         # Summary only
rtmx status -v      # Category breakdown
rtmx status -vv     # Requirement details
rtmx status -vvv    # Maximum detail
```

### rtmx backlog

Show prioritized incomplete requirements:

```bash
rtmx backlog              # All incomplete
rtmx backlog --phase 1    # Filter by phase
rtmx backlog --blocked    # Show only blocked items
```

### rtmx health

Run integration health checks:

```bash
rtmx health              # Standard output
rtmx health --format ci  # CI-friendly output
rtmx health --fix        # Auto-fix issues where possible
```

### rtmx deps

Analyze dependency relationships:

```bash
rtmx deps REQ-AUTH-001              # Show dependencies for a requirement
rtmx deps --all                     # Show full dependency graph
rtmx deps --critical-path           # Show critical path
```

### rtmx sync

Synchronize with external services:

```bash
rtmx sync github          # Sync to GitHub Issues
rtmx sync jira            # Sync to Jira
rtmx sync --dry-run       # Preview changes without applying
```

### rtmx mcp-server

Start MCP server for AI agent integration:

```bash
rtmx mcp-server                    # Start on stdio
rtmx mcp-server --port 8080        # Start on HTTP
```

## Global Options

| Option | Description |
|--------|-------------|
| `--config PATH` | Use specific config file |
| `--database PATH` | Use specific RTM database |
| `--verbose`, `-v` | Increase verbosity |
| `--quiet`, `-q` | Decrease verbosity |
| `--help`, `-h` | Show help message |
| `--version` | Show version |
