# muxbee demo
# Run from project root: vhs assets/demo.tape
# Prerequisites: go build -o muxbee . && ./muxbee init

Output assets/demo.gif

Set Shell "bash"
Set FontSize 16
Set Width 1000
Set Height 700
Set Padding 10
Set TypingSpeed 50ms

# Start muxbee
Set TypingSpeed 120ms
Type "./muxbee"
Set TypingSpeed 50ms
Enter
Sleep 3s

# === BRIDGES ===
Type "b"
Sleep 1500ms

# Enable signal (down 1 from whatsapp)
Down
Sleep 400ms
Enter
Sleep 800ms

# Queue bluesky while signal is pulling (down 3 more)
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Enter
Sleep 10s

# === DASHBOARD - see new services ===
Type "d"
Sleep 4s

# === BRIDGES - disable them ===
Type "b"
Sleep 1s

# Back to signal
Up
Sleep 200ms
Up
Sleep 200ms
Up
Sleep 200ms
Enter
Sleep 800ms

# Queue disable bluesky
Down
Sleep 300ms
Down
Sleep 300ms
Down
Sleep 300ms
Enter
Sleep 8s

# === DASHBOARD ===
Type "d"
Sleep 3s

# === BRIDGES - re-enable signal ===
Type "b"
Sleep 1s
Down
Sleep 300ms
Enter
Sleep 10s

# === DASHBOARD ===
Type "d"
Sleep 3s

# === LOGS ===
Type "l"
Sleep 3s

# === SETTINGS ===
Type "c"
Sleep 1500ms
Down
Sleep 400ms
Down
Sleep 400ms
Sleep 1s

# === DASHBOARD - final ===
Type "d"
Sleep 2s
