# REQ-REL-007: v1.0.0 Release

## Metadata
- **Category**: RELEASE
- **Subcategory**: v1.0.0
- **Priority**: P0
- **Phase**: 25
- **Status**: MISSING
- **Effort**: 0.5 weeks
- **Dependencies**: REQ-GO-047 (production readiness gate)
- **Blocks**: REQ-DIST-006, REQ-DIST-007, REQ-LAUNCH-002

## Requirement

RTMX shall tag and publish v1.0.0, signaling API stability per the
versioning policy in CLAUDE.md. The release gate (`rtmx release gate v1.0.0`)
must exit 0 before tagging. All 226 requirements must be COMPLETE.

## Rationale

v1.0 is the prerequisite for mainstream package submissions. Homebrew-core
and Debian both favor stable-versioned projects. The 100% completion
milestone and 42-command surface make this the right moment.

## Acceptance Criteria

1. `rtmx release gate v1.0.0` exits 0
2. All requirements are COMPLETE (0 PARTIAL, 0 MISSING)
3. CHANGELOG.md documents the v1.0.0 release
4. `git tag -s v1.0.0` is created with GPG signature
5. GoReleaser produces all artifacts (binaries, .deb, .rpm, Docker, Homebrew formula, Scoop manifest)
6. GitHub release page has install instructions and verification commands

## Pre-Tag Checklist

- [ ] All tests pass (`make test`)
- [ ] Lint clean (`make lint`)
- [ ] `rtmx verify --update` shows 0 changes
- [ ] `rtmx release gate v1.0.0` exits 0
- [ ] CHANGELOG.md updated
- [ ] README.md current (226 requirements, 100%)

## Files to Modify

- `CHANGELOG.md` -- v1.0.0 entry
- `.rtmx/database.csv` -- assign v1.0.0 to requirements
