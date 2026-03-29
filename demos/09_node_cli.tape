Require beetree

Source demos/_settings.tape

Output demos/segments/09_node_cli.mp4

# ── Feature: CLI Node Management ──
Hide
Type "clear"
Enter
# Set up a temp tree to edit
Type "cp examples/patrol.beetree.yaml /tmp/demo_tree.beetree.yaml"
Enter
Show

Sleep 300ms
Type "printf '\\n  \\033[1;36m── CLI Node Management ──\\033[0m\\n\\n'"
Enter
Sleep 1s

# List available node types
Type "beetree node list"
Sleep 300ms
Enter
Sleep 3s

# Add an action node under the root
Type "beetree node add /tmp/demo_tree.beetree.yaml patrol_loop alert --type action --node AlertNearby"
Sleep 300ms
Enter
Sleep 2s

# Render to see the change
Type "beetree render /tmp/demo_tree.beetree.yaml --format sigil"
Sleep 300ms
Enter
Sleep 3s

# Clean up
Hide
Type "rm -f /tmp/demo_tree.beetree.yaml"
Enter
Show

Sleep 500ms
