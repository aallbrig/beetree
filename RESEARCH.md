# BeeTree Research Notes

Research notes compiled from Game AI Pro literature on behavior tree design, implementation patterns, and best practices — written with the BeeTree platform in mind.

---

## Table of Contents

- [1. Core Behavior Tree Node Types](#1-core-behavior-tree-node-types)
- [2. The Tick/Return Status Model](#2-the-tickreturn-status-model)
- [3. Tree Traversal and Execution](#3-tree-traversal-and-execution)
- [4. Decorators and Specialized Composites](#4-decorators-and-specialized-composites)
- [5. The Blackboard Pattern](#5-the-blackboard-pattern)
- [6. Reactive and Event-Driven Patterns](#6-reactive-and-event-driven-patterns)
- [7. Utility-Based Decision Making](#7-utility-based-decision-making)
- [8. Data-Driven and Template-Based Design](#8-data-driven-and-template-based-design)
- [9. Combining Behavior Trees with State Machines](#9-combining-behavior-trees-with-state-machines)
- [10. Simulation, Testing, and Debugging](#10-simulation-testing-and-debugging)
- [11. Common Pitfalls and How to Avoid Them](#11-common-pitfalls-and-how-to-avoid-them)
- [12. Behavior Trees vs. Alternatives](#12-behavior-trees-vs-alternatives)
- [13. Code Generation and Scripting Patterns](#13-code-generation-and-scripting-patterns)
- [14. Implications for BeeTree Design](#14-implications-for-beetree-design)
- [Resources](#resources)

---

## 1. Core Behavior Tree Node Types

The literature converges on **six fundamental node types** that form the standard behavior tree vocabulary. BeeTree should treat these as first-class, built-in primitives.

### Leaf Nodes

**Action** — Modifies the world state. Examples: "move to cover," "fire weapon," "open door." Returns `Success` when the goal is achieved, `Failure` when blocked, or `Running` while still processing.

**Condition** — Checks world state without modifying it. Examples: "is enemy in range?", "is health low?" Operates in two modes: *instant check* (one-time evaluation) and *monitoring mode* (continuous re-evaluation each tick).

### Composite Nodes

**Sequence** (AND logic) — Executes children in order. Fails on the first child failure, succeeds only if all children succeed. Critical implementation detail: the next child must be processed immediately after the previous one succeeds (within the same frame) to avoid missing an entire frame before reaching a low-level action.

> "There's one important thing to note about this implementation; the next child behavior is processed immediately after the previous one succeeds. This is critical to make sure the BT does not miss an entire frame before having found a low-level action to run."
> — *The Behavior Tree Starter Kit*

**Selector** (OR / fallback logic) — Executes children in order until one succeeds or returns Running. Continues searching for fallback behaviors within the same update call, allowing the entire BT to handle failures within a single frame.

> "The selector keeps searching for fallback behaviors in the same update() until a suitable behavior is found or the selector fails. This allows the whole BT to deal with failures within a single frame without pausing."
> — *The Behavior Tree Starter Kit*

**Parallel** — Executes all children logically simultaneously. Two configurable policies for both success and failure:
- `RequireOne`: Terminate when *any* child meets the condition
- `RequireAll`: Terminate when *all* children meet the condition

Failure takes priority over success — the tree should assume worst case. This node must be "extremely precisely specified, so that it can be understood intuitively and relied upon without trying to second-guess the implementation."

### Wrapper Nodes

**Decorator** — Single-child wrapper that modifies the child's behavior. Common patterns include `Repeat(n)`, `Negate` (invert status), `Forever` (continue despite failure), `UntilFail`, and `Timeout`. Decorators introduce behavioral subtlety without tree duplication.

### BeeTree Takeaway

The standard set for BeeTree's core specification should be: **Action, Condition, Sequence, Selector, Parallel, Decorator**. These six nodes are Turing-complete with external memory (see Pitfalls section). Everything else is a specialization.

---

## 2. The Tick/Return Status Model

Every node returns one of three status codes:

| Status | Meaning |
|--------|---------|
| **Success** | Behavior completed as intended |
| **Failure** | Behavior encountered an obstacle or completed unsuccessfully |
| **Running** | Behavior still processing; requires another tick |

A fourth status, **Suspended**, appears in event-driven implementations where a behavior is waiting for an external event and should be ignored by the scheduler.

### Node Lifecycle API

The literature establishes a critical three-method contract:

```
onInitialize()  → Called once immediately before first update()
update()        → Called exactly once per tick until termination
onTerminate()   → Called once immediately after termination (receives final status)
```

The canonical tick wrapper:

```
tick():
  if status != Running then onInitialize()
  status = update()
  if status != Running then onTerminate(status)
  return status
```

> "This API is the core of any BT, and it's critical that you establish a clear specification for these operations."
> — *The Behavior Tree Starter Kit*

### BeeTree Takeaway

BeeTree's specification format must capture the three return statuses. The lifecycle contract (`onInitialize`, `update`, `onTerminate`) should be part of the code generation templates — each target engine (Unity, Unreal, Godot) will implement this contract in its own idiom.

---

## 3. Tree Traversal and Execution

### First-Generation: Full Tree Traversal

Depth-first traversal from the root every update. Simple to implement and understand. Many production games update behavior trees every other frame or at 5Hz to spread load.

> "It is often not necessary to update the behavior tree every game frame, with many games deciding to update each behavior tree every other frame or at 5Hz so that the load for updating the behavior trees of all characters can be spread across multiple game frames."
> — *The Behavior Tree Starter Kit*

### Second-Generation: Event-Driven

Maintains a scheduler with active behaviors in a queue. Behaviors remain active until they terminate. Parent composites use observers to manage children. Only replaces the active behavior list when a behavior terminates or world state changes. This approach enables partial tree evaluation for efficiency.

### Partial Evaluation (FPS Architecture)

The reactive FPS architecture uses partial tree evaluation — only active parallel nodes and the current action are maintained per agent, avoiding full reevaluation every frame.

> "This approach helps to create an extremely efficient BT, as we do not evaluate the complete tree every frame. It also keeps the reactiveness of the tree intact."
> — *A Reactive AI Architecture for Networked FPS Games*

### BeeTree Takeaway

BeeTree's specification should be execution-model agnostic. The tree definition captures structure and behavior; the generated code can implement either traversal strategy based on the target engine's needs. The spec should annotate whether a tree expects full or partial evaluation.

---

## 4. Decorators and Specialized Composites

Beyond the six core nodes, the literature describes several useful specializations:

### Filter
A Sequence with preconditions checked before actions. Conditions are prepended, actions appended. This is a design-time convenience — it's semantically equivalent to a Sequence.

### Monitor
A Parallel with read-only conditions and read-write actions. Prevents synchronization issues by keeping condition-checking separate from action execution. Useful for maintaining assumptions while a behavior is active.

> "Many behaviors tend to have assumptions that should be maintained while a behavior is active, and if those assumptions are found invalid the whole sub-tree should exit."
> — *The Behavior Tree Starter Kit*

### Active Selector
Re-checks decisions each tick, allowing high-priority behaviors to interrupt lower-priority ones. Implements reactivity without a full event-driven architecture.

### Interrupt Node
Forces immediate tree reevaluation when a high-priority event occurs (e.g., taking damage). Interrupts know nothing about AI behaviors — they simply trigger reevaluation.

> "Interrupts are basically events that will force the behavior tree logic to immediately invalidate the current action and initiate a full reevaluation of the tree."
> — *A Reactive AI Architecture for Networked FPS Games*

### BeeTree Takeaway

BeeTree should ship the six core nodes as built-in types and provide a clear extension mechanism for specialized composites. Users should be able to define custom node types (like Filter, Monitor, Active Selector) that compose the core primitives or introduce new traversal semantics.

---

## 5. The Blackboard Pattern

The blackboard is a shared data store that behavior tree nodes use to communicate. It provides a key-value interface for reading and writing world state data.

### Architecture Considerations

The FFXV system uses **local** (per-subtree) and **global** blackboards. The reactive FPS architecture uses **function-based variables** that auto-calculate values on query. The pitfalls chapter warns strongly against coupling the BT too tightly to the blackboard:

> "It is probably a mistake to tie the overall behavior tree logic too closely to the logic of the blackboard — and even more of a mistake during prototyping to make simple tasks dependent on a blackboard for communication."
> — *Overcoming Pitfalls in Behavior Tree Design*

The recommendation: the same behavior tree logic should work with a complex hierarchical blackboard, a simple struct with a handful of values, or no explicit blackboard at all.

### BeeTree Takeaway

BeeTree's specification should define blackboard variables as typed key-value pairs associated with a tree, but the blackboard implementation should be separate from the tree structure. Generated code should provide a default blackboard implementation per engine, while allowing users to substitute their own.

---

## 6. Reactive and Event-Driven Patterns

Reactivity — the ability to respond to changing conditions without waiting for the next scheduled evaluation — is a major concern in the literature.

### Approaches

1. **Active Selectors**: Re-check higher-priority branches each tick. Simple but potentially expensive.

2. **Interrupt System**: Condition nodes annotated with interrupt types (e.g., `"interrupt": "damage"`) only evaluate when their specified interrupt is being processed. When multiple interrupts occur on the same frame, use a predefined priority list and process only the highest.

3. **Parallel Conditions**: Conditions marked as parallel remain valid throughout action execution. If any condition fails, the action is invalidated and the tree reevaluates.

4. **HeBT Hints**: High-level trees send "hints" to reorder priorities in base trees dynamically, without modifying the base behavior. This enables safe prototyping and difficulty adaptation.

> "The key to this type of prototyping is that the new logic is completely optional. We can just remove the high-level tree, which will leave our original behavior untouched."
> — *Building a Risk-Free Environment to Enhance Prototyping*

### BeeTree Takeaway

BeeTree should support annotating nodes with reactivity semantics (interrupt triggers, parallel conditions). The specification format should capture these annotations so that code generators can produce appropriate reactive scaffolding for each target engine.

---

## 7. Utility-Based Decision Making

Standard behavior trees use static priority — the position in the tree determines evaluation order. Utility theory extends this with dynamic scoring.

### The Problem

> "In a standard behavior tree, priority is static. It is baked right into the tree. The simplicity is welcome, but in practice it can be frustratingly limiting."
> — *Building Utility Decisions into Your Existing Behavior Tree*

### The Solution: Utility Selector

A specialized selector that queries children for utility values instead of just binary validity. Selection methods include:

- **Highest utility**: Select child with maximum value
- **Weighted random**: Proportional random selection
- **Threshold-based random**: Unweighted random among children exceeding a threshold

Utility values must be **normalized to 0–1** for comparison. Common normalization functions: linear, exponential, sigmoidal, response curves.

> "Rather than creating variation through randomness or forcing agents to arbitrarily take one valid option over another, we can apply existing, well-documented techniques to deal with 'gray area' decisions in an elegant manner."
> — *Building Utility Decisions into Your Existing Behavior Tree*

### Integration Pattern

Utility decorators transform utility values as they propagate upward (multiply, scale, apply curves). The key insight is separating **evaluation** from **execution** — evaluate children before calculating utility, cache results, then execute the winner.

### BeeTree Takeaway

The Utility Selector should be a first-class extension node in BeeTree's library (not core, but officially supported). The specification should allow attaching utility scoring functions to any node, with the scoring function defined in the target language.

---

## 8. Data-Driven and Template-Based Design

Multiple chapters emphasize the importance of data-driven behavior trees — defining trees in data rather than code.

### Key Principle

> "The key functionality needed to make a data-driven behavior tree is an engine that turns text into objects... you need to be able to specify the tree — and specify what objects get produced from that tree."
> — *Template Tricks for Data-Driven Behavior Trees*

### Format Recommendations

> "I recommend that you choose a language which is more directly interpretable as standard programming language data structures, like JSON — or, better, something like Protocol Buffers... the important part is to let someone else do the heavy lifting of creating a parser, and focus on your tree."
> — *Template Tricks for Data-Driven Behavior Trees*

### The Registration Pattern

From Google's robotics team: each task type registers itself via a factory. The entire framework is hidden behind two operations:
- `REGISTER_TASK(YourTask)` — one-line registration
- `LoadScript(path)` — one-line loading

### Runtime Reloading

> "For AI development, fast iteration is one of the most important features to keep the AI improving until the end of development. As such, a user should be able to reload an AI Graph without compiling when they want to make a change."
> — *A Character Decision-Making System for FINAL FANTASY XV*

### BeeTree Takeaway

This is BeeTree's core value proposition. The specification format (JSON/YAML) is the data-driven representation. The CLI generates engine-specific code with proper registration patterns. The factory + registration pattern should be part of the generated scaffolding for each target engine.

---

## 9. Combining Behavior Trees with State Machines

The FFXV system demonstrates a powerful hybrid architecture.

### Architecture

Users build visual node graphs where:
- Top-layer state machines define high-level character states
- Behavior trees are embedded inside states at any level
- Nesting continues indefinitely

Every node has four unified methods: `start`, `update`, `finalize`, `terminationCondition`. The key difference: BT nodes terminate themselves via internal conditions; FSM nodes are terminated by external transition conditions.

> "After many discussions within the team, the idea to combine state machines and behavior trees was agreed upon. This allows our developers to leverage both techniques in a nested hierarchical node structure, enabling a very flexible architecture."
> — *A Character Decision-Making System for FINAL FANTASY XV*

### Three-Layer Character Architecture
1. **Intelligence Layer** (AI Graph) — Decision-making using combined BTs/FSMs
2. **Body Layer** — State machine for physical body state
3. **Animation Layer** — AnimGraph with FSM and blend trees

Communication between layers uses **messages** and **blackboard variables**.

### BeeTree Takeaway

BeeTree should support referencing sub-trees and potentially FSM states within a behavior tree definition. The specification format should allow composing trees from reusable sub-tree modules. Consider FSM integration as a future extension.

---

## 10. Simulation, Testing, and Debugging

### Simulation-Based Decision Making

The Star Wars Jedi system uses a hybrid approach where each action can simulate itself:

> "Behavior trees are great at allowing designers to define exactly what an AI can do, and planners are great at allowing designers to easily specify what an AI should do... This approach provides designers with all of the control of a behavior tree and all of the durability and flexibility of a planner."
> — *Simulating Behavior Trees*

### Multi-Valued Result Types

Rather than binary scoring, the Jedi system classifies simulation results as: Impossible, Hurtful, Irrelevant, Cosmetic, Beneficial, Urgent, Deadly. This provides more stable, predictable behavior than continuous 0–1 scoring.

### FakeSim Decorator

A special decorator wraps actions to insert incorrect information during simulation only. This allows AI to "miscalculate" and choose suboptimal actions, enabling designers to demonstrate mistakes without modifying action code.

### Debugging Visualization

> "While a game program runs, an AI Graph keeps a connection with the program. The active node currently being executed is highlighted in green. This enables a user to trace the active node in real time."
> — *A Character Decision-Making System for FINAL FANTASY XV*

### BeeTree Takeaway

BeeTree's TUI should include a tree visualization mode showing active nodes (similar to the FFXV debugger). The specification format should support metadata annotations for debugging labels, breakpoints, and simulation parameters. Generated code should include debug hooks.

---

## 11. Common Pitfalls and How to Avoid Them

The "Overcoming Pitfalls" chapter identifies three major traps:

### Pitfall 1: Creating Too Many Organizing Classes

Adding separate categories for every system need (Behaviors, Composites, Skills, Modules) creates unnecessary complexity.

**Solution**: Use a single unified base class. "It's Tasks, All the Way Down."

> "We ultimately adopted a mantra 'It's Tasks, All the Way Down.'"
> — *Overcoming Pitfalls in Behavior Tree Design*

### Pitfall 2: Implementing a Language Too Soon

Building a complete programming language into the BT (Not, And, Or, If-Then-Else, Cond, Loop, For, While) before understanding requirements leads to bloat.

**Solution**: Start with the five core primitives: Actions, Decorators, Sequences, Parallels, and Selectors. Expand only when proven necessary.

> "You can do a lot with a little; even very simple decision-making structures are Turing-complete... you can implement sophisticated control behaviors with very few node types."
> — *Overcoming Pitfalls in Behavior Tree Design*

### Pitfall 3: Forcing All Communication Through the Blackboard

Tying BT logic too closely to a centralized blackboard creates rigid dependencies.

**Solution**: Decouple. The same behavior tree logic should work with different communication mechanisms.

### Key Design Principle

> "A system which functionally breaks behaviors into atomic behaviors orchestrated by trees of composite behaviors may act like a behavior tree, but it will not be customizable or hackable if behaviors are too tightly coupled, nor will it be explicit or variable if the nodes are too large."
> — *Overcoming Pitfalls in Behavior Tree Design*

### BeeTree Takeaway

BeeTree's core specification should remain minimal (six node types). The extension mechanism should be clean and well-defined. The generated code should use a unified base class per target engine. Blackboard integration should be opt-in, not forced.

---

## 12. Behavior Trees vs. Alternatives

The literature positions behavior trees among several behavior selection approaches:

| Approach | Strengths | Weaknesses |
|----------|-----------|------------|
| **FSMs** | Simple, fast, intuitive | Don't scale; transition overload |
| **HFSMs** | Better scaling via nesting | Still complex implementation |
| **Behavior Trees** | Simple, modular, extensible, loosely coupled | Slower runtime (root traversal); stateless (can cause loops) |
| **Utility Systems** | Handle "gray area" decisions; emergent behavior | Unpredictable; difficult to tune |
| **GOAP** | Novel emergent solutions; streamlined development | Loss of authorial control; harder to inject specific behavior |
| **HTNs** | Natural hierarchy; designer-expressive | No emergent solutions; requires hand-built network |

Key BT advantage:

> "Behaviors can (and should) be written to be completely unaware of each other, so adding or removing behaviors from a character's behavior tree do not affect the running of the rest of the tree. This alleviates the problem common with FSMs, where every state must know the transition criteria for every other state."
> — *Behavior Selection Algorithms*

Key BT limitation:

> "Since the tree must run from the root every time behaviors are selected, the running time is generally greater than that of a finite-state machine."
> — *Behavior Selection Algorithms*

### BeeTree Takeaway

BeeTree is firmly in the behavior tree camp but should acknowledge hybrid approaches. The utility selector extension bridges the gap with utility systems. FSM integration (à la FFXV) could be a future extension. The specification should be rich enough to express these hybrid patterns.

---

## 13. Code Generation and Scripting Patterns

### Hybrid Architecture Pattern

The Kingdoms of Amalur: Reckoning team used C++/Lua where:
- Core algorithm in C++ (performance, engineering control)
- Behaviors in Lua (rapid iteration, wider team involvement)

> "Being able to edit and reload behaviors at runtime is a huge advantage when refining and debugging behaviors, and having a data-driven approach to behavior creation opens up the process to a much wider group of people on the team."
> — *Real-World Behavior Trees in Script*

### Selector Flexibility

> "Any kind of selection algorithm can work with a behavior tree, which is one of the major strengths of the system."
> — *Real-World Behavior Trees in Script*

Special selector types used in production:
- **Nonexclusive**: Continue checking siblings after running
- **Sequential**: Execute all children in order while parent's precondition is true
- **Stimulus**: React to in-game events

### Search-Enhanced Behavior Trees

Selector nodes can serve as "choice points" where look-ahead search explores different options:

> "By using scripts, it allows game designers to keep control over the range of behaviors the AI system can perform, whereas the adversarial look-ahead search enables it to better evaluate action outcomes."
> — *Combining Scripted Behavior with Game Tree Search*

### BeeTree Takeaway

Generated code for each target engine should follow this hybrid pattern: a core runtime in the engine's native language (C# for Unity, C++ for Unreal, GDScript for Godot) with data-driven tree definitions loaded at runtime. BeeTree's specification format IS the data-driven layer.

---

## 14. Implications for BeeTree Design

### Core Specification Node Set

Based on all reviewed literature, BeeTree's core specification should include exactly six node types:

1. **Action** — Leaf node; executes behavior
2. **Condition** — Leaf node; checks world state
3. **Sequence** — Composite; AND logic
4. **Selector** — Composite; OR/fallback logic
5. **Parallel** — Composite; concurrent execution with configurable policies
6. **Decorator** — Wrapper; modifies single child behavior

### Extension Library

Officially supported but not core:
- **Utility Selector** — Dynamic priority via scoring
- **Active Selector** — Reactive re-evaluation each tick
- **Filter** — Sequence with preconditions
- **Monitor** — Parallel with read-only conditions
- **Interrupt** — Event-driven reevaluation
- **Repeat**, **Negate**, **Timeout**, **Cooldown** (common decorators)

### Design Principles

1. **Start minimal, expand when proven** — Six core nodes, clean extension mechanism
2. **Data-driven specification** — JSON/YAML tree definitions as the central artifact
3. **Engine-agnostic authoring** — Define once, generate for Unity/Unreal/Godot
4. **Loose coupling** — Nodes unaware of each other; blackboard is opt-in
5. **Unified base class** — Generated code uses "tasks all the way down"
6. **Runtime reloadable** — Generated code supports hot-reloading where possible
7. **Debuggable** — Tree visualization with active node highlighting

### Key Quotes to Guide Development

> "Modular behaviors with simple (unique) responsibilities, to be combined together to form more complex ones... Behaviors that are very well specified (even unit tested) and easy to understand, even by nontechnical designers."
> — *The Behavior Tree Starter Kit*

> "The real strength of a behavior tree comes from its simplicity."
> — *Behavior Selection Algorithms*

> "This pattern — refining your abstractions until the complexity is squirreled away in a few files and the leaves of your functionality are dead simple — can be applied everywhere."
> — *Template Tricks for Data-Driven Behavior Trees*

> "By taking time to think carefully about the needs of the game, an AI system can be crafted to give the best player experience while maintaining the balance between development time and ease of creation."
> — *Behavior Selection Algorithms*

---

## Resources

### Game AI Pro — First Edition (2013)

1. **Chapter 4: Behavior Selection Algorithms** — *Dave Mark & Mike Dawe*
   Comprehensive taxonomy of FSMs, HFSMs, behavior trees, utility systems, GOAP, and HTNs with pros/cons analysis.
   https://www.gameaipro.com/GameAIPro/GameAIPro_Chapter04_Behavior_Selection_Algorithms.pdf

2. **Chapter 6: The Behavior Tree Starter Kit** — *Alex J. Champandard & Philip Dunstan*
   Foundational chapter establishing core BT node types, tick/status model, tree traversal, and both first-generation and event-driven implementations.
   https://www.gameaipro.com/GameAIPro/GameAIPro_Chapter06_The_Behavior_Tree_Starter_Kit.pdf

3. **Chapter 7: Real-World Behavior Trees in Script** — *Mike Dawe*
   Practical hybrid C++/Lua BT implementation from Kingdoms of Amalur: Reckoning. Covers data-driven design, runtime reloading, and team collaboration patterns.
   https://www.gameaipro.com/GameAIPro/GameAIPro_Chapter07_Real-World_Behavior_Trees_in_Script.pdf

4. **Chapter 8: Simulating Behavior Trees** — *Daniel Broder*
   Hybrid BT/planner system for Star Wars Jedi. Covers action simulation, multi-valued result types, constraint systems, and the FakeSim decorator for error injection.
   https://www.gameaipro.com/GameAIPro/GameAIPro_Chapter08_Simulating_Behavior_Trees.pdf

5. **Chapter 10: Building Utility Decisions into Your Existing Behavior Tree** — *Dave Mark*
   Integrating utility theory with BTs via utility selectors, normalization patterns, and scoring propagation through composites.
   https://www.gameaipro.com/GameAIPro/GameAIPro_Chapter10_Building_Utility_Decisions_into_Your_Existing_Behavior_Tree.pdf

### Game AI Pro 2 (2015)

6. **Chapter 10: Building a Risk-Free Environment to Enhance Prototyping** — *Alex�Molina Serrano*
   HeBT (Hinted-Execution Behavior Trees) for safe prototyping. Covers hint-based priority reordering, designer empowerment, and difficulty adaptation (Driver: San Francisco).
   https://www.gameaipro.com/GameAIPro2/GameAIPro2_Chapter10_Building_a_Risk-Free_Environment_to_Enhance_Prototyping.pdf

### Game AI Pro 3 (2017)

7. **Chapter 9: Overcoming Pitfalls in Behavior Tree Design** — *Anthony Francis*
   Three major BT pitfalls: too many organizing classes, premature language design, and over-coupling to the blackboard. Introduces "Tasks All the Way Down" philosophy.
   https://www.gameaipro.com/GameAIPro3/GameAIPro3_Chapter09_Overcoming_Pitfalls_in_Behavior_Tree_Design.pdf

8. **Chapter 10: A Reactive AI Architecture for Networked First-Person Shooter Games** — *Sumeet Jakatdar*
   Partial tree evaluation, interrupt system for reactivity, parallel conditions, and network-optimized AI state for multiplayer FPS games.
   https://www.gameaipro.com/GameAIPro3/GameAIPro3_Chapter10_A_Reactive_AI_Architecture_for_Networked_First-Person_Shooter_Games.pdf

9. **Chapter 11: A Character Decision-Making System for FINAL FANTASY XV** — *Youichiro Miyake et al.*
   Hybrid BT/FSM architecture (AI Graph) for FFXV. Covers hierarchical nesting, parallel thinking, real-time visual debugging, and three-layer character architecture.
   https://www.gameaipro.com/GameAIPro3/GameAIPro3_Chapter11_A_Character_Decision-Making_System_for_FINAL_FANTASY_XV_by_Combining_Behavior_Trees_and_State_Machines.pdf

10. **Chapter 14: Combining Scripted Behavior with Game Tree Search** — *Nicolas A. Barriga et al.*
    Using BT selector nodes as "choice points" for look-ahead search. Covers script-based AI with search enhancement and anytime algorithms.
    https://www.gameaipro.com/GameAIPro3/GameAIPro3_Chapter14_Combining_Scripted_Behavior_with_Game_Tree_Search_for_Stronger_More_Robust_Game_AI.pdf

### Game AI Pro Online Edition (2021)

11. **Chapter 13: Template Tricks for Data-Driven Behavior Trees** — *Anthony Francis*
    Production-grade data-driven BT system using Protocol Buffers and C++ templates. Covers factory pattern, registration pattern, serialization/deserialization, and the "hide complexity, expose simplicity" design philosophy. From Google's robotics team.
    https://www.gameaipro.com/GameAIProOnlineEdition2021/GameAIProOnlineEdition2021_Chapter13_Template_Tricks_for_Data-Driven_Behavior_Trees.pdf
