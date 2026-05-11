# REQ-LAUNCH-002: Show HN Launch Readiness

## Metadata
- **Category**: LAUNCH
- **Subcategory**: ShowHN
- **Priority**: HIGH
- **Phase**: 25
- **Status**: MISSING
- **Effort**: 0.5 weeks
- **Dependencies**: REQ-REL-007 (v1.0.0 tagged), REQ-DIST-006 (homebrew-core), REQ-DIST-007 (apt repo), REQ-LAUNCH-001 (README)

## Requirement

RTMX launch checklist shall be complete before the Show HN post. All
install methods shall work on a clean machine, the README shall be
current, and the blog post shall be published.

## Acceptance Criteria

1. `brew install rtmx` works (homebrew-core or tap fallback)
2. `apt install rtmx` works (apt repo set up)
3. `scoop install rtmx` works (existing Scoop bucket)
4. `go install github.com/rtmx-ai/rtmx/cmd/rtmx@v1.0.0` works
5. GitHub releases page has v1.0.0 with signed artifacts
6. README has correct version, requirement count, and install commands
7. rtmx.ai/blog/show-hn-rtmx is published and accessible
8. `rtmx init && rtmx setup && rtmx status` works on a fresh repo

## Verification Test

Inspection test validates all install commands are documented, release
exists, and README is consistent with the current state.

## Notes

- Show HN post should link to GitHub repo, not rtmx.ai
- Title format: "Show HN: RTMX -- track requirements as CSV in git, verified by tests"
- First comment should explain the motivation (AI agents + traceability)
