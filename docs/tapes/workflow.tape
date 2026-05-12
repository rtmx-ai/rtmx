# RTMX workflow demo -- dogfooded on itself, 233 requirements
# Run with: vhs docs/tapes/workflow.tape
# Output: docs/assets/rtmx-workflow.gif

Output docs/assets/rtmx-workflow.gif

Set Shell "bash"
Set FontSize 14
Set Width 1000
Set Height 620
Set Theme "Dracula"
Set Padding 16
Set TypingSpeed 50ms
Set CursorBlink false

# Show status -- the hero shot
Type "rtmx status"
Enter
Sleep 5s

# Health check -- all green
Type "rtmx health"
Enter
Sleep 4s

# Dependency graph -- top connected requirements
Type "rtmx deps"
Enter
Sleep 4s

# Verify -- close the loop between tests and requirements
Type "rtmx verify"
Enter
Sleep 5s
