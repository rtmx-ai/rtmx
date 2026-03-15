# REQ-MONO-001: Monorepo Structure and Vendoring

## Status: MISSING
## Priority: HIGH
## Phase: 22

## Description

RTMX shall establish a monorepo structure that vendors the Go-based `rtmx` CLI client, `rtmx-sync` real-time coordination service, `rtmx.ai` website, and `aegis-cli` compliance client. Each vendored project shall maintain its own `.rtmx/` directory for independent requirements traceability while being developed, tested, and released as one coherent platform.

## Rationale

The RTMX platform currently spans four repositories with independent release cycles, duplicated CI, and cross-repo dependency friction. A monorepo pattern enables:
- Atomic cross-project changes (API + client + docs in one PR)
- Unified CI/CD pipeline with shared test infrastructure
- Simplified dependency management (no version pinning across repos)
- Coherent platform releases
- Continued open-source contribution to rtmx and aegis-cli independently

## Acceptance Criteria

- [ ] Monorepo directory structure established:
  ```
  rtmx-platform/
  ├── rtmx/              # Go CLI client (open source, vendored)
  ├── rtmx-sync/         # Real-time coordination (proprietary)
  ├── website/           # rtmx.ai Astro site
  ├── aegis-cli/         # Compliance CLI (open source, vendored)
  ├── .rtmx/             # Platform-level RTM (aggregates children)
  ├── Makefile            # Unified build targets
  ├── go.work             # Go workspace for multi-module builds
  └── .github/workflows/  # Unified CI/CD
  ```
- [ ] Each vendored project retains its own `.rtmx/` folder with independent RTM database
- [ ] Platform-level `.rtmx/` aggregates requirements from all vendored projects
- [ ] `go.work` enables cross-module development without publishing intermediate versions
- [ ] Open-source projects (rtmx, aegis-cli) can be independently cloned and developed
- [ ] Proprietary components (rtmx-sync, website) are monorepo-only
- [ ] Git subtree or similar mechanism enables bidirectional sync with upstream open-source repos
- [ ] Unified `make` targets: `make build-all`, `make test-all`, `make release`

## Open Source Strategy

| Project | License | Monorepo Path | Upstream Repo | Sync Mechanism |
|---------|---------|---------------|---------------|----------------|
| rtmx | Apache-2.0 | `rtmx/` | github.com/rtmx-ai/rtmx | git subtree |
| aegis-cli | Apache-2.0 | `aegis-cli/` | github.com/rtmx-ai/aegis-cli | git subtree |
| rtmx-sync | Proprietary | `rtmx-sync/` | N/A (monorepo only) | N/A |
| website | Proprietary | `website/` | N/A (monorepo only) | N/A |

## Test Cases

1. `tests/test_monorepo.py::test_directory_structure` - All vendored projects present
2. `tests/test_monorepo.py::test_go_workspace_builds` - `go.work` multi-module build succeeds
3. `tests/test_monorepo.py::test_independent_rtmx_dirs` - Each project has own .rtmx/
4. `tests/test_monorepo.py::test_platform_aggregation` - Platform RTM aggregates all projects
5. `tests/test_monorepo.py::test_subtree_sync` - Open source subtree push/pull works
6. `tests/test_monorepo.py::test_unified_ci` - CI runs all project tests

## Dependencies

- REQ-MIG-002: Main branch migration (rtmx must be Go-based before vendoring)

## Blocks

- REQ-MONO-002: Unified build and release pipeline
- REQ-MONO-003: Open source / proprietary boundary enforcement
- REQ-HIER-001: Vendored .rtmx folder discovery

## Effort

5.0 weeks
