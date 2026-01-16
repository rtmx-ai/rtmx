# REQ-COMPL-005: Software Bill of Materials

## Requirement
RTMX shall generate CISA-compliant Software Bill of Materials.

## Phase
13 (Security/Compliance)

## Rationale
Executive Order 14028 and subsequent CISA guidance require SBOMs for software sold to the federal government. SBOMs enable vulnerability tracking, license compliance, and supply chain risk management. RTMX must generate SBOMs that meet the minimum requirements defined by NTIA and enhanced requirements from CISA's 2025 guidance.

## Acceptance Criteria
- [ ] SBOM generation in SPDX 2.3 format
- [ ] SBOM generation in CycloneDX 1.5+ format
- [ ] Component hash included per CISA 2025 requirements (SHA-256 minimum)
- [ ] License information for all direct and transitive dependencies
- [ ] Automated SBOM generation in CI/CD pipeline
- [ ] SBOM signed with project key (ML-DSA when REQ-SEC-012 complete)
- [ ] SBOM published with each release
- [ ] VEX (Vulnerability Exploitability eXchange) document capability
- [ ] PURL (Package URL) identifiers for all components

## NTIA Minimum Elements

| Element | Description | RTMX Implementation |
|---------|-------------|---------------------|
| Supplier Name | Entity that creates/distributes | ioTactical LLC |
| Component Name | Name of software component | From pyproject.toml/requirements |
| Version | Version identifier | Semantic version string |
| Unique Identifier | Globally unique ID | PURL + SHA-256 hash |
| Dependency Relationship | Upstream/downstream | Direct vs transitive |
| Author | Who created the data | Automated tooling |
| Timestamp | When SBOM was generated | ISO 8601 format |

## CISA 2025 Enhanced Requirements

| Element | Description |
|---------|-------------|
| Component Hash | SHA-256 of component package |
| Build Provenance | SLSA provenance attestation |
| Source Repository | Link to source code |
| Known Vulnerabilities | CVE/OSV identifiers at generation time |
| End of Life | Support/maintenance status |

## SBOM Generation Pipeline

```yaml
# .github/workflows/sbom.yml
name: SBOM Generation
on:
  release:
    types: [published]

jobs:
  sbom:
    runs-on: ubuntu-latest
    steps:
      - name: Generate SPDX SBOM
        uses: anchore/sbom-action@v0
        with:
          format: spdx-json
          output-file: rtmx-${{ github.ref_name }}.spdx.json

      - name: Generate CycloneDX SBOM
        uses: CycloneDX/gh-python-generate-sbom@v2
        with:
          input: pyproject.toml
          output: rtmx-${{ github.ref_name }}.cdx.json

      - name: Sign SBOM
        run: |
          cosign sign-blob --key cosign.key rtmx-*.json

      - name: Publish SBOM
        uses: softprops/action-gh-release@v1
        with:
          files: |
            rtmx-*.spdx.json
            rtmx-*.cdx.json
            rtmx-*.sig
```

## CLI Commands

```bash
# Generate SBOM for current project
rtmx sbom generate --format spdx

# Generate signed SBOM
rtmx sbom generate --format cyclonedx --sign

# Verify SBOM signature
rtmx sbom verify rtmx-1.0.0.cdx.json

# Generate VEX for known vulnerabilities
rtmx sbom vex --output rtmx-1.0.0.vex.json
```

## Technical Notes
- Use `cyclonedx-python` library for CycloneDX generation
- Use `spdx-tools` for SPDX generation
- Hash all wheel files and source distributions
- Include Python runtime version as component
- Document SBOM generation process for reproducibility
- Store SBOMs in `docs/sbom/` directory versioned with releases

## Sample SBOM Entry (CycloneDX)

```json
{
  "bomFormat": "CycloneDX",
  "specVersion": "1.5",
  "version": 1,
  "components": [
    {
      "type": "library",
      "name": "click",
      "version": "8.1.7",
      "purl": "pkg:pypi/click@8.1.7",
      "hashes": [
        {
          "alg": "SHA-256",
          "content": "ae74fb96c20a0277a1d615f1e4d73c8414f5a98db8b799a7931d1582f3390c28"
        }
      ],
      "licenses": [
        {
          "license": {
            "id": "BSD-3-Clause"
          }
        }
      ]
    }
  ]
}
```

## Test Cases
1. Verify SPDX SBOM passes NTIA minimum checker
2. Verify CycloneDX SBOM validates against schema
3. Verify all dependencies have hashes
4. Verify licenses are populated for all components
5. Verify SBOM signature validates with public key
6. Verify SBOM is generated in CI on release

## Dependencies
- REQ-SEC-012 (ML-DSA for signing) - for post-quantum signatures

## Effort
2.0 weeks
