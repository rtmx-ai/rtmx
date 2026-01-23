# REQ-ZT-002: OpenZiti Dark Service for rtmx-sync

## Status: NOT STARTED
## Priority: HIGH
## Phase: 10
## Effort: 3.5 weeks

## Description

rtmx-sync shall be deployed as an OpenZiti dark service with no public ports exposed. The service shall be accessible only via the Ziti overlay network, making it invisible to port scanners and immune to network-based attacks. This implements zero-trust network access where identity verification happens before any network connection.

## Acceptance Criteria

- [ ] rtmx-sync binds only to Ziti service, not TCP ports
- [ ] No public IP/port exposure (dark service)
- [ ] OpenZiti SDK integrated for service hosting
- [ ] Ziti identity enrollment from Zitadel JWT
- [ ] Service policies restrict access to authorized identities
- [ ] End-to-end encryption via Ziti (no TLS termination)
- [ ] Connection logging for audit trail
- [ ] Graceful degradation when Ziti controller unavailable

## Test Cases

- `tests/test_ziti.py::TestDarkService` - Service binding tests
- `tests/test_ziti.py::TestIdentityEnrollment` - JWT-based enrollment
- `tests/test_ziti.py::TestServicePolicies` - Access control tests
- `tests/test_ziti.py::TestE2EEncryption` - Encryption verification
- `tests/test_ziti.py::TestAuditLogging` - Connection logging tests

## Technical Notes

### Dark Service Architecture

```
Traditional Server:          Dark Service (OpenZiti):
┌─────────────────┐          ┌─────────────────┐
│  Load Balancer  │          │   Ziti Router   │
│  (public IP)    │          │  (Ziti overlay) │
├─────────────────┤          ├─────────────────┤
│    Firewall     │          │  Service Edge   │
│  (port 443)     │          │  (no ports)     │
├─────────────────┤          ├─────────────────┤
│   rtmx-sync     │          │   rtmx-sync     │
│  (0.0.0.0:8080) │          │  (Ziti bind)    │
└─────────────────┘          └─────────────────┘
     ↑ Exposed                    ↑ Dark
```

### Server Implementation

```python
# rtmx-sync/main.py
import openziti

def main():
    # Load Ziti identity (enrolled from Zitadel JWT)
    ztx = openziti.load('/etc/rtmx-sync/identity.json')

    # Zitify the server - no public ports
    cfg = dict(ztx=ztx, service="rtmx-sync")
    openziti.monkeypatch(bindings={('0.0.0.0', 8080): cfg})

    # Existing FastAPI/WebSocket code works unchanged
    # But now only reachable via Ziti overlay
    uvicorn.run(app, host="0.0.0.0", port=8080)
```

### Service Policy Example

```json
{
  "name": "rtmx-sync-access",
  "type": "Dial",
  "identityRoles": ["#rtmx-users"],
  "serviceRoles": ["@rtmx-sync"],
  "semantic": "AnyOf"
}
```

### Attack Surface Comparison

| Attack Vector | Traditional Server | Dark Service (Ziti) |
|---------------|-------------------|---------------------|
| Port scanning | Exposed ports visible | No listening ports |
| DDoS | Public endpoint vulnerable | No public endpoint |
| Credential stuffing | Login endpoint exposed | Only Ziti-auth'd clients |
| Man-in-the-middle | TLS terminates at LB | End-to-end via Ziti |
| Insider threat | Network = access | Identity-based policies |

## Files to Create/Modify

- `rtmx-sync/src/ziti/service.py` - Ziti service binding
- `rtmx-sync/src/ziti/identity.py` - Identity enrollment
- `rtmx-sync/src/ziti/policies.py` - Policy management
- `rtmx-sync/deployment/` - Kubernetes/Docker configs
- `tests/test_ziti.py` - Integration tests

## Dependencies

- REQ-ZT-001: Zitadel OIDC integration (for identity enrollment)

## Blocks

- REQ-ZT-003: JWT validation happens after Ziti connection
