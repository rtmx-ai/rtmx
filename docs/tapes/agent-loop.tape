# RTMX AI agent loop demo
# Shows an agent picking a requirement and completing it
# Run with: vhs docs/tapes/agent-loop.tape
# Output: docs/assets/rtmx-agent-loop.gif

Output docs/assets/rtmx-agent-loop.gif

Set Shell "bash"
Set FontSize 16
Set Width 960
Set Height 540
Set Theme "Dracula"
Set Padding 20
Set TypingSpeed 80ms

# Agent picks next requirement
Type "# Agent asks: what should I work on?"
Enter
Sleep 1s
Type "rtmx next --one --json"
Enter
Sleep 3s

# Agent writes a test (simulated)
Type "# Agent writes code and tests linked to the requirement"
Enter
Sleep 2s

# Agent verifies
Type "rtmx verify"
Enter
Sleep 5s

# Requirement status updates
Type "# Requirement status updated automatically"
Enter
Sleep 1s
Type "rtmx status"
Enter
Sleep 3s
