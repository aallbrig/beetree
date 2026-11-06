---
name: evaluate-website
description: Critically evaluate beetreecraft.com (home, builder, explore, learn, docs, forum) to find the most worthwhile improvement. Use when improving usability, the builder, docs quality, navigation or design.
---

# Evaluate the website

Surface: the Hugo site (`website/`) and the SPA (`web/`, served at `/app/`).

Probe, in order:
- Home page: does it say what the site is in five seconds, show something alive, and put sign up below the fold? Does it look modern in light and dark, on a phone and a wide screen?
- Builder: start from a template, change it, validate it, run it in 3D, save it to the agent inventory, export code. Count the clicks and note every hesitation.
- Learn and docs: follow a tutorial with the CLI side by side; every command and tree shown must work as written. Check links, the version footer and navigation from every page.
- Forum: post, report, reply; check that moderation exists and is visible.
- Accessibility basics: keyboard path, contrast, focus rings, labels.
- Speed: first load on a throttled connection.

## Method

1. **Use it as a stranger.** Run the real thing end to end as someone who has never seen it; capture the first confusing moment, the first dead end, the first thing that felt slow. Read `VISION.md` and `docs/AUDIENCES.md` for who that stranger is.
2. **Compare against the promise.** For each line of `VISION.md` this surface owns, mark it met, partly met or missing.
3. **Rank.** Score each gap by value to a visitor, how often they hit it, and effort. Keep the top 3.
4. **Pick one that ships tonight.** It must come with a gate: a test, a `check_site.sh` assertion, or a docs example CI validates with the real validator.
5. **Report** the ranked list, the chosen item and its gate, so the loop can file the rest as `loop-idea` issues.
