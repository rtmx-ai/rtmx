# REQ-DATA-002: Opt-in AC-matrix verify path

## Metadata
- **Category**: DATA
- **Subcategory**: DOCUMENT_MODEL
- **Priority**: HIGH
- **Phase**: 35
- **Status**: COMPLETE
- **Dependencies**: REQ-DATA-001
- **Blocks**:
- **External ID**:
- **ADR**: docs/adr/0007-requirement-document-model.md

## Requirement

Wire document-model AC evidence into an opt-in rtmx verify path using atdd-required-all-v0 without flipping default COMPLETE semantics.

## Acceptance Criteria

1. [ ] Traces to ADR-0007 Accepted decisions in DECISION_GATE_001f.
2. [ ] Tests bind `REQ-DATA-002`.
3. [ ] No managed-sync / sync-protocol-v1 break unless this is DATA-005 additive-only work.

## Out of Scope

Flipping rtmx 2.0 defaults in the same change without migrate + release notes.
