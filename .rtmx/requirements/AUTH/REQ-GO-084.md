# REQ-GO-084: CLI Team Checkout command

## Metadata
- **Category**: AUTH
- **Subcategory**: Billing
- **Priority**: P0
- **Phase**: 8
- **Status**: COMPLETE
- **Dependencies**: REQ-GO-082|REQ-GO-083
- **External ID**: rtmx-ai/REQ-MONO-021c

## Requirement

Go CLI SHALL provide `rtmx checkout` (name may alias) that ensures
managed login, ensures an org, creates a Stripe Checkout Session via
the managed billing API, and opens the Stripe URL in the system browser.

## Acceptance Criteria

1. [x] Command opens Checkout URL (browser opener injectable).
2. [x] Triggers managed login when unauthenticated.
3. [x] Org via `--org` / create prompt; seats default 1, `--seats` override.
4. [x] Messaging: entitlement after Stripe/webhook, not on command return.
5. [x] HTTP client injectable; unit tests mock billing API.

## Effort Estimate

1.0 week
