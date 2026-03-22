# REQ-E2E-001: Binary Smoke Test in CI

## Metadata
- **Category**: E2E
- **Subcategory**: Binary
- **Priority**: P0
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-010, REQ-GO-011, REQ-GO-012
- **Blocks**: REQ-GO-047

## Requirement

CI shall build the rtmx binary and run it against the actual repo's RTM database, verifying that `rtmx version`, `rtmx status`, `rtmx backlog`, `rtmx health`, and `rtmx verify` produce expected output and exit codes.

## Rationale

Unit tests use `go test` with in-process command execution. This doesn't catch issues where the compiled binary behaves differently - e.g., missing init() registrations, ldflags injection failures, or file path resolution differences. A smoke test runs the actual binary as a user would.

## Design

### CI Job

```yaml
smoke-test:
  name: Smoke Test
  needs: [test, lint]
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v5
    - uses: actions/setup-go@v6
      with:
        go-version: '1.25'
    - name: Build binary
      run: go build -o rtmx ./cmd/rtmx
    - name: Smoke test
      run: |
        ./rtmx version
        ./rtmx status
        ./rtmx status --json | jq .total
        ./rtmx status --fail-under 50
        ./rtmx backlog --json | jq .total_missing
        ./rtmx health
        ./rtmx health --json | jq .status
        ./rtmx deps | head -20
        ./rtmx cycles
        ./rtmx reconcile
        ./rtmx context
        ./rtmx install --agents --list
```

## Acceptance Criteria

1. CI runs the actual compiled binary (not `go test`)
2. All core commands execute without error against repo's own RTM
3. `--json` output is parseable by `jq`
4. `--fail-under` exits correctly
5. Smoke test runs on every push to main and every PR

## Files to Modify

- `.github/workflows/ci.yml` - Add `smoke-test` job
