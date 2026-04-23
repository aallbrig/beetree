# BeeTree Feature List

## Killer Features (Core Differentiators)

### 1. Interactive Simulation — "Walk the Tree"
Step through behavior trees and choose leaf outcomes interactively.

**Status:** Implemented in TUI and CLI (`simulate --override`).
**Gaps:** No scenario save/replay, no step-back, no trace export, limited automated scenario assertions.

### 2. Engine-Agnostic Spec → Native Code Generation
Define once in `.beetree.yaml`; generate Unity, Unreal, and Godot outputs.

**Status:** Implemented with stub preservation and generated engine README output.
**Gaps:** No generator plugin model, limited type mapping control, no formal compatibility policy for generated APIs.

### 3. TUI Behavior Tree Editor
Terminal editor for authoring, navigation, validation, and simulation.

**Status:** Implemented with add/edit/delete/move, undo/redo, params, blackboard editing, search, help overlay, validation and save flows.
**Gaps:** No metadata/custom_nodes/subtrees editor UX parity; advanced spec authoring still partially YAML-first.

### 4. Progressive Example Set
Examples teach BT concepts from patrol to full enemy AI.

**Status:** Implemented with extensive inline teaching comments.
**Gaps:** No linked walkthrough docs/tutorial flow that connects examples to full production workflow.

---

## Implemented Feature Snapshot

### CLI
- `init`, `new`, `validate`, `render`, `simulate`, `generate`, `builder`, `diff`, `doctor`, `version`
- `node` subcommands: `list`, `add`, `edit`, `move`, `remove`
- `registry` command set exists but is hidden pending readiness

### Spec + Validation
- YAML/JSON parsing
- Core + extension node types
- Blackboard, custom nodes, subtree refs
- Structural validation of major tree constraints

### Code Generation
- Unity C# generation
- Unreal C++ generation
- Godot GDScript generation
- Generated + stub file split with overwrite controls

### TUI
- Tree navigation and core node editing
- Param editor and blackboard CRUD
- Search, undo/redo, in-tool help
- Interactive simulation and validation feedback

### Quality Signals
- Go test coverage across core modules (~150+ tests)
- CI-focused test/lint patterns present in repo

---

## Planned / Missing Features

### Product Gaps
- [ ] Production-ready registry backend and workflow
- [ ] Scenario persistence/replay/export for simulation
- [ ] Full advanced spec editing parity in TUI (metadata/custom nodes/subtrees)
- [ ] Website docs completeness (quickstart, references, engine workflows)
- [ ] Explicit extension model for custom generators/validators

### Spec-Driven Development (SDD) Gaps
- [ ] Architecture spec covering module boundaries and ownership
- [ ] CLI behavior contract spec (command semantics, stability expectations)
- [ ] TUI interaction/state-machine spec
- [ ] Spec schema + validation-rule source-of-truth doc
- [ ] Simulation semantics spec (node evaluation and decorator behavior)
- [ ] Code generation contract docs per engine (inputs, outputs, preservation rules)
- [ ] Error taxonomy + message quality standard
- [ ] Compatibility/versioning policy for `.beetree.yaml` and generated outputs

### Suggested Priority for SDD Adoption
1. Spec schema + validation contract
2. Simulation semantics contract
3. Code generation contract (shared + engine-specific)
4. TUI interaction/state contract
5. CLI command behavior contract
6. Architecture + contribution decision record index
