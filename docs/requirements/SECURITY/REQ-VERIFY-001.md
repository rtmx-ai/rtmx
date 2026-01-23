# REQ-VERIFY-001: Property-Based Testing with Hypothesis

## Status: NOT STARTED
## Priority: HIGH
## Phase: 13
## Effort: 2.5 weeks

## Description

RTMX shall implement property-based testing using Hypothesis to exhaustively verify access control invariants. Stateful testing shall model the trust federation graph, verifying that no sequence of operations can violate security properties like privilege escalation or unauthorized access.

## Acceptance Criteria

- [ ] Hypothesis strategies for users, repos, grants, and requirements
- [ ] Stateful machine models trust federation operations
- [ ] Invariant: No privilege escalation beyond granted roles
- [ ] Invariant: Revoked grants immediately remove access
- [ ] Invariant: Shadow visibility respects grant constraints
- [ ] Invariant: Cross-repo deps only visible with appropriate grants
- [ ] Shrinking produces minimal failing examples
- [ ] CI runs property tests on every PR
- [ ] Coverage integrated with existing test infrastructure

## Test Cases

- `tests/property/test_access_control.py` - Access control invariants
- `tests/property/test_grant_delegation.py` - Grant delegation properties
- `tests/property/test_shadow_visibility.py` - Shadow requirement properties
- `tests/property/test_cross_repo.py` - Cross-repo access properties

## Technical Notes

### Stateful Testing Model

```python
# tests/property/test_access_control.py
from hypothesis import given, strategies as st
from hypothesis.stateful import RuleBasedStateMachine, rule, invariant

class TrustFederationMachine(RuleBasedStateMachine):
    """State machine for trust graph property testing."""

    def __init__(self):
        super().__init__()
        self.users: set[str] = set()
        self.repos: set[str] = set()
        self.grants: dict[tuple[str, str], set[str]] = {}  # (user, repo) -> roles

    @rule(user=st.text(min_size=1, max_size=20))
    def add_user(self, user: str):
        self.users.add(user)

    @rule(repo=st.text(min_size=1, max_size=50))
    def add_repo(self, repo: str):
        self.repos.add(repo)

    @rule(
        user=st.sampled_from(lambda self: list(self.users) or ["default"]),
        repo=st.sampled_from(lambda self: list(self.repos) or ["default"]),
        roles=st.sets(st.sampled_from(["viewer", "editor", "admin"]))
    )
    def grant_access(self, user: str, repo: str, roles: set[str]):
        if user in self.users and repo in self.repos:
            key = (user, repo)
            self.grants[key] = self.grants.get(key, set()) | roles

    @rule(
        user=st.sampled_from(lambda self: list(self.users) or ["default"]),
        repo=st.sampled_from(lambda self: list(self.repos) or ["default"])
    )
    def revoke_access(self, user: str, repo: str):
        key = (user, repo)
        self.grants.pop(key, None)

    @invariant()
    def no_privilege_escalation(self):
        """Users cannot access more than granted."""
        for (user, repo), roles in self.grants.items():
            actual_access = self.check_access(user, repo)
            assert actual_access <= roles, f"{user} escalated to {actual_access}"

    @invariant()
    def revocation_immediate(self):
        """Revoked grants have no access."""
        for user in self.users:
            for repo in self.repos:
                if (user, repo) not in self.grants:
                    assert not self.can_access(user, repo)

TestTrustFederation = TrustFederationMachine.TestCase
```

### Strategy Definitions

```python
# tests/property/generators.py
from hypothesis import strategies as st

# User strategies
users = st.text(
    alphabet=st.characters(whitelist_categories=("L", "N")),
    min_size=1,
    max_size=50
)

# Repository strategies
repos = st.from_regex(r"[a-z0-9_-]+/[a-z0-9_-]+", fullmatch=True)

# Role strategies
roles = st.sampled_from([
    "dependency_viewer",
    "status_observer",
    "requirement_editor",
    "admin"
])

# Grant strategies
grants = st.builds(
    Grant,
    grantor=repos,
    grantee=repos,
    roles=st.sets(roles, min_size=1),
    constraints=st.fixed_dictionaries({
        "categories": st.lists(st.text(min_size=1), max_size=5)
    })
)
```

### CI Integration

```yaml
# .github/workflows/property-tests.yml
property-tests:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Run property tests
      run: |
        pytest tests/property/ -v \
          --hypothesis-show-statistics \
          --hypothesis-seed=0 \
          -x
```

## Files to Create/Modify

- `tests/property/__init__.py` - Property test package
- `tests/property/generators.py` - Hypothesis strategies
- `tests/property/invariants.py` - Security invariant definitions
- `tests/property/test_access_control.py` - Access control tests
- `tests/property/test_grant_delegation.py` - Grant tests
- `.github/workflows/property-tests.yml` - CI workflow

## Dependencies

- REQ-COLLAB-003: Grant delegation (tested by property tests)
- REQ-ZT-003: JWT validation (tested by property tests)

## Blocks

- REQ-VERIFY-004: NIST compliance evidence from property tests
