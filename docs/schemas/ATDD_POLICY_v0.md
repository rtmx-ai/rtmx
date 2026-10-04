# ATDD completeness policy: `atdd-required-all-v0` (REQ-DATA-001b AC3)

## Aggregation rules

For a requirement with one or more **required** ACs:

| Condition | Aggregated status |
|---|---|
| Every required AC has evidence `pass` | `COMPLETE` |
| At least one required AC has `pass` or `fail`, but not all required ACs `pass` | `PARTIAL` |
| No required AC has `pass` or `fail` (all `gap`) | `MISSING` |

Optional ACs (`required: false`) do not block `COMPLETE`.

A required AC with status `fail` or `gap` **must not** yield silent `COMPLETE`.

## Req-level-only bindings (compat)

When a requirement has **no** `acs[]` entries, a binding with `req_id` (and no
`ac_id`) drives status:

- all such evidence pass → `COMPLETE`
- mix of pass/fail → `PARTIAL`
- no evidence → `MISSING`

## Non-goals

This policy is for the fixture spike path in `internal/docmodel/acverify` only.
It is not the default `rtmx verify` promotion rule until a later REQUIREMENT
explicitly wires it.
