#!/usr/bin/env bash
# Exercise the loop driver against a fake `claude`: it must wait out a usage limit, retry,
# and carry on, and it must not announce anything while testing.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TMP="$(mktemp -d)"; trap 'rm -rf "$TMP"' EXIT

git init -q --bare "$TMP/origin.git"
git clone -q "$TMP/origin.git" "$TMP/clone" 2>/dev/null
(cd "$TMP/clone" && git checkout -q -b main && git -c user.name=t -c user.email=t@example.com commit -q --allow-empty -m init && git push -q origin main)

mkdir -p "$TMP/bin"
cat > "$TMP/bin/claude" <<FAKE
#!/usr/bin/env bash
n=\$(cat "$TMP/calls" 2>/dev/null || echo 0); echo \$((n + 1)) > "$TMP/calls"
if [ "\$n" -lt 2 ]; then echo "Claude AI usage limit reached. Your limit will reset at 3am."; exit 1; fi
echo "Released v9.9.9. Fake iteration done."
FAKE
chmod +x "$TMP/bin/claude"

out="$(PATH="$TMP/bin:$PATH" LOOP_STATE="$TMP/state" LOOP_CLONE="$TMP/clone" LOOP_SKIP_SETUP=1 LOOP_NO_NOTIFY=1 \
  LOOP_LIMIT_WAIT=1 LOOP_PAUSE=0 LOOP_ONCE=1 "$ROOT/scripts/agent-loop.sh" 2>&1)"
echo "$out" | tail -8
[ "$(cat "$TMP/calls")" = 3 ] || { echo "expected 2 limit retries then success (3 calls), got $(cat "$TMP/calls")" >&2; exit 1; }
grep -q 'hit a usage limit' <<<"$out" || { echo "limit was not detected" >&2; exit 1; }
grep -q 'iteration 1 finished rc=0' <<<"$out" || { echo "iteration did not finish after the limit" >&2; exit 1; }
echo "loop limit handling ok"
