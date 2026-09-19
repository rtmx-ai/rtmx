# Stable `ac_id` scheme (REQ-DATA-001a AC4)

## Scheme name: `ordinal-AC-N`

Within a single `req_id`, acceptance criteria use identifiers of the form
`AC-1`, `AC-2`, … assigned in document order at first materialization.

## Stability rule

When rematerializing a requirement document (for example CSV+Markdown →
structured → CSV projection), an `ac_id` **MUST** be reused if the AC
`statement` text is unchanged (Unicode NFC, trimmed trailing whitespace).

If statement text changes, the rematerializer MAY:

1. Keep the same `ac_id` and treat it as an edit of that criterion, or
2. Retire the old `ac_id` and allocate the next ordinal for a new statement

Spikes MUST document which policy they implement; the default for this v0
schema is **(1) edit-in-place** when a single AC maps 1:1 by position and
text differs, and **reuse by identical statement text** when order shuffles.

## Content-hash policy (deferred)

A content-hash scheme (`AC-` + short hash of statement) remains an option for
REQ-DATA-001f if ordinal collisions appear in migration spikes. It is **not**
required to close REQ-DATA-001a.

## Schema reference

The `ac_id` property description on
`docs/schemas/requirement-document-v0.schema.json` names this scheme.
