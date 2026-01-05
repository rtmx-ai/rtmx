# REQ-SEC-006: Data Residency Selection

## Requirement
Users shall select data residency region.

## Phase
13 (Compliance)

## Rationale
Many organizations have legal or policy requirements about where data is stored. GDPR requires EU data to stay in EU, defense contractors may require US-only storage, and some organizations need on-premises deployment.

## Acceptance Criteria
- [ ] US region available (us-east, us-west)
- [ ] EU region available (eu-west, eu-central)
- [ ] Region selected at project creation
- [ ] Data never leaves selected region
- [ ] On-premises deployment option available
- [ ] Region displayed in project settings
- [ ] Migration between regions supported (with data export/import)

## Region Options

| Region | Location | Compliance |
|--------|----------|------------|
| us-east | Virginia, USA | FedRAMP, ITAR |
| us-west | Oregon, USA | FedRAMP |
| eu-west | Ireland | GDPR |
| eu-central | Frankfurt | GDPR |
| on-prem | Customer DC | All |

## Technical Notes
- Deploy sync servers in each region
- Use region-specific database instances
- DNS/routing directs traffic to correct region
- Cross-region replication explicitly disabled
- On-prem uses same codebase with customer-managed infrastructure

## Test Cases
1. Project created in US stores data in US
2. Project created in EU stores data in EU
3. Cross-region API calls are rejected
4. On-prem deployment functions identically

## Dependencies
- REQ-SEC-001 (TLS infrastructure)

## Effort
2.0 weeks
