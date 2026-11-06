---
title: "Terminal editor (TUI)"
description: "Debug and simulate behavior trees in the terminal: step through each leaf, then build and edit trees without leaving the keyboard."
weight: 25
---

`beetree builder` opens a full-screen editor for a `.beetree.yaml` file. Its best feature is debugging: press `r` and the tree pauses at every leaf so you decide whether it succeeds, fails or keeps running, then watch how each composite resolves.

```
beetree builder examples/full-enemy-ai.beetree.yaml
```

## Step through a simulation

Select the tree and press `r`. The walker visits leaves in tree order and stops at each one. Press `S`, `F` or `R` to answer Success, Failure or Running. Sequences, selectors, parallels and decorators resolve from your answers, so you can reproduce a bug path by hand. `Esc` stops the run. When the tree finishes, the final status is shown and `q` returns to editing.

![The TUI paused at a leaf during a step-through simulation](/img/tui-step-through.gif)

For a non-interactive run, `beetree simulate <file>` plays the whole tree at once, and `--override name=FAILURE` forces a node's result. See the [CLI reference](/docs/cli-reference/#beetree-simulate).

## Build and edit

Move with the arrow keys, press `a` to add a child, `e` to edit, `m` to move, `u` to undo. Press `v` to validate and `s` to save. Press `?` any time for the in-app help.

![Navigating a tree and opening the help overlay in the TUI](/img/tui-editor.gif)

## Key bindings

| Mode | Key | Action |
|---|---|---|
| Navigate | `a` | Add a child node |
| Navigate | `e` | Edit the selected node |
| Navigate | `d` | Delete the selected node |
| Navigate | `c` | Duplicate the selected node |
| Navigate | `m` | Start a move, then press m on the target to place it |
| Navigate | `p` | Edit node parameters |
| Navigate | `b` | Edit blackboard variables |
| Navigate | `u` | Undo the last change |
| Navigate | `U` | Redo |
| Navigate | `r` | Run the step-through simulation |
| Navigate | `v` | Validate the tree |
| Navigate | `/` | Search nodes by name |
| Navigate | `s` | Save |
| Navigate | `S` | Save as a new path |
| Navigate | `q` | Quit (prompts if unsaved) |
| Navigate | `?` | Show the help overlay |
| Navigate | `Up/Down` | Move the selection |
| Navigate | `Left/Right` | Collapse or expand the node |
| Move | `Up/Down` | Choose the target parent |
| Move | `m` | Place the node under the target |
| Move | `Esc` | Cancel the move |
| Simulate | `S` | Resolve the paused leaf as Success |
| Simulate | `F` | Resolve the paused leaf as Failure |
| Simulate | `R` | Resolve the paused leaf as Running |
| Simulate | `Esc` | Stop the simulation |
| Simulate | `q` | Leave the finished simulation |
