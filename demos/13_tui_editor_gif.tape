Require beetree

Source demos/_settings.tape

Output website/static/img/tui-editor.gif

# ── Feature: TUI Builder ──
Hide
Type "clear"
Enter
Show

Sleep 300ms

# Open the TUI builder
Type "beetree builder examples/full-enemy-ai.beetree.yaml"
Sleep 300ms
Enter
Sleep 2s

# Navigate down through the tree
Type "jjjj"
Sleep 2s

# Open help overlay
Type "?"
Sleep 3s

# Close help
Type "?"
Sleep 1s

# Navigate back up
Type "kk"
Sleep 1s

# Quit the TUI
Type "q"
Sleep 2s

Sleep 500ms
