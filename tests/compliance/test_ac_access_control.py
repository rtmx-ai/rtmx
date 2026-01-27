"""NIST 800-53 AC (Access Control) Family Tests.

Tests for Access Control requirements in RTMX trust federation.

Control mappings:
- AC-2: Account Management
- AC-3: Access Enforcement
- AC-4: Information Flow Enforcement
- AC-6: Least Privilege
- AC-17: Remote Access
- AC-21: Information Sharing
"""

from __future__ import annotations

import pytest

from rtmx.models import (
    DelegationRole,
    GrantConstraint,
    GrantDelegation,
    ShadowRequirement,
    Status,
    Visibility,
)


class TestAC2AccountManagement:
    """AC-2: Account Management.

    The organization:
    a. Identifies and selects types of accounts
    b. Assigns account managers
    c. Establishes conditions for group/role membership
    d. Specifies authorized users and access authorizations

    RTMX Implementation:
    - Users authenticated via Zitadel OIDC
    - Roles defined in DelegationRole enum
    - Grant delegations specify authorized access
    """

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_delegation_roles_are_defined(self) -> None:
        """AC-2(a): System defines distinct account/role types."""
        # Verify all expected roles exist
        expected_roles = {
            "dependency_viewer",
            "requirement_reader",
            "requirement_editor",
            "shadow_viewer",
        }

        actual_roles = {role.value for role in DelegationRole}
        assert expected_roles == actual_roles, "All required roles must be defined"

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_delegation_tracks_creator(self) -> None:
        """AC-2(b): Delegations track who created them (account manager)."""
        delegation = GrantDelegation(
            grantor="org/repo-a",
            grantee="org/repo-b",
            roles_delegated={DelegationRole.DEPENDENCY_VIEWER},
            created_by="admin@example.com",
        )

        assert delegation.created_by == "admin@example.com"
        assert delegation.created_at  # Timestamp auto-set

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_delegation_can_be_deactivated(self) -> None:
        """AC-2(k): Delegations can be disabled/terminated."""
        delegation = GrantDelegation(
            grantor="org/repo-a",
            grantee="org/repo-b",
            roles_delegated={DelegationRole.DEPENDENCY_VIEWER},
            active=True,
        )

        assert delegation.is_valid

        # Deactivate
        delegation.active = False
        assert not delegation.is_valid


class TestAC3AccessEnforcement:
    """AC-3: Access Enforcement.

    The information system enforces approved authorizations for
    logical access in accordance with applicable access control policies.

    RTMX Implementation:
    - Grant delegations enforce access between repos
    - Constraints limit access to specific categories/requirements
    - Shadow requirements provide partial visibility
    """

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_access_requires_valid_delegation(self) -> None:
        """AC-3: Access enforcement via delegation validation."""
        delegation = GrantDelegation(
            grantor="org/repo-a",
            grantee="org/repo-b",
            roles_delegated={DelegationRole.REQUIREMENT_READER},
        )

        # Should allow access with correct role
        assert delegation.allows_access(
            req_id="REQ-SW-001",
            category="SOFTWARE",
            role=DelegationRole.REQUIREMENT_READER,
        )

        # Should deny access with wrong role
        assert not delegation.allows_access(
            req_id="REQ-SW-001",
            category="SOFTWARE",
            role=DelegationRole.REQUIREMENT_EDITOR,
        )

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_inactive_delegation_denies_access(self) -> None:
        """AC-3: Inactive delegations do not grant access."""
        delegation = GrantDelegation(
            grantor="org/repo-a",
            grantee="org/repo-b",
            roles_delegated={DelegationRole.REQUIREMENT_READER},
            active=False,
        )

        assert not delegation.allows_access(
            req_id="REQ-SW-001",
            category="SOFTWARE",
            role=DelegationRole.REQUIREMENT_READER,
        )


class TestAC4InformationFlowEnforcement:
    """AC-4: Information Flow Enforcement.

    The information system enforces approved authorizations for
    controlling the flow of information within the system and
    between interconnected systems.

    RTMX Implementation:
    - Cross-repo dependencies have explicit flow controls
    - Shadow requirements limit information exposure
    - Visibility levels control data exposure
    """

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_shadow_visibility_limits_information_flow(self) -> None:
        """AC-4: Visibility levels control information exposure."""
        # Full visibility - all info flows
        shadow_full = ShadowRequirement(
            req_id="REQ-SW-001",
            external_repo="org/repo",
            shadow_hash="abc123",
            status=Status.COMPLETE,
            visibility=Visibility.FULL,
        )
        assert shadow_full.is_accessible

        # Shadow visibility - limited info
        shadow_limited = ShadowRequirement(
            req_id="REQ-SW-001",
            external_repo="org/repo",
            shadow_hash="abc123",
            status=Status.COMPLETE,
            visibility=Visibility.SHADOW,
        )
        assert not shadow_limited.is_accessible

        # Hash only - minimal info
        shadow_hash = ShadowRequirement(
            req_id="REQ-SW-001",
            external_repo="org/repo",
            shadow_hash="abc123",
            status=Status.COMPLETE,
            visibility=Visibility.HASH_ONLY,
        )
        assert not shadow_hash.is_accessible
        assert shadow_hash.is_verifiable


class TestAC6LeastPrivilege:
    """AC-6: Least Privilege.

    The organization employs the principle of least privilege,
    allowing only authorized accesses for users which are necessary
    to accomplish assigned tasks.

    RTMX Implementation:
    - Role hierarchy from shadow_viewer to requirement_editor
    - Constraints limit access to specific requirements
    - Category-based access restrictions
    """

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_role_hierarchy_supports_least_privilege(self) -> None:
        """AC-6: Roles provide graduated privilege levels."""
        # shadow_viewer has minimal access
        delegation_minimal = GrantDelegation(
            grantor="org/repo-a",
            grantee="org/repo-b",
            roles_delegated={DelegationRole.SHADOW_VIEWER},
        )

        # Should only allow shadow viewing
        assert delegation_minimal.has_role(DelegationRole.SHADOW_VIEWER)
        assert not delegation_minimal.has_role(DelegationRole.REQUIREMENT_READER)
        assert not delegation_minimal.has_role(DelegationRole.REQUIREMENT_EDITOR)

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_constraints_limit_access_scope(self) -> None:
        """AC-6: Constraints restrict access to minimum necessary."""
        constraint = GrantConstraint(
            categories={"SOFTWARE"},
            requirement_ids={"REQ-SW-001", "REQ-SW-002"},
        )

        # Should allow access to specified requirements
        assert constraint.allows_requirement("REQ-SW-001", "SOFTWARE")
        assert constraint.allows_requirement("REQ-SW-002", "SOFTWARE")

        # Should deny access to other requirements
        assert not constraint.allows_requirement("REQ-SW-003", "SOFTWARE")
        assert not constraint.allows_requirement("REQ-HW-001", "HARDWARE")

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_category_exclusions_enforce_separation(self) -> None:
        """AC-6: Category exclusions enforce need-to-know."""
        constraint = GrantConstraint(
            exclude_categories={"SECURITY", "INTERNAL"},
        )

        # Should allow access to non-excluded categories
        assert constraint.allows_requirement("REQ-SW-001", "SOFTWARE")
        assert constraint.allows_requirement("REQ-PERF-001", "PERFORMANCE")

        # Should deny access to excluded categories
        assert not constraint.allows_requirement("REQ-SEC-001", "SECURITY")
        assert not constraint.allows_requirement("REQ-INT-001", "INTERNAL")


class TestAC17RemoteAccess:
    """AC-17: Remote Access.

    The organization establishes and documents usage restrictions,
    configuration/connection requirements, and implementation guidance
    for each type of remote access allowed.

    RTMX Implementation:
    - OpenZiti provides zero-trust remote access
    - No public ports exposed (dark service)
    - All access via identity-verified overlay
    """

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_ziti_config_defines_remote_access(self) -> None:
        """AC-17: Remote access configuration is defined."""
        from rtmx.ziti import ZitiConfig

        config = ZitiConfig()

        # Remote access requires controller
        assert config.controller
        # Identity must be enrolled
        assert config.identity_dir
        # Services define what can be accessed
        assert "rtmx-sync" in config.services


class TestAC21InformationSharing:
    """AC-21: Information Sharing.

    The organization facilitates information sharing by enabling
    authorized users to determine whether access authorizations
    assigned to the sharing partner match the access restrictions
    on the information.

    RTMX Implementation:
    - Grant delegation model for explicit sharing
    - Shadow requirements for partial sharing
    - Hash verification for integrity
    """

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_delegation_enables_controlled_sharing(self) -> None:
        """AC-21: Delegations enable controlled information sharing."""
        # Create delegation for specific sharing
        delegation = GrantDelegation(
            grantor="company-a/requirements",
            grantee="company-b/product",
            roles_delegated={DelegationRole.DEPENDENCY_VIEWER},
            constraints=GrantConstraint(
                categories={"API", "INTERFACE"},
            ),
        )

        # Sharing is explicit and auditable
        assert delegation.grantor == "company-a/requirements"
        assert delegation.grantee == "company-b/product"
        assert delegation.created_at  # When sharing was established

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_unit
    @pytest.mark.technique_nominal
    @pytest.mark.env_simulation
    def test_shadow_hash_enables_verification(self) -> None:
        """AC-21: Hash allows sharing partners to verify data."""
        shadow = ShadowRequirement(
            req_id="REQ-API-001",
            external_repo="partner/repo",
            shadow_hash="abc123def456",
            status=Status.COMPLETE,
            visibility=Visibility.SHADOW,
        )

        # Partner can verify data integrity via hash
        assert shadow.is_verifiable
        assert shadow.shadow_hash == "abc123def456"
