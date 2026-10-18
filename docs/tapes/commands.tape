# RTMX additional commands demo
# Record only via: make gifs  (Docker; no host paths)
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

Type "rtmx health"
Enter
Sleep 4s

Type "rtmx backlog"
Enter
Sleep 4s

Type "rtmx next --one --json"
Enter
Sleep 4s
