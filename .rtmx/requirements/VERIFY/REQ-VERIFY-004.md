# REQ-VERIFY-004: Lenient and Strict Results JSON Parsing

## Metadata
- **Category**: VERIFY
- **Subcategory**: ResultsParsing
- **Priority**: HIGH
- **Phase**: 18
- **Status**: PARTIAL
- **Dependencies**: REQ-VERIFY-002
- **Blocks**: REQ-VERIFY-001

## Requirement

`rtmx verify --results <file>` shall parse the RTMX results JSON format
in both its canonical nested form and a convenience flat form, accept a
`status` string in place of the boolean `passed`, and shall reject
unknown fields with an actionable error.

## Rationale

A user reported (against v0.2.4) that the reproducer

```json
[{"req_id":"REQ-INGEST-030","test_name":"test_foo","test_file":"tests/unit/test_foo.py","status":"pass"}]
```

silently produced empty struct values: every entry was reported with
`invalid req_id ""` and missing `test_name`/`test_file`. Root cause: the
canonical schema nests marker fields under a `marker` object, but the
permissive `json.Unmarshal` accepted the flat top-level keys without
populating anything. The current parser neither honors the flat form
nor rejects unknown keys, so typos and naive payloads fail silently.

## Design

1. `Result` gains a custom `UnmarshalJSON` that:
   - Decodes via `json.Decoder` with `DisallowUnknownFields()` so typos
     surface immediately.
   - Accepts an auxiliary shape exposing both nested `marker` and
     top-level marker keys; when `marker` is absent the flat keys are
     promoted into `Result.Marker`. When both are supplied, nested wins
     and flat fills blanks.
   - Accepts a `status` string ("pass"/"passed"/"ok"/"success" → true,
     "fail"/"failed"/"error"/"errored"/"skip"/"skipped" → false) as an
     alternative to boolean `passed`. Unknown status strings produce a
     decode error.
2. `runVerifyFromResults` (in `internal/cmd/verify.go`) treats validation
   errors as fatal: if `results.Validate` returns any errors the command
   prints them and exits non-zero rather than silently producing zero
   matches.
3. `--results` help text gains a complete JSON example.

## Acceptance Criteria

1. Canonical nested payload parses and validates as before.
2. Flat payload parses, populates `Marker`, and validates without errors.
3. `"status":"pass"` and `"status":"fail"` map to `Passed` true/false.
4. Unknown top-level fields cause `Parse` to return an error mentioning
   the offending key.
5. Mixed nested+flat: nested values take precedence, flat fills blanks.
6. `rtmx verify --results <invalid>` exits non-zero and prints
   validation errors.
7. The bug-report reproducer succeeds end-to-end against a fixture
   database containing `REQ-INGEST-030`.

## Files to Create/Modify

- `internal/results/schema.go` — custom `UnmarshalJSON`
- `internal/results/schema_test.go` — table-driven tests for all forms
- `internal/cmd/verify.go` — fatal validation; expanded help text
- `.rtmx/requirements/VERIFY/REQ-VERIFY-002.md` — JSON payload example
- `features/verify_results.feature` — BDD spec
