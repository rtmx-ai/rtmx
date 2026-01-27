"""Chaos Engineering Tests for RTMX.

This package contains tests that validate RTMX behavior under
adverse network conditions and partial system failures.

Test Categories:
- Network partitions: Validate offline-first capabilities
- Service degradation: Test graceful degradation under load
- Split brain: Verify CRDT conflict resolution
- Token expiration: Ensure auth recovery

Requires toxiproxy for full network fault injection, but includes
simulation-based tests that work without external dependencies.
"""
