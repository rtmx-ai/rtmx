# REQ-DATA-001e: Sync and MCP impact spike

## Metadata
- **Category**: DATA
- **Subcategory**: DOCUMENT_MODEL
- **Priority**: MEDIUM
- **Phase**: 34
- **Status**: COMPLETE
- **Dependencies**: REQ-DATA-001a|REQ-DATA-001b
- **Blocks**: REQ-DATA-001f
- **External ID**:
- **ADR**: docs/adr/0007-requirement-document-model.md

## Requirement

Document how nested ACs and test bindings would appear in CRDT sync and MCP
projections; sketch sync-protocol-v2 or an additive `schema_version` extension
without implementing a server cutover. Preserve the constraint that managed-sync
launch does not require breaking sync-protocol-v1.

## Rationale

AC-first ATDD is incomplete if collaborative sync silently drops ACs or if MCP
tools blow token budgets on narrative while omitting the AC/binding surface
agents need.

## Acceptance Criteria

1. Gap analysis versus `system/contracts/sync-protocol-v1.json` `requirement_fields`.
2. Proposal chooses whether to sync AC status, full AC text, bindings, or a subset—with bandwidth and conflict tradeoffs.
3. MCP projection sketch: backlog/verify-style payloads that include AC gaps without bulk rationale.
4. Explicit statement: no breaking sync change required for managed-sync launch.
5. Sketch only—no production rtmx-sync schema migration in this spike.

## Test Strategy

Design review artifact checked in under `docs/` (or spike notes). Optional
contract stub JSON is allowed but not required to be enforced in CI yet.

## Out of Scope

Implementing Yjs nested types in production; changing live room schema.
