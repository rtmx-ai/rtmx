# REQ-GO-077: Automated Requirement PR Adjudication

## Metadata
- **Category**: SYNC
- **Subcategory**: CrossRepo
- **Priority**: MEDIUM
- **Phase**: 17
- **Status**: MISSING
- **Dependencies**: REQ-GO-076
- **Blocks**: (none)

## Requirement

rtmx-enabled projects shall support configurable adjudication rules in `.rtmx/config.yaml` that automatically validate, label, and optionally auto-merge incoming requirement PRs based on schema compliance and policy rules.

## Rationale

When multiple repos exchange requirements via PR (REQ-GO-076), the receiving project needs consistent acceptance criteria. Manual review of every requirement PR doesn't scale across a multi-repo ecosystem. Configurable rules allow teams to define what constitutes an acceptable requirement proposal — schema validity, required fields, naming conventions, phase alignment — and automate the routine checks while flagging exceptions for human review.

## Design

### Configuration

```yaml
# .rtmx/config.yaml
rtmx:
  adjudication:
    enabled: true
    rules:
      schema_valid: required     # Reject if CSV row doesn't match schema
      spec_file_present: required # Reject if no .md spec file included
      id_scheme: warn            # Warn if ID doesn't match project pattern
      phase_exists: required     # Reject if target phase doesn't exist
      no_circular_deps: required # Reject if new req creates dependency cycle
    auto_merge:
      enabled: false             # Default: require human approval
      conditions:
        - schema_valid
        - spec_file_present
        - no_circular_deps
    labels:
      pass: "requirement-accepted"
      fail: "requirement-needs-review"
      auto: "auto-merged"
```

### GitHub Action / Git Hook

A GitHub Action (or pre-merge hook) that:
1. Detects PRs modifying `.rtmx/database.csv` or requirement spec files
2. Runs adjudication rules against the diff
3. Posts check results as PR status/comment
4. Auto-merges if all `auto_merge.conditions` pass and `auto_merge.enabled`
5. Labels PR based on outcome

### CLI Command

```
rtmx adjudicate --pr 123
```

Manually run adjudication against a specific PR (for local testing or CI).

### Rule Evaluation

Each rule returns one of:
- `pass` — requirement meets criteria
- `warn` — non-blocking concern, added to PR comment
- `fail` — blocking, PR cannot merge until resolved

## Acceptance Criteria

1. `.rtmx/config.yaml` accepts `adjudication` configuration block
2. `schema_valid` rule rejects PRs with malformed CSV rows
3. `spec_file_present` rule rejects PRs missing spec markdown
4. `no_circular_deps` rule rejects PRs that create dependency cycles
5. `id_scheme` rule warns when requirement ID doesn't match project pattern
6. Auto-merge works when all conditions pass and is enabled
7. PR comment posted with rule evaluation results
8. Labels applied based on pass/fail outcome
9. `rtmx adjudicate` works from CLI for local testing
10. Disabled by default (opt-in)

## Test Strategy

- **Test Module**: `internal/sync/adjudicate_test.go`
- **Test Function**: `TestAdjudicateSchemaValid`, `TestAdjudicateAutoMerge`
- **Validation Method**: Integration Test
