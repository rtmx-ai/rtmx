---
title: Schema
description: RTM database schema and column definitions
---

The RTM database is a CSV file with standardized columns for requirements traceability.

## Core Schema

| Column | Required | Description |
|--------|----------|-------------|
| `req_id` | Yes | Unique identifier (e.g., `REQ-AUTH-001`) |
| `category` | Yes | High-level grouping |
| `subcategory` | No | Secondary grouping |
| `requirement_text` | Yes | Human-readable description |
| `acceptance_criteria` | No | Verification criteria |
| `test_module` | No | Test file path |
| `test_function` | No | Test function name |
| `verification_method` | No | How requirement is verified |
| `status` | Yes | Current status |
| `priority` | No | Priority level |
| `phase` | No | Development phase number |
| `notes` | No | Additional notes |
| `effort` | No | Estimated effort |
| `dependencies` | No | Requirements this depends on |
| `blocks` | No | Requirements this blocks |
| `owner` | No | Responsible party |
| `target_release` | No | Target release version |
| `actual_release` | No | Actual release version |
| `date_completed` | No | Completion date |
| `spec_path` | No | Path to requirement specification |

## Status Values

| Status | Description |
|--------|-------------|
| `COMPLETE` | Requirement fully implemented and tested |
| `PARTIAL` | Requirement partially implemented |
| `MISSING` | Requirement not yet implemented |
| `NOT_STARTED` | Requirement not yet started |
| `WIP` | Work in progress |
| `BLOCKED` | Blocked by dependencies |

## Priority Values

| Priority | Description |
|----------|-------------|
| `P0` | Critical - must have |
| `HIGH` | Important - should have |
| `MEDIUM` | Nice to have |
| `LOW` | Future consideration |

## Requirement ID Format

Requirement IDs follow the pattern: `REQ-{CATEGORY}-{NUMBER}`

Examples:
- `REQ-AUTH-001` - Authentication requirement 1
- `REQ-DATA-015` - Data requirement 15
- `REQ-UI-003` - UI requirement 3

## Dependencies

Dependencies are pipe-separated requirement IDs:

```csv
REQ-AUTH-002,AUTH,...,REQ-AUTH-001|REQ-DB-001,...
```

This means `REQ-AUTH-002` depends on both `REQ-AUTH-001` and `REQ-DB-001`.

## Schema Extensions

### Phoenix Schema

Extended schema for Phoenix-style projects:

```yaml
rtmx:
  schema: phoenix
```

Additional columns:
- `compliance_ref` - Compliance reference
- `risk_level` - Risk assessment
- `verification_level` - Verification level required

### Custom Fields

Define custom fields in configuration:

```yaml
rtmx:
  schema: core
  custom_fields:
    - name: customer_ref
      type: string
      required: false
    - name: sprint_id
      type: integer
      required: false
```

## Example RTM Entry

```csv
req_id,category,subcategory,requirement_text,acceptance_criteria,test_module,test_function,verification_method,status,priority,phase,notes,effort,dependencies,blocks,owner,target_release,actual_release,date_completed,spec_path
REQ-AUTH-001,AUTH,LOGIN,System shall support OAuth 2.0 authentication,OAuth flow completes successfully,tests/test_auth.py,test_oauth_login,Unit Test,COMPLETE,HIGH,1,Implemented using authlib,2.0,,,@alice,v0.1,v0.1,2024-01-15,docs/requirements/AUTH/REQ-AUTH-001.md
```

## Validation

Validate your RTM database:

```bash
rtmx health              # Full health check
rtmx config --validate   # Config validation only
```

Common validation errors:
- Duplicate requirement IDs
- Invalid status values
- Missing required fields
- Circular dependencies
- References to non-existent requirements
