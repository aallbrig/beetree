#!/usr/bin/env bash
# Unattended improvement loop. One short `claude -p` run per iteration, in its own clone
# so your dev checkout is never touched. Survives usage limits by sleeping and retrying,
# so work resumes as soon as the next window opens.
#
#   task agent:loop:start   run in a detached tmux session
#   task agent:loop:status  heartbeat, log tail, latest release
#   task agent:loop:stop    stop it
#
# Env: LOOP_MODEL (e.g. haiku), LOOP_PAUSE (s between iterations), LOOP_LIMIT_WAIT (s between
# limit retries), LOOP_LIMIT_RE, LOOP_TIMEOUT (s per iteration), LOOP_BUDGET (USD cap per
# iteration), LOOP_MAP_ISSUE, LOOP_ONCE=1, LOOP_DRY=1 (print the prompt, skip claude).
set -uo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STATE="$ROOT/.agent-loop"
CLONE="${LOOP_CLONE:-$ROOT/../beetree-loop}"
MAP="${LOOP_MAP_ISSUE:-16}"
PAUSE="${LOOP_PAUSE:-60}"
LIMIT_WAIT="${LOOP_LIMIT_WAIT:-600}"
LIMIT_RE="${LOOP_LIMIT_RE:-usage limit|rate limit|limit reached|limit will reset|quota|overloaded|429}"
SITE="https://beetreecraft.com"
FOCI=(builder docs website builder cli api docs mcp)

mkdir -p "$STATE"
exec 9>"$STATE/lock"
flock -n 9 || { echo "loop already running" >&2; exit 1; }

log() { printf '%s %s\n' "$(date -u +%FT%TZ)" "$*" | tee -a "$STATE/loop.log"; }
say() { [ "${LOOP_DRY:-}" = 1 ] && return 0; notify-send "BeeTreeCraft loop" "$1" 2>/dev/null || true; gh issue comment "$MAP" --body "$1" >/dev/null 2>&1 || true; }
release() { gh release list --limit 1 --json tagName -q '.[0].tagName' 2>/dev/null; }
heartbeat() { printf '{"at":"%s","iteration":%s,"state":"%s","release":"%s"}\n' "$(date -u +%FT%TZ)" "$1" "$2" "$(release)" > "$STATE/status.json"; }

production() {
  local code run
  code="$(curl -s -o /dev/null -w '%{http_code}' --max-time 20 "$SITE/" || echo 000)"
  run="$(gh run list --workflow 'Release deploy' --limit 1 --json conclusion,status -q '.[0] | "\(.status) \(.conclusion)"' 2>/dev/null)"
  if [ "$code" != 200 ]; then echo "RED: $SITE returned $code"
  elif [[ "$run" == completed\ failure* ]]; then echo "RED: the latest Release deploy run failed"
  else echo "green (site 200, latest deploy: ${run:-unknown})"; fi
}

setup_clone() {
  if [ ! -d "$CLONE/.git" ]; then
    git clone -q "$(git -C "$ROOT" remote get-url origin)" "$CLONE" || return 1
  fi
  (cd "$CLONE" && [ -d web/node_modules ] || (cd web && npm ci --silent)) || return 1
  (cd "$CLONE" && command -v pre-commit >/dev/null && pre-commit install >/dev/null 2>&1) || true
}

sync_clone() {
  cd "$CLONE" || return 1
  git fetch -q origin
  if [ -z "$(git status --porcelain)" ]; then git checkout -q main && git merge -q --ff-only origin/main; fi
}

build_prompt() {
  local n="$1" focus="$2"
  cat "$ROOT/scripts/loop-prompt.md"
  printf '\n## This iteration\n\n- Iteration: %s\n- Focus surface: %s (skill: evaluate-%s)\n- Map issue: #%s\n- Production: %s\n- Latest release: %s\n- Clone state:\n```\n%s\n```\n' \
    "$n" "$focus" "$( [ "$focus" = builder ] || [ "$focus" = docs ] && echo website || echo "$focus" )" "$MAP" "$(production)" "$(release)" "$(cd "$CLONE" && git status --short --branch | head -15)"
}

n="$(cat "$STATE/iteration" 2>/dev/null || echo 0)"
setup_clone || { log "clone setup failed"; exit 1; }
log "loop up (model=${LOOP_MODEL:-default}, clone=$CLONE)"
say "Loop is up (model: ${LOOP_MODEL:-default}). Production: $(production). Latest release: $(release)."

while :; do
  n=$((n + 1)); echo "$n" > "$STATE/iteration"
  focus="${FOCI[$(( (n - 1) % ${#FOCI[@]} ))]}"
  sync_clone || { log "sync failed"; sleep "$PAUSE"; continue; }
  before="$(release)"
  prompt="$(build_prompt "$n" "$focus")"
  if [ "${LOOP_DRY:-}" = 1 ]; then printf '%s\n' "$prompt"; exit 0; fi

  while :; do
    heartbeat "$n" "running:$focus"
    log "iteration $n ($focus) starting"
    out="$STATE/iter-$n.log"
    (cd "$CLONE" && timeout "${LOOP_TIMEOUT:-5400}" claude -p "$prompt" ${LOOP_MODEL:+--model "$LOOP_MODEL"} \
      ${LOOP_BUDGET:+--max-budget-usd "$LOOP_BUDGET"} --dangerously-skip-permissions --no-session-persistence) > "$out" 2>&1
    rc=$?
    if grep -Eiq "$LIMIT_RE" <<<"$(tail -c 600 "$out")" && { [ "$rc" -ne 0 ] || [ "$(wc -c < "$out")" -lt 800 ]; }; then
      heartbeat "$n" "waiting-for-limit"
      log "iteration $n hit a usage limit; retrying in ${LIMIT_WAIT}s"
      [ -f "$STATE/limit-announced" ] || { say "Loop paused at a usage limit; it retries every $((LIMIT_WAIT / 60)) min and resumes when the window opens."; touch "$STATE/limit-announced"; }
      sleep "$LIMIT_WAIT"; continue
    fi
    rm -f "$STATE/limit-announced"; break
  done

  after="$(release)"
  log "iteration $n finished rc=$rc release=$after"
  heartbeat "$n" "idle"
  if [ "$after" != "$before" ]; then
    say "Shipped $after (iteration $n, focus: $focus). $(tail -c 700 "$out" | tr '\n' ' ')"
  elif [ "$rc" -ne 0 ]; then
    log "iteration $n failed (rc=$rc); backing off"; sleep $((PAUSE * 5))
  fi
  [ "${LOOP_ONCE:-}" = 1 ] && exit 0
  sleep "$PAUSE"
done
