Require beetree

Source demos/_settings.tape

Output demos/segments/10_tui_editor.mp4

# ── Feature: TUI Builder ──
Hide
Type "clear"
Enter
Show

Sleep 300ms
Type "printf '\\n  \\033[1;36m── TUI Builder ──\\033[0m\\n\\n'"
Enter
Sleep 1s

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
