# REQ-INT-003: Git Branch Protection Integration

## Metadata
- **Category**: INTEGRITY
- **Subcategory**: Centralized
- **Priority**: MEDIUM
- **Phase**: 17
- **Status**: MISSING
- **Dependencies**: REQ-INT-001

## Requirement

RTMX shall provide tooling and documentation for enforcing database integrity via Git branch protection, where only CI/CD pipelines can merge status changes to protected branches.

## Rationale

For teams using centralized Git hosting (GitHub, GitLab, Bitbucket), branch protection rules provide enforceable access control without requiring local privilege escalation. This is the pragmatic near-term solution while CRDT-based enforcement matures.

## Design

### Workflow

```
1. Agent/developer works on feature branch
2. Agent proposes status change (MISSING → COMPLETE)
3. PR/MR created targeting protected branch
4. CI pipeline runs:
   a. Checkout branch
   b. Run `rtmx verify --update`
   c. If status differs from proposed, fail the check
   d. If status matches (tests pass), approve
5. Only CI bot can merge to protected branch
```

### Required Components

1. **GitHub Action / GitLab CI template**: Ready-to-use CI configuration
2. **Branch protection documentation**: Setup guide for each platform
3. **Status check enforcement**: CI job that validates proposed changes
4. **Merge bot configuration**: Automated merge on check pass

### Platform Support

| Platform | Branch Protection | Status Checks | Merge Bots |
|----------|-------------------|---------------|------------|
| GitHub | Yes | Yes | GitHub Actions |
| GitLab | Yes | Yes | GitLab CI |
| Bitbucket | Yes | Yes | Bitbucket Pipelines |
| Gitea | Partial | Yes | Drone CI |

## Acceptance Criteria

1. GitHub Action available in marketplace or as reusable workflow
2. Documentation covers setup for GitHub, GitLab, Bitbucket
3. Agent cannot directly push status changes to protected branch
4. CI pipeline correctly validates status changes via `rtmx verify`
5. False positive rate < 1% (legitimate changes blocked incorrectly)

## Test Strategy

- Integration tests with GitHub Actions (self-hosted runner)
- Manual testing of branch protection bypass attempts
- Documentation review by external users

## Limitations

- Requires centralized Git hosting with branch protection support
- Does not work for air-gapped or fully decentralized workflows
- Agents can still create valid-looking PRs (social engineering risk)

## References

- GitHub branch protection rules
- GitLab protected branches
- REQ-INT-002 for decentralized alternative
