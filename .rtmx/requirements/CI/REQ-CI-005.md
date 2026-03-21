# REQ-CI-005: Security Scanning

## Metadata
- **Category**: CI
- **Subcategory**: Security
- **Priority**: MEDIUM
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: (none)

## Requirement

CI pipeline shall include vulnerability scanning for Go dependencies and CodeQL analysis for security issues.

## Rationale

The Python CI includes pip-audit for dependency vulnerabilities and CodeQL for code security analysis. The Go CI has neither, creating a security gap.

## Design

### Dependency Scanning

```yaml
security:
  name: Security
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v5
    - uses: actions/setup-go@v6
      with:
        go-version: '1.22'
    - name: Install govulncheck
      run: go install golang.org/x/vuln/cmd/govulncheck@latest
    - name: Run govulncheck
      run: govulncheck ./...
```

### CodeQL Analysis

```yaml
codeql:
  name: CodeQL
  runs-on: ubuntu-latest
  permissions:
    security-events: write
  steps:
    - uses: actions/checkout@v5
    - uses: github/codeql-action/init@v3
      with:
        languages: go
    - uses: github/codeql-action/autobuild@v3
    - uses: github/codeql-action/analyze@v3
```

## Acceptance Criteria

1. `govulncheck` runs on every CI run
2. Build fails on known vulnerabilities in dependencies
3. CodeQL analysis runs and reports to Security tab
4. Weekly scheduled scan (in addition to push/PR triggers)

## Files to Modify

- `.github/workflows/ci.yml` - Add `security` and `codeql` jobs
