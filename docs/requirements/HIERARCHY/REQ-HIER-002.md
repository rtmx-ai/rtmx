# REQ-HIER-002: Aggregated Hierarchical RTM

## Status: MISSING
## Priority: HIGH
## Phase: 22

## Description

The `rtmx` CLI shall produce hierarchical RTM reports that aggregate requirements across all discovered vendored projects, providing platform-level visibility into completion status, phase progress, and health metrics while preserving per-project detail.

## Rationale

Platform engineers need to understand the overall health and progress of the RTMX platform, not just individual components. Hierarchical aggregation enables:
- Executive-level platform status dashboards
- Cross-project phase alignment tracking
- Platform-wide critical path analysis
- Compliance reporting across the full system of systems

## Acceptance Criteria

- [ ] `rtmx status` at platform root shows rollup: total requirements, completion percentage, per-project breakdown
- [ ] `rtmx backlog` aggregates and prioritizes across all projects, noting project origin
- [ ] `rtmx health` computes platform-wide health score from component health scores
- [ ] `rtmx graph` visualizes cross-project dependency graph
- [ ] Hierarchical output format:
  ```
  Platform Status: 62.3% (248/398 requirements)
    rtmx:        57.4% (53/95)
    rtmx-sync:   71.2% (89/125)
    aegis-cli:   65.0% (52/80)
    website:     55.1% (54/98)
  ```
- [ ] JSON export includes hierarchical structure for downstream tooling
- [ ] Phase alignment view shows which projects are on which phases
- [ ] Platform-level critical path identifies cross-project bottlenecks

## Test Cases

1. `tests/test_hierarchy.py::test_aggregated_status_rollup` - Platform-wide status calculation
2. `tests/test_hierarchy.py::test_per_project_breakdown` - Individual project stats in rollup
3. `tests/test_hierarchy.py::test_aggregated_backlog` - Cross-project backlog prioritization
4. `tests/test_hierarchy.py::test_platform_health_score` - Composite health metric
5. `tests/test_hierarchy.py::test_cross_project_graph` - Graph spans project boundaries
6. `tests/test_hierarchy.py::test_json_hierarchical_export` - Structured JSON output
7. `tests/test_hierarchy.py::test_phase_alignment_view` - Cross-project phase comparison

## Dependencies

- REQ-HIER-001: Vendored .rtmx folder discovery (must discover projects first)

## Blocks

None

## Effort

3.5 weeks
