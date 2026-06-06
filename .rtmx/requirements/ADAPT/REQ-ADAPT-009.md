# REQ-ADAPT-009: GitLab CI Pipeline Status Integration

## Metadata
- **Category**: ADAPT
- **Subcategory**: GitLab
- **Priority**: MEDIUM
- **Phase**: 27
- **Status**: MISSING
- **Dependencies**: REQ-ADAPT-007, REQ-ADAPT-008
- **Blocks**: (none)

## Requirement

The GitLab adapter shall support reading CI/CD pipeline status for the
configured project and mapping pipeline job results to RTMX requirement
verification status, enabling automated `rtmx verify` from GitLab CI
test results.

## Rationale

Teams using GitLab CI for testing can automatically update requirement
verification status based on pipeline results, closing the loop between
CI and requirements traceability without manual intervention. This is
particularly valuable for air-gapped deployments where the CI system
and RTMX database are in the same network.

## Design

### Pipeline Status Mapping

```yaml
rtmx:
  adapters:
    gitlab:
      ci_integration:
        enabled: true
        branch: "main"
        job_mapping:
          "test-go": "internal/cmd/*_test.go"
          "test-integration": "test/*_test.go"
```

### API Endpoints

| Endpoint | Purpose |
|----------|---------|
| `GET /projects/:id/pipelines` | List recent pipelines |
| `GET /projects/:id/pipelines/:id/jobs` | List jobs in pipeline |
| `GET /projects/:id/jobs/:id/artifacts` | Download test artifacts |

### Verification Flow

```bash
rtmx verify --adapter gitlab    # pull latest pipeline results
```

1. Fetch latest successful pipeline on configured branch
2. For each job in job_mapping, check pass/fail status
3. Download test result artifacts (JUnit XML) if available
4. Map test results to requirements via test_module/test_function matching
5. Update requirement status based on test results

## Acceptance Criteria

1. `rtmx verify --adapter gitlab` reads latest pipeline results.
2. Pipeline job pass/fail maps to requirement verification status.
3. JUnit XML artifacts are parsed for detailed test-to-requirement mapping.
4. Only the latest pipeline on the configured branch is used.
5. Failed pipeline jobs mark linked requirements as PARTIAL (not COMPLETE).
6. `--dry-run` shows proposed status changes without persisting.

## Files to Create/Modify

- `internal/adapters/gitlab.go` -- CI integration methods
- `internal/adapters/gitlab_test.go` -- Pipeline and artifact tests
- `internal/cmd/verify.go` -- --adapter flag for external verification

## Effort Estimate

0.5 weeks

## Test Strategy

- Mock pipeline API: simulate successful and failed pipelines
- JUnit parsing: verify test name extraction from XML artifacts
- Job mapping: verify correct test files matched to jobs
- Failed jobs: verify PARTIAL status assignment
- No pipeline: verify clean error when no pipelines exist
