---
title: "CLI guide"
description: "Install the beetree CLI and take a tree from idea to engine code."
weight: 15
---

The `beetree` CLI works on local `.beetree.yaml` files. It needs no account and no network.

## Install

Download a binary for Linux, macOS or Windows from the [latest release](https://github.com/aallbrig/beetree/releases/latest), put it on your `PATH`, and check it:

```bash
beetree version
```

Or build from source (needs Go):

```bash
git clone https://github.com/aallbrig/beetree
cd beetree && make install
```

## Your first tree

```bash
beetree init                     # creates a project with trees/ and generated/
beetree new guard                # creates trees/guard.beetree.yaml
beetree builder trees/guard.beetree.yaml   # interactive terminal editor
beetree validate trees/guard.beetree.yaml
beetree render trees/guard.beetree.yaml    # ASCII diagram
beetree simulate trees/guard.beetree.yaml  # dry run with a trace
beetree generate godot trees/guard.beetree.yaml
```

`generate` accepts `unity`, `unreal` or `godot`. Generated files marked as stubs are yours to edit and are never overwritten; everything else is regenerated each run.

## Use it in CI

`beetree validate` exits non-zero on an invalid tree, so it works as a gate:

```bash
beetree validate trees/*.beetree.yaml
```

Run `beetree doctor` to check that a project is set up correctly.

## Command reference

Every command and flag is listed in the [CLI reference](/docs/cli-reference/).
