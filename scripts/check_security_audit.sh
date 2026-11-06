#!/usr/bin/env bash
# Gate for the security-audit skill: its files exist, the validator's tests pass, and the
# validator accepts a sample findings file and rejects a malformed one.
set -euo pipefail
cd "$(dirname "$0")/.."
d=.claude/skills/security-audit
for f in SKILL.md RECONNAISSANCE.md HUNTING.md VALIDATION.md FINDINGS_SCHEMA.md validate_findings.py; do
  [ -f "$d/$f" ] || { echo "missing $d/$f" >&2; exit 1; }
done
(cd "$d" && python3 -B -m unittest -q test_validate_findings)
tmp="$(mktemp -d)"; trap 'rm -rf "$tmp"' EXIT
cat > "$tmp/ok.json" <<'JSON'
[{"id":"DRY-1","unit":"dry-run","title":"Harden example","verdict":"needs_validation","unresolved_fact":"whether the guard runs first"}]
JSON
echo '[{"id":"DRY-2","unit":"x","title":"t","verdict":"confirmed"}]' > "$tmp/bad.json"
python3 -B "$d/validate_findings.py" "$tmp/ok.json"
if python3 -B "$d/validate_findings.py" "$tmp/bad.json" 2>/dev/null; then echo "validator accepted a bad file" >&2; exit 1; fi
echo "security-audit skill ok"
