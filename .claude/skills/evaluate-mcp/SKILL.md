---
name: evaluate-mcp
description: Critically evaluate the BeeTreeCraft MCP server to find the most worthwhile improvement. Use when improving MCP tools, prompts, resources or their docs.
---

# Evaluate the MCP server

Surface: the MCP server (`cli/internal/api`, `/docs/mcp/`) as an AI client would use it.

Probe, in order:
- Connect a real MCP client and read only tool names and descriptions: would a model pick the right tool and fill its arguments correctly?
- Run the golden path: author a tree, validate it, simulate it, render it, generate engine code. Count the calls and the tokens spent on output that was not needed.
- Errors: do failures tell the model how to recover?
- Resources and prompts: are templates, the node reference and the format schema readable as resources? Is there a prompt for "design a behavior tree for this character"?
- Docs: `/docs/mcp/` setup steps must work verbatim for the common clients.

## Method

1. **Use it as a stranger.** Run the real thing end to end as someone who has never seen it; capture the first confusing moment, the first dead end, the first thing that felt slow. Read `VISION.md` and `docs/AUDIENCES.md` for who that stranger is.
2. **Compare against the promise.** For each line of `VISION.md` this surface owns, mark it met, partly met or missing.
3. **Rank.** Score each gap by value to a visitor, how often they hit it, and effort. Keep the top 3.
4. **Pick one that ships tonight.** It must come with a gate: a test, a `check_site.sh` assertion, or a docs example CI validates with the real validator.
5. **Report** the ranked list, the chosen item and its gate, so the loop can file the rest as `loop-idea` issues.
