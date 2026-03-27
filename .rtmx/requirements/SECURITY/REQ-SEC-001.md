# REQ-SEC-001: Results File Integrity via Source Attestation

## Metadata
- **Category**: SECURITY
- **Subcategory**: Integrity
- **Priority**: CRITICAL
- **Phase**: 20
- **Status**: MISSING
- **Effort**: 2 weeks

## Requirement

`rtmx verify --results` shall validate that results files include a source hash attestation proving the results were produced from the current test source files. Optionally, results may include a cryptographic signature from the runner identity for provenance.

## Design

### Source Hash Attestation

The results JSON includes a `source_hash` field computed from the test source files:

```json
{
  "rtmx_results": [...],
  "attestation": {
    "source_hash": "sha256:<hash of concatenated test source files>",
    "test_command": "pytest --rtmx-output results.json",
    "timestamp": "2026-03-26T00:00:00Z",
    "runner": "github-actions",
    "signature": "<optional Ed25519 or GPG signature>"
  }
}
```

`rtmx verify --results` recomputes the source hash from the test files referenced in the database and rejects results if the hash doesn't match.

### Signature (optional)

- In CI: results signed with runner identity (GitHub OIDC token or GPG key)
- Locally: signature optional (developer is trusted on their own machine)
- Config controls whether signature is required: `verify.require_signature: true`

### Behavior

- Missing attestation: warn but accept (backward compatibility for v0.x)
- Mismatched source hash: reject with clear error
- Missing signature when required: reject
- Valid attestation + signature: accept

## Acceptance Criteria

1. Results JSON schema includes attestation.source_hash field
2. verify --results rejects results with mismatched source hash
3. verify --results warns on missing attestation (backward compat)
4. Signature validation configurable (required in CI, optional locally)
