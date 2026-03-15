# REQ-MONO-003: Open Source and Proprietary Boundary Enforcement

## Status: MISSING
## Priority: HIGH
## Phase: 22

## Description

The RTMX monorepo shall enforce clear boundaries between open-source components (rtmx, aegis-cli) and proprietary components (rtmx-sync, website). Code, configuration, and secrets must not leak across boundaries. Git subtree operations must preserve licensing and contribution integrity.

## Rationale

RTMX maintains both open-source tools (Apache-2.0) and proprietary platform services. The monorepo must prevent:
- Proprietary code leaking into open-source upstream pushes
- Open-source components depending on proprietary modules
- Secrets or internal configuration appearing in public repositories
- License contamination between components

## Acceptance Criteria

- [ ] `.github/CODEOWNERS` enforces review requirements per component
- [ ] CI lint step validates no import paths cross the open/proprietary boundary
- [ ] Git subtree push pre-flight check scans for proprietary references before pushing to upstream
- [ ] Each component has explicit LICENSE file matching its licensing model
- [ ] Dependency analysis tool verifies open-source components have no proprietary dependencies
- [ ] Secret scanning configured per-component (open-source components scanned before subtree push)
- [ ] Contributing guide documents boundary rules for external contributors
- [ ] Go module boundaries prevent import of proprietary packages from open-source packages

## Test Cases

1. `tests/test_boundary.py::test_no_cross_boundary_imports` - No proprietary imports in open-source code
2. `tests/test_boundary.py::test_license_files_present` - Each component has correct LICENSE
3. `tests/test_boundary.py::test_subtree_push_clean` - Subtree push contains no proprietary content
4. `tests/test_boundary.py::test_codeowners_coverage` - All paths have owners
5. `tests/test_boundary.py::test_secret_scanning` - No secrets in open-source components

## Dependencies

- REQ-MONO-001: Monorepo structure

## Blocks

None

## Effort

2.0 weeks
