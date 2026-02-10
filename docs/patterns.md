# RTMX Patterns and Anti-Patterns

> **Version**: 1.0.0
> **Last Updated**: 2026-02-10
> **Canonical URL**: https://rtmx.ai/patterns

This document defines recommended patterns and anti-patterns for working with RTMX. Following these patterns ensures requirements remain traceable, verifiable, and trustworthy.

---

## Core Principle: Closed-Loop Verification

RTMX is built on a fundamental principle: **requirement status must be derived from evidence, not opinion**.

```
┌─────────────────────────────────────────────────────────────────┐
│                    CLOSED-LOOP VERIFICATION                      │
│                                                                  │
│   Requirements ──→ Tests ──→ Execution ──→ Status Update        │
│        ↑                                        │                │
│        └────────────────────────────────────────┘                │
│                                                                  │
│   Status reflects reality. Tests are the arbiter of truth.      │
└─────────────────────────────────────────────────────────────────┘
```

---

## Verification Patterns

### Pattern: Automated Status Updates

**Use `rtmx verify --update` to derive status from test results.**

```bash
# In CI/CD pipeline
rtmx verify --update

# Local development (dry-run first)
rtmx verify --dry-run
rtmx verify --update
```

The `verify` command:
1. Runs all tests with `@pytest.mark.req()` markers
2. Maps test outcomes to requirements
3. Updates status based on pass/fail results

| Test Result | Status Transition |
|-------------|-------------------|
| All pass | → COMPLETE |
| Some pass, none fail | → PARTIAL |
| Any fail | COMPLETE → PARTIAL (regression) |
| No tests | Status unchanged |

**Why this matters**: Status reflects what the code actually does, not what someone hopes it does.

---

### Anti-Pattern: Manual Status Edits

**Never manually edit the `status` field in the RTM database.**

```csv
# WRONG: Agent or human manually sets status
REQ-AUTH-001,AUTH,Login,User can log in,COMPLETE,...

# RIGHT: Let rtmx verify determine status from tests
rtmx verify --update
```

**Symptoms of this anti-pattern**:
- Status says COMPLETE but tests fail
- No test coverage for "complete" requirements
- Status changes without corresponding code changes
- Requirements marked complete before implementation

**Why it's harmful**: Manual status updates break the feedback loop. The RTM becomes a wish list instead of a verification record.

---

### Pattern: Test-Linked Requirements

**Every requirement should have at least one test with `@pytest.mark.req()`.**

```python
@pytest.mark.req("REQ-AUTH-001")
@pytest.mark.scope_unit
@pytest.mark.technique_nominal
@pytest.mark.env_simulation
def test_user_can_login():
    """Verify REQ-AUTH-001: User can log in."""
    user = create_test_user()
    result = login(user.email, user.password)
    assert result.success
    assert result.session_token is not None
```

**Sync test metadata to RTM**:
```bash
rtmx from-tests --update
```

This populates `test_module` and `test_function` columns, creating bidirectional traceability.

---

### Anti-Pattern: Orphan Tests

**Tests without requirement markers are untraceable.**

```python
# WRONG: No requirement linkage
def test_login_works():
    assert login("user", "pass").success

# RIGHT: Linked to requirement
@pytest.mark.req("REQ-AUTH-001")
def test_login_works():
    assert login("user", "pass").success
```

**Why it's harmful**: Orphan tests provide no evidence for requirement completion. They may test important functionality, but that functionality isn't tracked in the RTM.

**Detection**:
```bash
rtmx from-tests --missing  # Shows tests not linked to RTM requirements
```

---

## Development Workflow Patterns

### Pattern: Spec-First Development

**Write the requirement specification before writing code.**

```
1. Define requirement in RTM database
2. Create specification file (docs/requirements/CATEGORY/REQ-XXX.md)
3. Write acceptance criteria
4. Write failing tests
5. Implement to pass tests
6. Run rtmx verify --update
```

**Specification template**:
```markdown
# REQ-XXX-NNN: Requirement Title

## Description
Clear statement of what the system shall do.

## Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

## Test Cases
1. Nominal case
2. Edge case
3. Error case
```

---

### Anti-Pattern: Code-First, Spec-Never

**Writing code without defining requirements creates untraceable features.**

```
# WRONG workflow
1. Write code
2. Maybe write tests
3. Never create requirement
4. Feature exists but isn't tracked

# RIGHT workflow
1. Define requirement
2. Write spec
3. Write failing test
4. Write code
5. Verify
```

**Why it's harmful**: Features without requirements can't be verified, prioritized, or traced. When asked "what does the system do?", the answer becomes "read the code."

---

### Pattern: Phase Gates in CI

**Block releases until phase requirements are verified.**

```yaml
# .github/workflows/release.yml
- name: Verify Phase Requirements
  run: |
    rtmx verify --update
    rtmx status --json | jq -e '.phases["1"].complete == true'
```

**Phase progression**:
```
Phase 1: Core functionality (must be COMPLETE to release v0.1)
Phase 2: Extended features (must be COMPLETE to release v0.2)
Phase 3: Polish and optimization (must be COMPLETE to release v1.0)
```

---

### Anti-Pattern: Phase as Suggestion

**Treating phases as optional guidance rather than gates.**

```bash
# WRONG: Ignore phase gates
git tag v1.0.0  # Released with Phase 1 incomplete

# RIGHT: Enforce phase completion
rtmx status --phase 1  # Verify 100% before release
```

**Why it's harmful**: Releasing with incomplete phases means shipping unverified functionality. Phases exist to ensure quality gates.

---

## Agent Integration Patterns

### Pattern: Agent as Implementer, RTMX as Verifier

**Agents write code. `rtmx verify` determines completion.**

```
Agent Workflow:
1. Read requirement spec: docs/requirements/CATEGORY/REQ-XXX.md
2. Read existing tests
3. Implement code to pass tests
4. Run: rtmx verify --dry-run
5. If passing: rtmx verify --update
6. Commit changes
```

**Agent prompt guidance**:
```markdown
When implementing requirements:
- Never manually edit the status field in rtm_database.csv
- Always run `rtmx verify --update` after implementation
- Status is determined by test results, not by completion claims
```

---

### Anti-Pattern: Agent Status Claims

**Agents claiming completion without verification.**

```python
# WRONG: Agent edits CSV directly
db = RTMDatabase.load("docs/rtm_database.csv")
db.update("REQ-AUTH-001", status=Status.COMPLETE)  # Opinion-based
db.save()

# RIGHT: Agent runs verification
subprocess.run(["rtmx", "verify", "--update"])  # Evidence-based
```

**Why it's harmful**: Agent opinions about completion are unreliable. Tests may not exist, may fail, or may not cover the requirement. Only `rtmx verify` provides evidence-based status.

---

### Pattern: RTM as Development Contract

**Agents should read the RTM to understand what to build.**

```bash
# Agent discovers next task
rtmx backlog --phase 2 --limit 1

# Agent reads specification
cat docs/requirements/CATEGORY/REQ-XXX.md

# Agent checks dependencies
rtmx deps --req REQ-XXX

# Agent implements, then verifies
rtmx verify --update
```

**The RTM provides**:
- What to build (requirement text)
- How to verify (linked tests)
- What's blocking (dependencies)
- Priority order (phase, priority)

---

### Anti-Pattern: Ignoring Dependencies

**Implementing requirements before their dependencies are complete.**

```bash
# WRONG: Ignore blockedBy
rtmx backlog  # Shows REQ-B blocked by REQ-A
# Agent implements REQ-B anyway

# RIGHT: Respect dependencies
rtmx deps --req REQ-B  # Check what blocks this
# Implement REQ-A first, then REQ-B
```

**Why it's harmful**: Implementing blocked requirements often requires rework when the blocking requirement is completed. Dependencies exist for a reason.

---

## Continuous Integration Patterns

### Pattern: Verify on Every PR

**Run `rtmx verify` in CI to catch regressions.**

```yaml
# .github/workflows/ci.yml
jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install RTMX
        run: pip install rtmx
      - name: Verify Requirements
        run: rtmx verify --update
      - name: Check for Regressions
        run: |
          git diff --exit-code docs/rtm_database.csv || \
            echo "::warning::RTM status changed - review before merging"
```

---

### Pattern: RTM Diff in PRs

**Show requirement status changes in pull requests.**

```yaml
- name: RTM Diff
  run: |
    git fetch origin main
    rtmx diff origin/main HEAD --format markdown >> $GITHUB_STEP_SUMMARY
```

This surfaces:
- New requirements added
- Requirements completed
- Regressions (COMPLETE → PARTIAL)

---

## Quick Reference

### Commands for Closed-Loop Verification

| Command | Purpose | When to Use |
|---------|---------|-------------|
| `rtmx verify --dry-run` | Preview status changes | Before updating |
| `rtmx verify --update` | Update status from tests | After implementation |
| `rtmx from-tests --update` | Sync test metadata | After adding tests |
| `rtmx health` | Validate RTM integrity | Before releases |

### Status Sources

| Status Source | Trustworthy? | Notes |
|---------------|--------------|-------|
| `rtmx verify --update` | Yes | Evidence-based |
| Manual CSV edit | No | Opinion-based |
| Agent claim | No | Unverified |
| CI pipeline | Yes | Automated verification |

### Verification Checklist

Before claiming a requirement is complete:

- [ ] Test exists with `@pytest.mark.req("REQ-XXX")`
- [ ] Test passes locally
- [ ] `rtmx verify --update` shows COMPLETE
- [ ] CI pipeline passes
- [ ] Specification acceptance criteria met

---

## Summary

| Do This (Pattern) | Not This (Anti-Pattern) |
|-------------------|-------------------------|
| `rtmx verify --update` | Manual status edits |
| `@pytest.mark.req()` on all tests | Orphan tests |
| Spec-first development | Code-first, spec-never |
| Phase gates in CI | Phase as suggestion |
| Agent implements, RTMX verifies | Agent claims completion |
| Respect dependencies | Ignore blockedBy |

**Remember**: The RTM is a verification record, not a wish list. Status must be earned through passing tests, not claimed through manual edits.

---

*For more information, visit [rtmx.ai](https://rtmx.ai) or run `rtmx --help`.*
