Require beetree

Source demos/_settings.tape

Output demos/segments/04_render_diagrams.mp4

# ── Feature: Diagram Export ──
Hide
Type "clear"
Enter
Show

Sleep 300ms
Type "printf '\\n  \\033[1;36m── Diagram Export ──\\033[0m\\n\\n'"
Enter
Sleep 1s

# Mermaid flowchart (paste into GitHub markdown)
Type "beetree render examples/patrol.beetree.yaml --format mermaid"
Sleep 300ms
Enter
Sleep 3s

# Graphviz DOT (pipe to dot -Tpng)
Type "beetree render examples/patrol.beetree.yaml --format dot"
Sleep 300ms
Enter
Sleep 3s

Sleep 500ms
