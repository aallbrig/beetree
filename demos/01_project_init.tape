Require beetree

Source demos/_settings.tape

Output demos/segments/01_project_init.mp4

# ── Feature: Project Scaffolding ──
Hide
Type "clear"
Enter
Show

Sleep 300ms
Type "printf '\\n  \\033[1;36m── Project Scaffolding ──\\033[0m\\n\\n'"
Enter
Sleep 1s

# Initialize a new project
Type "beetree init --name MyGame"
Sleep 300ms
Enter
Sleep 2s

# Create a new behavior tree
Type "beetree new patrol_ai"
Sleep 300ms
Enter
Sleep 2s

# Show what was created
Type "cat trees/patrol_ai.beetree.yaml"
Sleep 300ms
Enter
Sleep 4s

# Clean up for next segment
Hide
Type "rm -rf beetree.yaml trees/ subtrees/"
Enter
Show

Sleep 500ms
