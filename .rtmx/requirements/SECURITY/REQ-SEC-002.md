# REQ-SEC-002: Sync Protocol Message Signing

## Metadata
- **Category**: SECURITY
- **Subcategory**: Protocol
- **Priority**: CRITICAL
- **Phase**: 20
- **Status**: MISSING
- **Effort**: 2 weeks

## Requirement

Sync protocol messages shall be signed with Ed25519 keys and include replay protection. Every node independently validates signatures from trusted peer public keys before applying updates.

## Design

### Decentralized Model

RTMX targets fully decentralized operation where every node is a valid CI runner. There is no centralized server to delegate authentication to. Therefore, authentication must happen at the message level, not the transport level.

### Message Signing

```go
type SignedSyncMessage struct {
    Message   SyncMessage `json:"message"`
    PublicKey string      `json:"public_key"`  // Ed25519 public key (base64)
    Signature string      `json:"signature"`   // Ed25519 signature of Message bytes
    Sequence  uint64      `json:"sequence"`     // Monotonic counter for replay protection
}
```

### Trust Configuration

Trusted peer public keys stored in project config:

```yaml
rtmx:
  sync:
    trusted_peers:
      - name: "alice"
        public_key: "base64-encoded-ed25519-public-key"
      - name: "ci-bot"
        public_key: "base64-encoded-ed25519-public-key"
```

### Verification

`ApplyUpdates()` accepts a `SignedSyncMessage` and:
1. Verifies the signature against the embedded public key
2. Checks the public key is in the trusted_peers list
3. Rejects messages with sequence <= last seen sequence for that peer
4. Then applies the updates

### Key Discovery

Public keys travel with the project via Git/CRDT sync. Adding a trusted peer is a config change committed to the repo.

## Acceptance Criteria

1. SyncMessage includes Ed25519 signature and public key
2. Unsigned messages rejected by ApplyUpdates
3. Messages signed by unknown peers rejected
4. Replayed messages (same or lower sequence) rejected
5. Trusted peer config stored in rtmx.yaml
