# BeeTree CLI — Design Specification

A Go-based CLI/TUI for creating, browsing, editing, and sharing engine-agnostic behavior trees, with code generation for Unity, Unreal Engine, and Godot.

---

## Table of Contents

- [1. Vision and Goals](#1-vision-and-goals)
- [2. Architecture Overview](#2-architecture-overview)
- [3. Behavior Tree Specification Format](#3-behavior-tree-specification-format)
- [4. Core Node Type Definitions](#4-core-node-type-definitions)
- [5. Extension Node Library](#5-extension-node-library)
- [6. Custom Node System](#6-custom-node-system)
- [7. CLI Command Design](#7-cli-command-design)
- [8. TUI Interactive Editor](#8-tui-interactive-editor)
- [9. Code Generation Pipeline](#9-code-generation-pipeline)
- [10. Sharing and Registry Platform](#10-sharing-and-registry-platform)
- [11. Go Libraries and Dependencies](#11-go-libraries-and-dependencies)
- [12. Project Structure](#12-project-structure)
- [13. GitHub Pages Documentation](#13-github-pages-documentation)
- [14. Future Considerations](#14-future-considerations)

---

## 1. Vision and Goals

BeeTree is a behavior tree ecosystem that makes it easy to **author**, **share**, and **deploy** behavior trees to any game engine.

### Core Principles

1. **Author Once, Deploy Anywhere** — Define behavior trees in an engine-agnostic specification, then generate native code for Unity (C#), Unreal (C++), and Godot (GDScript/C#).
2. **Start Minimal, Extend When Needed** — Six core node types cover the standard BT vocabulary. Users extend with custom nodes for domain-specific needs.
3. **Data-Driven by Default** — The `.beetree.yaml` specification file is the source of truth. Everything else is generated.
4. **Community Sharing** — Browse, share, and reuse behavior tree specs via a public registry (like npm/OpenAI GPT store), with support for private trees on pro accounts.
5. **Designer-Friendly** — The TUI provides an interactive visual editor. The CLI supports scripting and CI/CD integration.

### Target Users

- **Game AI Designers** — Create and iterate on behavior trees without writing engine code
- **Game Programmers** — Generate boilerplate, focus on leaf node implementation
- **AI Researchers** — Prototype and share BT architectures across engines
- **Indie Developers** — Rapidly scaffold AI systems for any engine

---

## 2. Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        BeeTree CLI                              │
│                                                                 │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌───────────────┐  │
│  │  Parser   │  │  Editor  │  │ Code Gen │  │   Registry    │  │
│  │ (YAML/   │  │  (TUI)   │  │ Pipeline │  │   Client      │  │
│  │  JSON)   │  │          │  │          │  │               │  │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘  └───────┬───────┘  │
│       │              │             │                │           │
│       └──────────────┴─────────────┘                │           │
│                      │                              │           │
│              ┌───────┴───────┐               ┌──────┴──────┐   │
│              │   Core Model  │               │  BeeTree    │   │
│              │  (Tree, Node, │               │  API Client │   │
│              │   Blackboard) │               │             │   │
│              └───────────────┘               └─────────────┘   │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │              Node Registry (Core + Extensions + Custom)   │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
               ┌──────────────┼──────────────┐
               │              │              │
         ┌─────┴─────┐ ┌─────┴─────┐ ┌─────┴─────┐
         │   Unity   │ │  Unreal   │ │   Godot   │
         │  (C#)     │ │  (C++)    │ │ (GDScript)│
         └───────────┘ └───────────┘ └───────────┘
```

### Data Flow

1. **Author** — User creates/edits a `.beetree.yaml` spec via CLI commands or TUI
2. **Validate** — CLI validates the spec against the node schema
3. **Generate** — Code generation pipeline produces engine-native code
4. **Share** — Optionally publish spec to the BeeTree registry
5. **Browse** — Discover and fork community behavior trees

---

## 3. Behavior Tree Specification Format

The specification uses YAML as the primary format (with JSON as an alternative). YAML is chosen for human readability and ease of manual editing.

### File Convention

- Extension: `.beetree.yaml` or `.beetree.json`
- A project may contain multiple spec files
- A `beetree.yaml` in the project root serves as the manifest

### Specification Schema

```yaml
# beetree.yaml — BeeTree Specification File
version: "1.0"
metadata:
  name: "enemy-patrol-and-engage"
  description: "Standard enemy patrol with engagement behavior"
  author: "aallbrig"
  license: "MIT"
  tags: ["patrol", "combat", "fps"]

# Blackboard variable declarations (typed)
blackboard:
  - name: "target"
    type: "Entity"
    default: null
    description: "Current combat target"
  - name: "health"
    type: "float"
    default: 100.0
  - name: "patrol_waypoints"
    type: "Entity[]"
    default: []
  - name: "is_alerted"
    type: "bool"
    default: false

# Custom node definitions (project-local)
custom_nodes:
  - name: "PatrolWaypoints"
    type: "action"
    description: "Move between patrol waypoints"
    parameters:
      - name: "speed"
        type: "float"
        default: 3.0
      - name: "wait_time"
        type: "float"
        default: 2.0
    blackboard_reads: ["patrol_waypoints"]
    blackboard_writes: []

  - name: "IsHealthLow"
    type: "condition"
    description: "Check if health is below threshold"
    parameters:
      - name: "threshold"
        type: "float"
        default: 30.0
    blackboard_reads: ["health"]

# The behavior tree definition
tree:
  type: "selector"
  name: "root"
  children:
    # Priority 1: Flee when health is low
    - type: "sequence"
      name: "flee_behavior"
      children:
        - type: "condition"
          name: "check_health"
          node: "IsHealthLow"
          parameters:
            threshold: 25.0
        - type: "action"
          name: "flee"
          node: "FleeFromTarget"

    # Priority 2: Engage enemy if one is visible
    - type: "sequence"
      name: "engage_behavior"
      children:
        - type: "condition"
          name: "has_target"
          node: "HasTarget"
        - type: "selector"
          name: "attack_options"
          children:
            - type: "sequence"
              name: "ranged_attack"
              children:
                - type: "condition"
                  name: "in_range"
                  node: "IsInRange"
                  parameters:
                    range: 20.0
                - type: "action"
                  name: "shoot"
                  node: "ShootAtTarget"
            - type: "action"
              name: "move_to_target"
              node: "MoveToTarget"

    # Priority 3: Patrol (default behavior)
    - type: "action"
      name: "patrol"
      node: "PatrolWaypoints"
      parameters:
        speed: 2.5
        wait_time: 3.0

# Sub-tree references (reusable)
subtrees:
  - name: "flee_subtree"
    file: "./subtrees/flee.beetree.yaml"
```

### Supported Types for Parameters and Blackboard

| Type | Description |
|------|-------------|
| `bool` | Boolean value |
| `int` | Integer |
| `float` | Floating-point number |
| `string` | Text string |
| `Entity` | Game engine entity/object reference |
| `Vector2` | 2D vector |
| `Vector3` | 3D vector |
| `Entity[]` | Array of entities |
| `string[]` | Array of strings |
| `any` | Untyped (engine-specific) |

---

## 4. Core Node Type Definitions

Six built-in node types form the standard vocabulary.

### Action

```yaml
type: "action"
```

Leaf node that executes a behavior and modifies world state. Returns `Success`, `Failure`, or `Running`.

- Has no children
- References a named action implementation
- May accept typed parameters
- May read/write blackboard variables

### Condition

```yaml
type: "condition"
```

Leaf node that checks world state without modifying it. Returns `Success` or `Failure` (never `Running`).

- Has no children
- References a named condition implementation
- May accept typed parameters
- Should only read blackboard variables (never write)

### Sequence

```yaml
type: "sequence"
```

Composite node implementing AND logic. Executes children left to right.

- Succeeds if **all** children succeed
- Fails on the **first** child failure (short-circuit)
- Returns `Running` if a child returns `Running`
- Resumes from the running child on next tick

### Selector

```yaml
type: "selector"
```

Composite node implementing OR/fallback logic. Tries children left to right.

- Succeeds on the **first** child success (short-circuit)
- Fails if **all** children fail
- Returns `Running` if a child returns `Running`
- Resumes from the running child on next tick

### Parallel

```yaml
type: "parallel"
success_policy: "require_all"   # or "require_one"
failure_policy: "require_one"   # or "require_all"
```

Composite node that executes all children logically simultaneously.

- `require_all` / `require_one` configurable for both success and failure
- Failure takes priority over success
- Default: succeed when all succeed, fail when one fails

### Decorator

```yaml
type: "decorator"
decorator: "repeat"        # Built-in decorator type
parameters:
  count: 3
```

Single-child wrapper that modifies the child's behavior.

Built-in decorator variants:

| Decorator | Behavior |
|-----------|----------|
| `repeat` | Repeat child N times |
| `negate` | Invert child's Success/Failure |
| `always_succeed` | Return Success regardless of child |
| `always_fail` | Return Failure regardless of child |
| `until_fail` | Repeat child until it returns Failure |
| `until_succeed` | Repeat child until it returns Success |
| `timeout` | Fail if child runs longer than duration |
| `cooldown` | Prevent re-execution for a duration after completion |
| `retry` | Retry child N times on failure |

---

## 5. Extension Node Library

Officially supported nodes beyond the core six. These are available in the BeeTree standard library.

### Utility Selector

```yaml
type: "utility_selector"
selection_method: "highest"   # or "weighted_random", "threshold_random"
threshold: 0.5                # for threshold_random
```

Queries each child for a utility score (0.0–1.0) and selects based on the configured method.

### Active Selector

```yaml
type: "active_selector"
```

Like a Selector, but re-evaluates higher-priority children every tick. If a higher-priority child becomes valid, the currently running lower-priority child is interrupted.

### Random Selector

```yaml
type: "random_selector"
```

Randomly shuffles children before evaluation (non-deterministic fallback).

### Random Sequence

```yaml
type: "random_sequence"
```

Randomly shuffles children before sequential execution.

### Subtree Reference

```yaml
type: "subtree"
ref: "flee_subtree"           # Name from subtrees section
# or
file: "./subtrees/flee.beetree.yaml"
```

Embeds another behavior tree as a node, enabling modular composition.

---

## 6. Custom Node System

Users define their own Action and Condition nodes that integrate with the BeeTree specification.

### Definition in Spec

```yaml
custom_nodes:
  - name: "DetectNearbyEnemy"
    type: "condition"
    description: "Scans for enemies within detection radius"
    category: "perception"
    parameters:
      - name: "detection_radius"
        type: "float"
        default: 15.0
      - name: "detection_angle"
        type: "float"
        default: 120.0
    blackboard_reads: []
    blackboard_writes: ["target", "is_alerted"]
```

### Code Generation Behavior

When generating code, custom nodes produce:
1. A **stub class/script** with the proper interface (lifecycle methods)
2. **Parameter declarations** as class properties
3. **Blackboard accessor helpers** for declared reads/writes
4. **TODO comments** marking where the user implements logic

Example generated Unity (C#) stub:

```csharp
// Auto-generated by BeeTree CLI — implement your logic in the marked sections
using BeeTree;

public class DetectNearbyEnemy : BTCondition
{
    [BTParameter] public float detectionRadius = 15.0f;
    [BTParameter] public float detectionAngle = 120.0f;

    public override void OnInitialize(Blackboard blackboard)
    {
        // TODO: Initialize detection state
    }

    public override BTStatus Evaluate(Blackboard blackboard)
    {
        // TODO: Implement enemy detection logic
        // Read: (none declared)
        // Write: blackboard.Set("target", ...), blackboard.Set("is_alerted", ...)
        return BTStatus.Failure;
    }
}
```

---

## 7. CLI Command Design

### Command Hierarchy

```
beetree
├── init              # Initialize a new .beetree.yaml project
├── new               # Create a new behavior tree spec
├── validate          # Validate a .beetree.yaml file
├── render            # Render tree as ASCII art
├── edit              # Open TUI interactive editor
├── generate          # Generate engine code
│   ├── unity         # Generate Unity C# code
│   ├── unreal        # Generate Unreal C++ code
│   └── godot         # Generate Godot GDScript code
├── node              # Manage custom nodes
│   ├── list          # List available nodes (core + extensions + custom)
│   ├── add           # Add a custom node to the spec
│   └── info          # Show details about a node type
├── registry          # Interact with BeeTree registry
│   ├── browse        # Browse public behavior trees
│   ├── search        # Search for behavior trees by tag/name
│   ├── pull          # Download a behavior tree spec
│   ├── push          # Publish a behavior tree spec
│   ├── fork          # Fork a behavior tree spec
│   └── login         # Authenticate with BeeTree registry
├── doctor            # Check environment and dependencies
└── version           # Show version info
```

### Command Details

#### `beetree init`

```bash
beetree init [--name <name>] [--engine <unity|unreal|godot>]
```

Creates a new BeeTree project:
- Generates a `beetree.yaml` manifest
- Creates directory structure (`trees/`, `subtrees/`, `generated/`)
- Optionally sets default target engine

#### `beetree new <name>`

```bash
beetree new patrol-and-engage --template <template-name>
```

Creates a new `.beetree.yaml` file from an optional template (blank, patrol, combat, etc.).

#### `beetree validate [file]`

```bash
beetree validate ./trees/enemy-ai.beetree.yaml
```

Validates the spec against the schema:
- Node type validation (known types, correct children cardinality)
- Parameter type checking
- Blackboard variable reference validation
- Subtree reference resolution
- Circular dependency detection

#### `beetree render [file]`

```bash
beetree render ./trees/enemy-ai.beetree.yaml [--format ascii|dot|mermaid]
```

Renders the tree visually. Output formats:
- `ascii` — Terminal-friendly tree diagram (default)
- `dot` — Graphviz DOT format for advanced visualization
- `mermaid` — Mermaid diagram syntax for documentation

#### `beetree edit [file]`

```bash
beetree edit ./trees/enemy-ai.beetree.yaml
```

Opens the TUI interactive editor (see §8).

#### `beetree generate <engine> [file]`

```bash
beetree generate unity ./trees/enemy-ai.beetree.yaml --output ./Assets/AI/
beetree generate unreal --all --output ./Source/AI/
beetree generate godot --all
```

Flags:
- `--output <dir>` — Output directory for generated code
- `--all` — Generate for all `.beetree.yaml` files in project
- `--dry-run` — Preview what would be generated
- `--overwrite` — Overwrite existing generated files (default: skip existing stubs)

#### `beetree registry browse`

```bash
beetree registry browse [--tag <tag>] [--sort popular|recent]
beetree registry search "patrol combat"
beetree registry pull aallbrig/enemy-patrol-and-engage
beetree registry push ./trees/enemy-ai.beetree.yaml [--public|--private]
```

#### `beetree node list`

```bash
beetree node list [--filter core|extension|custom] [--format table|json]
```

Lists all available node types with descriptions, parameters, and category.

---

## 8. TUI Interactive Editor

The TUI editor provides a visual interface for building and modifying behavior trees.

### Layout

```
┌─ BeeTree Editor ─────────────────────────────────────────────┐
│ ┌─ Tree View ──────────────────┐ ┌─ Properties ────────────┐ │
│ │ ⊟ [SEL] root                 │ │ Name: root              │ │
│ │   ├─ ⊟ [SEQ] flee_behavior   │ │ Type: selector          │ │
│ │   │    ├─ [CND] check_health │ │ Children: 3             │ │
│ │   │    └─ [ACT] flee         │ │                         │ │
│ │   ├─ ⊟ [SEQ] engage_behavior │ │ ─── Description ─────── │ │
│ │   │    ├─ [CND] has_target   │ │ Root selector for       │ │
│ │   │    └─ ⊟ [SEL] attack_opt │ │ enemy AI behavior       │ │
│ │   │         ├─ ⊟ [SEQ] range │ │                         │ │
│ │   │         │    ├─ [CND]    │ │ ─── Blackboard ──────── │ │
│ │   │         │    └─ [ACT]    │ │ target: Entity (null)   │ │
│ │   │         └─ [ACT] move    │ │ health: float (100.0)   │ │
│ │   └─ [ACT] patrol            │ │                         │ │
│ └──────────────────────────────┘ └─────────────────────────┘ │
│ ┌─ Commands ───────────────────────────────────────────────┐  │
│ │ [a]dd child  [d]elete  [m]ove  [e]dit  [s]ave  [q]uit   │  │
│ └──────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
```

### Key Bindings

| Key | Action |
|-----|--------|
| `↑/↓` | Navigate tree |
| `←/→` | Collapse/expand nodes |
| `a` | Add child node (opens type selector) |
| `d` | Delete selected node |
| `m` | Move node (cut/paste) |
| `e` | Edit node properties |
| `p` | Edit parameters |
| `Tab` | Switch focus between panes |
| `s` | Save to file |
| `Ctrl+r` | Render ASCII preview |
| `Ctrl+g` | Generate code |
| `Ctrl+v` | Validate tree |
| `q` | Quit |

### Node Type Selector

When adding a new node, a filterable list appears:

```
┌─ Select Node Type ───────────────┐
│ Filter: seq▏                     │
│                                  │
│ > [CORE] Sequence                │
│   [CORE] Selector                │
│   [EXT]  Random Sequence         │
│   [CUSTOM] MyCustomSequence      │
│                                  │
│ [Enter] Select  [Esc] Cancel     │
└──────────────────────────────────┘
```

---

## 9. Code Generation Pipeline

### Architecture

```
.beetree.yaml  →  Parse  →  Validate  →  Core Model  →  Template Engine  →  Generated Code
                                              │
                                    ┌─────────┼─────────┐
                                    │         │         │
                              Unity C#   Unreal C++ Godot GDScript
```

### Template Engine

Code generation uses Go's `text/template` with engine-specific template sets.

Each target engine has a template directory:

```
templates/
├── unity/
│   ├── tree.cs.tmpl           # BT runtime manager
│   ├── action.cs.tmpl         # Action node stub
│   ├── condition.cs.tmpl      # Condition node stub
│   ├── blackboard.cs.tmpl     # Blackboard class
│   └── tree_definition.cs.tmpl # Tree builder (factory)
├── unreal/
│   ├── tree.h.tmpl
│   ├── tree.cpp.tmpl
│   ├── task.h.tmpl            # BTTaskNode subclass
│   ├── task.cpp.tmpl
│   ├── decorator.h.tmpl
│   └── blackboard.h.tmpl
└── godot/
    ├── tree.gd.tmpl
    ├── action.gd.tmpl
    ├── condition.gd.tmpl
    └── blackboard.gd.tmpl
```

### Generation Strategy

**First run** — Generates everything including stubs. User implements logic in stubs.

**Subsequent runs** — Regenerates only the tree definition/factory code. Never overwrites user-edited stubs unless `--overwrite` is passed.

This is achieved via a **generated file header**:

```csharp
// ============================================================
// AUTO-GENERATED by BeeTree CLI — DO NOT EDIT
// Source: enemy-ai.beetree.yaml
// Regenerated on each `beetree generate` run
// ============================================================
```

vs. stubs that have:

```csharp
// ============================================================
// Generated by BeeTree CLI — EDIT THIS FILE
// Implement your custom logic below
// ============================================================
```

### Per-Engine Output

#### Unity (C#)

- Generates MonoBehaviour-compatible classes
- Uses ScriptableObject for tree definitions
- Integrates with Unity's serialization system
- Action/Condition nodes extend a `BTNode` base class
- Tree runner as a MonoBehaviour component

#### Unreal Engine (C++)

- Generates UBTTaskNode, UBTDecorator, UBTCompositeNode subclasses
- Compatible with Unreal's built-in Behavior Tree system
- Blackboard integration via UBlackboardComponent
- Header/source pairs per node

#### Godot (GDScript)

- Generates Node-based classes
- Tree runner as a GDScript node
- Blackboard as a Dictionary resource
- Compatible with Godot 4.x

---

## 10. Sharing and Registry Platform

### Concept

Like OpenAI's GPT store or npm, users can browse, share, and reuse behavior tree specs.

### Registry Features

| Feature | Free | Pro |
|---------|------|-----|
| Browse public trees | ✓ | ✓ |
| Pull public trees | ✓ | ✓ |
| Push public trees | ✓ | ✓ |
| Private trees | ✗ | ✓ |
| Team collaboration | ✗ | ✓ |
| Version history | 3 versions | Unlimited |
| Analytics | ✗ | ✓ |

### Registry API Endpoints (Future)

```
GET    /api/v1/trees                 # List/search trees
GET    /api/v1/trees/:owner/:name    # Get specific tree
POST   /api/v1/trees                 # Publish a tree
PUT    /api/v1/trees/:owner/:name    # Update a tree
DELETE /api/v1/trees/:owner/:name    # Delete a tree
POST   /api/v1/trees/:owner/:name/fork  # Fork a tree
POST   /api/v1/auth/login            # Authenticate
```

### Authentication

- OAuth via GitHub for CLI authentication
- API tokens for CI/CD integration
- `beetree registry login` triggers browser-based OAuth flow

---

## 11. Go Libraries and Dependencies

### Current Dependencies (Keep)

| Library | Import Path | Purpose |
|---------|------------|---------|
| Cobra | `github.com/spf13/cobra` | CLI framework with subcommands |
| tview | `github.com/rivo/tview` | TUI widget library (tree view, forms) |
| tcell | `github.com/gdamore/tcell/v2` | Terminal handling (tview backend) |
| looplab/fsm | `github.com/looplab/fsm` | Editor state management |

### New Dependencies (Add)

| Library | Import Path | Purpose |
|---------|------------|---------|
| Viper | `github.com/spf13/viper` | Configuration management (integrates with Cobra) |
| yaml.v3 | `gopkg.in/yaml.v3` | YAML parsing for `.beetree.yaml` specs |
| testify | `github.com/stretchr/testify` | Test assertions and mocking |
| lipgloss | `github.com/charmbracelet/lipgloss` | Terminal styling for CLI output |
| glamour | `github.com/charmbracelet/glamour` | Markdown rendering in terminal |
| go-jsonschema | `github.com/invopop/jsonschema` | JSON Schema generation for spec validation |

### Standard Library Usage

| Package | Purpose |
|---------|---------|
| `text/template` | Code generation engine |
| `encoding/json` | JSON serialization (alternative to YAML) |
| `os`, `path/filepath` | File system operations |
| `fmt`, `strings` | String formatting |
| `testing` | Test framework (with testify extensions) |

### Why These Choices

**Cobra + Viper** — The de facto standard for Go CLIs. Cobra powers kubectl, Hugo, and GitHub CLI. Viper integrates seamlessly for configuration.

**tview + tcell** — Already in use. tview's `TreeView` widget is ideal for behavior tree visualization. Rich enough for the full TUI editor.

**text/template** — Standard library is sufficient for code generation. No external dependency needed. Supports custom functions for engine-specific formatting.

**YAML (gopkg.in/yaml.v3)** — Human-readable specification format. Established library with struct tag support.

---

## 12. Project Structure

### Proposed CLI Layout

```
cli/
├── main.go                           # Entry point
├── go.mod
├── go.sum
│
├── cmd/                              # CLI commands (Cobra)
│   ├── root.go                       # Root command + global flags
│   ├── init.go                       # beetree init
│   ├── new.go                        # beetree new
│   ├── validate.go                   # beetree validate
│   ├── render.go                     # beetree render
│   ├── edit.go                       # beetree edit (TUI launcher)
│   ├── generate.go                   # beetree generate (parent)
│   ├── generate_unity.go             # beetree generate unity
│   ├── generate_unreal.go            # beetree generate unreal
│   ├── generate_godot.go             # beetree generate godot
│   ├── node.go                       # beetree node (parent)
│   ├── node_list.go                  # beetree node list
│   ├── node_add.go                   # beetree node add
│   ├── node_info.go                  # beetree node info
│   ├── registry.go                   # beetree registry (parent)
│   ├── registry_browse.go            # beetree registry browse
│   ├── registry_search.go            # beetree registry search
│   ├── registry_pull.go              # beetree registry pull
│   ├── registry_push.go              # beetree registry push
│   ├── doctor.go                     # beetree doctor
│   └── version.go                    # beetree version
│
├── internal/
│   ├── model/                        # Core domain model
│   │   ├── tree.go                   # TreeSpec, NodeSpec structs
│   │   ├── blackboard.go             # Blackboard variable definitions
│   │   ├── node_types.go             # Core node type constants + validation
│   │   └── status.go                 # Success, Failure, Running
│   │
│   ├── parser/                       # Spec file parsing
│   │   ├── yaml.go                   # YAML parser
│   │   ├── json.go                   # JSON parser
│   │   ├── text.go                   # Legacy text format (S(T(...)))
│   │   └── parser_test.go
│   │
│   ├── validator/                    # Spec validation
│   │   ├── validator.go              # Schema + semantic validation
│   │   └── validator_test.go
│   │
│   ├── renderer/                     # Tree rendering
│   │   ├── ascii.go                  # ASCII art output
│   │   ├── dot.go                    # Graphviz DOT output
│   │   ├── mermaid.go                # Mermaid diagram output
│   │   └── tui.go                    # tview tree rendering
│   │
│   ├── codegen/                      # Code generation pipeline
│   │   ├── generator.go              # Common generation interface
│   │   ├── unity.go                  # Unity C# generator
│   │   ├── unreal.go                 # Unreal C++ generator
│   │   ├── godot.go                  # Godot GDScript generator
│   │   ├── templates/                # Embedded template files
│   │   │   ├── unity/
│   │   │   ├── unreal/
│   │   │   └── godot/
│   │   └── codegen_test.go
│   │
│   ├── editor/                       # TUI editor
│   │   ├── editor.go                 # Main editor app
│   │   ├── tree_view.go              # Tree pane
│   │   ├── properties_view.go        # Properties pane
│   │   ├── node_selector.go          # Node type picker modal
│   │   └── keybindings.go            # Key binding configuration
│   │
│   ├── registry/                     # Registry API client
│   │   ├── client.go                 # HTTP client for registry
│   │   ├── auth.go                   # OAuth/token management
│   │   └── client_test.go
│   │
│   └── config/                       # CLI configuration
│       ├── config.go                 # Viper-based config
│       └── defaults.go               # Default settings
│
├── templates/                        # Code generation templates (if not embedded)
│   ├── unity/
│   ├── unreal/
│   └── godot/
│
└── testdata/                         # Test fixtures
    ├── valid/
    │   ├── simple-tree.beetree.yaml
    │   ├── complex-tree.beetree.yaml
    │   └── custom-nodes.beetree.yaml
    └── invalid/
        ├── missing-type.beetree.yaml
        └── circular-ref.beetree.yaml
```

### Monorepo Layout

```
beetree/
├── README.md                  # Project overview
├── RESEARCH.md                # This research document
├── DESIGN.md                  # This design specification
├── cli/                       # Go CLI application
├── website/                   # Hugo documentation site
└── .github/
    └── workflows/
        ├── hugo.yml           # GitHub Pages deployment
        ├── cli-test.yml       # CLI tests
        └── cli-release.yml    # CLI binary releases
```

---

## 13. GitHub Pages Documentation

### Site Structure (Hugo)

The website at `beetreecraft.com` serves as both marketing and documentation.

```
website/
├── hugo.toml
├── content/
│   ├── _index.md                    # Landing page
│   ├── docs/
│   │   ├── _index.md               # Documentation home
│   │   ├── getting-started/
│   │   │   ├── installation.md      # CLI installation guide
│   │   │   ├── quickstart.md        # 5-minute tutorial
│   │   │   └── concepts.md         # BT fundamentals
│   │   ├── specification/
│   │   │   ├── format.md           # .beetree.yaml spec reference
│   │   │   ├── core-nodes.md       # Core node type docs
│   │   │   ├── extension-nodes.md  # Extension library docs
│   │   │   ├── custom-nodes.md     # Custom node guide
│   │   │   └── blackboard.md       # Blackboard reference
│   │   ├── cli/
│   │   │   ├── commands.md          # Full CLI reference
│   │   │   └── configuration.md    # Config file reference
│   │   ├── code-generation/
│   │   │   ├── unity.md            # Unity integration guide
│   │   │   ├── unreal.md           # Unreal integration guide
│   │   │   └── godot.md           # Godot integration guide
│   │   └── registry/
│   │       ├── browsing.md          # How to discover trees
│   │       └── publishing.md       # How to share trees
│   └── blog/
│       └── _index.md               # Blog/changelog
├── layouts/
│   ├── index.html                   # Landing page template
│   └── docs/
│       └── single.html              # Documentation page template
└── static/
    ├── images/
    └── css/
```

### Hugo Theme

Recommend using **Hugo Book** (`github.com/alex-shpak/hugo-book`) or **Docsy** for documentation-focused sites. Both support:
- Sidebar navigation
- Search
- Dark mode
- Mobile responsive
- Versioned documentation

### Deployment

Already configured via `.github/workflows/hugo.yml`. No changes needed — pushes to `main` auto-deploy to GitHub Pages.

---

## 14. Future Considerations

### Phase 1 — Foundation (Current)
- Core specification format (.beetree.yaml)
- Six core node types
- CLI commands: init, new, validate, render, edit
- TUI editor with tree view and properties
- ASCII rendering

### Phase 2 — Code Generation
- Unity C# code generation
- Unreal C++ code generation
- Godot GDScript code generation
- Template engine with per-engine templates
- Generated code testing

### Phase 3 — Ecosystem
- Extension node library (Utility Selector, Active Selector, etc.)
- Custom node system with stub generation
- Subtree composition and references
- Mermaid/DOT rendering for documentation

### Phase 4 — Registry
- BeeTree registry API
- CLI commands: browse, search, pull, push
- GitHub OAuth authentication
- Public/private tree support

### Phase 5 — Advanced Features
- Visual debugging integration (engine plugins)
- Behavior tree simulation/dry-run in CLI
- FSM integration (hybrid BT/FSM à la FFXV)
- AI-assisted tree generation (LLM integration)
- Version diffing for tree specs
- Team collaboration features
