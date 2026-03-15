# REQ-MONO-002: Unified Build and Release Pipeline

## Status: MISSING
## Priority: HIGH
## Phase: 22

## Description

The RTMX monorepo shall have a unified CI/CD pipeline that builds, tests, and releases all platform components with proper dependency ordering. The pipeline shall support independent component releases as well as coordinated platform releases.

## Rationale

Without a unified pipeline, cross-project changes require coordinating releases across multiple repositories and CI systems. A unified pipeline ensures:
- Breaking changes are caught before they reach any component
- Platform releases are atomic and tested end-to-end
- Open-source components can still be released independently
- Release automation reduces manual coordination overhead

## Acceptance Criteria

- [ ] GitHub Actions workflow builds all components in dependency order
- [ ] Component-level CI triggers on path-specific changes (e.g., `rtmx/**` triggers rtmx tests only)
- [ ] Platform-level CI triggers on cross-cutting changes or release tags
- [ ] Release matrix:
  - `rtmx` → GoReleaser binaries + PyPI minimal plugin
  - `rtmx-sync` → Docker container + Helm chart
  - `website` → Vercel/Cloudflare deployment
  - `aegis-cli` → GoReleaser binaries
- [ ] Coordinated platform release tag (e.g., `platform-v1.0.0`) triggers all component releases
- [ ] Component-specific release tags (e.g., `rtmx-v1.2.0`) trigger single component release
- [ ] Shared test infrastructure (fixtures, helpers) available to all components
- [ ] Cross-component integration tests run on platform CI
- [ ] Release changelog aggregates component changes

## Test Cases

1. `tests/test_pipeline.py::test_path_specific_triggers` - Component CI triggers correctly
2. `tests/test_pipeline.py::test_dependency_order` - Build order respects dependencies
3. `tests/test_pipeline.py::test_platform_release` - Coordinated release works
4. `tests/test_pipeline.py::test_component_release` - Independent release works
5. `tests/test_pipeline.py::test_cross_component_integration` - Integration tests pass

## Dependencies

- REQ-MONO-001: Monorepo structure (must exist before pipeline)

## Blocks

None

## Effort

3.5 weeks
