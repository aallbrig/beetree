---
name: check-production
description: Check that beetreecraft.com and its latest release deploy are healthy, and fix production first when they are not. Use at the start of every session and loop iteration, and before any release.
---

# Check production

Production is healthy when every line below holds. Run them in parallel.

1. `curl -s -o /dev/null -w '%{http_code}' https://beetreecraft.com/` is 200, and so are `/app/`, `/docs/`, `/learn/building-blocks/` and `/schemas/beetree.schema.json`.
2. `https://beetreecraft.com/version.json` shows the latest release tag (`gh release list --limit 1`). A mismatch means the deploy is behind or failed.
3. `curl -s https://api.beetreecraft.com/v1/health` returns `{"status":"ok"}`; `/v1/me` and `POST /v1/trees` without a key return 401 (anonymous `GET /v1/trees` lists public trees and returns 200 by design), never 5xx.
4. The latest `Release deploy` run (`gh run list --workflow "Release deploy" --limit 1`) is `completed success` or still in progress.
5. Both `beetreecraft.com` and `www.beetreecraft.com` resolve (`dig +short`). A deploy with `DNS=false` or `WWW=false` deletes those records; `task deploy:backend` defaults to `DNS=true WWW=true`.

**Green**: report one line and carry on.

**Red**: `git stash push -u -m "loop: before production fix"`, read the failed run's logs (`gh run view <id> --log-failed`), fix the cause on a branch, run the gates, merge, release a patch, then check again. Pop the stash only when production is green.
