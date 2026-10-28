#!/usr/bin/env bash
# Manage the beetree tmuxinator session: up | attach | recycle | start | stop
set -euo pipefail

NAME=beetree
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CONFIG="$ROOT/.tmuxinator/$NAME.yml"

has_session() { tmux has-session -t "=$NAME" 2>/dev/null; }

case "${1:-up}" in
  up)
    if has_session; then tmux attach -t "=$NAME"; else tmuxinator start --project-config "$CONFIG"; fi
    ;;
  attach)
    has_session || { echo "no '$NAME' session; run 'task dev'" >&2; exit 1; }
    tmux attach -t "=$NAME"
    ;;
  recycle)
    has_session && tmux kill-session -t "=$NAME"
    tmuxinator start --project-config "$CONFIG"
    ;;
  start)
    has_session || tmuxinator start --project-config "$CONFIG" --no-attach
    ;;
  stop)
    has_session && tmux kill-session -t "=$NAME" || true
    ;;
  *)
    echo "usage: $0 {up|attach|recycle|start|stop}" >&2
    exit 2
    ;;
esac
