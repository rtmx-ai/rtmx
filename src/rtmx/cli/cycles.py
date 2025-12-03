"""RTMX cycles command.

Detect circular dependencies using Tarjan's algorithm.
"""

from __future__ import annotations

import sys
from pathlib import Path

from rtmx.formatting import Colors, header
from rtmx.models import RTMDatabase, RTMError


def run_cycles(rtm_csv: Path | None) -> None:
    """Run cycles command.

    Args:
        rtm_csv: Path to RTM CSV or None to auto-discover
    """
    try:
        db = RTMDatabase.load(rtm_csv)
    except RTMError as e:
        print(f"{Colors.RED}Error: {e}{Colors.RESET}", file=sys.stderr)
        sys.exit(1)

    print(header("Circular Dependency Analysis", "="))
    print()

    # Get graph statistics
    graph = db._get_graph()
    stats = graph.statistics()

    print(f"RTM Statistics:")
    print(f"  Total requirements: {stats['nodes']}")
    print(f"  Total dependencies: {stats['edges']}")
    print(f"  Average dependencies per requirement: {stats['avg_dependencies']:.2f}")
    print()

    # Find cycles
    cycles = db.find_cycles()

    if not cycles:
        print(f"{Colors.GREEN}✓ NO CIRCULAR DEPENDENCIES FOUND{Colors.RESET}")
        print()
        print("The dependency graph is acyclic (DAG). This is ideal for requirements management.")
        sys.exit(0)

    print(f"{Colors.RED}✗ FOUND {len(cycles)} CIRCULAR DEPENDENCY GROUP(S){Colors.RESET}")
    print()

    # Statistics
    total_in_cycles = sum(len(c) for c in cycles)
    cycles_sorted = sorted(cycles, key=len, reverse=True)

    print(f"Summary:")
    print(f"  Circular dependency groups: {len(cycles)}")
    print(f"  Requirements involved in cycles: {total_in_cycles}")
    print(f"  Largest cycle: {len(cycles_sorted[0])} requirements")
    print(f"  Smallest cycle: {len(cycles_sorted[-1])} requirements")
    print()

    # Show top cycles
    print(f"{'-' * 80}")
    print("TOP 10 LARGEST CIRCULAR DEPENDENCY GROUPS:")
    print(f"{'-' * 80}")

    for i, cycle in enumerate(cycles_sorted[:10], 1):
        cycle_set = set(cycle)
        path = graph.find_cycle_path(cycle_set)

        print(f"\n{i}. Cycle with {len(cycle)} requirements:")

        # Show path
        if len(path) <= 8:
            print(f"   Path: {' → '.join(path)}")
        else:
            print(f"   Path: {' → '.join(path[:4])} ... → {' → '.join(path[-3:])}")

        # Show all members if more than path
        if len(cycle) > len(path):
            print(f"   All members: {', '.join(sorted(cycle))}")

    # Recommendations
    print(f"\n{'=' * 80}")
    print("RECOMMENDATIONS:")
    print(f"{'=' * 80}")
    print("""
1. Review dependency direction:
   - Ensure parent requirements don't depend on child requirements
   - Component requirements should depend on system requirements, not vice versa

2. Examine the largest cycles first:
   - These likely indicate architectural issues
   - May need to split into layers or stages

3. Check for "blocks" vs "dependencies" confusion:
   - If A blocks B, then B depends on A (not both!)
   - Run: rtmx reconcile --execute

4. Consider adding phase constraints:
   - Requirements should not depend on later-phase requirements

5. Total effort to fix:
   - {total} requirements involved in {cycles} cycles
   - Suggest reviewing in batches: largest cycles first
    """.format(total=total_in_cycles, cycles=len(cycles)))

    sys.exit(1)  # Exit with error if cycles found
