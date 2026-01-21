# REQ-DEPLOY-002: Self-Hosted Enterprise

## Requirement
RTMX shall support self-hosted deployment for enterprise customers.

## Status: MISSING
## Priority: HIGH
## Phase: 13

## Rationale
Many enterprise customers require on-premises deployment due to data sovereignty, security policies, or network isolation requirements. Self-hosted deployment enables customers to maintain full control over their infrastructure and data while benefiting from RTMX capabilities.

## Acceptance Criteria
- [ ] Helm chart for Kubernetes deployment
- [ ] Docker Compose for single-node deployment
- [ ] Installation documentation complete
- [ ] License validation works offline
- [ ] Customer-provided database support (PostgreSQL)
- [ ] Configuration via environment variables and config files
- [ ] Upgrade path documented with rollback procedures

## Deployment Options

### Kubernetes (Production)

```yaml
# values.yaml example
rtmx:
  replicaCount: 3
  image:
    repository: ghcr.io/rtmx-ai/rtmx
    tag: "v1.0.0"
  database:
    external: true
    host: "postgres.internal"
    name: "rtmx"
  license:
    key: "${RTMX_LICENSE_KEY}"
    offlineMode: true
  persistence:
    enabled: true
    size: 10Gi
```

### Docker Compose (Development/Small Teams)

```yaml
# docker-compose.yml structure
services:
  rtmx-server:
    image: ghcr.io/rtmx-ai/rtmx:latest
    environment:
      - DATABASE_URL
      - RTMX_LICENSE_KEY
    volumes:
      - rtmx-data:/data
  postgres:
    image: postgres:16-alpine
    volumes:
      - postgres-data:/var/lib/postgresql/data
```

## Database Support

| Database | Version | Notes |
|----------|---------|-------|
| PostgreSQL | 14+ | Recommended, fully tested |
| PostgreSQL | 12-13 | Supported, limited testing |
| SQLite | 3.35+ | Single-node only, dev/small teams |

## License Validation

### Online Mode
- License key validated against ioTACTICAL license server
- Periodic re-validation (configurable interval)
- Graceful degradation on temporary network issues

### Offline Mode
- License key contains cryptographic signature
- No network calls required after initial activation
- Expiration date encoded in license
- Machine fingerprint validation (optional)

## Installation Documentation

Required documentation:
1. Prerequisites and system requirements
2. Quick start guide (5-minute deployment)
3. Production deployment guide
4. Database configuration
5. TLS/SSL certificate setup
6. Backup and restore procedures
7. Monitoring and alerting setup
8. Troubleshooting guide
9. Upgrade procedures

## Test Cases
1. Helm chart deploys to clean Kubernetes cluster
2. Docker Compose starts all services correctly
3. Offline license validation accepts valid license
4. Offline license validation rejects expired license
5. PostgreSQL connection with customer-provided credentials
6. Upgrade from previous version preserves data
7. Rollback restores previous version

## System Requirements

### Minimum (Single Node)
- 2 CPU cores
- 4 GB RAM
- 20 GB storage
- Docker 24.0+ or Kubernetes 1.27+

### Recommended (Production)
- 4+ CPU cores per node
- 8+ GB RAM per node
- 100 GB SSD storage
- Kubernetes 1.28+ with 3+ nodes

## Technical Notes
- Same container images as managed service
- Configuration differences only (no code forks)
- Support for air-gapped networks via REQ-DEPLOY-003
- Telemetry opt-in only, never required

## Dependencies
- None (standalone deployment)

## Blocks
- None

## Effort
4.0 weeks
