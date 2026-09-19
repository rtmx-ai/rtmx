# REQ-DATA-001: Requirement document model for AC-to-test ATDD (rtmx 2.0)

## Metadata
- **Category**: DATA
- **Subcategory**: DOCUMENT_MODEL
- **Priority**: HIGH
- **Phase**: 34
- **Status**: COMPLETE
- **Dependencies**: REQ-DATA-001f
- **Blocks**:
- **External ID**:
- **ADR**: docs/adr/0007-requirement-document-model.md

## Requirement

RTMX shall define a structured Requirement Document Model that makes acceptance
criteria and test bindings first-class data, enabling AC-to-test ATDD and
per-AC verify gaps for verifiable agentic engineering. Canonical on-disk
encoding is selected only after spikes REQ-DATA-001a–001e and decision
REQ-DATA-001f. JSONL is a trade-space option, not a predetermined outcome.
Any breaking default storage change targets rtmx 2.0.0 with CSV read/import
compatibility and an explicit migration path.

## Rationale

Today COMPLETE means “something under this req_id passed,” not “each AC is
proven.” Agents and humans need addressable AC IDs bound to tests so gaps are
visible and ATDD loops are machine-checkable. Storage format is secondary to
that outcome (ADR-0007).

## Acceptance Criteria

1. ADR-0007 is Accepted (or revised and Accepted) with the recorded outcome of REQ-DATA-001f.
2. A published JSON Schema describes Requirement, AcceptanceCriterion, and TestBinding.
3. Spike report (001f) answers the ADR-0007 gate questions with ATDD fitness as the selection criterion.
4. No production default format flip without `rtmx migrate` (or equivalent) and CSV read-compat plan.
5. This parent does not block managed-sync / monetization launch requirements.

## Test Strategy

Documentation and gate review: ADR status, schema artifact path, and 001f
decision record. No production CLI change required to close this parent beyond
the spike deliverables.

## Out of Scope

- Implementing 2.0 storage defaults before 001f.
- Monorepo-wide migration before fixture spikes succeed.
- Changing sync-protocol-v1 for managed-sync launch.
