# RTMX additional commands demo
# Run with: vhs docs/tapes/commands.tape
# Output: docs/assets/rtmx-commands.gif

Output docs/assets/rtmx-commands.gif

Set Shell "bash"
Set FontSize 14
Set Width 1000
Set Height 620
Set Theme "Dracula"
Set Padding 16
Set TypingSpeed 50ms
Set CursorBlink false

# Security audit
Type "rtmx security"
Enter
Sleep 4s

# Release scoping
Type "rtmx release scope v1.0.0"
Enter
Sleep 4s

# Status grouped by version
Type "rtmx status --by-version"
Enter
Sleep 4s

# JSON output for AI agents
Type "rtmx next --one --json"
Enter
Sleep 4s
