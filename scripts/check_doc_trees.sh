#!/usr/bin/env bash
# Fail when a ```yaml fence in the docs or README holds a tree (a top-level
# `tree:` key) that `beetree validate` rejects. Visitors copy these examples.
set -euo pipefail
cd "$(dirname "$0")/.."
bin=${BEETREE_BIN:-bin/beetree}
[ -x "$bin" ] || { echo "check_doc_trees: $bin missing; run 'task build:cli'" >&2; exit 1; }

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

fail=0 checked=0
while IFS= read -r f; do
  rm -f "$tmp"/*.beetree.yaml
  awk -v dir="$tmp" '
    /^```yaml[[:space:]]*$/ { infence = 1; n++; out = sprintf("%s/%d.beetree.yaml", dir, n); start[n] = NR; next }
    /^```/ && infence { infence = 0; close(out); next }
    infence { print > out }
  ' "$f"
  for ex in "$tmp"/*.beetree.yaml; do
    [ -e "$ex" ] || continue
    grep -q '^tree:' "$ex" || continue
    checked=$((checked + 1))
    if ! out=$("$bin" validate "$ex" 2>&1); then
      echo "$f: yaml example $(basename "$ex" .beetree.yaml) fails 'beetree validate':" >&2
      echo "$out" >&2
      fail=1
    fi
  done
done < <(git ls-files 'README.md' 'website/content/*.md')

[ "$checked" -gt 0 ] || { echo "check_doc_trees: no tree examples found" >&2; exit 1; }
echo "check_doc_trees: $checked tree examples validated"
exit $fail
