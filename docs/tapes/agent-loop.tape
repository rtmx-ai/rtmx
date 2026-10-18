# RTMX agent loop — Cursor-shaped: next → context → verify → status
# Record only via: make gifs  (Docker; no host paths)
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

Type "rtmx next --one --json"
Enter
Sleep 3s

Type "rtmx context"
Enter
Sleep 4s

Type "rtmx verify --results results.json"
Enter
Sleep 5s

Type "rtmx status"
Enter
Sleep 4s
