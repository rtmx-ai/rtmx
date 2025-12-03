"""RTMX reconcile command.

Check and fix dependency/blocks reciprocity.
"""

from __future__ import annotations

import sys
from pathlib import Path

from rtmx.formatting import Colors, header
from rtmx.models import RTMDatabase, RTMError
from rtmx.validation import check_reciprocity, fix_reciprocity


def run_reconcile(rtm_csv: Path | None, execute: bool) -> None:
    """Run reconcile command.

    Args:
        rtm_csv: Path to RTM CSV or None to auto-discover
        execute: If True, fix violations; otherwise dry-run
    """
    try:
        db = RTMDatabase.load(rtm_csv)
    except RTMError as e:
        print(f"{Colors.RED}Error: {e}{Colors.RESET}", file=sys.stderr)
        sys.exit(1)

    print(header("Reciprocity Check", "="))
    print()

    # Check for violations
    violations = check_reciprocity(db)

    if not violations:
        print(f"{Colors.GREEN}✓ No reciprocity violations found{Colors.RESET}")
        sys.exit(0)

    print(f"{Colors.YELLOW}Found {len(violations)} reciprocity violation(s):{Colors.RESET}")
    print()

    for req_id, related_id, issue in violations[:20]:
        print(f"  {Colors.RED}✗{Colors.RESET} {req_id} <-> {related_id}: {issue}")

    if len(violations) > 20:
        print(f"  ... and {len(violations) - 20} more")

    if execute:
        print()
        print(f"{Colors.BOLD}Fixing violations...{Colors.RESET}")
        fixed = fix_reciprocity(db)
        db.save()
        print(f"{Colors.GREEN}✓ Fixed {fixed} violation(s){Colors.RESET}")

        # Verify
        remaining = check_reciprocity(db)
        if remaining:
            print(f"{Colors.YELLOW}⚠ {len(remaining)} violations remain{Colors.RESET}")
        else:
            print(f"{Colors.GREEN}✓ All violations resolved{Colors.RESET}")
    else:
        print()
        print(f"{Colors.DIM}Run with --execute to fix violations{Colors.RESET}")
        sys.exit(1)
