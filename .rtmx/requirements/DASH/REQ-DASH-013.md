# REQ-DASH-013: Remove External CDN Dependencies

## Summary

The dashboard shall load all CSS and JS from the embedded static filesystem,
eliminating runtime CDN dependencies (Tailwind CDN, htmx CDN, Alpine.js CDN,
D3 CDN) for air-gapped and offline deployment.

## Acceptance Criteria

1. `layout.html` contains no `<script src="https://...">` or `<link href="https://...">` tags
2. htmx, Alpine.js, D3, and Tailwind (or equivalent utility CSS) are vendored into
   `internal/dashboard/static/vendor/`
3. All `<script>` and `<link>` tags reference `/static/vendor/` paths
4. Dashboard functions correctly with no internet access
5. Binary size increase is documented (htmx ~15KB, Alpine ~17KB, D3 ~250KB, CSS ~50KB)
6. Go binary remains under 20MB total

## Dependencies

- REQ-DASH-010 (design tokens may replace need for full Tailwind)

## Files to Modify

- `internal/dashboard/templates/layout.html`
- `internal/dashboard/static/vendor/` (new directory, vendored assets)
- `internal/dashboard/embed.go` (embed directive may need update)

## Test Strategy

- Unit test: layout HTML contains no external URLs
- Unit test: all vendored files exist in embed.FS
- Integration test: dashboard serves correctly with no network
- Binary size check: `ls -la bin/rtmx` < 20MB

## Effort

- 0.5 weeks

## Priority

- MEDIUM

## Phase

- 30

## Notes

Air-gapped and self-managed deployments require zero external dependencies at runtime.
This is a hard requirement for defense/classified environments using Zarf packaging.
