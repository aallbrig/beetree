Require beetree

Source demos/_settings.tape

Output demos/segments/03_render_sigil.mp4

# ── Feature: Sigil Notation ──
Hide
Type "clear"
Enter
Show

Sleep 300ms
Type "printf '\\n  \\033[1;36m── Sigil Notation (→ ? ! ¿ ◇ ⇒) ──\\033[0m\\n\\n'"
Enter
Sleep 1s

# Sigil tree — academic BT symbols
Type "beetree render examples/full-enemy-ai.beetree.yaml --format sigil"
Sleep 300ms
Enter
Sleep 4s

# Compact format — ideal for git diff
Type "beetree render examples/patrol.beetree.yaml --format compact"
Sleep 300ms
Enter
Sleep 3s

# Single-line S-expression
Type "beetree render examples/full-enemy-ai.beetree.yaml --format oneline"
Sleep 300ms
Enter
Sleep 3s

Sleep 500ms
