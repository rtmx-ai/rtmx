# REQ-GO-076: PR-Based Requirement Proposals

## Metadata
- **Category**: SYNC
- **Subcategory**: CrossRepo
- **Priority**: HIGH
- **Phase**: 17
- **Status**: MISSING
- **Dependencies**: REQ-GO-075
- **Blocks**: REQ-GO-077

## Requirement

`rtmx move` and `rtmx clone` shall support a `--branch` flag that creates a Git branch in the target repository and optionally opens a pull request, enabling the target project's maintainers to review and accept incoming requirements.

## Rationale

Directly writing to another project's main branch violates standard code review practices. Requirement transfers should follow the same PR workflow as code changes — proposed via branch, reviewed by maintainers, merged when accepted. This is especially important in regulated environments (FedRAMP, CMMC) where requirement changes require documented approval.

## Design

### Branch Workflow

```
rtmx move REQ-WEB-005 --to /path/to/rtmx --branch rtmx-sync/migrate-web-005
```

1. All steps from REQ-GO-075 execute against a new branch in the target repo
2. Branch created from target's default branch (main/master)
3. CSV row and spec file committed to the branch
4. Commit message references source repo and requirement ID

### PR Creation (optional)

```
rtmx move REQ-WEB-005 --to /path/to/rtmx --branch ... --pr
```

When `--pr` is specified:
1. Push branch to target remote
2. Create PR via GitHub API (`gh pr create`) or GitLab API
3. PR title: `req: Accept REQ-WEB-005 from rtmx-sync`
4. PR body: Requirement description, provenance link, rationale for transfer
5. Labels: `requirement`, `cross-repo`
6. Return PR URL

### Flags

- `--branch NAME` — create branch in target repo (required for PR workflow)
- `--pr` — create pull request after pushing branch
- `--reviewer USER` — request PR review from specific user
- `--label LABEL` — additional PR labels (repeatable)

## Acceptance Criteria

1. `--branch` creates a new branch in target repo
2. Changes committed to branch with descriptive message
3. `--pr` creates a GitHub/GitLab PR with structured body
4. PR body includes requirement description and provenance
5. PR URL returned to caller
6. Error if branch already exists (no silent overwrite)
7. Error if target has no remote configured and `--pr` is specified
8. Works with both local paths and remote URLs

## Test Strategy

- **Test Module**: `internal/sync/crossrepo_test.go`
- **Test Function**: `TestMoveToBranch`, `TestMoveWithPR`
- **Validation Method**: Integration Test
