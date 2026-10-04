# REQ-DATA-006: MCP AC-gap projections

## Metadata
- **Category**: DATA
- **Subcategory**: DOCUMENT_MODEL
- **Priority**: MEDIUM
- **Phase**: 35
- **Status**: COMPLETE
- **Dependencies**: REQ-DATA-001|REQ-DATA-002
- **Blocks**:
- **External ID**:
- **ADR**: docs/adr/0007-requirement-document-model.md

## Requirement

Extend MCP backlog/verify-style payloads with AC gaps and bindings while omitting rationale by default.

## Acceptance Criteria

1. [ ] Traces to ADR-0007 Accepted decisions in DECISION_GATE_001f.
2. [ ] Tests bind `REQ-DATA-006`.
3. [ ] No managed-sync / sync-protocol-v1 break unless this is DATA-005 additive-only work.

## Out of Scope

Flipping rtmx 2.0 defaults in the same change without migrate + release notes.
