#!/usr/bin/env bash
# Fail when the "Your first tree" block in docs/cli.md does not run as written,
# or a path named in a "creates ..." comment is absent after its command ran.
set -euo pipefail
cd "$(dirname "$0")/.."
bin=$(realpath "${BEETREE_BIN:-bin/beetree}")
doc=website/content/docs/cli.md
[ -x "$bin" ] || { echo "check_doc_tutorial: $bin missing; run 'task build:cli'" >&2; exit 1; }

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
block=$(awk '/^## Your first tree/ {s=1; next} s && /^```bash/ {f=1; next} f && /^```/ {exit} f' "$doc")
[ -n "$block" ] || { echo "check_doc_tutorial: no bash block under 'Your first tree'" >&2; exit 1; }

fail=0 ran=0
cd "$tmp"
while IFS= read -r line; do
  cmd=${line%%#*}
  comment=
  [[ $line == *'#'* ]] && comment=${line#*#}
  set -- $cmd
  [ "${1:-}" = beetree ] || continue
  [ "${2:-}" = builder ] && continue # interactive TUI
  shift
  ran=$((ran + 1))
  if ! out=$("$bin" "$@" 2>&1); then
    echo "$doc: 'beetree $*' fails:" >&2; echo "$out" >&2; fail=1; continue
  fi
  if [[ $comment == *creates* ]]; then
    for tok in $(echo "${comment#*creates}" | tr ',' ' '); do
      case $tok in
        */|*.yaml) [ -e "$tok" ] || { echo "$doc: 'beetree $*' did not create $tok" >&2; fail=1; } ;;
      esac
    done
  fi
done <<< "$block"
[ "$ran" -gt 0 ] || { echo "check_doc_tutorial: ran no commands" >&2; exit 1; }
[ "$fail" -eq 0 ] && echo "check_doc_tutorial: $ran commands ok"
exit "$fail"
