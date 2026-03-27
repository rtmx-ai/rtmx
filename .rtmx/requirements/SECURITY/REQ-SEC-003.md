# REQ-SEC-003: Grant Enforcement in ApplyUpdates

## Metadata
- **Category**: SECURITY
- **Subcategory**: AccessControl
- **Priority**: CRITICAL
- **Phase**: 20
- **Status**: MISSING
- **Effort**: 1 week

## Requirement

`ApplyUpdates()` shall enforce grant constraints before applying any update. Each node independently checks the sender's grant against the target requirement, ensuring consistency across the decentralized network.

## Design

### Decentralized Enforcement

Every node in the network applies the same deterministic grant rules. If two nodes have the same grant config and receive the same signed message, they make the same accept/reject decision. No coordination or centralized authority required.

### Integration with Message Signing (REQ-SEC-002)

1. Signed message arrives with sender's public key
2. Look up sender identity from trusted_peers config
3. Look up sender's grant from grants config
4. For each RequirementUpdate in the message:
   - Load the target requirement from the database
   - Check `ConstraintAllows(grant, requirement)`
   - Check `IsGrantActive(grant)`
   - Reject if constraint or expiry fails
5. Apply only the authorized updates
6. Return rejected updates as errors

### ApplyUpdates Signature Change

```go
// Before:
func ApplyUpdates(db *database.Database, updates []RequirementUpdate) *SyncResult

// After:
func ApplyUpdates(db *database.Database, updates []RequirementUpdate, opts ...ApplyOption) *SyncResult

type ApplyOption func(*applyConfig)
func WithGrant(grant config.SyncGrant) ApplyOption
func WithRequireGrant(require bool) ApplyOption
```

### Behavior

- No grant provided + RequireGrant false: apply all (local operations, backward compat)
- No grant provided + RequireGrant true: reject all
- Grant provided: apply only updates that pass ConstraintAllows + IsGrantActive
- Rejected updates reported in SyncResult.Errors

## Acceptance Criteria

1. ApplyUpdates accepts optional grant configuration
2. Updates violating category constraints are rejected
3. Updates violating requirement ID constraints are rejected
4. Updates from expired grants are rejected
5. Rejected updates reported in SyncResult.Errors with reason
6. No grant + RequireGrant=false allows all (backward compat)
7. Deterministic: same inputs produce same accept/reject on any node
