---
name: evaluate-api
description: Critically evaluate the BeeTreeCraft REST API to find the most worthwhile improvement. Use when improving endpoints, the OpenAPI spec, API docs, errors or scaling.
---

# Evaluate the REST API

Surface: `cli/internal/api`, the OpenAPI spec and `/docs/api/`.

Probe, in order:
- Follow `/docs/api/` as a developer with only curl: sign up, get a token, create, fetch, update and delete a tree. Every example must run as written.
- Consistency: naming, pagination, filtering, error shape, status codes, idempotency.
- Spec truth: the OpenAPI document must match behavior; try each documented response, including the failures.
- Abuse and scale: rate limits, payload size limits, authorization on every route, hot DynamoDB access patterns.
- Capability gaps against the website and CLI: templates, agent inventory, forum reports, moderation.

## Method

1. **Use it as a stranger.** Run the real thing end to end as someone who has never seen it; capture the first confusing moment, the first dead end, the first thing that felt slow. Read `VISION.md` and `docs/AUDIENCES.md` for who that stranger is.
2. **Compare against the promise.** For each line of `VISION.md` this surface owns, mark it met, partly met or missing.
3. **Rank.** Score each gap by value to a visitor, how often they hit it, and effort. Keep the top 3.
4. **Pick one that ships tonight.** It must come with a gate: a test, a `check_site.sh` assertion, or a docs example CI validates with the real validator.
5. **Report** the ranked list, the chosen item and its gate, so the loop can file the rest as `loop-idea` issues.
