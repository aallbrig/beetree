# Findings schema

A findings file is a JSON array. `validate_findings.py` enforces it.

| field | required | notes |
|-------|----------|-------|
| `id` | yes | unique, for example `AUTH-1` |
| `unit` | yes | the audit unit |
| `title` | yes | neutral wording |
| `verdict` | yes | `confirmed`, `needs_validation` or `rejected` |
| `severity` | confirmed only | `low`, `medium`, `high`, `critical`; forbidden on the other verdicts |
| `trace` | confirmed only | list of `file:line` strings, entry point to sink |
| `observed` | confirmed only | the bounded observed result |
| `unresolved_fact` | needs_validation only | the one fact that decides it |
| `reason` | rejected only | why it was disproved |
| `fix` | confirmed only | neutral-wording fix plan and the regression test |

Text fields must not contain AWS account IDs (12 digits), absolute home paths or real email addresses.
