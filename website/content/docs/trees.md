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

## One tree, several agents

A squad tree drives several agents at once. The tree ticks once per frame, and each leaf names the agent it acts on with `parameters.slot`. The slots are `leader`, `flanker` and `support`. A leaf with no `slot` acts on the leader, so trees written for a single agent keep working. The validator rejects any other slot name.

```yaml
version: "1.0"
metadata:
  name: squad
  description: One tree, three roles
tree:
  type: parallel
  name: squad_ai
  children:
    - {type: action, name: hold_position, parameters: {slot: leader}}
    - {type: action, name: circle_target, parameters: {slot: flanker}}
    - {type: action, name: follow_leader, parameters: {slot: support}}
```

A slot with no agent in the scene makes its leaf fail.

Explore's Demo scene 3 runs this pattern; its tree is [`examples/squad.beetree.yaml`](https://github.com/aallbrig/beetreecraft/blob/main/examples/squad.beetree.yaml).

## Blackboard and custom nodes

Optional top-level `blackboard` (typed shared variables) and `custom_nodes` (your own node types with typed parameters) generate richer engine code. See `examples/full-enemy-ai.beetree.yaml` in the [repository](https://github.com/aallbrig/beetreecraft/tree/main/examples) for a complete example.

## JSON Schema

The format is described by a [JSON Schema](https://beetreecraft.com/schemas/beetree.schema.json) (draft 2020-12), hosted at:

```
https://beetreecraft.com/schemas/beetree.schema.json
```

Point your editor at it to get completion and inline errors while you write a tree. In VS Code with the YAML extension, add this as the first line of your file:

```yaml
# yaml-language-server: $schema=https://beetreecraft.com/schemas/beetree.schema.json
```

The schema checks structure: required fields, the six node types, and child counts (leaves take none, sequence, selector and parallel take at least one, a decorator takes exactly one). Run `beetree validate` for the full set of checks, including duplicate blackboard variables and subtree references.
