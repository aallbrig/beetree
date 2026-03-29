# Design: CLI Feature Demo Video System

**Date:** 2026-03-29
**Status:** Accepted

## Problem

BeeTree needs visual documentation showing its features in action — the sigil
notation, TUI editor, code generation, simulation. Static screenshots go stale,
hand-recorded screencasts are inconsistent, and there's no automated way to
detect when the video needs re-recording after code changes.

## Solution

Use [VHS](https://github.com/charmbracelet/vhs) to script deterministic demo
recordings as `.tape` files. Individual feature segments are recorded separately,
then stitched into a single MP4 via `ffmpeg`. A nightly CI job detects code
changes and regenerates the video automatically.

---

## Architecture

```
demos/                   # Directory of .tape files
  _settings.tape         # Shared VHS settings (included by all tapes)
  01_project_init.tape   # One tape per feature segment
  02_render_formats.tape
  ...
  NN_<feature>.tape
    |
    v
scripts/
  record-demo.sh         # Orchestrator: records each tape, stitches into MP4
    |
    v
dist/
  demo.mp4               # Final combined video
```

## Tape Conventions

### Shared Settings (`demos/_settings.tape`)

Visual settings reused across all tapes. Individual tapes do NOT set global
visuals — they only set `Output` and `Require`.

### Feature Tapes (`NN_<name>.tape`)

Each tape records one feature segment:
- Two-digit prefix for ordering (`01_`, `02_`, etc.)
- Descriptive snake_case name after the prefix
- Starts with ANSI banner identifying the feature
- `Hide`/`Show` wraps the `clear` command for clean transitions

Pattern:
```tape
Require beetree
Source demos/_settings.tape
Output demos/segments/NN_<name>.mp4

Hide
Type "clear"
Enter
Show

Sleep 300ms
Type "printf '\\n  \\033[1;36m── Feature Name ──\\033[0m\\n\\n'"
Enter
Sleep 1s

Type "<demo command>"
Sleep 300ms
Enter
Sleep 3s
Sleep 500ms
```

## Features to Demo

Based on `docs/FEATURES.md`, these features are demonstrated in order:

| # | Tape File | Feature | Demo Commands |
|---|-----------|---------|---------------|
| 01 | `01_project_init.tape` | Project scaffold | `beetree init`, `beetree new patrol_ai` |
| 02 | `02_render_ascii.tape` | ASCII tree rendering | `beetree render examples/full-enemy-ai.beetree.yaml` |
| 03 | `03_render_sigil.tape` | Sigil notation | `beetree render --format sigil`, `--format compact`, `--format oneline` |
| 04 | `04_render_diagrams.tape` | Diagram export | `beetree render --format mermaid`, `--format dot` |
| 05 | `05_validate.tape` | Spec validation | `beetree validate examples/patrol.beetree.yaml` |
| 06 | `06_simulate.tape` | Batch simulation | `beetree simulate` with `--override` |
| 07 | `07_generate_code.tape` | Code generation | `beetree generate unity --dry-run`, `generate godot` |
| 08 | `08_diff.tape` | Tree comparison | `beetree diff examples/patrol.beetree.yaml examples/combat.beetree.yaml` |
| 09 | `09_node_cli.tape` | CLI tree editing | `beetree node list`, `node add`, `render` |
| 10 | `10_tui_editor.tape` | Interactive TUI | `beetree builder` with navigation + simulation |
| 11 | `11_version.tape` | Version & completion | `beetree version`, `beetree completion bash` |

## CI Workflow

Nightly job compares source paths against last recorded commit (tracked via
`demo-recorded-<sha>` git tags). Re-records only when CLI source, tape files,
or build config have changed.

## Makefile Integration

```makefile
demo:        ## Record all VHS tapes and produce dist/demo.mp4
demo-clean:  ## Remove recorded segments and final video
```
