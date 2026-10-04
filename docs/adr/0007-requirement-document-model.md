# ADR-0007: Requirement Document Model (AC-First ATDD)

## Status

Accepted

## Context

ADR-0001 chose CSV for the RTM spine (git diffs, Sheets, agent parseability).
ADR-0004 and the VERIFY/LANG families link tests to requirements primarily at
`req_id` granularity (`test_function`, markers, results JSON).

In practice:

1. CSV holds one row per requirement and at most one primary `test_function`.
2. Detailed acceptance criteria live in Markdown (`requirement_file`) as prose lists.
3. `rtmx verify` promotes or demotes at requirement level, not per AC.
4. CRDT sync (`sync-protocol-v1`) models mostly scalar requirement fields;
   nested AC arrays and long narrative are not first-class on the wire.

### Guiding principle

**Acceptance criteria become first-class data; verify reports per-AC gaps.**

### Primary outcome

**AC-to-test ATDD for verifiable agentic engineering:** each AC is an
addressable ID with explicit test bindings so agents and humans can see which
behaviors are proven, which are missing, and what remains before a requirement
may be COMPLETE.

Storage encoding is a means. Selection criterion is **AC-to-test ATDD / per-AC
verify gaps**, not preference for a single file extension.

Spikes REQ-DATA-001a–001e completed; decision record:
`docs/schemas/DECISION_GATE_001f.md` (REQ-DATA-001f).

## Decision

1. Introduce a **Requirement Document Model** with explicit entities:
   - `Requirement` (identity, status, deps, priority, …)
   - `AcceptanceCriterion` (`ac_id`, statement, optional severity / required flag)
   - `TestBinding` (`ac_id` and/or `req_id` → test identity / marker / results key)
2. Treat structured AC and binding fields as the **source of truth for verification**.
3. Treat narrative (rationale, implementation history) as **optional prose** in a
   **Markdown companion** (`narrative_path` / `requirement_file`), not embedded in
   the default agent/verify projection.
4. **Canonical structured on-disk encoding for rtmx 2.0.0:** JSONL (NDJSON)
   records holding identity, `acs[]`, and `test_bindings[]`, plus optional MD
   companion — justified by the 001c authoring scorecard (ATDD maintainability).
5. **Default completeness policy** for AC-aware verify: `atdd-required-all-v0`
   (`docs/schemas/ATDD_POLICY_v0.md`).
6. Target **rtmx 2.0.0** for any breaking default storage change.
7. **v1 compatibility for at least one major cycle:** read CSV projects; export
   CSV projection; migrate CSV+MD → structured store via `rtmx migrate`.
8. **Sync:** keep sync-protocol-v1 for managed-sync launch; add nested ACs later
   via an **additive** extension (see `SYNC_MCP_IMPACT_SPIKE.md`). No silent drop
   of ACs once clients advertise the capability.
9. Published schema: `docs/schemas/requirement-document-v0.schema.json`.

## Non-goals (this ADR)

- Replacing sync-protocol-v1 in the monetization / managed-sync launch path.
- Monorepo-wide migration before follow-on implementation requirements ship.
- Dropping Markdown GitHub rendering.
- Flipping default `rtmx verify` to AC-complete before REQ-DATA-002.

## Spike gate

| Question | Pass criteria | Status |
|---|---|---|
| Schema | JSON Schema validates fixture docs; ACs have stable IDs | **Met** (001a) |
| Verify | AC-level evidence can drive status under a documented ATDD policy | **Met** (001b) |
| Authoring | Humans/agents can edit without pathological diffs | **Met** (001c) |
| Migration | Round-trip CSV+MD → structured → CSV projection is idempotent on identity/deps/status | **Met** (001d) |
| Sync | Sketch of protocol delta; no silent data loss for nested ACs | **Met** (001e) |
| MCP/token | Projection exists that omits bulk narrative while retaining AC+binding surface | **Met** (001c/001e) |

## Alternatives considered

### A. Keep CSV + Markdown; add AC sidecar only
- **Pro**: Smallest change; ADR-0001 intact.
- **Con**: Three artifacts to keep consistent; easy drift between MD lists and sidecar.
- **Outcome**: Rejected as sole 2.0 model; CSV remains v1 + import/export.

### B. All-JSONL (RTM + full prose in every line)
- **Pro**: One parser; specs are data.
- **Con**: Huge lines, weak prose review UX, CRDT/token pressure.
- **Outcome**: Rejected for prose; structured JSONL **without** bulk rationale wins.

### C. Markdown with YAML frontmatter as canonical
- **Pro**: Familiar authoring; GitHub renders body.
- **Con**: Frontmatter drift; weaker blame for status churn; dual list ID risk.
- **Outcome**: Runner-up / import-export view only (001c).

### D. SQLite / embedded DB
- Rejected for primary store (ADR-0001 still holds for git/Sheets).
- Allowed later as *derived* cache/export only.

## Consequences

### Positive
- ACs become first-class; verify can report per-AC gaps.
- Agents can drive ATDD loops against concrete AC IDs.
- Clear MAJOR boundary and migration story.
- Does not block managed-sync launch (CSV remains production path).

### Negative / risks
- Dual-write period and dual readers until 2.0 default flips.
- Sync schema needs an additive version for nested ACs.
- Completeness policy is more complex (req-level vs AC-level).

### Mitigations
- Follow-ons REQ-DATA-002–006 implement verify, migrate, IO, sync, MCP.
- CSV remains import/export and v1 default until 2.0.0.
- Completeness policy remains explicit and configurable.

## Follow-on requirements

REQ-DATA-002 (AC verify opt-in), REQ-DATA-003 (migrate), REQ-DATA-004 (structured
IO), REQ-DATA-005 (sync 1.1-ac), REQ-DATA-006 (MCP AC projections).

## References

- ADR-0001 (CSV over SQLite)
- ADR-0004 (markers / req_id linking)
- system/contracts/sync-protocol-v1.json (`requirement_fields`)
- docs/schemas/DECISION_GATE_001f.md
- REQ-DATA-001 through REQ-DATA-001f; REQ-DATA-002–006
