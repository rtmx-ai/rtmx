# REQ-CI-001: Closed-Loop Verification in CI

## Metadata
- **Category**: CI
- **Subcategory**: Verify
- **Priority**: P0
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-018
- **Blocks**: REQ-GO-047

## Requirement

CI pipeline shall run `rtmx verify --update` on main branch pushes and auto-commit RTM database status changes with signed commits.

## Rationale

The Python CI pipeline (ci.yml job `verify-requirements`) runs `rtmx verify --update` after tests pass, then auto-commits changes to `docs/rtm_database.csv`. This ensures requirement status is always derived from test results, never from manual edits. The Go CI currently has no equivalent.

## Design

### Workflow Addition (ci.yml)

New job `verify-requirements` that:
1. Depends on `test` and `lint` jobs passing
2. Runs only on `push` to `main` (not PRs)
3. Uses GitHub App token for signed commits
4. Runs `rtmx verify --update --verbose`
5. If `.rtmx/database.csv` changed, commits via GitHub API

```yaml
verify-requirements:
  name: Verify Requirements
  needs: [test, lint]
  if: github.event_name == 'push' && github.ref == 'refs/heads/main'
  runs-on: ubuntu-latest
  permissions:
    contents: write
  steps:
    - uses: actions/create-github-app-token@v1
      id: app-token
      with:
        app-id: ${{ secrets.APP_ID }}
        private-key: ${{ secrets.APP_PRIVATE_KEY }}
    - uses: actions/checkout@v5
      with:
        token: ${{ steps.app-token.outputs.token }}
    - uses: actions/setup-go@v6
      with:
        go-version: '1.22'
    - run: go build -o rtmx ./cmd/rtmx
    - run: ./rtmx verify --update --verbose
    - run: ./rtmx status
    - name: Commit RTM updates
      run: |
        git diff --quiet .rtmx/database.csv && exit 0
        git add .rtmx/database.csv
        git commit -m "ci: Auto-update RTM status from test results"
        git push
```

## Infrastructure Required

- GitHub App for signed commits (same as Python repo: `APP_ID`, `APP_PRIVATE_KEY`)

## Acceptance Criteria

1. `rtmx verify --update` runs on every main push after tests pass
2. Changed `.rtmx/database.csv` is auto-committed
3. Commits are signed (GitHub App token)
4. Job skips on PRs (verify only, no commit on PRs)
5. Pipeline status visible in GitHub Actions

## Files to Create/Modify

- `.github/workflows/ci.yml` - Add `verify-requirements` job
