# RTMX workflow demo — status → next → verify (matches SITE-018 / LAUNCH-003)
# Record only via: make gifs  (Docker; no host paths)
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

# Hero: completion dashboard
Type "rtmx status"
Enter
Sleep 4s

# Pick the next unblocked requirement
Type "rtmx next --one"
Enter
Sleep 4s

# Cross-check tests against the RTM (fixture results — no Go toolchain in VHS image)
Type "rtmx verify --results results.json"
Enter
Sleep 5s
