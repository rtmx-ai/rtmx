# RTMX 30-second workflow demo
# Records: status -> next -> verify -> status
# Run with: vhs docs/tapes/workflow.tape
# Output: docs/assets/rtmx-workflow.gif

Output docs/assets/rtmx-workflow.gif

Set Shell "bash"
Set FontSize 16
Set Width 960
Set Height 540
Set Theme "Dracula"
Set Padding 20
Set TypingSpeed 80ms

# Show status dashboard
Type "rtmx status"
Enter
Sleep 3s

# Pick the next requirement
Type "rtmx next --one"
Enter
Sleep 3s

# Show what tests to write (verify before implementation)
Type "rtmx verify"
Enter
Sleep 5s

# Show updated status after tests pass
Type "rtmx status"
Enter
Sleep 3s
