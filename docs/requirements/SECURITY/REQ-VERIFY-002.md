# REQ-VERIFY-002: TLA+ Formal Specification

## Status: NOT STARTED
## Priority: MEDIUM
## Phase: 13
## Effort: 3.0 weeks

## Description

RTMX shall have a TLA+ formal specification for the trust federation protocol, allowing model checking to verify safety and liveness properties. The specification shall prove that no sequence of operations can violate access control invariants, providing mathematical evidence of correctness for CISO review.

## Acceptance Criteria

- [ ] TLA+ specification models trust federation state machine
- [ ] Specification includes users, repos, grants, and operations
- [ ] Safety property: No privilege escalation
- [ ] Safety property: Grants bounded by grantor permissions
- [ ] Liveness property: Eventually consistent after network heal
- [ ] TLC model checker verifies all properties
- [ ] TLAPS proof checked for core invariants
- [ ] CI runs TLC on specification changes
- [ ] Documentation explains spec to non-TLA+ readers

## Test Cases

- `specs/TrustFederation.tla` - Main specification
- `specs/TrustFederationMC.tla` - Model checking configuration
- `specs/TrustFederation.toolbox/` - TLA+ Toolbox project

## Technical Notes

### TLA+ Specification

```tla
---- MODULE TrustFederation ----
EXTENDS Integers, Sequences, FiniteSets

CONSTANTS Users, Repos, Roles

VARIABLES
    grants,     \* Function from (user, repo) to set of roles
    shadows,    \* Set of (user, repo, req_id) shadow visibility
    pending     \* Set of pending grant operations

TypeInvariant ==
    /\ grants \in [Users \X Repos -> SUBSET Roles]
    /\ shadows \in SUBSET (Users \X Repos \X ReqIds)
    /\ pending \in SUBSET GrantOperations

------------------------------------------------------------
\* Actions

GrantAccess(user, repo, role) ==
    /\ role \in Roles
    /\ grants' = [grants EXCEPT ![user, repo] = @ \cup {role}]
    /\ UNCHANGED <<shadows, pending>>

RevokeAccess(user, repo) ==
    /\ grants' = [grants EXCEPT ![user, repo] = {}]
    /\ shadows' = shadows \ {<<user, repo, r>> : r \in ReqIds}
    /\ UNCHANGED pending

------------------------------------------------------------
\* Invariants

NoPrivilegeEscalation ==
    \A u \in Users, r \in Repos, role \in Roles:
        CanAccess(u, r, role) => HasGrant(u, r, role)

DelegationBounded ==
    \A g \in pending:
        g.roles \subseteq GetRoles(g.grantor, g.repo)

ShadowVisibilityRespected ==
    \A <<u, r, req>> \in shadows:
        HasGrant(u, r, "dependency_viewer")

------------------------------------------------------------
\* Specification

Init ==
    /\ grants = [u \in Users, r \in Repos |-> {}]
    /\ shadows = {}
    /\ pending = {}

Next ==
    \/ \E u \in Users, r \in Repos, role \in Roles:
        GrantAccess(u, r, role)
    \/ \E u \in Users, r \in Repos:
        RevokeAccess(u, r)

Spec == Init /\ [][Next]_<<grants, shadows, pending>>

------------------------------------------------------------
\* Properties to verify

THEOREM Spec => []TypeInvariant
THEOREM Spec => []NoPrivilegeEscalation
THEOREM Spec => []DelegationBounded
THEOREM Spec => []ShadowVisibilityRespected

====
```

### Model Checking Configuration

```tla
---- MODULE TrustFederationMC ----
EXTENDS TrustFederation

\* Small model for exhaustive checking
MCUsers == {"alice", "bob", "charlie"}
MCRepos == {"rtmx", "rtmx-sync"}
MCRoles == {"viewer", "editor", "admin"}

====
```

### CI Integration

```yaml
# .github/workflows/formal-verification.yml
formal-verification:
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - name: Install TLA+ tools
      run: |
        wget https://github.com/tlaplus/tlaplus/releases/download/v1.8.0/tla2tools.jar
    - name: Run TLC model checker
      run: |
        java -jar tla2tools.jar -workers auto specs/TrustFederationMC.tla
```

### Documentation

The specification shall be accompanied by:
1. English prose explanation of each invariant
2. Mapping from spec to implementation code
3. CISO-friendly summary of verified properties
4. Counter-example analysis if model checking fails

## Files to Create/Modify

- `specs/TrustFederation.tla` - Main TLA+ specification
- `specs/TrustFederationMC.tla` - Model checking config
- `specs/README.md` - Specification documentation
- `.github/workflows/formal-verification.yml` - CI workflow
- `docs/security/formal-verification.md` - CISO documentation

## Dependencies

- REQ-COLLAB-003: Grant delegation (modeled in spec)
- REQ-ZT-003: JWT validation (modeled in spec)

## Blocks

- REQ-VERIFY-004: NIST compliance uses TLA+ as evidence
