# REQ-TEST-009: Large Scale Performance Tests

## Status: NOT_STARTED
## Priority: LOW
## Phase: 4
## Effort: 0.5 weeks

## Description

Add performance/stress tests for large RTM databases to ensure acceptable performance at scale.

## Acceptance Criteria

- [ ] Test status command with 1000+ requirements completes in <5s
- [ ] Test backlog command with 1000+ requirements completes in <5s
- [ ] Test cycles command with complex dependency graphs completes in <10s
- [ ] Test from-tests with 500+ test files completes in <30s
- [ ] Memory usage stays under 500MB for 10,000 requirement database
- [ ] At least 5 new technique_stress tests

## Test Scenarios

### Database Scale
1. Generate 1000 requirement database
2. Generate 10000 requirement database
3. Complex dependency graph (average 5 deps per requirement)

### Command Performance
1. `rtmx status` on 1000 reqs < 5s
2. `rtmx backlog` on 1000 reqs < 5s
3. `rtmx cycles` on 1000 reqs with 100 edges < 10s
4. `rtmx deps` on 1000 reqs < 5s

### Memory Usage
1. Peak memory during status on 10000 reqs < 500MB
2. No memory leaks over repeated operations

## Files to Create

- `tests/test_performance.py`

## Test Implementation

```python
@pytest.mark.technique_stress
@pytest.mark.scope_system
def test_status_1000_requirements(large_database):
    """Status command completes in under 5 seconds with 1000 requirements."""
    start = time.time()
    result = run_rtmx("status", cwd=large_database)
    elapsed = time.time() - start
    assert result.returncode == 0
    assert elapsed < 5.0, f"Status took {elapsed:.2f}s, expected <5s"
```

## Notes

Performance thresholds are initial targets and may be adjusted based on baseline measurements.
