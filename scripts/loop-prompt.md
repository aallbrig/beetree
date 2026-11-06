# Overnight loop iteration

You are one iteration of an unattended loop improving BeeTreeCraft (beetreecraft.com). Work alone in this clone. Finish with one shipped patch release and a short report.

Read first: `CLAUDE.md` (conventions, release policy), `VISION.md` (what to aspire to), `GLOSSARY.md`, and the map issue named below.

## Steps

1. **Production first.** The driver reports production status below. When it says RED, `git stash` any work in progress, fix production (read the failed `Release deploy` run's logs), release a patch, then continue. Unfinished work in this clone is listed by `git status`; resume it when production is green.
2. **Pick the work.** Use the focus surface below. Invoke its `evaluate-*` skill to use the product critically and rank what to improve. File the top 2 or 3 ideas as GitHub issues labelled `loop-idea` (one line of why each helps a user), then take the best one. Map tickets that are unblocked take priority over new ideas when they fit the focus.
3. **Build it with its gate.** Every change ships with a test, a `check_site.sh` assertion, or a docs example that CI validates, so the gates stay representative of the feature. Run `task format:check` and `task ci` before committing.
4. **Ship.** Branch, commit (conventional commits, trailers from the harness), push, open a PR whose `## Summary` is written for a reader who was away (it becomes the release notes). Wait until `gh pr checks` lists the checks, then `task pr:merge -- N`, `git checkout main && git pull`, and `task release -- vX.Y.Z` (patch by default; bigger when the change deserves it). Chain with `&&` under `set -o pipefail`.
5. **Move on.** The deploy runs in CI; do not wait for it. If it fails the next iteration fixes it.
6. **Report.** Finish with one paragraph: the version released, what a visitor can now do, and the gate that proves it.

Keep an iteration to one releasable feature. The UI uses no emojis. Infrastructure, IAM and DNS changes go through `task deploy:backend` with `DNS=true WWW=true`.
