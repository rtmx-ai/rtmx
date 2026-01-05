# REQ-SEC-002: End-to-End Encryption

## Requirement
CRDT payloads shall be end-to-end encrypted.

## Phase
9 (CRDT) - MVP Security

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
