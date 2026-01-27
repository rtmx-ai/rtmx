"""Hypothesis strategies for trust graph testing.

Provides strategies for generating:
- Requirement IDs
- Repository references
- Access grants
- Trust graph states
"""

from __future__ import annotations

from typing import TYPE_CHECKING

from hypothesis import strategies as st

if TYPE_CHECKING:
    from hypothesis.strategies import SearchStrategy


# Repository identifiers
@st.composite
def repos(draw: st.DrawFn) -> str:
    """Generate repository identifiers like 'org/repo'."""
    org = draw(st.text(alphabet="abcdefghijklmnopqrstuvwxyz", min_size=2, max_size=10))
    repo = draw(st.text(alphabet="abcdefghijklmnopqrstuvwxyz-", min_size=2, max_size=15))
    return f"{org}/{repo}"


# Requirement IDs
@st.composite
def requirement_ids(draw: st.DrawFn) -> str:
    """Generate requirement IDs like 'REQ-SW-001'."""
    category = draw(st.sampled_from(["SW", "HW", "PERF", "SEC", "SYNC", "CLI", "ZT"]))
    number = draw(st.integers(min_value=1, max_value=999))
    return f"REQ-{category}-{number:03d}"


# Cross-repo requirement references
@st.composite
def cross_repo_refs(draw: st.DrawFn) -> str:
    """Generate cross-repo requirement references like 'org/repo:REQ-SW-001'."""
    repo = draw(repos())
    req_id = draw(requirement_ids())
    return f"{repo}:{req_id}"


# Permission levels
permissions: SearchStrategy[str] = st.sampled_from(
    [
        "read",
        "write",
        "dependency_viewer",
        "requirement_editor",
        "admin",
    ]
)


# Users
@st.composite
def users(draw: st.DrawFn) -> str:
    """Generate user identifiers."""
    name = draw(st.text(alphabet="abcdefghijklmnopqrstuvwxyz", min_size=3, max_size=12))
    return f"user_{name}"


# Access grant
@st.composite
def grants(draw: st.DrawFn) -> dict:
    """Generate an access grant."""
    return {
        "grantor": draw(repos()),
        "grantee": draw(repos()),
        "user": draw(users()),
        "permission": draw(permissions),
        "requirements": draw(st.lists(requirement_ids(), min_size=0, max_size=5)),
    }


# Trust graph state
@st.composite
def trust_graph_states(
    draw: st.DrawFn,
    max_repos: int = 5,
    max_users: int = 5,
    max_grants: int = 10,
) -> dict:
    """Generate a complete trust graph state for testing.

    Returns:
        Dictionary with:
        - repos: Set of repository identifiers
        - users: Set of user identifiers
        - grants: List of access grants
        - requirements: Dict mapping repo to list of requirement IDs
    """
    # Generate repos
    num_repos = draw(st.integers(min_value=1, max_value=max_repos))
    repo_list = draw(st.lists(repos(), min_size=num_repos, max_size=num_repos, unique=True))

    # Generate users
    num_users = draw(st.integers(min_value=1, max_value=max_users))
    user_list = draw(st.lists(users(), min_size=num_users, max_size=num_users, unique=True))

    # Generate grants between repos and users
    grant_list = []
    num_grants = draw(st.integers(min_value=0, max_value=max_grants))
    for _ in range(num_grants):
        if len(repo_list) >= 2:
            grantor = draw(st.sampled_from(repo_list))
            grantee_options = [r for r in repo_list if r != grantor]
            if grantee_options:
                grantee = draw(st.sampled_from(grantee_options))
                grant_list.append(
                    {
                        "grantor": grantor,
                        "grantee": grantee,
                        "user": draw(st.sampled_from(user_list)),
                        "permission": draw(permissions),
                    }
                )

    # Generate requirements per repo
    requirements: dict[str, list[str]] = {}
    for repo in repo_list:
        num_reqs = draw(st.integers(min_value=1, max_value=10))
        requirements[repo] = draw(
            st.lists(requirement_ids(), min_size=num_reqs, max_size=num_reqs, unique=True)
        )

    return {
        "repos": set(repo_list),
        "users": set(user_list),
        "grants": grant_list,
        "requirements": requirements,
    }
