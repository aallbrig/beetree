# Overnight loop iteration

You are one iteration of an unattended loop improving BeeTreeCraft (beetreecraft.com). Work alone in this clone, ship one release, report, and exit so the driver can start the next iteration.

Read first: `CLAUDE.md` (conventions, release policy), `VISION.md` (what to aspire to), `GLOSSARY.md`, the map issue named below.

## Steps

1. **Check production.** Invoke the `check-production` skill. On red, fix production first, as that skill says.
2. **Backlog first.** `gh issue list --label loop-idea --state open`. With 3 or more open ideas, skip to step 4 and take the best one for the focus surface below.
3. **Evaluate (backlog thin).** Dispatch up to four read-only subagents in parallel with the `Agent` tool, each with `model: "haiku"` and a prompt that tells it to invoke one skill and return its ranked ideas in under 300 words: `evaluate-builder`, `evaluate-website`, `evaluate-cli` (CLI and TUI), `evaluate-api` (API and MCP, with `evaluate-mcp` for the MCP half). Triage what comes back: drop duplicates of open issues, rank by value to a visitor against `VISION.md`, file the best 3 to 5 as issues labelled `loop-idea` (one line of why each helps a user, and the gate that would prove it). Keep the evaluators read-only; you write the issues.
4. **Choose and shape.** Take the best idea for this iteration's focus surface; map tickets that are unblocked win ties. When it is bigger than one iteration, split it into child issues (one releasable slice each, labelled `loop-idea`, the first one marked in its title with `[next]`) and ship the first slice now. Wayfinder's grilling needs a human, so leave its decision tickets to people.
5. **Build with its gate.** Every change ships with a test, a `check_site.sh` assertion, or a docs example that CI validates, so the gates stay representative of the feature. Run `task format:check` and `task ci` before committing.
6. **Ship.** Branch, commit (conventional commits, trailers from the harness), push, open a PR whose `## Summary` is written for a reader who was away (it becomes the release notes). Wait until `gh pr checks` lists the checks, then `task pr:merge -- N`, `git checkout main && git pull`, and `task release -- vX.Y.Z` with the next free patch (`gh release list --limit 1`; another run may have released meanwhile). Chain with `&&` under `set -o pipefail`. Close the issue you finished with a comment naming the release.
7. **Exit.** The deploy runs in CI; do not wait for it. Finish with one paragraph: the version released, what a visitor can now do, and the gate that proves it.

Keep an iteration to one releasable feature. The UI uses no emojis. Never commit secrets, API keys, account IDs or personal data (`CLAUDE.md`, Public Repo Hygiene). Infrastructure, IAM and DNS changes go through `task deploy:backend` with `DNS=true WWW=true`.
