# REQ-COLLAB-002: Shadow Requirements for Partial Visibility

## Status: NOT STARTED
## Priority: HIGH
## Phase: 10
## Effort: 2.0 weeks

## Description

RTMX shall support shadow requirements that provide partial visibility into cross-repository dependencies when the user lacks full access. Shadow requirements show enough information to understand dependency relationships without exposing sensitive requirement details.

## Acceptance Criteria

- [ ] `ShadowRequirement` dataclass stores minimal cross-repo requirement info
- [ ] Shadow requirements include: req_id, external_repo, status, shadow_hash
- [ ] Shadow hash is SHA-256 of full requirement for verification without disclosure
- [ ] Visibility levels: `full` (all fields), `shadow` (status only), `hash_only` (existence only)
- [ ] `rtmx status` and `rtmx backlog` show shadow requirements with clear indicators
- [ ] Shadow requirements can be refreshed when online
- [ ] Cache shadow data locally for offline access
- [ ] UI shows "[SHADOW]" indicator and "Request Access" option

## Test Cases

- `tests/test_shadow.py::TestShadowRequirement` - ShadowRequirement dataclass tests
- `tests/test_shadow.py::TestShadowHash` - Hash generation and verification
- `tests/test_shadow.py::TestShadowVisibility` - Visibility level enforcement
- `tests/test_shadow.py::TestShadowCache` - Local caching behavior
- `tests/test_shadow.py::TestShadowDisplay` - CLI display formatting

## Technical Notes

### Shadow Requirement Model

```python
@dataclass
class ShadowRequirement:
    req_id: str                    # "REQ-COLLAB-007"
    external_repo: str             # "rtmx-ai/rtmx"
    shadow_hash: str               # SHA-256 for verification
    status: Status                 # COMPLETE/PARTIAL/MISSING
    visibility: str = "shadow"     # full | shadow | hash_only
    last_verified: datetime | None = None
```

### User Experience

```
$ rtmx backlog
REQ-SYNC-001  [MISSING]  WebSocket sync server
  depends on: rtmx-ai/rtmx:REQ-COLLAB-007 [SHADOW]
               Status: MISSING (verified 2h ago)
               Hash: abc123... (for verification)
               [Request Access] to view details
```

### Security Properties

1. Shadow hash allows verification without disclosure
2. Status updates require valid JWT with appropriate grants
3. Hash changes indicate requirement modification
4. No sensitive text/notes exposed in shadow view

## Files to Create/Modify

- `src/rtmx/models.py` - Add `ShadowRequirement` dataclass
- `src/rtmx/shadow.py` - Shadow resolution and caching logic
- `src/rtmx/cli/status.py` - Update for shadow display
- `src/rtmx/cli/backlog.py` - Update for shadow display
- `tests/test_shadow.py` - Comprehensive tests

## Dependencies

- REQ-COLLAB-001: Cross-repo dependency references

## Blocks

- REQ-ZT-003: JWT validation for shadow status updates
