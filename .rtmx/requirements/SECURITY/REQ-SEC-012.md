# REQ-SEC-012: Security Posture Check Command

## Metadata
- **Category**: SECURITY
- **Subcategory**: Governance
- **Priority**: HIGH
- **Phase**: 20
- **Status**: MISSING
- **Effort**: 2 weeks
- **Dependencies**: REQ-SEC-004, REQ-SEC-010, REQ-SEC-011

## Requirement

`rtmx security` shall audit the security posture of the repository and RTMX configuration, reporting PASS/WARN/FAIL for each control and optionally applying recommended fixes.

## Rationale

`rtmx health` checks internal RTM data consistency. `rtmx security` checks the external controls that protect the RTM from tampering. Together they give a complete picture of project integrity.

The command should work with GitHub (MVP), with extensibility to GitLab, Gitea, and BitBucket. Platform detection is automatic via git remote URL parsing.

## Design

### Check Categories

**Repository Controls (platform-specific, GitHub MVP):**

| Check | Method | Fix |
|-------|--------|-----|
| Branch protection on default branch | `gh api repos/{owner}/{repo}/rules` | Enable via API |
| Signed commit requirement | `gh api repos/{owner}/{repo}/rules` | Enable via API |
| Secret scanning enabled | `gh api repos/{owner}/{repo}` | Enable via API |
| Dependabot enabled | Check `.github/dependabot.yml` exists | Create file |
| CODEOWNERS for `.rtmx/` | Check `CODEOWNERS` or `.github/CODEOWNERS` | Create file |
| Vulnerability alerts enabled | `gh api repos/{owner}/{repo}/vulnerability-alerts` | Enable via API |

**RTMX Controls (platform-independent):**

| Check | Method | Fix |
|-------|--------|-----|
| Verify thresholds configured | Read `rtmx.yaml` verify.thresholds | Set defaults |
| CI verify-requirements job exists | Parse workflow files | Create job |
| GitHub Actions pinned to SHAs | Scan workflow `uses:` directives | Pin SHAs |
| GPG signing in release workflow | Parse release workflow | Remove conditional |
| Install script GPG verification | Parse `scripts/install.sh` | Add verification |
| Trusted peers configured | Read `rtmx.yaml` sync.trusted_peers | Generate keypair |
| Grants configured | Read `rtmx.yaml` sync.grants | Prompt setup |
| Source attestation in results | Read `rtmx.yaml` verify.require_attestation | Enable |

### CLI Interface

```
rtmx security                    # Run all checks, report PASS/WARN/FAIL
rtmx security --json             # Machine-readable output
rtmx security --fix              # Interactive: prompt to apply fixes
rtmx security --fix --yes        # Non-interactive: apply all recommended fixes
rtmx security --platform github  # Override platform detection
```

### Output Format

```
RTMX Security Posture Check

Repository Controls (GitHub: rtmx-ai/rtmx)
  [PASS] Branch protection enabled on main
  [WARN] Signed commits not required
  [PASS] Secret scanning enabled
  [PASS] Dependabot configured
  [WARN] CODEOWNERS missing for .rtmx/ paths
  [PASS] Vulnerability alerts enabled

RTMX Controls
  [PASS] Verify thresholds configured (warn=5, fail=15)
  [PASS] CI verify-requirements job exists
  [PASS] GitHub Actions pinned to SHAs (41/41)
  [PASS] GPG signing mandatory in release workflow
  [PASS] Install script verifies GPG signatures
  [WARN] No trusted peers configured
  [WARN] No grants configured

Score: 11/14 passed, 3 warnings, 0 failures

Apply recommended fixes? [a]ll / [s]elect / [n]one:
```

### Exit Codes

- 0: All checks pass or only warnings
- 1: Any check fails (FAIL, not WARN)
- `--strict` flag: treats warnings as failures

### JSON Output

```json
{
  "platform": "github",
  "repository": "rtmx-ai/rtmx",
  "checks": [
    {
      "category": "repository",
      "name": "branch_protection",
      "status": "PASS",
      "message": "Branch protection enabled on main",
      "fixable": false
    },
    {
      "category": "rtmx",
      "name": "verify_thresholds",
      "status": "PASS",
      "message": "Thresholds configured (warn=5, fail=15)",
      "fixable": true,
      "fix_description": "Set verify.thresholds in rtmx.yaml"
    }
  ],
  "summary": {
    "passed": 11,
    "warnings": 3,
    "failed": 0,
    "score_percent": 78.6
  }
}
```

### Platform Extensibility

```go
type PlatformChecker interface {
    DetectPlatform(remoteURL string) bool
    CheckBranchProtection(ctx context.Context) CheckResult
    CheckSignedCommits(ctx context.Context) CheckResult
    CheckSecretScanning(ctx context.Context) CheckResult
    FixBranchProtection(ctx context.Context) error
    // ...
}
```

GitHub implementation uses `gh` CLI or GitHub API. Future implementations for GitLab, Gitea, BitBucket implement the same interface.

### Files to Create

- `internal/cmd/security.go` -- Command and RTMX checks
- `internal/cmd/security_test.go` -- Tests
- `internal/adapters/platform_github.go` -- GitHub platform checker (MVP)

### Files to Modify

- `internal/cmd/root.go` -- Register security command

## Acceptance Criteria

1. `rtmx security` runs all checks and reports PASS/WARN/FAIL per check
2. `rtmx security --json` produces machine-readable output
3. `rtmx security --fix` interactively offers to apply fixes
4. `rtmx security --fix --yes` applies all fixable checks non-interactively
5. GitHub platform detected automatically from git remote URL
6. Exit code 1 on any FAIL check
7. `--strict` treats warnings as failures
8. RTMX controls checked without platform API access (local file parsing)
9. Repository controls degrade gracefully when API unavailable (skip with warning)

## Test Strategy

- **Test Module**: `internal/cmd/security_test.go`
- **Test Function**: `TestSecurityCommand`
- **Validation Method**: Integration Test

### Test Cases

1. RTMX controls checked against temp project with various configs
2. JSON output validates against expected schema
3. --strict flag changes exit code behavior
4. Missing platform API gracefully skips repository checks
5. --fix creates CODEOWNERS and sets thresholds
