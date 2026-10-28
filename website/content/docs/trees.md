---
title: "The tree format"
description: "Node types, decorators and the .beetree.yaml file format."
weight: 30
---

A behavior tree decides what an agent does by walking a tree from the root every moment. Each node reports **success**, **failure** or **running**.

## A complete example

```yaml
version: "1.0"
metadata:
  name: guard
  description: Patrol, chase and attack
  tags: [enemy, melee]
tree:
  type: selector
  name: guard_ai
  children:
    - type: sequence
      name: combat
      children:
        - {type: condition, name: can_see_target}
        - {type: action, name: chase_target}
    - {type: action, name: patrol}
```

Read it as: *try to fight; if that fails, patrol.*

## Node types

| Type | Children | Meaning |
|---|---|---|
| `action` | none | Does something (move, attack, wait). May return *running* while it works. |
| `condition` | none | Checks something and returns success or failure immediately. |
| `sequence` | 1 or more | Runs children in order and stops at the first that does not succeed. "Do all of these." |
| `selector` | 1 or more | Runs children in order and stops at the first that does not fail. "Try these until one works." |
| `parallel` | 1 or more | Runs all children each tick. Fails if any fails. |
| `decorator` | exactly 1 | Changes its child's behaviour (see below). |

Every node needs a unique `name`. Leaves may also set `node`, the implementation id your engine provides, plus `description` and `parameters`.

## Decorators

Set `decorator:` on a `decorator` node.

| Decorator | Effect |
|---|---|
| `repeat` | Run the child forever |
| `negate` | Swap success and failure |
| `always_succeed` / `always_fail` | Force the result |
| `until_fail` / `until_succeed` | Keep running until the child gets that result |
| `retry` | Re-run a failing child (`attempts`, default 3) |
| `timeout` | Fail if the child runs longer than `seconds` |
| `cooldown` | After the child finishes, fail for `seconds` before allowing it again |

## Reactive evaluation

BeeTreeCraft's runtime re-evaluates the tree from the root every tick. Put urgent checks (like "am I about to die?") first and they will interrupt whatever lower-priority branch was running.

## Blackboard and custom nodes

Optional top-level `blackboard` (typed shared variables) and `custom_nodes` (your own node types with typed parameters) generate richer engine code. See `examples/full-enemy-ai.beetree.yaml` in the [repository](https://github.com/aallbrig/beetree/tree/main/examples) for a complete example.
