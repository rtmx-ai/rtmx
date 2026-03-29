# RTMX Database Integrity Framework

## Design Analysis

### Problem

RTMX's core guarantee is that requirement status is derived from test evidence, not asserted by humans or agents. In agentic development, AI agents have the same filesystem permissions as developers -- any file the user can modify, the agent can modify. Detection-based approaches (signatures, audit logs) are insufficient because they can be bypassed.

### Enforcement Mechanisms Analysis

| Mechanism | Prevents Agent Bypass | No Sudo | Works Offline | Air-Gapped | Cross-Platform | Chosen |
|-----------|----------------------|---------|---------------|------------|----------------|--------|
| OS-level file permissions | Yes (with privilege separation) | No | Yes | Yes | Partial | No |
| Remote attestation service | Yes | Yes | No | No | Yes | No |
| Git branch protection | Partial (CI-only writes) | Yes | No | No | Yes | Yes (INT-003) |
| Interactive confirmation | Yes | Yes | Yes | Yes | Yes | No (breaks automation) |
| Hardware token attestation | Yes | Depends | Yes | Yes | Partial | No (adoption barrier) |
| CRDT proof-of-verification | Yes (at sync boundary) | Yes | Yes | Yes | Yes | Yes (INT-002) |

### Chosen Architecture: Defense in Depth

Two complementary enforcement mechanisms, selectable per deployment:

**1. CRDT Proof-of-Verification (REQ-INT-002) -- Decentralized**

Status change operations in the sync protocol require a cryptographic proof:
- The proof includes a hash of the test output that justifies the status change
- The proof is signed by the verifier's Ed25519 key (from REQ-SEC-002)
- During CRDT merge, operations without valid proofs are rejected
- This works offline, in air-gapped environments, and across trust boundaries

Trust policies scale from single developer to enterprise federation:
- `self`: only local key can verify (solo developer)
- `team`: configured team keys (small team)
- `delegated`: org-designated verifiers (enterprise)
- `web-of-trust`: N-of-M attestation (federation)

**2. Git Branch Protection (REQ-INT-003) -- Centralized**

For teams using GitHub/GitLab, status changes to `database.csv` are restricted:
- Only the CI pipeline (via `rtmx verify --update`) can modify status fields
- Branch protection rules prevent direct pushes to main
- The `rtmx security` command (REQ-SEC-012) validates this is configured
- This is the pragmatic near-term solution for Git-hosted projects

### Migration Path

1. **Current state**: status changes unenforceable (detection via SEC-013 staleness warning)
2. **Phase 1**: Git branch protection (INT-003) -- CI-only writes, immediate for Git users
3. **Phase 2**: CRDT proofs (INT-002) -- decentralized enforcement, future sync protocol
4. **Phase 3**: Combined -- branch protection + proofs for defense in depth

### Design Decisions

1. **No sudo required**: both mechanisms work with user-level privileges
2. **Offline-first**: CRDT proofs work without network; Git protection degrades gracefully
3. **Agent-safe**: agents cannot forge proofs (private key required) or bypass CI gates
4. **Configurable**: teams choose their enforcement level via `rtmx.yaml`

### Configuration

```yaml
rtmx:
  integrity:
    enforcement: "branch-protection"  # or "crdt-proofs" or "both"
    trust_policy: "team"              # self, team, delegated, web-of-trust
    require_proof: false              # require proof for local verify --update
```

### Adversarial Analysis

| Attack | Branch Protection | CRDT Proofs | Both |
|--------|------------------|-------------|------|
| Agent edits CSV directly | Blocked by CI gate | Detected at sync, rejected | Blocked + rejected |
| Forge test results | Blocked (CI runs real tests) | Proof hash won't match | Both block |
| Compromise CI runner | Succeeds (runner is trusted) | Proof valid (runner has key) | Succeeds (known limitation) |
| Compromise verifier key | N/A | Succeeds until key revoked | CI gate still holds |
| Replay old proof | N/A | Rejected (nonce/sequence) | Both block |

The remaining attack vector (compromised CI runner) is mitigated by:
- Pinning GitHub Actions to commit SHAs (REQ-SEC-004)
- Verify throughput thresholds (REQ-SEC-011)
- Security posture checks (REQ-SEC-012)
