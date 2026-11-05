---
title: "Behavior trees and their history"
description: "What a behavior tree is, why game studios adopted them, and how they spread beyond games."
weight: 10
---

A **behavior tree** (BT) is a way to organize what an autonomous agent does. The agent might be a guard in a stealth game, a companion in an RPG or a warehouse robot. The tree breaks the agent's behavior into small tasks and says how to switch between them. Many times a second the agent walks the tree from the root. Each node reports one of three results:

- **Success**: the node did its job.
- **Failure**: it could not.
- **Running**: it is still working on it.

Parents use those results to decide what to try next. That is the whole idea. The [building blocks](/learn/building-blocks/) page covers the six node types.

## Why not a state machine?

Before behavior trees, most game characters were built as **finite state machines** (FSMs): a set of states such as *Patrol*, *Chase* and *Attack*, with arrows between them for every allowed transition.

FSMs are easy to start with. They get hard to live with as the character grows. For a character to be *reactive* (to drop what it is doing and flee when hurt), many states need arrows to many other states. Each arrow is a one-way jump. If you delete or change a state, you have to revisit every arrow that points at it. Colledanchise and Ögren, in [*Behavior Trees in Robotics and AI*](https://arxiv.org/abs/1709.00084), compare this to the `goto` statement in early programming languages, which Dijkstra argued against in 1968. A behavior tree instead uses two-way hand-offs, like function calls: a parent asks a child to do something and gets an answer back. A tree can therefore be reactive and still be built from parts that you develop, test and reuse separately.

Two properties explain most of the appeal:

- **Modular**: a sub-tree such as *Pick up ball* does not need to know what runs before or after it. You can lift it out and reuse it in another tree.
- **Reactive**: because the tree is re-evaluated from the root, a higher-priority branch (flee when hurt) can interrupt a lower one (patrol) without any special wiring.

## A short history

**Before trees (1980s to 1990s).** Game AI is mostly finite state machines and scripts. In robotics, Rodney Brooks proposes the *subsumption architecture* in 1986: layers of simple behaviors where higher-priority layers override lower ones. It is an ancestor of the "try the most important thing first" idea that behavior trees later formalized.

**Mid-2000s: the game industry.** Behavior trees take their modern form in game studios, which needed their NPCs to be built from reusable parts. The best-known early account is the Halo 2 AI. Damián Isla, who worked on Bungie's *Halo 2* (2004), presented the design at the Game Developers Conference in 2005 in the talk *Handling Complexity in the Halo 2 AI*. Its tree of prioritized behaviors was widely copied. Alex Champandard's AiGameDev site then spread the technique and its vocabulary across the industry through articles and talks.

**2010s: engines make it standard.** Studios shipped their own trees, and engines began to include them. Unreal Engine 4 shipped a built-in behavior tree editor, and Unity and Godot gained community and asset-store implementations. Today it is one of the most common ways to build NPC behavior.

**2010s to now: robotics and research.** The book above describes how the approach was picked up outside games: robotic manipulation at Carnegie Mellon, end-user robot programming, UAV mission control and formal verification of behavior. The authors observe that the "modularity" that appealed to game studios is also the main reason robotics researchers adopted trees.

## Where BeeTreeCraft fits

Every engine's behavior tree has its own node names, editor and runtime. BeeTreeCraft sits above those differences. You describe a tree in one engine-agnostic file, simulate it, and generate code for [Unity, Unreal or Godot](/docs/cli/).

## Further reading

- Colledanchise and Ögren, [*Behavior Trees in Robotics and AI: An Introduction*](https://arxiv.org/abs/1709.00084) (CRC Press, 2018). Chapter 1 covers the history above, and Chapter 2 shows how trees generalize state machines and subsumption.
- Isla, *Handling Complexity in the Halo 2 AI*, GDC 2005.
- Next: the [building blocks](/learn/building-blocks/).
