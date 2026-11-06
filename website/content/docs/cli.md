---
title: "CLI guide"
description: "Install the beetree CLI and take a tree from idea to engine code."
weight: 15
---

The `beetree` CLI works on local `.beetree.yaml` files. It needs no account and no network.

## Install

On Linux and macOS, one line installs the latest release after checking its checksum:

```bash
curl -fsSL https://beetreecraft.com/install.sh | sh
beetree version
```

Prefer to download by hand, or on Windows? The [Download page](/get/) has an archive per platform and the steps to verify it. You can also build from source (needs Go):

{{< repoclone >}}

## Your first tree

```bash
beetree init                     # creates beetree.yaml, trees/ and subtrees/
beetree new guard                # creates trees/guard.beetree.yaml
beetree builder trees/guard.beetree.yaml   # interactive terminal editor
beetree validate trees/guard.beetree.yaml
beetree render trees/guard.beetree.yaml    # ASCII diagram
beetree simulate trees/guard.beetree.yaml  # dry run with a trace
beetree generate godot trees/guard.beetree.yaml
```

`generate` accepts `unity`, `unreal` or `godot`. Generated files marked as stubs are yours to edit and are never overwritten; everything else is regenerated each run.

## Step through a tree

A dry run shows one trace. To see how a tree reacts when a leaf fails, run `beetree simulate trees/guard.beetree.yaml --step`. It pauses at each leaf and asks you to choose Success, Failure or Running, then shows how the status moves up through the composites, so you can test the branch you worry about instead of the one the dry run happens to take. It uses the same walker as the TUI.

## Use it in CI

`beetree validate` exits non-zero on an invalid tree, so it works as a gate:

```bash
beetree validate trees/*.beetree.yaml
```

Run `beetree doctor` to check that a project is set up correctly.

## Command reference

Every command and flag is listed in the [CLI reference](/docs/cli-reference/).
