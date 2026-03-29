Require beetree

Source demos/_settings.tape

Output demos/segments/07_generate_code.mp4

# ── Feature: Code Generation ──
Hide
Type "clear"
Enter
Show

Sleep 300ms
Type "printf '\\n  \\033[1;36m── Code Generation ──\\033[0m\\n\\n'"
Enter
Sleep 1s

# Dry-run: see what Unity would generate
Type "beetree generate unity examples/patrol.beetree.yaml --dry-run"
Sleep 300ms
Enter
Sleep 3s

# Dry-run: Godot (GDScript)
Type "beetree generate godot examples/patrol.beetree.yaml --dry-run"
Sleep 300ms
Enter
Sleep 3s

# Dry-run: Unreal (C++)
Type "beetree generate unreal examples/patrol.beetree.yaml --dry-run"
Sleep 300ms
Enter
Sleep 3s

Sleep 500ms
