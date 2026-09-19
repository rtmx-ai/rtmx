# REQ-DATA-001a: Document model schema spike

## Metadata
- **Category**: DATA
- **Subcategory**: DOCUMENT_MODEL
- **Priority**: HIGH
- **Phase**: 34
- **Status**: MISSING
- **Dependencies**:
- **Blocks**: REQ-DATA-001b|REQ-DATA-001c|REQ-DATA-001d|REQ-DATA-001e|REQ-DATA-001f
- **External ID**:
- **ADR**: docs/adr/0007-requirement-document-model.md

## Requirement

Define a minimal JSON Schema for one requirement document object: identity
fields aligned with CSV columns needed for graph and verify; `acs[]` with
stable `ac_id`; `test_bindings[]` linking AC (and optionally req) to tests;
optional narrative pointer or embedded prose. Encoding on disk is not decided
here—only the logical document shape that enables AC-to-test ATDD.

## Rationale

Without a shared schema, spikes diverge and migration cannot be idempotent.
Stable `ac_id` is the prerequisite for per-AC verify gaps and agentic ATDD.

## Acceptance Criteria

1. JSON Schema validates a hand-built fixture of at least three requirements including nested ACs and bindings.
2. Invalid documents fail validation with JSON Pointer (or equivalent) paths to the failing field.
3. A mapping table documents CSV column → schema field for identity, status, deps, and related scalars.
4. Schema documents that `ac_id` values are stable across rematerialization when AC text is unchanged (scheme named: e.g. `AC-1` ordinal or content-hash policy).
5. Explicitly out of scope: rewriting `internal/database` production path or flipping CLI defaults.

## Test Strategy

Schema unit tests against valid/invalid fixtures under `testdata/` or
`docs/schemas/` companion fixtures. No default `rtmx verify` behavior change.

## Out of Scope

Production CSV reader changes; sync protocol changes; authoring UX bake-off (001c).
