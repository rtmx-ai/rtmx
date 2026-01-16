# REQ-WEB-009: Re-enable Security and Enterprise Pages

## Status: MISSING
## Priority: MEDIUM
## Phase: 10
## Effort: 0.5 weeks

## Description

The Security and Enterprise pages shall be re-enabled on rtmx.ai when the RTMX Sync service launches and security features are being actively implemented.

Currently these pages are disabled (renamed to `_security.astro` and `_enterprise.astro`) because they describe aspirational features that are not yet implemented. Re-enabling them prematurely would be misleading to users.

## Acceptance Criteria

- [ ] RTMX Sync service is deployed and operational (REQ-COLLAB-001)
- [ ] At least one security feature from the Security Roadmap is implemented
- [ ] Security page content is updated to reflect actual implemented vs planned features
- [ ] Enterprise page content is updated to reflect actual deployment options
- [ ] Pages are renamed from `_security.astro` to `security.astro` and `_enterprise.astro` to `enterprise.astro`
- [ ] Navigation links are restored in `astro.config.mjs`
- [ ] Pages are accessible at `/security` and `/enterprise`

## Test Cases

- `tests/test_website.py::test_security_page_accessible` - Security page returns 200
- `tests/test_website.py::test_enterprise_page_accessible` - Enterprise page returns 200
- `tests/test_website.py::test_security_content_accurate` - Security page reflects implemented features

## Technical Notes

Files to modify:
- `website/src/pages/_security.astro` → `website/src/pages/security.astro`
- `website/src/pages/_enterprise.astro` → `website/src/pages/enterprise.astro`
- `website/astro.config.mjs` - Restore nav links for Security and Enterprise

## Dependencies

- REQ-COLLAB-001: CRDT sync server operational

## Blocks

None
