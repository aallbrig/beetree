# Behavior Tree Shorthand Notation Research

**Date:** 2026-03-28 16:00:09 UTC
**Author:** BeeTree Team
**Status:** Research complete — ready for implementation

## Problem Statement

BeeTree's TUI currently displays node names without type indicators in the tree view. The user must select a node and read the Properties panel to determine whether it's a sequence, selector, action, condition, etc. This makes it impossible to understand tree structure at a glance.

**Goals:**
1. Make node types legible at a glance in the TUI tree view
2. Support a condensed multi-line tree format (for diffing behavior tree changes)
3. Support a single-line tree description (for commit messages, search, quick reference)
4. Allow users to define custom notation for their own node types via YAML config

## Survey of Existing Notations

### 1. Academic Mathematical Notation (Colledanchise & Ögren, 2018)

The most cited formal notation comes from *Behavior Trees in Robotics and AI* by Michele Colledanchise and Petter Ögren. It uses mathematical symbols in visual diagrams:

| Node Type | Symbol | Notes |
|-----------|--------|-------|
| Sequence | → | Right arrow; children execute left-to-right, all must succeed |
| Fallback/Selector | ? | Question mark; tries alternatives until one succeeds |
| Parallel | ⇉ | Double arrow; ticks children concurrently |
| Action | Named box | Rectangle with action name |
| Condition | Named oval | Ellipse with condition name |
| Decorator | ◇ (diamond) | Rhombus wrapping a single child |

**Source:** Colledanchise & Ögren, *Behavior Trees in Robotics and AI*, 2018. Also confirmed by the Grokipedia article on behavior trees: "Sequence nodes, often denoted by the symbol →" and "Selector nodes, also known as fallback nodes and often denoted by the symbol ?"

**Assessment:** This is the closest thing to a standard. The → and ? symbols are widely recognized in academic BT literature. However, this notation was designed for visual diagrams, not text files. It doesn't specify symbols for actions vs conditions (both are just named shapes).

### 2. BehaviorTree.CPP / BT.CPP (XML Format)

BT.CPP uses XML as its serialization format:

```xml
<Sequence name="root_sequence">
    <SaySomething message="Hello"/>
    <OpenGripper name="open_gripper"/>
    <ApproachObject name="approach_object"/>
</Sequence>
```

It also offers "compact" vs "explicit" syntax — compact uses the node class as the tag name, explicit uses `<Action ID="SaySomething"/>`.

**Assessment:** XML is verbose and not suitable for at-a-glance reading. Node types (Sequence, Action, Condition) are visible as tag names, but the format is designed for tooling (Groot editor), not human scanning.

### 3. py_trees (Python, ROS)

py_trees provides `ascii_tree()` and `unicode_tree()` for terminal display:

```
[-] Sequence
    --> Action 1 [✓]
    --> Action 2 [*]
    --> Action 3 [-]
```

Uses `[-]`, `[o]`, `[*]`, `[✓]`, `[✕]` to show runtime status. Node types are shown via class name (Sequence, Selector, etc.) but there's no sigil-based shorthand for type.

**Assessment:** py_trees focuses on runtime status visualization, not type notation. No sigil system exists.

### 4. BeeTree's Current ASCII Renderer

BeeTree already has abbreviated type tags in its ASCII render format:

```
[SEL] enemy_ai
├── [SEQ] flee
│   ├── [CND] is_dying
│   ├── [ACT] play_hurt
│   └── [ACT] run_to_cover
├── [SEQ] combat
│   └── [PAR] engage
│       ├── [ACT] face_target
│       └── [SEL] attack_or_close
└── [DEC:repeat] patrol_forever
```

Current abbreviation map:

| Type | Abbreviation |
|------|-------------|
| action | `[ACT]` |
| condition | `[CND]` |
| sequence | `[SEQ]` |
| selector | `[SEL]` |
| parallel | `[PAR]` |
| decorator | `[DEC]` |
| utility_selector | `[UTL]` |
| active_selector | `[ASL]` |
| random_selector | `[RSL]` |
| random_sequence | `[RSQ]` |
| subtree | `[SUB]` |

**Assessment:** This is functional for the `render --format ascii` command but is NOT used in the TUI tree view (the user's primary complaint). The `[TAG]` format is 5 characters wide, which adds bulk. It works well for multi-line tree dumps but is too verbose for a single-line format.

### 5. Industry/Game Dev Practice

No standardized text notation exists in game development. Most studios use visual editors (Unreal's BT editor, Unity's node-graph tools, BT.CPP's Groot). When trees are serialized to text, it's usually JSON or XML — not designed for human reading.

## Conclusion: No Standard Short-Form Text Notation Exists

There is no standardized short-form text notation for behavior trees. The academic convention (→ and ?) is the closest to a standard but only covers two node types and was designed for visual diagrams. Every existing tool either uses verbose markup (XML/JSON) or full-word class names in tree displays.

**This is an opportunity.** BeeTree can define a practical sigil-based notation that:
- Builds on the academic convention (→ for sequence, ? for selector)
- Extends it to cover ALL node types
- Works in monospace terminals and plain-text files
- Supports user-defined custom sigils

---

## Proposed Notation: BeeTree Sigil System (BSS)

### Design Principles

1. **One symbol per type** — recognize the node type from a single character/glyph
2. **Academic compatibility** — use → and ? where convention exists
3. **Visual distinctness** — composites, leaves, and decorators should look different from each other
4. **Two tiers** — Unicode-preferred with ASCII fallback for legacy terminals
5. **Extensible** — users can define sigils for custom node types

### Core Sigil Table

| Node Type | Unicode Sigil | ASCII Fallback | Mnemonic |
|-----------|:---:|:---:|----------|
| **Composites** | | | |
| sequence | `→` (U+2192) | `->` | "then do" — academic standard |
| selector | `?` (U+003F) | `?` | "try this?" — academic standard |
| parallel | `⇒` (U+21D2) | `=>` | "simultaneously" — double arrow |
| **Leaves** | | | |
| action | `!` (U+0021) | `!` | "do this!" — imperative |
| condition | `¿` (U+00BF) | `??` | "is this true?" — query |
| **Decorator** | | | |
| decorator | `◇` (U+25C7) | `<>` | wraps/modifies child — diamond shape |
| **Extensions** | | | |
| utility_selector | `⚖` (U+2696) | `?$` | weighing options — scales |
| active_selector | `⚡` (U+26A1) | `?!` | reactive check — lightning |
| random_selector | `🎲` (U+1F3B2) | `?~` | random choice — die |
| random_sequence | `🔀` (U+1F500) | `->~` | random order — shuffle |
| subtree | `↗` (U+2197) | `>>` | "go to" — jump arrow |

#### Decorator Sub-Sigils

When a decorator's specific behavior matters, append a modifier to `◇`:

| Decorator | Full Sigil | ASCII | Mnemonic |
|-----------|:---:|:---:|----------|
| repeat | `◇∞` | `<>R` | loops forever |
| negate | `◇¬` | `<>!` | inverts result |
| always_succeed | `◇✓` | `<>S` | forces success |
| always_fail | `◇✗` | `<>F` | forces failure |
| until_fail | `◇⤓` | `<>UF` | runs until failure |
| until_succeed | `◇⤒` | `<>US` | runs until success |
| timeout | `◇⏱` | `<>T` | time limit |
| cooldown | `◇⏳` | `<>C` | delay between runs |
| retry | `◇↻` | `<>r` | try again |

### Format 1: Sigil Tree (Multi-Line, for TUI and Diffing)

This is the primary format for the TUI tree view and `beetree render --format sigil`:

```
? enemy_ai
├── → flee
│   ├── ¿ is_dying
│   ├── ! play_hurt
│   ├── ! find_cover
│   └── ! run_to_cover
├── → combat
│   ├── ¿ has_target
│   ├── ¿ can_see_target
│   ├── ! go_full_alert
│   └── ⇒ engage
│       ├── ! face_target
│       └── ? attack_or_close
│           ├── → ranged_attack
│           │   ├── ¿ in_range
│           │   └── ! shoot
│           └── ! chase
├── → investigate
│   ├── ¿ is_alerted
│   ├── ! look_around
│   ├── ! go_to_noise
│   ├── ! search_area
│   └── ! calm_down
└── ◇∞ patrol_forever
    └── → patrol_cycle
        ├── ¿ has_waypoints
        ├── ! walk_to_waypoint
        ├── ! wait_briefly
        └── ! advance_waypoint
```

**Benefits over current `[TAG]` format:**
- 1-2 characters vs 5 characters per tag → significantly more compact
- Visual shape differences between `→ ? ! ¿ ◇ ⇒` are immediately apparent
- Composites have "directional" feel (`→ ? ⇒`), leaves have "punctuation" feel (`! ¿`)
- Diffs are cleaner because lines are shorter and sigils don't change when node names change

**Comparison:**

```
# Current [TAG] format (68 chars widest line)
[SEL] enemy_ai
├── [SEQ] combat
│   ├── [CND] has_target
│   └── [PAR] engage
│       ├── [ACT] face_target
│       └── [ACT] shoot

# Proposed sigil format (50 chars widest line)
? enemy_ai
├── → combat
│   ├── ¿ has_target
│   └── ⇒ engage
│       ├── ! face_target
│       └── ! shoot
```

### Format 2: Compact Tree (Indentation-Only, for Diffing)

Strips box-drawing characters for minimal, diff-friendly output. Uses `beetree render --format compact`:

```
? enemy_ai
  → flee
    ¿ is_dying
    ! play_hurt
    ! find_cover
    ! run_to_cover
  → combat
    ¿ has_target
    ¿ can_see_target
    ! go_full_alert
    ⇒ engage
      ! face_target
      ? attack_or_close
        → ranged_attack
          ¿ in_range
          ! shoot
        ! chase
  → investigate
    ¿ is_alerted
    ! look_around
    ! go_to_noise
    ! search_area
    ! calm_down
  ◇∞ patrol_forever
    → patrol_cycle
      ¿ has_waypoints
      ! walk_to_waypoint
      ! wait_briefly
      ! advance_waypoint
```

This format is ideal for `git diff`:
```diff
  ? enemy_ai
    → flee
      ¿ is_dying
      ! play_hurt
-     ! find_cover
-     ! run_to_cover
+     ! dodge_roll
+     ! sprint_to_cover
```

### Format 3: Single-Line (S-expression Style)

For commit messages, search indices, and quick tree summaries. Uses parenthetical grouping:

```
?(→(¿is_dying !play_hurt !find_cover !run_to_cover) →(¿has_target ¿can_see ⇒(!face ?(!shoot !chase))) →(¿alerted !look !go_to_noise !search !calm) ◇∞(→(¿has_wp !walk !wait !next)))
```

**Rules:**
- Composite nodes: `sigil name(child child child)`
- Leaf nodes: `sigil name` (no parens)
- Decorator nodes: `sigil name(child)`
- Spaces separate siblings within parens
- Names can be omitted for unnamed composites: `→(¿check !act)`

**Shortened variant** (omit composite names, keep only leaves):
```
?(→(¿is_dying !play_hurt !find_cover !run_to_cover) →(¿has_target ¿can_see ⇒(!face ?(!shoot !chase))))
```

This is ~120 chars for the enemy AI combat subtree — fits in a wide terminal line.

---

## Custom Node Notation

### User-Defined Sigils via YAML

Users can define custom sigils in their `.beetree.yaml` files under a new `notation` key:

```yaml
version: "1.0"
metadata:
  name: my_game_ai

notation:
  # Override default sigils for built-in types (optional)
  type_sigils:
    sequence: "»"    # prefer guillemet over arrow
    parallel: "‖"    # pipe-style parallel

  # Define sigils for custom node classes
  node_sigils:
    PlayAnimation: "▶"
    SetAlertLevel: "⚠"
    SpawnParticle: "✦"
    PlaySound: "♪"

custom_nodes:
  - name: PlayAnimation
    type: action
    description: "Play a named animation clip"
    parameters:
      - name: clip_name
        type: string
      - name: speed
        type: float
        default: 1.0

tree:
  # ...
```

The sigil resolution order:
1. `notation.node_sigils[node.Node]` — per-node-class custom sigil
2. `notation.type_sigils[node.Type]` — per-type custom sigil
3. Built-in sigil table — default sigils from BSS

### Software Model Changes Required

**1. Add `Sigil` field to `CustomNodeDef`:**

```go
type CustomNodeDef struct {
    Name             string         `yaml:"name" json:"name"`
    Type             string         `yaml:"type" json:"type"`
    Sigil            string         `yaml:"sigil,omitempty" json:"sigil,omitempty"` // NEW
    Description      string         `yaml:"description,omitempty" json:"description,omitempty"`
    // ... existing fields ...
}
```

**2. Add `NotationConfig` to `TreeSpec`:**

```go
type NotationConfig struct {
    TypeSigils map[string]string `yaml:"type_sigils,omitempty" json:"type_sigils,omitempty"`
    NodeSigils map[string]string `yaml:"node_sigils,omitempty" json:"node_sigils,omitempty"`
}

type TreeSpec struct {
    // ... existing fields ...
    Notation    NotationConfig  `yaml:"notation,omitempty" json:"notation,omitempty"` // NEW
}
```

**3. Add sigil resolver function:**

```go
func ResolveSigil(node *NodeSpec, notation NotationConfig) string {
    // 1. Check node-class sigil
    if sigil, ok := notation.NodeSigils[node.Node]; ok {
        return sigil
    }
    // 2. Check type sigil override
    if sigil, ok := notation.TypeSigils[node.Type]; ok {
        return sigil
    }
    // 3. Fall back to built-in BSS table
    return defaultSigils[node.Type]
}
```

---

## Impact on Existing BeeTree Components

| Component | Change | Effort |
|-----------|--------|--------|
| `model/model.go` | Add `NotationConfig`, `Sigil` field | Small |
| `renderer/spec_renderer.go` | Add `RenderSigil`, `RenderCompact`, `RenderOneline` formats; add sigil resolution | Medium |
| `tui/editor_view.go` | Use sigil in tree node labels instead of bare names | Small |
| `cmd/render.go` | Add `sigil`, `compact`, `oneline` format options | Small |
| `validator/` | Validate custom sigils (non-empty, reasonable length) | Small |
| `spec/` | Parse `notation` block from YAML | Automatic (struct tag) |

---

## Alternatives Considered

### Alternative A: Emoji-Only Sigils
Using emoji for all types (`🔀 🎯 ❓ ⚡ 🎲`). Rejected because:
- Emoji are double-width in most terminals, breaking alignment
- Not all terminals render emoji reliably
- Harder to type in editors

### Alternative B: Bracket Tags Only (Current System)
Keep `[SEQ]`, `[ACT]`, etc. This works but:
- 5 chars wide per tag is bulky
- All tags look similar (bracket-text-bracket) — no visual shape difference
- Not suitable for single-line format

### Alternative C: No Notation — Rely on Color
Use ANSI colors to distinguish types in the TUI. Rejected as sole solution because:
- Colors don't work in plain text (diff output, log files, piped output)
- Colorblind users can't distinguish
- But: color + sigils together would be excellent (future enhancement)

---

## Recommendation

**We know enough to implement.** The proposed design is concrete, the code touch-points are well-understood, and the notation builds on established academic convention.

### Suggested Implementation Order

1. **Add sigil table and resolver to `renderer/`** — define the built-in sigil map, add `ResolveSigil()` function
2. **Add `RenderSigil()` format** — multi-line sigil tree (replaces `[TAG]` format, or offered alongside)
3. **Integrate sigils into TUI tree view** — show `→ patrol_cycle` instead of bare `patrol_cycle`
4. **Add `RenderCompact()` format** — indent-only, no box-drawing
5. **Add `RenderOneline()` format** — S-expression single-line
6. **Add `NotationConfig` to model** — `notation` YAML block with `type_sigils` and `node_sigils`
7. **Add `Sigil` field to `CustomNodeDef`** — per-node-class override
8. **Wire custom sigil resolution into renderer** — notation config flows through to sigil resolver
9. **Add `--format sigil|compact|oneline` to render command** — expose new formats

### Open Questions (Non-Blocking)

These can be decided during implementation:

- **Should `sigil` become the default TUI display?** Probably yes, with the `[TAG]` format available via `render --format ascii` for backward compat.
- **Should the ASCII fallback tier be auto-detected?** Could check `$LANG` or terminal capabilities, or just offer a `--ascii` flag.
- **Should single-line format show composite names?** The shortened variant (omit composite names) is more compact but less informative. Could offer both via a `--verbose` flag on `oneline`.

---

## References

1. Colledanchise, M. & Ögren, P. (2018). *Behavior Trees in Robotics and AI: An Introduction*. CRC Press.
2. Grokipedia. "Behavior tree (artificial intelligence, robotics and control)." https://grokipedia.com/page/Behavior_tree_(artificial_intelligence%2C_robotics_and_control)
3. BehaviorTree.CPP Documentation. "XML Format." https://www.behaviortree.dev/docs/learn-the-basics/xml_format
4. BehaviorTree.CPP Documentation. "BT Basics." https://www.behaviortree.dev/docs/learn-the-basics/BT_basics
5. py_trees Documentation. "Visualisation." https://py-trees.readthedocs.io/en/devel/visualisation.html
6. BeeTree internal: `cli/internal/renderer/spec_renderer.go` — current `typeAbbreviations` map
7. BeeTree internal: `cli/internal/model/node_types.go` — node type taxonomy
