# REQ-E2E-002: Cross-Language Verification E2E

## Metadata
- **Category**: E2E
- **Subcategory**: CrossLang
- **Priority**: HIGH
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-VERIFY-001, REQ-LANG-004

## Requirement

CI shall run Python tests with `--rtmx-output`, then verify the results with the Go CLI's `rtmx verify --results`, validating the full cross-language dogfood loop.

## Design

### CI Job

```yaml
cross-language-e2e:
  name: Cross-Language E2E
  needs: [test]
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v5
    - uses: actions/setup-go@v6
      with:
        go-version: '1.25'
    - uses: actions/setup-python@v5
      with:
        python-version: '3.12'
    - name: Build Go CLI
      run: go build -o rtmx ./cmd/rtmx
    - name: Install Python rtmx
      run: pip install rtmx
    - name: Create test file
      run: |
        cat > /tmp/test_cross.py << 'EOF'
        import pytest
        @pytest.mark.req("REQ-VERIFY-001")
        def test_cross_lang():
            assert True
        EOF
    - name: Run Python tests with RTMX output
      run: pytest /tmp/test_cross.py --rtmx-output=/tmp/results.json
    - name: Verify with Go CLI
      run: |
        cat /tmp/results.json | python3 -m json.tool
        ./rtmx verify --results /tmp/results.json --dry-run --verbose
```

## Acceptance Criteria

1. Python pytest produces valid RTMX results JSON
2. Go CLI parses and validates the Python output
3. Cross-language pipeline runs in CI on every push
4. Failure in either direction fails the CI job

## Files to Modify

- `.github/workflows/ci.yml` - Add `cross-language-e2e` job
