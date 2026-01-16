# REQ-DEPLOY-004: Multi-Tenant Architecture

## Requirement
RTMX managed service shall support multi-tenant isolation.

## Status: MISSING
## Priority: HIGH
## Phase: 13

## Rationale
A managed service must securely isolate tenant data and configurations. Defense and enterprise customers require strong guarantees that their data cannot be accessed by other tenants. Multi-tenant architecture also enables efficient resource utilization and simplified operations for the managed service.

## Acceptance Criteria
- [ ] Tenant data isolation at database level
- [ ] Tenant-specific encryption keys
- [ ] Tenant configuration isolation
- [ ] Cross-tenant access prevention verified
- [ ] Tenant provisioning automated
- [ ] Tenant deprovisioning with data cleanup
- [ ] Audit logging per tenant

## Isolation Architecture

### Database Isolation

| Isolation Level | Implementation | Use Case |
|-----------------|----------------|----------|
| Schema | Separate PostgreSQL schema per tenant | Standard tier |
| Database | Separate database instance per tenant | Enterprise tier |
| Cluster | Dedicated database cluster | Government tier |

```sql
-- Schema isolation example
CREATE SCHEMA tenant_abc123;
CREATE TABLE tenant_abc123.requirements (...);
CREATE TABLE tenant_abc123.audit_log (...);

-- Row-level security (additional layer)
ALTER TABLE requirements ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_isolation ON requirements
  USING (tenant_id = current_setting('app.tenant_id'));
```

### Encryption Key Management

```
                    +------------------+
                    |   Master Key     |
                    |  (HSM-backed)    |
                    +--------+---------+
                             |
              +--------------+--------------+
              |              |              |
     +--------v-----+ +------v------+ +-----v-------+
     | Tenant A Key | | Tenant B Key| | Tenant C Key|
     |    (DEK)     | |    (DEK)    | |    (DEK)    |
     +--------------+ +-------------+ +-------------+
              |              |              |
     +--------v-----+ +------v------+ +-----v-------+
     | Tenant A     | | Tenant B    | | Tenant C    |
     | Data         | | Data        | | Data        |
     +--------------+ +-------------+ +-------------+
```

### Network Isolation

- Tenant-specific Kubernetes namespaces
- Network policies preventing cross-namespace traffic
- Dedicated ingress per tenant (optional, enterprise tier)
- Service mesh mTLS between components

## Tenant Provisioning

### Automated Workflow

```yaml
# Tenant provisioning request
tenant:
  id: "abc123"
  name: "Acme Defense Corp"
  tier: "enterprise"
  region: "us-gov-east-1"
  admin_email: "admin@acme.defense"
  settings:
    encryption: "customer-managed"
    kms_key_arn: "arn:aws:kms:..."
```

### Provisioning Steps
1. Validate tenant configuration
2. Create database schema/instance
3. Generate tenant encryption key
4. Create Kubernetes namespace
5. Deploy tenant-specific resources
6. Configure DNS/routing
7. Send admin invitation email
8. Create audit log entry

## Configuration Isolation

Each tenant has isolated:
- RTM database configuration
- Schema extensions (custom fields)
- User/role definitions
- Integration settings (Jira, GitHub)
- Notification preferences
- Branding/theming (enterprise)

## Cross-Tenant Prevention

### Technical Controls
- Database connection strings include tenant schema
- API authentication includes tenant claim
- Object storage paths prefixed with tenant ID
- All queries filtered by tenant context

### Verification Tests
1. API request with Tenant A token cannot access Tenant B data
2. Database query without tenant context fails
3. Object storage access denied across tenant boundaries
4. WebSocket connections scoped to tenant room

## Test Cases
1. Tenant provisioning creates all required resources
2. Tenant A cannot query Tenant B data via API
3. Tenant A cannot access Tenant B files in storage
4. Tenant deprovisioning removes all data
5. Encryption keys are tenant-specific
6. Audit logs are tenant-isolated
7. Resource quotas enforced per tenant

## Tenant Lifecycle

```
  +----------+     +-----------+     +--------+     +-------------+
  |  Request | --> | Provision | --> | Active | --> | Decommission|
  +----------+     +-----------+     +--------+     +-------------+
                         |               |                |
                         v               v                v
                   [Create DB]    [Normal Ops]     [Backup Data]
                   [Gen Keys]     [Billing]        [Delete Data]
                   [Setup NS]    [Support]         [Revoke Keys]
```

## Technical Notes
- Tenant ID is immutable after creation
- Tenant context propagated via JWT claims
- Background jobs include tenant context
- Metrics and logs tagged with tenant ID
- Rate limiting applied per tenant

## Dependencies
- REQ-SEC-002 (E2E encryption for tenant keys)

## Blocks
- None

## Effort
3.0 weeks
