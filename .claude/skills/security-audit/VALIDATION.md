# Validation

A fresh verifier (stronger model, no hunter context beyond the candidate) tries to disprove each candidate.

1. Re-trace the claim from entry point to sink yourself. Look for the guard the hunter missed: middleware, validation, an earlier return, an IAM condition.
2. Prefer proof by an existing or new test in the repo (the authorization matrix is the usual home). Never probe production or any live service.
3. Assign exactly one verdict:
   - `confirmed`: a complete source trace and a bounded observed result (a failing test, a precise reading of the code that no guard contradicts).
   - `needs_validation`: one exact fact is unresolved. Name it in `unresolved_fact`. No severity.
   - `rejected`: a guard exists, the input is not attacker-reachable, or it is only a missing defense-in-depth layer. Keep the reason; later runs skip it.
4. For `confirmed`, set severity as likelihood times impact and write the fix plan with neutral wording. A second fresh verifier re-checks every `confirmed` finding before the fix ships.
5. Disclosure: confirmed details stay in `.agent-loop/security/` until the fix is deployed.
