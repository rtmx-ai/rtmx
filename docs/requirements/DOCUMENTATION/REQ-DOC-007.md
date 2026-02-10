# REQ-DOC-007: RTMX Patterns and Anti-Patterns Whitepaper

## Requirement

RTMX Patterns and Anti-Patterns whitepaper shall document best practices.

## Description

Create a comprehensive whitepaper documenting recommended patterns and common anti-patterns when using RTMX. This serves as the authoritative guide for teams adopting RTMX and for AI agents working in RTMX-enabled projects.

## Acceptance Criteria

- [ ] PDF generated from Typst source
- [ ] Published to rtmx.ai/whitepapers/patterns
- [ ] Content derived from docs/patterns.md (single source of truth)
- [ ] Expanded examples beyond the markdown source
- [ ] Visual diagrams for key patterns
- [ ] Case studies or scenarios
- [ ] 15-25 pages in length

## Outline

1. **Introduction**
   - Purpose of this guide
   - The cost of anti-patterns

2. **Core Principle: Closed-Loop Verification**
   - Visual diagram
   - Why evidence > opinion

3. **Verification Patterns**
   - Pattern: Automated Status Updates
   - Anti-Pattern: Manual Status Edits
   - Pattern: Test-Linked Requirements
   - Anti-Pattern: Orphan Tests

4. **Development Workflow Patterns**
   - Pattern: Spec-First Development
   - Anti-Pattern: Code-First, Spec-Never
   - Pattern: Phase Gates in CI
   - Anti-Pattern: Phase as Suggestion

5. **Agent Integration Patterns**
   - Pattern: Agent as Implementer, RTMX as Verifier
   - Anti-Pattern: Agent Status Claims
   - Pattern: RTM as Development Contract
   - Anti-Pattern: Ignoring Dependencies

6. **CI/CD Patterns**
   - Pattern: Verify on Every PR
   - Pattern: RTM Diff in PRs
   - Example: GitHub Actions workflow

7. **Case Study: Real-World Anti-Pattern Recovery**
   - Scenario: Agent manually updating status
   - Diagnosis: Tests weren't running
   - Solution: CI enforcement

8. **Quick Reference**
   - Pattern summary table
   - Command cheat sheet

## Source Document

The canonical source is `docs/patterns.md` in the rtmx repository. This whitepaper expands on that content with additional context, diagrams, and examples.

## Dependencies

- REQ-DOC-005: Typst template
