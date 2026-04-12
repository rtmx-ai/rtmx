# REQ-BENCH-008: Issue-Backend Sync Validation in Benchmarks

## Metadata
- **Category**: BENCH
- **Subcategory**: Sync
- **Priority**: MEDIUM
- **Phase**: 22
- **Status**: MISSING
- **Dependencies**: REQ-BENCH-001, REQ-GO-040
- **Blocks**: (none)

## Requirement

RTMX benchmark framework shall validate issue-backend sync by importing GitHub Issues from exemplar projects into the local RTM, demonstrating the full traceability loop: tests -> requirements -> roadmap.

## Rationale

Benchmarks currently validate only the scanner (from-tests) and verification (verify) pipeline. The sync adapter -- which connects local requirements to GitHub Issues and Jira -- is exercised only by unit tests with mock HTTP servers. Running `rtmx sync --import --dry-run` against real public issue trackers proves the adapters work against live APIs and demonstrates rtmx's most differentiating capability: bringing project roadmaps into local developer context.

## Design

### Sync Step in Benchmark Pipeline

After the existing scan and verify steps in `run-benchmark.sh`, add:

```bash
# Step 6: Sync with issue backend (read-only)
if [ -n "${SYNC_SERVICE}" ]; then
    echo "  Syncing with ${SYNC_SERVICE}..."
    rtmx sync --service "${SYNC_SERVICE}" --import --dry-run
fi
```

### Extended Config Schema

Add optional sync fields to benchmark configs:

```yaml
sync:
  service: github          # github or jira
  repo: cli/cli            # for GitHub: owner/repo
  query:                   # optional filter
    labels: enhancement
    state: open
  expected_items: 10       # minimum items imported
```

### Exemplar Sync Targets

| Language | Exemplar | Sync Service | Expected Items |
|----------|----------|--------------|----------------|
| Go | cli/cli | GitHub Issues | >= 50 |
| Python | psf/requests | GitHub Issues | >= 20 |
| JS/TS | sindresorhus/got | GitHub Issues | >= 10 |
| Java | google/gson | GitHub Issues | >= 10 |
| C# | jbogard/MediatR | GitHub Issues | >= 5 |
| Rust | rtmx-ai/aegis-cli | GitHub Issues | >= 5 |

### Sync Result Schema

Extend `BenchmarkResult` with sync fields:

```json
{
  "sync_service": "github",
  "sync_items_found": 47,
  "sync_status": "pass"
}
```

### Regression Detection

A sync regression occurs when:
- Item count drops below expected (API parsing broke)
- Sync status changes from "pass" to "fail" (adapter broke)
- Previously importable fields are missing (schema change)

### Authentication

Public GitHub repos allow unauthenticated read access to issues (rate-limited to 60 req/hr). For nightly CI, this is sufficient. If rate limiting becomes a problem, use `GITHUB_TOKEN` from the workflow's `secrets.GITHUB_TOKEN`.

### Dry-Run Only

All sync operations in benchmarks MUST use `--dry-run` to prevent writing to external services. The benchmark validates that sync *can* import, not that it *does* import.

## Acceptance Criteria

1. `run-benchmark.sh` includes a sync step when config has `sync:` section
2. At least 3 exemplar configs include sync configuration
3. `rtmx sync --import --dry-run` succeeds against at least one live GitHub repo in CI
4. Sync results are captured in the benchmark result JSON
5. Regression detection covers sync item count drops

## Files to Create/Modify

- `benchmarks/scripts/run-benchmark.sh` -- add sync step
- `benchmarks/configs/*.yaml` -- add sync sections
- `internal/benchmark/config.go` -- add SyncConfig struct
- `internal/benchmark/regression.go` -- add sync regression checks

## Effort Estimate

1.5 weeks

## Test Strategy

- Unit test: `internal/benchmark/config_test.go` validates SyncConfig parsing
- Unit test: `internal/benchmark/regression_test.go` validates sync regression detection
- Integration: `make -C benchmarks run LANG=go` includes sync step
- E2E: nightly CI workflow exercises live GitHub Issues API
