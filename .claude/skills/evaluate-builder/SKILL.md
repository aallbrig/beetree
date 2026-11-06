---
name: evaluate-builder
description: Critically evaluate the behavior tree builder, agent inventory and templates to find the most worthwhile improvement. Use when improving how people author trees on beetreecraft.com.
---

# Evaluate the builder

Surface: the builder, scene studio and dashboard in the SPA (`web/src/views/builder.ts`, `scenes.ts`, `dashboard.ts`, `build.ts`) and the templates in `web/src/samples.ts`. This is the product's core promise, so hold it to the highest bar.

Probe, in order:
- **Blank to working tree**: sign up, open the builder, and author a guard that patrols and chases without reading docs. Count clicks and every hesitation.
- **Templates**: there should be 4 to 5 trope-inspired trees (patrolling guard, skittish critter, boss with phases, shopkeeper routine, loyal companion). For each: valid, runs in 3D, teaches one idea, has a plain-language description.
- **Editing power**: add, rename, reorder, duplicate, delete and reparent nodes; decorators and parallel policies; undo and redo; keyboard use; large trees.
- **Feedback**: validation messages that point at the node and say how to fix it; live node status during simulation; a way to see why a node ran.
- **Inventory**: an agent holds a tree; people can list, find, fork, rename and delete their agents.
- **Export**: copy or download YAML and Unity, Unreal and Godot code that compiles.
- **Persistence**: nothing is lost on refresh or sign-out.

## Method

1. Use it as a stranger who has never made a behavior tree (`VISION.md`, `docs/AUDIENCES.md`).
2. Compare with the builder lines of `VISION.md`: met, partly met or missing.
3. Rank gaps by value to a first-time builder, frequency and effort; keep the top 3.
4. Pick one that ships in this iteration, with a gate: a unit test on the spec logic, a `check_site` assertion, or a template that CI validates with the real validator.
5. Report the ranked list and the chosen item so the loop can file the rest as `loop-idea` issues.
