# VHS tape — gera GIF demo do pgcraft.
# Install: https://github.com/charmbracelet/vhs
# Run:     vhs vhs/demo.tape
# Output:  vhs/demo.gif

Output vhs/demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 1400
Set Height 800
Set Theme "Dracula"
Set TypingSpeed 50ms

# Connect (assume DATABASE_URL set previamente)
Type "./pgcraft postgres://dmt_user:dmt_dev_password@localhost:5433/dmt_dev"
Enter
Sleep 2s

# Browse schemas → tables
Tab
Sleep 500ms
Down@30ms 10
Sleep 1s

# Filter tables
Type "/"
Sleep 300ms
Type "django"
Sleep 1s
Enter
Sleep 500ms

# Open table details
Down
Down
Sleep 1s

# EXPLAIN
Type "e"
Sleep 2s
Type "x"

# Stats
Type "T"
Sleep 2s
Type "x"

# Activity
Type "A"
Sleep 2s
Type "x"

# SQL editor
Type "s"
Sleep 500ms
Type "SELECT email, is_active FROM \"user\" LIMIT 5;"
Sleep 1s
Ctrl+J
Sleep 2s

# Quit
Type "x"
Type "q"
Sleep 500ms
