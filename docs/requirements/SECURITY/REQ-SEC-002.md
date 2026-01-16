# REQ-SEC-002: End-to-End Encryption

## Requirement
CRDT payloads shall be end-to-end encrypted.

## Phase
10 (Collaboration) - Sync Server Security

## Rationale
End-to-end encryption ensures that even the sync server operator cannot read requirement content. This is critical for defense contractors, regulated industries, and any organization with sensitive IP in their requirements.

## Acceptance Criteria
- [ ] CRDT updates are encrypted client-side before transmission
- [ ] Server stores and relays only ciphertext
- [ ] Decryption keys never leave client devices
- [ ] AES-256-GCM or ChaCha20-Poly1305 cipher used
- [ ] Key derivation uses Argon2id or HKDF
- [ ] Project keys can be rotated without data loss

## Technical Notes
- Leverage Yjs encryption providers or implement custom encryption layer
- Consider `y-webrtc` for peer-to-peer sync bypassing server entirely
- Key exchange via X25519 (Curve25519) for forward secrecy
- Store encrypted key material in user's keychain/credential store

## Crypto-Agility Architecture

This requirement is the foundation for RTMX's crypto-agility layer. The encryption implementation must support runtime algorithm selection to enable:

### Algorithm Provider Pattern
```python
class CryptoProvider(Protocol):
    def encrypt(self, plaintext: bytes, key: bytes) -> bytes: ...
    def decrypt(self, ciphertext: bytes, key: bytes) -> bytes: ...
    def key_exchange(self, peer_public: bytes) -> tuple[bytes, bytes]: ...
    def sign(self, message: bytes, private_key: bytes) -> bytes: ...
    def verify(self, message: bytes, signature: bytes, public_key: bytes) -> bool: ...
```

### Supported Algorithm Profiles

| Profile | Symmetric | Key Exchange | Signature | Use Case |
|---------|-----------|--------------|-----------|----------|
| `classic` | AES-256-GCM | X25519 | Ed25519 | Default, maximum compatibility |
| `pqc-hybrid` | AES-256-GCM | X25519 + ML-KEM-768 | Ed25519 + ML-DSA-65 | Quantum-resistant transition |
| `fips` | AES-256-GCM (FIPS) | ECDH P-384 | ECDSA P-384 | Government compliance |
| `pqc-only` | AES-256-GCM | ML-KEM-768 | ML-DSA-65 | Future post-quantum native |

### Configuration
```yaml
crypto:
  profile: classic  # classic | pqc-hybrid | fips | pqc-only
  fips_mode: false
  allow_fallback: true  # Fall back to classic if PQC unavailable
```

### Key Rotation with Algorithm Migration
When rotating keys, the crypto layer must:
1. Re-encrypt data with new algorithm if profile changed
2. Maintain ability to decrypt old data during transition period
3. Log algorithm changes for compliance auditing

## Security Properties
- **Confidentiality**: Server sees only encrypted blobs
- **Integrity**: GCM/Poly1305 provides authenticated encryption
- **Forward secrecy**: Compromised keys don't expose past data

## Test Cases
1. Verify payload is encrypted before network transmission
2. Verify server cannot decrypt stored data
3. Verify key rotation re-encrypts existing data
4. Verify tampered ciphertext is rejected

## Dependencies
- REQ-CRDT-001 (CRDT layer exists)

## Effort
2.0 weeks
