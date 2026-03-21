# REQ-CI-003: Marker Compliance Gate

## Metadata
- **Category**: CI
- **Subcategory**: Quality
- **Priority**: HIGH
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-017
- **Blocks**: REQ-GO-047

## Requirement

CI pipeline shall enforce that at least 80% of test functions have requirement markers (`rtmx.Req()`), failing the build if compliance drops below threshold.

## Rationale

The Python CI (ci.yml job `marker-compliance`) counts tests with `@pytest.mark.req` markers and fails if coverage is below 80%. This prevents tests from being written without requirement linkage, maintaining the closed-loop traceability guarantee.

## Design

### Workflow Addition (ci.yml)

New job `marker-compliance`:

```yaml
marker-compliance:
  name: Marker Compliance
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v5
    - uses: actions/setup-go@v6
      with:
        go-version: '1.22'
    - name: Check marker compliance
      run: |
        TOTAL=$(go test -list '.*' ./... 2>/dev/null | grep -c '^Test')
        MARKED=$(grep -r 'rtmx\.Req(t,' internal/ test/ pkg/ --include='*_test.go' -l | wc -l)
        PCT=$((MARKED * 100 / TOTAL))
        echo "Total test files: $TOTAL"
        echo "Files with rtmx.Req: $MARKED"
        echo "Compliance: ${PCT}%"
        if [ "$PCT" -lt 80 ]; then
          echo "::error::Marker compliance ${PCT}% is below 80% threshold"
          exit 1
        fi
```

### Alternative: Use `rtmx from-tests`

```bash
rtmx from-tests --show-missing
# Fail if too many tests lack markers
```

## Acceptance Criteria

1. CI job counts test files with `rtmx.Req()` markers
2. Build fails if marker coverage < 80%
3. Compliance percentage visible in CI output
4. Works with Go test naming conventions

## Files to Modify

- `.github/workflows/ci.yml` - Add `marker-compliance` job
