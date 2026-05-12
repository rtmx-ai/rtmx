# REQ-DIST-008: Homebrew-core PR Submission and Acceptance

## Metadata
- **Category**: DIST
- **Subcategory**: Homebrew
- **Priority**: HIGH
- **Phase**: 26
- **Status**: MISSING
- **Effort**: 1 week
- **Dependencies**: REQ-DIST-006 (formula exists and passes audit)
- **Blocks**: REQ-LAUNCH-002

## Requirement

RTMX formula shall be submitted to Homebrew/homebrew-core via PR and pass
all CI checks including `brew audit --strict --new`, `brew install --build-from-source`,
and `brew test`. The PR must follow homebrew-core contribution guidelines.

## Rationale

REQ-DIST-006 created the formula and validated it locally. This requirement
covers the actual submission: forking homebrew-core, creating the PR, passing
their CI, and getting it merged.

## Submission Workflow

1. Fork Homebrew/homebrew-core on GitHub
2. Create branch with formula at Formula/r/rtmx.rb (homebrew-core uses first-letter subdirs)
3. Run `brew audit --strict --new rtmx` locally
4. Run `brew install --build-from-source rtmx` from the local branch
5. Run `brew test rtmx` to verify test stanza
6. Submit PR with description following homebrew-core template
7. Address reviewer feedback if any
8. Formula merged and available via `brew update && brew install rtmx`

## Acceptance Criteria

1. PR submitted to Homebrew/homebrew-core
2. homebrew-core CI passes (audit, install, test)
3. Formula builds rtmx from source using Go toolchain
4. `brew install rtmx` works on macOS without custom tap after merge
5. GoReleaser `brew bump-formula-pr` automation configured for future releases

## Verification

Test validates formula structure and local audit. Actual PR acceptance
verified by inspection (merge into homebrew-core).

## Files to Modify

- `Formula/rtmx.rb` -- ensure formula matches homebrew-core conventions
- `.goreleaser.yaml` -- configure `brew bump-formula-pr` for homebrew-core

## Notes

- homebrew-core requires the project to be "notable" (stars, users, press)
- Show HN launch (REQ-LAUNCH-001) provides the notability signal
- Custom tap (rtmx-ai/homebrew-tap) remains as fallback
- After merge, updates go through `brew bump-formula-pr` or Homebrew bot
