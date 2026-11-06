# Reconnaissance

Goal: one unit, fully understood, with a ledger that proves what was and was not covered.

1. Read the threat model in `docs/security/` and any earlier ledger in `.agent-loop/security/`. Reuse rows; do not start over.
2. Name the unit and its files. List entry points (routes, tools, workflows, scripts), who can reach each (anonymous, signed-in user, owner, API key by scope, CI), and the data each touches.
3. Mark trust boundaries: where untrusted bytes enter (request bodies, headers, claims, issue text, file uploads) and where they reach an interpreter, a query, a template, the filesystem, a shell, or another user's browser.
4. Write the ledger to `.agent-loop/security/<UTC timestamp>_<unit>_ledger.md`, one row per hunting target:

   | id | file or route | entry actor | untrusted input | sink | status |
   |----|---------------|-------------|-----------------|------|--------|

   Status starts `open` and becomes `hunted`, `candidate`, `confirmed`, `needs_validation`, `rejected` or `skipped (reason)`.
5. Keep the unit small enough to finish in one iteration. Rows left `open` stay in the ledger for the next run.
