Require beetree

Source demos/_settings.tape

Output demos/segments/11_version.mp4

# ── Feature: Version & Completions ──
Hide
Type "clear"
Enter
Show

Sleep 300ms
Type "printf '\\n  \\033[1;36m── Version & Shell Completion ──\\033[0m\\n\\n'"
Enter
Sleep 1s

# Version
Type "beetree version"
Sleep 300ms
Enter
Sleep 2s

# Shell completion (show first 20 lines)
Type "beetree completion bash | head -20"
Sleep 300ms
Enter
Sleep 3s

Sleep 500ms
