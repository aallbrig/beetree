---
name: evaluate-cli
description: Critically evaluate the beetree CLI and TUI to find the most worthwhile improvement. Use when improving the CLI, the TUI editor, command help, errors or codegen workflow.
---

# Evaluate the CLI and TUI

Surface: `beetree` commands (`cli/cmd`) and the TUI editor (`cli/internal/tui`).

Probe, in order:
- `beetree --help` and each subcommand's `--help`: are names guessable, examples present, flags consistent?
- First run with no files: does `init` or `examples` get a newcomer to a valid tree in under two minutes?
- Error messages: feed it broken YAML, a missing file, an unknown node type; each error should say what to do next.
- The TUI: open a tree, add, move and delete nodes, undo, save; look for dead keys, missing hints and unreadable layouts at 80 columns.
- The loop from tree to engine code: `validate`, `simulate`, `render`, `codegen`; the output must compile in the target engine's conventions.
- Parity: anything the website or API can do that the CLI cannot, and the reverse.

## Method

1. **Use it as a stranger.** Run the real thing end to end as someone who has never seen it; capture the first confusing moment, the first dead end, the first thing that felt slow. Read `VISION.md` and `docs/AUDIENCES.md` for who that stranger is.
2. **Compare against the promise.** For each line of `VISION.md` this surface owns, mark it met, partly met or missing.
3. **Rank.** Score each gap by value to a visitor, how often they hit it, and effort. Keep the top 3.
4. **Pick one that ships tonight.** It must come with a gate: a test, a `check_site.sh` assertion, or a docs example CI validates with the real validator.
5. **Report** the ranked list, the chosen item and its gate, so the loop can file the rest as `loop-idea` issues.
