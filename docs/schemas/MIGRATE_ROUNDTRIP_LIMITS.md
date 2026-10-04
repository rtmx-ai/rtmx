# Migration round-trip limits (REQ-DATA-001d AC5)

What **does** round-trip in this spike:

- `req_id`, `requirement_text`, `status`, `priority`
- `dependencies` / `blocks` (pipe-joined in CSV projection)
- AC `statement` text + stable `ac_id` (ordinal-AC-N) when statement unchanged

What **does not** round-trip:

- `## Implementation History`, `## Rationale`, and other free-form Markdown sections
- Checkbox checked/unchecked state (`[x]` vs `[ ]`) — only the statement text is kept
- Nested numbering depth / original list style (checkbox vs numbered)
- CSV columns not in the spike projection (`test_module`, assignee, dates, notes, …)
- Live `.rtmx/database.csv` tree migration (explicitly out of scope)

Production `rtmx migrate` UX polish is out of scope for this requirement.
