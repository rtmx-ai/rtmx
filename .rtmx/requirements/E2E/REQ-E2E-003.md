# REQ-E2E-003: GoReleaser Snapshot Dry-Run

## Metadata
- **Category**: E2E
- **Subcategory**: Release
- **Priority**: HIGH
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-043

## Requirement

CI shall run `goreleaser release --snapshot --clean` to validate the release configuration before tag push, catching signing, SBOM, and archive errors early.

## Design

### CI Job

```yaml
release-dry-run:
  name: Release Dry Run
  needs: [test, lint]
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v5
      with:
        fetch-depth: 0
    - uses: actions/setup-go@v6
      with:
        go-version: '1.25'
    - name: Install syft
      run: curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin
    - name: GoReleaser snapshot
      uses: goreleaser/goreleaser-action@v6
      with:
        distribution: goreleaser
        version: '~> v2'
        args: release --snapshot --clean --skip=sign,publish
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    - name: Verify artifacts
      run: |
        ls -la dist/
        test -f dist/checksums.txt
        test -f dist/rtmx-go_*_linux_amd64.tar.gz
        test -f dist/rtmx_*_linux_amd64.deb
```

## Acceptance Criteria

1. GoReleaser snapshot builds complete without error
2. All expected archives produced (tar.gz, zip, deb, rpm)
3. Checksums generated
4. SBOM generated (when configured)
5. Runs on every push to main (catches config drift early)

## Files to Modify

- `.github/workflows/ci.yml` - Add `release-dry-run` job
