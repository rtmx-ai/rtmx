# REQ-HIER-001: Vendored .rtmx Folder Discovery

## Status: MISSING
## Priority: HIGH
## Phase: 22

## Description

The `rtmx` CLI shall automatically discover `.rtmx/` folders in vendored subdirectories and present them as part of a hierarchical requirements structure. When run from a parent directory (e.g., the monorepo root), `rtmx` shall walk the directory tree, discover child `.rtmx/` databases, and build an aggregated view of the entire platform's requirements.

## Rationale

In a monorepo with multiple vendored projects, each project maintains its own `.rtmx/` folder with independent requirements. Engineers need a unified view across the platform without manually aggregating databases. This enables platform-level status tracking, cross-project dependency validation, and hierarchical reporting.

## Acceptance Criteria

- [ ] `rtmx` walks directory tree from CWD to discover `.rtmx/database.csv` files in subdirectories
- [ ] Discovery respects `.rtmxignore` patterns (similar to `.gitignore`) to exclude directories
- [ ] Each discovered database is loaded with a namespace prefix derived from its relative path
- [ ] `rtmx status` at monorepo root shows aggregated status across all discovered projects
- [ ] `rtmx status --project rtmx` filters to a single vendored project
- [ ] Discovery depth is configurable via `rtmx.yaml` (`discovery.max_depth`, default: 3)
- [ ] Performance: discovery completes in under 500ms for up to 20 vendored projects
- [ ] Discovered projects listed via `rtmx projects` command
- [ ] Configuration in parent `rtmx.yaml`:
  ```yaml
  rtmx:
    discovery:
      enabled: true
      max_depth: 3
      ignore:
        - node_modules
        - .git
        - vendor
  ```

## Technical Design

### Discovery Algorithm

1. Walk directory tree from CWD up to `max_depth`
2. At each level, check for `.rtmx/database.csv`
3. Skip directories matching `.rtmxignore` or `discovery.ignore` patterns
4. Load each discovered database with namespace = relative path
5. Cache discovery results in `.rtmx/.discovery-cache.json` (invalidate on mtime change)

### Namespace Format

Requirements from vendored projects are namespaced:
- `rtmx/REQ-DIST-002` (from `rtmx/.rtmx/database.csv`)
- `rtmx-sync/REQ-SYNC-001` (from `rtmx-sync/.rtmx/database.csv`)
- `aegis-cli/REQ-AEGIS-001` (from `aegis-cli/.rtmx/database.csv`)

Local (root) requirements have no namespace prefix.

## Test Cases

1. `tests/test_discovery.py::test_discovers_nested_rtmx_dirs` - Finds .rtmx/ in subdirectories
2. `tests/test_discovery.py::test_respects_rtmxignore` - Skips ignored directories
3. `tests/test_discovery.py::test_namespace_prefix` - Requirements namespaced by path
4. `tests/test_discovery.py::test_aggregated_status` - Status aggregates all projects
5. `tests/test_discovery.py::test_project_filter` - `--project` flag filters correctly
6. `tests/test_discovery.py::test_max_depth_config` - Depth limit respected
7. `tests/test_discovery.py::test_discovery_performance` - Completes under 500ms
8. `tests/test_discovery.py::test_projects_command` - Lists discovered projects

## Dependencies

- REQ-MONO-001: Monorepo structure (provides the vendored project layout)

## Blocks

- REQ-HIER-002: Aggregated hierarchical RTM
- REQ-HIER-003: Cross-project dependency tracking

## Effort

3.0 weeks
