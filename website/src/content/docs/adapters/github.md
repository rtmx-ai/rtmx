---
title: GitHub Adapter
description: Sync requirements with GitHub Issues
---

The GitHub adapter synchronizes your RTM requirements with GitHub Issues, enabling issue tracking integration.

## Setup

### 1. Configure GitHub Token

Set the `GITHUB_TOKEN` environment variable:

```bash
export GITHUB_TOKEN=ghp_your_token_here
```

Required scopes:
- `repo` (for private repos)
- `public_repo` (for public repos only)

### 2. Enable in Configuration

```yaml
rtmx:
  github:
    enabled: true
    repo: owner/repo
    label_prefix: "req:"
```

## Sync Commands

### Push to GitHub

```bash
rtmx sync github              # Sync all requirements
rtmx sync github --dry-run    # Preview changes
rtmx sync github --filter status=MISSING
```

### Behavior

For each requirement:

1. **Creates** a GitHub issue if none exists
2. **Updates** existing issue if requirement changed
3. **Labels** issues with `req:{category}` and status

### Issue Format

Created issues follow this template:

```markdown
# REQ-AUTH-001: OAuth 2.0 Authentication

**Status**: COMPLETE
**Priority**: HIGH
**Phase**: 1

## Description
System shall support OAuth 2.0 authentication.

## Acceptance Criteria
OAuth flow completes successfully.

## Dependencies
- REQ-DB-001

---
_Managed by RTMX. Do not edit this section._
```

## Labels

The adapter creates labels automatically:

| Label | Description |
|-------|-------------|
| `req:AUTH` | Category label |
| `status:COMPLETE` | Status label |
| `phase:1` | Phase label |
| `priority:HIGH` | Priority label |

## Configuration Options

```yaml
rtmx:
  github:
    enabled: true
    repo: owner/repo
    label_prefix: "req:"
    create_missing: true        # Create issues for new requirements
    close_completed: false      # Close issues for COMPLETE requirements
    assignee_field: owner       # Map RTM owner to GitHub assignee
    milestone_field: phase      # Map RTM phase to GitHub milestone
```

## CI Integration

Sync automatically on push:

```yaml
# .github/workflows/rtmx-sync.yml
name: RTMX Sync
on:
  push:
    paths:
      - 'docs/rtm_database.csv'
jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: pip install rtmx
      - run: rtmx sync github
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

## Bidirectional Sync

Currently, sync is **one-way** (RTM → GitHub). The RTM CSV is the source of truth.

To incorporate GitHub changes:
1. Review issues in GitHub
2. Update `rtm_database.csv` manually or via PR
3. Run `rtmx sync github` to push updates

## Troubleshooting

### Rate Limits

GitHub has API rate limits. RTMX uses delta tracking to minimize calls.

If you hit limits:
```bash
# Check remaining quota
gh api rate_limit

# Sync only specific requirements
rtmx sync github --filter category=AUTH
```

### Authentication Errors

```bash
# Verify token
gh auth status

# Test API access
gh api repos/owner/repo
```
