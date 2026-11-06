# BeeTreeCraft vision

The aspirational brief for product work. When choosing what to build next, ask which line below the site does not yet live up to.

## Essence

BeeTreeCraft is where people who care about behavior trees want to be: a place to build them, learn them, and talk about them. It should feel modern, welcoming and alive, and a visitor with no prior knowledge should leave knowing more than when they arrived.

## What it offers

1. **Home page.** Invites people in. A clear promise above the fold, a live taste of the product, and sign up or sign in below the fold. Modern, attractive, fast.
2. **Agent inventory and builder.** Signed-in people keep an inventory of the agents they have made, each with its tree. Builder and inventory are enough to author real trees without leaving the site. Ships with 4 to 5 **templates** inspired by common video game tropes (for example: patrolling guard, skittish critter, boss with phases, shopkeeper's daily routine, loyal companion), each runnable in the live 3D preview.
3. **Forum.** Where people discuss behavior trees. Moderated: reporting, a review queue, rate limits and roles. Built on serverless AWS (SAM: Lambda, DynamoDB, API Gateway, Cognito) so it scales without servers to tend.
4. **Learning resource.** The best place to start and to go deeper. Original explainers plus a curated library of articles, centred on the Game AI Pro chapters and the book Behavior Trees in Robotics and AI. Credits live on the references page.
5. **Engine output.** One engine-agnostic tree format, native code for Unity, Unreal and Godot, and a CLI, REST API and MCP server that expose the same capabilities.

## Techniques for writing behavior trees

A comprehensive view, taught in Learn and reflected in the builder. Each technique gets an explainer, a runnable example tree and an engine-code sample.

- Recursive composition: subtrees as reusable nodes, trees that reference trees
- Event-driven trees: reacting to events instead of ticking everything every frame
- Graph structures: DAGs, shared subtrees, blackboards as data flow
- Concurrency: parallel nodes, success and failure policies, running state, interruption and abort
- Reactive versus memory nodes, and when each is right
- Decorators as policy: retries, cooldowns, timeouts, guards
- Hybrids: utility scoring, state machines, planners (GOAP, HTN) next to trees
- Data-driven trees: authoring in files, hot reload, versioning
- Testing and simulation: deterministic ticks, golden traces, regression of behavior
- Debugging and visualization: seeing why a node ran
- Scale: many agents, shared computation, budgeting ticks
- Newer ideas: LLM-assisted authoring and tree repair

## Quality bar

A feature counts when a gate proves it. Every change adds or updates a test, a `check_site` assertion, or a documentation example that CI validates with the real validator, and the gates stay representative of the features. Docs describe what the CLI, API and MCP actually do.

## Operating principles

- Ship small and often; the footer shows the live version.
- Release notes say what changed in plain language so the owner can review a night's work in the morning.
- Production health comes before new work.
