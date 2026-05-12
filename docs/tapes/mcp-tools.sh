#!/bin/bash
# List RTMX MCP tools via stdio transport
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | rtmx mcp-server --stdio 2>/dev/null | jq -r '.result.tools[].name'
