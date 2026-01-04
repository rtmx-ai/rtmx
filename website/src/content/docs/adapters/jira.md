---
title: Jira Adapter
description: Sync requirements with Jira tickets
---

The Jira adapter synchronizes your RTM requirements with Jira, enabling enterprise issue tracking integration.

## Setup

### 1. Configure Jira Credentials

Set environment variables:

```bash
export JIRA_URL=https://your-org.atlassian.net
export JIRA_EMAIL=your-email@company.com
export JIRA_API_TOKEN=your_api_token
```

Generate an API token at: https://id.atlassian.com/manage-profile/security/api-tokens

### 2. Enable in Configuration

```yaml
rtmx:
  jira:
    enabled: true
    url: https://your-org.atlassian.net
    project_key: PROJ
    issue_type: Story
```

## Sync Commands

### Push to Jira

```bash
rtmx sync jira              # Sync all requirements
rtmx sync jira --dry-run    # Preview changes
rtmx sync jira --filter phase=1
```

### Behavior

For each requirement:

1. **Creates** a Jira issue if none exists
2. **Updates** existing issue if requirement changed
3. **Sets labels** for category and status

## Field Mapping

| RTM Field | Jira Field |
|-----------|------------|
| `req_id` | Summary prefix |
| `requirement_text` | Summary |
| `acceptance_criteria` | Description |
| `status` | Custom field / Label |
| `priority` | Priority |
| `owner` | Assignee |
| `phase` | Sprint / Version |

## Configuration Options

```yaml
rtmx:
  jira:
    enabled: true
    url: https://your-org.atlassian.net
    project_key: PROJ
    issue_type: Story
    create_missing: true
    status_mapping:
      COMPLETE: Done
      PARTIAL: In Progress
      MISSING: To Do
      WIP: In Progress
      BLOCKED: Blocked
    priority_mapping:
      P0: Highest
      HIGH: High
      MEDIUM: Medium
      LOW: Low
    custom_fields:
      req_id: customfield_10001
      acceptance_criteria: customfield_10002
```

## Issue Format

Created issues include:

- **Summary**: `[REQ-AUTH-001] OAuth 2.0 Authentication`
- **Description**: Full requirement details with acceptance criteria
- **Labels**: `rtmx`, `req-AUTH`, `phase-1`

## Workflow Integration

### Link to Confluence

Reference RTMX requirements in Confluence:

```
{jira:REQ-AUTH-001}
```

### JQL Queries

Find RTMX-managed issues:

```sql
project = PROJ AND labels = rtmx
```

Find incomplete requirements:

```sql
project = PROJ AND labels = rtmx AND status != Done
```

## Troubleshooting

### Permission Errors

Ensure your API token has:
- Browse Projects
- Create Issues
- Edit Issues
- Transition Issues

### Custom Field Access

Some custom fields require admin setup:

```bash
# List available fields
rtmx sync jira --list-fields
```

### Rate Limits

Jira Cloud has rate limits. Use `--batch-size` for large syncs:

```bash
rtmx sync jira --batch-size 50
```
