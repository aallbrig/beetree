#!/usr/bin/env bash
# Fail when the threat model names a file that does not exist, or lost its key sections.
set -euo pipefail
cd "$(dirname "$0")/.."

doc="$(ls docs/security/*_threat_model.md 2>/dev/null | tail -1 || true)"
[ -n "$doc" ] || { echo "no docs/security/*_threat_model.md" >&2; exit 1; }

for h in "## Assets" "## Actors" "## Trust boundaries" "## Top threats and controls"; do
  grep -qF "$h" "$doc" || { echo "$doc is missing section: $h" >&2; exit 1; }
done

fail=0
paths="$(grep -o '`[A-Za-z0-9_./-]*`' "$doc" | tr -d '`' | grep -E '^[A-Za-z0-9_.-]+(/[A-Za-z0-9_.-]+)*$' | grep -E '/|\.(md|go|yaml|yml|sh|toml)$' | sort -u)"
[ -n "$paths" ] || { echo "no file references found in $doc" >&2; exit 1; }
while IFS= read -r p; do
  [ "$p" = ".agent-loop/" ] && continue
  if [ ! -e "$p" ] && [ ! -e "$(dirname "$doc")/$p" ]; then echo "$doc names a missing file: $p" >&2; fail=1; fi
done <<< "$paths"
[ "$fail" = 0 ] && echo "threat model ok: $(wc -l <<< "$paths") references checked"
exit "$fail"
