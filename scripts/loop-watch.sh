#!/usr/bin/env bash
# Show what the loop is doing right now: the latest iteration's messages and tool calls.
# Usage: scripts/loop-watch.sh [N lines, default 40]   (task agent:loop:watch)
set -uo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
STATE="${LOOP_STATE:-$ROOT/.agent-loop}"
cat "$STATE/status.json" 2>/dev/null || echo "no heartbeat yet"
latest="$(ls -t "$STATE"/iter-*.jsonl 2>/dev/null | head -1)"
[ -n "$latest" ] || { echo "no iteration log yet"; exit 0; }
echo "--- $(basename "$latest")"
jq -r 'select(.type=="assistant") | .message.content[]?
  | if .type=="text" then "SAY  " + (.text | gsub("\n"; " ") | .[0:220])
    elif .type=="tool_use" then "TOOL " + .name + " " + ((.input | tostring) | .[0:140])
    else empty end' "$latest" 2>/dev/null | tail -n "${1:-40}"
