# REQ-REL-003: Release Notes with Verification Instructions

## Metadata
- **Category**: RELEASE
- **Subcategory**: Documentation
- **Priority**: HIGH
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-REL-001
- **Blocks**: REQ-GO-047

## Requirement

Release notes shall include curated highlights, verification instructions, breaking changes, known issues, and upgrade path from Python CLI.

## Rationale

Current release notes are auto-generated commit logs. Professional releases need user-facing documentation that helps adopters install, verify, and upgrade safely.

## Design

### Release Note Template

```markdown
## What's New
- [Summary of user-facing changes]

## Installation

### Verify Your Download
gpg --verify checksums.txt.sig checksums.txt
sha256sum -c <(grep linux_amd64 checksums.txt)

### Install
[Platform-specific instructions]

## Breaking Changes
[None / List]

## Upgrade from Python CLI
pip install --upgrade rtmx  # Minimal pytest plugin
brew install rtmx-ai/tap/rtmx  # Go binary

## Known Issues
[List]

## Security
[CVEs addressed, if any]
```

### Implementation

Add `release.header` and `release.footer` to `.goreleaser.yaml`, or use a `release_notes.md` template file.

## Acceptance Criteria

1. Release notes include verification instructions
2. Release notes include user-facing highlights (not just commit log)
3. Breaking changes documented (or explicitly stated as none)
4. Upgrade path from Python CLI documented
5. Known issues listed

## Files to Modify

- `.goreleaser.yaml` - Add release note template
- `scripts/release-notes.sh` or similar - Template generation
