# RTMX MCP server setup demo
# Shows one-command setup for Claude Code with successful connection
# Run with: vhs docs/tapes/mcp-setup.tape
# Output: docs/assets/rtmx-mcp-setup.gif

Output docs/assets/rtmx-mcp-setup.gif

Set Shell "bash"
Set FontSize 14
Set Width 1000
Set Height 620
Set Theme "Dracula"
Set Padding 16
Set TypingSpeed 50ms
Set CursorBlink false

# One-command setup for Claude Code
Type "claude mcp add rtmx -- rtmx mcp-server --stdio"
Enter
Sleep 3s

# Verify successful connection (filter to rtmx only)
Type "claude mcp list 2>&1 | grep -v claude.ai"
Enter
Sleep 8s

# List all 10 tools via stdio transport
Type "bash docs/tapes/mcp-tools.sh"
Enter
Sleep 5s
