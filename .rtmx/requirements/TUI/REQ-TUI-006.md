# REQ-TUI-006: Live Refresh with File Watch

## Metadata
- **Category**: TUI
- **Subcategory**: Data
- **Priority**: MEDIUM
- **Phase**: 26
- **Status**: MISSING
- **Dependencies**: REQ-TUI-001
- **Blocks**: (none)

## Requirement

The TUI shall automatically detect changes to the RTM database CSV file
and refresh all views within 500ms of the file being modified, using
filesystem event notifications (fsnotify) with a polling fallback.

## Rationale

When multiple developers or AI agents are working on a project, the
database changes frequently. A TUI that shows stale data forces manual
refresh and creates confusion during concurrent work. File-watch-based
live refresh keeps the dashboard current without polling overhead.

## Design

### Watch Mechanism

```go
// Primary: fsnotify for instant detection
watcher, _ := fsnotify.NewWatcher()
watcher.Add(dbPath)

// Fallback: 2-second stat() polling for NFS/CIFS/FUSE mounts
ticker := time.NewTicker(2 * time.Second)
```

The watcher sends a Bubble Tea message (`DatabaseChangedMsg`) to the
application model, which triggers a reload of the in-memory database
and a re-render of all views.

### Debouncing

Multiple rapid writes (e.g., `rtmx verify --update` touching many rows)
are debounced with a 200ms window. Only one reload occurs per debounce
window.

### State Preservation

On reload, the TUI preserves:
- Active view (tab selection)
- Scroll position (as close as possible if items shift)
- Active filters and sort settings
- Selected item (by req_id, not index)

### Manual Refresh

The `r` key forces an immediate reload regardless of file watch state.
A brief "Refreshed" indicator appears in the status bar for 2 seconds.

## Acceptance Criteria

1. Database changes are detected within 500ms on local filesystem.
2. All views update to reflect the new database state.
3. Rapid consecutive writes produce a single debounced reload.
4. Active view, filters, and selection are preserved across reloads.
5. `r` key forces immediate manual refresh.
6. Status bar shows "Refreshed" indicator after reload.
7. Polling fallback works on non-inotify filesystems.
8. Watch errors are logged but do not crash the TUI.

## Files to Create/Modify

- `internal/tui/watcher.go` -- File watcher with debounce
- `internal/tui/watcher_test.go` -- Watch and debounce tests
- `internal/tui/app.go` -- Handle DatabaseChangedMsg
- `go.mod` -- Add fsnotify dependency

## Effort Estimate

0.5 weeks

## Test Strategy

- Write to database file, verify reload message sent within timeout
- Rapid writes: verify single reload per debounce window
- State preservation: filter active, reload, verify filter still applied
- Manual refresh: send `r` key, verify reload
- Fallback: disable fsnotify, verify polling detects changes
