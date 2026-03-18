# REQ-PAR-005: Config Label Format Compatibility

## Metadata
- **Category**: PARITY
- **Subcategory**: Config
- **Priority**: HIGH
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-008

## Requirement

Go CLI shall parse Python-format config files without breaking, specifically the `labels` field which uses list format in Python but struct format in Go.

## Problem

Python config:
```yaml
github:
  labels: ["requirement", "feature"]
```

Go config:
```yaml
github:
  labels:
    requirement: "requirement"
```

Existing Python users' config files will fail to parse in Go.

## Acceptance Criteria

1. Go CLI parses both list and struct label formats
2. Existing Python config files load without error
3. Go writes labels in struct format (forward-looking)
4. Warning emitted if legacy list format detected

## Files to Modify

- `internal/config/config.go` - Add flexible label parsing
- `internal/config/config_test.go` - Both format tests
