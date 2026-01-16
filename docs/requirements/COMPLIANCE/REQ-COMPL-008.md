# REQ-COMPL-008: CMMC Level 2 Alignment

## Requirement
RTMX documentation and processes shall align with CMMC 2.0 Level 2.

## Phase
13 (Security/Compliance)

## Rationale
The Cybersecurity Maturity Model Certification (CMMC) 2.0 is required for defense contractors handling CUI. CMMC Level 2 aligns with NIST SP 800-171 R2 and requires third-party assessment. RTMX customers in the DIB (Defense Industrial Base) need tools that support their own CMMC compliance. RTMX must demonstrate its own alignment and provide features that help customers achieve compliance.

## Acceptance Criteria
- [ ] NIST SP 800-171 R2 control mapping complete (110 controls)
- [ ] CUI handling procedures documented
- [ ] Assessment scope boundary defined for RTMX as a tool
- [ ] Self-assessment capability for customers using RTMX
- [ ] Evidence collection automated via RTM database
- [ ] SSP template available for customers using RTMX
- [ ] Customer responsibility matrix for CUI handling
- [ ] CMMC assessment documentation package

## CMMC 2.0 Level 2 Overview

| Domain | Practice Count | RTMX Relevance |
|--------|---------------|----------------|
| Access Control (AC) | 22 | Authentication, RBAC |
| Awareness & Training (AT) | 3 | Customer responsibility |
| Audit & Accountability (AU) | 9 | Audit logging |
| Configuration Management (CM) | 9 | Version control, schema |
| Identification & Authentication (IA) | 11 | SSO, MFA |
| Incident Response (IR) | 3 | Event detection |
| Maintenance (MA) | 6 | Update procedures |
| Media Protection (MP) | 9 | Data encryption |
| Personnel Security (PS) | 2 | Customer responsibility |
| Physical Protection (PE) | 6 | CSP/Customer responsibility |
| Risk Assessment (RA) | 3 | Vulnerability management |
| Security Assessment (CA) | 4 | Self-assessment |
| System & Communications Protection (SC) | 16 | TLS, encryption |
| System & Information Integrity (SI) | 7 | Input validation |

## NIST SP 800-171 R2 Mapping

Key controls with RTMX implementation status:

| Control | Description | RTMX Implementation |
|---------|-------------|---------------------|
| 3.1.1 | Limit access to authorized users | RBAC via REQ-SEC-004 |
| 3.1.2 | Limit access to transactions | Role-based permissions |
| 3.1.5 | Least privilege | Viewer/Editor/Owner roles |
| 3.3.1 | Create audit records | REQ-SEC-005 |
| 3.3.2 | Unique user traceability | User ID in audit logs |
| 3.5.1 | Identify system users | REQ-SEC-003 SSO |
| 3.5.2 | Authenticate users | MFA via IdP |
| 3.13.1 | Monitor communications | TLS inspection capability |
| 3.13.8 | Cryptographic mechanisms | REQ-SEC-008 FIPS |
| 3.14.1 | Identify flaws | REQ-COMPL-004 POA&M |

## CUI Handling in RTMX

RTMX may contain CUI when customers track requirements for CUI-handling systems:

```
CUI Categories Potentially in RTMX:
├── ITAR-controlled technical data (defense articles)
├── Export-controlled source code references
├── System security architecture details
├── Vulnerability information (pre-disclosure)
└── Proprietary contractor information
```

### CUI Protection Measures

| Measure | Implementation |
|---------|----------------|
| Marking | Requirement metadata field for CUI marking |
| Encryption | E2E encryption (REQ-SEC-002) |
| Access Control | RBAC with CUI awareness |
| Audit | All CUI access logged |
| Sharing | CUI sharing restrictions enforced |
| Disposal | Cryptographic erasure |

## Self-Assessment Features

RTMX provides features to help customers with their own CMMC self-assessment:

```bash
# Generate CMMC evidence from RTM database
rtmx compliance evidence --framework cmmc-l2 --output evidence/

# Generate SSP sections from RTMX configuration
rtmx compliance ssp --framework 800-171 --output ssp/

# Track CMMC practices as requirements
rtmx import --template cmmc-l2-practices.csv

# Validate CMMC-related requirements are complete
rtmx validate --category CMMC
```

## Evidence Collection Automation

RTMX can automatically collect evidence for CMMC practices:

| Practice | Evidence Source |
|----------|-----------------|
| AU-2 Audit events | RTMX audit log export |
| CM-2 Baseline config | rtmx.yaml version history |
| CM-3 Change control | Git commit history |
| IR-4 Incident handling | POA&M database |
| RA-5 Vulnerability scan | SBOM + VEX documents |

## Technical Notes
- Create CMMC practice tracker as requirement category
- Map CMMC practices to existing RTMX features
- Provide Jinja2 templates for SSP sections
- Automate SPRS score calculation
- Support both self-assessment and C3PAO documentation

## Assessment Scope

When RTMX is in a customer's assessment scope:

```
Customer CUI Boundary
├── Customer Development Environment
│   ├── Source Code (CUI)
│   ├── Requirements Database (RTMX) ◄── In scope
│   └── Documentation
├── Customer Infrastructure
│   └── [Customer's other systems]
└── RTMX Sync Service (if used)
    └── Inherits from RTMX's authorization
```

## Test Cases
1. Verify 800-171 mapping covers all 110 controls
2. Verify evidence export produces valid documentation
3. Verify CUI marking field is available in schema
4. Verify CUI access is logged appropriately
5. Verify SSP template generates valid output
6. Verify SPRS score calculation is accurate

## Dependencies
- REQ-COMPL-001 (NIST 800-53 mapping)

## Effort
4.0 weeks
