# muxbee demo - simple navigation tour
# Run with: vhs demo-simple.tape

Output demo.gif

Set Shell "bash"
Set FontSize 16
Set Width 1000
Set Height 700
Set Padding 10

# Start muxbee
Type "muxbee"
Enter
Sleep 3s

# === DASHBOARD ===
Sleep 2s

# === BRIDGES ===
Type "b"
Sleep 2s

# Scroll through bridges
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Down
Sleep 400ms
Sleep 1s

# === DASHBOARD ===
Type "d"
Sleep 2s

# === LOGS ===
Type "l"
Sleep 3s

# === SETTINGS ===
Type "s"
Sleep 2s

# Navigate settings
Down
Sleep 500ms
Down
Sleep 500ms
Up
Sleep 500ms
Up
Sleep 500ms
Sleep 1s

# === DASHBOARD ===
Type "d"
Sleep 2s

# Quit
Type "q"
