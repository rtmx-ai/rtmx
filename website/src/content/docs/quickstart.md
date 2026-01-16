---
title: Quickstart
description: Get up and running with RTMX in 5 minutes
---

## 1. Set up RTMX in your project

```bash
rtmx setup
```

This creates everything you need:
- `rtmx.yaml` — Configuration file
- `docs/rtm_database.csv` — Requirements database
- `docs/requirements/` — Requirement specification files
- Makefile targets (`make rtm`, `make backlog`)
- AI agent configs (CLAUDE.md, .cursorrules)

For existing projects, use `--branch` to review changes before merging:

```bash
rtmx setup --branch    # Creates a git branch for review
rtmx setup --pr        # Creates branch + opens PR
```

## 2. Check your status

```bash
rtmx status -v    # Verbose output with category breakdown
rtmx status -vv   # Very verbose with requirement details
rtmx status -vvv  # Maximum detail
```

## 3. See what to work on next

```bash
rtmx backlog
```

The backlog shows requirements sorted by priority and phase, with dependency information.

## 4. Run health checks

```bash
rtmx health
```

Health checks verify:
- RTM database can be loaded
- All requirement IDs are valid
- Dependencies reference existing requirements
- Test files and functions exist

## 5. Link tests to requirements

```python
import pytest

@pytest.mark.req("REQ-AUTH-001")
@pytest.mark.scope_unit
def test_oauth_login():
    """Validates REQ-AUTH-001: OAuth 2.0 authentication."""
    assert authenticate_user(token) == expected_user
```

Then scan your tests:

```bash
rtmx from-tests
```

## Using the Makefile

After `rtmx setup`, you get these Makefile targets:

```bash
make rtm       # Quick status check
make backlog   # See what's next
make health    # Run health checks
```

## Next Steps

- [CLI Reference](/docs/guides/cli-reference) — All available commands
- [Test Markers](/docs/guides/markers) — pytest integration
- [Schema](/docs/guides/schema) — RTM database schema
- [MCP Server](/docs/adapters/mcp) — AI agent integration
