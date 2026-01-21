---
title: Installation
description: Install RTMX and set up your project
---

## Requirements

- Python 3.10 or higher
- pip or pipx

## Install from PyPI

```bash
pip install rtmx
```

Or with pipx for isolated installation:

```bash
pipx install rtmx
```

## Verify Installation

```bash
rtmx --version
```

## Development Installation

For contributing to RTMX:

```bash
git clone https://github.com/rtmx-ai/rtmx.git
cd rtmx
make dev      # Install with dev dependencies
make test     # Run tests
make lint     # Run linter
```

## Optional Dependencies

RTMX has optional features that require additional packages:

```bash
# GitHub integration
pip install rtmx[github]

# Jira integration
pip install rtmx[jira]

# Web UI dashboard (rtmx serve)
pip install rtmx[web]

# Terminal UI (rtmx tui)
pip install rtmx[tui]

# Rich formatting for status output
pip install rtmx[rich]

# MCP server for AI agent integration
pip install rtmx[mcp]

# All optional dependencies
pip install rtmx[all]
```
