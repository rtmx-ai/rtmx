# REQ-DEPLOY-003: Air-Gap Deployment (Zarf)

## Requirement
RTMX shall support fully air-gapped deployment via Zarf package.

## Status: MISSING
## Priority: HIGH
## Phase: 13

## Rationale
Classified and high-security environments operate without internet connectivity. Zarf provides a proven, DoD-approved method for deploying containerized applications into air-gapped Kubernetes clusters. This enables RTMX adoption in SCIF environments, classified networks, and disconnected operational environments.

## Acceptance Criteria
- [ ] Zarf package includes all dependencies
- [ ] No external network calls required
- [ ] SBOM included in package
- [ ] Iron Bank base images used where available
- [ ] Installation on k3s/RKE2 tested
- [ ] Package signed with verifiable signature
- [ ] Deployment tested on Big Bang cluster

## Zarf Package Structure

```
rtmx-zarf-package/
├── zarf.yaml                 # Package definition
├── manifests/
│   ├── namespace.yaml
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── configmap.yaml
│   └── secrets.yaml
├── images/
│   ├── rtmx-server.tar       # Container image
│   ├── postgres.tar          # Database image
│   └── busybox.tar           # Init container
├── charts/
│   └── rtmx/                 # Helm chart
├── sbom/
│   ├── rtmx-server.spdx.json
│   └── postgres.spdx.json
└── README.md
```

## Container Images

### Iron Bank Images (Preferred)
| Component | Iron Bank Image | Fallback |
|-----------|-----------------|----------|
| PostgreSQL | registry1.dso.mil/ironbank/opensource/postgres/postgresql16 | postgres:16-alpine |
| Init Container | registry1.dso.mil/ironbank/redhat/ubi/ubi9-minimal | busybox:stable |
| Base Image | registry1.dso.mil/ironbank/redhat/ubi/ubi9 | python:3.12-slim |

### RTMX Images
| Image | Description |
|-------|-------------|
| rtmx-server | Main RTMX server application |
| rtmx-sync | CRDT sync server |
| rtmx-web | Web dashboard |

## SBOM Requirements

Each container image includes:
- SPDX 2.3 format SBOM
- Complete dependency tree
- CVE scan results at build time
- Cryptographic hash of all components

## Deployment Targets

### k3s
```bash
# Install Zarf
zarf init

# Deploy RTMX package
zarf package deploy rtmx-vX.Y.Z-amd64.tar.zst --confirm
```

### RKE2
```bash
# With Big Bang installed
zarf package deploy rtmx-vX.Y.Z-amd64.tar.zst \
  --set ISTIO_ENABLED=true \
  --confirm
```

## Test Cases
1. Zarf package builds without network access
2. Package deploys to clean k3s cluster
3. Package deploys to RKE2 with Big Bang
4. All functionality works without internet
5. SBOM validates against known CVE database
6. Package signature verifies correctly
7. Update package applies cleanly

## Package Signing

```bash
# Sign package with GPG key
zarf package sign rtmx-vX.Y.Z-amd64.tar.zst \
  --key /path/to/signing-key.asc

# Verify signature before deployment
zarf package verify rtmx-vX.Y.Z-amd64.tar.zst \
  --key /path/to/public-key.asc
```

## Technical Notes
- Package size target: < 500 MB compressed
- Support both amd64 and arm64 architectures
- Include database migration scripts in package
- Offline license validation required (REQ-DEPLOY-002)
- FIPS mode supported (REQ-SEC-008)

## Big Bang Integration

When deployed on Big Bang:
- Istio service mesh integration
- Kiali observability
- Jaeger tracing
- Keycloak SSO integration
- Policy enforcement via Gatekeeper

## Update Procedure

```bash
# Export data from running instance
rtmx export --format json > backup.json

# Deploy new package version
zarf package deploy rtmx-vX.Y.Z-amd64.tar.zst --confirm

# Verify deployment
kubectl get pods -n rtmx

# Import data if needed
rtmx import backup.json
```

## Dependencies
- REQ-COMPL-005 (SBOM generation - to be created)

## Blocks
- None

## Effort
4.0 weeks
