# BeeTree — Claude Code Project Context

## What This Is
BeeTree is a CLI/TUI for authoring, simulating, and exporting engine-agnostic behavior trees for video games. Target engines: Unity (C#), Unreal (C++), Godot (GDScript).

## Build & Test
```bash
make build          # Build binary to ./bin/beetree
make test           # Run all tests
make install        # Install to $GOPATH/bin
```

All Go code lives under `cli/`. Run commands from the repo root (Makefile handles `cd cli`).

```bash
cd cli && go test -race ./...          # Tests with race detector
cd cli && go vet ./...                 # Lint
cd cli && go build -o ../bin/beetree . # Manual build
```

## Project Structure
- `cli/cmd/` — Cobra CLI commands
- `cli/internal/model/` — TreeSpec, NodeSpec, Status types
- `cli/internal/tui/` — TUI editor (tview-based, model/view split)
- `cli/internal/codegen/` — Code generation (templates in `codegen/templates/`)
- `cli/internal/simulator/` — Batch simulation
- `cli/internal/treeedit/` — Tree mutation functions (add/remove/move/update)
- `cli/internal/validator/` — Spec validation
- `cli/internal/renderer/` — ASCII/Mermaid/DOT rendering
- `cli/internal/registry/` — Registry client (stub backend)
- `examples/` — Example .beetree.yaml files
- `docs/` — Audience profiles, feature list, introspect checkpoints
- `website/` — Hugo site (WIP)

## Conventions
- Test files live next to their source files
- Use `testify/assert` and `testify/require` for assertions
- Commit messages follow conventional commits: `feat:`, `fix:`, `ci:`, `docs:`
- The TUI uses a model/view split: EditorModel holds state, EditorView renders with tview

## Target Audiences
See `docs/AUDIENCES.md` — beginners, intermediates, and experts in game AI/BT design.

## Introspect Checkpoints
An **introspect checkpoint** is a periodic deep analysis of the tool's usability alignment with its target audiences. Run via `/introspect checkpoint`. Checkpoints produce:
- `docs/AUDIENCES.md` — Updated audience profiles
- `docs/FEATURES.md` — Current feature list with status and gaps
- `docs/introspect/<timestamp>_CHECKPOINT.md` — Full analysis with gap identification and improvement plan

The technique: read the entire codebase from the perspective of each audience tier, evaluate first-run experience, command discoverability, TUI completeness, simulation UX, code generation workflow, error messages, example quality, and documentation coverage. Be brutally honest. Identify what sucks and why.
