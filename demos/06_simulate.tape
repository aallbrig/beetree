Require beetree

Source demos/_settings.tape

Output demos/segments/06_simulate.mp4

# ── Feature: Batch Simulation ──
Hide
Type "clear"
Enter
Show

Sleep 300ms
Type "printf '\\n  \\033[1;36m── Batch Simulation ──\\033[0m\\n\\n'"
Enter
Sleep 1s

# Simulate the patrol tree (all nodes succeed by default)
Type "beetree simulate examples/patrol.beetree.yaml"
Sleep 300ms
Enter
Sleep 3s

# Simulate with an override — force a failure
Type "beetree simulate examples/full-enemy-ai.beetree.yaml --override is_dying=FAILURE"
Sleep 300ms
Enter
Sleep 4s

Sleep 500ms
