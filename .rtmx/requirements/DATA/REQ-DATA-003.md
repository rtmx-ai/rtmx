# REQ-DATA-003: Migrate CSV+MD to structured store

## Metadata
- **Category**: DATA
- **Subcategory**: DOCUMENT_MODEL
- **Priority**: HIGH
- **Phase**: 35
- **Status**: COMPLETE
- **Dependencies**: REQ-DATA-001|REQ-DATA-002
- **Blocks**:
- **External ID**:
- **ADR**: docs/adr/0007-requirement-document-model.md

## Requirement

Provide rtmx migrate from CSV plus Markdown ACs into JSONL structured records with optional MD companion and stable ac_id idempotence.

## Acceptance Criteria

1. [ ] Traces to ADR-0007 Accepted decisions in DECISION_GATE_001f.
2. [ ] Tests bind `REQ-DATA-003`.
3. [ ] No managed-sync / sync-protocol-v1 break unless this is DATA-005 additive-only work.

## Out of Scope

Flipping rtmx 2.0 defaults in the same change without migrate + release notes.
