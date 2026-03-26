# REQ-GO-074: Health Check for Status Consistency

## Metadata
- **Category**: CLI
- **Subcategory**: Health
- **Priority**: HIGH
- **Phase**: 16
- **Status**: MISSING
- **Dependencies**: REQ-GO-012
- **Blocks**: (none)

## Requirement

`rtmx health` shall detect status consistency violations where a COMPLETE requirement depends on a MISSING or PARTIAL requirement.

## Rationale

When a requirement is marked COMPLETE but one of its dependencies is still MISSING, this indicates either:
1. The dependency was incorrectly modeled (should be removed)
2. The COMPLETE status is premature (should be downgraded)
3. The dependency was completed but its status wasn't updated

This was discovered during dependency validation when REQ-GO-070 (COMPLETE) depended on REQ-GO-067 (MISSING). The existing health checks catch orphaned deps, reciprocity, and cycles, but not this class of semantic inconsistency.

## Design

### New Health Check: `status_consistency`

```
[WARN] status_consistency: REQ-GO-070 is COMPLETE but depends on REQ-GO-067 (MISSING)
```

Severity: WARNING (not ERROR), because there may be legitimate reasons for this state during development (e.g., a dependency was relaxed but not yet removed from the database).

### Check Logic

```go
for _, req := range db.All() {
    if req.Status != database.StatusComplete {
        continue
    }
    for _, depID := range req.Dependencies {
        dep := db.Get(depID)
        if dep != nil && dep.Status == database.StatusMissing {
            warn("status_consistency", "%s is COMPLETE but depends on %s (%s)",
                req.ReqID, depID, dep.Status)
        }
    }
}
```

### JSON Output

```json
{
  "name": "status_consistency",
  "status": "WARN",
  "message": "1 status consistency issue(s) found",
  "details": [
    {
      "req_id": "REQ-GO-070",
      "status": "COMPLETE",
      "dependency": "REQ-GO-067",
      "dep_status": "MISSING"
    }
  ]
}
```

## Acceptance Criteria

1. `rtmx health` includes a `status_consistency` check
2. COMPLETE requirements depending on MISSING requirements produce a WARNING
3. COMPLETE requirements depending on PARTIAL requirements produce a WARNING
4. COMPLETE requirements depending on COMPLETE requirements produce no warning
5. MISSING/PARTIAL requirements depending on MISSING are not flagged (expected state)
6. `--json` output includes the new check in the checks array
7. Check name appears in health summary

## Files to Modify

- `internal/cmd/health.go` - Add `checkStatusConsistency` function
- `internal/cmd/health_test.go` - Tests for the new check

## Test Strategy

- Unit test: database with COMPLETE → MISSING dependency (should warn)
- Unit test: database with COMPLETE → COMPLETE dependency (should pass)
- Unit test: database with MISSING → MISSING dependency (should pass)
- Unit test: database with COMPLETE → PARTIAL dependency (should warn)
- Integration test: `rtmx health` output includes the check
