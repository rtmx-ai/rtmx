# REQ-COMPL-002: Information Assurance Boundary Diagram

## Requirement
RTMX shall maintain authoritative IA boundary diagrams for all deployment modes.

## Phase
13 (Security/Compliance)

## Rationale
Every ATO package requires clear boundary diagrams showing what is in scope for authorization, where data flows, and how systems interconnect. Without authoritative boundary diagrams, security assessors cannot evaluate the system, and customers cannot integrate RTMX into their own authorization boundaries.

## Acceptance Criteria
- [ ] Boundary diagram for cloud/managed service deployment
- [ ] Boundary diagram for self-hosted/on-prem deployment
- [ ] Boundary diagram for air-gap (Zarf) deployment
- [ ] Data flow diagrams showing encryption points
- [ ] Network segmentation documented
- [ ] All external interfaces identified
- [ ] Trust boundaries clearly marked
- [ ] Diagrams available in editable format (draw.io, Mermaid, or similar)
- [ ] Diagrams versioned and maintained with releases

## Deployment Mode Boundaries

### Cloud/Managed Service
```
┌─────────────────────────────────────────────────────────────┐
│                    RTMX Authorization Boundary              │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐     │
│  │ RTMX Web UI │◄──►│ RTMX Sync   │◄──►│ RTMX API    │     │
│  │  (FastAPI)  │    │  (CRDT)     │    │  Gateway    │     │
│  └─────────────┘    └─────────────┘    └─────────────┘     │
│         │                 │                   │             │
│         ▼                 ▼                   ▼             │
│  ┌───────────────────────────────────────────────────┐     │
│  │              Encrypted Data Store                  │     │
│  │              (PostgreSQL + CRDT State)             │     │
│  └───────────────────────────────────────────────────┘     │
├─────────────────────────────────────────────────────────────┤
│                    CSP Infrastructure                       │
│              (AWS/Azure/GCP - Inherited Controls)           │
└─────────────────────────────────────────────────────────────┘
         ▲                                        ▲
         │ TLS 1.3                                │ TLS 1.3
         ▼                                        ▼
    ┌──────────┐                           ┌──────────┐
    │ RTMX CLI │                           │ IDE Ext  │
    │ (Customer│                           │ (Customer│
    │  Workst.)│                           │  Workst.)│
    └──────────┘                           └──────────┘
```

### Self-Hosted/On-Prem
```
┌─────────────────────────────────────────────────────────────┐
│              Customer Authorization Boundary                 │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────┐           │
│  │         RTMX Container (Customer Managed)     │           │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐       │           │
│  │  │ Web UI  │  │  Sync   │  │  API    │       │           │
│  │  └─────────┘  └─────────┘  └─────────┘       │           │
│  │                    │                          │           │
│  │                    ▼                          │           │
│  │            ┌─────────────┐                    │           │
│  │            │ Local Store │                    │           │
│  │            └─────────────┘                    │           │
│  └──────────────────────────────────────────────┘           │
│                                                              │
│              Customer Network Infrastructure                 │
└─────────────────────────────────────────────────────────────┘
```

### Air-Gap (Zarf)
```
┌─────────────────────────────────────────────────────────────┐
│         DISN / Classified Network Boundary                  │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────┐           │
│  │         RTMX Zarf Package (Offline)           │           │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐       │           │
│  │  │ Web UI  │  │  Local  │  │  CLI    │       │           │
│  │  │         │  │  Sync   │  │         │       │           │
│  │  └─────────┘  └─────────┘  └─────────┘       │           │
│  │                    │                          │           │
│  │                    ▼                          │           │
│  │         ┌─────────────────┐                   │           │
│  │         │ Local Git + CSV │                   │           │
│  │         └─────────────────┘                   │           │
│  └──────────────────────────────────────────────┘           │
│                                                              │
│     No external network connections - Git-only sync         │
└─────────────────────────────────────────────────────────────┘
```

## Data Flow Documentation

| Flow | Source | Destination | Protocol | Encryption |
|------|--------|-------------|----------|------------|
| CLI to Sync | Developer workstation | RTMX Sync Server | WebSocket | TLS 1.3 + E2E |
| Web UI to API | Browser | RTMX API Gateway | HTTPS | TLS 1.3 |
| Sync to Store | RTMX Server | Database | Internal | AES-256 |
| IDE to Sync | IDE Extension | RTMX Sync Server | WebSocket | TLS 1.3 + E2E |

## Technical Notes
- Use Mermaid for diagrams in documentation (renders in GitHub)
- Maintain draw.io source files for detailed architecture
- Update diagrams with each major release
- Include FIPS mode encryption details where applicable

## Test Cases
1. Verify boundary diagrams exist for all three deployment modes
2. Verify all external interfaces are documented
3. Verify encryption points are marked on data flow diagrams
4. Verify diagrams are parseable (valid Mermaid/draw.io)
5. Verify version matches current release

## Dependencies
None (foundational documentation)

## Effort
2.0 weeks
