# REQ-COMPL-003: Ports, Protocols, and Services Mapping

## Requirement
RTMX shall document all required ports, protocols, and services for PPSM compliance.

## Phase
13 (Security/Compliance)

## Rationale
The Ports, Protocols, and Services Management (PPSM) process is required for all DoD systems. Network administrators and security personnel must know exactly which ports RTMX requires to configure firewalls, review network traffic, and approve the system for deployment. Incomplete PPSM documentation blocks ATO approval.

## Acceptance Criteria
- [ ] Complete PPSM matrix for cloud deployment mode
- [ ] Complete PPSM matrix for self-hosted deployment mode
- [ ] Complete PPSM matrix for air-gap deployment mode
- [ ] Justification documented for each required port/protocol
- [ ] TLS versions documented with cipher suite preferences
- [ ] Firewall rule recommendations provided
- [ ] PPSM matrix available in machine-readable format (CSV/JSON)
- [ ] Documentation updated with each network-affecting change

## PPSM Matrix - Cloud Deployment

| Port | Protocol | Direction | Service | Justification |
|------|----------|-----------|---------|---------------|
| 443 | TCP | Outbound | HTTPS/WSS | RTMX Sync API and WebSocket connections |
| 443 | TCP | Inbound | HTTPS | Web dashboard access (optional) |
| 22 | TCP | Outbound | SSH | Git remote operations (GitHub/GitLab) |

## PPSM Matrix - Self-Hosted Deployment

| Port | Protocol | Direction | Service | Justification |
|------|----------|-----------|---------|---------------|
| 443 | TCP | Inbound | HTTPS/WSS | RTMX Sync API and WebSocket connections |
| 8080 | TCP | Inbound | HTTP | Health checks and metrics (configurable) |
| 5432 | TCP | Internal | PostgreSQL | Database connections (internal only) |
| 6379 | TCP | Internal | Redis | Session/cache (optional, internal only) |

## PPSM Matrix - Air-Gap Deployment

| Port | Protocol | Direction | Service | Justification |
|------|----------|-----------|---------|---------------|
| None | - | - | - | No network ports required |
| - | - | - | Git (local) | File-based sync via local git only |

## TLS Configuration

### Required TLS Version
- Minimum: TLS 1.2 (for legacy system compatibility)
- Preferred: TLS 1.3
- FIPS Mode: TLS 1.2+ with FIPS-approved cipher suites only

### Cipher Suite Preferences (TLS 1.3)
```
TLS_AES_256_GCM_SHA384
TLS_AES_128_GCM_SHA256
TLS_CHACHA20_POLY1305_SHA256 (non-FIPS mode only)
```

### Cipher Suite Preferences (TLS 1.2 FIPS Mode)
```
TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384
TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256
```

### Prohibited Protocols
- SSLv2, SSLv3, TLS 1.0, TLS 1.1
- Export cipher suites
- NULL cipher suites
- Anonymous cipher suites
- RC4, DES, 3DES

## Firewall Recommendations

### Cloud Deployment (Outbound Rules)
```
# RTMX Sync
ALLOW TCP/443 to sync.rtmx.io

# Git Operations
ALLOW TCP/443 to github.com, gitlab.com (or enterprise hosts)
ALLOW TCP/22 to github.com, gitlab.com (optional, for SSH git)

# Package Registry (for updates)
ALLOW TCP/443 to pypi.org, files.pythonhosted.org
```

### Self-Hosted (Inbound Rules)
```
# RTMX Services
ALLOW TCP/443 from authorized networks

# Health Checks (from monitoring)
ALLOW TCP/8080 from monitoring subnet only

# Block all other inbound
DENY all
```

## Technical Notes
- Use infrastructure-as-code (Terraform) for firewall rule templates
- Document port requirements in container healthcheck commands
- Provide network policy templates for Kubernetes deployments
- Include AWS Security Group / Azure NSG / GCP Firewall examples

## Test Cases
1. Verify PPSM matrix covers all deployment modes
2. Verify all ports have documented justification
3. Verify TLS configuration documentation is complete
4. Verify machine-readable export is valid
5. Verify firewall recommendations are actionable

## Dependencies
- REQ-SEC-001 (TLS 1.3)

## Effort
1.0 weeks
