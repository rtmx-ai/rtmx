# REQ-SYNC-002: y-websocket Client for rtmx-sync

## Metadata
- **Category**: SYNC
- **Subcategory**: CRDT
- **Priority**: P0
- **Phase**: 33
- **Status**: COMPLETE
- **Dependencies**: REQ-GO-035|REQ-GO-042
- **Blocks**: REQ-MONO-016a
- **External ID**: rtmx-ai/REQ-MONO-016a

## Requirement

The Go CLI shall open a y-websocket connection to rtmx-sync using
auth-token-v1 credentials, push local RTM state, pull merged state, and
persist the result to the CSV database. `rtmx serve --sync-url` shall use
the same client instead of only printing the URL.

## Rationale

REQ-GO-042 described this client and is marked COMPLETE, but
`internal/sync` only defines JSON message types. There is no
`websocket.Dial`, no y-websocket handshake, and no CSV write-back from a
server merge. Monetization E2E (REQ-MONO-016a) cannot pass without a real
client.

## Acceptance Criteria

1. Client dials `ws(s)://.../sync/{org_slug}/{room}` with Bearer token or API key.
2. Handshake completes per sync-protocol-v1 (SYNC_STEP1 / STEP2 / UPDATE).
3. Local CSV changes are sent as updates; remote updates are written to CSV.
4. `rtmx serve --sync-url` actually connects; live dashboard is not polling-only
   when sync-url is set.
5. Token expiry or 4401/4402/4403/4409 are surfaced to the user without crashing.
6. Offline operation is unchanged when sync-url is unset.

## Test Strategy

`internal/sync/room_test.go` — `TestSyncClientRoundTrip` against a ygo
y-websocket server, so the client is checked against an independent
implementation of the protocol rather than a mock that agrees with it.
`internal/sync/follow_test.go` covers the long-lived subscription and
`internal/cmd/serve_room_test.go` covers AC4. Convergence with the shipped
Python server is the monorepo process test, REQ-MONO-016a.

## Implementation

- `internal/sync/room.go` — `RoomClient`: dial, handshake, publish, snapshot.
  Close codes 4401/4402/4403/4409 become a `RoomError` that explains itself.
- `internal/sync/follow.go` — `Follow`: keeps a room open for `rtmx serve`,
  retrying transient drops but returning refusals the user must act on.
- `internal/cmd/sync_room.go` — `rtmx sync --sync-url` with `--pull`,
  `--push`, and `--set REQ-ID=STATUS`, writing merged state back to the CSV.
- `internal/cmd/serve_room.go` — `rtmx serve --sync-url` applies remote room
  state to the served database and persists it.

Interop required a server fix: rtmx-sync framed its SyncStep2 reply with a
second `YMessageType.SYNC` byte, which pycrdt peers tolerate and every other
y-websocket client rejects. Regression test in rtmx-sync
`tests/test_collab.py::test_sync_step1_reply_is_a_single_y_websocket_envelope`.
