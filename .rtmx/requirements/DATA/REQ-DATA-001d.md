# REQ-DATA-001d: Migration round-trip spike

## Metadata
- **Category**: DATA
- **Subcategory**: DOCUMENT_MODEL
- **Priority**: HIGH
- **Phase**: 34
- **Status**: MISSING
- **Dependencies**: REQ-DATA-001a
- **Blocks**: REQ-DATA-001f
- **External ID**:
- **ADR**: docs/adr/0007-requirement-document-model.md

## Requirement

Prototype migrate (CLI stub or script) from CSV + Markdown `## Acceptance
Criteria` sections into structured documents conforming to REQ-DATA-001a, then
project back to CSV. Identity, dependencies, status, and AC text must
round-trip within documented limits; `ac_id` assignment must be stable when
content is unchanged.

## Rationale

rtmx 2.0 requires full backward compatibility and migration. Without
idempotent round-trip, AC-to-test ATDD cannot be adopted on existing projects.

## Acceptance Criteria

1. Parser handles representative AC shapes (numbered lists and checkbox lists), including samples patterned on REQ-VERIFY-013 and REQ-SYNC-002.
2. Emits stable `ac_id`s under the scheme named in 001a when AC text is unchanged.
3. CSV export drops AC nesting without losing req-level identity, status, deps, and other mapped scalars.
4. Second migrate does not churn `ac_id`s when content is unchanged (idempotence).
5. Documented limits list what does not round-trip (e.g. free-form Implementation history).

## Test Strategy

Fixture directory with CSV+MD in → structured out → CSV out; golden or
equality assertions on req_id, status, deps, and ac text/`ac_id`.

## Out of Scope

Migrating the live `.rtmx/database.csv` tree; production `rtmx migrate` polish UX.
