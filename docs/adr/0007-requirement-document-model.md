# ADR-0007: Requirement Document Model (AC-First ATDD)

## Status

Proposed

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

Storage encoding (JSONL, Markdown+frontmatter, CSV+sidecar, etc.) is a means
in the trade space. Uniform *schema* matters more than a single file extension.
JSONL is a leading candidate for structured records, not a predetermined end state.

This ADR does **not** authorize a wholesale format cutover. It authorizes a
document model and a spike series that must complete before any 2.0.0
implementation.

## Decision (provisional)

1. Introduce a **Requirement Document Model** with explicit entities:
   - `Requirement` (identity, status, deps, priority, …)
   - `AcceptanceCriterion` (`ac_id`, statement, optional severity / required flag)
   - `TestBinding` (`ac_id` and/or `req_id` → test identity / marker / results key)
2. Treat structured AC and binding fields as the **source of truth for verification**.
3. Treat narrative (rationale, implementation history) as **optional prose** —
   companion Markdown path or embedded Markdown string — chosen by spike.
4. Target **rtmx 2.0.0** for any breaking default storage change.
5. **v1 compatibility for at least one major cycle:** read CSV projects; export
   CSV projection; migrate CSV+MD → structured store via `rtmx migrate`.
6. Canonical on-disk encoding for the structured store is **TBD by spikes**
   (JSONL is in the trade space, not assumed).

## Non-goals (this ADR)

- Replacing sync-protocol-v1 in the monetization / managed-sync launch path.
- Migrating the monorepo RTM before spikes conclude.
- Dropping Markdown GitHub rendering without an authoring story.
- Making COMPLETE mean “all ACs green” without an explicit completeness policy.
- Choosing JSONL (or any encoding) before the spike gate in REQ-DATA-001f.

## Spike gate

Before implementing 2.0 storage defaults, spikes REQ-DATA-001a–001e must produce
a written decision in REQ-DATA-001f covering:

| Question | Pass criteria |
|---|---|
| Schema | JSON Schema validates fixture docs; ACs have stable IDs |
| Verify | AC-level evidence can drive status under a documented ATDD policy |
| Authoring | Humans/agents can edit without pathological diffs |
| Migration | Round-trip CSV+MD → structured → CSV projection is idempotent on identity/deps/status |
| Sync | Sketch of protocol delta; no silent data loss for nested ACs |
| MCP/token | Projection exists that omits bulk narrative while retaining AC+binding surface |

## Alternatives considered

### A. Keep CSV + Markdown; add AC sidecar only
- **Pro**: Smallest change; ADR-0001 intact.
- **Con**: Three artifacts to keep consistent; easy drift between MD lists and sidecar.

### B. All-JSONL (RTM + full prose in every line)
- **Pro**: One parser; specs are data.
- **Con**: Huge lines, weak prose review UX, CRDT/token pressure.
- **Note**: May still win for *structured* fields even if prose stays elsewhere.

### C. Markdown with YAML frontmatter as canonical
- **Pro**: Familiar authoring; GitHub renders body.
- **Con**: Frontmatter drift; weaker “one line = one req” blame for status churn.

### D. SQLite / embedded DB
- Rejected for primary store (ADR-0001 still holds for git/Sheets).
- Allowed later as *derived* cache/export only.

**Preferred exploration path:** structured records (JSONL candidate) + prose
companion (Markdown candidate) unless spikes prove a single-surface approach
superior for AC-to-test ATDD.

## Consequences

### Positive
- ACs become first-class; verify can report per-AC gaps.
- Agents can drive ATDD loops against concrete AC IDs, not only req-level COMPLETE.
- Clear MAJOR boundary and migration story.
- Does not block managed-sync launch (CSV remains production path).

### Negative / risks
- Dual-write period and dual readers until 2.0 default flips.
- Sync schema likely needs a version bump for nested ACs.
- Completeness policy becomes more complex (req-level vs AC-level).

### Mitigations
- Spikes on fixtures, not monorepo-wide migration.
- CSV remains import/export and v1 default until gate passes.
- Completeness policy remains explicit and configurable.
- Encoding choice deferred to REQ-DATA-001f with ATDD outcome as the selection criterion.

## References

- ADR-0001 (CSV over SQLite)
- ADR-0004 (markers / req_id linking)
- system/contracts/sync-protocol-v1.json (`requirement_fields`)
- REQ-DATA-001 through REQ-DATA-001f
