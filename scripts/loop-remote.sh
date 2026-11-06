#!/usr/bin/env bash
# Interactive Claude session named "beetree.com loop" with Remote Control, so the unattended
# loop can be watched and steered from the Claude phone app. The headless driver
# (scripts/agent-loop.sh) does the work; this session reads its state and runs the
# control-loop skill (status, pause, resume, finish).
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
NAME="beetree.com loop"
PROMPT='You are the remote console for the unattended beetreecraft loop. Use the control-loop skill: report the loop status now (heartbeat, recent log, latest release), then wait for the owner. Do not edit code here; the loop runs in its own clone.'
while true; do
  claude --remote-control "$NAME" --name "$NAME" "$PROMPT" || true
  echo "remote console exited; restarting in 30 s (Ctrl-C to stop)"
  sleep 30
done
