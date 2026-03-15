# REQ-MIG-002: Main Branch Migration to rtmx-go

## Status: MISSING
## Priority: CRITICAL
## Phase: 21

## Description

The `rtmx` repository main branch shall be migrated to track the `rtmx-go` Go implementation, deprecating the original Python trunk. The Python codebase shall be archived to a `legacy/python` branch and the Go codebase from `rtmx-go` shall become the new main branch content. The `rtmx-go` repository shall be archived after the migration is complete.

## Rationale

Maintaining two implementations creates unsustainable development overhead. Consolidating to a single Go-based implementation in the canonical `rtmx` repository simplifies the development workflow, reduces confusion about which repository is authoritative, and aligns with the platform's distribution strategy (single binary, zero dependencies).

## Acceptance Criteria

- [ ] Python codebase preserved on `legacy/python` branch with full git history
- [ ] Go codebase from rtmx-go merged into rtmx main branch preserving git history
- [ ] All CI/CD pipelines updated for Go build (goreleaser, GitHub Actions)
- [ ] PyPI package `rtmx` updated to minimal pytest plugin only (REQ-PYTEST-001)
- [ ] Go module path set to `github.com/rtmx-ai/rtmx` (not rtmx-go)
- [ ] All documentation updated to reflect Go as primary implementation
- [ ] CHANGELOG.md documents the migration with clear version boundary
- [ ] rtmx-go repository archived with pointer to rtmx
- [ ] rtmx.ai submodule reference updated to track new main branch
- [ ] rtmx-sync dependency updated from Python rtmx to Go rtmx (or vendored)
- [ ] GitHub releases continue from existing version sequence (no version reset)
- [ ] Migration guide published for users transitioning from Python CLI

## Migration Procedure

1. Create `legacy/python` branch from current main
2. Merge rtmx-go history into rtmx using `git merge --allow-unrelated-histories`
3. Resolve any path conflicts (README, Makefile, CI)
4. Update go.mod module path to `github.com/rtmx-ai/rtmx`
5. Update all CI/CD workflows for Go
6. Tag migration release (e.g., v1.0.0-go)
7. Archive rtmx-go repository
8. Update downstream dependencies (rtmx-sync, rtmx.ai)

## Test Cases

1. `tests/test_migration_go.py::test_legacy_branch_preserves_history` - Python history preserved
2. `tests/test_migration_go.py::test_go_module_path_correct` - Module path is github.com/rtmx-ai/rtmx
3. `tests/test_migration_go.py::test_ci_pipelines_pass` - All CI checks green post-migration
4. `tests/test_migration_go.py::test_pypi_minimal_plugin` - PyPI package is pytest plugin only
5. `tests/test_migration_go.py::test_downstream_deps_updated` - rtmx-sync builds against new rtmx

## Dependencies

- REQ-MIG-001: Feature parity validation (must pass before migration)
- REQ-PYTEST-001: Minimal pytest plugin ready for standalone release

## Blocks

- REQ-MIG-003: Legacy Python package deprecation
- REQ-MONO-001: Monorepo structure (needs Go-based rtmx as input)

## Effort

4.0 weeks
