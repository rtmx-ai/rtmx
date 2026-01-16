# REQ-CRDT-005: CRDT operations shall work offline

## Status: COMPLETE
## Priority: MEDIUM
## Phase: 9

## Description
CRDT operations shall work offline with local-first architecture. All edits are persisted locally and queued for sync when connectivity is restored.

## Acceptance Criteria
- [x] CRDT document state can be saved to local file
- [x] CRDT document state can be loaded from local file
- [x] Pending updates are queued when sync server unavailable
- [x] Queued updates are applied when connectivity restored
- [x] Local state survives CLI restarts
- [x] State file location is configurable via state_dir parameter

## Test Cases
- `tests/test_crdt.py::TestOfflineOperations::test_save_state_to_file`
- `tests/test_crdt.py::TestOfflineOperations::test_load_state_from_file`
- `tests/test_crdt.py::TestOfflineOperations::test_state_roundtrip`
- `tests/test_crdt.py::TestOfflineOperations::test_queue_pending_updates`
- `tests/test_crdt.py::TestOfflineOperations::test_apply_pending_to_document`
- `tests/test_crdt.py::TestOfflineOperations::test_sync_from_csv`
- `tests/test_crdt.py::TestSyncState::test_initial_state`
- `tests/test_crdt.py::TestSyncState::test_mark_synced`
- `tests/test_crdt.py::TestSyncState::test_mark_offline`

## Implementation Details

### OfflineStore Class
Located in `src/rtmx/sync/offline.py`:
- `save_state(doc)` - Atomic write of CRDT state to file
- `load_state()` - Load CRDT state from file
- `queue_update(update)` - Queue update for later sync
- `get_pending_updates()` - Get all queued updates in order
- `clear_pending_updates()` - Clear queue after successful sync
- `sync_from_csv(csv_path)` - Combined load/create with pending updates

### Local State Storage
- Default location: `.rtmx/sync/state.crdt` (binary CRDT state)
- Pending updates: `.rtmx/sync/pending/` (one file per update)
- Configurable via `state_dir` parameter

### State File Format
- Binary format using pycrdt's native serialization
- Includes full document state for fast recovery
- Atomic writes using temp file + rename to prevent corruption

### Pending Updates Queue
- Each update saved as timestamped file (microseconds)
- Applied in order when sync resumes
- Deleted after successful sync

## Notes
This enables the local-first workflow where users can work without network connectivity. All changes are captured locally and synchronized when the sync server becomes available.

## Dependencies
- REQ-CRDT-002: Y.Map structures

## Blocks
- None
