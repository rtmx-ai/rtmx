# REQ-DOC-006: RTMX 101 Whitepaper

## Requirement

RTMX 101 whitepaper shall introduce RTMX concepts and quick start.

## Description

Create an introductory whitepaper that serves as the primary onboarding document for new RTMX users. The document should explain the "why" of requirements traceability and demonstrate RTMX's value proposition.

## Acceptance Criteria

- [ ] PDF generated from Typst source
- [ ] Published to rtmx.ai/whitepapers/rtmx-101
- [ ] Covers problem statement (why RTM matters)
- [ ] Explains core RTMX concepts
- [ ] Includes quick start guide
- [ ] Documents closed-loop verification principle
- [ ] Provides visual diagrams of workflow
- [ ] 8-12 pages in length

## Outline

1. **Introduction**
   - The requirements traceability problem
   - Why traditional tools fail for modern development

2. **What is RTMX?**
   - CSV-first, Git-native, AI-friendly
   - Core philosophy: verification over claims

3. **Core Concepts**
   - Requirements database (RTM)
   - Test markers (@pytest.mark.req)
   - Status lifecycle
   - Closed-loop verification

4. **Quick Start**
   - Installation
   - rtmx init
   - rtmx status
   - rtmx verify --update

5. **The Closed Loop**
   - Visual diagram of verify workflow
   - Why status must be earned, not claimed

6. **Next Steps**
   - Link to Patterns whitepaper
   - Link to documentation

## Dependencies

- REQ-DOC-005: Typst template
