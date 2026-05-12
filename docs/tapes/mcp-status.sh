#!/bin/bash
# Call RTMX MCP status tool via stdio transport
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"status"}}' \
  | rtmx mcp-server --stdio 2>/dev/null \
  | jq '.result.content[0].text | fromjson | {total, complete, completion_pct, categories: [.categories[] | {name, pct}]}'
