# RTMX workflow demo — status (stale) → verify → status → next
# Closed-loop: status warns verify is old; verify refreshes evidence; status improves.
# Record only via: make gifs  (Docker; no host paths)
# Output: docs/assets/rtmx-workflow.gif
# Fixture covers a small requirement set (not the dogfooded 219-req RTM).

Output docs/assets/rtmx-workflow.gif

Set Shell "bash"
Set FontSize 14
Set Width 1000
Set Height 620
Set Theme "Dracula"
Set Padding 16
Set TypingSpeed 50ms
Set CursorBlink false

# Stale dashboard: warns that status has never been verified
Type "rtmx status"
Enter
Sleep 4s

# Closed-loop verification (fixture results — no Go toolchain in VHS image)
Type "rtmx verify --results results.json --update --no-demote"
Enter
Sleep 5s

# Status reflects fresh evidence (warning gone, completion up)
Type "rtmx status"
Enter
Sleep 4s

# Pick the next unblocked requirement
Type "rtmx next --one"
Enter
Sleep 4s
