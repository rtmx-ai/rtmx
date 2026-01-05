# REQ-SEC-004: Role-Based Access Control

## Requirement
Projects shall enforce role-based access control.

## Phase
10 (Collaboration) - Enterprise Security

## Rationale
Multi-user collaboration requires granular permissions. Different team members need different access levels - some can edit requirements, others can only view, and project owners manage membership.

## Acceptance Criteria
- [ ] Three roles defined: Owner, Editor, Viewer
- [ ] Owner can manage project membership
- [ ] Owner can delete project
- [ ] Editor can create/update/delete requirements
- [ ] Viewer can only read requirements
- [ ] Role changes take effect immediately
- [ ] API endpoints enforce role checks
- [ ] UI reflects user's permissions

## Role Definitions

| Role | Create | Read | Update | Delete | Manage Members | Delete Project |
|------|--------|------|--------|--------|----------------|----------------|
| Owner | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |
| Editor | ✓ | ✓ | ✓ | ✓ | ✗ | ✗ |
| Viewer | ✗ | ✓ | ✗ | ✗ | ✗ | ✗ |

## Technical Notes
- Store roles in project membership table
- Check permissions in middleware/decorator
- Cache role lookups for performance
- Consider attribute-based access control (ABAC) for future flexibility

## Test Cases
1. Owner can invite new members
2. Editor can modify requirements
3. Viewer cannot modify requirements
4. Role change is enforced immediately
5. Removed user loses access immediately

## Dependencies
- REQ-SEC-003 (authentication exists)

## Effort
2.0 weeks
