# RTMX MCP setup for Cursor (not Claude Code)
# Record only via: make gifs  (Docker; no host paths)
# Output: docs/assets/rtmx-mcp-setup.gif

Output docs/assets/rtmx-mcp-setup.gif

Set Shell "bash"
Set FontSize 14
Set Width 1000
Set Height 620
Set Theme "Dracula"
Set Padding 16
Set TypingSpeed 45ms
Set CursorBlink false

# Show Cursor MCP config (file pre-seeded in the recording container)
Type "cat .cursor/mcp.json"
Enter
Sleep 3s

# List tools over the same stdio transport Cursor uses
Type `list-rtmx-mcp-tools`
Enter
Sleep 5s
