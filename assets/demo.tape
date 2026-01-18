# muxchat demo
# Run with: vhs demo.tape

Output demo.gif

Set Shell "bash"
Set FontSize 16
Set Width 1000
Set Height 700
Set Padding 10
Set TypingSpeed 50ms

# Make sure muxchat is built and config exists
# Prerequisites: go build -o muxchat . && ./muxchat init

# Start muxchat
Type "./muxchat"
Enter
Sleep 2s

# === DASHBOARD ===
# We start on the dashboard - pause to show it
Sleep 3s

# === BRIDGES ===
# Press 'b' to go to bridges
Type "b"
Sleep 2s

# Navigate to signal (index 1) and toggle it
Down
Sleep 300ms
Enter
Sleep 2s

# Navigate to bluesky (index 4) - need to go down 3 more
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Enter
Sleep 2s

# Navigate to meta (index 9) - need to go down 5 more
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Enter
Sleep 2s

# Navigate to linkedin (index 11) - need to go down 2 more
Down
Sleep 300ms
Down
Sleep 300ms
Enter
Sleep 2s

# Pause to show bridges enabled
Sleep 2s

# === BACK TO DASHBOARD ===
Type "d"
Sleep 3s

# === BACK TO BRIDGES ===
Type "b"
Sleep 2s

# Disable the bridges we just enabled (they're still selected positions)
# Go back to signal (index 1)
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Down
Sleep 300ms
Enter
Sleep 1500ms

# === DASHBOARD AGAIN ===
Type "d"
Sleep 2s

# === LOGS ===
Type "l"
Sleep 3s

# === SETTINGS ===
Type "s"
Sleep 3s

# Show we can navigate in settings but don't change mode
Down
Sleep 500ms
Down
Sleep 500ms
Up
Sleep 500ms
Up
Sleep 500ms

# Pause to show settings
Sleep 2s

# === BACK TO DASHBOARD ===
Type "d"
Sleep 2s

# Quit
Type "q"
Sleep 500ms
