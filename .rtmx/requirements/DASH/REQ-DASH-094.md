# REQ-DASH-094: Graph Performance Test Suite

## Summary

Automated performance budget tests ensuring graph operations stay within acceptable time bounds. Tests cover dagre layout computation (< 500ms for 300 nodes), cluster layout (< 100ms for 500 nodes all collapsed), and per-frame render budget (< 16ms). Tests run as part of CI using Go benchmarks for backend operations and documented browser timing assertions for frontend operations.

## Acceptance Criteria

1. A Go benchmark test verifies that graph JSON enrichment for 300 nodes completes in under 500ms.
2. A Go benchmark test verifies that graph JSON enrichment for 500 nodes with category grouping completes in under 100ms.
3. Performance thresholds are documented as constants in the test file for easy adjustment.
4. Tests use generated fixture data (not production data) to ensure reproducibility.
5. Frontend performance assertions are documented as manual test procedures with expected timing from browser Performance API (`performance.now()`).
6. CI runs the Go benchmark tests on every PR; frontend timing tests are manual regression checks.
7. Test failures indicate which operation exceeded its budget and by how much.

## Dependencies

- REQ-DASH-048 (dagre layout computation is the primary target)
- REQ-DASH-044 (enriched JSON generation is the backend target)

## Blocks

- None currently identified

## Files to Modify

- `internal/cmd/serve_dashboard_bench_test.go` (new file: Go benchmark tests for graph JSON enrichment)

## Test Strategy

- Go benchmark: BenchmarkGraphEnrichment300Nodes asserts < 500ms
- Go benchmark: BenchmarkGraphEnrichment500NodesGrouped asserts < 100ms
- Documentation: frontend performance test procedure with expected timings

## Effort

- 0.50 weeks

## Priority

- MEDIUM

## Phase

- 32
