# RTMX MCP server setup demo
# Shows one-command setup for Claude Code and Cursor
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

# Set up for Claude Code -- one command
Type "claude mcp add rtmx -- rtmx mcp-server --stdio"
Enter
Sleep 4s

# Show it registered
Type "claude mcp list"
Enter
Sleep 4s

# Quick stdio demo: list available tools
Type "bash docs/tapes/mcp-tools.sh"
Enter
Sleep 5s

# Show the help with Cursor setup example
Type "rtmx mcp-server --help"
Enter
Sleep 5s
