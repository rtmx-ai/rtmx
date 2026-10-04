# REQ-AUTH-001: Authenticate before protected rooms

## Rationale

Unauthenticated clients must not observe room CRDT state. Close code 4401
matches entitlement-style signaling already used for unpaid orgs (4402).
Bearer session tokens keep the CLI and website on one auth path without
pasting long-lived API keys into shell history.

## Notes

Structured ACs and bindings live in the companion JSONL record. Edit AC
statements and `ac_id` values there; keep this file for human narrative only.
