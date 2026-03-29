# BeeTree Feature List

## Killer Features (Core Differentiators)

### 1. Interactive Simulation — "Walk the Tree"
The user steps through their behavior tree node by node, choosing the return status of each leaf node (Success/Failure/Running). The tree evaluates composites automatically, showing exactly how the tree would behave given those inputs. This turns an abstract YAML spec into something tangible — you can *see* the AI think.

**Status:** Implemented in TUI (press `r`). CLI batch mode also works with `--override` flags.
**Gaps:** No scenario save/replay, no step-back, no breakpoints, no pre-set overrides in TUI mode, no trace export.

### 2. Engine-Agnostic Spec → Native Code Generation
Define behavior trees once in `.beetree.yaml`, generate native code for Unity (C#), Unreal Engine (C++), and Godot (GDScript). First run generates stubs for user implementation; subsequent runs regenerate only the tree definition, preserving user code.

**Status:** Implemented for all three engines. Templates are embedded in the binary. Each generate run produces a README.md with integration instructions.
**Gaps:** No custom generator plugin system, no type mapping customization.

### 3. TUI Behavior Tree Editor
A full terminal-based tree editor: navigate, add/edit/delete/move nodes, undo (50 levels), save/save-as, and launch simulation — all without leaving the terminal.

**Status:** Implemented with tview. Keybindings: `a`dd, `e`dit, `d`elete, `m`ove, `p`arams, `b`board, `u`ndo, `U` redo, `c`opy, `r`un sim, `v`alidate, `/`search, `s`ave, `q`uit, `?` help.
**Gaps:** Can't edit metadata/custom_nodes in TUI.

### 4. Accessible to All Skill Levels
Progressive examples (patrol → combat → full enemy AI) with inline comments that teach BT concepts. Templates and guided commands lower the barrier to entry.

**Status:** Three well-documented examples. `beetree init` and `beetree new` provide scaffolding.
**Gaps:** No quickstart wizard, no in-tool tutorials, no concept explainer, website documentation is empty.

---

## Implemented Features

### CLI Commands
- `beetree init` — Scaffold a new project (beetree.yaml, trees/, subtrees/)
- `beetree new <name>` — Create a new tree from template (`--template`: default, patrol, combat, utility, blank)
- `beetree validate <file>` — Validate spec against schema (9 checks)
- `beetree render <file>` — Visualize as ASCII, Mermaid, or DOT/Graphviz
- `beetree simulate <file>` — Batch simulation with optional `--override` flags
- `beetree generate <engine> <file>` — Code generation (unity/unreal/godot)
- `beetree builder [file]` — Launch interactive TUI editor
- `beetree node list` — List available node types (core + extension)
- `beetree node add/remove/move` — CLI tree editing commands
- `beetree diff <a> <b>` — Compare two tree specs
- `beetree doctor` — Project health check
- `beetree registry browse/search/pull/push` — Registry commands (stub backend)
- `beetree version` — Version info

### Spec Format
- YAML and JSON parsing
- 6 core node types: action, condition, sequence, selector, parallel, decorator
- 5 extension types: utility_selector, active_selector, random_selector, random_sequence, subtree
- 9 built-in decorators: repeat, negate, always_succeed, always_fail, until_fail, until_succeed, timeout, cooldown, retry
- Blackboard with typed variables and defaults
- Custom node definitions with typed parameters
- Subtree references (file-based composition)

### Code Generation
- Unity: C# MonoBehaviour-based runtime, blackboard class, tree definition builder, action/condition stubs
- Unreal: C++ BTTaskNode/BTDecorator classes, tree factory, blackboard header
- Godot: GDScript node-based runtime, tree builder, action/condition stubs
- 27 embedded templates total
- Stub preservation: existing user code not overwritten on regeneration (unless `--overwrite`)
- Enriched stubs: generated stubs include node descriptions, parameters, and blackboard variable hints
- Dry-run mode for previewing output

### TUI Editor
- 3-pane layout: tree view, properties, status bar
- Full tree navigation with expand/collapse
- Add node with type selector modal
- Edit node properties (name, type, node class, decorator)
- Parameter editor: add/edit/remove key-value parameters
- Blackboard editor: add/edit/remove blackboard variables with types and defaults
- Delete, move (cut/paste), undo (50-level stack), redo (Shift+U)
- Copy/duplicate node (`c` key, deep-clones subtree with unique names)
- Search/find node by name or class (`/` key, live filtering, cycle matches)
- In-TUI validation (`v` key, also warns on save)
- Save, save-as (prompt for path), quit confirmation on unsaved changes
- Interactive simulation with step-through and visual trace
- Color-coded nodes during simulation (green=success, red=failure, yellow=running, cyan=current)

### Testing & CI
- 14 test suites, ~150+ tests
- Race detector enabled in CI
- go vet + staticcheck linting
- 6-platform cross-compilation (linux/darwin/windows x amd64/arm64)
- Coverage reporting

---

## Planned / Missing Features

### High Priority
- [x] In-TUI help overlay (`?` key)
- [x] In-TUI validation before save (`v` key)
- [x] Parameter editing in TUI (`p` key)
- [x] Blackboard editing in TUI (`b` key)
- [ ] Website documentation (getting started, CLI reference, engine guides)
- [ ] Quickstart wizard command
- [x] Template selection for `beetree new` (`--template` flag)

### Medium Priority
- [x] Redo stack (Shift+U)
- [x] Search/find node in TUI (`/` key)
- [x] Copy/duplicate node (`c` key)
- [ ] Simulation scenario save/replay
- [ ] Simulation step-back
- [ ] Trace export
- [x] Code generation integration guides per engine
- [x] Enriched generated stubs with descriptions and blackboard hints
- [ ] Node name validation (identifier rules)

### Low Priority / Future
- [ ] Real registry backend (currently stub)
- [ ] Custom generator plugin system
- [ ] Type mapping configuration for code generation
- [ ] `beetree generate --watch` for live regeneration
- [ ] goreleaser for distribution (Homebrew, GitHub Releases)
- [ ] Video tutorials
- [ ] VS Code extension for .beetree.yaml
