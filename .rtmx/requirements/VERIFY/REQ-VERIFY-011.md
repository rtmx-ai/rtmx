# REQ-VERIFY-011: Configurable Requirement-ID Pattern

## Metadata
- **Category**: VERIFY
- **Subcategory**: COMPLETENESS
- **Priority**: MEDIUM
- **Phase**: 29
- **Status**: MISSING
- **Dependencies**: REQ-VERIFY-010
- **Blocks**:
- **External ID**:

## Requirement

The RTMX results validator shall accept a project-configured requirement-ID
regular expression, supplied via `rtmx.req_id_pattern` in config. When set, the
configured pattern replaces the built-in pattern for results validation; when
unset, the built-in default is used. The built-in default shall accept a
single- or multi-segment uppercase-alphanumeric category prefix followed by a
numeric index (e.g. `REQ-SW-009`, `REQ-E2E-010`, `REQ-INFRA-DT-002`,
`REQ-MODE-S-006`).

## Rationale

The results schema validated requirement IDs against a fixed pattern
`^REQ-[A-Z]+-[0-9]+$`, which accepts only a single alphabetic category segment.
This rejects two legitimate and common conventions: alphanumeric categories
that the adapters already support (e.g. `REQ-E2E-010`) and multi-segment
category prefixes that hierarchical projects use (e.g. the Phoenix radar RTM's
`REQ-INFRA-DT-002`, `REQ-SW-DSP-015`, and `REQ-MODE-S-006`). Because the pattern
is enforced as a hard validation error in the cross-language `rtmx verify
--results` path, such projects could never close the verification loop, even
though their markers were otherwise well-formed.

Broadening the built-in default fixes the common case with zero configuration,
and exposing the pattern in config (the validation-layer complement to the
configurable dimension vocabularies of REQ-VERIFY-010) lets projects with any
other identifier convention define their own grammar without forking the tool.

## Acceptance Criteria

- [ ] The built-in default accepts single- and multi-segment category prefixes
      (`REQ-SW-009`, `REQ-INFRA-DT-002`, `REQ-MODE-S-006`) and still rejects
      malformed IDs (lowercase, missing number, digit-leading category).
- [ ] A configured `rtmx.req_id_pattern` overrides the default for results
      validation.
- [ ] An unset pattern reproduces the default validation behavior exactly.
- [ ] An invalid configured pattern is reported as a single validation error
      and falls back to the default rather than aborting.
- [ ] Verified by `internal/results/schema_reqid_test.go::TestDefaultReqIDPattern`
      and `::TestConfigurableReqIDPattern`.
