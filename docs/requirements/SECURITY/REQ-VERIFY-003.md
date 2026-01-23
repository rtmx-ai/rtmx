# REQ-VERIFY-003: Chaos Engineering Test Suite

## Status: NOT STARTED
## Priority: MEDIUM
## Phase: 13
## Effort: 2.5 weeks

## Description

RTMX shall include a chaos engineering test suite that verifies security properties under adverse network conditions. Using toxiproxy for fault injection, tests shall verify that access control remains correct during network partitions, latency spikes, and split-brain scenarios.

## Acceptance Criteria

- [ ] Toxiproxy integration for network fault injection
- [ ] Test: Offline access preserves cached grants
- [ ] Test: Revocations propagate after partition heals
- [ ] Test: Split-brain scenarios resolve correctly
- [ ] Test: Latency doesn't cause auth bypass
- [ ] Test: Retry storms don't cause privilege escalation
- [ ] Test: Clock skew doesn't invalidate valid tokens
- [ ] All chaos tests run in isolated network namespace
- [ ] CI runs chaos tests on merge to main

## Test Cases

- `tests/chaos/test_network_partition.py` - Partition tolerance tests
- `tests/chaos/test_split_brain.py` - Split-brain resolution tests
- `tests/chaos/test_latency.py` - High latency behavior tests
- `tests/chaos/test_retry_storms.py` - Retry behavior tests

## Technical Notes

### Toxiproxy Setup

```python
# tests/chaos/conftest.py
import pytest
from toxiproxy import Toxiproxy

@pytest.fixture
def toxiproxy():
    """Network fault injection fixture."""
    proxy = Toxiproxy()

    # Create proxy for rtmx-sync
    sync_proxy = proxy.create(
        name="rtmx-sync",
        listen="localhost:18080",
        upstream="localhost:8080"
    )

    yield proxy

    # Cleanup
    proxy.destroy("rtmx-sync")

@pytest.fixture
def partition(toxiproxy):
    """Create network partition."""
    def _partition(name: str):
        toxiproxy.get(name).add_toxic(
            "down",
            type="timeout",
            attributes={"timeout": 0}
        )
    return _partition

@pytest.fixture
def heal(toxiproxy):
    """Heal network partition."""
    def _heal(name: str):
        toxiproxy.get(name).remove_toxic("down")
    return _heal
```

### Partition Tolerance Tests

```python
# tests/chaos/test_network_partition.py
import pytest
from rtmx.sync import SyncClient

@pytest.mark.chaos
def test_offline_access_preserved(toxiproxy, client):
    """Verify cached grants work during partition."""
    # Grant access while online
    client.grant("alice", "rtmx", ["viewer"])

    # Verify access works
    assert client.can_access("alice", "rtmx")

    # Simulate network partition
    toxiproxy.get("rtmx-sync").add_toxic(
        "partition",
        type="timeout",
        attributes={"timeout": 60000}
    )

    # Should still work from cache
    assert client.can_access("alice", "rtmx")

@pytest.mark.chaos
def test_revocation_propagates_after_heal(toxiproxy, client, partition, heal):
    """Verify revocations sync after partition heals."""
    # Grant access
    client.grant("alice", "rtmx", ["viewer"])
    assert client.can_access("alice", "rtmx")

    # Partition network
    partition("rtmx-sync")

    # Revoke during partition (queued locally)
    client.revoke("alice", "rtmx")

    # Still has cached access during partition
    # (This is acceptable - eventually consistent)

    # Heal partition
    heal("rtmx-sync")

    # Wait for sync
    client.sync()

    # Revocation should now be effective
    assert not client.can_access("alice", "rtmx")
```

### Split-Brain Tests

```python
# tests/chaos/test_split_brain.py
@pytest.mark.chaos
def test_split_brain_grant_conflict(toxiproxy, client_a, client_b):
    """Verify conflicting grants resolve correctly."""
    # Create split-brain: A and B can't communicate
    toxiproxy.get("a-to-b").add_toxic("down", type="timeout")
    toxiproxy.get("b-to-a").add_toxic("down", type="timeout")

    # A grants viewer, B grants editor (conflict)
    client_a.grant("alice", "rtmx", ["viewer"])
    client_b.grant("alice", "rtmx", ["editor"])

    # Heal split
    toxiproxy.get("a-to-b").remove_toxic("down")
    toxiproxy.get("b-to-a").remove_toxic("down")

    # Sync both clients
    client_a.sync()
    client_b.sync()

    # Should converge to union (CRDT behavior)
    assert client_a.get_grants("alice", "rtmx") == {"viewer", "editor"}
    assert client_b.get_grants("alice", "rtmx") == {"viewer", "editor"}
```

### CI Configuration

```yaml
# .github/workflows/chaos-tests.yml
chaos-tests:
  runs-on: ubuntu-latest
  services:
    toxiproxy:
      image: ghcr.io/shopify/toxiproxy
      ports:
        - 8474:8474
  steps:
    - uses: actions/checkout@v4
    - name: Run chaos tests
      run: |
        pytest tests/chaos/ -v \
          --toxiproxy-host localhost \
          -m chaos
```

## Files to Create/Modify

- `tests/chaos/__init__.py` - Chaos test package
- `tests/chaos/conftest.py` - Toxiproxy fixtures
- `tests/chaos/test_network_partition.py` - Partition tests
- `tests/chaos/test_split_brain.py` - Split-brain tests
- `tests/chaos/test_latency.py` - Latency tests
- `.github/workflows/chaos-tests.yml` - CI workflow

## Dependencies

- REQ-ZT-003: JWT validation (chaos tests verify auth under faults)
- REQ-VERIFY-001: Property tests (chaos tests are complementary)

## Blocks

- REQ-VERIFY-004: NIST compliance uses chaos test evidence
