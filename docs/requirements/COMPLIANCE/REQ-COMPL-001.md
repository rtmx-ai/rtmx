# REQ-COMPL-001: NIST 800-53 Rev 5 Controls Mapping

## Requirement
RTMX shall provide documented mapping to NIST SP 800-53 Rev 5 security controls.

## Phase
13 (Security/Compliance)

## Rationale
NIST SP 800-53 Rev 5 is the foundation for federal cybersecurity compliance. All FedRAMP authorizations, DoD ATOs, and many civilian agency systems require documented control implementations mapped to 800-53. Without this mapping, RTMX cannot be deployed in federal environments or support customers seeking their own ATOs.

## Acceptance Criteria
- [ ] Control mapping document created for all applicable control families
- [ ] Mapping covers AC (Access Control) family controls
- [ ] Mapping covers AU (Audit and Accountability) family controls
- [ ] Mapping covers CA (Assessment, Authorization, and Monitoring) family controls
- [ ] Mapping covers CM (Configuration Management) family controls
- [ ] Mapping covers CP (Contingency Planning) family controls
- [ ] Mapping covers IA (Identification and Authentication) family controls
- [ ] Mapping covers IR (Incident Response) family controls
- [ ] Mapping covers MA (Maintenance) family controls
- [ ] Mapping covers MP (Media Protection) family controls
- [ ] Mapping covers PE (Physical and Environmental Protection) family controls
- [ ] Mapping covers PL (Planning) family controls
- [ ] Mapping covers PM (Program Management) family controls
- [ ] Mapping covers PS (Personnel Security) family controls
- [ ] Mapping covers RA (Risk Assessment) family controls
- [ ] Mapping covers SA (System and Services Acquisition) family controls
- [ ] Mapping covers SC (System and Communications Protection) family controls
- [ ] Mapping covers SI (System and Information Integrity) family controls
- [ ] Mapping covers SR (Supply Chain Risk Management) family controls
- [ ] Inheritance model for CSP-provided controls documented
- [ ] Customer responsibility matrix (CRM) available
- [ ] Machine-readable mapping available in JSON/OSCAL format

## Control Family Coverage

| Family | Description | RTMX Relevance |
|--------|-------------|----------------|
| AC | Access Control | RBAC, authentication, session management |
| AU | Audit and Accountability | Audit logging, immutable logs, SIEM integration |
| CA | Assessment, Authorization, Monitoring | Continuous monitoring, POA&M, self-assessment |
| CM | Configuration Management | Schema validation, version control integration |
| CP | Contingency Planning | Backup procedures, disaster recovery |
| IA | Identification and Authentication | SSO, MFA, PKI support |
| IR | Incident Response | Security event detection, alerting |
| MA | Maintenance | Update mechanisms, patch management |
| MP | Media Protection | Data encryption at rest |
| PE | Physical/Environmental | CSP-inherited for cloud deployments |
| PL | Planning | Security documentation, architecture diagrams |
| PM | Program Management | Risk management framework |
| PS | Personnel Security | Customer-inherited |
| RA | Risk Assessment | Vulnerability scanning, threat modeling |
| SA | System and Services Acquisition | SBOM, secure SDLC |
| SC | System and Communications Protection | TLS, E2E encryption, boundary protection |
| SI | System and Information Integrity | Input validation, malware protection |
| SR | Supply Chain Risk Management | Dependency verification, SBOM |

## Technical Notes
- Use OSCAL (Open Security Controls Assessment Language) for machine-readable format
- Align with FedRAMP Moderate baseline (325 controls)
- Document which controls are inherited from CSP vs. customer responsibility
- Leverage existing REQ-SEC-* implementations where applicable

## Inheritance Model
```
CSP Provided (Inherited):
- PE-* (Physical security)
- Most PS-* (Personnel security for infrastructure)
- Some SC-* (Network boundary protection)

Shared Responsibility:
- AC-* (Platform RBAC + customer user management)
- AU-* (Platform logging + customer log review)
- CM-* (Platform baseline + customer configuration)

Customer Responsibility:
- PS-* (Customer personnel)
- AT-* (Customer security awareness)
- Customer-specific policies
```

## Test Cases
1. Verify control mapping document exists with all 20 families
2. Verify JSON/OSCAL export produces valid schema
3. Verify CRM clearly delineates responsibilities
4. Verify inheritance model covers cloud/on-prem/air-gap scenarios
5. Verify control implementations reference RTMX features

## Dependencies
- REQ-SEC-008 (FIPS mode)

## Effort
4.0 weeks
