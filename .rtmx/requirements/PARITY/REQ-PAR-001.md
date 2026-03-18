# REQ-PAR-001: JSON Output Flag for Status and Backlog

## Metadata
- **Category**: PARITY
- **Subcategory**: Output
- **Priority**: P0
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-010, REQ-GO-011
- **Blocks**: REQ-GO-047

## Requirement

`rtmx status --json` and `rtmx backlog --json` shall produce machine-readable JSON output matching the Python CLI schema, enabling CI/CD pipeline integration.

## Rationale

CI/CD pipelines, dashboards, and automation tools consume JSON output from `rtmx status --json`. Without this flag, Go CLI cannot replace Python CLI in automated workflows.

## Design

### CLI Interface

```bash
rtmx status --json                    # JSON to stdout
rtmx status --json --output report.json  # JSON to file
rtmx backlog --json                   # JSON to stdout
```

### Status JSON Schema

```json
{
  "total": 81,
  "complete": 55,
  "partial": 0,
  "missing": 26,
  "completion_pct": 67.9,
  "phases": [
    {"phase": 1, "name": "Foundation", "total": 3, "complete": 3, "pct": 100.0}
  ],
  "categories": [
    {"name": "CLI", "total": 20, "complete": 18, "pct": 90.0}
  ]
}
```

### Backlog JSON Schema

```json
{
  "total_missing": 26,
  "estimated_effort_weeks": 48.0,
  "critical_path": [
    {"req_id": "REQ-GO-047", "description": "...", "effort": 2.0, "blocks": 0}
  ],
  "remaining": [
    {"req_id": "REQ-GO-034", "description": "...", "priority": "MEDIUM", "blocked": false}
  ]
}
```

## Acceptance Criteria

1. `rtmx status --json` outputs valid JSON to stdout
2. `rtmx backlog --json` outputs valid JSON to stdout
3. JSON schema matches Python CLI output (field names, types, nesting)
4. `--json` suppresses all non-JSON output (no headers, progress bars)
5. JSON output is parseable by `jq` without errors
6. Exit codes unchanged (0 on success)

## Files to Modify

- `internal/cmd/status.go` - Add `--json` flag and JSON rendering
- `internal/cmd/backlog.go` - Add `--json` flag and JSON rendering
- `internal/cmd/status_test.go` - JSON output tests
- `internal/cmd/backlog_test.go` - JSON output tests
