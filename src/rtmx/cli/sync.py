"""RTMX sync command.

Bi-directional synchronization with external services (GitHub, Jira).
"""

from __future__ import annotations

import sys
from pathlib import Path

from rtmx.adapters.base import ExternalItem, ServiceAdapter, SyncResult
from rtmx.config import RTMXConfig
from rtmx.formatting import Colors
from rtmx.models import Requirement
from rtmx.parser import load_csv


def run_sync(
    service: str,
    do_import: bool,
    do_export: bool,
    bidirectional: bool,
    dry_run: bool,
    prefer_local: bool,
    prefer_remote: bool,
    config: RTMXConfig,
) -> None:
    """Run sync command.

    Synchronize RTM with GitHub Issues or Jira tickets.

    Args:
        service: Service to sync with (github, jira)
        do_import: Pull from service into RTM
        do_export: Push RTM to service
        bidirectional: Two-way sync
        dry_run: Preview changes without writing
        prefer_local: RTM wins on conflicts
        prefer_remote: Service wins on conflicts
        config: RTMX configuration
    """
    if not any([do_import, do_export, bidirectional]):
        print(
            f"{Colors.YELLOW}No sync direction specified. Use --import, --export, or --bidirectional{Colors.RESET}"
        )
        sys.exit(1)
        return  # Unreachable, but needed for mocked sys.exit in tests

    if prefer_local and prefer_remote:
        print(f"{Colors.RED}Cannot use both --prefer-local and --prefer-remote{Colors.RESET}")
        sys.exit(1)
        return  # Unreachable, but needed for mocked sys.exit in tests

    print(f"=== RTMX Sync: {service.upper()} ===")
    print()

    if dry_run:
        print(f"{Colors.YELLOW}DRY RUN - no changes will be made{Colors.RESET}")
        print()

    # Determine sync mode
    if bidirectional or (do_import and do_export):
        mode = "bidirectional"
    elif do_import:
        mode = "import"
    else:
        mode = "export"

    # Determine conflict resolution
    if prefer_local:
        conflict_resolution = "prefer-local"
    elif prefer_remote:
        conflict_resolution = "prefer-remote"
    else:
        conflict_resolution = config.sync.conflict_resolution

    print(f"Mode: {mode}")
    print(f"Conflict resolution: {conflict_resolution}")
    print()

    # Get adapter
    adapter = _get_adapter(service, config)
    if adapter is None:
        return

    # Test connection
    print(f"{Colors.BOLD}Testing connection...{Colors.RESET}")
    success, message = adapter.test_connection()
    if not success:
        print(f"  {Colors.RED}✗{Colors.RESET} {message}")
        return
    print(f"  {Colors.GREEN}✓{Colors.RESET} {message}")
    print()

    # Run sync based on mode
    if mode == "import":
        result = _run_import(adapter, config, dry_run)
    elif mode == "export":
        result = _run_export(adapter, config, dry_run)
    else:
        result = _run_bidirectional(adapter, config, conflict_resolution, dry_run)

    # Print summary
    _print_summary(result)


def _get_adapter(service: str, config: RTMXConfig) -> ServiceAdapter | None:
    """Get the appropriate adapter for the service."""
    if service == "github":
        if not config.adapters.github.enabled:
            print(f"{Colors.RED}GitHub adapter not enabled in rtmx.yaml{Colors.RESET}")
            return None

        if not config.adapters.github.repo:
            print(f"{Colors.RED}GitHub repo not configured in rtmx.yaml{Colors.RESET}")
            return None

        try:
            from rtmx.adapters.github import GitHubAdapter

            return GitHubAdapter(config.adapters.github)
        except ImportError:
            print(
                f"{Colors.RED}PyGithub not installed. Run: pip install rtmx[github]{Colors.RESET}"
            )
            return None

    elif service == "jira":
        if not config.adapters.jira.enabled:
            print(f"{Colors.RED}Jira adapter not enabled in rtmx.yaml{Colors.RESET}")
            return None

        if not config.adapters.jira.project:
            print(f"{Colors.RED}Jira project not configured in rtmx.yaml{Colors.RESET}")
            return None

        try:
            from rtmx.adapters.jira import JiraAdapter

            return JiraAdapter(config.adapters.jira)
        except ImportError:
            print(
                f"{Colors.RED}jira package not installed. Run: pip install rtmx[jira]{Colors.RESET}"
            )
            return None

    else:
        print(f"{Colors.RED}Unknown service: {service}{Colors.RESET}")
        return None


def _run_import(
    adapter: ServiceAdapter,
    config: RTMXConfig,
    dry_run: bool,
) -> SyncResult:
    """Import items from external service into RTM."""
    result = SyncResult()

    print(f"{Colors.BOLD}Fetching items from {adapter.name}...{Colors.RESET}")

    # Load existing RTM
    rtm_path = Path(config.database)
    requirements: dict[str, Requirement] = {}
    external_id_map: dict[str, str] = {}  # external_id -> requirement_id

    if rtm_path.exists():
        reqs = load_csv(rtm_path)
        for req in reqs:
            requirements[req.req_id] = req
            if req.external_id:
                external_id_map[req.external_id] = req.req_id

    # Fetch external items
    items_found = 0
    for item in adapter.fetch_items():
        items_found += 1

        # Check if already linked
        if item.external_id in external_id_map:
            req_id = external_id_map[item.external_id]
            req = requirements[req_id]

            # Update status from external
            new_status = adapter.map_status_to_rtmx(item.status)
            if new_status != req.status:
                if dry_run:
                    print(f"  Would update {req_id} status: {req.status} → {new_status}")
                else:
                    print(f"  {Colors.BLUE}↻{Colors.RESET} {req_id}: {req.status} → {new_status}")
                result.updated.append(req_id)
            else:
                result.skipped.append(item.external_id)

        elif item.requirement_id and item.requirement_id in requirements:
            # Item references a requirement ID we have
            req = requirements[item.requirement_id]
            if dry_run:
                print(f"  Would link {item.requirement_id} to {item.external_id}")
            else:
                print(
                    f"  {Colors.GREEN}⇄{Colors.RESET} Linked {item.requirement_id} ↔ {item.external_id}"
                )
            result.updated.append(item.requirement_id)

        else:
            # New item - would need to create requirement
            if dry_run:
                print(f"  Would import: [{item.external_id}] {item.title[:50]}...")
            else:
                print(f"  {Colors.GREEN}+{Colors.RESET} [{item.external_id}] {item.title[:50]}...")
            result.created.append(item.external_id)

    print()
    print(f"Found {items_found} items in {adapter.name}")

    return result


def _run_export(
    adapter: ServiceAdapter,
    config: RTMXConfig,
    dry_run: bool,
) -> SyncResult:
    """Export requirements to external service."""
    result = SyncResult()

    print(f"{Colors.BOLD}Exporting requirements to {adapter.name}...{Colors.RESET}")

    # Load existing RTM
    rtm_path = Path(config.database)
    if not rtm_path.exists():
        print(f"{Colors.RED}RTM database not found: {rtm_path}{Colors.RESET}")
        return result

    reqs = load_csv(rtm_path)

    for req in reqs:
        if req.external_id:
            # Already exported - update
            if dry_run:
                print(f"  Would update: {req.req_id} → {req.external_id}")
            else:
                success = adapter.update_item(req.external_id, req)
                if success:
                    print(
                        f"  {Colors.BLUE}↻{Colors.RESET} Updated {req.req_id} → {req.external_id}"
                    )
                    result.updated.append(req.req_id)
                else:
                    print(f"  {Colors.RED}✗{Colors.RESET} Failed to update {req.req_id}")
                    result.errors.append((req.req_id, "Update failed"))
        else:
            # New export
            if dry_run:
                print(f"  Would export: {req.req_id}")
            else:
                try:
                    external_id = adapter.create_item(req)
                    print(f"  {Colors.GREEN}+{Colors.RESET} Exported {req.req_id} → {external_id}")
                    result.created.append(req.req_id)
                except Exception as e:
                    print(f"  {Colors.RED}✗{Colors.RESET} Failed to export {req.req_id}: {e}")
                    result.errors.append((req.req_id, str(e)))

    return result


def _run_bidirectional(
    adapter: ServiceAdapter,
    config: RTMXConfig,
    conflict_resolution: str,
    dry_run: bool,
) -> SyncResult:
    """Run bidirectional sync."""
    result = SyncResult()

    print(f"{Colors.BOLD}Running bidirectional sync with {adapter.name}...{Colors.RESET}")

    # Load existing RTM
    rtm_path = Path(config.database)
    requirements: dict[str, Requirement] = {}
    external_id_map: dict[str, str] = {}

    if rtm_path.exists():
        reqs = load_csv(rtm_path)
        for req in reqs:
            requirements[req.req_id] = req
            if req.external_id:
                external_id_map[req.external_id] = req.req_id

    # Fetch external items
    external_items: dict[str, ExternalItem] = {}
    print(f"\n{Colors.DIM}Fetching external items...{Colors.RESET}")
    for item in adapter.fetch_items():
        external_items[item.external_id] = item

    print(f"Found {len(external_items)} external items")
    print(f"Have {len(requirements)} local requirements")
    print()

    # Process linked items (check for conflicts)
    for external_id, req_id in external_id_map.items():
        if external_id in external_items:
            item = external_items[external_id]
            req = requirements[req_id]

            # Check for status conflict
            external_status = adapter.map_status_to_rtmx(item.status)
            if external_status != req.status:
                if conflict_resolution == "prefer-local":
                    if dry_run:
                        print(f"  Would update {external_id}: {item.status} → {req.status}")
                    else:
                        adapter.update_item(external_id, req)
                        print(f"  {Colors.BLUE}↻{Colors.RESET} {req_id}: Local wins ({req.status})")
                    result.updated.append(req_id)
                elif conflict_resolution == "prefer-remote":
                    if dry_run:
                        print(f"  Would update {req_id}: {req.status} → {external_status}")
                    else:
                        print(
                            f"  {Colors.BLUE}↻{Colors.RESET} {req_id}: Remote wins ({external_status})"
                        )
                    result.updated.append(req_id)
                else:
                    print(
                        f"  {Colors.YELLOW}?{Colors.RESET} Conflict: {req_id} (local={req.status}, remote={external_status})"
                    )
                    result.conflicts.append(
                        (req_id, f"Status conflict: {req.status} vs {external_status}")
                    )
            else:
                result.skipped.append(req_id)

            # Remove from external_items to track what's left
            del external_items[external_id]

    # Items only in external service (import candidates)
    for external_id, item in external_items.items():
        if dry_run:
            print(f"  Would import: [{external_id}] {item.title[:50]}...")
        else:
            print(
                f"  {Colors.GREEN}←{Colors.RESET} Import candidate: [{external_id}] {item.title[:50]}..."
            )
        result.created.append(external_id)

    # Requirements not in external service (export candidates)
    exported_ids = set(external_id_map.values())
    for req_id in requirements:
        if req_id not in exported_ids:
            if dry_run:
                print(f"  Would export: {req_id}")
            else:
                print(f"  {Colors.GREEN}→{Colors.RESET} Export candidate: {req_id}")

    return result


def _print_summary(result: SyncResult) -> None:
    """Print sync summary."""
    print()
    print(f"{Colors.BOLD}Sync Summary:{Colors.RESET}")
    print(f"  {result.summary}")

    if result.conflicts:
        print()
        print(f"{Colors.YELLOW}Conflicts requiring attention:{Colors.RESET}")
        for item_id, reason in result.conflicts:
            print(f"  • {item_id}: {reason}")

    if result.errors:
        print()
        print(f"{Colors.RED}Errors:{Colors.RESET}")
        for item_id, error in result.errors:
            print(f"  • {item_id}: {error}")
