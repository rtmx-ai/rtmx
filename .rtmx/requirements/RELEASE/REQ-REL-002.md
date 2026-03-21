# REQ-REL-002: Software Bill of Materials (SBOM)

## Metadata
- **Category**: RELEASE
- **Subcategory**: Supply Chain
- **Priority**: HIGH
- **Phase**: 14
- **Status**: MISSING
- **Dependencies**: REQ-GO-043
- **Blocks**: REQ-GO-047

## Requirement

Each release shall include an SBOM in SPDX JSON format listing all dependencies, enabling enterprise supply chain auditing.

## Rationale

NIST EO 14028 and CISA guidance require SBOM for software supply chain security. Enterprises need to audit transitive dependencies for vulnerabilities and license compliance. RTMX has only 3 direct dependencies (cobra, viper, yaml.v3) which is excellent, but this must be electronically verifiable.

## Design

### GoReleaser Configuration

```yaml
sboms:
  - artifacts: archive
    documents:
      - "{{ .ArtifactName }}.sbom.spdx.json"
```

### Generated Artifact

```json
{
  "spdxVersion": "SPDX-2.3",
  "name": "rtmx-go",
  "packages": [
    {"name": "github.com/spf13/cobra", "version": "v1.8.0"},
    {"name": "gopkg.in/yaml.v3", "version": "v3.0.1"}
  ]
}
```

## Acceptance Criteria

1. SBOM generated for each release archive
2. SBOM in SPDX JSON format
3. Lists all direct and transitive dependencies
4. Attached to GitHub Release page
5. Machine-parseable by standard SBOM tools (e.g., `syft`, `grype`)

## Files to Modify

- `.goreleaser.yaml` - Add `sboms:` section
