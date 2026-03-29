Require beetree

Source demos/_settings.tape

Output demos/segments/05_validate.mp4

# ── Feature: Spec Validation ──
Hide
Type "clear"
Enter
Show

Sleep 300ms
Type "printf '\\n  \\033[1;36m── Spec Validation ──\\033[0m\\n\\n'"
Enter
Sleep 1s

# Validate a correct spec
Type "beetree validate examples/patrol.beetree.yaml"
Sleep 300ms
Enter
Sleep 2s

# Validate the full enemy AI
Type "beetree validate examples/full-enemy-ai.beetree.yaml"
Sleep 300ms
Enter
Sleep 2s

Sleep 500ms
