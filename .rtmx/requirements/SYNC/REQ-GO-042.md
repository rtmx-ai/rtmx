# REQ-GO-042: CRDT Client for Offline-First Sync

## Metadata
- **Category**: SYNC
- **Subcategory**: CRDT
- **Priority**: HIGH
- **Phase**: 13
- **Status**: MISSING
- **Effort**: 3 weeks
- **Dependencies**: REQ-GO-035 (Shadow requirements)
- **Blocks**: REQ-INT-001 (Integrity framework), REQ-INT-002 (Proof-of-verification)

## Requirement

Go CLI shall implement a WebSocket client that connects to an rtmx-sync server to synchronize the local RTM database via the y-websocket CRDT protocol. All merge logic is delegated to the server; the CLI is responsible for sending local state, receiving merged state, and updating the local CSV database.

## Rationale

The rtmx-sync server already implements CRDT merge semantics via pycrdt/Yrs. Rather than duplicating that complexity in Go, the CLI acts as a thin client that syncs local state to the server and receives the merged result. This preserves the local-first philosophy: the CLI works fully offline with CSV + Git, and optionally connects to rtmx-sync for real-time multi-user collaboration.

## Design

### Architecture

```
rtmx-go CLI                          rtmx-sync Server

 CSV Database                         CRDT Document (Room)
     |                                     |
     v                                     v
 SyncClient  ---[WebSocket]-->  y-websocket Handler
     |        <--[WebSocket]---        |
     v                                     v
 CSV Writer                          Yrs Merge Engine
```

### Data Flow

1. `rtmx sync push` -- Send local database state to server
   - Read local CSV database
   - Convert requirements to JSON updates
   - Send via WebSocket to server room
   - Server applies updates to CRDT document

2. `rtmx sync pull` -- Receive merged state from server
   - Connect to server room via WebSocket
   - Receive current CRDT state as requirement updates
   - Convert updates back to CSV rows
   - Write updated database to disk
   - Report changes (added, updated, conflicts)

3. `rtmx sync watch` -- Live sync (optional, for long-running sessions)
   - Maintain persistent WebSocket connection
   - Watch local CSV for changes, push updates
   - Receive remote changes, update local CSV
   - Display real-time change notifications

### Core Components

```go
// internal/sync/client.go

// SyncClient manages WebSocket connections to rtmx-sync server.
type SyncClient struct {
    ServerURL string           // e.g., "wss://sync.rtmx.ai"
    Room      string           // Room/document identifier
    Token     string           // Auth token (from rtmx auth login)
    OnUpdate  func(updates []RequirementUpdate)
}

// Connect establishes a WebSocket connection to the sync server.
func (c *SyncClient) Connect(ctx context.Context) error

// Push sends local database state to the server.
func (c *SyncClient) Push(db *database.Database) (*SyncResult, error)

// Pull receives the current merged state from the server.
func (c *SyncClient) Pull() (*database.Database, *SyncResult, error)

// Close terminates the WebSocket connection.
func (c *SyncClient) Close() error

// RequirementUpdate represents a change received from the server.
type RequirementUpdate struct {
    ReqID     string
    Action    string // "added", "updated", "removed"
    Fields    map[string]string
    Source    string // Which client made the change
    Timestamp time.Time
}
```

### Configuration

```yaml
rtmx:
  sync:
    server: "wss://sync.rtmx.ai"
    room: "rtmx-ai/rtmx"    # Default: GitHub repo identifier
    auto_sync: false             # Enable live sync on verify/status
```

### CLI Commands

```
rtmx sync push [--server URL] [--room NAME]    # Push local state to server
rtmx sync pull [--server URL] [--room NAME]    # Pull merged state from server
rtmx sync watch [--server URL] [--room NAME]   # Live bidirectional sync
rtmx sync status                                # Show sync connection status
```

### Files to Create

- `internal/sync/client.go` -- WebSocket client, push/pull/watch operations
- `internal/sync/client_test.go` -- Tests with mock WebSocket server
- `internal/sync/protocol.go` -- Message serialization (JSON requirement updates)
- `internal/sync/protocol_test.go` -- Protocol tests

### Files to Modify

- `internal/config/config.go` -- Add server/room/auto_sync to SyncConfig
- `internal/cmd/sync.go` -- Add push/pull/watch subcommands

## Acceptance Criteria

1. `rtmx sync push` sends local database state to rtmx-sync server via WebSocket
2. `rtmx sync pull` receives merged state and updates local CSV
3. Pull reports changes: requirements added, updated, or with conflicts
4. Connection requires auth token (from `rtmx auth login` or `RTMX_TOKEN` env)
5. Graceful handling of server unreachable (clear error, no data loss)
6. Local database is never corrupted by sync (atomic write with backup)
7. Works with the existing `sync:ALIAS:REQ-ID` shadow requirement format
8. `--dry-run` flag on push/pull shows what would change without writing

## Test Strategy

- **Test Module**: `internal/sync/crdt_test.go`
- **Test Function**: `TestCRDTMerge`
- **Validation Method**: Unit Test

### Test Cases

1. Push serializes database to correct JSON protocol format
2. Pull deserializes server response to correct requirements
3. Pull writes updated CSV atomically (backup created)
4. Pull reports added/updated/removed requirements correctly
5. Connection failure returns clear error without data loss
6. Dry-run mode shows changes without writing
7. Auth token passed in WebSocket headers
8. Reconnection after disconnect (watch mode)
9. Empty database push/pull (edge case)
10. Large database (100+ requirements) serialization performance
