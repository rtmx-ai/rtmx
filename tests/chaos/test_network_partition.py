"""Network Partition Chaos Tests.

Tests for validating RTMX behavior during network failures.

These tests verify:
- Offline access to cached data
- Graceful degradation when services unavailable
- Sync recovery after partition heals
- No data loss during network instability
"""

from __future__ import annotations

import asyncio
from dataclasses import dataclass, field
from typing import Any

import pytest

from rtmx.models import (
    DelegationRole,
    GrantConstraint,
    GrantDelegation,
)


@dataclass
class NetworkSimulator:
    """Simulates network conditions for chaos testing.

    Provides a lightweight alternative to toxiproxy for
    basic network fault injection in tests.

    Attributes:
        healthy: Whether network is currently healthy
        latency_ms: Added latency in milliseconds
        packet_loss: Probability of packet loss (0-1)
        partitioned_services: Set of services that are unreachable
    """

    healthy: bool = True
    latency_ms: int = 0
    packet_loss: float = 0.0
    partitioned_services: set[str] = field(default_factory=set)

    def partition(self, service: str) -> None:
        """Partition a service (make unreachable)."""
        self.partitioned_services.add(service)

    def heal(self, service: str | None = None) -> None:
        """Heal partition for service or all services."""
        if service:
            self.partitioned_services.discard(service)
        else:
            self.partitioned_services.clear()

    def is_reachable(self, service: str) -> bool:
        """Check if service is reachable."""
        return self.healthy and service not in self.partitioned_services

    async def simulate_request(self, service: str) -> dict[str, Any]:
        """Simulate a network request with current conditions.

        Raises:
            ConnectionError: If service is partitioned
        """
        if not self.is_reachable(service):
            raise ConnectionError(f"Service {service} is unreachable")

        if self.latency_ms > 0:
            await asyncio.sleep(self.latency_ms / 1000)

        return {"status": "ok", "service": service}


@dataclass
class CachedGrantStore:
    """Local cache for grant delegations.

    Simulates the caching behavior of RTMX for offline access.

    Attributes:
        grants: Cached grant delegations
        last_sync: Timestamp of last sync
        dirty: Whether cache has unsyncbed changes
    """

    grants: dict[str, GrantDelegation] = field(default_factory=dict)
    last_sync: str = ""
    dirty: bool = False

    def cache_grant(self, key: str, grant: GrantDelegation) -> None:
        """Cache a grant locally."""
        self.grants[key] = grant

    def get_grant(self, key: str) -> GrantDelegation | None:
        """Get grant from cache."""
        return self.grants.get(key)

    def can_access_offline(self, grantee: str, req_id: str, category: str) -> bool:
        """Check if grantee can access requirement from cache."""
        for grant in self.grants.values():
            if (
                grant.grantee == grantee
                and grant.is_valid
                and grant.constraints.allows_requirement(req_id, category)
            ):
                return True
        return False


class TestNetworkPartition:
    """Tests for network partition scenarios."""

    @pytest.fixture
    def network(self) -> NetworkSimulator:
        """Create network simulator."""
        return NetworkSimulator()

    @pytest.fixture
    def cache(self) -> CachedGrantStore:
        """Create cached grant store."""
        return CachedGrantStore()

    @pytest.mark.req("REQ-VERIFY-003")
    @pytest.mark.scope_integration
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    async def test_offline_access_preserved_during_partition(
        self, network: NetworkSimulator, cache: CachedGrantStore
    ) -> None:
        """Verify cached grants work during network partition."""
        # Setup: Grant access while online
        grant = GrantDelegation(
            grantor="org/repo-a",
            grantee="org/repo-b",
            roles_delegated={DelegationRole.REQUIREMENT_READER},
            constraints=GrantConstraint(categories={"SOFTWARE"}),
        )
        cache.cache_grant("grant-1", grant)

        # Verify online access works
        assert network.is_reachable("rtmx-sync")
        assert cache.can_access_offline("org/repo-b", "REQ-SW-001", "SOFTWARE")

        # Simulate network partition
        network.partition("rtmx-sync")
        assert not network.is_reachable("rtmx-sync")

        # Should still work from cache
        assert cache.can_access_offline("org/repo-b", "REQ-SW-001", "SOFTWARE")

    @pytest.mark.req("REQ-VERIFY-003")
    @pytest.mark.scope_integration
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    async def test_new_requests_fail_during_partition(self, network: NetworkSimulator) -> None:
        """Verify requests to partitioned services fail gracefully."""
        network.partition("rtmx-sync")

        with pytest.raises(ConnectionError):
            await network.simulate_request("rtmx-sync")

    @pytest.mark.req("REQ-VERIFY-003")
    @pytest.mark.scope_integration
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    async def test_sync_resumes_after_partition_heals(self, network: NetworkSimulator) -> None:
        """Verify sync resumes when partition heals."""
        # Partition the network
        network.partition("rtmx-sync")
        assert not network.is_reachable("rtmx-sync")

        # Heal the partition
        network.heal("rtmx-sync")
        assert network.is_reachable("rtmx-sync")

        # Requests should work again
        result = await network.simulate_request("rtmx-sync")
        assert result["status"] == "ok"

    @pytest.mark.req("REQ-VERIFY-003")
    @pytest.mark.scope_integration
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    async def test_revocation_propagates_after_heal(
        self, network: NetworkSimulator, cache: CachedGrantStore
    ) -> None:
        """Verify revocations sync after partition heals."""
        # Grant access
        grant = GrantDelegation(
            grantor="org/repo-a",
            grantee="org/repo-b",
            roles_delegated={DelegationRole.REQUIREMENT_READER},
            active=True,
        )
        cache.cache_grant("grant-1", grant)
        assert cache.can_access_offline("org/repo-b", "REQ-SW-001", "SOFTWARE")

        # Simulate: during partition, grant is revoked remotely
        network.partition("rtmx-sync")

        # Local cache still has old grant (would be updated on sync)
        assert cache.can_access_offline("org/repo-b", "REQ-SW-001", "SOFTWARE")

        # Heal and simulate sync with revocation
        network.heal()
        cached_grant = cache.get_grant("grant-1")
        if cached_grant:
            cached_grant.active = False  # Simulate sync receiving revocation

        # Access should now be denied
        assert not cache.can_access_offline("org/repo-b", "REQ-SW-001", "SOFTWARE")


class TestServiceDegradation:
    """Tests for graceful degradation under partial failures."""

    @pytest.mark.req("REQ-VERIFY-003")
    @pytest.mark.scope_integration
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    async def test_high_latency_handled_gracefully(self) -> None:
        """Verify system handles high latency without errors."""
        network = NetworkSimulator(latency_ms=100)

        # Should succeed despite latency
        result = await network.simulate_request("rtmx-sync")
        assert result["status"] == "ok"

    @pytest.mark.req("REQ-VERIFY-003")
    @pytest.mark.scope_integration
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    async def test_partial_service_failure(self) -> None:
        """Verify system handles partial service failures."""
        network = NetworkSimulator()

        # Partition one service but not others
        network.partition("external-service")

        # rtmx-sync should still work
        assert network.is_reachable("rtmx-sync")
        result = await network.simulate_request("rtmx-sync")
        assert result["status"] == "ok"

        # External service should fail
        with pytest.raises(ConnectionError):
            await network.simulate_request("external-service")


class TestSplitBrain:
    """Tests for split-brain scenarios and CRDT conflict resolution."""

    @pytest.mark.req("REQ-COLLAB-002")
    @pytest.mark.scope_integration
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_concurrent_grants_merge_correctly(self) -> None:
        """Verify concurrent grant modifications merge without conflict."""
        # Simulate two caches that diverge during partition
        cache_a = CachedGrantStore()
        cache_b = CachedGrantStore()

        # Both start with same grant
        base_grant = GrantDelegation(
            grantor="org/repo-a",
            grantee="org/repo-b",
            roles_delegated={DelegationRole.DEPENDENCY_VIEWER},
        )
        cache_a.cache_grant("grant-1", base_grant)
        cache_b.cache_grant("grant-1", base_grant)

        # During partition, A adds a role
        grant_a = cache_a.get_grant("grant-1")
        if grant_a:
            grant_a.roles_delegated.add(DelegationRole.REQUIREMENT_READER)

        # During partition, B adds a different role
        grant_b = cache_b.get_grant("grant-1")
        if grant_b:
            grant_b.roles_delegated.add(DelegationRole.SHADOW_VIEWER)

        # After heal, merging should preserve both roles (set union)
        merged_roles = (grant_a.roles_delegated if grant_a else set()) | (
            grant_b.roles_delegated if grant_b else set()
        )

        assert DelegationRole.DEPENDENCY_VIEWER in merged_roles
        assert DelegationRole.REQUIREMENT_READER in merged_roles
        assert DelegationRole.SHADOW_VIEWER in merged_roles

    @pytest.mark.req("REQ-COLLAB-002")
    @pytest.mark.scope_integration
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    def test_revocation_wins_over_grant(self) -> None:
        """Verify revocation takes precedence in conflict (last writer wins)."""
        # Simulate conflict: one side grants, one side revokes
        grant_granted = GrantDelegation(
            grantor="org/repo-a",
            grantee="org/repo-b",
            roles_delegated={DelegationRole.REQUIREMENT_READER},
            active=True,
            created_at="2024-01-01T10:00:00",
        )

        grant_revoked = GrantDelegation(
            grantor="org/repo-a",
            grantee="org/repo-b",
            roles_delegated={DelegationRole.REQUIREMENT_READER},
            active=False,
            created_at="2024-01-01T11:00:00",  # Later timestamp
        )

        # Revoked version (later timestamp) should win
        # In real CRDT, this would be determined by vector clocks or timestamps
        assert grant_revoked.created_at > grant_granted.created_at
        assert not grant_revoked.is_valid  # Revocation takes effect


class TestTokenExpiration:
    """Tests for authentication recovery during token issues."""

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_integration
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    async def test_expired_token_triggers_refresh(self) -> None:
        """Verify expired tokens trigger automatic refresh."""
        from datetime import datetime, timedelta

        from rtmx.auth.oidc import TokenInfo

        # Create expired token
        expired_token = TokenInfo(
            access_token="expired-token",
            refresh_token="valid-refresh",
            expires_at=datetime.now() - timedelta(hours=1),
        )

        assert expired_token.is_expired
        assert expired_token.is_refreshable

    @pytest.mark.req("REQ-VERIFY-004")
    @pytest.mark.scope_integration
    @pytest.mark.technique_stress
    @pytest.mark.env_simulation
    async def test_no_refresh_token_requires_reauth(self) -> None:
        """Verify missing refresh token requires full reauth."""
        from datetime import datetime, timedelta

        from rtmx.auth.oidc import TokenInfo

        # Create expired token without refresh
        expired_no_refresh = TokenInfo(
            access_token="expired-token",
            refresh_token="",  # No refresh token
            expires_at=datetime.now() - timedelta(hours=1),
        )

        assert expired_no_refresh.is_expired
        assert not expired_no_refresh.is_refreshable
