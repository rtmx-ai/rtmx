# REQ-SYNC-002 (fixture excerpt)

## Acceptance Criteria

1. Client dials `ws(s)://.../sync/{org_slug}/{room}` with Bearer token or API key.
2. Handshake completes per sync-protocol-v1 (SYNC_STEP1 / STEP2 / UPDATE).
3. Local CSV changes are sent as updates; remote updates are written to CSV.
4. `rtmx serve --sync-url` actually connects; live dashboard is not polling-only
   when sync-url is set.
5. Token expiry or 4401/4402/4403/4409 are surfaced to the user without crashing.
6. Offline operation is unchanged when sync-url is unset.

## Rationale

Narrative that does not round-trip.
