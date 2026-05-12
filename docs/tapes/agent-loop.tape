# RTMX AI agent loop demo
# Shows the commands an AI agent uses: next, context, verify
# Run with: vhs docs/tapes/agent-loop.tape
# Output: docs/assets/rtmx-agent-loop.gif

Output docs/assets/rtmx-agent-loop.gif

Set Shell "bash"
Set FontSize 14
Set Width 1000
Set Height 620
Set Theme "Dracula"
Set Padding 16
Set TypingSpeed 50ms
Set CursorBlink false

# Agent asks: what should I work on?
Type "rtmx next --one --json"
Enter
Sleep 3s

# Agent gets context for the requirement
Type "rtmx context"
Enter
Sleep 4s

# Agent runs verify to check test coverage
Type "rtmx verify"
Enter
Sleep 5s

# Agent checks overall status
Type "rtmx status"
Enter
Sleep 4s
