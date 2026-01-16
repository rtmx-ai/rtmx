# REQ-COMPL-007: DoD IL4/IL5 Authorization

## Requirement
RTMX shall support deployment at DoD Impact Level 4 and 5.

## Phase
13 (Security/Compliance)

## Rationale
Defense contractors and DoD components require systems authorized at Impact Level 4 (Controlled Unclassified Information - CUI) or Impact Level 5 (CUI in a National Security System context). IL4/IL5 authorization enables RTMX to support defense development teams tracking requirements for CUI-handling systems.

## Acceptance Criteria
- [ ] FedRAMP High baseline implemented (all 421 controls)
- [ ] DoD SRG (Security Requirements Guide) additional controls addressed
- [ ] US person access controls enforced with verification
- [ ] Physical isolation requirements documented for on-prem
- [ ] GovCloud deployment tested on AWS GovCloud
- [ ] GovCloud deployment tested on Azure Government
- [ ] GovCloud deployment tested on Google Cloud for Government
- [ ] STIG (Security Technical Implementation Guide) compliance documented
- [ ] CAC/PIV authentication supported
- [ ] Data isolation between tenants verified

## Impact Level Requirements

| Requirement | IL4 | IL5 |
|-------------|-----|-----|
| Data Classification | CUI | CUI in NSS context |
| FedRAMP Baseline | High | High |
| US Person Access | Required | Required |
| Data Location | US only | US only (specific regions) |
| Physical Isolation | Logical | May require physical |
| Encryption | FIPS 140-3 | FIPS 140-3 + CNSA 2.0 |
| Background Check | Moderate risk | High risk / TS |

## DoD SRG Additional Controls

Beyond FedRAMP High, the DoD Cloud Computing SRG requires:

| Control | Description | RTMX Implementation |
|---------|-------------|---------------------|
| Access Control | US person verification | USCIS E-Verify integration |
| Audit | Enhanced logging retention | 1 year minimum, 7 years archive |
| Incident Response | DoD CISO notification | 72-hour notification requirement |
| Media Protection | Sanitization procedures | Cryptographic erasure documented |
| Personnel Security | Background investigations | Customer responsibility matrix |
| System Protection | Boundary defense | Network segmentation documented |

## GovCloud Deployment Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    AWS GovCloud (US)                        │
│                    FedRAMP High Region                      │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────┐        │
│  │              RTMX VPC (IL4/IL5)                 │        │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐      │        │
│  │  │ EKS Node │  │ EKS Node │  │ EKS Node │      │        │
│  │  │ (RTMX)   │  │ (RTMX)   │  │ (RTMX)   │      │        │
│  │  └──────────┘  └──────────┘  └──────────┘      │        │
│  │         │            │            │             │        │
│  │         ▼            ▼            ▼             │        │
│  │  ┌─────────────────────────────────────┐        │        │
│  │  │     RDS PostgreSQL (Encrypted)      │        │        │
│  │  │     FIPS 140-3 Encryption           │        │        │
│  │  └─────────────────────────────────────┘        │        │
│  │                                                  │        │
│  │  ┌─────────────────────────────────────┐        │        │
│  │  │     S3 (State/Backups, SSE-KMS)     │        │        │
│  │  └─────────────────────────────────────┘        │        │
│  └─────────────────────────────────────────────────┘        │
│                         │                                    │
│                         ▼                                    │
│  ┌─────────────────────────────────────────────────┐        │
│  │     AWS PrivateLink (No Internet Egress)        │        │
│  └─────────────────────────────────────────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

## US Person Access Controls

```python
# Example access control implementation
class USPersonVerification:
    """Verify US person status for IL4/IL5 access."""

    def verify_access(self, user: User) -> bool:
        """Check US person status before granting access."""
        # Verify citizenship/residency status
        if not self.verify_us_person_status(user):
            raise AccessDenied("IL4/IL5 requires US person status")

        # Verify background check level
        if not self.verify_background_check(user, level="moderate"):
            raise AccessDenied("IL4/IL5 requires background investigation")

        # Log access attempt for audit
        self.audit_log.record(user, "IL4_ACCESS_GRANTED")
        return True
```

## CNSA 2.0 Requirements for IL5

For IL5 systems (National Security Systems context), CNSA 2.0 compliance is required:

- AES-256 for symmetric encryption
- ML-KEM-1024 for key encapsulation (quantum-resistant)
- ML-DSA-87 for digital signatures (quantum-resistant)
- SHA-384 minimum for hashing
- See REQ-SEC-014 for implementation details

## Technical Notes
- Use AWS GovCloud, Azure Government, or GCP for Government
- Implement tenant isolation at network and data layers
- Document STIG compliance for all infrastructure components
- Support CAC/PIV authentication for defense users
- Configure audit log shipping to customer SIEM

## Test Cases
1. Verify GovCloud deployment completes successfully
2. Verify FIPS mode is enforced in GovCloud environment
3. Verify US person access controls reject non-US persons
4. Verify data remains in US regions only
5. Verify tenant isolation prevents cross-tenant access
6. Verify CAC/PIV authentication flow works

## Dependencies
- REQ-COMPL-006 (FedRAMP authorization)
- REQ-SEC-014 (CNSA 2.0 compliance)

## Effort
8.0 weeks (assumes FedRAMP High already complete)
