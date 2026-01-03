# REQ-SCHEMA-001: RTM database shall support named phases

## Status: MISSING
## Priority: HIGH
## Phase: 4

## Description
RTM database shall support named phases, allowing phase numbers to be mapped to human-readable names. This provides meaningful context for project managers and stakeholders.

## Acceptance Criteria
- [ ] rtmx.yaml supports phase definitions with names
- [ ] Phase names displayed in status output instead of/alongside numbers
- [ ] Phase names displayed in backlog output
- [ ] Phase names displayed in rich progress output
- [ ] Phase names optional - defaults to "Phase N" if not defined
- [ ] CLI commands support --phase by name or number

## Configuration Schema

```yaml
rtmx:
  phases:
    1: "Foundation"
    2: "Core Features"
    3: "Testing"
    4: "Developer Experience"
    5: "CLI UX"
    6: "Web UI"
    7: "Real-Time"
    8: "Website"
    9: "CRDT"
    10: "Collaboration"
    11: "IDE"
    12: "Distribution"
```

## Display Format

```
╭─ Phase Progress ─────────────────────────────────────────╮
│ Phase 1 (Foundation):     ████████████████████████ 100% │
│ Phase 2 (Core Features):  ████████████████████████ 100% │
│ Phase 3 (Testing):        ████████████████████████ 100% │
│ Phase 4 (Developer Exp):  ██████████████████████░░  91% │
│ Phase 5 (CLI UX):         ██████████░░░░░░░░░░░░░░  33% │
│ Phase 6 (Web UI):         ░░░░░░░░░░░░░░░░░░░░░░░░   0% │
╰──────────────────────────────────────────────────────────╯
```

## Test Cases
- `tests/test_models.py::test_phase_names`
- `tests/test_models.py::test_phase_names_fallback`
- `tests/test_cli_commands.py::test_status_phase_names`
- `tests/test_cli_commands.py::test_backlog_phase_by_name`

## Implementation Notes
- Add `phases` dict to RTMXConfig
- Update RTMDatabase to load phase names from config
- Update status/backlog/health formatters to display phase names
- Support both `--phase 5` and `--phase "CLI UX"` in CLI
