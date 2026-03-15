# REQ-MIG-001: Python-Go Feature Parity Validation

## Status: MISSING
## Priority: CRITICAL
## Phase: 21

## Description

RTMX shall validate full feature parity between the Python CLI (`rtmx`) and Go CLI (`rtmx-go`) before migrating the main branch. A comprehensive parity test suite shall compare command output, database operations, and behavioral equivalence across all CLI commands.

## Rationale

Switching the main branch from Python to Go is a one-way door. Without validated feature parity, users will lose functionality they depend on. Automated parity validation ensures no regressions during the migration.

## Acceptance Criteria

- [ ] Parity test matrix covering all CLI commands (`status`, `backlog`, `health`, `validate`, `graph`, `from-tests`, `remote`, `init`, `verify`)
- [ ] Golden file tests comparing Python and Go output for identical inputs
- [ ] CSV round-trip tests ensuring Go CLI reads/writes identical database format
- [ ] Configuration loading parity (rtmx.yaml parsing produces identical behavior)
- [ ] Graph algorithm parity (Tarjan SCC, topological sort, critical path produce identical results)
- [ ] Exit code parity for all error conditions
- [ ] Deprecation warning system functional in Go CLI
- [ ] All edge cases from Python test suite have Go equivalents
- [ ] Parity report generated as CI artifact showing command-by-command coverage

## Test Cases

1. `tests/test_parity.py::test_status_output_matches` - Status command output parity
2. `tests/test_parity.py::test_backlog_output_matches` - Backlog command output parity
3. `tests/test_parity.py::test_health_output_matches` - Health command output parity
4. `tests/test_parity.py::test_csv_round_trip_parity` - Database read/write parity
5. `tests/test_parity.py::test_graph_algorithms_parity` - Graph computation parity
6. `tests/test_parity.py::test_config_loading_parity` - Configuration parity
7. `tests/test_parity.py::test_exit_codes_parity` - Error handling parity
8. `tests/test_parity.py::test_parity_report_generated` - CI parity report artifact

## Dependencies

- REQ-DIST-002: Go CLI binary distribution (Go CLI must exist)
- REQ-LANG-003: Go testing integration

## Blocks

- REQ-MIG-002: Main branch migration (cannot migrate without validated parity)

## Effort

3.0 weeks
