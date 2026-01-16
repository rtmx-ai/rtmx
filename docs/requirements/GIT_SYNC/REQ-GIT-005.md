# REQ-GIT-005: Offline-first operation

## Status: NOT STARTED
## Priority: HIGH
## Phase: 15
## Effort: 1.5 weeks

## Description

All RTMX CLI commands shall work without network connectivity, using Git as the only synchronization primitive. The system shall operate in a fully offline mode by default, with optional network features that gracefully degrade when unavailable.

## Acceptance Criteria

- [ ] All core RTMX commands work without network access
- [ ] `rtmx status`, `rtmx health`, `rtmx validate` never require network
- [ ] `rtmx sync` uses Git fetch/push as only network operation
- [ ] No implicit network calls in any command
- [ ] GitHub adapter operations fail gracefully with clear error
- [ ] Jira adapter operations queue locally when offline
- [ ] `--offline` flag forces offline mode even with network
- [ ] Local CRDT state persists across CLI invocations
- [ ] Git is the sole synchronization mechanism for RTM data

## Test Cases

- `tests/test_offline.py::test_status_no_network` - Status works offline
- `tests/test_offline.py::test_health_no_network` - Health check works offline
- `tests/test_offline.py::test_validate_no_network` - Validation works offline
- `tests/test_offline.py::test_sync_uses_git` - Sync only uses git
- `tests/test_offline.py::test_github_graceful_fail` - GitHub errors handled
- `tests/test_offline.py::test_jira_offline_queue` - Jira queues locally
- `tests/test_offline.py::test_offline_flag` - Explicit offline mode
- `tests/test_offline.py::test_crdt_persistence` - State survives restart

## Technical Notes

### Offline-First Architecture

```
+-------------------+     +------------------+
|    Local CLI      |     |   Remote Repo    |
|                   |     |                  |
| +---------------+ |     | +-------------+  |
| |   RTM CSV     | |<--->| |  RTM CSV    |  |
| +---------------+ | git | +-------------+  |
|        |          |     |                  |
| +---------------+ |     +------------------+
| |  CRDT State   | |
| +---------------+ |
|        |          |
| +---------------+ |
| | Offline Queue | |
| +---------------+ |
+-------------------+
```

### Command Categories

1. **Always Offline** (never touch network)
   - `rtmx status` - Show RTM status
   - `rtmx health` - Validate RTM health
   - `rtmx validate` - Schema validation
   - `rtmx graph` - Dependency analysis
   - `rtmx from-tests` - Update from test results

2. **Git-Based Sync** (network via Git only)
   - `rtmx sync` - Equivalent to `git pull && git push`
   - `rtmx diff --branch` - Uses `git diff`

3. **Optional Network** (graceful degradation)
   - `rtmx github sync` - Sync with GitHub issues
   - `rtmx jira sync` - Sync with Jira
   - `rtmx ai suggest` - AI-powered suggestions

### Offline Queue Implementation

For adapters that require network:

```python
class OfflineQueue:
    """Queue operations when offline for later replay."""

    def __init__(self, queue_dir: Path):
        self.queue_dir = queue_dir

    def enqueue(self, adapter: str, operation: str, data: dict) -> str:
        """Queue an operation for later execution."""
        entry = {
            "id": uuid4().hex,
            "adapter": adapter,
            "operation": operation,
            "data": data,
            "queued_at": datetime.utcnow().isoformat(),
        }
        path = self.queue_dir / f"{entry['id']}.json"
        path.write_text(json.dumps(entry))
        return entry["id"]

    def replay(self, adapter_factory: Callable) -> list[Result]:
        """Replay queued operations in order."""
        results = []
        for path in sorted(self.queue_dir.glob("*.json")):
            entry = json.loads(path.read_text())
            adapter = adapter_factory(entry["adapter"])
            result = adapter.execute(entry["operation"], entry["data"])
            if result.success:
                path.unlink()
            results.append(result)
        return results
```

### Network Detection

```python
def is_offline() -> bool:
    """Check if we should operate in offline mode."""
    # Explicit offline flag
    if os.environ.get("RTMX_OFFLINE") == "1":
        return True

    # Check for network (optional, don't fail if check fails)
    try:
        socket.create_connection(("github.com", 443), timeout=1)
        return False
    except (socket.timeout, OSError):
        return True
```

### Git as Sync Primitive

All RTM synchronization flows through Git:

```bash
# Sync is just Git operations
rtmx sync
# Equivalent to:
git fetch origin
git merge origin/main  # Uses RTMX merge driver
git push origin main
```

## Files to Create/Modify

- `src/rtmx/offline.py` - Offline queue and detection
- `src/rtmx/cli/sync.py` - Git-based sync command
- `src/rtmx/adapters/github.py` - Add offline handling
- `src/rtmx/adapters/jira.py` - Add offline queue
- `tests/test_offline.py` - Offline operation tests

## Dependencies

- REQ-CRDT-001: CRDT layer for local state
- REQ-CRDT-005: CRDT offline operations

## Blocks

- REQ-GIT-006: Branch-aware requirement status
