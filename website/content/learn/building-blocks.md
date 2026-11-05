---
title: "Building blocks"
description: "The six node types every behavior tree is built from, with pseudocode for Unity, Godot and Unreal."
weight: 20
---

Every behavior tree, in any engine, is made from the same six node types. This page describes each one and shows it in code. The code is **pseudocode you can adapt**: it is the smallest version that works, not a production library. Pick your engine and the code samples will follow you down the page.

{{< langpicker >}}

## The contract: tick and status

Every node has one operation, **tick**. A parent ticks a child and gets back a **status**.

| Status | Meaning |
|---|---|
| **Success** | The node achieved its goal. |
| **Failure** | The node could not achieve its goal. |
| **Running** | The node is still working. Tick it again next frame. |

The root is ticked on a schedule, usually once per frame or on an AI update interval. A node only runs when its parent ticks it.

{{% lang unity "Unity C#" %}}
```csharp
public enum Status { Success, Failure, Running }

public abstract class Node
{
    public abstract Status Tick();
}
```
{{% /lang %}}
{{% lang godot "Godot GDScript" %}}
```gdscript
class_name BTNode extends RefCounted

enum Status { SUCCESS, FAILURE, RUNNING }

func tick() -> Status:
    return Status.FAILURE   # override in each node type
```
{{% /lang %}}
{{% lang unreal "Unreal C++" %}}
```cpp
enum class EStatus : uint8 { Success, Failure, Running };

class FNode
{
public:
    virtual ~FNode() = default;
    virtual EStatus Tick() = 0;
};
```
{{% /lang %}}

All the examples below build on this contract. The six types fall into two groups: **leaves** that do the work (Action and Condition) and **control-flow nodes** that decide which children to tick (Sequence, Fallback, Parallel and Decorator).

## Summary

| Node | Symbol | Succeeds | Fails | Running |
|---|---|---|---|---|
| **Action** | box with a label | on completion | if it cannot complete | while in progress |
| **Condition** | box with a label | if true | if false | never |
| **Sequence** | → | if all children succeed | if one child fails | if one child is running |
| **Fallback** | ? | if one child succeeds | if all children fail | if one child is running |
| **Parallel** | ⇒ | if at least M children succeed | if more than N − M children fail | otherwise |
| **Decorator** | ◇ | custom | custom | custom |

*Source: Table 1.1 of Colledanchise and Ögren, [Behavior Trees in Robotics and AI](https://arxiv.org/abs/1709.00084). Fallback is called **Selector** in many engines.*

## Action

An **action** does something in the world: move, play an animation, fire a weapon. It returns **Running** while it works, **Success** when it finishes and **Failure** if it cannot be done.

{{% lang unity "Unity C#" %}}
```csharp
public class ActionNode : Node
{
    readonly System.Func<Status> work;
    public ActionNode(System.Func<Status> work) { this.work = work; }
    public override Status Tick() => work();
}

// e.g. move toward the target; Running until we arrive
new ActionNode(() => agent.remainingDistance > 0.5f ? Status.Running : Status.Success);
```
{{% /lang %}}
{{% lang godot "Godot GDScript" %}}
```gdscript
class Action extends BTNode:
    var work: Callable
    func _init(w: Callable): work = w
    func tick() -> Status: return work.call()

# e.g. move toward the target; RUNNING until we arrive
var move := Action.new(func():
    return Status.RUNNING if nav.distance_to_target() > 0.5 else Status.SUCCESS)
```
{{% /lang %}}
{{% lang unreal "Unreal C++" %}}
```cpp
class FAction : public FNode
{
public:
    TFunction<EStatus()> Work;
    explicit FAction(TFunction<EStatus()> InWork) : Work(MoveTemp(InWork)) {}
    EStatus Tick() override { return Work(); }
};

// e.g. move toward the target; Running until we arrive
FAction Move([this]{ return Distance() > 50.f ? EStatus::Running : EStatus::Success; });
```
{{% /lang %}}

## Condition

A **condition** checks something and never takes time. It returns **Success** if the check is true and **Failure** if it is false. It never returns Running. Typical examples are "can I see the player?" and "is my health low?".

{{% lang unity "Unity C#" %}}
```csharp
public class ConditionNode : Node
{
    readonly System.Func<bool> check;
    public ConditionNode(System.Func<bool> check) { this.check = check; }
    public override Status Tick() => check() ? Status.Success : Status.Failure;
}

new ConditionNode(() => health < 0.25f);
```
{{% /lang %}}
{{% lang godot "Godot GDScript" %}}
```gdscript
class Condition extends BTNode:
    var check: Callable
    func _init(c: Callable): check = c
    func tick() -> Status:
        return Status.SUCCESS if check.call() else Status.FAILURE

var low_health := Condition.new(func(): return health < 0.25)
```
{{% /lang %}}
{{% lang unreal "Unreal C++" %}}
```cpp
class FCondition : public FNode
{
public:
    TFunction<bool()> Check;
    explicit FCondition(TFunction<bool()> InCheck) : Check(MoveTemp(InCheck)) {}
    EStatus Tick() override { return Check() ? EStatus::Success : EStatus::Failure; }
};

FCondition LowHealth([this]{ return Health < 0.25f; });
```
{{% /lang %}}

## Sequence

A **sequence** runs its children **in order, left to right**, and stops at the first one that does not succeed. It means "do this, *and then* this, *and then* this". It succeeds only if every child succeeds. If a child fails, the sequence fails. If a child is running, the sequence is running and picks up from the top on the next tick.

{{% lang unity "Unity C#" %}}
```csharp
public class Sequence : Node
{
    readonly Node[] children;
    public Sequence(params Node[] children) { this.children = children; }

    public override Status Tick()
    {
        foreach (var child in children)
        {
            var status = child.Tick();
            if (status != Status.Success) return status; // Failure or Running stops here
        }
        return Status.Success;
    }
}
```
{{% /lang %}}
{{% lang godot "Godot GDScript" %}}
```gdscript
class Sequence extends BTNode:
    var children: Array[BTNode]
    func _init(c: Array[BTNode]): children = c

    func tick() -> Status:
        for child in children:
            var status := child.tick()
            if status != Status.SUCCESS:
                return status   # FAILURE or RUNNING stops here
        return Status.SUCCESS
```
{{% /lang %}}
{{% lang unreal "Unreal C++" %}}
```cpp
class FSequence : public FNode
{
public:
    TArray<TSharedPtr<FNode>> Children;

    EStatus Tick() override
    {
        for (const TSharedPtr<FNode>& Child : Children)
        {
            const EStatus Status = Child->Tick();
            if (Status != EStatus::Success) return Status; // Failure or Running stops here
        }
        return EStatus::Success;
    }
};
```
{{% /lang %}}

## Fallback

A **fallback** (also called a **selector**) is the mirror image: it tries its children **in order until one works**. It means "try this, *or else* this, *or else* this", which makes it the natural way to express priorities. It succeeds as soon as one child succeeds and fails only if every child fails.

{{% lang unity "Unity C#" %}}
```csharp
public class Fallback : Node
{
    readonly Node[] children;
    public Fallback(params Node[] children) { this.children = children; }

    public override Status Tick()
    {
        foreach (var child in children)
        {
            var status = child.Tick();
            if (status != Status.Failure) return status; // Success or Running stops here
        }
        return Status.Failure;
    }
}
```
{{% /lang %}}
{{% lang godot "Godot GDScript" %}}
```gdscript
class Fallback extends BTNode:
    var children: Array[BTNode]
    func _init(c: Array[BTNode]): children = c

    func tick() -> Status:
        for child in children:
            var status := child.tick()
            if status != Status.FAILURE:
                return status   # SUCCESS or RUNNING stops here
        return Status.FAILURE
```
{{% /lang %}}
{{% lang unreal "Unreal C++" %}}
```cpp
class FFallback : public FNode
{
public:
    TArray<TSharedPtr<FNode>> Children;

    EStatus Tick() override
    {
        for (const TSharedPtr<FNode>& Child : Children)
        {
            const EStatus Status = Child->Tick();
            if (Status != EStatus::Failure) return Status; // Success or Running stops here
        }
        return EStatus::Failure;
    }
};
```
{{% /lang %}}

Sequence and fallback combine into the most common shape in a game tree: a fallback of sequences, where each sequence is "if *condition*, then do *action*". The first branch whose condition holds wins.

## Parallel

A **parallel** ticks **all** its children on every tick. You choose a threshold **M** out of **N** children. It succeeds when at least M children have succeeded. It fails when so many have failed that M successes are no longer possible, which means more than N − M failures. Otherwise it is running. Use it for things that happen together, such as walking while aiming.

{{% lang unity "Unity C#" %}}
```csharp
public class Parallel : Node
{
    readonly Node[] children;
    readonly int successThreshold; // M
    public Parallel(int m, params Node[] children) { successThreshold = m; this.children = children; }

    public override Status Tick()
    {
        int successes = 0, failures = 0;
        foreach (var child in children)   // every child is ticked every time
        {
            var status = child.Tick();
            if (status == Status.Success) successes++;
            else if (status == Status.Failure) failures++;
        }
        if (successes >= successThreshold) return Status.Success;
        if (failures > children.Length - successThreshold) return Status.Failure;
        return Status.Running;
    }
}
```
{{% /lang %}}
{{% lang godot "Godot GDScript" %}}
```gdscript
class Parallel extends BTNode:
    var children: Array[BTNode]
    var success_threshold: int   # M
    func _init(m: int, c: Array[BTNode]): success_threshold = m; children = c

    func tick() -> Status:
        var successes := 0
        var failures := 0
        for child in children:       # every child is ticked every time
            match child.tick():
                Status.SUCCESS: successes += 1
                Status.FAILURE: failures += 1
        if successes >= success_threshold: return Status.SUCCESS
        if failures > children.size() - success_threshold: return Status.FAILURE
        return Status.RUNNING
```
{{% /lang %}}
{{% lang unreal "Unreal C++" %}}
```cpp
class FParallel : public FNode
{
public:
    TArray<TSharedPtr<FNode>> Children;
    int32 SuccessThreshold = 1; // M

    EStatus Tick() override
    {
        int32 Successes = 0, Failures = 0;
        for (const TSharedPtr<FNode>& Child : Children) // every child is ticked every time
        {
            const EStatus Status = Child->Tick();
            if (Status == EStatus::Success) ++Successes;
            else if (Status == EStatus::Failure) ++Failures;
        }
        if (Successes >= SuccessThreshold) return EStatus::Success;
        if (Failures > Children.Num() - SuccessThreshold) return EStatus::Failure;
        return EStatus::Running;
    }
};
```
{{% /lang %}}

## Decorator

A **decorator** wraps **one child** and changes how it behaves. It can change the child's result, limit how often the child runs, or decide whether to tick it at all. What it does is up to you, which is why the table says "custom". Common examples are **Inverter** (swap success and failure), **Retry N times**, **Timeout** and **Cooldown**.

{{% lang unity "Unity C#" %}}
```csharp
public class Inverter : Node
{
    readonly Node child;
    public Inverter(Node child) { this.child = child; }

    public override Status Tick() => child.Tick() switch
    {
        Status.Success => Status.Failure,
        Status.Failure => Status.Success,
        _ => Status.Running,
    };
}

public class Retry : Node   // let the child fail up to N times
{
    readonly Node child; readonly int max; int failures;
    public Retry(Node child, int max) { this.child = child; this.max = max; }

    public override Status Tick()
    {
        if (failures >= max) return Status.Failure;   // give up without ticking the child
        var status = child.Tick();
        if (status == Status.Failure) failures++;
        return status == Status.Failure && failures < max ? Status.Running : status;
    }
}
```
{{% /lang %}}
{{% lang godot "Godot GDScript" %}}
```gdscript
class Inverter extends BTNode:
    var child: BTNode
    func _init(c: BTNode): child = c

    func tick() -> Status:
        match child.tick():
            Status.SUCCESS: return Status.FAILURE
            Status.FAILURE: return Status.SUCCESS
            _: return Status.RUNNING

class Retry extends BTNode:   # let the child fail up to N times
    var child: BTNode
    var max_failures: int
    var failures := 0
    func _init(c: BTNode, n: int): child = c; max_failures = n

    func tick() -> Status:
        if failures >= max_failures: return Status.FAILURE   # give up without ticking the child
        var status := child.tick()
        if status == Status.FAILURE:
            failures += 1
            return Status.RUNNING if failures < max_failures else Status.FAILURE
        return status
```
{{% /lang %}}
{{% lang unreal "Unreal C++" %}}
```cpp
class FInverter : public FNode
{
public:
    TSharedPtr<FNode> Child;

    EStatus Tick() override
    {
        switch (Child->Tick())
        {
        case EStatus::Success: return EStatus::Failure;
        case EStatus::Failure: return EStatus::Success;
        default:               return EStatus::Running;
        }
    }
};

class FRetry : public FNode   // let the child fail up to N times
{
public:
    TSharedPtr<FNode> Child;
    int32 MaxFailures = 3;
    int32 Failures = 0;

    EStatus Tick() override
    {
        if (Failures >= MaxFailures) return EStatus::Failure; // give up without ticking the child
        const EStatus Status = Child->Tick();
        if (Status == EStatus::Failure)
        {
            ++Failures;
            return Failures < MaxFailures ? EStatus::Running : EStatus::Failure;
        }
        return Status;
    }
};
```
{{% /lang %}}

## Putting it together

A guard that flees when hurt, attacks when it sees the player and otherwise patrols is a fallback of three branches:

```yaml
tree:
  type: selector            # Fallback: first branch that works wins
  name: guard_ai
  children:
    - type: sequence        # branch 1: flee when hurt
      children:
        - { type: condition, name: low_health }
        - { type: action, name: run_to_cover }
    - type: sequence        # branch 2: attack when the player is visible
      children:
        - { type: condition, name: sees_player }
        - { type: action, name: attack }
    - { type: action, name: patrol }   # branch 3: default
```

That is the same shape you will build in the [Build hub](/app/#/build). For the exact file syntax, see [The tree format](/docs/trees/).

## Engine notes

The code above is the classical formulation, chosen for clarity. Real engines differ in details:

- **Unreal Engine** has its own built-in behavior tree system: the editor asset, Blackboards, Services, and decorators that act as conditions. Its composites are called *Selector* and *Sequence*, and it is event-driven rather than re-ticked from the root every frame. The C++ above shows the concept. To use Unreal's own system, subclass `UBTTaskNode`, `UBTDecorator` and `UBTService`.
- **Unity** has no built-in behavior tree runtime, so teams use a package or write their own. The C# above is close to what a small hand-written one looks like.
- **Godot** has no built-in behavior tree either. Add-ons such as LimboAI and Beehave provide one, and the GDScript above is the same idea in miniature.
- **Memory.** These simple versions re-tick from the first child on each tick. Many libraries instead remember which child was running ("memory" nodes) to save work. Both are valid, and they behave differently when something higher in the tree changes.
