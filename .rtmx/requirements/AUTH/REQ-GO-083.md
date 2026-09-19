# REQ-GO-083: Sync uses stored managed auth

## Metadata
- **Category**: AUTH
- **Subcategory**: Managed
- **Priority**: P0
- **Phase**: 8
- **Status**: MISSING
- **Dependencies**: REQ-GO-082
- **External ID**: rtmx-ai/REQ-MONO-021b

## Requirement

Go CLI sync commands SHALL use the stored managed credential for the
managed sync host without requiring `--token` paste in the happy path.

## Acceptance Criteria

1. [ ] Stored session authenticates sync to managed host.
2. [ ] Clear error + `rtmx login` hint when credential missing/expired.
3. [ ] `--token` / env override still wins when set.
4. [ ] Tests cover stored vs override paths.

## Effort Estimate

0.75 weeks
