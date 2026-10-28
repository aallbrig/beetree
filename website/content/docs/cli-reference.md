---
title: "CLI reference"
description: "Every beetree command, generated from the binary."
weight: 20
---

{{< generated >}}

## `beetree`

BeeTree is a behavior tree ecosystem that makes it easy to author, share,
and deploy behavior trees to any game engine.

Define behavior trees in an engine-agnostic .beetree.yaml specification,
then generate native code for Unity (C#), Unreal (C++), and Godot (GDScript).

Quick start:

```
  beetree init                    Initialize a new project
  beetree new <name>              Create a behavior tree
  beetree builder [file]          Launch the interactive TUI editor
  beetree validate <file>         Validate a tree spec
  beetree generate unity <file>   Generate engine code (unity/unreal/godot)
```

```
beetree
```

### `beetree builder`

Launch the interactive behavior tree editor.

If a .beetree.yaml file is provided, it will be loaded for editing.
Otherwise, a new empty tree is created.

Key bindings:

```
  ↑/↓        Navigate tree
  ←/→        Collapse/expand
  a           Add child node
  e           Edit selected node properties
  d           Delete selected node
  m           Move node (press m again on target to place, Esc to cancel)
  u           Undo last change
  r           Run interactive simulation (step through tree, choose each result)
  s           Save to file (prompts for path if new)
  q           Quit (confirms if unsaved changes)
```

Simulation mode:

```
  S           Node returns Success
  F           Node returns Failure
  R           Node returns Running
  Esc         Stop simulation
```

```
beetree builder [file]
```

### `beetree diff`

Show structural differences between two .beetree.yaml specs.

```
beetree diff <file1> <file2>
```

### `beetree doctor`

Run diagnostics on the current BeeTree project.

```
beetree doctor
```

### `beetree generate`

Generate native game engine code from a .beetree.yaml specification.

First run generates everything including stubs. Subsequent runs regenerate
tree definitions but skip existing stubs unless --overwrite is passed.

Use --all to generate for all .beetree.yaml files found in the trees/ directory.

```
beetree generate
```

#### `beetree generate godot`

Generate Godot GDScript code from a .beetree.yaml spec

```
beetree generate godot [file] [flags]
```

**Flags**

```
--all             Generate for all .beetree.yaml files in trees/
      --dry-run         Preview what would be generated
      --output string   Output directory for generated code
      --overwrite       Overwrite existing stub files
```

#### `beetree generate unity`

Generate Unity C# code from a .beetree.yaml spec

```
beetree generate unity [file] [flags]
```

**Flags**

```
--all             Generate for all .beetree.yaml files in trees/
      --dry-run         Preview what would be generated
      --output string   Output directory for generated code
      --overwrite       Overwrite existing stub files
```

#### `beetree generate unreal`

Generate Unreal C++ code from a .beetree.yaml spec

```
beetree generate unreal [file] [flags]
```

**Flags**

```
--all             Generate for all .beetree.yaml files in trees/
      --dry-run         Preview what would be generated
      --output string   Output directory for generated code
      --overwrite       Overwrite existing stub files
```

### `beetree init`

Creates a beetree.yaml manifest and project directory structure
in the current directory.

```
beetree init [flags]
```

**Flags**

```
--name string   Project name (defaults to directory name)
```

### `beetree new`

Creates a new .beetree.yaml file in the trees/ directory.

Available templates (--template):

```
  default   Starter tree with selector, sequence, condition, and action
  patrol    NPC patrol loop with waypoints
  combat    Priority-based combat AI with flee/attack/chase/idle
  utility   Utility-selector based decision making
  blank     Minimal empty tree (just a root selector)
```

```
beetree new <name> [flags]
```

**Flags**

```
--template string   tree template (default, patrol, combat, utility, blank) (default "default")
```

### `beetree node`

Manage behavior tree node types

```
beetree node
```

#### `beetree node add`

Add a new child node under the specified parent.

Example:

```
  beetree node add trees/enemy.beetree.yaml combat reload --type action --node Reload
  beetree node add trees/enemy.beetree.yaml root flee_check --type condition --node FleeCheck
```

```
beetree node add <file> <parent-name> <node-name> [flags]
```

**Flags**

```
--decorator string   Decorator type (for decorator nodes)
      --node string        Node class name (e.g., Patrol, HasTarget)
      --type string        Node type (action, condition, sequence, selector, parallel, decorator) (default "action")
```

#### `beetree node edit`

Edit a node's name, type, node class, or decorator in place.

Example:

```
  beetree node edit trees/enemy.beetree.yaml patrol --name guard
  beetree node edit trees/enemy.beetree.yaml attack --type condition --node HasAmmo
  beetree node edit trees/enemy.beetree.yaml flee --decorator negate
```

```
beetree node edit <file> <node-name> [flags]
```

**Flags**

```
--decorator string   New decorator type
      --name string        New node name
      --node string        New node class name
      --type string        New node type
```

#### `beetree node list`

List available node types

```
beetree node list [flags]
```

**Flags**

```
--filter string   Filter by category: core, extension
```

#### `beetree node move`

Detach a node (and its subtree) and reattach under a new parent.

Example:

```
  beetree node move trees/enemy.beetree.yaml patrol --to combat
```

```
beetree node move <file> <node-name> --to <parent-name> [flags]
```

**Flags**

```
--to string   Destination parent node name
```

#### `beetree node remove`

Remove a node and all its children from the tree.

Example:

```
  beetree node remove trees/enemy.beetree.yaml patrol
```

```
beetree node remove <file> <node-name>
```

### `beetree render`

Render a .beetree.yaml specification file as a visual tree diagram.

Supported formats:

```
  ascii   - Indented ASCII tree with [TAG] labels (default)
  sigil   - Unicode sigil tree (→ ? ! ¿ ◇ ⇒)
  compact - Sigil tree with indentation only (ideal for diffing)
  oneline - Single-line S-expression (for commit messages, search)
  mermaid - Mermaid flowchart syntax
  dot     - Graphviz DOT digraph
```

```
beetree render <file> [flags]
```

**Flags**

```
--format string   Output format: ascii, sigil, compact, oneline, mermaid, dot (default "ascii")
```

### `beetree simulate`

Walk the behavior tree and simulate execution, showing node traversal
and resulting statuses. Use --override to force specific node outcomes.

Example:

```
  beetree simulate trees/enemy-ai.beetree.yaml
  beetree simulate trees/enemy-ai.beetree.yaml --override check_health=FAILURE
```

```
beetree simulate <file> [flags]
```

**Flags**

```
--override strings   Force node outcome: name=STATUS (can repeat)
```

### `beetree validate`

Validate a .beetree.yaml specification file

```
beetree validate <file>
```

### `beetree version`

Print the version of beetree

```
beetree version
```
