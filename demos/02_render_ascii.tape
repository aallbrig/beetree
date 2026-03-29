Require beetree

Source demos/_settings.tape

Output demos/segments/02_render_ascii.mp4

# ── Feature: ASCII Tree Rendering ──
Hide
Type "clear"
Enter
Show

Sleep 300ms
Type "printf '\\n  \\033[1;36m── ASCII Tree Rendering ──\\033[0m\\n\\n'"
Enter
Sleep 1s

# Render a simple patrol tree
Type "beetree render examples/patrol.beetree.yaml"
Sleep 300ms
Enter
Sleep 3s

# Render the full enemy AI tree
Type "beetree render examples/full-enemy-ai.beetree.yaml"
Sleep 300ms
Enter
Sleep 4s

Sleep 500ms
