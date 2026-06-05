# REQ-VERIFY-010: Configurable Test-Dimension Vocabularies

## Metadata
- **Category**: VERIFY
- **Subcategory**: COMPLETENESS
- **Priority**: MEDIUM
- **Phase**: 29
- **Status**: MISSING
- **Dependencies**: REQ-VERIFY-009
- **Blocks**:
- **External ID**:

## Requirement

The RTMX results validator shall accept project-configured vocabularies for the
`scope`, `technique`, and `env` marker dimensions, supplied via
`rtmx.completeness.vocabulary` in config. When a dimension's vocabulary is set,
its values augment the built-in accepted set for that dimension; when unset, the
built-in vocabulary is used unchanged.

## Rationale

The results schema validates dimension values against fixed built-in
vocabularies (`scope`: unit/integration/system/acceptance; `technique`:
nominal/parametric/monte_carlo/stress/boundary; `env`:
simulation/hil/anechoic/field). Schemas that legitimately use finer-grained
values are rejected — for example the `phoenix` schema splits `env` into
`static_field` and `dynamic_field`, which the built-in `field`-only vocabulary
does not accept. Custom schemas defined via `.rtmx/schema.yaml` have the same
need. Making the vocabularies configurable lets these projects emit and validate
their own dimension values while preserving the safe built-in defaults for
everyone else. This is the validation-layer complement to REQ-VERIFY-009.

## Acceptance Criteria

- [ ] Configured vocabulary values validate without error.
- [ ] Built-in vocabularies remain accepted when a dimension is unset.
- [ ] An unset configuration reproduces prior validation behavior exactly.
- [ ] Verified by `internal/results/schema_vocab_test.go::TestValidateWithVocabulary`.
