# REQ-MIG-002: Trunk Migration to Go

## Metadata
- **Category**: MIGRATION
- **Subcategory**: Trunk
- **Priority**: CRITICAL
- **Phase**: 21
- **Status**: MISSING
- **Effort**: 4 weeks
- **Dependencies**: REQ-MIG-001, REQ-GO-045, REQ-GO-046
- **Blocks**: REQ-MIG-003

## Requirement

The rtmx-ai/rtmx main branch shall migrate to track the Go implementation. Python trunk archived to legacy/python branch.

## Design

### Pre-migration (in rtmx-go repo)
1. Rename Go module path: `github.com/rtmx-ai/rtmx` -> `github.com/rtmx-ai/rtmx`
2. Update all import paths, ldflags, install scripts, README, GoReleaser config
3. Run full test suite, verify CI passes
4. Tag final rtmx-go release

### Migration (in rtmx repo)
1. Create `legacy/python` branch from current main
2. Tag Python CLI final release with deprecation notice
3. Push Go codebase to `go-migration` branch
4. Open PR for audit trail, merge to main
5. Transfer secrets (GPG, Homebrew, Scoop, GitHub App)
6. Tag v1.0.0 to trigger first release from new main

### Post-migration
1. Archive rtmx-go repo (read-only)
2. Update rtmx-go README to redirect to rtmx

## Acceptance Criteria

1. Go codebase is on rtmx-ai/rtmx main branch
2. Python code preserved on legacy/python branch
3. CI pipeline passes on new main
4. Tag release produces signed binaries via GoReleaser
5. Homebrew/Scoop formulas update correctly
6. rtmx-go repo archived with redirect notice
