# REQ-HIER-003: Cross-Project Dependency Tracking

## Status: MISSING
## Priority: HIGH
## Phase: 22

## Description

The `rtmx` CLI shall support cross-project requirement dependencies within the monorepo, enabling requirements in one vendored project to declare dependencies on or block requirements in another vendored project. The dependency graph shall span project boundaries for critical path analysis and blocking detection.

## Rationale

In a system of systems, components have real dependencies on each other. For example, `rtmx-sync` cannot implement real-time collaboration without the `rtmx` CLI providing the database format it syncs. Cross-project dependencies must be:
- Explicitly declared and validated
- Visible in the platform-level dependency graph
- Factored into critical path calculations
- Checked during `rtmx validate` for missing or circular dependencies

## Acceptance Criteria

- [ ] Cross-project dependency syntax in RTM database:
  ```
  rtmx/REQ-DIST-002    # Depends on requirement in rtmx project
  aegis-cli/REQ-AEGIS-001  # Depends on requirement in aegis-cli project
  ```
- [ ] `rtmx validate` detects broken cross-project references (missing requirement, missing project)
- [ ] `rtmx graph` renders cross-project edges with distinct styling (dashed lines, different color)
- [ ] Critical path analysis considers cross-project dependencies
- [ ] `rtmx backlog` marks requirements as blocked when cross-project dependency is incomplete
- [ ] Circular dependency detection spans project boundaries
- [ ] `rtmx validate --cross-project` specifically validates all cross-project references
- [ ] Graceful degradation when a referenced project's `.rtmx/` is unavailable (warning, not error)
- [ ] Builds on existing cross-repo reference format from REQ-COLLAB-001 but uses local discovery instead of remote configuration

## Relationship to REQ-COLLAB-001

REQ-COLLAB-001 established cross-**repository** dependencies using remote configuration (`rtmx remote add`). This requirement extends that concept for **local vendored projects** discovered by REQ-HIER-001. The reference format is compatible:

| Context | Reference Format | Resolution |
|---------|-----------------|------------|
| Cross-repo (remote) | `sync:REQ-SYNC-001` | Remote config lookup |
| Cross-project (local) | `rtmx-sync/REQ-SYNC-001` | Discovery-based lookup |

When a project is both a configured remote and a discovered local project, local discovery takes precedence.

## Test Cases

1. `tests/test_cross_project.py::test_cross_project_dependency_syntax` - Reference parsing
2. `tests/test_cross_project.py::test_broken_reference_detection` - Missing project/requirement
3. `tests/test_cross_project.py::test_cross_project_graph_edges` - Graph visualization
4. `tests/test_cross_project.py::test_critical_path_cross_project` - Critical path spans projects
5. `tests/test_cross_project.py::test_blocked_by_cross_project` - Blocking detection
6. `tests/test_cross_project.py::test_circular_dependency_cross_project` - Cycle detection
7. `tests/test_cross_project.py::test_unavailable_project_graceful` - Graceful degradation
8. `tests/test_cross_project.py::test_local_precedence_over_remote` - Discovery beats remote config

## Dependencies

- REQ-HIER-001: Vendored .rtmx folder discovery (must discover projects to resolve references)
- REQ-COLLAB-001: Cross-repository dependency references (foundational reference format)

## Blocks

None

## Effort

2.5 weeks
