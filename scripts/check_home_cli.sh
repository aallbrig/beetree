#!/usr/bin/env bash
# Fail when the CLI transcript on the home page (#home-cli) differs from what the
# binary prints for the tree shown beside it (#home-tree). Visitors read it as real output.
set -euo pipefail
cd "$(dirname "$0")/.."
bin=$(realpath "${BEETREE_BIN:-bin/beetree}")
page=website/layouts/index.html
[ -x "$bin" ] || { echo "check_home_cli: $bin missing; run 'task build:cli'" >&2; exit 1; }

block() { # print the unescaped text of the <pre id=$1> block
  awk -v id="$1" '
    index($0, "id=\"" id "\"") { f = 1 }
    f { print }
    f && /<\/code><\/pre>/ { exit }' "$page" |
    sed -e 's/^.*<code>//' -e 's/<\/code><\/pre>.*$//' -e 's/&lt;/</g' -e 's/&gt;/>/g' -e 's/&quot;/"/g' -e 's/&amp;/\&/g'
}

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
block home-tree > "$tmp/guard.beetree.yaml"
[ -s "$tmp/guard.beetree.yaml" ] || { echo "check_home_cli: no #home-tree block in $page" >&2; exit 1; }
block home-cli > "$tmp/shown.txt"

cd "$tmp"
: > actual.txt
while IFS= read -r line; do
  case $line in
    '$ beetree '*)
      [ -s actual.txt ] && echo >> actual.txt
      echo "$line" >> actual.txt
      # shellcheck disable=SC2086
      "$bin" ${line#\$ beetree } >> actual.txt 2>&1 ;;
  esac
done < shown.txt
if ! diff -u shown.txt actual.txt; then
  echo "check_home_cli: home page CLI transcript is stale; paste the actual output above" >&2
  exit 1
fi
echo "check_home_cli: transcript matches"
