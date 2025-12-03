"""RTMX MCP server implementation.

Exposes RTMX tools via the Model Context Protocol for AI agent integration.
"""

from __future__ import annotations

import json
from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from rtmx.config import RTMXConfig


def create_server(config: RTMXConfig | None = None):
    """Create an MCP server with RTMX tools.

    Args:
        config: RTMX configuration

    Returns:
        MCP server instance

    Raises:
        ImportError: If mcp package is not installed
    """
    try:
        from mcp.server import Server
        from mcp.server.stdio import stdio_server
        from mcp.types import TextContent, Tool
    except ImportError as e:
        raise ImportError(
            "MCP package is required for MCP server. "
            "Install with: pip install rtmx[mcp]"
        ) from e

    from rtmx.adapters.mcp.tools import RTMXTools

    # Create tools instance
    tools = RTMXTools(config)

    # Create MCP server
    server = Server("rtmx")

    @server.list_tools()
    async def list_tools() -> list[Tool]:
        """List available RTMX tools."""
        return [
            Tool(
                name="rtmx_status",
                description="Get RTM completion status. Returns total/complete/partial/missing counts.",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "verbose": {
                            "type": "integer",
                            "description": "Verbosity level: 0=summary, 1=categories, 2=all requirements",
                            "default": 0,
                        },
                    },
                },
            ),
            Tool(
                name="rtmx_backlog",
                description="Get prioritized backlog of incomplete requirements.",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "phase": {
                            "type": "integer",
                            "description": "Filter by phase number",
                        },
                        "critical_only": {
                            "type": "boolean",
                            "description": "Only show critical priority items",
                            "default": False,
                        },
                        "limit": {
                            "type": "integer",
                            "description": "Maximum items to return",
                            "default": 20,
                        },
                    },
                },
            ),
            Tool(
                name="rtmx_get_requirement",
                description="Get details for a specific requirement by ID.",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "req_id": {
                            "type": "string",
                            "description": "Requirement ID (e.g., REQ-SW-001)",
                        },
                    },
                    "required": ["req_id"],
                },
            ),
            Tool(
                name="rtmx_update_status",
                description="Update the status of a requirement.",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "req_id": {
                            "type": "string",
                            "description": "Requirement ID",
                        },
                        "status": {
                            "type": "string",
                            "description": "New status: MISSING, PARTIAL, or COMPLETE",
                            "enum": ["MISSING", "PARTIAL", "COMPLETE"],
                        },
                    },
                    "required": ["req_id", "status"],
                },
            ),
            Tool(
                name="rtmx_deps",
                description="Get dependency information for a requirement.",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "req_id": {
                            "type": "string",
                            "description": "Requirement ID",
                        },
                    },
                    "required": ["req_id"],
                },
            ),
            Tool(
                name="rtmx_search",
                description="Search requirements by text.",
                inputSchema={
                    "type": "object",
                    "properties": {
                        "query": {
                            "type": "string",
                            "description": "Search query",
                        },
                        "limit": {
                            "type": "integer",
                            "description": "Maximum results",
                            "default": 10,
                        },
                    },
                    "required": ["query"],
                },
            ),
        ]

    @server.call_tool()
    async def call_tool(name: str, arguments: dict[str, Any]) -> list[TextContent]:
        """Handle tool invocations."""
        result = None

        if name == "rtmx_status":
            result = tools.get_status(verbose=arguments.get("verbose", 0))
        elif name == "rtmx_backlog":
            result = tools.get_backlog(
                phase=arguments.get("phase"),
                critical_only=arguments.get("critical_only", False),
                limit=arguments.get("limit", 20),
            )
        elif name == "rtmx_get_requirement":
            result = tools.get_requirement(arguments["req_id"])
        elif name == "rtmx_update_status":
            result = tools.update_status(arguments["req_id"], arguments["status"])
        elif name == "rtmx_deps":
            result = tools.get_dependencies(arguments["req_id"])
        elif name == "rtmx_search":
            result = tools.search_requirements(
                arguments["query"],
                limit=arguments.get("limit", 10),
            )
        else:
            return [TextContent(type="text", text=f"Unknown tool: {name}")]

        # Format result
        if result.success:
            return [TextContent(type="text", text=json.dumps(result.data, indent=2))]
        else:
            return [TextContent(type="text", text=f"Error: {result.error}")]

    return server, stdio_server


async def run_server(config: RTMXConfig | None = None) -> None:
    """Run the MCP server.

    Args:
        config: RTMX configuration
    """
    server, stdio = create_server(config)

    async with stdio(server.run):
        pass
