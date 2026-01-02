# REQ-DX-004: Auto-Generate Documentation

## Status: NOT_STARTED
## Priority: MEDIUM
## Phase: 4
## Effort: 0.5 weeks

## Description

Add `rtmx docs` command to generate schema, config, and API documentation automatically.

## Acceptance Criteria

- [ ] `rtmx docs schema` generates schema.md from rtmx.schema module
- [ ] `rtmx docs config` generates config reference from config schema
- [ ] `rtmx docs api` generates API docs using pdoc
- [ ] Output to `.rtmx/cache/` by default
- [ ] `--output` flag to specify custom location
- [ ] Generated docs include version and timestamp

## Commands

```bash
# Generate schema documentation
rtmx docs schema
# Output: .rtmx/cache/schema.md

# Generate configuration reference
rtmx docs config
# Output: .rtmx/cache/configuration.md

# Generate API documentation
rtmx docs api
# Output: .rtmx/cache/api-docs/

# Specify custom output location
rtmx docs schema --output docs/schema.md

# Generate all documentation
rtmx docs all
```

## Schema Documentation Content

```markdown
# RTMX Database Schema

Generated: 2026-01-02 10:15:00
Version: 0.0.2

## Columns

| Column | Type | Required | Description |
|--------|------|----------|-------------|
| req_id | string | Yes | Unique requirement identifier |
| category | string | Yes | Requirement category |
...
```

## Config Documentation Content

```markdown
# RTMX Configuration Reference

Generated: 2026-01-02 10:15:00
Version: 0.0.2

## Configuration File Location

- `.rtmx/config.yaml` (preferred)
- `rtmx.yaml` (legacy)

## Options

### database
Path to RTM database CSV file.
- Type: string
- Default: `.rtmx/database.csv`
- Environment: `RTMX_DATABASE`
...
```

## Files to Create/Modify

- NEW `src/rtmx/cli/docs.py` - Docs command
- `src/rtmx/cli/main.py` - Register docs command

## Dependencies

- REQ-DX-001 (uses .rtmx/cache/)
- pdoc already in dev dependencies

## Notes

This eliminates manual documentation maintenance and ensures docs stay in sync with code.
