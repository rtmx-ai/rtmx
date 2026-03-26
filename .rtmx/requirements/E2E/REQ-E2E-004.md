# REQ-E2E-004: Install Script Verification

## Metadata
- **Category**: E2E
- **Subcategory**: Install
- **Priority**: MEDIUM
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-REL-006

## Requirement

CI shall test `scripts/install.sh` in a clean environment to verify it downloads, verifies checksums, and installs the binary correctly.

## Design

### CI Job

```yaml
install-script-test:
  name: Install Script Test
  needs: [test]
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v5
    - name: Test install script syntax
      run: bash -n scripts/install.sh
    - name: Test install script with mock
      run: |
        # Verify the script handles missing release gracefully
        RTMX_VERSION=v999.999.999 bash scripts/install.sh 2>&1 && exit 1 || true
        echo "Install script correctly fails for non-existent version"
```

Note: Full install test requires a published release. The syntax and error-handling tests run without network.

## Acceptance Criteria

1. Install script has valid bash syntax
2. Script handles missing versions gracefully (non-zero exit)
3. Script handles unsupported platforms gracefully
4. After release publish: script installs working binary

## Files to Modify

- `.github/workflows/ci.yml` - Add install script validation
