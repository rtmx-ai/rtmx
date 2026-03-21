# REQ-CI-006: Reusable Validation Workflow

## Metadata
- **Category**: CI
- **Subcategory**: Ecosystem
- **Priority**: MEDIUM
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-012

## Requirement

Repository shall provide a reusable GitHub Actions workflow that other RTMX-enabled repositories can call to validate their RTM databases.

## Rationale

The Python repo provides `rtmx-validate.yml` as a `workflow_call` that other repos (rtmx-sync, rtmx.ai) can reference. This enables ecosystem-wide RTM validation without duplicating CI configuration. The Go CLI should provide an equivalent.

## Design

### New Workflow: `.github/workflows/rtmx-validate.yml`

```yaml
name: RTMX Validate
on:
  workflow_call:
    inputs:
      rtm-csv-path:
        description: 'Path to RTM database CSV'
        required: false
        default: '.rtmx/database.csv'
        type: string
      go-version:
        description: 'Go version'
        required: false
        default: '1.22'
        type: string
    outputs:
      status:
        description: 'Health status'
        value: ${{ jobs.validate.outputs.status }}
```

### Caller Usage (from other repos)

```yaml
# In rtmx-sync/.github/workflows/ci.yml
validate-rtm:
  uses: rtmx-ai/rtmx-go/.github/workflows/rtmx-validate.yml@main
  with:
    rtm-csv-path: '.rtmx/database.csv'
```

## Acceptance Criteria

1. Workflow callable via `workflow_call`
2. Configurable RTM database path
3. Outputs health status (healthy/degraded/unhealthy)
4. Uploads health report as artifact
5. Fails on unhealthy status

## Files to Create

- `.github/workflows/rtmx-validate.yml` - Reusable validation workflow
