"""RTMX MCP tool definitions.

Defines the tools exposed by the RTMX MCP server.
"""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Any

from rtmx.config import RTMXConfig, load_config
from rtmx.models import RTMDatabase, Status


@dataclass
class ToolResult:
    """Result of a tool invocation."""

    success: bool
    data: Any
    error: str | None = None


class RTMXTools:
    """RTMX tools for MCP server.

    Provides tool implementations that can be exposed via MCP protocol.
    """

    def __init__(self, config: RTMXConfig | None = None) -> None:
        """Initialize RTMX tools.

        Args:
            config: RTMX configuration (loads from file if not provided)
        """
        self._config = config or load_config()
        self._db: RTMDatabase | None = None

    def _get_db(self) -> RTMDatabase:
        """Get or load the RTM database."""
        if self._db is None:
            db_path = Path(self._config.database)
            if not db_path.exists():
                raise FileNotFoundError(f"RTM database not found: {db_path}")
            self._db = RTMDatabase.load(db_path)
        return self._db

    def _reload_db(self) -> RTMDatabase:
        """Force reload the RTM database."""
        self._db = None
        return self._get_db()

    def get_status(self, verbose: int = 0) -> ToolResult:
        """Get RTM completion status.

        Args:
            verbose: Verbosity level (0=summary, 1=categories, 2=requirements)

        Returns:
            ToolResult with status data
        """
        try:
            db = self._get_db()

            # Calculate overall stats
            total = len(db)
            complete = len([r for r in db.all() if r.status == Status.COMPLETE])
            partial = len([r for r in db.all() if r.status == Status.PARTIAL])
            missing = len([r for r in db.all() if r.status == Status.MISSING])

            result: dict[str, Any] = {
                "total": total,
                "complete": complete,
                "partial": partial,
                "missing": missing,
                "completion_pct": round(complete / total * 100, 1) if total > 0 else 0,
            }

            if verbose >= 1:
                # Add category breakdown
                categories: dict[str, dict[str, int]] = {}
                for req in db.all():
                    cat = req.category or "Uncategorized"
                    if cat not in categories:
                        categories[cat] = {"total": 0, "complete": 0}
                    categories[cat]["total"] += 1
                    if req.status == Status.COMPLETE:
                        categories[cat]["complete"] += 1
                result["categories"] = categories

            if verbose >= 2:
                # Add individual requirements
                requirements = []
                for req in db.all():
                    requirements.append(
                        {
                            "id": req.req_id,
                            "category": req.category,
                            "status": req.status.value,
                            "text": req.requirement_text[:100],
                        }
                    )
                result["requirements"] = requirements

            return ToolResult(success=True, data=result)

        except Exception as e:
            return ToolResult(success=False, data=None, error=str(e))

    def get_backlog(
        self,
        phase: int | None = None,
        critical_only: bool = False,
        limit: int = 20,
    ) -> ToolResult:
        """Get prioritized backlog of incomplete requirements.

        Args:
            phase: Filter by phase number
            critical_only: Only show critical priority items
            limit: Maximum number of items to return

        Returns:
            ToolResult with backlog data
        """
        try:
            db = self._get_db()

            # Filter incomplete requirements
            incomplete = [r for r in db.all() if r.status != Status.COMPLETE]

            # Apply filters
            if phase is not None:
                incomplete = [r for r in incomplete if r.phase == phase]

            if critical_only:
                from rtmx.models import Priority

                incomplete = [r for r in incomplete if r.priority == Priority.P0]

            # Sort by priority and blocking count
            def sort_key(req):
                priority_order = {"CRITICAL": 0, "HIGH": 1, "MEDIUM": 2, "LOW": 3}
                return (
                    priority_order.get(req.priority.value, 4),
                    -len(req.blocks),  # More blockers = higher priority
                    req.phase or 999,
                )

            incomplete.sort(key=sort_key)

            # Build result
            backlog = []
            for req in incomplete[:limit]:
                backlog.append(
                    {
                        "id": req.req_id,
                        "text": req.requirement_text[:100],
                        "priority": req.priority.value,
                        "phase": req.phase,
                        "status": req.status.value,
                        "blocks": list(req.blocks),
                        "dependencies": list(req.dependencies),
                    }
                )

            return ToolResult(
                success=True,
                data={
                    "total_incomplete": len(incomplete),
                    "showing": len(backlog),
                    "items": backlog,
                },
            )

        except Exception as e:
            return ToolResult(success=False, data=None, error=str(e))

    def get_requirement(self, req_id: str) -> ToolResult:
        """Get details for a specific requirement.

        Args:
            req_id: Requirement ID (e.g., REQ-SW-001)

        Returns:
            ToolResult with requirement data
        """
        try:
            db = self._get_db()
            req = db.get(req_id)

            return ToolResult(
                success=True,
                data={
                    "id": req.req_id,
                    "category": req.category,
                    "subcategory": req.subcategory,
                    "text": req.requirement_text,
                    "status": req.status.value,
                    "priority": req.priority.value,
                    "phase": req.phase,
                    "target_value": req.target_value,
                    "test_module": req.test_module,
                    "test_function": req.test_function,
                    "validation_method": req.validation_method,
                    "dependencies": list(req.dependencies),
                    "blocks": list(req.blocks),
                    "assignee": req.assignee,
                    "sprint": req.sprint,
                    "notes": req.notes,
                    "requirement_file": req.requirement_file,
                    "external_id": req.external_id,
                },
            )

        except Exception as e:
            return ToolResult(success=False, data=None, error=str(e))

    def update_status(self, req_id: str, status: str) -> ToolResult:
        """Update the status of a requirement.

        Args:
            req_id: Requirement ID
            status: New status (MISSING, PARTIAL, COMPLETE)

        Returns:
            ToolResult indicating success/failure
        """
        try:
            db = self._get_db()

            # Validate status
            try:
                new_status = Status.from_string(status)
            except ValueError:
                return ToolResult(
                    success=False,
                    data=None,
                    error=f"Invalid status: {status}. Must be MISSING, PARTIAL, or COMPLETE",
                )

            # Update requirement
            db.update(req_id, status=new_status)
            db.save()

            # Reload to get fresh data
            self._reload_db()

            return ToolResult(
                success=True,
                data={"id": req_id, "status": new_status.value},
            )

        except Exception as e:
            return ToolResult(success=False, data=None, error=str(e))

    def get_dependencies(self, req_id: str) -> ToolResult:
        """Get dependency information for a requirement.

        Args:
            req_id: Requirement ID

        Returns:
            ToolResult with dependency data
        """
        try:
            db = self._get_db()
            req = db.get(req_id)

            # Get dependency details
            deps = []
            for dep_id in req.dependencies:
                try:
                    dep = db.get(dep_id)
                    deps.append(
                        {
                            "id": dep.req_id,
                            "status": dep.status.value,
                            "text": dep.requirement_text[:50],
                        }
                    )
                except Exception:
                    deps.append({"id": dep_id, "status": "NOT_FOUND", "text": ""})

            # Get blocks details
            blocks = []
            for block_id in req.blocks:
                try:
                    block = db.get(block_id)
                    blocks.append(
                        {
                            "id": block.req_id,
                            "status": block.status.value,
                            "text": block.requirement_text[:50],
                        }
                    )
                except Exception:
                    blocks.append({"id": block_id, "status": "NOT_FOUND", "text": ""})

            return ToolResult(
                success=True,
                data={
                    "id": req_id,
                    "depends_on": deps,
                    "blocks": blocks,
                    "is_blocked": any(d["status"] != "COMPLETE" for d in deps),
                },
            )

        except Exception as e:
            return ToolResult(success=False, data=None, error=str(e))

    def search_requirements(self, query: str, limit: int = 10) -> ToolResult:
        """Search requirements by text.

        Args:
            query: Search query
            limit: Maximum results to return

        Returns:
            ToolResult with matching requirements
        """
        try:
            db = self._get_db()
            query_lower = query.lower()

            matches = []
            for req in db.all():
                # Search in ID, text, category, notes
                searchable = " ".join(
                    [
                        req.req_id,
                        req.requirement_text,
                        req.category,
                        req.subcategory,
                        req.notes,
                    ]
                ).lower()

                if query_lower in searchable:
                    matches.append(
                        {
                            "id": req.req_id,
                            "text": req.requirement_text[:100],
                            "status": req.status.value,
                            "category": req.category,
                        }
                    )

                if len(matches) >= limit:
                    break

            return ToolResult(
                success=True,
                data={"query": query, "count": len(matches), "results": matches},
            )

        except Exception as e:
            return ToolResult(success=False, data=None, error=str(e))

    def get_spec(self, req_id: str) -> ToolResult:
        """Get the full specification file content for a requirement.

        Args:
            req_id: Requirement ID (e.g., REQ-MCP-001)

        Returns:
            ToolResult with specification markdown content
        """
        try:
            db = self._get_db()
            req = db.get(req_id)

            # Check if requirement has a spec file
            if not req.requirement_file:
                return ToolResult(
                    success=False,
                    data=None,
                    error=f"Requirement {req_id} has no specification file defined",
                )

            # Resolve the spec file path relative to the database location
            db_path = Path(self._config.database)
            spec_path = db_path.parent / req.requirement_file

            if not spec_path.exists():
                return ToolResult(
                    success=False,
                    data=None,
                    error=f"Specification file not found: {spec_path}",
                )

            # Read the spec file content
            content = spec_path.read_text(encoding="utf-8")

            return ToolResult(
                success=True,
                data={
                    "id": req_id,
                    "spec_file": str(req.requirement_file),
                    "content": content,
                },
            )

        except Exception as e:
            return ToolResult(success=False, data=None, error=str(e))
