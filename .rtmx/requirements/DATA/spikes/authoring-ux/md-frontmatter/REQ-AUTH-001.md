---
req_id: REQ-AUTH-001
category: AUTH
subcategory: Login
requirement_text: Users shall authenticate before accessing protected rooms.
target_value: "401 without token; 200 with session"
status: PARTIAL
priority: HIGH
phase: 8
effort_weeks: 1.0
dependencies: []
blocks:
  - REQ-AUTH-002
acs:
  - ac_id: AC-1
    statement: Unauthenticated WebSocket joins receive close code 4401.
    required: true
    severity: must
  - ac_id: AC-2
    statement: Valid Bearer session tokens are accepted on /auth/me.
    required: true
    severity: must
test_bindings:
  - binding_id: TB-1
    ac_id: AC-1
    test_module: internal/sync/room_test.go
    test_function: TestRoomRejectsMissingToken
  - binding_id: TB-2
    ac_id: AC-2
    test_module: internal/cmd/auth_test.go
    test_function: TestAuthStatusWithValidToken
---

# REQ-AUTH-001: Authenticate before protected rooms

## Rationale

Unauthenticated clients must not observe room CRDT state. Close code 4401
matches entitlement-style signaling already used for unpaid orgs (4402).
Bearer session tokens keep the CLI and website on one auth path without
pasting long-lived API keys into shell history.

## Day-to-day editing

- Change AC text or add an AC by editing the YAML `acs` list (stable `ac_id`).
- Bindings stay in frontmatter next to the AC they prove.
- Body Markdown is narrative only and is omitted from agent verify projections.
