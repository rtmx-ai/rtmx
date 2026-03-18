# REQ-E2E-001: Local Development Stack via Zarf + Kind

## Status: MISSING
## Priority: HIGH
## Phase: 10

## Description

RTMX shall provide a Zarf package that deploys all platform components to a local Kind (Kubernetes-in-Docker) cluster for end-to-end development and testing. The package shall include Zitadel (identity provider), CockroachDB (Zitadel backing store), OpenZiti (zero-trust network controller and router), and rtmx-sync (coordination service). Using Zarf for local development ensures what we test locally is exactly what ships to air-gapped and production environments.

## Rationale

- **Single deployment tool** — same `zarf package deploy` for local dev, CI, and air-gapped customer sites
- **What you test is what you ship** — eliminates Docker Compose ↔ Kubernetes divergence
- **Validates the Zarf package itself** — every local dev session tests the deployment artifact
- **No cloud costs** — Kind runs a full K8s cluster on a single machine
- **Air-gapped ready** — Zarf packages are self-contained tarballs with all images vendored
- Aligns with Defense Industrial Base deployment expectations

## Acceptance Criteria

- [ ] `make e2e-up` creates Kind cluster and deploys Zarf package
- [ ] `make e2e-down` deletes Kind cluster and cleans up
- [ ] `make e2e-test` runs the E2E test suite against the local cluster
- [ ] `make e2e-status` shows health of all deployed components
- [ ] `make zarf-build` builds the Zarf package tarball
- [ ] Zarf package (`zarf-package-rtmx-platform-*.tar.zst`) includes:
  ```yaml
  # zarf.yaml
  kind: ZarfPackageConfig
  metadata:
    name: rtmx-platform
    description: RTMX real-time collaboration platform
  components:
    - name: crdb
      description: CockroachDB (Zitadel backing store)
      charts:
        - name: cockroachdb
      images:
        - cockroachdb/cockroach:latest

    - name: zitadel
      description: Identity provider (OIDC)
      charts:
        - name: zitadel
      images:
        - ghcr.io/zitadel/zitadel:latest

    - name: openziti
      description: Zero-trust network overlay
      charts:
        - name: ziti-controller
        - name: ziti-router
      images:
        - openziti/ziti-controller:latest
        - openziti/ziti-router:latest

    - name: rtmx-sync
      description: Real-time coordination service
      charts:
        - name: rtmx-sync
      images:
        - ghcr.io/rtmx-ai/rtmx-sync:latest
  ```
- [ ] Stack reaches healthy state within 120 seconds on first deploy
- [ ] Auto-provisioning via Zarf actions configures:
  - Zitadel project, OIDC application (rtmx-cli), and test users
  - OpenZiti identities, services, and service policies
  - rtmx-sync service registration with Ziti
- [ ] Host-based rtmx CLI can authenticate against port-forwarded Zitadel
- [ ] Host-based rtmx CLI can sync via local OpenZiti overlay
- [ ] Health checks on all pods with proper readiness/liveness probes
- [ ] Works on Linux and macOS (Docker Desktop or Colima for Kind)
- [ ] Stack resource usage under 4GB RAM total
- [ ] Package size under 2GB compressed (all images vendored)

## Directory Structure

```
e2e/
├── zarf.yaml                 # Zarf package definition
├── kind-config.yaml          # Kind cluster configuration (port mappings)
├── charts/
│   ├── rtmx-sync/            # Helm chart for rtmx-sync
│   │   ├── Chart.yaml
│   │   ├── values.yaml
│   │   └── templates/
│   ├── zitadel-overrides.yaml  # Zitadel Helm values for local dev
│   ├── ziti-overrides.yaml     # OpenZiti Helm values for local dev
│   └── crdb-overrides.yaml     # CockroachDB Helm values for local dev
├── actions/
│   ├── zitadel-seed.py       # Provision OIDC project/app/users via Zitadel API
│   ├── ziti-seed.sh          # Provision Ziti identities/services/policies
│   └── wait-for-healthy.sh   # Health check orchestration
├── fixtures/
│   ├── test-users.json       # Test user definitions
│   ├── test-policies.json    # Ziti service policies
│   └── test-database.csv     # Sample RTM database for sync testing
└── Makefile                  # e2e-up, e2e-down, e2e-test, e2e-status, zarf-build
```

## Test Cases

1. `tests/e2e/test_stack_health.py::test_all_pods_healthy` - All pods running and ready
2. `tests/e2e/test_stack_health.py::test_stack_boot_time` - Stack ready within 120s
3. `tests/e2e/test_stack_health.py::test_resource_usage` - Under 4GB RAM
4. `tests/e2e/test_stack_health.py::test_zarf_package_size` - Under 2GB compressed
5. `tests/e2e/test_auth_e2e.py::test_cli_login_local_zitadel` - OIDC PKCE against local Zitadel
6. `tests/e2e/test_auth_e2e.py::test_cli_logout` - Token cleanup
7. `tests/e2e/test_sync_e2e.py::test_cli_sync_via_ziti` - Sync through OpenZiti overlay
8. `tests/e2e/test_sync_e2e.py::test_concurrent_sync` - Multiple CLI instances syncing
9. `tests/e2e/test_provisioning.py::test_auto_seed_zitadel` - OIDC app provisioned correctly
10. `tests/e2e/test_provisioning.py::test_auto_seed_ziti` - Ziti policies applied correctly

## Dependencies

None (foundation requirement for E2E testing infrastructure)

## Blocks

- REQ-E2E-002: Zitadel local instance testing
- REQ-E2E-003: E2E test suite
- REQ-ZT-001: Zitadel OIDC integration (provides local instance to develop against)
- REQ-ZT-002: OpenZiti dark service (provides local Ziti to develop against)

## Effort

3.0 weeks
