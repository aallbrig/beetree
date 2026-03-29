Require beetree

Source demos/_settings.tape

Output demos/segments/08_diff.mp4

# ── Feature: Tree Comparison ──
Hide
Type "clear"
Enter
Show

Sleep 300ms
Type "printf '\\n  \\033[1;36m── Tree Comparison (Diff) ──\\033[0m\\n\\n'"
Enter
Sleep 1s

# Diff two different tree specs
Type "beetree diff examples/patrol.beetree.yaml examples/combat.beetree.yaml"
Sleep 300ms
Enter
Sleep 4s

Sleep 500ms
