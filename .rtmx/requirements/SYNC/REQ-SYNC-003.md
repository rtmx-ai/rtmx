# REQ-SYNC-003: CLI Connects With Issued Managed Credentials

## Metadata
- **Category**: SYNC
- **Subcategory**: CRDT
- **Priority**: P0
- **Phase**: 33
- **Status**: MISSING
- **Dependencies**: REQ-SYNC-002
- **Blocks**: REQ-MONO-011e
- **External ID**: rtmx-ai/REQ-MONO-011e

## Requirement

The Go CLI shall join a managed rtmx-sync room using the sync URL and
token issued at PaaS fulfillment (`rtmx sync --sync-url --token`), and
shall persist merged RTM state to CSV. Behavior is the same client as
REQ-SYNC-002; this requirement is that fulfillment credentials are
sufficient with no extra hidden config.

## Rationale

RoomClient already exists. Public Checkout is not complete if the
success page prints a URL the CLI cannot use, or if a hosted URL needs
undocumented flags.

## Acceptance Criteria

1. `--sync-url` plus `--token` (or `RTMX_SYNC_TOKEN`) authenticates to
   `/sync/{org}/{room}` on the issued host.
2. A write and a subsequent pull from a second process converge.
3. 4402 after dunning is reported as entitlement, not a generic close.

## Test Strategy

Monorepo PaaS fulfillment scenario (REQ-MONO-011e). Unit coverage is
REQ-SYNC-002's RoomClient tests.
