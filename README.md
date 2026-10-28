# BeeTree 🐝🌳

A behavior tree ecosystem for game developers. Create, validate, simulate, and deploy behavior trees to Unity, Unreal, and Godot — from the command line or from [beetreecraft.com](https://beetreecraft.com), where you can build trees visually, watch agents run them in a 3D scene, and share them with the community.

## Repository layout

| Path | What it is |
|---|---|
| `cli/` | The `beetree` CLI (Go) and the serverless API: `cli/internal/api` (REST + MCP + OpenAPI), `cli/lambda/*` (Lambda entry points) |
| `web/` | The web app (Vite + TypeScript + three.js): builder, 3D explore, scenes, community, forum, accounts |
| `website/` | The Hugo site: home page and all documentation |
| `infra/` | The SAM/CloudFormation stack: CloudFront, S3, API Gateway, Lambda, DynamoDB, Cognito |

## Develop

```bash
task setup        # one-time: git hooks, Go and npm dependencies
task ci           # every quality gate (no live services needed)
task dev          # tmuxinator session
task deploy       # deploy backend + site to AWS (needs AWS credentials)
task smoke        # end-to-end test of the deployed stack
```

## Features

- **Spec-driven design** — Define behavior trees in YAML with a rich node type system (6 core + 5 extension + 9 decorators + custom nodes)
- **Code generation** — Generate native engine code for Unity (C#), Unreal (C++), and Godot (GDScript)
- **Simulation** — Dry-run behavior tree execution with override support for scenario testing
- **Visualization** — Render trees as ASCII, Mermaid diagrams, or DOT/Graphviz
- **Spec diffing** — Structural comparison between tree versions
- **Registry** — Browse, search, pull, and push behavior tree specs
- **Validation** — Catch errors early: node types, blackboard refs, subtree references, custom nodes
- **Project health** — `beetree doctor` checks your setup

## Quick Start

```bash
# Initialize a project
beetree init

# Create a new behavior tree
beetree new enemy-ai

# Validate the spec
beetree validate trees/enemy-ai.beetree.yaml

# Simulate execution
beetree simulate trees/enemy-ai.beetree.yaml

# Generate Unity code
beetree generate unity trees/enemy-ai.beetree.yaml

# Render as Mermaid diagram
beetree render trees/enemy-ai.beetree.yaml --format mermaid
```

## CLI Commands

| Command | Description |
|---------|-------------|
| `beetree init` | Initialize a BeeTree project |
| `beetree new <name>` | Create a new behavior tree spec |
| `beetree validate <file>` | Validate a tree spec |
| `beetree render <file>` | Render tree (ascii/mermaid/dot) |
| `beetree generate <engine> <file>` | Generate engine code (unity/unreal/godot) |
| `beetree simulate <file>` | Simulate tree execution |
| `beetree diff <a> <b>` | Compare two tree specs |
| `beetree node list` | List available node types |
| `beetree registry browse` | Browse published trees |
| `beetree registry search <q>` | Search for trees |
| `beetree registry pull <name>` | Download a tree |
| `beetree registry push <file>` | Publish a tree |
| `beetree doctor` | Check project health |

## Spec Format

```yaml
version: "1.0"
metadata:
  name: enemy-ai
  description: Basic enemy AI with patrol and combat
  author: yourname

blackboard:
  - name: target
    type: Entity
  - name: health
    type: float
    default: 100.0

tree:
  type: selector
  name: root
  children:
    - type: sequence
      name: combat
      children:
        - type: condition
          name: has_target
          node: HasTarget
        - type: action
          name: attack
          node: Attack
    - type: action
      name: patrol
      node: Patrol
```

## Supported Engines

| Engine | Language | Output |
|--------|----------|--------|
| Unity | C# | MonoBehaviour runner, action/condition stubs, blackboard |
| Unreal | C++ | BTTaskNode/BTDecorator subclasses, header/source pairs |
| Godot | GDScript | Node classes, tree runner, blackboard dictionary |

## Development

```bash
cd cli
go test -count=1 ./...    # Run all 151 tests
go build -o beetree .      # Build the binary
```

## Website

Documentation hosted at [beetreecraft.com](https://beetreecraft.com) via GitHub Pages.

## Documentation

- [DESIGN.md](DESIGN.md) — Full specification and architecture
- [RESEARCH.md](RESEARCH.md) — Game AI Pro research notes
