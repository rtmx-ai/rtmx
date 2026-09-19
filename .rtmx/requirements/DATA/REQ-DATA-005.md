# REQ-DATA-005: Additive sync 1.1-ac maps

## Metadata
- **Category**: DATA
- **Subcategory**: DOCUMENT_MODEL
- **Priority**: MEDIUM
- **Phase**: 35
- **Status**: COMPLETE
- **Dependencies**: REQ-DATA-001e|REQ-DATA-001
- **Blocks**:
- **External ID**:
- **ADR**: docs/adr/0007-requirement-document-model.md

## Requirement

Implement additive requirement_acs and requirement_bindings CRDT maps behind a capability flag; v1 clients remain compatible.

## Acceptance Criteria

1. [ ] Traces to ADR-0007 Accepted decisions in DECISION_GATE_001f.
2. [ ] Tests bind `REQ-DATA-005`.
3. [ ] No managed-sync / sync-protocol-v1 break unless this is DATA-005 additive-only work.

## Out of Scope

Flipping rtmx 2.0 defaults in the same change without migrate + release notes.
