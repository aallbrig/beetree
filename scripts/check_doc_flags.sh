#!/usr/bin/env bash
# Fail when a doc names a flag that the built binary's --help does not list.
# Covers inline `beetree <cmd> ... --flag` mentions in README.md and website/content,
# and the flag blocks under each heading of the generated CLI reference.
set -euo pipefail
cd "$(dirname "$0")/.."
bin=${BEETREE_BIN:-bin/beetree}
[ -x "$bin" ] || { echo "check_doc_flags: $bin missing; run 'task build:cli'" >&2; exit 1; }

declare -A helpcache
help_for() { # help_for "simulate" -> help text, empty when not a command
  local key=$1
  if [ -z "${helpcache[$key]+x}" ]; then
    # shellcheck disable=SC2086
    helpcache[$key]=$("$bin" $key --help 2>/dev/null || true)
  fi
  printf '%s' "${helpcache[$key]}"
}

# Longest leading run of words that is a real command path.
cmd_path() {
  local path="" w
  for w in "$@"; do
    case $w in -*|*[/\<\[\(\"\'.]*) break ;; esac
    w=${w//[^a-z0-9-]/}
    [ -n "$w" ] || break
    local next="${path:+$path }$w"
    if [ -n "$(help_for "$next")" ] && "$bin" $next --help >/dev/null 2>&1; then path=$next; else break; fi
  done
  printf '%s' "$path"
}

fail=0 checked=0
check_flag() { # file line path flag
  checked=$((checked + 1))
  if ! help_for "$3" | grep -qE -- "(^|[[:space:],])$4([[:space:]=,]|$)"; then
    echo "$1:$2: 'beetree $3' has no flag $4" >&2
    fail=1
  fi
}

while IFS= read -r f; do
  heading=""
  n=0
  while IFS= read -r line; do
    n=$((n + 1))
    if [[ $line =~ ^#+\ \`beetree([^\`]*)\` ]]; then
      heading=$(echo "${BASH_REMATCH[1]}" | xargs); continue
    fi
    if [[ $line =~ ^[[:space:]]*(-[a-zA-Z],[[:space:]]+)?(--[a-z][a-z0-9-]*)[[:space:]] && $f == *cli-reference.md ]]; then
      check_flag "$f" "$n" "$heading" "${BASH_REMATCH[2]}"; continue
    fi
    if [[ $line == *beetree* && $line =~ --[a-z] ]]; then
      rest=${line#*beetree }
      [ "$rest" != "$line" ] || continue
      # shellcheck disable=SC2086
      path=$(cmd_path ${rest//\`/ })
      [ -n "$path" ] || continue
      for flag in $(grep -oE -- '--[a-z][a-z0-9-]*' <<<"$line"); do
        check_flag "$f" "$n" "$path" "$flag"
      done
    fi
  done <"$f"
done < <(git ls-files 'README.md' 'website/content/*.md')

echo "check_doc_flags: $checked flag mentions checked"
exit $fail
