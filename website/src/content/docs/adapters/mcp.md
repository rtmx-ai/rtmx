---
title: MCP Server
description: AI agent integration via Model Context Protocol
---

RTMX includes an MCP (Model Context Protocol) server that enables AI agents like Claude to interact with your requirements database.

## What is MCP?

[Model Context Protocol](https://modelcontextprotocol.io/) is an open standard for connecting AI assistants to external data sources and tools.

With RTMX MCP, agents can:
- Query project status
- View backlog and blockers
- Update requirement status
- Analyze dependencies

## Quick Start

### Start the MCP Server

```bash
rtmx mcp-server
```

This starts the server on stdio (standard input/output), the default MCP transport.

### Configure Claude Desktop

Add to your Claude Desktop config (`~/.config/claude/claude_desktop_config.json`):

```json
{
  "mcpServers": {
    "rtmx": {
      "command": "rtmx",
      "args": ["mcp-server"],
      "cwd": "/path/to/your/project"
    }
  }
}
```

## Available Tools

### rtmx_status

Get project completion status.

```json
{
  "name": "rtmx_status",
  "description": "Get RTM completion status",
  "parameters": {
    "verbose": {
      "type": "boolean",
      "description": "Include detailed breakdown"
    }
  }
}
```

### rtmx_backlog

Get prioritized incomplete requirements.

```json
{
  "name": "rtmx_backlog",
  "parameters": {
    "phase": {
      "type": "integer",
      "description": "Filter by phase number"
    },
    "limit": {
      "type": "integer",
      "description": "Maximum items to return"
    }
  }
}
```

### rtmx_get_requirement

Get details for a specific requirement.

```json
{
  "name": "rtmx_get_requirement",
  "parameters": {
    "req_id": {
      "type": "string",
      "description": "Requirement ID (e.g., REQ-AUTH-001)"
    }
  }
}
```

### rtmx_update_status

Update a requirement's status.

```json
{
  "name": "rtmx_update_status",
  "parameters": {
    "req_id": {
      "type": "string",
      "description": "Requirement ID"
    },
    "status": {
      "type": "string",
      "enum": ["COMPLETE", "PARTIAL", "MISSING", "WIP", "BLOCKED"]
    }
  }
}
```

### rtmx_deps

Get dependency information.

```json
{
  "name": "rtmx_deps",
  "parameters": {
    "req_id": {
      "type": "string",
      "description": "Requirement ID"
    },
    "direction": {
      "type": "string",
      "enum": ["upstream", "downstream", "both"]
    }
  }
}
```

### rtmx_search

Search requirements by text or filters.

```json
{
  "name": "rtmx_search",
  "parameters": {
    "query": {
      "type": "string",
      "description": "Search text"
    },
    "category": {
      "type": "string",
      "description": "Filter by category"
    },
    "status": {
      "type": "string",
      "description": "Filter by status"
    }
  }
}
```

## Configuration

Enable/disable specific tools in `rtmx.yaml`:

```yaml
rtmx:
  mcp:
    enabled: true
    tools:
      - status
      - backlog
      - get_requirement
      - update_status
      - deps
      - search
```

## Use Cases

### Agent-Assisted Development

Claude can help implement requirements:

```
User: "What should I work on next?"
Claude: [calls rtmx_backlog] "The highest priority item is REQ-AUTH-002:
        Add password reset flow. It has no blockers."

User: "I finished implementing it"
Claude: [calls rtmx_update_status with status=COMPLETE]
        "Updated REQ-AUTH-002 to COMPLETE. 3 requirements are now unblocked."
```

### Progress Reporting

```
User: "How's the project going?"
Claude: [calls rtmx_status] "Phase 1 is 95% complete with 2 remaining items.
        Overall progress is 67% across all phases."
```

### Dependency Analysis

```
User: "What's blocking REQ-DATA-005?"
Claude: [calls rtmx_deps] "REQ-DATA-005 is blocked by REQ-DB-001 (MISSING)
        and REQ-AUTH-001 (COMPLETE). You need to complete REQ-DB-001 first."
```

## Security Considerations

- MCP server runs locally with your project permissions
- No network exposure by default (stdio transport)
- Status updates are saved to your local CSV
- Review changes before git commits
