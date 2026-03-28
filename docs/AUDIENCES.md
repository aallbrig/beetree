# BeeTree Target Audiences

## Audience 1: Beginner Game Developer / Designer

**Who they are:** Students, hobbyists, or designers touching game AI for the first time. May be comfortable in Unity/Godot/Unreal but have never built a behavior tree. Might have read a blog post or GDC talk slide deck about BTs but haven't implemented one.

**What they know:**
- Basic game dev concepts (game loop, entities, physics)
- One game engine (probably Unity or Godot)
- What "AI" means in a game context (enemies that react)
- Possibly: state machines, if/else trees in code

**What they don't know:**
- Selector vs. sequence semantics
- Why BTs are better than FSMs for complex AI
- What a blackboard is and why it exists
- How to decompose behavior into actions and conditions
- What "tick" means in BT context

**What they need from BeeTree:**
- Conceptual onboarding: "What is a behavior tree and why should I use one?"
- Guided creation: templates, wizards, progressive examples
- Immediate visual feedback: see the tree, simulate it, understand the flow
- Clear integration path: "I made a tree, now how does my game use it?"
- Forgiving UX: undo, validation before save, helpful error messages

**Success metric:** Can go from `beetree init` to running generated code in their game engine within 30 minutes.

---

## Audience 2: Intermediate Game Developer

**Who they are:** Working game devs who have shipped at least one game or prototype with AI. May have hand-rolled a BT system before or used a visual BT editor (Unreal's built-in, NodeCanvas, Behavior Designer). Comfortable reading YAML and working in a terminal.

**What they know:**
- BT fundamentals (selector, sequence, action, condition)
- At least one game engine's BT system
- When to use BTs vs. other AI approaches
- How to structure AI for different agent types

**What they don't know:**
- Advanced BT patterns (utility selectors, parallel policies, subtree composition)
- How to make trees portable across engines
- Best practices for large-scale BT management (dozens of trees, shared subtrees)
- How to test/validate BT behavior systematically

**What they need from BeeTree:**
- Productivity: fast editing, batch operations, project-wide management
- Power features: diff, simulate with scenarios, validate across files
- Engine-specific guidance: how generated code maps to their engine's idioms
- Composability: subtrees, custom nodes, parameterized behaviors
- Export quality: generated code that follows engine conventions

**Success metric:** Can convert an existing game's AI from hand-written code to beetree-managed specs, with generated stubs that integrate cleanly.

---

## Audience 3: Expert / AI Architect

**Who they are:** Senior game programmers, AI leads, or researchers who design BT architectures for teams. May be evaluating BeeTree against existing tools (Unreal BT editor, BehaviorTree.CPP, py_trees). Want to standardize their team's workflow.

**What they know:**
- Deep BT theory (reactive vs. non-reactive, memory policies, concurrency models)
- Multiple game engines and their BT implementations
- Performance implications of BT design choices
- How to extend and customize BT frameworks

**What they don't know (and want to evaluate):**
- Whether BeeTree's spec format is expressive enough for their needs
- Whether code generation is customizable for their engine setup
- Whether the tool can scale to their team's workflow (CI/CD, version control, reviews)
- Whether the registry/sharing model works for internal team use

**What they need from BeeTree:**
- Specification completeness: all standard BT semantics covered
- Extensibility: custom generators, node types, validation rules
- CI/CD integration: validate in pipelines, generate in build scripts, diff in PRs
- Team workflow: registry for internal sharing, tree versioning, review tooling
- Performance: handle large trees without slowdown

**Success metric:** Can adopt BeeTree as the team's standard BT authoring tool, replacing engine-specific editors with a portable workflow.

---

## Cross-Cutting Needs (All Audiences)

| Need | Beginner Priority | Intermediate Priority | Expert Priority |
|------|:-:|:-:|:-:|
| Clear documentation | Critical | High | Medium |
| Working examples | Critical | Medium | Low |
| Interactive TUI | High | Critical | Medium |
| CLI scriptability | Low | High | Critical |
| Code generation | High | Critical | Critical |
| Error messages | Critical | High | Medium |
| Simulation | High | Critical | High |
| Registry/sharing | Low | Medium | High |
| Extensibility | Low | Low | Critical |
| CI/CD integration | Low | Medium | Critical |
