# REQ-DATA-001f: Decision gate for 2.0 document model

## Metadata
- **Category**: DATA
- **Subcategory**: DOCUMENT_MODEL
- **Priority**: HIGH
- **Phase**: 34
- **Status**: COMPLETE
- **Dependencies**: REQ-DATA-001a|REQ-DATA-001b|REQ-DATA-001c|REQ-DATA-001d|REQ-DATA-001e
- **Blocks**:
- **External ID**:
- **ADR**: docs/adr/0007-requirement-document-model.md

## Requirement

Record go/no-go for the structured document model as the path to AC-to-test
ATDD: canonical encoding (JSONL or otherwise), prose strategy, AC-level
completeness policy, and 2.0 migration CLI scope. Update ADR-0007 Status to
Accepted (or revise then Accept). Encoding choice is justified by ATDD
fitness, not by uniformity for its own sake.

## Rationale

Spikes without a decision leave 2.0 ambiguous. The gate converts evidence into
an implementation backlog or an explicit deferral.

## Acceptance Criteria

1. Written decision updates ADR-0007 Decision and Status (Accepted or revised+Accepted).
2. Decision answers all ADR-0007 spike-gate questions.
3. If go: implementation backlog sliced into follow-on requirements (IDs listed).
4. If no-go or partial: explicit deferral list and what remains v1.
5. Selection criterion for encoding is stated as AC-to-test ATDD / per-AC verify gaps—not “one file extension.”

## Test Strategy

Review checklist against ADR-0007 gate table; parent REQ-DATA-001 closes when
this gate and schema publication ACs are met.

## Out of Scope

Implementing the 2.0 default flip in the same change as the decision record.
