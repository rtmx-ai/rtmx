# REQ-COMPL-004: POA&M Management

## Requirement
RTMX shall support Plan of Action and Milestones tracking for security findings.

## Phase
13 (Security/Compliance)

## Rationale
Every ATO package includes a POA&M documenting known security findings, remediation plans, and target completion dates. Organizations must track and report POA&M status to their Authorizing Officials (AO). RTMX should use its own traceability capabilities to track security findings, demonstrating dogfooding while providing essential ATO documentation.

## Acceptance Criteria
- [ ] POA&M template following DoD/FedRAMP format available
- [ ] Integration with vulnerability scanning results (SARIF, CycloneDX VEX)
- [ ] Milestone tracking with due dates and status
- [ ] Risk acceptance workflow with approval chain
- [ ] Export to standard formats (Excel/XLSX, JSON, OSCAL)
- [ ] POA&M items tracked in RTMX using REQ-POAM-* prefix
- [ ] Automatic severity classification based on CVSS scores
- [ ] Overdue milestone alerting
- [ ] Historical tracking of POA&M completion rates

## POA&M Template Fields

| Field | Description | Required |
|-------|-------------|----------|
| ID | Unique identifier (POAM-YYYY-NNN) | Yes |
| Finding | Description of vulnerability/weakness | Yes |
| Source | Scanner, assessment, or audit source | Yes |
| Severity | Critical/High/Medium/Low | Yes |
| CVSS | CVSS 3.1 score if applicable | No |
| Control | Related NIST 800-53 control | Yes |
| Status | Open/In Progress/Completed/Risk Accepted | Yes |
| Milestone | Remediation milestone description | Yes |
| Due Date | Target completion date | Yes |
| Resources | Required resources/budget | No |
| Risk Level | Residual risk if not remediated | Yes |
| Justification | Risk acceptance justification (if applicable) | Conditional |
| Approver | Risk acceptance approver (if applicable) | Conditional |

## Risk Acceptance Workflow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Finding   │────►│  Analysis   │────►│  Remediate  │
│  Identified │     │  & Triage   │     │  or Accept  │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                    ┌──────────────────────────┼──────────────────────────┐
                    │                          │                          │
                    ▼                          ▼                          ▼
            ┌─────────────┐           ┌─────────────┐           ┌─────────────┐
            │  Remediate  │           │   Mitigate  │           │    Risk     │
            │   (Fix it)  │           │  (Reduce    │           │  Acceptance │
            │             │           │   Impact)   │           │ (AO Approve)│
            └─────────────┘           └─────────────┘           └─────────────┘
                    │                         │                          │
                    ▼                         ▼                          ▼
            ┌─────────────┐           ┌─────────────┐           ┌─────────────┐
            │  Validate   │           │  Document   │           │  Document   │
            │   & Close   │           │  Mitigation │           │ Justification│
            └─────────────┘           └─────────────┘           └─────────────┘
```

## Integration Points

### Vulnerability Scanners
- Import SARIF format results (GitHub Advanced Security, Snyk, etc.)
- Import CycloneDX VEX format
- Import Trivy JSON output
- Import OWASP Dependency-Check XML

### Export Formats
- FedRAMP POA&M Template (XLSX)
- DoD eMASS-compatible format
- OSCAL POA&M JSON
- CSV for analysis

## CLI Commands

```bash
# Import vulnerability scan results
rtmx poam import --format sarif results.sarif

# List all POA&M items
rtmx poam list --status open

# Update POA&M status
rtmx poam update POAM-2025-001 --status "In Progress" --milestone "Patch pending"

# Export for ATO package
rtmx poam export --format xlsx poam.xlsx

# Risk acceptance workflow
rtmx poam accept POAM-2025-001 --justification "Compensating control in place" --approver "ISSM"
```

## Technical Notes
- Store POA&M as requirements with special category in RTM database
- Use validation schema for required fields
- Integrate with existing `rtmx status` and `rtmx backlog` commands
- Support filtering by severity, status, and due date
- Calculate POA&M closure rate metrics

## Test Cases
1. Verify POA&M template has all required FedRAMP fields
2. Verify SARIF import creates POA&M items correctly
3. Verify Excel export matches FedRAMP template format
4. Verify risk acceptance requires approver field
5. Verify overdue items are flagged in status output
6. Verify OSCAL export produces valid schema

## Dependencies
- REQ-SEC-005 (Audit logging)

## Effort
3.0 weeks
